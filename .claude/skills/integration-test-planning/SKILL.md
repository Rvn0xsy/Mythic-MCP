---
name: integration-test-planning
description: Designs comprehensive integration and end-to-end test strategies with real services, test data management, and Docker orchestration. Use when planning tests, testing strategy, implementing test coverage, setting up test environments, or validating integration points. Focuses on real service testing over mocking.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash
---

# Integration Test Planning

Comprehensive guide for designing and implementing integration and end-to-end tests that validate real system behavior.

## When to Use This Skill

Invoke this skill when:
- Planning test strategy for a feature
- Implementing integration tests
- Setting up test environments
- Designing test data management
- Troubleshooting test failures
- Validating service integrations
- Ensuring test coverage

## Test Strategy Framework

### Test Level Definitions

**Integration Tests:**
- Test interactions between 2+ services/components
- Use real databases, APIs, caches
- Validate data flows across boundaries
- Test service contracts and interfaces
- Shorter than E2E, focused on subsystems

**End-to-End Tests:**
- Test complete user workflows
- Involve all system components
- Validate full stack integration
- Test from entry point to data persistence
- Realistic user scenarios

**Unit Tests (Minimal in CI-First):**
- Only for complex business logic
- Algorithms, calculations, transformations
- NOT for database queries, API calls, or integrations
- Should be a small portion of total tests

## Test Environment Setup

### Docker Compose for Real Services

**Example: Web Application Test Environment**

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.test
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
    environment:
      NODE_ENV: test
      DATABASE_URL: postgresql://testuser:testpass@postgres:5432/testdb
      REDIS_URL: redis://redis:6379/0
      JWT_SECRET: test-secret-key
      API_PORT: 3000
    ports:
      - "3000:3000"

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
      POSTGRES_DB: testdb
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U testuser"]
      interval: 5s
      timeout: 5s
      retries: 5
    tmpfs:
      - /var/lib/postgresql/data  # Fast in-memory storage for tests

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  # Optional: External API mock
  mock-api:
    image: mockserver/mockserver:latest
    environment:
      MOCKSERVER_INITIALIZATION_JSON_PATH: /config/expectations.json
    volumes:
      - ./tests/mocks:/config
```

**Key Principles:**
- Use `depends_on` with health checks for startup ordering
- Use `tmpfs` for fast test databases
- Use environment variables for configuration
- Use real service images, not mocks (except for external APIs you don't control)

### Database Migrations and Seeding

**Migration Strategy:**

```bash
# Run migrations before tests
docker-compose -f docker-compose.test.yml up -d postgres
sleep 5  # Wait for healthy
docker-compose -f docker-compose.test.yml exec -T postgres psql -U testuser -d testdb < schema.sql

# Or use migration tool
npm run migrate:test
# or
go run migrations/migrate.go up
```

**Seed Data Strategy:**

```sql
-- tests/fixtures/seed.sql
-- Minimal shared data for all tests

-- Test admin user
INSERT INTO users (id, email, role, created_at)
VALUES (1, 'admin@test.com', 'admin', NOW());

-- Test organization
INSERT INTO organizations (id, name, created_at)
VALUES (1, 'Test Org', NOW());

-- Don't seed test-specific data here
-- Use EnsureXExists helpers in individual tests
```

## Test Data Management

### The EnsureXExists Pattern

**Purpose:** Automatically create test prerequisites without skipping tests

**Template:**

```go
func EnsureXExists(t *testing.T, client *Client) *X {
    t.Helper()

    // 1. Check if shared resource already exists
    if sharedX != nil {
        // Verify it's still valid
        if err := client.PingX(ctx, sharedX.ID); err == nil {
            return sharedX
        }
    }

    // 2. Create the resource
    x, err := client.CreateX(ctx, &XRequest{
        // Use deterministic test data
        Name: "test-x-" + t.Name(),
        // Required fields only
    })
    require.NoError(t, err, "Failed to create test X")

    // 3. Store as shared resource
    sharedX = x

    // 4. Cleanup strategy (IMPORTANT)
    t.Cleanup(func() {
        // Only clean up LOCAL files created
        if x.LocalFile != "" {
            os.Remove(x.LocalFile)
        }
        // DON'T delete x from external system
        // It's shared across tests
    })

    return x
}
```

**JavaScript/TypeScript Example:**

```typescript
let sharedUser: User | null = null;

