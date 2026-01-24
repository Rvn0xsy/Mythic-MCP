# CI-First Development Plugin

A Claude Code plugin that guides development using GitHub Actions CI/CD-first methodology with comprehensive integration and end-to-end testing instead of isolated unit tests.

## Philosophy

**CI-First Development** shifts from traditional unit-test-driven TDD to a proven approach where:

1. **Reason Before Coding** - Understand system behavior, integration points, and desired outcomes before implementation
2. **Integration Over Isolation** - Prioritize integration and E2E tests that validate real system behavior
3. **CI as Source of Truth** - All validation happens through GitHub Actions with real services
4. **Fast Feedback Loops** - Rapid iteration with comprehensive CI coverage catches real-world issues early

### Why CI-First?

Traditional unit-test-driven TDD often misses real-world issues:
- Database constraint violations
- API contract mismatches
- Service interaction bugs
- Configuration errors
- Race conditions
- Deployment issues

**CI-First testing with real services catches these before production.**

## Features

This plugin provides:

- **CI-First Philosophy Skill** - Guides Claude's overall development approach with systematic thinking and real service integration
- **Integration Test Planning Skill** - Designs comprehensive test strategies with Docker orchestration and test data management
- **GitHub Actions Setup Skill** - Implements optimized CI/CD pipelines with caching, parallelization, and deployment automation
- **CI Architect Agent** - Specialized assistant for complex CI/CD architecture and troubleshooting

## Installation

### From Marketplace

```bash
/plugin marketplace add claude-lab
/plugin install ci-first-dev@claude-lab
```

### Local Installation (Development)

```bash
/claude-dev:install-local ci-first-dev@claude-lab
```

## Components

### Skills (Auto-Discovered)

#### 1. CI-First Philosophy

**Triggers:** "implement", "build", "create", "add functionality", "develop"

Guides Claude to:
- Analyze system requirements before coding
- Design test data and helper functions
- Use EnsureXExists patterns for test prerequisites
- Organize tests by dependency phases
- Treat test skips as failures
- Share expensive resources (connections, payloads)
- Use rapid CI feedback loops

**Key Patterns:**
- `EnsureXExists(t, client)` - Never skip tests due to missing data
- Phase-based test organization (0: Schema, 1-2: Core, 3: Integration, 4: Advanced, 5: Edge cases)
- Small focused commits with clear messages
- Pattern matching from existing code

#### 2. Integration Test Planning

**Triggers:** "test", "testing", "coverage", "validation", "integration"

Guides Claude to:
- Set up Docker Compose test environments
- Design integration tests with real databases
- Implement E2E tests for complete workflows
- Manage test data with shared/per-test strategies
- Handle database migrations and seeding
- Structure tests by dependency levels

**Key Patterns:**
- Real services via Docker (PostgreSQL, Redis, not mocks)
- Shared expensive resources, per-test cheap resources
- Test data helpers that auto-create prerequisites
- Integration tests validate service boundaries
- E2E tests validate complete user journeys

#### 3. GitHub Actions Setup

**Triggers:** "CI", "pipeline", "GitHub Actions", "workflow", "deployment"

Guides Claude to:
- Design fast-feedback CI pipelines
- Implement caching strategies
- Configure matrix testing
- Set up real service dependencies
- Optimize parallel execution
- Handle secrets and environments
- Debug CI failures systematically

**Key Patterns:**
- Fast checks first (lint, type-check)
- Services via GitHub Actions services
- Cache dependencies and build artifacts
- Parallel independent jobs
- Environment protection for deployments

### Agent

#### CI Architect

**Use when:** Planning CI/CD strategy, implementing workflows, debugging CI issues, designing test infrastructure

A specialized agent with expertise in:
- GitHub Actions workflow design
- Integration testing with real services
- End-to-end test orchestration
- Test strategy and coverage
- Docker-based test environments
- CI/CD optimization

**Invocation:**
```
Can you help design a CI pipeline for my Node.js API with PostgreSQL?
# Claude will invoke the ci-architect agent
```

## Usage Examples

### Example 1: Implementing a New Feature

