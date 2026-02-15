# Installation

## Build from Source

```bash
git clone https://github.com/nbaertsch/Mythic-MCP.git
cd Mythic-MCP
go build -o mythic-mcp ./cmd/mythic-mcp
```

## Docker

Build and run alongside your Mythic stack:

```bash
# Build the image
docker build -t mythic-mcp:latest .

# Run on the Mythic Docker network
docker run -d \
  --name mythic-mcp \
  --restart unless-stopped \
  --network mythic_default \
  -p 3333:3333 \
  -v /path/to/mythic/documentation-docker/content:/docs:ro \
  -e MYTHIC_URL=https://mythic_nginx:7443 \
  -e MYTHIC_API_TOKEN=your-api-token \
  -e MCP_TRANSPORT=http \
  -e MCP_HTTP_PORT=3333 \
  -e MYTHIC_SKIP_TLS_VERIFY=true \
  -e MYTHIC_DOCS_PATH=/docs \
  mythic-mcp:latest
```

## Verify

```bash
# stdio mode — list tools
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./mythic-mcp

# HTTP mode — health check
curl http://localhost:3333/sse
```
