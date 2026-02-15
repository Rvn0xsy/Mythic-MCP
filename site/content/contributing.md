# Contributing

Contributions are welcome! Here's how to get started.

## Development Setup

```bash
git clone https://github.com/nbaertsch/Mythic-MCP.git
cd Mythic-MCP
go mod download
```

## Code Standards

- **Go 1.23** — use modern idioms, no legacy patterns
- Run `golangci-lint run` before committing
- All tool files follow the pattern in `pkg/server/tools_*.go`
- Tabs for indentation (Go standard)

## Adding a New Tool

1. Pick the appropriate `tools_<category>.go` file (or create one)
2. Add the `mcp.AddTool` call with a clear, AI-readable description
3. Define the args struct with `json` and `jsonschema` tags
4. Implement the handler, calling the SDK and returning `*mcp.CallToolResult`
5. Write an E2E test in `tests/`
6. The CI workflow will regenerate the documentation site automatically

## Pull Request Guidelines

- Small, focused PRs
- Link to related issues
- Pass all CI checks (`make lint && make test`)
- Don't skip tests — every test creates its own dependencies

## Documentation Site

The tool reference is **auto-generated** from source code. Do not edit
`site/content/tools/*.md` by hand — they will be overwritten by the schema
generator on each build.

To preview locally:

```bash
pip install -r requirements-docs.txt
go run ./tools/gen-schema-docs
mkdocs serve
```
