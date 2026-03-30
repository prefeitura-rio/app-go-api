# Integration and Stress Test Suite

This directory contains comprehensive integration and stress tests for the app-go-api system.

## Test Categories

### 1. Enrollment Workflow Tests (`enrollment_workflow_test.go`)
Tests complete enrollment workflows with multiple service interactions:

- **Complete Flow**: CreateCurso → CreateInscricao → UpdateStatus → SendEmail (1 test)
- **Multiple Enrollments**: Testing concurrent enrollments to same course (1 test)
- **Duplicate Prevention**: Testing duplicate enrollment detection (1 test)
- **Status Transitions**: Testing valid state transitions (1 test)
- **Bulk Operations**: Testing bulk status updates (1 test)
- **Auto-Approval**: Testing auto-approval workflow (1 test)
- **Schedule Changes**: Testing schedule change workflow (1 test)
- **Validation**: Testing enrollment period and course status validation (2 tests)
- **Data Consistency**: Testing data integrity across operations (1 test)

**Total**: 10 tests

### 2. MEI Proposal Workflow Tests (`mei_workflow_test.go`)
Tests MEI opportunity and proposal workflows:

- **Complete Flow**: CreateOportunidade → CreateProposta → Validate CNAE → UpdateStatus → Notify (1 test)
- **Ownership Validation**: Testing MEI ownership validation (1 test)
- **Duplicate Prevention**: Testing duplicate proposal detection (1 test)
- **Status Transitions**: Testing valid state transitions (1 test)
- **Bulk Operations**: Testing bulk status updates (1 test)
- **Field Updates**: Testing proposal field updates (1 test)
- **Validation**: Testing inactive opportunity validation (1 test)
- **List Operations**: Testing listing by opportunity and by CNPJ (2 tests)
- **Data Consistency**: Testing data integrity (1 test)

**Total**: 10 tests

### 3. Job Application Workflow Tests (`job_workflow_test.go`)
Tests job posting and application workflows:

- **Complete Flow**: CreateVaga → CreateCandidatura → UpdateEtapa → Notify (1 test)
- **Concurrent Applications**: Testing concurrent applications to same job (1 test)
- **Duplicate Prevention**: Testing duplicate application detection (1 test)
- **Status Transitions**: Testing valid state transitions (1 test)
- **Etapa Updates**: Testing application stage updates (1 test)
- **Validation**: Testing expired job validation (1 test)
- **Curriculum Snapshot**: Testing curriculum snapshot creation (1 test)
- **Bulk Operations**: Testing bulk status updates (1 test)
- **Status Counting**: Testing status aggregation (1 test)
- **Data Consistency**: Testing data integrity (1 test)

**Total**: 10 tests

### 4. Curriculum Building Tests (`curriculo_workflow_test.go`)
Tests curriculum building and management:

- **Complete Flow**: Creating all curriculum sections in sequence (1 test)
- **ReplaceAll Formacoes**: Testing bulk replace for formacoes (1 test)
- **ReplaceAll Experiencias**: Testing bulk replace for experiencias (1 test)
- **Validation**: Testing validation on create (1 test)
- **CRUD Operations**: Testing update and delete (1 test)

**Total**: 5 tests

### 5. Stress Tests (`stress_test.go`)
Load and performance testing:

#### Concurrent Operations (3 tests)
- **100 Simultaneous Enrollments**: Race conditions, capacity checks
- **Capacity Enforcement**: Testing vacancy limits under load
- **100 Concurrent Job Applications**: Testing application processing under load

#### Bulk Operations (3 tests)
- **1000+ Enrollment Import**: Testing bulk import performance
- **500+ Status Updates**: Testing bulk update performance
- **100+ Curriculum Items**: Testing ReplaceAll with many items

#### Performance Tests (4 tests)
- **Memory Usage**: Testing memory under high concurrent load
- **Large Result Sets**: Testing pagination performance with 1000+ records
- **High-Frequency Updates**: Testing rapid status changes (50 updates)
- **Connection Pool**: Testing 200+ concurrent database connections

**Total**: 10 tests

### 6. Error Recovery Tests (`error_recovery_test.go`)
Testing system resilience and error handling:

#### Error Scenarios (10 tests)
- **Invalid IDs**: Testing handling of non-existent resources
- **Invalid Data**: Testing validation of required fields and formats
- **Transaction Rollback**: Testing rollback on partial failures
- **Status Validation**: Testing invalid state transitions
- **Duplicate Keys**: Testing unique constraint violations
- **Null Pointers**: Testing null safety
- **Concurrent Updates**: Testing concurrent modification conflicts
- **Foreign Keys**: Testing referential integrity violations
- **Context Cancellation**: Testing context timeout handling
- **Data Integrity**: Testing integrity after errors

