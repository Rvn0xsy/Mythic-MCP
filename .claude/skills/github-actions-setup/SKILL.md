---
name: github-actions-setup
description: Implements and optimizes GitHub Actions CI/CD pipelines with comprehensive testing, caching, parallelization, and deployment automation. Use when setting up CI/CD, creating workflows, implementing pipelines, configuring GitHub Actions, optimizing CI performance, or automating deployments.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash
---

# GitHub Actions CI/CD Setup

Complete guide for implementing production-ready GitHub Actions workflows with comprehensive testing and deployment automation.

## When to Use This Skill

Invoke this skill when:
- Setting up CI/CD for a new project
- Creating GitHub Actions workflows
- Implementing test automation in CI
- Optimizing CI performance (speed, cost)
- Adding deployment pipelines
- Debugging CI failures
- Implementing matrix testing
- Setting up secrets and environment variables

## Workflow Design Philosophy

### Fast Feedback First

**Priority order for jobs:**

1. **Lint/Format** (30s-1min) - Catch style issues immediately
2. **Type Check** (1-2min) - Catch type errors before testing
3. **Unit Tests** (1-3min) - Quick validation if used
4. **Integration Tests** (3-10min) - Core validation
5. **E2E Tests** (5-20min) - Full system validation
6. **Build** (2-10min) - Compilation/bundling
7. **Deploy** (1-5min) - Only after all tests pass

**Why this order?**
- Fail fast on cheap checks
- Expensive tests run only if cheap ones pass
- Developers get quick feedback
- CI resources used efficiently

### Comprehensive Workflow Template

**.github/workflows/ci.yml**

```yaml
name: CI Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]
  workflow_dispatch:  # Manual trigger

# Cancel in-progress runs for same PR/branch
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

env:
  NODE_VERSION: '18'
  GO_VERSION: '1.21'
  PYTHON_VERSION: '3.11'

jobs:
  # ============================================
  # Phase 1: Fast Checks (Fail Fast)
  # ============================================

  lint:
    name: Lint and Format
    runs-on: ubuntu-latest
    timeout-minutes: 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run ESLint
        run: npm run lint

      - name: Check formatting
        run: npm run format:check

  type-check:
    name: TypeScript Type Check
    runs-on: ubuntu-latest
    timeout-minutes: 5

    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Type check
        run: npm run type-check

  # ============================================
  # Phase 2: Unit Tests (if used)
  # ============================================

  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest
    timeout-minutes: 10
    needs: [lint, type-check]  # Only run if fast checks pass

    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run unit tests
        run: npm run test:unit -- --coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage/coverage-final.json
          flags: unit

  # ============================================
  # Phase 3: Integration Tests
  # ============================================

  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest
    timeout-minutes: 20
    needs: [lint, type-check]  # Run in parallel with unit tests

    # Real service dependencies
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: testuser
          POSTGRES_PASSWORD: testpass
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run database migrations
        run: npm run migrate:test
        env:
          DATABASE_URL: postgresql://testuser:testpass@localhost:5432/testdb

      - name: Seed test data
        run: npm run seed:test
        env:
          DATABASE_URL: postgresql://testuser:testpass@localhost:5432/testdb

      - name: Run integration tests
        run: npm run test:integration
        env:
          DATABASE_URL: postgresql://testuser:testpass@localhost:5432/testdb
          REDIS_URL: redis://localhost:6379
          NODE_ENV: test

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage/coverage-final.json
          flags: integration

  # ============================================
  # Phase 4: E2E Tests
  # ============================================

  e2e-tests:
    name: E2E Tests
    runs-on: ubuntu-latest
    timeout-minutes: 30
    needs: [integration-tests]  # Only run if integration passes

    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Start services with docker-compose
        run: docker-compose -f docker-compose.test.yml up -d

      - name: Wait for services to be healthy
        run: |
          timeout 60 bash -c 'until docker-compose -f docker-compose.test.yml ps | grep -q "healthy"; do sleep 2; done'

      - name: Run E2E tests
        run: npm run test:e2e
        env:
          BASE_URL: http://localhost:3000

      - name: Upload test artifacts
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: e2e-screenshots
          path: tests/e2e/screenshots/
          retention-days: 7

      - name: Upload E2E videos
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: e2e-videos
          path: tests/e2e/videos/
          retention-days: 7

      - name: Stop services
        if: always()
        run: docker-compose -f docker-compose.test.yml down -v

  # ============================================
  # Phase 5: Build
  # ============================================

  build:
    name: Build Application
    runs-on: ubuntu-latest
    timeout-minutes: 15
    needs: [integration-tests]  # Build in parallel with E2E

    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Build application
        run: npm run build

      - name: Upload build artifacts
        uses: actions/upload-artifact@v4
        with:
          name: build-artifacts
          path: dist/
          retention-days: 7

  # ============================================
  # Phase 6: Security Scanning
  # ============================================

  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    timeout-minutes: 10
    needs: [lint]

    steps:
      - uses: actions/checkout@v4

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

      - name: Run npm audit
        run: npm audit --audit-level=moderate

  # ============================================
  # Phase 7: Deploy (only on main branch)
  # ============================================

  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    timeout-minutes: 10
    needs: [unit-tests, integration-tests, e2e-tests, build]
    if: github.ref == 'refs/heads/develop' && github.event_name == 'push'

    environment:
      name: staging
      url: https://staging.example.com

    steps:
      - uses: actions/checkout@v4

      - name: Download build artifacts
        uses: actions/download-artifact@v4
        with:
          name: build-artifacts
          path: dist/

      - name: Deploy to staging
        run: |
          # Your deployment script here
          ./scripts/deploy.sh staging
        env:
          DEPLOY_KEY: ${{ secrets.STAGING_DEPLOY_KEY }}
          DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}

  deploy-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    timeout-minutes: 15
    needs: [unit-tests, integration-tests, e2e-tests, build]
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'

    environment:
      name: production
      url: https://example.com

    steps:
      - uses: actions/checkout@v4

      - name: Download build artifacts
        uses: actions/download-artifact@v4
        with:
          name: build-artifacts
          path: dist/

      - name: Deploy to production
        run: |
          # Your deployment script here
          ./scripts/deploy.sh production
        env:
          DEPLOY_KEY: ${{ secrets.PRODUCTION_DEPLOY_KEY }}
          DATABASE_URL: ${{ secrets.PRODUCTION_DATABASE_URL }}

      - name: Run smoke tests
        run: npm run test:smoke
        env:
          BASE_URL: https://example.com
```

