# Architecture

## System Overview

Mythic MCP Server is a **bridge** between the
[Model Context Protocol](https://modelcontextprotocol.io) and the
[Mythic C2 Framework](https://github.com/its-a-feature/Mythic).

```mermaid
flowchart TB
    subgraph clients ["MCP Clients"]
        claude["Claude Desktop"]
        chat["ChatGPT / other"]
        custom["Custom MCP Client"]
    end

    subgraph mcp_server ["Mythic MCP Server (Go)"]
        transport["Transport Layer<br/>stdio · HTTP/SSE"]
        router["Tool Router<br/>148 tools, 19 categories"]
        sdk["Mythic Go SDK"]
    end

    subgraph mythic ["Mythic C2 Framework"]
        api["GraphQL API · REST"]
        agents["Payload Types<br/>Xenon · Poseidon · Forge"]
        c2["C2 Profiles<br/>httpx · tcp · http"]
        data["Operational Data<br/>callbacks · tasks · files"]
    end

    clients -->|"MCP (JSON-RPC)"| transport
    transport --> router
    router --> sdk
    sdk -->|"HTTPS + GraphQL"| api
    api --> agents & c2 & data

    style clients fill:#7c3aed,color:#fff,stroke:none
    style mcp_server fill:#4f46e5,color:#fff,stroke:none
    style mythic fill:#f59e0b,color:#000,stroke:none
```

## Transport Modes

The server supports two MCP transport modes:

| Mode | Use Case | Config |
|------|----------|--------|
| **stdio** | Claude Desktop, local CLI tools | Default — no extra config |
| **HTTP/SSE** | Remote clients, Docker containers, shared access | `MCP_TRANSPORT=http MCP_HTTP_PORT=3333` |

## Tool Registration

At startup the server registers every tool with the MCP SDK. Each tool
declaration includes:

- **Name** — a stable identifier like `mythic_issue_task`
- **Description** — natural-language explanation the AI reads to decide when to call it
- **Input Schema** — JSON Schema for parameters (auto-derived from Go structs via `jsonschema` tags)

The tool reference pages in this site are **generated directly from the source
code** to stay in sync automatically.

## Authentication Flow

```mermaid
sequenceDiagram
    participant AI as AI Assistant
    participant MCP as MCP Server
    participant Mythic as Mythic API

    alt API Token (recommended)
        AI->>MCP: Any tool call
        MCP->>Mythic: Header: apitoken / Bearer
        Mythic-->>MCP: Response
    else Username / Password
        AI->>MCP: mythic_login(user, pass)
        MCP->>Mythic: POST /auth
        Mythic-->>MCP: JWT access + refresh
        MCP-->>AI: "Authenticated as user"
        AI->>MCP: Any tool call
        MCP->>Mythic: Header: Bearer JWT
    end
```

The SDK automatically detects whether `MYTHIC_API_TOKEN` contains a real
API token or a JWT, and picks the correct auth header format.

## Project Layout

```
Mythic-MCP/
├── cmd/mythic-mcp/       # Entrypoint
├── pkg/
│   ├── config/           # Env-based configuration
│   └── server/
│       ├── server.go     # MCP server lifecycle
│       ├── tools_*.go    # Tool registration + handlers (18 files)
│       └── errors.go     # SDK → MCP error translation
├── tools/
│   └── gen-schema-docs/  # Schema → Markdown generator
├── site/                 # This documentation site
├── tests/                # E2E integration tests
└── mkdocs.yml            # Site configuration
```
