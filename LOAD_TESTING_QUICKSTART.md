# Load Testing Quick Start Guide

This guide will help you set up and run load tests for the Skillsphere application.

## Prerequisites

- PostgreSQL database running (default: `localhost:5470`)
- Go 1.21 or higher
- Database migrations applied

## Step 1: Seed Test Data

The load tests require realistic test data with 100+ users. You have two options:

### Option A: Run test with auto-seeding (Recommended)

The tests can create their own minimal test data. Just run:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5470/myapp?sslmode=disable"
export TEST_DATABASE_URL="$DATABASE_URL"

# Run a single test that will create basic data
go test -v ./internal/app/domain/matches -run TestMatchingAlgorithmLoadPerformance
```

### Option B: Seed comprehensive test data manually

For full load testing with 150 users, sessions, reviews, and chat data:

```bash
# Apply seed migrations using psql
PGPASSWORD=postgres psql -h localhost -p 5470 -U postgres -d myapp <<EOF
-- Create 150 test users (simplified version)
INSERT INTO users (email, username, hashed_password, display_name, is_active, created_at, updated_at)
SELECT
    'testuser' || i || '@loadtest.example.com',
    'testuser' || i,
    '\$2a\$10\$rQZ5YJ3p5Y8Z5Z5Z5Z5Z5uH5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5Z5a',
    'Test User ' || i,
    true,
    NOW(),
    NOW()
FROM generate_series(1, 150) AS i;

-- Initialize user stats
INSERT INTO user_stats (user_id, total_sessions, completed_sessions, average_rating, total_reviews)
SELECT id, 0, 0, 0.0, 0 FROM users WHERE email LIKE '%@loadtest.example.com';

-- Get existing skills and assign to users
DO \$\$
DECLARE
    user_rec RECORD;
    skill_ids UUID[];
    skill_id UUID;
    num_skills INT;
BEGIN
    SELECT ARRAY_AGG(id) INTO skill_ids FROM skills LIMIT 20;

    FOR user_rec IN SELECT id FROM users WHERE email LIKE '%@loadtest.example.com' LOOP
        -- Assign 3-5 offered skills
        num_skills := 3 + FLOOR(RANDOM() * 3)::INT;
        FOR i IN 1..num_skills LOOP
            skill_id := skill_ids[1 + FLOOR(RANDOM() * array_length(skill_ids, 1))];
            INSERT INTO user_skills (user_id, skill_id, skill_type, proficiency, created_at, updated_at)
            VALUES (user_rec.id, skill_id, 'offered', 1 + FLOOR(RANDOM() * 10)::INT, NOW(), NOW())
            ON CONFLICT DO NOTHING;
        END LOOP;

        -- Assign 2-4 wanted skills
        num_skills := 2 + FLOOR(RANDOM() * 3)::INT;
        FOR i IN 1..num_skills LOOP
            skill_id := skill_ids[1 + FLOOR(RANDOM() * array_length(skill_ids, 1))];
            INSERT INTO user_skills (user_id, skill_id, skill_type, created_at, updated_at)
            VALUES (user_rec.id, skill_id, 'wanted', NOW(), NOW())
            ON CONFLICT DO NOTHING;
        END LOOP;
    END LOOP;
END \$\$;

-- Refresh materialized view for matching
REFRESH MATERIALIZED VIEW user_skill_vectors;

-- Verify
SELECT COUNT(*) as test_users FROM users WHERE email LIKE '%@loadtest.example.com';
EOF
```

## Step 2: Verify Test Data

```bash
# Check user count
PGPASSWORD=postgres psql -h localhost -p 5470 -U postgres -d myapp -c \
  "SELECT COUNT(*) FROM users WHERE email LIKE '%@loadtest.example.com';"

# Should show 150 users
```

## Step 3: Run Load Tests

### Run All Tests

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5470/myapp?sslmode=disable"
export TEST_DATABASE_URL="$DATABASE_URL"

# Run all load tests
go test -v \
  ./internal/app/domain/matches \
  ./internal/app/domain/dashboard \
  ./internal/app/domain/chat \
  ./internal/tests \
  -run "Load|Memory" \
  -timeout 15m
```

