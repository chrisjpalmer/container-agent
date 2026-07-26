package main

import (
	"context"
	"fmt"

	"github.com/chrisjpalmer/container-agent/container-agent/internal/dagger"
)

type Claude struct {
	AnthropicAuthToken *dagger.Secret
	AnthropicBaseUrl   string
	AwsProfile         string
	AwsCredentials     *dagger.Secret
	DefaultModel       string
	GitName            string
	GitEmail           string
	GithubToken        *dagger.Secret
	LlamaCpp           *dagger.Service
	Src                *dagger.Directory
	ToolVersions       *dagger.File
	SSH                *dagger.Socket
	SSHRewrite         string
	GoPrivate          string
}

func (m *ContainerAgent) Claude(
	// +optional
	awsProfile string,
	// +optional
	awsCredentials *dagger.Secret,
	// +optional
	defaultModel string,
	// +optional
	githubToken *dagger.Secret,
	// +optional
	gitName string,
	// +optional
	gitEmail string,
	// +optional
	llamacpp *dagger.Service,
	// +optional
	authToken *dagger.Secret,
	// +optional
	baseUrl string,
	// +optional
	ssh *dagger.Socket,
	// +optional
	sshRewrite string,
	// +optional
	goPrivate string,
) *Claude {
	return &Claude{
		AnthropicAuthToken: authToken,
		AnthropicBaseUrl:   baseUrl,
		AwsProfile:         awsProfile,
		AwsCredentials:     awsCredentials,
		DefaultModel:       defaultModel,
		GitName:            gitName,
		GitEmail:           gitEmail,
		GithubToken:        githubToken,
		LlamaCpp:           llamacpp,
		Src:                m.Src,
		ToolVersions:       m.ToolVersions,
		SSH:                ssh,
		SSHRewrite:         sshRewrite,
		GoPrivate:          goPrivate,
	}
}

// +cache="never"
func (m *Claude) Agent(ctx context.Context) (*dagger.Changeset, error) {
	c, err := m.base(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base container: %w", err)
	}

	args := []string{"claude", "--permission-mode=acceptEdits"}

	if m.DefaultModel != "" {
		args = append(args, "--model="+m.DefaultModel)
	}

	return changesFromSrc(
		c.Terminal(dagger.ContainerTerminalOpts{Cmd: args, ExperimentalPrivilegedNesting: true}),
		m.Src,
	), nil
}

// +cache="never"
func (m *Claude) Sh(ctx context.Context) (*dagger.Changeset, error) {
	c, err := m.base(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base container: %w", err)
	}

	return changesFromSrc(c.Terminal(dagger.ContainerTerminalOpts{ExperimentalPrivilegedNesting: true}), m.Src), nil
}

func (m *Claude) base(ctx context.Context) (*dagger.Container, error) {
	ctr, err := base(ctx)
	if err != nil {
		return nil, fmt.Errorf("error configuring base: %w", err)
	}

	ctr = withClaude(ctr)

	ctr = withAsdfPlugins(ctr)

	ctr = withToolVersions(ctr, m.ToolVersions)

	ctr = withAsdfInstall(ctr)

	ctr, err = withGoSetup(ctx, ctr)
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

	if m.GithubToken != nil {
		ctr = ctr.WithSecretVariable("GITHUB_TOKEN", m.GithubToken)
	}

	if m.SSH != nil {
		ctr, err = withGithubSSH(ctr, m.SSH, m.GoPrivate, m.SSHRewrite)
		if err != nil {
			return nil, fmt.Errorf("failed to set up github ssh correctly: %w", err)
		}
	}

	if m.GitName != "" && m.GitEmail != "" {
		ctr = withGitConfig(ctr, m.GitName, m.GitEmail)
	}

	if m.AwsProfile != "" && m.AwsCredentials != nil {
		ctr = withAwsCredentials(ctr, m.AwsProfile, m.AwsCredentials)
	}

	ctr = ctr.WithDirectory("/root/.claude", claude)

	return withSource(ctr, m.Src), nil
}

func withClaude(c *dagger.Container) *dagger.Container {
	return c.WithExec([]string{"sh", "-c", "curl -fsSL https://claude.ai/install.sh | bash"})
}
