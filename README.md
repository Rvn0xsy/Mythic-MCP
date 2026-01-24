# Mythic MCP Server

A production-grade Model Context Protocol (MCP) server that wraps the [Mythic C2 Framework](https://github.com/its-a-feature/Mythic), enabling AI assistants to interact with Mythic for red team operations.

**Status:** Design Complete - Implementation In Progress

---

## Overview

This project creates an MCP server that exposes all 204+ Mythic C2 operations as MCP tools, allowing AI assistants like Claude to:

- Authenticate with Mythic servers
- Build and deploy payloads
- Manage callbacks and execute tasks
- Handle file operations
- Query MITRE ATT&CK mappings
- Generate reports and analytics
- Automate red team workflows

### Key Features

- **Complete Coverage**: All 204 Mythic SDK methods wrapped as MCP tools
- **Type-Safe**: Leverages Go's strong typing throughout
- **Production-Ready**: >90% test coverage, comprehensive error handling
- **CI-First**: All functionality validated through integration tests against real Mythic
- **Well-Documented**: Comprehensive documentation and examples

---

## Architecture

```
AI Assistant (Claude)
        ↓
   MCP Protocol
        ↓
   MCP Server (this project)
        ↓
   Mythic Go SDK
        ↓
Mythic C2 Framework
```

The MCP server acts as a thin wrapper around the [Mythic Go SDK](https://github.com/nbaertsch/mythic-sdk-go), translating MCP tool calls into Mythic SDK operations and formatting responses for AI consumption.

---

## Documentation

Comprehensive design documents are available in the `docs/` directory:

1. **[Architecture](docs/01-ARCHITECTURE.md)** - System design and component architecture
2. **[API Mapping](docs/02-API-MAPPING.md)** - Complete mapping of 204 Mythic SDK methods to MCP tools
3. **[Test Strategy](docs/03-TEST-STRATEGY.md)** - CI-First testing approach with >90% coverage
4. **[CI/CD Design](docs/04-CI-CD-DESIGN.md)** - GitHub Actions pipeline for automated testing
5. **[Implementation Roadmap](docs/05-IMPLEMENTATION-ROADMAP.md)** - Phased implementation plan

---

## Project Status

### Design Phase ✅ Complete

- [x] Architecture design
- [x] API mapping (all 204 tools)
- [x] Test strategy
- [x] CI/CD pipeline design
- [x] Implementation roadmap

### Implementation Phase 📋 Planned

**Timeline:** 6-8 weeks

- [ ] **Phase 0: Foundation** (Week 1)
  - Repository setup
  - CI/CD pipeline
  - Test infrastructure

- [x] **Phase 1: Authentication** (Week 2) ✅
  - 7 authentication tools
  - E2E tests ready
  - Coverage: 3.4% (7/204 tools)

- [ ] **Phase 2: Core Operations** (Weeks 3-4)
  - 48 core tools (Operations, Files, Operators, Tags, Credentials)
  - Cumulative coverage: 27%

- [ ] **Phase 3: Agent Operations** (Weeks 5-6)
  - 60 agent tools (Payloads, Callbacks, Tasks, C2 Profiles)
  - Cumulative coverage: 56%

- [ ] **Phase 4: Advanced Features** (Week 7)
  - 40 advanced tools (MITRE, Processes, Hosts, Screenshots)
  - Cumulative coverage: 76%

- [ ] **Phase 5: Specialized** (Week 8)
  - 43 specialized tools (Eventing, Containers, Alerts, Reporting)
  - Cumulative coverage: 97%

- [ ] **Phase 6: Polish & Release** (Week 8)
  - Final 6 tools
  - Documentation completion
  - v1.0.0 release
  - Coverage: 100%

---

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Mythic C2 Server (3.3.0+)
- Docker (for testing)

### Installation

```bash
# Clone repository
git clone https://github.com/YOUR_ORG/mythic-mcp.git
cd mythic-mcp

# Install dependencies
go mod download

# Build server
go build -o mythic-mcp ./cmd/mythic-mcp
```

### Configuration

Set environment variables:

```bash
export MYTHIC_URL="https://mythic.example.com:7443"
export MYTHIC_API_TOKEN="your-api-token"
# OR
export MYTHIC_USERNAME="mythic_admin"
export MYTHIC_PASSWORD="your-password"

# Optional
export MYTHIC_SKIP_TLS_VERIFY="false"
export LOG_LEVEL="info"
```

### Running the Server

```bash
./mythic-mcp
```

The MCP server will:
1. Connect to the configured Mythic instance
2. Authenticate using provided credentials
3. Register all 204 MCP tools
4. Listen for MCP protocol requests

### Usage with Claude Desktop

Add to Claude Desktop configuration (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "mythic": {
      "command": "/path/to/mythic-mcp",
      "env": {
        "MYTHIC_URL": "https://mythic.example.com:7443",
        "MYTHIC_API_TOKEN": "your-api-token"
      }
    }
  }
}
```

---

## Development

### Development Principles

This project follows **CI-First Development Philosophy**:

1. **Reason Before Coding** - Understand the system before implementing
2. **Integration Over Isolation** - Test against real Mythic, not mocks
3. **CI as Source of Truth** - If it passes in CI, it works
4. **No Test Skips** - Every test creates its own dependencies

See [CI-First Philosophy](.claude/skills/ci-first-philosophy/SKILL.md) for details.

### Setting Up Development Environment

```bash
# Clone repository
git clone https://github.com/YOUR_ORG/mythic-mcp.git
cd mythic-mcp

# Install development dependencies
go mod download
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Set up Mythic for testing (Docker required)
docker-compose -f docker-compose.test.yml up -d

# Run tests
make test
```

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run E2E tests (requires Mythic)
make test-e2e

# Generate coverage report
make coverage
```

### Code Quality

```bash
# Run linter
make lint

# Format code
make fmt

# Run all checks (lint + test + coverage)
make check
```

---

## MCP Tools Reference

### Tool Categories

The MCP server exposes 204 tools organized into categories:

**Core Operations**
- **Authentication** (7 tools) - Login, logout, token management
- **Operations** (11 tools) - Operation CRUD, settings, event logging
- **Files** (10 tools) - Upload, download, bulk operations
- **Operators** (12 tools) - User management, preferences
- **Tags** (9 tools) - Tag types and categorization

**Agent Operations**
- **Payloads** (14 tools) - Build, download, manage payloads
- **Callbacks** (14 tools) - Callback management, P2P
- **Tasks** (20 tools) - Task execution and responses
- **C2 Profiles** (9 tools) - Profile management, IOCs
- **Commands** (4 tools) - Command management

**Advanced Features**
- **MITRE ATT&CK** (7 tools) - Technique queries, mappings
- **Processes** (6 tools) - Process enumeration, tree
- **Hosts** (6 tools) - Host discovery, network mapping
- **Screenshots** (6 tools) - Screenshot capture, timeline
- **Keylogs** (3 tools) - Keylog retrieval
- **Tokens** (3 tools) - Token enumeration
- **File Browser** (3 tools) - Browse agent filesystems
- **RPFWD/Proxy** (6 tools) - Port forwarding, proxies

**Specialized**
- **Eventing** (14 tools) - Workflow automation, webhooks
- **Containers** (4 tools) - Container file operations
- **Alerts** (6 tools) - Alert management
- **Reporting** (3 tools) - Report generation
- **Build Parameters** (6 tools) - Parameter queries
- **Utilities** (8 tools) - Misc utilities

See [API Mapping](docs/02-API-MAPPING.md) for complete tool reference.

---

## Testing

### Test Strategy

This project follows a comprehensive testing strategy:

**Test Pyramid (Inverted for CI-First):**
```
┌─────────────────────────┐
│  E2E Workflow Tests     │  80% - Test against real Mythic
├─────────────────────────┤
│  Integration Tests      │  15% - Test MCP server integration
├─────────────────────────┤
│  Unit Tests             │  5% - Test formatters, validators
└─────────────────────────┘
```

### Coverage Requirements

- **Tool Coverage**: 204/204 tools (100%)
- **Code Coverage**: >90%
- **Skip Rate**: 0% (no skipped tests)
- **Pass Rate**: 100% in CI

### Test Execution

```bash
# Full test suite (including E2E with Mythic)
make test-all

# Quick tests (unit + integration, no Mythic)
make test-quick

# E2E tests only
make test-e2e

# Coverage report
make coverage-report
```

See [Test Strategy](docs/03-TEST-STRATEGY.md) for details.

---

## CI/CD Pipeline

### GitHub Actions Workflows

**Test Pipeline** (`.github/workflows/test.yml`)
- Runs on every PR and push to main
- Stages: Lint → Unit Tests → Integration Tests → E2E Tests
- Duration: ~13 minutes
- Requires: Docker, Mythic Framework, Poseidon agent

**Release Pipeline** (`.github/workflows/release.yml`)
- Triggered by version tags (e.g., `v1.0.0`)
- Builds binaries for Linux, macOS, Windows (amd64, arm64)
- Creates GitHub release with artifacts

### CI Status

[![Test](https://github.com/YOUR_ORG/mythic-mcp/workflows/Test/badge.svg)](https://github.com/YOUR_ORG/mythic-mcp/actions)
[![Coverage](https://codecov.io/gh/YOUR_ORG/mythic-mcp/branch/main/graph/badge.svg)](https://codecov.io/gh/YOUR_ORG/mythic-mcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/YOUR_ORG/mythic-mcp)](https://goreportcard.com/report/github.com/YOUR_ORG/mythic-mcp)

See [CI/CD Design](docs/04-CI-CD-DESIGN.md) for pipeline details.

---

## Contributing

We welcome contributions! Please follow these guidelines:

1. **Read the Documentation**
   - Familiarize yourself with the [Architecture](docs/01-ARCHITECTURE.md)
   - Understand the [CI-First Philosophy](.claude/skills/ci-first-philosophy/SKILL.md)

2. **Follow the Process**
   - Write tests first (TDD)
   - Ensure CI passes before submitting PR
   - Maintain >90% coverage
   - No skipped tests allowed

3. **Code Standards**
   - Follow Go best practices
   - Run `make lint` before committing
   - Write clear commit messages
   - Update documentation

4. **Pull Request Guidelines**
   - Small, focused PRs
   - Link to related issues
   - Include test coverage
   - Pass all CI checks

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Mythic C2 Framework](https://github.com/its-a-feature/Mythic) by @its-a-feature
- [Mythic Go SDK](https://github.com/nbaertsch/mythic-sdk-go) (upstream dependency)
- [Model Context Protocol](https://modelcontextprotocol.io) by Anthropic
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)

---

## Related Projects

- [Mythic C2 Framework](https://github.com/its-a-feature/Mythic)
- [Mythic Go SDK](https://github.com/nbaertsch/mythic-sdk-go)
- [Model Context Protocol](https://modelcontextprotocol.io)
- [MCP Servers](https://github.com/modelcontextprotocol/servers)

---

## Support

- **Issues**: [GitHub Issues](https://github.com/YOUR_ORG/mythic-mcp/issues)
- **Discussions**: [GitHub Discussions](https://github.com/YOUR_ORG/mythic-mcp/discussions)
- **Mythic Community**: [Mythic Slack](https://bloodhoundgang.herokuapp.com/)

---

**Built with ❤️ for the red team community**
**Following CI-First Development Philosophy**