### Run Individual Test Suites

```bash
# Matching algorithm tests (should complete in <2s)
go test -v ./internal/app/domain/matches -run TestMatchingAlgorithmLoadPerformance

# Dashboard tests (should load in <1s)
go test -v ./internal/app/domain/dashboard -run TestDashboardLoadPerformance

# Chat WebSocket tests (should handle 10+ connections)
go test -v ./internal/app/domain/chat -run TestChatWebSocketLoadPerformance

# Memory leak tests
go test -v ./internal/tests -run TestMemoryUsageUnderSustainedLoad -timeout 10m
```

## Step 4: Run Benchmarks (Optional)

```bash
# Matching benchmarks
go test -bench=BenchmarkMatchingAlgorithm ./internal/app/domain/matches -benchmem -benchtime=5s

# Dashboard benchmarks
go test -bench=BenchmarkDashboardQueries ./internal/app/domain/dashboard -benchmem -benchtime=5s

# Chat benchmarks
go test -bench=BenchmarkChatOperations ./internal/app/domain/chat -benchmem -benchtime=5s
```

## Expected Results

### Success Criteria

All tests should pass with these performance metrics:

✅ **Matching Algorithm**: Finds matches in <2 seconds with 100+ users
✅ **Dashboard**: Loads in <1 second
✅ **WebSocket Chat**: Handles 10+ concurrent connections
✅ **Memory**: No leaks (<15MB growth under sustained load)

### Example Output

```
=== RUN   TestMatchingAlgorithmLoadPerformance
=== RUN   TestMatchingAlgorithmLoadPerformance/SingleUserMatching
    matching_load_test.go:45: Testing matching algorithm with 150 users
    matching_load_test.go:53: Found 20 matches in 847ms
--- PASS: TestMatchingAlgorithmLoadPerformance (0.85s)
    --- PASS: TestMatchingAlgorithmLoadPerformance/SingleUserMatching (0.85s)
```

## Cleanup Test Data

```bash
# Remove all test data
PGPASSWORD=postgres psql -h localhost -p 5470 -U postgres -d myapp <<EOF
DELETE FROM user_skills WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@loadtest.example.com');
DELETE FROM user_stats WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@loadtest.example.com');
DELETE FROM sessions WHERE initiator_id IN (SELECT id FROM users WHERE email LIKE '%@loadtest.example.com');
DELETE FROM conversations WHERE user_a_id IN (SELECT id FROM users WHERE email LIKE '%@loadtest.example.com');
DELETE FROM users WHERE email LIKE '%@loadtest.example.com';
REFRESH MATERIALIZED VIEW user_skill_vectors;
EOF
```

## Troubleshooting

### Database Connection Failed

```bash
# Check if PostgreSQL is running
docker-compose ps

# Start database if not running
docker-compose up -d db
```

### Test Fails: "need at least 100 users"

Run the seed script from Step 1 again, or check that existing test data hasn't been deleted.

### Performance Tests Fail

This could indicate actual performance issues. Check:

1. **Database indexes**: Ensure all migrations have been applied
2. **Connection pool**: Check `MaxConns` setting in database config
3. **Hardware**: Tests assume reasonable hardware (4+ CPU cores, 8GB+ RAM)
4. **Load**: Close other applications using significant resources

### Memory Leak Tests Fail

This is expected in development if:
- Debugger is attached
- Many background processes are running
- Running on a resource-constrained system

## Advanced Usage

### Profile Memory Usage

```bash
go test -memprofile=mem.prof ./internal/app/domain/matches -bench=.
go tool pprof mem.prof
```

### Profile CPU Usage

```bash
go test -cpuprofile=cpu.prof ./internal/app/domain/matches -bench=.
go tool pprof cpu.prof
```

### Run with Race Detector

```bash
go test -race ./internal/app/domain/chat -run TestChatWebSocketLoadPerformance
```

## For More Information

See the comprehensive [Load Testing Documentation](docs/LOAD_TESTING.md) for:
- Detailed test descriptions
- Architecture overview
- CI/CD integration
- Performance monitoring tips
- Best practices