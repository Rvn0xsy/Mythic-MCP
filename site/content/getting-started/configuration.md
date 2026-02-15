# Configuration

All configuration is via **environment variables**. No config files needed.

## Required

| Variable | Description | Example |
|----------|-------------|---------|
| `MYTHIC_URL` | Full URL to your Mythic instance | `https://mythic.lab:7443` |
| `MYTHIC_API_TOKEN` | API token **or** JWT | `abc123…` |

!!! tip "API Token vs. JWT"
    If you pass a JWT (starts with `eyJ`), the SDK auto-detects it and uses
    `Authorization: Bearer` headers instead of `apitoken` headers.

## Optional — Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `MYTHIC_USERNAME` | — | Username for login-based auth |
| `MYTHIC_PASSWORD` | — | Password for login-based auth |
| `MYTHIC_SKIP_TLS_VERIFY` | `false` | Skip TLS certificate verification |

## Optional — Transport

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_TRANSPORT` | `stdio` | `stdio` or `http` |
| `MCP_HTTP_PORT` | `3333` | Port for HTTP/SSE mode |

## Optional — Documentation

| Variable | Default | Description |
|----------|---------|-------------|
| `MYTHIC_DOCS_PATH` | `/root/mythic/documentation-docker/content` | Path to Mythic docs content directory |
| `LOG_LEVEL` | `info` | Logging verbosity (`debug`, `info`, `warn`, `error`) |
