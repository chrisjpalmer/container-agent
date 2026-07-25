// A generated module for ClaudeAgent functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/pi-agent/internal/dagger"
	"fmt"
	"strings"
	"time"
)

type ClaudeAgent struct {
	AnthropicAuthToken *dagger.Secret
	AnthropicBaseUrl   string
	Src                *dagger.Directory
	ToolVersions       *dagger.File
}

func New(
	ws *dagger.Workspace,
	anthropicAuthToken *dagger.Secret,
	anthropicBaseUrl string,
) *ClaudeAgent {
	src := ws.Directory("/", dagger.WorkspaceDirectoryOpts{Exclude: []string{"claude-agent", "dagger.json"}})

	toolVersions := ws.File(".tool-versions")

	return &ClaudeAgent{
		AnthropicAuthToken: anthropicAuthToken,
		AnthropicBaseUrl:   anthropicBaseUrl,
		Src:                src,
		ToolVersions:       toolVersions,
	}
}

// +cache="never"
func (m *ClaudeAgent) Claude(ctx context.Context) (*dagger.Changeset, error) {
	c, err := m.base(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base container: %w", err)
	}

	return m.changes(
		c.Terminal(dagger.ContainerTerminalOpts{Cmd: []string{"claude", "--permission-mode=acceptEdits"}}),
	), nil
}

// +cache="never"
func (m *ClaudeAgent) Sh(ctx context.Context) (*dagger.Changeset, error) {
	c, err := m.base(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base container: %w", err)
	}

	return m.changes(
		c.Terminal(),
	), nil
}

func (m *ClaudeAgent) changes(c *dagger.Container) *dagger.Changeset {
	return c.WithExec([]string{"rm", "-rf", "/src"}).
		WithExec([]string{"cp", "-R", "/work", "/src"}).
		WithoutMount("/work").
		Directory("/src").
		Changes(m.Src)
}

func (m *ClaudeAgent) base(ctx context.Context) (*dagger.Container, error) {
	cache := dag.CacheVolume("my-work" + time.Now().String())

	settings := `
{
  "theme":"dark",
  "permissions": {
    "allow": [
      "Bash(*)",
      "Read(*)",
      "Edit(*)",
      "Cd(*)"
    ]
  }
}`

	ctr := dag.Container().
		From("ubuntu:latest@sha256:3131b4cc82a783df6c9df078f86e01819a13594b865c2cad47bd1bca2b7063bb").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "curl", "git", "-y"}).
		WithExec([]string{"sh", "-c", "curl -fsSL https://claude.ai/install.sh | bash"}).
		WithFile("/usr/bin/asdf", asdf()).
		WithEnvVariable("PATH", "/root/.asdf/shims:${PATH}", dagger.ContainerWithEnvVariableOpts{Expand: true}).
		WithEnvVariable("PATH", "/root/.local/bin:${PATH}", dagger.ContainerWithEnvVariableOpts{Expand: true}).

		// asdf install
		WithExec([]string{"asdf", "plugin", "add", "golang"}).
		WithWorkdir("/src").
		WithFile(".tool-versions", m.ToolVersions).
		WithExec([]string{"asdf", "install"})

	envStr, err := ctr.WithExec([]string{"go", "env"}).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run go env: %w", err)
	}

	env := parseEnv(envStr)

	// goModCache, err := ctr.EnvVariable(ctx, "GOMODCACHE")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to retrieve env variable: %w", err)
	// }

	// goCache, err := ctr.EnvVariable(ctx, "GOCACHE")
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to retrieve env variable: %w", err)
	// }

	return ctr.
		// set up go caches
		WithMountedCache(env["GOMODCACHE"], dag.CacheVolume("go-mod-cache")).
		WithMountedCache(env["GOCACHE"], dag.CacheVolume("go-build-cache")).

		// install go tools
		WithExec([]string{"go", "install", "golang.org/x/tools/gopls@latest"}).

		// claude setup
		WithSecretVariable("ANTHROPIC_AUTH_TOKEN", m.AnthropicAuthToken).
		WithEnvVariable("ANTHROPIC_BASE_URL", m.AnthropicBaseUrl).
		WithEnvVariable("CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY", "1").
		WithNewFile("/root/.claude/settings.json", settings).

		// source
		WithDirectory(".", m.Src).
		WithMountedCache("/work", cache).
		WithExec([]string{"cp", "-R", "/src/.", "/work"}).
		WithWorkdir("/work"), nil
}

func asdf() *dagger.File {
	return dag.Container().
		From("golang:latest@sha256:3aff6657219a4d9c14e27fb1d8976c49c29fddb70ba835014f477e1c70636647").
		WithExec([]string{"go", "install", "github.com/asdf-vm/asdf/cmd/asdf@v0.20.0"}).
		File("/go/bin/asdf")
}

func parseEnv(env string) map[string]string {

	ll := strings.Lines(env)

	m := make(map[string]string)

	for l := range ll {
		parts := strings.Split(l, "=")
		if len(parts) != 2 {
			continue
		}
		key := strings.Trim(parts[0], "\n")
		value := strings.Trim(parts[1], "'\n")

		m[key] = value
	}

	return m
}