## Optimization Patterns

### Caching Strategies

**Dependencies caching:**

```yaml
- name: Setup Node.js with caching
  uses: actions/setup-node@v4
  with:
    node-version: '18'
    cache: 'npm'  # Automatic caching

# Or manual caching for more control
- name: Cache node modules
  uses: actions/cache@v4
  with:
    path: |
      ~/.npm
      node_modules
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-
```

**Build artifact caching:**

```yaml
- name: Cache build artifacts
  uses: actions/cache@v4
  with:
    path: |
      dist/
      .next/cache
      .turbo
    key: ${{ runner.os }}-build-${{ github.sha }}
    restore-keys: |
      ${{ runner.os }}-build-
```

**Docker layer caching:**

```yaml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3

- name: Cache Docker layers
  uses: actions/cache@v4
  with:
    path: /tmp/.buildx-cache
    key: ${{ runner.os }}-buildx-${{ github.sha }}
    restore-keys: |
      ${{ runner.os }}-buildx-

- name: Build Docker image
  uses: docker/build-push-action@v5
  with:
    context: .
    cache-from: type=local,src=/tmp/.buildx-cache
    cache-to: type=local,dest=/tmp/.buildx-cache-new,mode=max
```

### Matrix Testing

**Test across multiple versions:**

```yaml
test-matrix:
  name: Test on ${{ matrix.os }} with ${{ matrix.node-version }}
  runs-on: ${{ matrix.os }}

  strategy:
    fail-fast: false  # Don't cancel other jobs if one fails
    matrix:
      os: [ubuntu-latest, windows-latest, macos-latest]
      node-version: [16, 18, 20]
      exclude:
        # Skip expensive combinations
        - os: macos-latest
          node-version: 16

  steps:
    - uses: actions/checkout@v4

    - name: Setup Node.js ${{ matrix.node-version }}
      uses: actions/setup-node@v4
      with:
        node-version: ${{ matrix.node-version }}

    - name: Run tests
      run: npm test
```

**Database matrix testing:**

```yaml
integration-tests:
  name: Integration Tests (PostgreSQL ${{ matrix.postgres-version }})
  runs-on: ubuntu-latest

  strategy:
    matrix:
      postgres-version: [13, 14, 15]

  services:
    postgres:
      image: postgres:${{ matrix.postgres-version }}
      env:
        POSTGRES_USER: testuser
        POSTGRES_PASSWORD: testpass
        POSTGRES_DB: testdb
      options: >-
        --health-cmd pg_isready
        --health-interval 10s
        --health-timeout 5s
        --health-retries 5
      ports:
        - 5432:5432

  steps:
    - uses: actions/checkout@v4
    - name: Run tests
      run: npm run test:integration
```

### Parallel Test Execution

**Split tests into shards:**