export async function ensureUserExists(client: Client): Promise<User> {
  // 1. Check if shared resource exists
  if (sharedUser) {
    try {
      await client.getUser(sharedUser.id);
      return sharedUser;
    } catch (e) {
      sharedUser = null;  // Invalidated, recreate
    }
  }

  // 2. Create the resource
  sharedUser = await client.createUser({
    email: `test-${Date.now()}@example.com`,
    name: "Test User",
    password: "testpass123",
  });

  return sharedUser;
}
```

### Shared vs Per-Test Resources

**Decision Matrix:**

| Resource Type | Share? | Reason |
|--------------|--------|--------|
| Database connection | ✅ Yes | Expensive to establish |
| Authenticated session | ✅ Yes | Slow to create |
| User accounts | ✅ Yes | Limited by rate limits |
| Container instances | ✅ Yes | Slow to start |
| Built artifacts/payloads | ✅ Yes | Time-consuming |
| Individual records | ❌ No | Fast to create, test isolation |
| Test-specific data | ❌ No | Unique per test |
| Error scenarios | ❌ No | May corrupt shared state |

**Example: API Client SDK Tests**

```go
var (
    // Shared - expensive
    sharedClient     *mythic.Client
    sharedCallback   *mythic.Callback
    sharedPayload    *mythic.Payload

    // Not shared - cheap and per-test
    // tasks, individual files, responses
)

func EnsureCallbackExists(t *testing.T) *mythic.Callback {
    if sharedCallback != nil {
        return sharedCallback
    }

    // Requires payload (also expensive)
    payload := EnsurePayloadExists(t)

    // Build and execute payload to get callback
    callback := buildAndExecutePayload(t, payload)
    sharedCallback = callback

    return callback
}

func TestE2E_IssueTask(t *testing.T) {
    // Reuse expensive shared callback
    callback := EnsureCallbackExists(t)

    // Create cheap per-test task
    task, err := sharedClient.IssueTask(ctx, &TaskRequest{
        CallbackID: callback.ID,
        Command:    "shell",
        Params:     "whoami",
    })
    require.NoError(t, err)

    // Test-specific assertions
    assert.Equal(t, callback.ID, task.CallbackID)
}
```

## Integration Test Patterns

### Pattern: API Integration Test

**Tests API endpoint with real database:**

```javascript
describe('User API Integration', () => {
  let client;
  let db;

  beforeAll(async () => {
    // Connect to test database
    db = await connectToTestDB();
    await db.migrate.latest();

    // Create test client
    client = new APIClient('http://localhost:3000');
  });

  beforeEach(async () => {
    // Clean slate for each test
    await db('users').truncate();
    await db('sessions').truncate();
  });

  it('creates user and persists to database', async () => {
    // 1. Create via API
    const response = await client.post('/api/users', {
      email: 'test@example.com',
      name: 'Test User',
      password: 'password123',
    });

    expect(response.status).toBe(201);
    const userId = response.data.id;

    // 2. Verify in database directly
    const dbUser = await db('users').where({ id: userId }).first();
    expect(dbUser).toBeDefined();
    expect(dbUser.email).toBe('test@example.com');

    // 3. Verify via API retrieval
    const getResponse = await client.get(`/api/users/${userId}`);
    expect(getResponse.status).toBe(200);
    expect(getResponse.data.email).toBe('test@example.com');
  });

  it('validates email uniqueness constraint', async () => {
    // Create first user
    await client.post('/api/users', {
      email: 'duplicate@example.com',
      name: 'User 1',
      password: 'pass123',
    });

    // Try to create duplicate - should fail at DB level
    await expect(
      client.post('/api/users', {
        email: 'duplicate@example.com',  // Same email
        name: 'User 2',
        password: 'pass456',
      })
    ).rejects.toMatchObject({
      response: {
        status: 409,
        data: { error: expect.stringContaining('email') },
      },
    });
  });
});
```

### Pattern: Service-to-Service Integration

**Tests interaction between microservices:**

```python
class TestOrderPaymentIntegration:
    """
    Tests integration between Order Service and Payment Service
    Uses real database and real service instances
    """

    @classmethod
    def setup_class(cls):
        # Start services with docker-compose
        subprocess.run(['docker-compose', '-f', 'docker-compose.test.yml', 'up', '-d'])
        time.sleep(10)  # Wait for services to be healthy

        cls.order_client = OrderServiceClient('http://localhost:8001')
        cls.payment_client = PaymentServiceClient('http://localhost:8002')
        cls.db = connect_to_test_db()

    def test_order_creation_triggers_payment(self):
        # 1. Create order through Order Service
        order_response = self.order_client.create_order({
            'user_id': 123,
            'items': [
                {'product_id': 1, 'quantity': 2, 'price': 29.99}
            ],
            'total': 59.98
        })

        order_id = order_response['id']
        assert order_response['status'] == 'pending_payment'

        # 2. Order Service should have called Payment Service
        # Check payment was created
        time.sleep(2)  # Allow async processing

        payments = self.payment_client.get_payments_for_order(order_id)
        assert len(payments) == 1
        assert payments[0]['amount'] == 59.98
        assert payments[0]['status'] == 'pending'

        # 3. Verify database state
        db_order = self.db.query('SELECT * FROM orders WHERE id = ?', [order_id])
        assert db_order['payment_id'] == payments[0]['id']

    def test_payment_success_updates_order(self):
        # 1. Create order (triggers payment creation)
        order = self.order_client.create_order({
            'user_id': 123,
            'items': [{'product_id': 1, 'quantity': 1, 'price': 10.00}],
            'total': 10.00
        })
        time.sleep(2)

        # 2. Complete payment through Payment Service
        payments = self.payment_client.get_payments_for_order(order['id'])
        payment_id = payments[0]['id']

        self.payment_client.complete_payment(payment_id, {
            'payment_method': 'card',
            'transaction_id': 'txn_123456'
        })

        # 3. Payment Service should have notified Order Service
        # Order status should be updated
        time.sleep(2)  # Allow webhook/event processing

        updated_order = self.order_client.get_order(order['id'])
        assert updated_order['status'] == 'confirmed'
        assert updated_order['payment_status'] == 'paid'
