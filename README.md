<p align="center">
  <img src="site/assets/mythic-mcp-logo.svg" width="120" alt="Mythic MCP Logo"/>
</p>

<h1 align="center">Mythic MCP Server</h1>

<p align="center">
  An MCP server that gives AI assistants full operational control of the
  <a href="https://github.com/its-a-feature/Mythic">Mythic C2 Framework</a>.
</p>

<p align="center">
  <a href="https://github.com/nbaertsch/Mythic-MCP/actions/workflows/test.yml"><img src="https://github.com/nbaertsch/Mythic-MCP/actions/workflows/test.yml/badge.svg" alt="Tests"></a>
  <a href="https://github.com/nbaertsch/Mythic-MCP/actions/workflows/deploy-docs.yml"><img src="https://github.com/nbaertsch/Mythic-MCP/actions/workflows/deploy-docs.yml/badge.svg" alt="Docs"></a>
  <a href="https://goreportcard.com/report/github.com/nbaertsch/Mythic-MCP"><img src="https://goreportcard.com/badge/github.com/nbaertsch/Mythic-MCP" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/nbaertsch/Mythic-MCP" alt="License"></a>
</p>

<p align="center">
  <a href="https://nbaertsch.github.io/Mythic-MCP/"><strong>Documentation</strong></a> ·
  <a href="https://nbaertsch.github.io/Mythic-MCP/tools/"><strong>Tool Reference</strong></a> ·
  <a href="https://nbaertsch.github.io/Mythic-MCP/showcase/lab-walkthrough/"><strong>Lab Walkthrough</strong></a>
</p>

---

**147 tools · 18 categories · Go · Mythic ≥ 3.3.0**

Mythic MCP Server exposes every meaningful operation in the [Mythic C2 Framework](https://github.com/its-a-feature/Mythic) as a structured [Model Context Protocol](https://modelcontextprotocol.io) tool. Connect it to Claude, ChatGPT, or any MCP client and operate Mythic through natural language.

```
AI Assistant  ──MCP──▶  Mythic MCP Server  ──SDK──▶  Mythic C2
```

### Capabilities

- **Auth** — login, API tokens, session management
- **Payloads** — build, download, inspect agents (Xenon, Poseidon, Forge, …)
- **Callbacks** — list, task, pivot, manage active sessions
- **C2 Profiles** — start/stop listeners, IOCs, configuration
- **Tasks** — issue commands, read output, wait for completion
- **Files** — upload, download, preview, bulk export
- **MITRE ATT&CK** — technique lookup, coverage mapping
- **Operations** — campaign management, event logs, global settings
- **+ 10 more categories** — credentials, artifacts, hosts, processes, screenshots, keylogs, tags, operators, commands, documentation

Full schema for every tool: **[Tool Reference →](https://nbaertsch.github.io/Mythic-MCP/tools/)**

---

## Quick Start

```bash
# Build
go build -o mythic-mcp ./cmd/mythic-mcp

# Run (stdio — Claude Desktop)
MYTHIC_URL=https://mythic.lab:7443 \
MYTHIC_API_TOKEN=your-token \
  ./mythic-mcp

# Run (HTTP/SSE — Docker / remote)
MCP_TRANSPORT=http MCP_HTTP_PORT=3333 \
MYTHIC_URL=https://mythic.lab:7443 \
MYTHIC_API_TOKEN=your-token \
  ./mythic-mcp
```

### Docker

```bash
docker build -t mythic-mcp:latest .
docker run -d --name mythic-mcp --network mythic_default \
  -p 3333:3333 \
  -e MYTHIC_URL=https://mythic_nginx:7443 \
  -e MYTHIC_API_TOKEN=your-token \
  -e MCP_TRANSPORT=http -e MCP_HTTP_PORT=3333 \
  -e MYTHIC_SKIP_TLS_VERIFY=true \
  mythic-mcp:latest
```

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "mythic": {
      "command": "/path/to/mythic-mcp",
      "env": {
        "MYTHIC_URL": "https://mythic.lab:7443",
        "MYTHIC_API_TOKEN": "your-token"
      }
    }
  }
}
```

See the full [Getting Started guide](https://nbaertsch.github.io/Mythic-MCP/getting-started/) for all options.

---

## Configuration

| Variable | Required | Description |
|----------|:--------:|-------------|
| `MYTHIC_URL` | ✓ | Mythic server URL |
| `MYTHIC_API_TOKEN` | ✓¹ | API token or JWT |
| `MYTHIC_USERNAME` / `MYTHIC_PASSWORD` | ✓¹ | Alternative: credential-based auth |
| `MYTHIC_SKIP_TLS_VERIFY` | | Skip TLS verification (default `false`) |
| `MCP_TRANSPORT` | | `stdio` (default) or `http` |
| `MCP_HTTP_PORT` | | HTTP/SSE listen port (default `3333`) |
| `MYTHIC_DOCS_PATH` | | Path to Mythic docs content for `mythic_get_documentation` |

<sub>¹ One of API token or username/password is required.</sub>

---

## Architecture

```
┌──────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  MCP Client  │────▶│  Mythic MCP      │────▶│  Mythic C2      │
│  (Claude,    │ MCP │  Server           │ SDK │  Framework      │
│   ChatGPT)   │◀────│  147 tools        │◀────│  GraphQL + REST │
└──────────────┘     └──────────────────┘     └─────────────────┘
```

The server is a thin, type-safe translation layer. Each tool validates inputs against a JSON Schema derived from Go struct tags, calls the [Mythic Go SDK](https://github.com/nbaertsch/mythic-sdk-go), and returns structured results.

Full architecture docs: **[Architecture →](https://nbaertsch.github.io/Mythic-MCP/architecture/)**

---

## Documentation Site

The documentation at [nbaertsch.github.io/Mythic-MCP](https://nbaertsch.github.io/Mythic-MCP/) is auto-generated and deployed on every push to `main`. The [Tool Reference](https://nbaertsch.github.io/Mythic-MCP/tools/) pages are built by a Go program (`tools/gen-schema-docs`) that parses the source code directly — tool names, descriptions, and parameter schemas are always in sync.

```bash
# Preview docs locally
pip install -r requirements-docs.txt
go run ./tools/gen-schema-docs
mkdocs serve
```

---

## Contributing

```bash
git clone https://github.com/nbaertsch/Mythic-MCP.git
cd Mythic-MCP && go mod download
```

- Run `golangci-lint run` before committing
- Tool files follow the pattern in `pkg/server/tools_*.go`
- Tool reference pages are auto-generated — don't edit `site/content/tools/*.md` by hand

See [Contributing →](https://nbaertsch.github.io/Mythic-MCP/contributing/)

---

## Related Projects

| Project | Description |
|---------|-------------|
| [Mythic C2 Framework](https://github.com/its-a-feature/Mythic) | The C2 framework this server wraps |
| [Mythic Go SDK](https://github.com/nbaertsch/mythic-sdk-go) | Go SDK for the Mythic API (upstream dep) |
| [Model Context Protocol](https://modelcontextprotocol.io) | The protocol spec |
| [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) | Official Go MCP SDK |

## License

[MIT](LICENSE)