```yaml
e2e-tests:
  name: E2E Tests (Shard ${{ matrix.shard }})
  runs-on: ubuntu-latest

  strategy:
    matrix:
      shard: [1, 2, 3, 4]  # Split into 4 parallel runs

  steps:
    - uses: actions/checkout@v4

    - name: Run E2E tests (shard ${{ matrix.shard }}/4)
      run: npm run test:e2e -- --shard=${{ matrix.shard }}/4
```

## Language-Specific Patterns

### Go Project

```yaml
name: Go CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: testpass
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true  # Automatic Go module caching

      - name: Install dependencies
        run: go mod download

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

      - name: Run tests with coverage
        run: go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
        env:
          DATABASE_URL: postgresql://postgres:testpass@localhost:5432/postgres

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.txt
```

### Python Project

```yaml
name: Python CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    strategy:
      matrix:
        python-version: ['3.9', '3.10', '3.11']

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: testpass
        options: >-
          --health-cmd pg_isready
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v4

      - name: Setup Python ${{ matrix.python-version }}
        uses: actions/setup-python@v5
        with:
          python-version: ${{ matrix.python-version }}
          cache: 'pip'

      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -r requirements.txt
          pip install -r requirements-dev.txt

      - name: Run black formatter check
        run: black --check .

      - name: Run flake8
        run: flake8 .

      - name: Run mypy type check
        run: mypy src/

      - name: Run pytest with coverage
        run: pytest --cov=src --cov-report=xml
        env:
          DATABASE_URL: postgresql://postgres:testpass@localhost:5432/postgres

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.xml
```

## Docker-Based Testing

**Full environment in Docker Compose:**

```yaml
e2e-tests-docker:
  name: E2E Tests (Docker)
  runs-on: ubuntu-latest

  steps:
    - uses: actions/checkout@v4

    - name: Create .env file
      run: |
        cat << EOF > .env.test
        DATABASE_URL=postgresql://testuser:testpass@postgres:5432/testdb
        REDIS_URL=redis://redis:6379
        JWT_SECRET=test-secret
        EOF

    - name: Start services
      run: docker-compose -f docker-compose.test.yml up -d

    - name: Wait for healthy services
      run: |
        timeout 120 bash -c 'until docker-compose -f docker-compose.test.yml ps | grep -q "healthy"; do sleep 5; done'

    - name: Run database migrations
      run: docker-compose -f docker-compose.test.yml exec -T app npm run migrate

    - name: Run E2E tests
      run: docker-compose -f docker-compose.test.yml exec -T app npm run test:e2e

    - name: Collect logs on failure
      if: failure()
      run: |
        docker-compose -f docker-compose.test.yml logs > docker-logs.txt

    - name: Upload logs
      if: failure()
      uses: actions/upload-artifact@v4
      with:
        name: docker-logs
        path: docker-logs.txt

    - name: Cleanup
      if: always()
      run: docker-compose -f docker-compose.test.yml down -v
```

## Secrets and Environment Management

### Setting up secrets

```bash
# Via GitHub CLI
gh secret set DATABASE_URL --body "postgresql://user:pass@host:5432/db"
gh secret set API_KEY --body "sk-1234567890"

# Via GitHub UI:
# Settings → Secrets and variables → Actions → New repository secret
```

### Using secrets in workflows

```yaml
steps:
  - name: Deploy to production
    run: ./deploy.sh
    env:
      # Repository secrets
      DATABASE_URL: ${{ secrets.DATABASE_URL }}
      API_KEY: ${{ secrets.API_KEY }}

      # Environment secrets (more secure)
      DEPLOY_KEY: ${{ secrets.DEPLOY_KEY }}

      # Public environment variables
      NODE_ENV: production
      LOG_LEVEL: info
```

### Environment-specific configs

```yaml
deploy-staging:
  environment:
    name: staging
    url: https://staging.example.com
  steps:
    - run: echo "Deploying to staging"
      env:
        # Secrets specific to staging environment
        DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}

deploy-production:
  environment:
    name: production
    url: https://example.com
  steps:
    - run: echo "Deploying to production"
      env:
        # Secrets specific to production environment
        DATABASE_URL: ${{ secrets.PRODUCTION_DATABASE_URL }}
```

## Debugging CI Failures

### Get CI run information

```bash
# List recent runs
gh run list --limit 10

# View specific run
gh run view <run_id>

# View run logs
gh run view <run_id> --log

# View specific job logs
gh run view <run_id> --log --job=<job_id>

# Watch run in real-time
gh run watch <run_id> --exit-status

# Extract errors
gh run view <run_id> --log | grep -E "(Error|FAIL)" | head -20
```

### Enable debug logging

```yaml
# In workflow file, add:
env:
  ACTIONS_STEP_DEBUG: true
  ACTIONS_RUNNER_DEBUG: true
```

