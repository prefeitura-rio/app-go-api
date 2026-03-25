# Test Coverage Improvements - Phase 1

## Summary

This document outlines the test coverage improvements made to app-go-api and provides a roadmap for reaching the 60% coverage target.

## Current State

### Before
- **Overall Coverage**: 6.2%
- **Test Files**: 23
- **Critical Gaps**: Middlewares (0%), Auth (0%), Most handlers (0%), Services (10.9%)

### After Phase 1
- **Overall Coverage**: 8.3%
- **Test Files**: 29 (+6 new files)
- **Key Improvements**:
  - Auth package: 0% → 94.6%
  - Middlewares package: 0% → 74.5%
  - Models package: 11.4% → 19.4%
  - Cache package: 0% → 13.4%

### Files Added

1. **internal/auth/service_account_test.go** (94.6% coverage)
   - Tests for service account token manager
   - Concurrent access tests
   - Token expiration and refresh tests
   - HTTP error handling tests

2. **internal/middlewares/user_context_test.go** (60.4% coverage)
   - JWT decoding tests
   - User context extraction tests
   - Role-based authorization tests
   - Ownership/admin middleware tests

3. **internal/middlewares/cors_test.go**
   - CORS preflight request tests
   - CORS actual request tests
   - Header validation tests

4. **internal/middlewares/timeout_test.go**
   - Request timeout tests
   - Context cancellation tests
   - Success/failure scenarios

5. **internal/models/curso_validation_test.go** (19.4% coverage)
   - Modalidade validation and normalization
   - StatusCurso validation and normalization
   - CourseManagementType validation

6. **tests/e2e/api_test.go**
   - Comprehensive E2E endpoint tests
   - Health check structure validation
   - CORS header validation
   - Response time tests
   - Swagger documentation tests

### Coverage by Package

| Package | Before | After | Change |
|---------|--------|-------|--------|
| internal/auth | 0.0% | 94.6% | +94.6% |
| internal/middlewares | 0.0% | 74.5% | +74.5% |
| internal/models | 11.4% | 19.4% | +8.0% |
| internal/cache | 0.0% | 13.4% | +13.4% |
| internal/handlers/common | 61.3% | 61.3% | 0% |
| internal/handlers/v1 | 9.7% | 9.7% | 0% |
| internal/services | 10.9% | 10.9% | 0% |
| internal/utils | 95.4% | 95.4% | 0% |

## Challenges Encountered

### 1. Lack of Dependency Injection
The codebase uses concrete types instead of interfaces, making it difficult to mock dependencies:
- Services are instantiated with concrete Repository types
- Handlers are instantiated with concrete Service types
- No interfaces for mocking external dependencies

### 2. Missing Test Infrastructure
- No existing mocking framework configured
- No testcontainers setup for database tests
- No test fixtures or helpers for common scenarios

### 3. Complex Dependencies
- Many handlers depend on multiple services
- Services depend on repositories with database connections
- Circular dependencies in some packages

## Roadmap to 60% Coverage

### Phase 2: Service Layer (Target: +15% coverage)

**Priority Packages**:
1. **internal/services** (currently 10.9%)
   - Add interface-based mocking
   - Test business logic validation
   - Test error handling paths

2. **internal/services/empregabilidade** (currently 22.9%)
   - Comprehensive service tests
   - Validation logic tests

**Estimated Effort**: 3-4 days
**Approach**: Introduce interfaces for repositories, create mock implementations

### Phase 3: Handler Layer (Target: +20% coverage)

**Priority Packages**:
1. **internal/handlers/v1** (currently 9.7%)
   - Test all HTTP handlers
   - Test request validation
   - Test response formatting
   - Test error responses

2. **internal/handlers/v1/empregabilidade** (currently 0%)
   - Complete handler test suite

**Estimated Effort**: 4-5 days
**Approach**: Use httptest, create service mocks

### Phase 4: Repository Layer (Target: +10% coverage)

**Priority Packages**:
1. **internal/repository** (currently 0%)
2. **internal/repository/empregabilidade** (currently 0%)

**Estimated Effort**: 3-4 days
**Approach**: Use testcontainers for database tests

### Phase 5: Integration Tests (Target: +7% coverage)

**Focus Areas**:
1. End-to-end API tests
2. Database integration tests
3. External service integration tests

**Estimated Effort**: 2-3 days

## Recommendations

### Immediate Actions (To reach 60%)

1. **Refactor for Testability**
   - Introduce interfaces for services and repositories
   - Implement dependency injection
   - Extract complex functions into testable units

2. **Add Test Infrastructure**
   - Set up testcontainers for Postgres
   - Configure mocking framework (testify/mock)
   - Create test fixtures and helpers

3. **Prioritize High-Value Tests**
   - Focus on business logic (services)
   - Test critical paths (authentication, authorization)
   - Test error handling

4. **Incremental Approach**
   - Set coverage targets by package (e.g., services: 70%, handlers: 60%)
   - Enforce coverage in PR quality gates
   - Add tests for new code (prevent regression)

### Long-Term Improvements

1. **CI/CD Integration**
   - Add coverage reporting in PRs
   - Enforce minimum coverage thresholds
   - Track coverage trends over time

2. **Test Quality**
   - Write meaningful tests (not just for coverage numbers)
   - Test edge cases and error paths
   - Avoid brittle tests

3. **Documentation**
   - Document test patterns and best practices
   - Create testing guidelines for contributors
   - Maintain test coverage roadmap

## E2E Test Improvements

### Current E2E Tests
- Basic health check script (shell)
- Minimal Go E2E tests

### Added E2E Tests
- **tests/e2e/api_test.go**: Comprehensive API endpoint testing
  - Health endpoint structure validation
  - CORS header validation
  - Response time validation
  - Swagger documentation validation
  - Multiple endpoint smoke tests

### Recommended E2E Additions

1. **Authentication Flow Tests**
   - JWT token validation
   - Role-based access control
   - Unauthorized access scenarios

2. **Business Logic Tests**
   - Course creation/update flow
   - Enrollment flow
   - MEI proposal flow
   - Empregabilidade features

3. **Data Validation Tests**
   - Input validation
   - Error response formats
   - Data persistence verification

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Run specific package
go test ./internal/auth -v

# Run with race detection
go test -race ./...

# Run E2E tests
TEST_BASE_URL=http://localhost:8080 go test ./tests/e2e -v
```

## Conclusion

Phase 1 establishes a foundation for improved test coverage with high-quality tests for critical components (auth, middlewares). Reaching 60% coverage will require:

1. **Architectural changes** to support better testability
2. **Investment in test infrastructure** (mocking, test containers)
3. **Systematic approach** to testing each layer (repository → service → handler)
4. **Team commitment** to writing and maintaining tests

**Estimated total effort to 60%**: 12-16 days of focused work

**Key Metrics**:
- Coverage increased: 6.2% → 8.3% (+2.1% in Phase 1)
- Tests added: 120+ new test cases
- Files added: 6 new test files
- Critical gaps filled: Auth (94.6%), Middlewares (74.5%)
