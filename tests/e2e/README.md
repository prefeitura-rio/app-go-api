# E2E Test Suite

Comprehensive end-to-end tests for validating deployed instances of app-go-api.

## Tests Included

### `health_check.sh`

Validates the deployed application is healthy and responsive with the following checks:

1. **Service Readiness** - Waits up to 60 seconds (30 retries × 2s) for service to become available
2. **Health Endpoint** - Validates `/v1/health` returns HTTP 200
3. **JSON Response** - Validates health endpoint returns valid JSON
4. **Status Field** - Validates health response includes `status` field
5. **Database Connectivity** - Validates database connection status (if reported)
6. **Swagger Docs** - Validates `/swagger/index.html` is accessible
7. **OpenAPI Spec** - Validates `/swagger/doc.json` is accessible
8. **Response Time** - Validates health endpoint responds within 2 seconds

## Usage

### Local Testing

```bash
# Test against local development server
API_URL=http://localhost:8080 ./tests/e2e/health_check.sh

# Test with custom retry settings
API_URL=http://localhost:8080 MAX_RETRIES=60 RETRY_DELAY=1 ./tests/e2e/health_check.sh
```

### CI/CD Integration

The E2E tests are automatically run in the following scenarios:

1. **Staging Deployment** - After blue-green preview is created (via `deploy-staging.yaml`)
2. **Production Deployment** - Could be added to `release.yaml` for canary validation

#### Environment Variables

- `API_URL` - Target API URL (default: `http://localhost:8080`)
- `MAX_RETRIES` - Maximum number of readiness check attempts (default: `30`)
- `RETRY_DELAY` - Delay between retry attempts in seconds (default: `2`)

### Example Output

```
[INFO] Starting E2E Health Check Test Suite
[INFO] Target: http://localhost:8080

[INFO] Waiting for service to be ready at http://localhost:8080...
[PASS] Service is ready (attempt 1/30)

[INFO] Running E2E tests...

[INFO] Test: Health endpoint returns HTTP 200
[PASS] Health endpoint returned HTTP 200
[INFO] Test: Health endpoint returns valid JSON
[PASS] Health endpoint returned valid JSON
[INFO] Test: Health endpoint includes 'status' field
[PASS] Health endpoint includes 'status' field: healthy
[INFO] Test: Database connectivity (via health endpoint)
[PASS] Database is connected: up
[INFO] Test: Swagger documentation is accessible
[PASS] Swagger documentation is accessible
[INFO] Test: OpenAPI spec is accessible
[PASS] OpenAPI spec is accessible
[INFO] Test: Response time is acceptable (<2s)
[PASS] Response time is 45ms (acceptable)

==================================
E2E Test Suite Summary
==================================
Tests Run:    7
Tests Passed: 7
Tests Failed: 0
==================================
[PASS] All tests passed!
```

## Exit Codes

- `0` - All tests passed
- `1` - One or more tests failed or service failed to become ready

## Adding New Tests

To add a new E2E test:

1. Create a new test function following the pattern:
```bash
test_my_new_check() {
    TESTS_RUN=$((TESTS_RUN + 1))
    log_info "Test: Description of what you're testing"

    # Perform test logic
    if [ test_condition ]; then
        log_success "Test passed"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "Test failed"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}
```

2. Call the test function in `main()`:
```bash
test_my_new_check || true
```

3. Update this README with the new test description

## Integration with Deployment Workflows

### Staging (Blue-Green)

The `deploy-staging.yaml` workflow runs basic smoke tests (health + swagger). To use the full E2E suite:

```yaml
- name: Run E2E tests against preview
  run: |
    API_URL=http://localhost:8080 ./tests/e2e/health_check.sh
```

### Production (Canary)

Could be integrated into `release.yaml` for canary validation:

```yaml
- name: Validate canary deployment
  run: |
    # Get canary service endpoint
    CANARY_URL=$(kubectl get svc go-canary -n go -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    API_URL=http://$CANARY_URL ./tests/e2e/health_check.sh
```

## Dependencies

- `curl` - For HTTP requests
- `jq` - For JSON parsing
- `bash` - Shell interpreter

All dependencies are typically available in CI/CD environments (GitHub Actions, GitLab CI, etc.).
