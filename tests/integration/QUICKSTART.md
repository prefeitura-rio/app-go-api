# Integration Tests - Quick Start Guide

## Prerequisites

1. **Database**: PostgreSQL instance running
2. **Environment Variables**: Set up database connection

```bash
export RUN_INTEGRATION_TESTS=1
export DATABASE_URL="postgresql://user:password@localhost:5432/dbname?sslmode=disable"
```

## Quick Commands

### Run All Tests
```bash
# All integration tests (30 min timeout)
just test-integration

# With race detection
just test-integration-race
```

### Run Specific Suites
```bash
# Workflow tests only (~10 min)
just test-integration-workflows

# Stress tests only (~20 min)
just test-stress

# Error recovery tests only (~10 min)
just test-error-recovery
```

### Run Individual Workflows
```bash
# Enrollment workflow (10 tests)
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run TestEnrollmentWorkflow

# MEI workflow (10 tests)
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run TestMEIWorkflow

# Job workflow (10 tests)
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run TestJobWorkflow

# Curriculum workflow (5 tests)
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run TestCurriculoWorkflow
```

### Run Specific Tests
```bash
# Single test
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run TestEnrollmentWorkflow_CompleteFlow

# Pattern matching
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run "Concurrent"
```

## Test Output

### Successful Test
```
=== RUN   TestEnrollmentWorkflow_CompleteFlow
--- PASS: TestEnrollmentWorkflow_CompleteFlow (0.52s)
PASS
```

### Stress Test Output
```
=== RUN   TestStress_ConcurrentEnrollments
    stress_test.go:45: Concurrent enrollments: 100/100 succeeded in 8.2s
--- PASS: TestStress_ConcurrentEnrollments (8.20s)
```

### Skipped Test (No DB)
```
=== RUN   TestEnrollmentWorkflow_CompleteFlow
    test_helpers.go:18: Skipping integration test: set RUN_INTEGRATION_TESTS=1 or DATABASE_URL to run
--- SKIP: TestEnrollmentWorkflow_CompleteFlow (0.00s)
```

## Performance Expectations

| Test Suite | Duration | Operations |
|------------|----------|------------|
| Enrollment Workflow | 1-2 min | 10 tests |
| MEI Workflow | 1-2 min | 10 tests |
| Job Workflow | 2-3 min | 10 tests |
| Curriculum Workflow | 1 min | 5 tests |
| Stress Tests | 10-20 min | 1000+ ops |
| Error Recovery | 2-3 min | 15 tests |
| **Total** | **20-40 min** | **70 tests** |

## Troubleshooting

### Database Connection Failed
```bash
# Verify database is running
psql -h localhost -U user -d dbname -c "SELECT 1"

# Check DATABASE_URL format
echo $DATABASE_URL
# Should be: postgresql://user:pass@host:port/dbname?sslmode=disable
```

### Tests Timeout
```bash
# Increase timeout (default: 30m)
RUN_INTEGRATION_TESTS=1 go test -v -timeout 60m ./tests/integration/...

# Run with verbose output
RUN_INTEGRATION_TESTS=1 go test -v -timeout 30m ./tests/integration/... | tee test_output.log
```

### Race Detector Issues
```bash
# Run without race detector
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration/...

# Run specific test with race
RUN_INTEGRATION_TESTS=1 go test -v -race ./tests/integration -run TestEnrollmentWorkflow_CompleteFlow
```

### Memory Issues
```bash
# Run tests sequentially (no parallel)
RUN_INTEGRATION_TESTS=1 go test -v -p 1 ./tests/integration/...

# Limit concurrent tests
RUN_INTEGRATION_TESTS=1 go test -v -parallel 2 ./tests/integration/...
```

## Development Workflow

### Before Committing
```bash
# Run workflow tests to validate changes
just test-integration-workflows

# Check for race conditions
just test-integration-race
```

### Before Deploying
```bash
# Run all tests including stress tests
just test-integration

# Verify error handling
just test-error-recovery
```

### Performance Testing
```bash
# Run stress tests and monitor
just test-stress | tee stress_results.log

# Check for performance degradation
grep "took" stress_results.log
```

## CI/CD Integration

### GitHub Actions Example
```yaml
- name: Run Integration Tests
  env:
    RUN_INTEGRATION_TESTS: 1
    DATABASE_URL: ${{ secrets.DATABASE_URL }}
  run: |
    just test-integration
```

### Local CI Simulation
```bash
# Simulate CI environment
docker-compose up -d postgres
export RUN_INTEGRATION_TESTS=1
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/test?sslmode=disable"
just test-integration
```

## Tips

1. **Run Frequently**: Run workflow tests during development
2. **Use Patterns**: Use `-run` to filter tests by pattern
3. **Check Logs**: Tests log performance metrics
4. **Monitor Resources**: Watch memory and CPU during stress tests
5. **Clean Up**: Tests clean up after themselves, but verify DB state
6. **Parallel Execution**: Tests can run in parallel (default)
7. **Race Detection**: Always run with `-race` before committing

## Common Patterns

### Debug Single Test
```bash
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run TestEnrollmentWorkflow_CompleteFlow -test.v
```

### Run Subset of Tests
```bash
# All "Complete Flow" tests
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run "CompleteFlow"

# All bulk operation tests
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run "Bulk"
```

### Performance Profiling
```bash
# CPU profile
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run TestStress -cpuprofile cpu.prof

# Memory profile
RUN_INTEGRATION_TESTS=1 go test -v ./tests/integration -run TestStress -memprofile mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Success Criteria

All tests should:
- ✅ Pass without errors
- ✅ Complete within timeout
- ✅ Show no race conditions
- ✅ Maintain data integrity
- ✅ Meet performance benchmarks

Example successful output:
```
=== RUN   TestEnrollmentWorkflow
=== RUN   TestEnrollmentWorkflow_CompleteFlow
=== RUN   TestEnrollmentWorkflow_MultipleEnrollments
=== RUN   TestEnrollmentWorkflow_DuplicateEnrollment
...
--- PASS: TestEnrollmentWorkflow (10.25s)
    --- PASS: TestEnrollmentWorkflow_CompleteFlow (0.52s)
    --- PASS: TestEnrollmentWorkflow_MultipleEnrollments (1.23s)
    --- PASS: TestEnrollmentWorkflow_DuplicateEnrollment (0.31s)
    ...
PASS
ok      github.com/prefeitura-rio/app-go-api/tests/integration  10.256s
```

## Need Help?

- See `README.md` for detailed documentation
- Check test comments for expected behavior
- Review `test_helpers.go` for test utilities
- Ask in team chat for support