```

### Pattern: Database Transaction Testing

**Tests transaction boundaries and rollback:**

```go
func TestE2E_TransactionRollback(t *testing.T) {
    db := getTestDB(t)
    service := NewUserService(db)

    // Count users before
    countBefore, err := db.Query("SELECT COUNT(*) FROM users").Scan()
    require.NoError(t, err)

    // Try to create user with invalid data (should rollback)
    _, err = service.CreateUserWithProfile(ctx, &UserRequest{
        Email: "test@example.com",
        Name:  "Test User",
        Profile: &ProfileRequest{
            Bio: strings.Repeat("x", 10000),  // Exceeds max length
        },
    })

    // Should fail
    require.Error(t, err)

    // Verify NO user was created (transaction rolled back)
    countAfter, err := db.Query("SELECT COUNT(*) FROM users").Scan()
    require.NoError(t, err)
    assert.Equal(t, countBefore, countAfter, "Transaction should have rolled back")

    // Verify NO profile was created either
    profiles, err := db.Query("SELECT COUNT(*) FROM profiles WHERE user_id IS NULL").Scan()
    require.NoError(t, err)
    assert.Equal(t, 0, profiles, "No orphaned profiles should exist")
}
```

## End-to-End Test Patterns

### Pattern: Complete User Journey

**Tests full workflow from UI to database:**

```javascript
// Using Playwright for browser automation
describe('E2E: User Registration and Login Flow', () => {
  let page;

  beforeAll(async () => {
    // Ensure test environment is clean
    await resetTestDatabase();
    page = await browser.newPage();
  });

  it('user can register, receive email, verify, and login', async () => {
    // 1. Navigate to registration
    await page.goto('http://localhost:3000/register');

    // 2. Fill registration form
    await page.fill('[name="email"]', 'newuser@example.com');
    await page.fill('[name="password"]', 'SecurePass123!');
    await page.fill('[name="confirmPassword"]', 'SecurePass123!');
    await page.fill('[name="name"]', 'New User');

    // 3. Submit form
    await page.click('[type="submit"]');

    // 4. Verify redirect to verification page
    await expect(page).toHaveURL(/\/verify-email/);
    await expect(page.locator('text=Check your email')).toBeVisible();

    // 5. Get verification token from test email service
    const token = await getVerificationToken('newuser@example.com');
    expect(token).toBeDefined();

    // 6. Visit verification link
    await page.goto(`http://localhost:3000/verify?token=${token}`);

    // 7. Verify redirect to login
    await expect(page).toHaveURL(/\/login/);
    await expect(page.locator('text=Email verified')).toBeVisible();

    // 8. Login with new account
    await page.fill('[name="email"]', 'newuser@example.com');
    await page.fill('[name="password"]', 'SecurePass123!');
    await page.click('[type="submit"]');

    // 9. Verify logged in (redirected to dashboard)
    await expect(page).toHaveURL(/\/dashboard/);
    await expect(page.locator('[data-testid="user-menu"]')).toContainText('New User');

    // 10. Verify session persists after reload
    await page.reload();
    await expect(page).toHaveURL(/\/dashboard/);
    await expect(page.locator('[data-testid="user-menu"]')).toBeVisible();

    // 11. Verify database state
    const db = await connectToTestDB();
    const user = await db('users').where({ email: 'newuser@example.com' }).first();
    expect(user).toBeDefined();
    expect(user.email_verified).toBe(true);
    expect(user.name).toBe('New User');

    // 12. Verify session in database
    const sessions = await db('sessions').where({ user_id: user.id });
    expect(sessions.length).toBeGreaterThan(0);
  });
});
```

### Pattern: Error Scenario Testing

**Tests failure paths and recovery:**

```python
def test_e2e_payment_failure_recovery():
    """
    Test complete flow when payment fails:
    1. Create order
    2. Payment fails
    3. Retry payment
    4. Order completes
    """
    client = get_test_client()

    # 1. Create order
    order = client.create_order({
        'user_id': 123,
        'items': [{'product_id': 1, 'quantity': 1, 'price': 50.00}],
        'total': 50.00
    })

    # 2. Simulate payment failure (using test payment provider)
    payment_result = client.process_payment(order['id'], {
        'payment_method': 'card',
        'card_number': '4000000000000002',  # Test card that always fails
    })

    assert payment_result['status'] == 'failed'
    assert payment_result['error'] == 'card_declined'

    # 3. Verify order status updated
    order = client.get_order(order['id'])
    assert order['status'] == 'payment_failed'
    assert order['retry_count'] == 1

    # 4. Retry with valid payment method
    retry_result = client.process_payment(order['id'], {
        'payment_method': 'card',
        'card_number': '4242424242424242',  # Test card that succeeds
    })

    assert retry_result['status'] == 'success'

    # 5. Verify order completed
    order = client.get_order(order['id'])
    assert order['status'] == 'confirmed'
    assert order['retry_count'] == 2

    # 6. Verify payment audit trail
    payment_attempts = client.get_payment_attempts(order['id'])
    assert len(payment_attempts) == 2
    assert payment_attempts[0]['status'] == 'failed'
    assert payment_attempts[1]['status'] == 'success'
