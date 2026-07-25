package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chrisjpalmer/container-agent/container-agent/internal/dagger"
)

type ContainerAgent struct {
	ToolVersions *dagger.File
	Src          *dagger.Directory
}

func New(ws *dagger.Workspace) *ContainerAgent {
	return &ContainerAgent{
		ToolVersions: ws.File(".tool-versions"),
		Src:          ws.Directory("/", dagger.WorkspaceDirectoryOpts{Exclude: []string{"container-agent", "dagger.json"}}),
	}
}

func base() *dagger.Container {
	return dag.Container().
		From("ubuntu:latest@sha256:3131b4cc82a783df6c9df078f86e01819a13594b865c2cad47bd1bca2b7063bb").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "curl", "git", "-y"}).
		WithFile("/usr/bin/asdf", asdf()).
		WithEnvVariable("PATH", "/root/.asdf/shims:${PATH}", dagger.ContainerWithEnvVariableOpts{Expand: true}).
		WithEnvVariable("PATH", "/root/.local/bin:${PATH}", dagger.ContainerWithEnvVariableOpts{Expand: true}).
		WithWorkdir("/src")
}

func asdf() *dagger.File {
	return dag.Container().
		From("golang:latest@sha256:3aff6657219a4d9c14e27fb1d8976c49c29fddb70ba835014f477e1c70636647").
		WithExec([]string{"go", "install", "github.com/asdf-vm/asdf/cmd/asdf@v0.20.0"}).
		File("/go/bin/asdf")
}

func withToolVersions(c *dagger.Container, toolVersions *dagger.File) *dagger.Container {
	return c.WithFile(".tool-versions", toolVersions)
}

func withAsdfPlugins(c *dagger.Container) *dagger.Container {
	return c.WithExec([]string{"asdf", "plugin", "add", "golang"}).
		WithExec([]string{"asdf", "plugin", "add", "nodejs"})
}

func withAsdfInstall(c *dagger.Container) *dagger.Container {
	return c.WithExec([]string{"asdf", "install"})
}

func withGoSetup(ctx context.Context, c *dagger.Container) (*dagger.Container, error) {
	envStr, err := c.WithExec([]string{"go", "env"}).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run go env: %w", err)
	}

	env := parseEnv(envStr)

	return c.
		// set up go caches
		WithMountedCache(env["GOMODCACHE"], dag.CacheVolume("go-mod-cache")).
		WithMountedCache(env["GOCACHE"], dag.CacheVolume("go-build-cache")).

		// install go tools
		WithExec([]string{"go", "install", "golang.org/x/tools/gopls@latest"}), nil
}

func withSource(c *dagger.Container, src *dagger.Directory) *dagger.Container {
	return c.WithDirectory(".", src).
		WithMountedCache("/work", dag.CacheVolume("my-work"+time.Now().String())).
		WithExec([]string{"cp", "-R", "/src/.", "/work"}).
		WithWorkdir("/work")
}

func changesFromSrc(c *dagger.Container, src *dagger.Directory) *dagger.Changeset {
	return c.WithExec([]string{"rm", "-rf", "/src"}).
		WithExec([]string{"cp", "-R", "/work", "/src"}).
		WithoutMount("/work").
		Directory("/src").
		Changes(src)
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
