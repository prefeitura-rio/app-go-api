#!/bin/bash
# E2E Health Check Test Suite
# Validates the deployed application is healthy and responsive

set -e

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
MAX_RETRIES="${MAX_RETRIES:-30}"
RETRY_DELAY="${RETRY_DELAY:-2}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

# Wait for service to be ready
wait_for_service() {
    log_info "Waiting for service to be ready at $API_URL..."

    for i in $(seq 1 $MAX_RETRIES); do
        if curl -sf "$API_URL/health" > /dev/null 2>&1; then
            log_success "Service is ready (attempt $i/$MAX_RETRIES)"
            return 0
        fi

        log_info "Service not ready yet (attempt $i/$MAX_RETRIES), retrying in ${RETRY_DELAY}s..."
        sleep $RETRY_DELAY
    done

    log_error "Service failed to become ready after $MAX_RETRIES attempts"
    return 1
}

# Test: Health endpoint returns 200
test_health_endpoint() {
    TESTS_RUN=$((TESTS_RUN + 1))
    log_info "Test: Health endpoint returns HTTP 200"

    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/health")

    if [ "$HTTP_CODE" = "200" ]; then
        log_success "Health endpoint returned HTTP 200"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "Health endpoint returned HTTP $HTTP_CODE (expected 200)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Test: Health endpoint returns valid JSON
test_health_json() {
    TESTS_RUN=$((TESTS_RUN + 1))
    log_info "Test: Health endpoint returns valid JSON"

    RESPONSE=$(curl -s "$API_URL/health")

    if echo "$RESPONSE" | jq empty > /dev/null 2>&1; then
        log_success "Health endpoint returned valid JSON"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "Health endpoint returned invalid JSON: $RESPONSE"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Test: Health endpoint includes status field
test_health_status_field() {
    TESTS_RUN=$((TESTS_RUN + 1))
    log_info "Test: Health endpoint includes 'status' field"

    RESPONSE=$(curl -s "$API_URL/health")
    STATUS=$(echo "$RESPONSE" | jq -r '.status // empty')

    if [ -n "$STATUS" ]; then
        log_success "Health endpoint includes 'status' field: $STATUS"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "Health endpoint missing 'status' field: $RESPONSE"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Test: Database connectivity (if health endpoint reports it)
test_database_connectivity() {
    TESTS_RUN=$((TESTS_RUN + 1))
    log_info "Test: Database connectivity (via health endpoint)"

    RESPONSE=$(curl -s "$API_URL/health")
    DB_STATUS=$(echo "$RESPONSE" | jq -r '.database // .db // empty')

    if [ -n "$DB_STATUS" ]; then
        if [ "$DB_STATUS" = "up" ] || [ "$DB_STATUS" = "connected" ] || [ "$DB_STATUS" = "ok" ]; then
            log_success "Database is connected: $DB_STATUS"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            return 0
        else
            log_error "Database status is not healthy: $DB_STATUS"
            TESTS_FAILED=$((TESTS_FAILED + 1))
            return 1
        fi
    else
        log_info "Database status not reported in health endpoint (skipping check)"
        TESTS_RUN=$((TESTS_RUN - 1))
        return 0
    fi
}

# Test: Swagger documentation is accessible
test_swagger_docs() {
    TESTS_RUN=$((TESTS_RUN + 1))
    log_info "Test: Swagger documentation is accessible"

    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/swagger/index.html")

    if [ "$HTTP_CODE" = "200" ]; then
        log_success "Swagger documentation is accessible"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "Swagger documentation returned HTTP $HTTP_CODE (expected 200)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Test: OpenAPI spec is accessible
test_openapi_spec() {
    TESTS_RUN=$((TESTS_RUN + 1))
    log_info "Test: OpenAPI v3 spec is accessible"

    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/swagger/doc.json")

    if [ "$HTTP_CODE" = "200" ]; then
        log_success "OpenAPI spec is accessible"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "OpenAPI spec returned HTTP $HTTP_CODE (expected 200)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Test: Response time is acceptable
test_response_time() {
    TESTS_RUN=$((TESTS_RUN + 1))
    log_info "Test: Response time is acceptable (<2s)"

    START_TIME=$(date +%s%N)
    curl -sf "$API_URL/health" > /dev/null
    END_TIME=$(date +%s%N)

    DURATION_MS=$(( (END_TIME - START_TIME) / 1000000 ))

    if [ $DURATION_MS -lt 2000 ]; then
        log_success "Response time is ${DURATION_MS}ms (acceptable)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "Response time is ${DURATION_MS}ms (exceeds 2000ms threshold)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Print test summary
print_summary() {
    echo ""
    echo "=================================="
    echo "E2E Test Suite Summary"
    echo "=================================="
    echo "Tests Run:    $TESTS_RUN"
    echo "Tests Passed: $TESTS_PASSED"
    echo "Tests Failed: $TESTS_FAILED"
    echo "=================================="

    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All tests passed!"
        return 0
    else
        log_error "Some tests failed!"
        return 1
    fi
}

# Main execution
main() {
    log_info "Starting E2E Health Check Test Suite"
    log_info "Target: $API_URL"
    echo ""

    # Wait for service to be ready
    if ! wait_for_service; then
        log_error "Service failed to become ready. Aborting tests."
        exit 1
    fi

    echo ""
    log_info "Running E2E tests..."
    echo ""

    # Run all tests (continue even if some fail)
    test_health_endpoint || true
    test_health_json || true
    test_health_status_field || true
    test_database_connectivity || true
    test_swagger_docs || true
    test_openapi_spec || true
    test_response_time || true

    # Print summary and exit with appropriate code
    if ! print_summary; then
        exit 1
    fi
}

# Run main function
main
