# Mythic MCP Server CI/CD Design

**Author:** Claude Code
**Date:** 2026-01-24
**Purpose:** GitHub Actions CI/CD pipeline for automated testing and deployment
**Philosophy:** CI-First - If it passes in CI, it works

---

## Table of Contents

1. [CI/CD Overview](#cicd-overview)
2. [GitHub Actions Workflows](#github-actions-workflows)
3. [Test Pipeline](#test-pipeline)
4. [Release Pipeline](#release-pipeline)
5. [Caching Strategy](#caching-strategy)
6. [Security Considerations](#security-considerations)

---

## CI/CD Overview

### Pipeline Goals

1. **Automated Testing** - All tests run on every PR
2. **Fast Feedback** - Results in <15 minutes
3. **Reproducible** - Same results every time
4. **Comprehensive** - Unit, integration, and E2E tests
5. **Secure** - No credentials in code, secure secrets management

### Pipeline Architecture

```
┌─────────────────────────────────────────────────────┐
│              GitHub Pull Request                     │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│          Lint & Format Check                         │
│  - golangci-lint                                     │
│  - go fmt                                            │
│  Duration: ~30 seconds                               │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│          Unit Tests                                  │
│  - pkg/server/*_test.go                              │
│  - tests/unit/**/*_test.go                           │
│  Duration: ~10 seconds                               │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│          Integration Tests                           │
│  - tests/integration/*_test.go                       │
│  - Mock transport (no Mythic needed)                 │
│  Duration: ~30 seconds                               │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│          E2E Tests (with Mythic)                     │
│  1. Clone Mythic Framework                           │
│  2. Build mythic-cli                                 │
│  3. Start Mythic                                     │
│  4. Install Poseidon agent                           │
│  5. Run E2E test suite                               │
│  Duration: ~10-12 minutes                            │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│          Build & Release (on tag)                    │
│  - Build binaries (Linux, macOS, Windows)            │
│  - Create GitHub release                             │
│  - Upload artifacts                                  │
│  Duration: ~2 minutes                                │
└──────────────────────────────────────────────────────┘
```

---

## GitHub Actions Workflows

### Workflow 1: Test Pipeline

**File:** `.github/workflows/test.yml`

```yaml
name: Test

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]
  workflow_dispatch:

jobs:
  lint:
    name: Lint & Format
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

      - name: Check formatting
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "Go code is not formatted:"
            gofmt -d .
            exit 1
          fi

  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Run unit tests
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic \
            ./pkg/... \
            ./tests/unit/...

      - name: Check coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 90" | bc -l) )); then
            echo "Coverage is below 90%"
            exit 1
          fi

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
          flags: unit

  integration-tests:
    name: Integration Tests (No Mythic)
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: Download dependencies
        run: go mod download

      - name: Run integration tests
        run: |
          go test -v -race \
            -tags=integration \
            -run 'TestIntegration_' \
            ./tests/integration/...

  e2e-tests:
    name: E2E Tests (with Mythic)
    runs-on: ubuntu-latest
    needs: [unit-tests, integration-tests]
    timeout-minutes: 25
    steps:
      - name: Checkout MCP Server code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: Clone Mythic Framework
        run: |
          git clone --depth 1 https://github.com/its-a-feature/Mythic.git /tmp/mythic
          cd /tmp/mythic
          echo "Mythic version: $(cat VERSION)"

      - name: Build Mythic CLI
        run: |
          cd /tmp/mythic
          sudo make
          sudo chmod +x mythic-cli

      - name: Start Mythic
        run: |
          cd /tmp/mythic
          sudo ./mythic-cli start
          echo "Mythic started"

      - name: Install Poseidon Agent
        run: |
          cd /tmp/mythic
          echo "Installing Poseidon agent..."
          sudo ./mythic-cli install github https://github.com/MythicAgents/poseidon

          echo "Waiting for Poseidon container..."
          timeout=180
          attempt=0
          while [ $attempt -lt $timeout ]; do
            if sudo docker ps | grep -q poseidon; then
              echo "✓ Poseidon ready after $attempt seconds"
              break
            fi
            sleep 5
            attempt=$((attempt + 5))
          done

          if [ $attempt -ge $timeout ]; then
            echo "✗ Poseidon failed to start"
            sudo ./mythic-cli status
            sudo docker ps -a
            exit 1
          fi

      - name: Wait for Mythic Ready
        run: |
          echo "Waiting for Mythic to be ready..."
          timeout=240
          attempt=0
          while [ $attempt -lt $timeout ]; do
            if curl -k -s https://127.0.0.1:7443 > /dev/null 2>&1; then
              echo "✓ Mythic is ready after $attempt seconds"
              break
            fi
            sleep 5
            attempt=$((attempt + 5))
          done

          if [ $attempt -ge $timeout ]; then
            echo "✗ Mythic failed to start within timeout"
            cd /tmp/mythic
            sudo ./mythic-cli status
            exit 1
          fi

      - name: Extract Mythic Credentials
        id: mythic-creds
        run: |
          PASSWORD=$(grep MYTHIC_ADMIN_PASSWORD /tmp/mythic/.env | cut -d'=' -f2 | tr -d '"')
          echo "::add-mask::$PASSWORD"
          echo "password=$PASSWORD" >> $GITHUB_OUTPUT

      - name: Build MCP Server
        run: |
          go build -v -o mythic-mcp ./cmd/mythic-mcp

      - name: Run E2E Tests
        env:
          MYTHIC_URL: "https://127.0.0.1:7443"
          MYTHIC_USERNAME: "mythic_admin"
          MYTHIC_PASSWORD: ${{ steps.mythic-creds.outputs.password }}
          MYTHIC_SKIP_TLS_VERIFY: "true"
        run: |
          set -o pipefail
          go test -v \
            -tags=e2e \
            -timeout 20m \
            -coverprofile=coverage-e2e.out \
            ./tests/integration/... \
            2>&1 | tee e2e-test-output.log

      - name: Check E2E Skip Rate
        run: |
          SKIPS=$(grep -c "SKIP" e2e-test-output.log || true)
          echo "Skip count: $SKIPS"
          if [ $SKIPS -gt 0 ]; then
            echo "ERROR: $SKIPS tests skipped - all E2E tests must run!"
            exit 1
          fi

      - name: Upload E2E coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage-e2e.out
          flags: e2e

      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: e2e-test-results
          path: e2e-test-output.log
          retention-days: 7

      - name: Collect Mythic logs on failure
        if: failure()
        run: |
          cd /tmp/mythic
          sudo ./mythic-cli logs > mythic-logs.txt 2>&1 || true
          sudo docker ps -a > docker-containers.txt || true

      - name: Upload failure logs
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: mythic-failure-logs
          path: |
            /tmp/mythic/mythic-logs.txt
            /tmp/mythic/docker-containers.txt
          retention-days: 7

      - name: Cleanup Mythic
        if: always()
        run: |
          cd /tmp/mythic
          sudo ./mythic-cli stop || true
          sudo docker system prune -af || true

  all-tests-passed:
    name: All Tests Passed
    runs-on: ubuntu-latest
    needs: [unit-tests, integration-tests, e2e-tests]
    steps:
      - name: Success
        run: echo "All tests passed!"
```

---

### Workflow 2: Release Pipeline

**File:** `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  build:
    name: Build Release Binaries
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build binary
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          OUTPUT_NAME="mythic-mcp-${GOOS}-${GOARCH}"
          if [ "${GOOS}" = "windows" ]; then
            OUTPUT_NAME="${OUTPUT_NAME}.exe"
          fi

          go build \
            -v \
            -ldflags "-X main.Version=${GITHUB_REF#refs/tags/}" \
            -o "${OUTPUT_NAME}" \
            ./cmd/mythic-mcp

          # Create archive
          if [ "${GOOS}" = "windows" ]; then
            zip "${OUTPUT_NAME}.zip" "${OUTPUT_NAME}"
          else
            tar czf "${OUTPUT_NAME}.tar.gz" "${OUTPUT_NAME}"
          fi

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: mythic-mcp-${{ matrix.goos }}-${{ matrix.goarch }}
          path: mythic-mcp-*

  release:
    name: Create GitHub Release
    runs-on: ubuntu-latest
    needs: build
    permissions:
      contents: write
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Download artifacts
        uses: actions/download-artifact@v4
        with:
          path: artifacts

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: artifacts/**/*
          generate_release_notes: true
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### Workflow 3: Dependency Update

**File:** `.github/workflows/dependencies.yml`

```yaml
name: Update Dependencies

on:
  schedule:
    # Run weekly on Monday at 9 AM UTC
    - cron: '0 9 * * 1'
  workflow_dispatch:

jobs:
  update-deps:
    name: Update Go Dependencies
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Update dependencies
        run: |
          go get -u ./...
          go mod tidy

      - name: Run tests
        run: |
          go test -v ./...

      - name: Create Pull Request
        uses: peter-evans/create-pull-request@v6
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
          commit-message: 'chore: update Go dependencies'
          title: 'chore: update Go dependencies'
          body: |
            Automated dependency update.

            Changes:
            - Updated all Go dependencies to latest versions
            - Ran go mod tidy
            - All tests passing

          branch: deps/update-go-dependencies
          delete-branch: true
```

---

## Test Pipeline

### Pipeline Stages

**Stage 1: Lint & Format (30s)**
- Run `golangci-lint` with full rule set
- Check `go fmt` compliance
- Validate `go.mod` and `go.sum`

**Stage 2: Unit Tests (10s)**
- Run tests in `./pkg/...` and `./tests/unit/...`
- Generate coverage report
- Require >90% coverage
- Upload to Codecov

**Stage 3: Integration Tests (30s)**
- Run integration tests with mocked transport
- No external dependencies
- Verify tool registration
- Test error handling

**Stage 4: E2E Tests (10-12min)**
- Clone Mythic Framework
- Build and start Mythic
- Install Poseidon agent
- Run full E2E test suite
- Verify 0% skip rate
- Generate coverage report

### Parallel Execution

```yaml
# Jobs run in parallel when possible
jobs:
  lint:          # 30s - no dependencies
  unit-tests:    # 10s - depends on lint
  integration:   # 30s - depends on lint
  e2e-tests:     # 12m - depends on unit+integration
```

**Total Pipeline Time:** ~13 minutes (with parallelization)

---

## Release Pipeline

### Release Process

**Trigger:** Push tag matching `v*.*.*` (e.g., `v1.0.0`)

**Steps:**
1. Build binaries for all platforms
2. Create GitHub release
3. Upload artifacts
4. Generate release notes

### Build Matrix

| OS | Architecture | Output |
|----|--------------|--------|
| Linux | amd64 | `mythic-mcp-linux-amd64.tar.gz` |
| Linux | arm64 | `mythic-mcp-linux-arm64.tar.gz` |
| macOS | amd64 | `mythic-mcp-darwin-amd64.tar.gz` |
| macOS | arm64 | `mythic-mcp-darwin-arm64.tar.gz` |
| Windows | amd64 | `mythic-mcp-windows-amd64.zip` |

### Version Embedding

```go
// cmd/mythic-mcp/main.go
package main

var (
    // Version is set via ldflags during build
    Version = "dev"
)

func main() {
    if len(os.Args) > 1 && os.Args[1] == "version" {
        fmt.Printf("mythic-mcp version %s\n", Version)
        os.Exit(0)
    }
    // ...
}
```

Build command:
```bash
go build -ldflags "-X main.Version=v1.0.0" ./cmd/mythic-mcp
```

---

## Caching Strategy

### Go Module Cache

```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.23'
    cache: true  # Automatically caches go mod and build cache
```

**Benefit:** Speeds up dependency downloads from ~30s to ~5s

### Docker Layer Cache

```yaml
- name: Cache Mythic Docker layers
  uses: actions/cache@v4
  with:
    path: /var/lib/docker
    key: ${{ runner.os }}-docker-${{ hashFiles('/tmp/mythic/Dockerfile') }}
    restore-keys: |
      ${{ runner.os }}-docker-
```

**Benefit:** Speeds up Mythic startup from ~5min to ~2min

### Build Artifact Cache

```yaml
- name: Cache Go build
  uses: actions/cache@v4
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

**Benefit:** Speeds up builds from ~60s to ~10s

---

## Security Considerations

### Secrets Management

**GitHub Secrets:**
- `CODECOV_TOKEN` - Codecov upload token
- `GITHUB_TOKEN` - Automatically provided, used for releases

**Runtime Secrets:**
- Mythic password extracted from `.env` file
- Masked in logs using `::add-mask::`
- Never committed to repo

### Dependency Security

```yaml
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'fs'
    scan-ref: '.'
    format: 'sarif'
    output: 'trivy-results.sarif'

- name: Upload Trivy results to GitHub Security
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: 'trivy-results.sarif'
```

### Code Scanning

```yaml
- name: Initialize CodeQL
  uses: github/codeql-action/init@v3
  with:
    languages: go

- name: Perform CodeQL Analysis
  uses: github/codeql-action/analyze@v3
```

---

## Monitoring & Notifications

### Test Failure Notifications

```yaml
- name: Notify on failure
  if: failure()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    text: 'E2E tests failed in ${{ github.repository }}'
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

### Coverage Tracking

- Codecov integration
- Coverage badge in README
- Block PRs if coverage drops

### Performance Tracking

```yaml
- name: Benchmark
  run: |
    go test -bench=. -benchmem ./... | tee bench-output.txt

- name: Store benchmark result
  uses: benchmark-action/github-action-benchmark@v1
  with:
    tool: 'go'
    output-file-path: bench-output.txt
    github-token: ${{ secrets.GITHUB_TOKEN }}
    auto-push: true
```

---

## CI Workflow Best Practices

### 1. Fail Fast

- Run linting first (catches most issues in 30s)
- Run unit tests before E2E (faster feedback)
- Use `timeout-minutes` to prevent hanging jobs

### 2. Clear Error Messages

```yaml
- name: Check coverage
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "::notice::Coverage is ${COVERAGE}%"
    if (( $(echo "$COVERAGE < 90" | bc -l) )); then
      echo "::error::Coverage is below 90% threshold"
      exit 1
    fi
```

### 3. Artifact Retention

```yaml
- name: Upload logs
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: test-logs
    path: '**/*.log'
    retention-days: 7  # Don't store forever
```

### 4. Resource Cleanup

```yaml
- name: Cleanup
  if: always()
  run: |
    sudo docker system prune -af
    sudo rm -rf /tmp/mythic
```

---

## Success Criteria

### Pipeline Requirements

- [ ] All tests pass in <15 minutes
- [ ] No flaky tests (>99% reliability)
- [ ] Coverage >90% maintained
- [ ] Zero skip rate on E2E tests
- [ ] Clear failure messages
- [ ] Automatic dependency updates
- [ ] Security scanning integrated
- [ ] Release automation working

### Developer Experience

- [ ] Fast feedback on PRs (<2 minutes for quick checks)
- [ ] Clear CI status in PR
- [ ] Easy to reproduce failures locally
- [ ] Detailed logs available
- [ ] Coverage reports visible

---

**Status:** CI/CD Design Complete
**Next Document:** [05-IMPLEMENTATION-ROADMAP.md](05-IMPLEMENTATION-ROADMAP.md)
