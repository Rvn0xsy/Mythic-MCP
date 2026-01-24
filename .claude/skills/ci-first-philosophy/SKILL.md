---
name: ci-first-philosophy
description: Guides CI-first development methodology where comprehensive integration and end-to-end testing in GitHub Actions CI/CD validates all changes. Use when implementing features, building applications, adding functionality, creating systems, or developing code. Emphasizes reasoning before coding, real service integration over mocking, and CI as the source of truth.
allowed-tools: Read, Grep, Glob, Bash, TaskCreate, TaskUpdate, TaskList, AskUserQuestion
---

# CI-First Development Philosophy

A proven methodology that shifts from isolated unit-test-driven TDD to comprehensive CI-validated, integration-first development.

## Core Principles

### 1. Reason Before Coding
**Always start by understanding the system, not just the function:**

- What is the **desired outcome** for the user/system?
- What are the **integration points** (databases, APIs, services)?
- What **edge cases** exist at integration boundaries?
- What **test data** is needed for realistic testing?
- How will this work in the **full system context**?

**Pattern:**
```
User Request → System Thinking → Integration Analysis → Test Planning → Implementation
     ↓              ↓                    ↓                    ↓              ↓
  "Add login"   Auth flow         DB, sessions         Test users    Code + Tests
```

### 2. Integration Over Isolation

**Test Pyramid - Inverted for CI-First:**
```
┌─────────────────────────────┐
│    End-to-End Tests         │  ← MOST IMPORTANT
│    (Full system flows)      │     Test real user journeys
├─────────────────────────────┤
│   Integration Tests         │  ← CORE VALIDATION
│   (Service interactions)    │     Real DB, APIs, services
├─────────────────────────────┤
│      Unit Tests             │  ← TARGETED ONLY
│   (Complex logic only)      │     Not exhaustive
└─────────────────────────────┘
```

**Key Insight:** Unit tests alone miss real-world issues. Integration and E2E tests catch:
- Database constraint violations
- API contract mismatches
- Race conditions
- Configuration errors
- Service interaction bugs
- Deployment issues

### 3. CI as Source of Truth

**If it passes in CI, it works. If it fails in CI, it's broken.**

- CI tests run in production-like environment
- CI validates full system integration
- CI catches environment-specific issues
- Local tests can lie (wrong versions, cached state, etc.)

**Workflow:**
```bash
1. Make change
2. Commit and push
3. Watch CI run
4. If fails → analyze, fix, repeat
5. If passes → trust it
```

### 4. Treat Test Skips as Failures

**Every skipping test represents missing test coverage.**

❌ **Bad - Skip when missing data:**
```go
func TestUserProfile(t *testing.T) {
    if testUser == nil {
        t.Skip("No test user available")
    }
    // Test never runs in CI
}
```

✅ **Good - Create data automatically:**
```go
func TestUserProfile(t *testing.T) {
    user := EnsureTestUserExists(t, client)
    // Test always runs, full coverage
}
```

**Pattern - EnsureXExists helpers:**
```go
func EnsureTestUserExists(t *testing.T, client *Client) *User {
    // 1. Check if shared resource exists
    if sharedUser != nil {
        return sharedUser
    }

    // 2. Create the resource
    user, err := client.CreateUser(ctx, &UserRequest{
        Email: "test@example.com",
        Name:  "Test User",
    })
    require.NoError(t, err)
    sharedUser = user

    // 3. Cleanup (local files only, not shared resources)
    t.Cleanup(func() {
        // Clean up temp files if any
        // DON'T delete user - it's shared across tests
    })

    return user
}
```

## When to Use This Skill

Apply CI-first philosophy when:
- **Implementing features** - "Add user authentication"
- **Building applications** - "Create a blog platform"
- **Adding functionality** - "Add payment processing"
- **Creating systems** - "Build an API service"
- **Developing code** - Any non-trivial development work

## CI-First Development Process

### Phase 1: System Analysis (BEFORE coding)

**Use TodoWrite for complex features (3+ steps):**

Only use TodoWrite when the task has multiple distinct steps or affects multiple files. Create tasks that are specific and actionable.

**Pattern:**
```markdown
1. Analyze system requirements and integration points
2. Design test data and helper functions
3. Implement integration tests
4. Implement E2E tests
5. Implement core functionality
6. Verify all tests pass in CI
```

