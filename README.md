# Container Agent

Run agents in a containerised sandbox that only embeds your project.
Rest assured that the agents can only access the tools and files it needs:

Use with an LLM gateway:
```sh
dagger call -m https://github.com/chrisjpalmer/container-agent claude --anthropic-base-url=https://blahblah.com --anthropic-auth-token=env://ANTHROPIC_AUTH_TOKEN agent
```

Or a local llamacpp server:
```sh
dagger call -m https://github.com/chrisjpalmer/container-agent claude --llamacpp=tcp://localhost:8080 agent
dagger call -m https://github.com/chrisjpalmer/container-agent opencode --llamacpp=tcp://localhost:8080 agent
dagger call -m https://github.com/chrisjpalmer/container-agent pi --llamacpp=tcp://localhost:8080 agent
```

## Prerequisites

- dagger 0.21.7