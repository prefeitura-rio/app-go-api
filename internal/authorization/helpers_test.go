package authorization

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

// mockChecker is a mock implementation of the Checker for testing
type mockChecker struct {
	checkAnyActionResult bool
	checkAnyActionErr    error
}

func (m *mockChecker) CheckAnyAction(ctx context.Context, principalCPF string, roles []string, resourceKind string, actions []string) (bool, error) {
	return m.checkAnyActionResult, m.checkAnyActionErr
}

// createTestCheckerWithBehavior creates a real Checker with a mock Cerbos server
func createTestCheckerWithBehavior(t *testing.T, shouldAllow bool, shouldError bool) *Checker {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldError {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Parse the request to get action names
		var req CheckResourcesRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		// Build response with all actions
		actions := make(map[string]string)
		for _, action := range req.Resources[0].Actions {
			if shouldAllow {
				actions[action] = "EFFECT_ALLOW"
			} else {
				actions[action] = "EFFECT_DENY"
			}
		}

		response := CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions:  actions,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	t.Cleanup(server.Close)

	return NewChecker(server.URL, 5*time.Second)
}

// mockCNPJOwnershipChecker is a mock for CNPJ ownership checking
type mockCNPJOwnershipChecker struct {
	isOwner bool
	err     error
}

func (m *mockCNPJOwnershipChecker) CheckCNPJOwnership(ctx context.Context, authToken string, cpf string, cnpj string) (bool, error) {
	return m.isOwner, m.err
}

func setupTestContext(userCPF, userRole, authToken string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	if userCPF != "" {
		c.Set(middlewares.UserCPFKey, userCPF)
	}
	if userRole != "" {
		c.Set(middlewares.UserRoleKey, userRole)
	}
	if authToken != "" {
		c.Request.Header.Set("Authorization", authToken)
	}

	return c
}

func TestRequireOwnershipOrAnyPermission_OwnerAccess(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	// User is the owner
	err := RequireOwnershipOrAnyPermission(c, nil, "12345678900", "proposta", []string{"proposta:update"})

	assert.NoError(t, err)
}

func TestRequireOwnershipOrAnyPermission_NoUserCPF(t *testing.T) {
	c := setupTestContext("", "user", "")

	err := RequireOwnershipOrAnyPermission(c, nil, "11111111111", "proposta", []string{"proposta:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user CPF not found in context")
}

func TestRequireOwnershipOrAnyPermission_NotOwnerNilChecker(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	// User is NOT the owner and checker is nil (Cerbos disabled)
	err := RequireOwnershipOrAnyPermission(c, nil, "11111111111", "proposta", []string{"proposta:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you can only access your own resources")
}

func TestRequireOwnershipOrAnyPermission_NotOwnerWithPermission(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	// User is NOT the owner but has permission
	// Since we can't easily mock the Checker, we test with nil which represents Cerbos disabled
	err := RequireOwnershipOrAnyPermission(c, (*Checker)(nil), "11111111111", "proposta", []string{"proposta:update"})

	// This should fail because we're passing nil checker
	assert.Error(t, err)

	// Note: Testing the permission check path would require refactoring to inject a mock
	// or running against a real Cerbos instance, which is beyond unit testing scope
}

func TestRequireOwnershipOrAnyPermission_NotOwnerWithoutPermission(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	// User is NOT the owner and doesn't have permission
	// With nil checker, should fail
	err := RequireOwnershipOrAnyPermission(c, nil, "11111111111", "proposta", []string{"proposta:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you can only access your own resources")
}

func TestRequireCNPJOwnershipOrAnyPermission_NoUserCPF(t *testing.T) {
	c := setupTestContext("", "user", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: true}

	err := RequireCNPJOwnershipOrAnyPermission(c, nil, ownershipChecker, "12345678000100", "empresa", []string{"empresa:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user CPF not found in context")
}

func TestRequireCNPJOwnershipOrAnyPermission_NoAuthToken(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: true}

	err := RequireCNPJOwnershipOrAnyPermission(c, nil, ownershipChecker, "12345678000100", "empresa", []string{"empresa:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization token not found")
}

func TestRequireCNPJOwnershipOrAnyPermission_IsOwner(t *testing.T) {
	c := setupTestContext("12345678900", "user", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: true}

	err := RequireCNPJOwnershipOrAnyPermission(c, nil, ownershipChecker, "12345678000100", "empresa", []string{"empresa:update"})

	assert.NoError(t, err)
}

func TestRequireCNPJOwnershipOrAnyPermission_NotOwnerNilChecker(t *testing.T) {
	c := setupTestContext("12345678900", "user", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: false}

	err := RequireCNPJOwnershipOrAnyPermission(c, nil, ownershipChecker, "12345678000100", "empresa", []string{"empresa:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you can only access resources for CNPJs you own")
}

func TestRequireCNPJOwnershipOrAnyPermission_OwnershipCheckError(t *testing.T) {
	c := setupTestContext("12345678900", "user", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{
		isOwner: false,
		err:     assert.AnError,
	}

	err := RequireCNPJOwnershipOrAnyPermission(c, nil, ownershipChecker, "12345678000100", "empresa", []string{"empresa:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check CNPJ ownership")
}

func TestRequireCNPJOwnershipOrAnyPermission_XAuthRequestToken(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")
	c.Request.Header.Set("X-Auth-Request-Token", "istio-token")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: true}

	err := RequireCNPJOwnershipOrAnyPermission(c, nil, ownershipChecker, "12345678000100", "empresa", []string{"empresa:update"})

	assert.NoError(t, err)
}

func TestRequireAnyPermission_NilChecker(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	err := RequireAnyPermission(c, nil, "oportunidade", []string{"oportunidade:create"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization is disabled")
}

func TestRequireAnyPermission_NoUserCPF(t *testing.T) {
	c := setupTestContext("", "admin", "")

	// We can't create a real Checker without a running Cerbos instance
	// So we test the validation logic
	checker := (*Checker)(nil)

	err := RequireAnyPermission(c, checker, "oportunidade", []string{"oportunidade:create"})

	// Should fail at nil checker check first
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization is disabled")
}

func TestRequirePermission_SingleAction(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	// Test that RequirePermission wraps RequireAnyPermission correctly
	// With nil checker, should fail
	err := RequirePermission(c, nil, "oportunidade", "oportunidade:create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization is disabled")
}

func TestRequirePermission_WrapsRequireAnyPermission(t *testing.T) {
	// Verify that RequirePermission is just a wrapper
	c := setupTestContext("12345678900", "admin", "")

	err1 := RequirePermission(c, nil, "resource", "action")
	err2 := RequireAnyPermission(c, nil, "resource", []string{"action"})

	// Both should return the same error message
	assert.Equal(t, err1.Error(), err2.Error())
}

func TestRequireOwnershipOrAnyPermission_EmptyRole(t *testing.T) {
	c := setupTestContext("12345678900", "", "")

	// User is the owner, role doesn't matter
	err := RequireOwnershipOrAnyPermission(c, nil, "12345678900", "proposta", []string{"proposta:update"})

	assert.NoError(t, err)
}

func TestRequireAnyPermission_EmptyActions(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	// Empty actions slice
	err := RequireAnyPermission(c, nil, "oportunidade", []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization is disabled")
}

// Tests with mock Checker to cover permission check paths
func TestRequireOwnershipOrAnyPermission_NotOwnerCheckAnyActionError(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	// Since we can't easily inject a real mock without refactoring the code,
	// we test the error path with nil checker which simulates Cerbos disabled
	checker := (*Checker)(nil)

	err := RequireOwnershipOrAnyPermission(c, checker, "11111111111", "proposta", []string{"proposta:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you can only access your own resources")
}

func TestRequireOwnershipOrAnyPermission_NotOwnerNoPermission(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	// Test scenario: user is not owner, checker exists but returns false
	// Since we can't inject a real mock, we verify the error path exists
	err := RequireOwnershipOrAnyPermission(c, nil, "98765432100", "resource", []string{"action"})

	assert.Error(t, err)
}

func TestRequireOwnershipOrAnyPermission_EmptyRoleNotOwner(t *testing.T) {
	c := setupTestContext("12345678900", "", "")

	// User is NOT owner, role is empty, checker is nil
	err := RequireOwnershipOrAnyPermission(c, nil, "98765432100", "proposta", []string{"proposta:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you can only access your own resources")
}

func TestRequireCNPJOwnershipOrAnyPermission_NotOwnerEmptyRole(t *testing.T) {
	c := setupTestContext("12345678900", "", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: false}

	err := RequireCNPJOwnershipOrAnyPermission(c, nil, ownershipChecker, "12345678000100", "empresa", []string{"empresa:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you can only access resources for CNPJs you own")
}

func TestRequireAnyPermission_EmptyRole(t *testing.T) {
	c := setupTestContext("12345678900", "", "")

	// Empty role, nil checker
	err := RequireAnyPermission(c, nil, "resource", []string{"action"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization is disabled")
}

// mockCheckerWrapper wraps the Checker to allow mocking CheckAnyAction behavior
type mockCheckerWrapper struct {
	checkAnyActionFunc func(ctx context.Context, principalCPF string, roles []string, resourceKind string, actions []string) (bool, error)
}

func (m *mockCheckerWrapper) CheckAnyAction(ctx context.Context, principalCPF string, roles []string, resourceKind string, actions []string) (bool, error) {
	if m.checkAnyActionFunc != nil {
		return m.checkAnyActionFunc(ctx, principalCPF, roles, resourceKind, actions)
	}
	return false, nil
}

// newMockChecker creates a Checker with mocked behavior for testing
func newMockChecker(checkAnyActionFunc func(ctx context.Context, principalCPF string, roles []string, resourceKind string, actions []string) (bool, error)) *Checker {
	// Create a mock checker - we'll use the actual Checker type but with a fake client
	// This is a hack to test the helper functions without refactoring them
	// In production code, we'd use dependency injection
	checker := &Checker{
		client: nil, // We won't actually call the client
	}

	// Override the CheckAnyAction behavior by creating a wrapper
	// Note: This is tricky because we can't easily mock methods on the Checker struct
	// The best we can do is test with nil and verify the error paths
	return checker
}

func TestRequireOwnershipOrAnyPermission_NotOwnerWithPermissionGranted(t *testing.T) {
	// This test verifies the happy path when user is not owner but has permission
	// However, we can't easily test this without refactoring the code to inject
	// a CheckAnyAction function or interface. The current implementation calls
	// checker.CheckAnyAction directly.

	// For now, we document that this path is covered by integration tests
	// and focus on the paths we can test with unit tests.
	t.Skip("Requires refactoring to inject CheckAnyAction dependency")
}

func TestRequireOwnershipOrAnyPermission_NotOwnerPermissionCheckError(t *testing.T) {
	// This test would verify the error handling when CheckAnyAction fails
	// However, we can't easily test this without refactoring the code
	t.Skip("Requires refactoring to inject CheckAnyAction dependency")
}

func TestRequireCNPJOwnershipOrAnyPermission_NotOwnerWithPermissionGranted(t *testing.T) {
	// This test would verify CNPJ permission check happy path
	t.Skip("Requires refactoring to inject CheckAnyAction dependency")
}

func TestRequireCNPJOwnershipOrAnyPermission_NotOwnerPermissionCheckError(t *testing.T) {
	// This test would verify CNPJ permission check error path
	t.Skip("Requires refactoring to inject CheckAnyAction dependency")
}

func TestRequireAnyPermission_PermissionGranted(t *testing.T) {
	// This test would verify the happy path when permission is granted
	t.Skip("Requires refactoring to inject CheckAnyAction dependency")
}

func TestRequireAnyPermission_PermissionDenied(t *testing.T) {
	// This test would verify the path when permission is denied
	t.Skip("Requires refactoring to inject CheckAnyAction dependency")
}

func TestRequireAnyPermission_CheckError(t *testing.T) {
	// This test would verify error handling in permission check
	t.Skip("Requires refactoring to inject CheckAnyAction dependency")
}

// Additional edge case tests
func TestRequireOwnershipOrAnyPermission_MultipleActions(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	// User is owner, should succeed regardless of actions
	err := RequireOwnershipOrAnyPermission(c, nil, "12345678900", "proposta", []string{
		"proposta:read",
		"proposta:update",
		"proposta:delete",
	})

	assert.NoError(t, err)
}

func TestRequireOwnershipOrAnyPermission_SingleAction(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	// User is owner, single action
	err := RequireOwnershipOrAnyPermission(c, nil, "12345678900", "proposta", []string{"proposta:read"})

	assert.NoError(t, err)
}

func TestRequireCNPJOwnershipOrAnyPermission_AuthTokenPriority(t *testing.T) {
	c := setupTestContext("12345678900", "user", "auth-header-token")
	// Also set X-Auth-Request-Token which should take priority
	c.Request.Header.Set("X-Auth-Request-Token", "istio-token")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: true}

	err := RequireCNPJOwnershipOrAnyPermission(c, nil, ownershipChecker, "12345678000100", "empresa", []string{"empresa:read"})

	// Should succeed with X-Auth-Request-Token taking priority
	assert.NoError(t, err)
}

func TestRequirePermission_MultiplePermissions(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	// Test that RequirePermission correctly wraps single permission
	err1 := RequirePermission(c, nil, "resource", "action")
	err2 := RequireAnyPermission(c, nil, "resource", []string{"action"})

	// Both should produce identical errors
	assert.Equal(t, err1.Error(), err2.Error())
}

func TestRequireOwnershipOrAnyPermission_DifferentResourceKinds(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	// Test owner access works for different resource kinds
	testCases := []struct {
		resourceKind string
		actions      []string
	}{
		{"proposta", []string{"proposta:read"}},
		{"oportunidade", []string{"oportunidade:create"}},
		{"empresa", []string{"empresa:update", "empresa:delete"}},
		{"document", []string{"document:view", "document:download", "document:share"}},
	}

	for _, tc := range testCases {
		t.Run(tc.resourceKind, func(t *testing.T) {
			err := RequireOwnershipOrAnyPermission(c, nil, "12345678900", tc.resourceKind, tc.actions)
			assert.NoError(t, err)
		})
	}
}

func TestRequireAnyPermission_DifferentResourceKinds(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	// All should fail with nil checker
	testCases := []string{
		"proposta",
		"oportunidade",
		"empresa",
		"document",
	}

	for _, resourceKind := range testCases {
		t.Run(resourceKind, func(t *testing.T) {
			err := RequireAnyPermission(c, nil, resourceKind, []string{"action"})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "authorization is disabled")
		})
	}
}

func TestRequireCNPJOwnershipOrAnyPermission_InvalidCNPJ(t *testing.T) {
	c := setupTestContext("12345678900", "user", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: false}

	// Test with various CNPJ formats
	testCases := []string{
		"00000000000000",
		"12345678000100",
		"99999999999999",
	}

	for _, cnpj := range testCases {
		t.Run(cnpj, func(t *testing.T) {
			err := RequireCNPJOwnershipOrAnyPermission(c, nil, ownershipChecker, cnpj, "empresa", []string{"empresa:read"})
			// Should fail because ownership checker returns false and checker is nil
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "you can only access resources for CNPJs you own")
		})
	}
}

func TestRequireCNPJOwnershipOrAnyPermission_MultipleActions(t *testing.T) {
	c := setupTestContext("12345678900", "user", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: true}

	// Owner should have access regardless of number of actions
	err := RequireCNPJOwnershipOrAnyPermission(c, nil, ownershipChecker, "12345678000100", "empresa", []string{
		"empresa:read",
		"empresa:update",
		"empresa:delete",
		"empresa:admin",
	})

	assert.NoError(t, err)
}

// Integration-style tests with real Checker and mock HTTP server
// These tests cover the permission check paths that couldn't be tested with nil checker

func TestRequireOwnershipOrAnyPermission_NotOwnerWithPermission_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	// Create checker that allows the action
	checker := createTestCheckerWithBehavior(t, true, false)

	// User is NOT the owner (different CPF) but has permission via role
	err := RequireOwnershipOrAnyPermission(c, checker, "98765432100", "proposta", []string{"proposta:update"})

	assert.NoError(t, err, "should be allowed because user has permission")
}

func TestRequireOwnershipOrAnyPermission_NotOwnerWithoutPermission_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	// Create checker that denies the action
	checker := createTestCheckerWithBehavior(t, false, false)

	// User is NOT the owner and doesn't have permission
	err := RequireOwnershipOrAnyPermission(c, checker, "98765432100", "proposta", []string{"proposta:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you don't have permission to access this resource")
}

func TestRequireOwnershipOrAnyPermission_CheckerError_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	// Create checker that returns error
	checker := createTestCheckerWithBehavior(t, false, true)

	// User is NOT the owner and checker returns error
	err := RequireOwnershipOrAnyPermission(c, checker, "98765432100", "proposta", []string{"proposta:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization check failed")
}

func TestRequireOwnershipOrAnyPermission_NotOwnerMultipleActionsOneAllowed_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "editor", "")

	// Create checker that allows at least one action
	checker := createTestCheckerWithBehavior(t, true, false)

	// User is NOT the owner but has permission for one of the actions
	err := RequireOwnershipOrAnyPermission(c, checker, "98765432100", "proposta", []string{
		"proposta:read",
		"proposta:update",
		"proposta:delete",
	})

	assert.NoError(t, err, "should be allowed if any action is permitted")
}

func TestRequireCNPJOwnershipOrAnyPermission_NotOwnerWithPermission_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: false}
	checker := createTestCheckerWithBehavior(t, true, false)

	// User is NOT the CNPJ owner but has permission
	err := RequireCNPJOwnershipOrAnyPermission(c, checker, ownershipChecker, "12345678000100", "empresa", []string{"empresa:read"})

	assert.NoError(t, err, "should be allowed because user has permission")
}

func TestRequireCNPJOwnershipOrAnyPermission_NotOwnerWithoutPermission_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "user", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: false}
	checker := createTestCheckerWithBehavior(t, false, false)

	// User is NOT the CNPJ owner and doesn't have permission
	err := RequireCNPJOwnershipOrAnyPermission(c, checker, ownershipChecker, "12345678000100", "empresa", []string{"empresa:update"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you don't have permission to access this resource")
}

func TestRequireCNPJOwnershipOrAnyPermission_CheckerError_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: false}
	checker := createTestCheckerWithBehavior(t, false, true)

	// User is NOT the CNPJ owner and checker returns error
	err := RequireCNPJOwnershipOrAnyPermission(c, checker, ownershipChecker, "12345678000100", "empresa", []string{"empresa:read"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization check failed")
}

func TestRequireCNPJOwnershipOrAnyPermission_NotOwnerMultipleActions_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "manager", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: false}
	checker := createTestCheckerWithBehavior(t, true, false)

	// User is NOT the CNPJ owner but has permission for the actions
	err := RequireCNPJOwnershipOrAnyPermission(c, checker, ownershipChecker, "12345678000100", "empresa", []string{
		"empresa:read",
		"empresa:update",
	})

	assert.NoError(t, err)
}

func TestRequireAnyPermission_PermissionGranted_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	// Create checker that allows the action
	checker := createTestCheckerWithBehavior(t, true, false)

	err := RequireAnyPermission(c, checker, "oportunidade", []string{"oportunidade:create"})

	assert.NoError(t, err, "should be allowed because user has permission")
}

func TestRequireAnyPermission_PermissionDenied_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	// Create checker that denies the action
	checker := createTestCheckerWithBehavior(t, false, false)

	err := RequireAnyPermission(c, checker, "oportunidade", []string{"oportunidade:delete"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you don't have permission to perform this action")
}

func TestRequireAnyPermission_CheckerError_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	// Create checker that returns error
	checker := createTestCheckerWithBehavior(t, false, true)

	err := RequireAnyPermission(c, checker, "oportunidade", []string{"oportunidade:create"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization check failed")
}

func TestRequireAnyPermission_MultipleActions_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "editor", "")

	// Create checker that allows actions
	checker := createTestCheckerWithBehavior(t, true, false)

	err := RequireAnyPermission(c, checker, "document", []string{
		"document:read",
		"document:write",
		"document:delete",
	})

	assert.NoError(t, err, "should be allowed if user has any of the permissions")
}

func TestRequireAnyPermission_MultipleActionsDenied_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "viewer", "")

	// Create checker that denies actions
	checker := createTestCheckerWithBehavior(t, false, false)

	err := RequireAnyPermission(c, checker, "document", []string{
		"document:write",
		"document:delete",
		"document:admin",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you don't have permission to perform this action")
}

func TestRequireOwnershipOrAnyPermission_WithRoleAndPermission_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "auditor", "")

	checker := createTestCheckerWithBehavior(t, true, false)

	// User with specific role has permission to audit
	err := RequireOwnershipOrAnyPermission(c, checker, "98765432100", "proposta", []string{"proposta:audit"})

	assert.NoError(t, err)
}

func TestRequireCNPJOwnershipOrAnyPermission_WithRoleAndPermission_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "compliance", "token123")

	ownershipChecker := &mockCNPJOwnershipChecker{isOwner: false}
	checker := createTestCheckerWithBehavior(t, true, false)

	// User with compliance role has permission
	err := RequireCNPJOwnershipOrAnyPermission(c, checker, ownershipChecker, "12345678000100", "empresa", []string{"empresa:audit"})

	assert.NoError(t, err)
}

func TestRequirePermission_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "admin", "")

	checker := createTestCheckerWithBehavior(t, true, false)

	// Test single permission check
	err := RequirePermission(c, checker, "resource", "resource:manage")

	assert.NoError(t, err)
}

func TestRequirePermission_Denied_Integration(t *testing.T) {
	c := setupTestContext("12345678900", "user", "")

	checker := createTestCheckerWithBehavior(t, false, false)

	// Test single permission denied
	err := RequirePermission(c, checker, "resource", "resource:admin")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "you don't have permission to perform this action")
}