**CRITICAL: Only mark ONE task as in_progress at a time.** Complete it, mark completed, then move to the next.

**Questions to answer:**
- What services does this integrate with?
- What test data is required?
- What are the dependencies between tests?
- What can be shared vs per-test resources?

### Phase 2: Test Data Strategy

**Shared vs Per-Test Resources:**

**Share expensive resources:**
- Database connections
- Authenticated sessions
- Built payloads/artifacts
- Long-lived test accounts
- Container instances

**Create per-test:**
- Individual records/entities
- Test-specific data
- Error condition scenarios
- Cleanup validation

**Helper Function Pattern:**
```go
// Shared resource - expensive to create
func EnsureCallbackExists(t *testing.T, client *Client) *Callback {
    if sharedCallback != nil {
        return sharedCallback
    }

    callback := createCallback(t, client)
    sharedCallback = callback

    // No cleanup - shared across tests
    return callback
}

// Per-test resource - cheap to create
func CreateTask(t *testing.T, client *Client, callbackID int) *Task {
    task, err := client.IssueTask(ctx, &TaskRequest{
        CallbackID: callbackID,
        Command:    "shell",
        Params:     "whoami",
    })
    require.NoError(t, err)

    // Per-test cleanup if needed
    t.Cleanup(func() {
        // Clean up task-specific resources
    })

    return task
}
```

### Phase 3: Test Organization by Phases

**Structure tests by dependency level:**

- **Phase 0:** Schema validation (no external services needed)
- **Phase 1-2:** Core APIs (authentication, basic operations)
- **Phase 3:** Integration (service-to-service communication)
- **Phase 4:** Advanced features (complex workflows)
- **Phase 5:** Edge cases (timeouts, errors, large data)

**Benefits:**
- Failures in earlier phases block later phases (correct dependency order)
- Easy to identify which layer is broken
- Progressive validation (basic → advanced)

**Example:**
```go
func TestE2E_Phase1_Authentication(t *testing.T) {
    // Must pass before anything else works
}

func TestE2E_Phase3_UserWorkflow(t *testing.T) {
    // Depends on auth working
    user := EnsureAuthenticatedUser(t, client)
    // ...
}
```

### Phase 4: Implementation with Real Services

**Use Docker Compose for test environment:**

```yaml
# docker-compose.test.yml
version: '3.8'
services:
  app:
    build: .
    depends_on:
      - postgres
      - redis
    environment:
      DATABASE_URL: postgresql://test:test@postgres:5432/testdb
      REDIS_URL: redis://redis:6379

  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: testdb

  redis:
    image: redis:7-alpine
```

**Integration test against real services:**
```go
func TestE2E_UserCreationAndRetrieval(t *testing.T) {
    client := getAuthenticatedClient(t)

    // Create user (hits real DB)
    user, err := client.CreateUser(ctx, &UserRequest{
        Email: "test@example.com",
        Name:  "Test User",
    })
    require.NoError(t, err)

    // Retrieve user (queries real DB)
    retrieved, err := client.GetUser(ctx, user.ID)
    require.NoError(t, err)

    // Verify integration worked
    assert.Equal(t, user.Email, retrieved.Email)
    assert.NotZero(t, retrieved.CreatedAt) // DB generated this
}
```

### Phase 5: Rapid CI Feedback Loop

**Iterative development with CI:**

```bash
# 1. Implement one test + code
# 2. Commit with clear message
git add -A
git commit -m "Add user creation with DB integration test"

# 3. Push and watch CI
git push
gh run watch --exit-status

# 4. If fails, analyze quickly
gh run view <run_id> --log | grep -E "(FAIL|Error)"

# 5. Categorize error:
#    - Schema issue? Pattern match from working code
#    - Test assertion wrong? Update test
#    - Type error? Fix types
#    - Missing data? Add helper

# 6. Fix and repeat
```

## Common Patterns

### Pattern: Schema Discovery Through Code

When APIs fail with "field not found" errors:

```bash
# 1. Identify the error
# Error: field 'callback_id' not found in type 'filemeta'

# 2. Search for similar working code
grep -r "filemeta" pkg/ | grep -v "_test"

# 3. Find the pattern
# Found: artifact.go uses "artifact_text" not "artifact"
# Pattern: Fields use _text suffix in this version

# 4. Apply the pattern
# callback_id → callback_id_text
# OR use relationship: filemeta → task → callback_id
```