```

## Test Organization and Naming

### Phase-Based Organization

```
tests/
├── integration/
│   ├── phase0_schema_test.go        # Schema validation
│   ├── phase1_auth_test.go          # Authentication
│   ├── phase2_users_test.go         # User management
│   ├── phase3_orders_test.go        # Business logic
│   ├── phase4_payments_test.go      # External integrations
│   └── phase5_edge_cases_test.go    # Error scenarios
└── e2e/
    ├── user_registration_flow_test.js
    ├── checkout_flow_test.js
    └── admin_workflow_test.js
```

**Why phases matter:**
- Failures in Phase 0/1 indicate fundamental issues
- Phases 2-3 are core functionality
- Phases 4-5 are advanced features
- Failed phase indicates which layer is broken

### Naming Conventions

**Integration tests:**
```
Test[E2E|Integration]_<ServiceName>_<Action>_<Scenario>

Examples:
- TestIntegration_Users_Create_ValidData
- TestIntegration_Users_Create_DuplicateEmail
- TestE2E_Orders_CreateAndPay_Success
- TestE2E_Orders_CreateAndPay_PaymentFailure
```

**Test function structure:**
```go
func TestE2E_Orders_CreateAndPay_Success(t *testing.T) {
    // Setup: Arrange test data
    user := EnsureUserExists(t, client)
    product := EnsureProductExists(t, client)

    // Execute: Perform the operation (Act)
    order, err := client.CreateOrder(ctx, &OrderRequest{
        UserID: user.ID,
        Items:  []Item{{ProductID: product.ID, Quantity: 1}},
    })
    require.NoError(t, err)

    payment, err := client.ProcessPayment(ctx, order.ID, validCard)
    require.NoError(t, err)

    // Verify: Assert expectations
    assert.Equal(t, "confirmed", order.Status)
    assert.Equal(t, "success", payment.Status)

    // Verify side effects
    dbOrder := getOrderFromDB(t, order.ID)
    assert.Equal(t, "confirmed", dbOrder.Status)
}
```

## Troubleshooting Test Failures

### Debugging Integration Tests

**Step 1: Identify the failure point**
```bash
# Get test output with verbose logging
go test -v ./tests/integration/... 2>&1 | tee test.log

