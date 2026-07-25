package main

import (
	"context"
	"fmt"

	"github.com/chrisjpalmer/container-agent/container-agent/internal/dagger"
)

type Claude struct {
	AnthropicAuthToken *dagger.Secret
	AnthropicBaseUrl   string
	LlamaCpp           *dagger.Service
	Src                *dagger.Directory
	ToolVersions       *dagger.File
}

func (m *ContainerAgent) Claude(
	// +optional
	llamacpp *dagger.Service,
	// +optional
	authToken *dagger.Secret,
	// +optional
	baseUrl string,
) *Claude {
	return &Claude{
		AnthropicAuthToken: authToken,
		AnthropicBaseUrl:   baseUrl,
		LlamaCpp:           llamacpp,
		Src:                m.Src,
		ToolVersions:       m.ToolVersions,
	}
}

// +cache="never"
func (m *Claude) Agent(ctx context.Context) (*dagger.Changeset, error) {
	c, err := m.base(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base container: %w", err)
	}

	return changesFromSrc(
		c.Terminal(dagger.ContainerTerminalOpts{Cmd: []string{"claude", "--permission-mode=acceptEdits"}}),
		m.Src,
	), nil
}

// +cache="never"
func (m *Claude) Sh(ctx context.Context) (*dagger.Changeset, error) {
	c, err := m.base(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base container: %w", err)
	}

	return changesFromSrc(c.Terminal(), m.Src), nil
}

func (m *Claude) base(ctx context.Context) (*dagger.Container, error) {
	ctr := base()

	ctr = withAsdfPlugins(ctr)

	ctr = withClaude(ctr)

	ctr = withToolVersions(ctr, m.ToolVersions)

	ctr = withAsdfInstall(ctr)

	ctr, err := withGoSetup(ctx, ctr)
	if err != nil {
		return nil, fmt.Errorf("failed to set up go correctly: %w", err)
	}

	// claude setup
	claude := dag.CurrentModule().Source().Directory(".claude")

	if m.LlamaCpp != nil {
		ctr = ctr.WithEnvVariable("CLAUDE_CODE_ATTRIBUTION_HEADER", "0").
			WithEnvVariable("CLAUDE_CODE_ENABLE_TELEMETRY", "0").
			WithEnvVariable("ANTHROPIC_AUTH_TOKEN", "x").
			WithEnvVariable("ANTHROPIC_BASE_URL", "http://llamacpp:8080").
			WithServiceBinding("llamacpp", m.LlamaCpp)
	} else {
		ctr = ctr.WithSecretVariable("ANTHROPIC_AUTH_TOKEN", m.AnthropicAuthToken).
			WithEnvVariable("ANTHROPIC_BASE_URL", m.AnthropicBaseUrl).
			WithEnvVariable("CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY", "1")
	}

	ctr = ctr.WithDirectory("/root/.claude", claude)

	return withSource(ctr, m.Src), nil
}

func withClaude(c *dagger.Container) *dagger.Container {
	return c.WithExec([]string{"sh", "-c", "curl -fsSL https://claude.ai/install.sh | bash"})
}