### Pattern: Test vs Implementation Bug

**Decision tree:**
```
Test fails with assertion error
├─ Did SDK behavior change recently? (check git log)
│  ├─ Yes → Likely test bug (assertions outdated)
│  └─ No → Continue investigating
├─ Does manual testing work?
│  ├─ Yes → Test bug (assertions wrong)
│  └─ No → SDK bug (implementation wrong)
└─ Is the behavior documented/expected?
   ├─ Yes → Test bug
   └─ No → SDK bug
```

### Pattern: Parallel vs Sequential Operations

**Parallel (independent operations):**
```python
# Reading multiple unrelated files
Read("pkg/api/users.go")
Read("pkg/api/auth.go")
Read("tests/integration/users_test.go")

# Checking git state
Bash("git status")
Bash("git log -1")
Bash("git diff")
```

**Sequential (dependent operations):**
```python
# Write file, THEN commit it
Write("new_file.go", content)
→ wait for completion →
Bash("git add new_file.go && git commit -m 'Add feature'")

# Create resource, THEN use it
result = CreateResource()
→ wait for completion →
UseResource(result.id)
```

## Debugging Strategies

### Layered Debugging Approach

**Start broad, narrow systematically:**

1. **CI Level:** Which job/phase failed?
2. **Test Level:** Which specific test failed?
3. **Operation Level:** Which function/API call failed?
4. **Error Level:** What was the error message?
5. **Code Level:** What code caused the error?

**Example:**
```
CI shows Phase 3 failed
→ Check Phase 3 logs
→ TestE2E_UserWorkflow failed
→ GetUserProfile() returned error
→ Error: "field 'profile_image' not found"
→ Check users.go GraphQL query
→ Query uses profile_image field
→ Field doesn't exist in schema (search working code)
→ Found: other code uses profile_image_url
→ Fix: profile_image → profile_image_url
```

### Pattern Matching from Existing Code

**Learn from what already works:**

```bash
# Find similar working implementations
grep -r "similar_method" pkg/

# Find field naming patterns
grep -r "graphql:\".*name\"" pkg/ | head -10

# Find test patterns
grep -r "TestE2E_" tests/integration/ | grep "User"
```

## Anti-Patterns to Avoid

❌ **Over-Engineering:**
- Don't add error handling for impossible scenarios
- Don't create abstractions for one-time use
- Don't add features beyond requirements
- Don't refactor code that isn't changing

❌ **Mocking Everything:**
- Don't mock databases in integration tests (use real DB)
- Don't mock APIs you control (test actual integration)
- Don't mock when Docker can provide real service

❌ **Skipping CI:**
- Don't assume local tests are sufficient
- Don't skip CI validation
- Don't merge without green CI

❌ **Batch Commits:**
- Don't combine unrelated changes
- Don't commit before CI passes
- Don't push large changesets without incremental validation

## Success Checklist

Before declaring a feature complete:

- [ ] All tests pass in CI (no local-only passes)
- [ ] No skipping tests (all tests run and pass)
- [ ] Integration tests use real services (not mocks)
- [ ] E2E tests cover complete user flows
- [ ] Test data is created automatically (EnsureX helpers)
- [ ] Shared resources are reused (not recreated per test)
- [ ] Small, focused commits with clear messages
- [ ] Root cause fixed (not symptoms)
- [ ] Code follows existing patterns
- [ ] Documentation explains "why" not just "what"

## Key Insights from Real Projects

### What Makes CI-First Successful:

1. **Systematic approach** - Break problems into manageable pieces
2. **CI as rapid feedback** - Push frequently, iterate quickly
3. **Pattern recognition** - Find solutions in existing code
4. **Test quality focus** - Treat skips as failures
5. **Incremental progress** - Small commits, frequent pushes
6. **Root cause analysis** - Fix problems, not symptoms
7. **Clear documentation** - Explain why, not just what

### Common Pitfalls:

1. **Over-fixing** - Changing more than necessary
2. **Assumption-based** - Not verifying root cause
3. **Incomplete cleanup** - Leaving failing tests
4. **Poor naming** - Unclear function/variable names
5. **Missing context** - Commits without explanation

## The 10 Commandments of CI-First Development

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

---

Remember: **The goal is not 100% unit test coverage. The goal is comprehensive integration and E2E test coverage that validates the system actually works in production-like conditions.**
