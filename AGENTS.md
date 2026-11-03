# Agent Instructions for app-go-api

## What is app-go-api?

app-go-api is a RESTful API service built in Go for the Prefeitura do Rio (Rio de Janeiro City Hall) platform. It manages courses (cursos), job opportunities (empregos), MEI (Microempreendedor Individual) opportunities, course enrollments (inscrições), and a full employability (empregabilidade) module covering resumes, vacancies, and candidacies.

The service sits behind Istio in the Kubernetes cluster. Authentication is handled externally by Istio/Keycloak — the API trusts the `X-Auth-Request-Token` header injected by Istio and decodes the JWT payload without re-validating the signature. Role resolution is delegated to the Heimdall service, and fine-grained authorization on MEI proposals uses Cerbos.

The service is also consumed by `superapp` (citizen-facing frontend) and by internal admin tools. It synchronizes data from the `app-rmi` service via background workers.

## Architecture

### Technology Stack

- **Language**: Go 1.27
- **Framework**: Gin (gin-gonic/gin v1.10)
- **Database**: PostgreSQL 16+ via GORM (gorm.io/gorm v1.30) with `gorm.io/driver/postgres`
- **Cache**: Redis 7+ via `redis/go-redis/v9`
- **Search**: Typesense (`typesense-go/v3`) for full-text search on courses and jobs
- **Migrations**: Goose (`pressly/goose`) for SQL migration files in `internal/db/migrations/`
- **Observability**: OpenTelemetry tracing via OTLP/gRPC to SigNoz; GORM and Redis are both instrumented
- **Configuration**: Viper (`spf13/viper`) with `.env` file and environment variable fallback
- **API Docs**: Swaggo (`swaggo/swag`) auto-generated from code annotations
- **Auth**: Keycloak service account token manager (`internal/auth/`) for calling RMI on behalf of the service; Cerbos PDP for policy-based authorization
- **External clients**: `clients/rmi_client.go` (app-rmi), `clients/data_relay_client.go` (email notifications)

### Project Structure

```
app-go-api/
├── cmd/
│   ├── server/              # Application entry point (main.go)
│   └── tools/
│       └── import_courses/  # One-off import utility
├── internal/
│   ├── auth/                # Keycloak service account token manager
│   ├── authorization/       # Cerbos policy checker
│   ├── cache/               # Redis cache wrappers (legal entities, reference data, courses)
│   ├── clients/             # HTTP clients for external services (RMI, Data Relay)
│   ├── config/              # Configuration management (Viper singleton)
│   ├── db/
│   │   └── migrations/      # SQL migration files (Goose)
│   ├── handlers/
│   │   ├── common/          # Shared handler utilities
│   │   └── v1/              # HTTP handlers for API v1
│   │       └── empregabilidade/ # Employability-specific handlers
│   ├── jobs/                # Async job processor (bulk enrollment imports)
│   ├── middlewares/         # Gin middleware (CORS, timeout, user context, Swagger)
│   ├── models/              # GORM models and domain types
│   │   └── empregabilidade/ # Employability-specific models
│   ├── observability/       # OpenTelemetry tracer initialization
│   ├── repository/          # GORM data access layer
│   │   └── empregabilidade/ # Employability-specific repositories
│   ├── router/              # Gin router setup and dependency wiring
│   ├── services/            # Business logic layer
│   │   └── empregabilidade/ # Employability-specific services
│   ├── utils/               # Shared utilities
│   └── workers/             # Background workers (orgao sync, citizen sync)
├── docs/                    # Swagger (auto-generated: docs.go, swagger.*) + handwritten docs (e.g. observability.md)
├── k8s/
│   ├── staging/             # Argo Rollout (blue-green) + Services
│   └── prod/                # Argo Rollout (canary) + AnalysisTemplate + Services
├── scripts/                 # Shell scripts (extract-coverage.sh, etc.)
├── tests/
│   └── e2e/                 # E2E health check suite (health_check.sh)
├── .env.example             # All supported environment variables with defaults
├── .golangci.yml            # Linter configuration
├── Dockerfile               # Multi-stage Docker build
├── docker-compose.yml       # Local dev (PostgreSQL + Redis)
├── justfile                 # Task runner commands
└── flake.nix                # Nix development environment
```

