package main

import (
	"context"
	"fmt"

	"github.com/chrisjpalmer/container-agent/container-agent/internal/dagger"
)

type Pi struct {
	LlamaCpp     *dagger.Service
	Src          *dagger.Directory
	ToolVersions *dagger.File
}

func (m *ContainerAgent) Pi(
	llamacpp *dagger.Service,
) *Pi {
	return &Pi{
		LlamaCpp:     llamacpp,
		Src:          m.Src,
		ToolVersions: m.ToolVersions,
	}
}

// +cache="never"
func (m *Pi) Agent(ctx context.Context) (*dagger.Changeset, error) {
	c, err := m.base(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base container: %w", err)
	}

	return changesFromSrc(
		c.Terminal(dagger.ContainerTerminalOpts{Cmd: []string{"pi"}}),
		m.Src,
	), nil
}

// +cache="never"
func (m *Pi) Sh(ctx context.Context) (*dagger.Changeset, error) {
	c, err := m.base(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup base container: %w", err)
	}

	return changesFromSrc(c.Terminal(), m.Src), nil
}

func (m *Pi) base(ctx context.Context) (*dagger.Container, error) {
	ctr, err := base(ctx)
	if err != nil {
		return nil, fmt.Errorf("error configuring base: %w", err)
	}

	ctr = withAsdfPlugins(ctr)

	ctr = withPi(ctr)

	ctr = withToolVersions(ctr, m.ToolVersions)

	ctr = withAsdfInstall(ctr)

	ctr, err = withGoSetup(ctx, ctr)
	if err != nil {
		return nil, fmt.Errorf("failed to set up go correctly: %w", err)
	}

	// pi setup
	pi := dag.CurrentModule().Source().Directory(".pi")

	ctr = ctr.WithDirectory("/root/.pi", pi).
		WithServiceBinding("llamacpp", m.LlamaCpp)

	return withSource(ctr, m.Src), nil
}

func withPi(c *dagger.Container) *dagger.Container {
	c = c.WithNewFile("/root/.tool-versions", "nodejs 24.18.0")

	c = withAsdfInstall(c)

	return c.WithExec([]string{"npm", "install", "-g", "--ignore-scripts", "@earendil-works/pi-coding-agent@0.82.0"})
}