#### Resilience Tests (5 tests)
- **Partial Failures**: Testing bulk operation partial failures
- **Email Failures**: Testing graceful degradation
- **System Recovery**: Testing recovery after multiple errors
- **Transaction Safety**: Testing ACID properties
- **Consistency**: Testing eventual consistency

**Total**: 15 tests

## Test Statistics

### Total Tests by Phase

| Phase | Tests | Description |
|-------|-------|-------------|
| Phase 1: Enrollment Workflow | 10 | Complete enrollment integration tests |
| Phase 2: MEI Workflow | 10 | MEI proposal integration tests |
| Phase 3: Job Workflow | 10 | Job application integration tests |
| Phase 4: Curriculum Workflow | 5 | Curriculum building integration tests |
| Phase 5: Stress Tests | 10 | Load and performance tests |
| Phase 6: Error Recovery | 15 | Resilience and error handling tests |

**Grand Total**: **60 integration and stress tests**

### Coverage by Component

- **Curso Service**: Fully tested
- **Inscricao Service**: Fully tested
- **PropostaMEI Service**: Fully tested
- **Vaga Service**: Fully tested
- **Candidatura Service**: Fully tested
- **Curriculo Service**: Fully tested
- **Email Notification**: Tested (graceful degradation)
- **CNAE Validation**: Tested (mocked)

## Running the Tests

### Prerequisites

1. Database connection configured
2. Environment variable `RUN_INTEGRATION_TESTS=1` or `DATABASE_URL` set

### Run All Integration Tests

```bash
just test-integration
```

### Run Specific Test Suites

```bash
# Enrollment workflow tests
go test -v ./tests/integration -run TestEnrollmentWorkflow

# MEI workflow tests
go test -v ./tests/integration -run TestMEIWorkflow

# Job workflow tests
go test -v ./tests/integration -run TestJobWorkflow

# Curriculum workflow tests
go test -v ./tests/integration -run TestCurriculoWorkflow

# Stress tests
go test -v ./tests/integration -run TestStress

# Error recovery tests
go test -v ./tests/integration -run TestErrorRecovery
```

### Run with Database

```bash
# Set environment variable
export RUN_INTEGRATION_TESTS=1
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"

# Run tests
go test -v ./tests/integration
```

### Performance Benchmarking

```bash
# Run stress tests with timing
go test -v -timeout 30m ./tests/integration -run TestStress
```

## Test Design Principles

### Integration Test Characteristics

1. **Real Services**: Tests use real service implementations with real repositories
2. **Database**: Tests run against real database (requires setup)
3. **Mocked External Dependencies**: Email service, RMI client are mocked
4. **Transaction Isolation**: Each test cleans up after itself
5. **Realistic Scenarios**: Tests simulate real user workflows

### Stress Test Characteristics

1. **Concurrent Operations**: Uses goroutines to simulate concurrent users
2. **Performance Metrics**: Measures throughput and latency
3. **Resource Usage**: Tests memory and connection pool behavior
4. **Scalability**: Tests system behavior under increasing load
5. **Race Detection**: Can be run with `-race` flag

### Error Recovery Test Characteristics

1. **Failure Scenarios**: Tests various failure modes
2. **Graceful Degradation**: Validates system continues operating
3. **Data Integrity**: Ensures data remains consistent after errors
4. **Transaction Safety**: Tests ACID properties
5. **System Resilience**: Tests recovery capabilities

## Expected Results

### Success Criteria

- All integration tests pass ✓
- Stress tests complete without panics ✓
- Error recovery tests validate graceful handling ✓
- No data corruption under load ✓
- Performance meets acceptable thresholds ✓

### Performance Benchmarks

- **Concurrent Enrollments**: 100 enrollments in < 10s
- **Bulk Import**: 1000 enrollments in < 30s
- **Bulk Update**: 500 updates in < 5s
- **Concurrent Applications**: 100 applications in < 10s
- **Database Connections**: 200 concurrent reads in < 5s

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Ensure `DATABASE_URL` is set correctly
   - Verify database is running and accessible

2. **Tests Timeout**
   - Increase timeout: `go test -timeout 30m`
   - Check database performance

3. **Race Conditions**
   - Run with race detector: `go test -race`
   - Review concurrent test logic

4. **Foreign Key Violations**
   - Ensure test data setup creates all required dependencies
   - Check test cleanup procedures

## Future Enhancements

1. **Cache Testing**: Add cache performance and invalidation tests
2. **Typesense Integration**: Add search integration tests
3. **RMI Integration**: Add real RMI integration tests (when available)
4. **Retry Logic**: Add tests for exponential backoff
5. **Circuit Breakers**: Add tests for circuit breaker patterns
6. **Monitoring**: Add observability and metrics tests

## Contributing

When adding new integration tests:

1. Follow existing test patterns
2. Use helper functions from `test_helpers.go`
3. Clean up test data after each test
4. Add meaningful assertions
5. Document expected behavior
6. Update this README
