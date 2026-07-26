package main

import (
	"context"
	"fmt"

	"github.com/chrisjpalmer/container-agent/container-agent/internal/dagger"
)

type Opencode struct {
	LlamaCpp     *dagger.Service
	Src          *dagger.Directory
	ToolVersions *dagger.File
}

func (m *ContainerAgent) Opencode(
	llamacpp *dagger.Service,
) *Opencode {
	return &Opencode{
		LlamaCpp:     llamacpp,
		Src:          m.Src,
		ToolVersions: m.ToolVersions,
	}
}

// +cache="never"
func (m *Opencode) Agent(ctx context.Context) (*dagger.Changeset, error) {
	c, err := m.base(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base container: %w", err)
	}

	return changesFromSrc(
		c.Terminal(dagger.ContainerTerminalOpts{Cmd: []string{"opencode", "--model", "llama.cpp/default"}}),
		m.Src,
	), nil
}

// +cache="never"
func (m *Opencode) Sh(ctx context.Context) (*dagger.Changeset, error) {
	c, err := m.base(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base container: %w", err)
	}

	return changesFromSrc(c.Terminal(), m.Src), nil
}

func (m *Opencode) base(ctx context.Context) (*dagger.Container, error) {
	ctr, err := base(ctx)
	if err != nil {
		return nil, fmt.Errorf("error configuring base: %w", err)
	}

	ctr = withAsdfPlugins(ctr)

	ctr = withOpencode(ctr)

	ctr = withToolVersions(ctr, m.ToolVersions)

	ctr = withAsdfInstall(ctr)

	ctr, err = withGoSetup(ctx, ctr)
	if err != nil {
		return nil, fmt.Errorf("failed to set up go correctly: %w", err)
	}

	// opencode setup
	cfg := dag.CurrentModule().Source().Directory("opencode")

	ctr = ctr.WithDirectory("/root/.config/opencode", cfg).
		WithServiceBinding("llamacpp", m.LlamaCpp)

	return withSource(ctr, m.Src), nil
}

func withOpencode(c *dagger.Container) *dagger.Container {
	c = c.WithNewFile("/root/.tool-versions", "nodejs 24.18.0")

	c = withAsdfInstall(c)

	return c.WithExec([]string{"npm", "install", "-g", "opencode-ai"})
}