### Architecture Patterns

**Layered Architecture**: Handlers receive HTTP requests and call services. Services contain business logic and call repositories. Repositories interact with GORM/PostgreSQL or Redis.

**Dependency Injection via Constructor**: All dependencies are wired in `internal/router/router.go`. There is no DI framework — the router instantiates repositories, passes them to services, passes services to handlers. This is the single wiring point.

**Interface-Driven Design**: Services depend on repository interfaces (defined in `internal/services/interfaces.go`), not concrete structs. This makes unit testing possible without a real database.

**Background Workers**: `workers/orgao_sync_worker.go` and `workers/citizen_sync_worker.go` run as goroutines started in `main.go` and `router.go` respectively. They use Redis-based distributed locking so only one pod executes the sync at a time.

**User Context via Middleware**: `middlewares/ExtractUserContext` decodes the JWT from `X-Auth-Request-Token` (Istio) or `Authorization`, optionally queries Heimdall for roles, and injects CPF, role, user ID, name, and email into the Gin context. Downstream handlers read these via `middlewares.GetUserCPF(c)`, `middlewares.IsAdmin(c)`, etc.

**Configuration Singleton**: `config.Load()` reads from a `.env` file and environment variables. `config.Get()` returns the singleton; it is safe for concurrent reads.

**Redis Caching**: Reference data (categories, accessibility types, education levels) uses a 1-hour TTL. Course data uses a 5-minute TTL. Legal entity lookups (CNPJ → LegalEntity) use a 30-minute TTL.

## Development Practices

### Test-Driven Development (TDD)

**TDD is MANDATORY for all code changes.** Follow the Red-Green-Refactor cycle:

1. **Red**: Write a failing test first, then run it to confirm it fails:
   ```bash
   go test -v -run TestYourNewTest ./internal/handlers/v1/...
   ```

2. **Green**: Write the minimal production code to make the test pass:
   ```bash
   go test -v -run TestYourNewTest ./internal/handlers/v1/...
   ```

3. **Refactor**: Clean up while keeping tests green:
   ```bash
   go test -race ./...
   ```

**TDD Guidelines**:
- Write tests BEFORE writing production code
- No production code without a corresponding test
- Each test should test one behavior
- Use table-driven tests for multiple input scenarios
- Mock external dependencies using the `testify/mock` library or manual structs implementing the relevant interface

### Development Workflow

High-level git/PR steps are in `README.md`. For agents: branch from `main`, write the failing test first, implement, then `just ci` before opening the PR.

### Code Quality Standards

**Required before creating a PR**:
- All tests pass with `-race` flag (`just test`)
- Coverage does not decrease beyond 0.1% tolerance
- Code is formatted (`just fmt`)
- Linter passes (`just lint`)
- No CRITICAL or HIGH security vulnerabilities (Trivy)
- TDD approach followed (test written before implementation)

**Active linters** (from `.golangci.yml`):
- `errcheck` — unchecked errors (disabled in `_test.go` files)
- `govet` — suspicious constructs (atomic, bools, buildtag, errorsas, nilfunc, printf)
- `gofmt` / `goimports` — formatting
- `sqlclosecheck` / `rowserrcheck` — SQL resource leaks

**Naming Conventions**:
- Handlers: `XxxHandler` struct with methods `Create`, `List`, `GetByID`, `Update`, `Delete`
- Services: `XxxService` interface + constructor `NewXxxService(repo XxxRepositoryInterface) *xxxService`
- Repositories: concrete struct `xxxRepository` constructed by `NewXxxRepository(db *gorm.DB) *xxxRepository`; implement the matching interface defined in `internal/services/interfaces.go`
- Tests: `TestFunctionName_Scenario_ExpectedResult` (e.g., `TestCourseHandler_Create_Success`)

**Testing Patterns**:

