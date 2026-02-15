# Claude Desktop Integration

Add the following to your Claude Desktop configuration file.

## Find the config file

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |

## stdio mode (local binary)

```json
{
  "mcpServers": {
    "mythic": {
      "command": "/path/to/mythic-mcp",
      "env": {
        "MYTHIC_URL": "https://mythic.lab:7443",
        "MYTHIC_API_TOKEN": "your-api-token",
        "MYTHIC_SKIP_TLS_VERIFY": "true"
      }
    }
  }
}
```

## HTTP/SSE mode (remote server)

If the MCP server is running remotely (e.g. in Docker):

```json
{
  "mcpServers": {
    "mythic": {
      "url": "http://mythic-mcp-host:3333/sse"
    }
  }
}
```

## Verify

After restarting Claude Desktop, open a new conversation and ask:

> *"List the available Mythic tools"*

Claude should respond with the full tool catalog — 148 tools across 19
categories.
