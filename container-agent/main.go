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
		Src:          ws.Directory("/"),
	}
}

func base(ctx context.Context) (*dagger.Container, error) {
	ctr := dag.Container().
		From("ubuntu:latest@sha256:3131b4cc82a783df6c9df078f86e01819a13594b865c2cad47bd1bca2b7063bb").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "build-essential", "curl", "git", "wget", "-y"})

	ctr = withGithubCLI(ctr)

	ctr, err := withTools(ctx, ctr)
	if err != nil {
		return nil, fmt.Errorf("error while adding tools: %w", err)
	}

	return ctr.WithFile("/usr/bin/asdf", asdf()).
		WithEnvVariable("PATH", "/root/.asdf/shims:${PATH}", dagger.ContainerWithEnvVariableOpts{Expand: true}).
		WithEnvVariable("PATH", "/root/.local/bin:${PATH}", dagger.ContainerWithEnvVariableOpts{Expand: true}).
		WithWorkdir("/src"), nil
}

func withGitConfig(ctr *dagger.Container, name, email string) *dagger.Container {
	return ctr.WithExec([]string{"git", "config", "--global", "user.name", name}).
		WithExec([]string{"git", "config", "--global", "user.email", email}).
		WithExec([]string{"git", "config", "--global", "push.autoSetupRemote", "true"})
}

func withGithubSSH(ctr *dagger.Container, ssh *dagger.Socket, goPrivate, sshRewrite string) (*dagger.Container, error) {
	const ghHost = "github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl"

	ctr = ctr.WithExec([]string{"mkdir", "-p", "/root/.ssh"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("echo %s >> /root/.ssh/known_hosts", ghHost)})

	ctr, err := withSSHRewrite(ctr, sshRewrite)
	if err != nil {
		return nil, err
	}

	return ctr.WithExec([]string{"go", "env", "-w", "GOPRIVATE=" + goPrivate}).
		WithUnixSocket("/sock/ssh", ssh).
		WithEnvVariable("SSH_AUTH_SOCK", "/sock/ssh"), nil
}

// github.com/chrisjpalmer
func withSSHRewrite(ctr *dagger.Container, sshRewrite string) (*dagger.Container, error) {
	const ghPrefix = "github.com/"
	if !strings.HasPrefix(sshRewrite, ghPrefix) {
		return nil, fmt.Errorf("expected sshRewrite to start with %q", ghPrefix)
	}

	org := strings.TrimPrefix(sshRewrite, ghPrefix)

	return ctr.WithExec([]string{
		"git",
		"config",
		"--global",
		fmt.Sprintf("url.git@github.com:%s/.insteadOf", org),
		fmt.Sprintf("https://github.com/%s/", org),
	}), nil
}

func withGithubCLI(c *dagger.Container) *dagger.Container {
	return c.WithExec([]string{"sh", "-c", `mkdir -p -m 755 /etc/apt/keyrings \
&& out=$(mktemp) && wget -nv -O$out https://cli.github.com/packages/githubcli-archive-keyring.gpg \
&& cat $out | tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null \
&& chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
&& mkdir -p -m 755 /etc/apt/sources.list.d \
&& echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
&& apt update \
&& apt install gh -y`})

}

func withAwsCredentials(ctr *dagger.Container, awsProfile string, awsCredentials *dagger.Secret) *dagger.Container {
	return ctr.WithEnvVariable("AWS_PROFILE", awsProfile).
		WithMountedSecret("/root/.aws/credentials", awsCredentials)
}

func withTools(ctx context.Context, c *dagger.Container) (*dagger.Container, error) {
	tools, err := buildTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("error while trying to build tools: %w", err)
	}

	return c.WithFiles("/usr/bin", tools), nil
}

func buildTools(ctx context.Context) ([]*dagger.File, error) {
	tools, err := dag.CurrentModule().Source().Directory("tools").Entries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list available tools: %w", err)
	}

	out := make([]*dagger.File, 0, len(tools))
	for _, t := range tools {
		out = append(out, buildTool(t))
	}

	return out, nil
}

func buildTool(tool string) *dagger.File {
	src := dag.CurrentModule().Source().Directory("tools").Directory(tool)

	return golang().
		WithWorkdir("/app").
		WithDirectory(".", src).
		WithExec([]string{"go", "build", "-o", "tool", "."}).
		File("tool").
		WithName(strings.Trim(tool, "/"))
}

func golang() *dagger.Container {
	return dag.Container().
		From("golang:latest@sha256:3aff6657219a4d9c14e27fb1d8976c49c29fddb70ba835014f477e1c70636647")
}

func asdf() *dagger.File {
	return golang().
		WithExec([]string{"go", "install", "github.com/asdf-vm/asdf/cmd/asdf@v0.20.0"}).
		File("/go/bin/asdf")
}

func withToolVersions(c *dagger.Container, toolVersions *dagger.File) *dagger.Container {
	return c.WithFile(".tool-versions", toolVersions)
}

func withAsdfPlugins(c *dagger.Container) *dagger.Container {
	return c.WithExec([]string{"asdf", "plugin", "add", "golang"}).
		WithExec([]string{"asdf", "plugin", "add", "nodejs"}).
		WithExec([]string{"asdf", "plugin", "add", "dagger"})
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