# Or for JavaScript
npm test -- --verbose 2>&1 | tee test.log

# Find the exact failure
grep "FAIL" test.log
grep "Error:" test.log
```

**Step 2: Categorize the error**

| Error Pattern | Likely Cause | Solution |
|--------------|--------------|----------|
| "connection refused" | Service not running | Check docker-compose, health checks |
| "field 'X' not found" | Schema mismatch | Check migrations, search for patterns |
| "expected X, got Y" | Test assertion wrong | Verify SDK behavior, update test |
| "timeout" | Service too slow | Increase timeout, check service logs |
| "constraint violation" | Test data issue | Check for duplicate data, clean state |

**Step 3: Inspect service logs**
```bash
# View logs for all services
docker-compose -f docker-compose.test.yml logs

# View specific service
docker-compose -f docker-compose.test.yml logs postgres

# Follow logs in real-time
docker-compose -f docker-compose.test.yml logs -f app
```

**Step 4: Check database state**
```bash
# Connect to test database
docker-compose -f docker-compose.test.yml exec postgres psql -U testuser -d testdb

# Query relevant tables
SELECT * FROM users ORDER BY created_at DESC LIMIT 5;
SELECT * FROM orders WHERE status = 'failed';
```

## Best Practices Summary

### Do's ✅

- **Use real services** - Docker containers for databases, caches, queues
- **Automate test data** - EnsureXExists helpers, no manual setup
- **Share expensive resources** - Connections, sessions, built artifacts
- **Test realistic scenarios** - Real user flows, not just happy paths
- **Clean between tests** - Truncate tables, reset state
- **Run locally first** - `docker-compose up` before CI
- **Fast feedback** - Optimize test speed, parallel execution
- **Clear assertions** - Specific error messages
- **Document setup** - README with environment requirements

### Don'ts ❌

- **Don't skip tests** - Treat skips as failures, fix prerequisites
- **Don't mock databases** - Use real DB in integration tests
- **Don't share mutable state** - Isolate per-test data
- **Don't hardcode IDs** - Use dynamic test data
- **Don't test implementation** - Test behavior, not internals
- **Don't ignore flaky tests** - Fix root cause of intermittent failures
- **Don't leave orphaned resources** - Clean up after tests
- **Don't commit .env secrets** - Use example files, document setup

## Quick Reference

### Test Data Helper Template

```go
func EnsureXExists(t *testing.T, client *Client) *X {
    t.Helper()
    if sharedX != nil { return sharedX }
    x, err := client.CreateX(ctx, &XRequest{/*...*/})
    require.NoError(t, err)
    sharedX = x
    t.Cleanup(func() { /* cleanup local files only */ })
    return x
}
```

### Integration Test Template

```go
func TestIntegration_X(t *testing.T) {
    prereq := EnsurePrereqExists(t, client)
    result, err := client.Method(ctx, prereq.ID)
    require.NoError(t, err)
    assert.Equal(t, expected, result.Field)
    verifyDatabaseState(t, result.ID)
}
```

### E2E Test Template

```javascript
it('complete user flow', async () => {
    await setupTestData();
    await performUserAction();
    await verifyUIState();
    await verifyDatabaseState();
    await verifyExternalSideEffects();
});
```

---

**Remember:** Integration tests catch the bugs that unit tests miss. Test the integration, not the isolation.