```
User: "Add user authentication to the API"

Claude (guided by ci-first-philosophy):
1. Analyzes system requirements:
   - What auth method? (JWT, sessions)
   - Database tables needed (users, sessions)
   - Integration points (login endpoint, middleware)

2. Plans test strategy:
   - Integration tests: Auth endpoints + DB
   - E2E tests: Full login/logout flow
   - Test data: EnsureUserExists helper

3. Implements tests first (with real DB via Docker)
4. Implements feature
5. Validates in CI with GitHub Actions
```

### Example 2: Setting Up CI

```
User: "Set up CI for my project"

Claude (uses github-actions-setup skill):
1. Creates docker-compose.test.yml with real services
2. Implements .github/workflows/ci.yml with:
   - Fast checks (lint, type-check)
   - Integration tests with PostgreSQL service
   - E2E tests with full docker-compose
   - Caching for dependencies
   - Parallel job execution
3. Documents how to run tests locally
```

### Example 3: Debugging Test Failures

```
User: "Tests are failing in CI"

Claude (uses ci-first-philosophy):
1. Gets CI logs: gh run view <run_id> --log
2. Categorizes error:
   - "field not found" → Schema issue
   - "expected X, got Y" → Test assertion issue
   - "timeout" → Environment issue
3. Pattern matches from working code
4. Applies fix and validates in CI
```

## Core Principles

### The 10 Commandments

1. **Thou shalt reason before coding**
2. **Thou shalt treat test skips as failures**
3. **Thou shalt use CI as thy truth**
4. **Thou shalt test against real services**
5. **Thou shalt fix root causes, not symptoms**
6. **Thou shalt make small, focused commits**
7. **Thou shalt document the "why" not just the "what"**
8. **Thou shalt share expensive resources**
9. **Thou shalt not over-engineer**
10. **Thou shalt verify every change with CI**

### Test Pyramid - Inverted for CI-First

```
┌─────────────────────────────┐
│    End-to-End Tests         │  ← MOST IMPORTANT
│    (Full system flows)      │     Real user journeys
├─────────────────────────────┤
│   Integration Tests         │  ← CORE VALIDATION
│   (Service interactions)    │     Real DB, APIs
├─────────────────────────────┤
│      Unit Tests             │  ← TARGETED ONLY
│   (Complex logic only)      │     Not exhaustive
└─────────────────────────────┘
```

### Key Patterns

#### EnsureXExists Pattern

```go
func EnsureUserExists(t *testing.T, client *Client) *User {
    if sharedUser != nil {
        return sharedUser
    }

    user, err := client.CreateUser(ctx, &UserRequest{...})
    require.NoError(t, err)
    sharedUser = user

    t.Cleanup(func() {
        // Clean up local files only, not shared resources
    })

    return user
}
```

#### Phase-Based Test Organization

```
Phase 0: Schema validation
Phase 1-2: Core APIs (auth, basic operations)
Phase 3: Integration (service interactions)
Phase 4: Advanced features
Phase 5: Edge cases (errors, timeouts)
```

#### Rapid CI Feedback Loop

```bash
1. Make targeted change
2. Commit with clear message
3. Push and watch CI
4. If fails → categorize error → fix → repeat
5. If passes → continue
```

## Real-World Success

This methodology was proven effective in building an enterprise SDK with:
- Full integration test coverage with real Mythic instances
- Zero skipping tests (all prerequisites auto-created)
- Phase-based test organization
- Rapid CI feedback loops (small commits, frequent pushes)
- Pattern matching for schema compatibility
- Systematic error categorization and fixing

## Requirements

- Claude Code CLI
- Docker and docker-compose (for local testing)
- GitHub account (for Actions)
- Git

## Support

For issues or questions about this plugin:
- GitHub: https://github.com/noahbaertsch/ci-first-dev
- Email: noahbaertsch@github.com

For Claude Code itself:
- Documentation: https://docs.anthropic.com/claude-code
- Issues: https://github.com/anthropics/claude-code/issues

## License

MIT License - see LICENSE file for details

---

**Remember:** The goal is not 100% unit test coverage. The goal is comprehensive integration and E2E test coverage that validates the system actually works in production-like conditions.