Or set via secrets:
```bash
gh secret set ACTIONS_STEP_DEBUG --body "true"
gh secret set ACTIONS_RUNNER_DEBUG --body "true"
```

### SSH into runner for debugging

```yaml
- name: Setup tmate session
  if: failure()
  uses: mxschmitt/action-tmate@v3
  timeout-minutes: 30
```

## Performance Optimization

### Job dependencies optimization

**Bad - Sequential (slow):**
```yaml
jobs:
  lint:
    runs-on: ubuntu-latest
    steps: [...]

  test:
    needs: lint  # Waits for lint
    runs-on: ubuntu-latest
    steps: [...]

  build:
    needs: test  # Waits for test
    runs-on: ubuntu-latest
    steps: [...]
```

**Good - Parallel (fast):**
```yaml
jobs:
  lint:
    runs-on: ubuntu-latest
    steps: [...]

  test:
    needs: lint  # Only depends on lint
    runs-on: ubuntu-latest
    steps: [...]

  build:
    needs: lint  # Also only depends on lint, runs parallel with test
    runs-on: ubuntu-latest
    steps: [...]

  deploy:
    needs: [test, build]  # Waits for both
    runs-on: ubuntu-latest
    steps: [...]
```

### Conditional job execution

```yaml
- name: Run expensive check
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
  run: npm run expensive-check

- name: Skip on dependabot
  if: github.actor != 'dependabot[bot]'
  run: npm run full-test
```

### Fail fast for PRs

```yaml
on:
  pull_request:
    branches: [main]

jobs:
  quick-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm run lint
      - run: npm run type-check

      # If these pass, trigger full suite
      - name: Trigger full tests
        if: success()
        run: echo "Quick checks passed, full suite will run"
```

## Monitoring and Notifications

### Slack notifications

```yaml
- name: Notify Slack on failure
  if: failure()
  uses: slackapi/slack-github-action@v1
  with:
    webhook-url: ${{ secrets.SLACK_WEBHOOK }}
    payload: |
      {
        "text": "CI Failed for ${{ github.repository }}",
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": "*CI Failure*\nRun: ${{ github.run_id }}\nBranch: ${{ github.ref }}\nCommit: ${{ github.sha }}"
            }
          }
        ]
      }
```

### Status badges

Add to README.md:
```markdown
![CI](https://github.com/username/repo/workflows/CI/badge.svg)
```

## Best Practices Summary

### Do's ✅

- **Fail fast** - Run cheap checks first (lint, type-check)
- **Cache dependencies** - Use actions/cache or built-in caching
- **Use matrix testing** - Test multiple versions in parallel
- **Set timeouts** - Prevent hanging jobs
- **Upload artifacts** - Save test results, screenshots, logs
- **Use services** - Real databases via GitHub Actions services
- **Parallel jobs** - Run independent jobs concurrently
- **Environment protection** - Use environments for production deploys
- **Security scanning** - Integrate Trivy, CodeQL, dependency checks

### Don'ts ❌

- **Don't run E2E on every commit** - Use for main branch / PRs only
- **Don't skip caching** - Wastes time and GitHub Actions minutes
- **Don't use large runners unnecessarily** - Standard runners usually sufficient
- **Don't commit secrets** - Use GitHub Secrets
- **Don't ignore flaky tests** - Fix root cause
- **Don't run all tests sequentially** - Parallelize where possible
- **Don't forget cleanup** - Always clean up Docker resources

## Quick Reference

### Essential Commands

```bash
# View recent runs
gh run list --limit 5

# Watch current run
gh run watch --exit-status

# View logs
gh run view <run_id> --log

# Re-run failed jobs
gh run rerun <run_id> --failed

# Cancel run
gh run cancel <run_id>

# View workflow file
gh workflow view ci.yml

# List workflows
gh workflow list
```

### Common Workflow Triggers

```yaml
on:
  push:
    branches: [main, develop]
    paths-ignore:
      - '**.md'
      - 'docs/**'
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC
  workflow_dispatch:  # Manual trigger
```

### Template Checklist

- [ ] Fast feedback (lint/type-check first)
- [ ] Real services (postgres, redis via services)
- [ ] Caching (dependencies, build artifacts)
- [ ] Parallel execution (independent jobs)
- [ ] Timeouts (prevent hanging)
- [ ] Artifacts upload (test results, logs)
- [ ] Security scanning (Trivy, npm audit)
- [ ] Deploy only from main (environment protection)
- [ ] Secrets properly configured
- [ ] Matrix testing if needed

---

**Remember:** CI is your rapid feedback loop. Optimize for speed and reliability, not exhaustiveness. The faster CI runs, the more frequently developers push, the earlier bugs are caught.