Handler tests use `httptest.NewRecorder()` and a test-mode Gin engine. Dependencies are mocked either with `testify/mock` (preferred) or with manual structs implementing the service interface:

```go
// Using testify/mock (preferred for services with many methods)
type MockCursoService struct {
    mock.Mock
}

func (m *MockCursoService) Create(ctx context.Context, curso *models.Curso) (int, error) {
    args := m.Called(ctx, curso)
    return args.Int(0), args.Error(1)
}

func TestCourseHandler_Create_Success(t *testing.T) {
    gin.SetMode(gin.TestMode)
    mockService := new(MockCursoService)
    mockService.On("Create", mock.Anything, mock.AnythingOfType("*models.Curso")).Return(1, nil)

    handler := NewCourseHandler(mockService, nil, nil)
    r := gin.New()
    r.POST("/api/v1/courses", handler.Create)

    body := `{"titulo": "Test Course"}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/courses", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusCreated, w.Code)
    mockService.AssertExpectations(t)
}
```

```go
// Manual mock struct (used in acessibilidade, categoria, escolaridade tests)
type mockRepoForHandler struct {
    entity    *models.Acessibilidade
    createID  int
    listItems []*models.Acessibilidade
    listTotal int
}

func (m *mockRepoForHandler) Create(_ context.Context, _ *models.Acessibilidade) (int, error) {
    return m.createID, nil
}
// ... implement all interface methods
```

Service and repository unit tests follow table-driven patterns:

```go
func TestSomeService_Method(t *testing.T) {
    tests := []struct {
        name     string
        input    SomeInput
        wantErr  bool
        wantID   int
    }{
        {"success case", validInput, false, 1},
        {"not found", badInput, true, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            // Act
            // Assert
        })
    }
}
```

**Integration tests** (repository layer only) are gated by the `RUN_REPOSITORY_INTEGRATION=1` environment variable and require `DATABASE_URL`. They are never run in CI PRs — only locally:
```bash
RUN_REPOSITORY_INTEGRATION=1 DATABASE_URL="postgres://..." go test ./internal/repository/...
```

## GitOps / CI (agent notes)

Branch strategy, PR quality gate, staging blue-green, and production canary are documented in `README.md` and live under `.github/workflows/`. Agent-specific details:

- Coverage baseline lives on `gh-pages` as `coverage-baseline.txt`; PRs fail if coverage drops more than 0.1%.
- Staging: namespace `go`, ArgoCD app `go`, blue-green (`go` / `go-preview`), Infisical `go-secrets`.
- Production: same namespace/app; canary via Istio VirtualService `go-virtual-service` with analysis template `go-success-rate`.
- Canary steps: `10% → 2m → 20% → 3m → 40% → 5m → 60% → 5m → 80% → 5m → 100%` (Prometheus analysis from step 2; NaN-safe queries on `istio_requests_total`).
- Secrets/vars: `GCP_CREDENTIALS_JSON`, `GCP_PROJECT_ID`, optional `DISCORD_WEBHOOK_URL`, auto `GITHUB_TOKEN`.

## Working with the Codebase

### Development Environment Setup

**Use nix and direnv for a reproducible development environment.**

This project provides a `flake.nix` and `.envrc` to ensure every developer and CI run uses the exact same versions of all tools.

#### Prerequisites

Install nix (if not already installed):
```bash
# Multi-user installation (recommended)
sh <(curl -L https://nixos.org/nix/install) --daemon

# Or single-user installation
sh <(curl -L https://nixos.org/nix/install) --no-daemon
```

Enable flakes in your nix configuration:
```bash
mkdir -p ~/.config/nix
echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf
```

Install direnv (if not already installed):
```bash
# macOS
brew install direnv

# Linux — follow https://direnv.net/docs/installation.html
```

Add the direnv hook to your shell:
```bash
# bash
echo 'eval "$(direnv hook bash)"' >> ~/.bashrc

# zsh
echo 'eval "$(direnv hook zsh)"' >> ~/.zshrc

# fish
echo 'direnv hook fish | source' >> ~/.config/fish/config.fish
```

#### Activating the Environment

```bash
direnv allow   # one-time approval; re-run after flake.nix changes
```

After `direnv allow`, every `cd` into the project directory automatically loads the nix shell. Leaving the directory unloads it. No manual version management is needed.

#### What the Nix Shell Provides

The `flake.nix` `devShells.default` includes:

| Tool | Purpose |
|---|---|
| `go_1_27` | Go compiler and toolchain |
| `gopls` | Go language server (editor support) |
| `gotools` | `goimports` and other Go utilities |
| `golangci-lint` | Linter (matches CI version) |
| `delve` | Go debugger |
| `air` | Hot reload for local development |
| `postgresql_16` | PostgreSQL client tools |
| `goose` | Database migration runner |
| `docker` / `docker-compose` | Container runtime |
| `jq` | JSON processing |
| `gnumake` / `just` | Task runners |

`swag` (Swagger doc generator) is installed on first shell entry via the `shellHook` if not already present.

#### Benefits

- **Reproducibility**: All developers and CI use identical tool versions
- **Isolation**: Project tools do not conflict with system-level installations
- **Fast onboarding**: New developers get a working environment with two commands (`direnv allow`)
- **CI parity**: The same `flake.nix` can be referenced in CI to match the local environment

#### Troubleshooting

**direnv not loading automatically**:
```bash
direnv allow
```

**Nix shell not activating**:
```bash
# Reload the environment
direnv reload

# Or enter the nix shell manually
nix develop
```

**Stale or broken dependencies**:
```bash
nix-collect-garbage -d
direnv reload
```

---

### Running Locally

See `README.md` for setup (`just setup`, `.env`, `docker-compose`, `just migrate-up`, `just dev` / `just run`).

### Environment Variables Reference

Key variables (see `.env.example` for the full list):

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | Raw environment string, compared as-is by `cfg.App.Environment` / `IsDevelopment()` / `IsProduction()` / `IsTest()` — see note below, the deployed values are **not** `staging`/`production` |
| `RBAC_ENABLED` | `true` | Enables course/vaga RBAC middlewares (`CourseAuthorization`, `VagaAuthorization`, ownership checks, `ExtractSecretariaOrgaoIDs`). Independent of `APP_ENV`. Set `false` in production until RBAC is ready there. |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `app_go_api` | PostgreSQL database name |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `SERVER_PORT` | `8080` | HTTP server port |
| `RUN_MIGRATIONS` | `false` | Auto-run GORM AutoMigrate on startup |
| `TYPESENSE_HOST` | `localhost` | Typesense search engine host |
| `TYPESENSE_API_KEY` | `` | Typesense API key |
| `RMI_BASE_URL` | `` | URL for the app-rmi service |
| `HEIMDALL_BASE_URL` | `` | URL for Heimdall role resolution |
| `KEYCLOAK_URL` | `` | Keycloak URL (required for service account) |
| `KEYCLOAK_CLIENT_ID` | `` | Service account client ID |
| `KEYCLOAK_CLIENT_SECRET` | `` | Service account client secret |
| `CERBOS_ENABLED` | `false` | Enable Cerbos policy-based authorization |
| `ORGAO_SYNC_ENABLED` | `true` | Enable background org sync worker |
| `CITIZEN_SYNC_ENABLED` | `true` | Enable background citizen sync worker |
| `TRACING_ENABLED` | `false` | Enable OpenTelemetry tracing |

> **Real `APP_ENV` values per deployed environment** (confirmed via `kubectl logs -n go <pod>`, which prints `Ambiente: <value>` on startup when `APP_DEBUG=true`; the value is not set in the committed `k8s/*/resources.yaml` — it comes from the Infisical-managed `go-secrets` SecretRef):
>
> - **Local dev**: `development` (the `.env.example` default)
> - **Staging**: `development` — **not** `staging`. The `go-secrets` for staging sets `APP_ENV=development`, so any code that branches on `cfg.App.Environment == "development"` (or `IsDevelopment()`) is also active in staging. Prefer dedicated env vars (e.g. `RBAC_ENABLED`) for feature flags instead of gating on `APP_ENV`.
> - **Production**: `prod` — **not** `production`. `AppSettings.IsProduction()` compares against both `"prod"` and `"production"`.
>
> Before writing a new environment-gated feature flag, prefer an explicit env var (like `RBAC_ENABLED`) over branching on `APP_ENV`. If you must branch on environment, check the real deployed value first (`kubectl logs -n go <pod> | grep Ambiente` in both the staging and production contexts) instead of assuming `staging`/`production` are used literally.

### Running Tests

Prefer `just test`, `just test-coverage`, `just test-pkg <pkg>`, and `just test-e2e` (see `README.md` / `justfile`). Coverage HTML is written under `coverage/` (gitignored).

```bash
# Named unit test
go test -v -race -run TestCourseHandler_Create_Success ./internal/handlers/v1/...

# Repository integration (local only; never in CI PRs)
DATABASE_URL="postgres://postgres:postgres@localhost:5432/app_db?sslmode=disable" \
  RUN_REPOSITORY_INTEGRATION=1 \
  go test -v ./internal/repository/...
```

### Common Development Tasks

#### Adding a New Endpoint (TDD Approach)

Follow this exact order:

**Step 1 — Write the handler test first**

```go
// internal/handlers/v1/feature_handler_test.go
package v1_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
)

func TestFeatureHandler_GetByID_Success(t *testing.T) {
    gin.SetMode(gin.TestMode)
    // Mock the service
    mockSvc := &mockFeatureService{entity: &models.Feature{ID: 1}}
    handler := v1.NewFeatureHandler(mockSvc)

    r := gin.New()
    r.GET("/api/v1/features/:id", handler.GetByID)

    req := httptest.NewRequest(http.MethodGet, "/api/v1/features/1", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}
```

**Step 2 — Confirm the test fails**

```bash
go test -v -run TestFeatureHandler_GetByID_Success ./internal/handlers/v1/...
```

**Step 3 — Write the service interface and repository interface** in `internal/services/interfaces.go`

**Step 4 — Implement handler, service, and repository** (in that order, with tests at each layer)

**Step 5 — Register the route** in `internal/router/router.go`

**Step 6 — Generate Swagger docs**

```bash
just swagger
```

**Step 7 — Run quality gate locally**

```bash
just ci  # fmt + lint + test
```

#### Adding a Database Migration

```bash
# Create a new migration file (Goose SQL format)
just migrate-create add_feature_column

# Apply migrations
just migrate-up

# Roll back last migration
just migrate-down
```

Migration files live in `internal/db/migrations/` as `YYYYMMDDHHMMSS_name.sql`.

#### Updating API Documentation

Swagger annotations live as comments above handler functions. After adding or modifying endpoints, regenerate docs:

```bash
just swagger
```

Swagger outputs (`docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`) are committed and must be regenerated before building the Docker image (CI does this automatically). Handwritten docs under `docs/` (e.g. `observability.md`) are edited normally.

#### Updating Dependencies

```bash
just deps-update   # go get -u ./... && go mod tidy
just test          # Verify nothing breaks
```

### Understanding the Request Lifecycle

1. Gin router matches the path and calls middlewares in order:
   - `gin.Recovery()` — panic recovery
   - `gin.Logger()` — request logging
   - `otelgin.Middleware(...)` — trace propagation (if TRACING_ENABLED)
   - `middlewares.TimeoutMiddleware(...)` — per-request deadline
   - `middlewares.CorsMiddleware()` — CORS headers
   - `middlewares.ExtractUserContext(heimdallBaseURL)` — JWT decode + Heimdall role lookup
2. Handler receives `*gin.Context`, validates input, calls service method
3. Service applies business logic, calls repository method(s)
4. Repository executes GORM query against PostgreSQL
5. Handler writes JSON response

### Debugging

**Check logs in the cluster**:
```bash
# Staging
kubectl config use-context <staging-context>
kubectl logs -n go -l app=go --tail=100 -f

# Production
kubectl config use-context <prod-context>
kubectl logs -n go -l app=go --tail=100 -f
```

**Check deployment status**:
```bash
kubectl get pods -n go
kubectl describe pod <pod-name> -n go
kubectl argo rollouts get rollout go -n go
```

**Check ArgoCD sync status**:
```bash
kubectl get application go -n argocd -o jsonpath='{.status.sync.status}'
kubectl get application go -n argocd -o jsonpath='{.status.health.status}'
```

**Check recent CI runs**:
```bash
gh run list --limit 10
gh run view <run-id> --log
```

**Check PR quality gate**:
```bash
gh pr view <pr-number>
gh pr checks <pr-number>
```

## Best Practices

1. **Always follow TDD**: Write the test, watch it fail, write minimal code, watch it pass, refactor
2. **Define interfaces for testability**: If you add a new external dependency, add an interface in `internal/services/interfaces.go`
3. **Mock at the service boundary in handler tests**: Handlers should only know about service interfaces, never concrete types
4. **Test error paths explicitly**: Do not only test the happy path — also test 404, 400, 409, 500 scenarios
5. **Use table-driven tests**: Multiple scenarios in one function with `t.Run`
6. **Keep handlers thin**: Input parsing and validation in handlers, all business logic in services
7. **Use context propagation**: Pass `ctx context.Context` through every layer; use `c.Request.Context()` in handlers
8. **Validate input at the handler boundary**: Bind and validate JSON before calling the service
9. **Return structured errors**: Use `gin.H{"error": "..."}` with appropriate HTTP status codes
10. **Do not hard-code configuration**: Use `cfg` (passed from `config.Load()`) for all tuneable values
11. **Handle Redis cache misses gracefully**: Cache is a performance optimization; the service must work if Redis is unavailable
12. **Use user context helpers**: Use `middlewares.GetUserCPF(c)`, `middlewares.IsAdmin(c)`, etc. — never parse headers manually in handlers
13. **Use nix + direnv**: Always develop within the nix shell (`direnv allow`) for reproducibility and CI parity

## Common Pitfalls to Avoid

- Writing production code before a failing test exists (violates TDD)
- Not running tests with `-race` flag (masks data races)
- Ignoring or suppressing errors instead of returning them
- Not calling `c.Abort()` after `c.JSON(...)` in middleware that should stop the chain
- Committing stale Swagger outputs without regenerating with `swag init` / `just swagger`
- Editing `docs/docs.go` / `docs/swagger.*` by hand (they are auto-generated)
- Adding new fields to GORM models without a corresponding migration
- Forgetting to register a new route in `internal/router/router.go`
- Hardcoding environment-specific values (URLs, ports, timeouts) in application code
- Ignoring lint warnings and relying on CI to catch them (run `just lint` locally first)
- Not checking the `gh-pages` branch coverage baseline when estimating test effort
- Performing kubectl write operations in CI without read-only intent (use `kubectl patch` only in deploy workflows)

## Troubleshooting CI/CD

For lint, coverage, security, and local test failures, see `README.md` (§ Troubleshooting CI/CD). Cluster/ops notes:

**Test race flakes**:
```bash
go test -race -count=3 ./internal/...
```

**Rollout stuck in staging**:
```bash
kubectl argo rollouts get rollout go -n go
kubectl get pods -n go
kubectl describe pod <pod-name> -n go
```

**ArgoCD not syncing**:
```bash
kubectl get application go -n argocd -o yaml
```

**Canary analysis NaN in production**:
The `go-success-rate` AnalysisTemplate uses `> 0` filters on the denominator and `failureLimit: 5` to tolerate low-traffic gaps.

## Resources

- **Coverage Baseline**: `gh-pages` branch → `coverage-baseline.txt`
- **CI/CD Workflows**: `.github/workflows/`
- **API Documentation**: `/docs/swagger/index.html` when running locally, or at the deployed `/docs/` path
- **Environment Variables**: `.env.example`
- **Migration Files**: `internal/db/migrations/`
- **OpenTelemetry**: `docs/observability.md`
- **Golden Standard Reference**: sibling repo `app-rmi`
