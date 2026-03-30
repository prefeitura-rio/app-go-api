package authorization

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCerbosServer creates a test HTTP server that mimics Cerbos responses
func mockCerbosServer(t *testing.T, responseFunc func(req *CheckResourcesRequest) *CheckResourcesResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and content type
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request
		var req CheckResourcesRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Generate response
		response := responseFunc(&req)

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
}

func TestNewChecker(t *testing.T) {
	endpoint := "http://cerbos:3592/api/check/resources"
	timeout := 10 * time.Second

	checker := NewChecker(endpoint, timeout)

	assert.NotNil(t, checker)
	assert.NotNil(t, checker.client)
	assert.Equal(t, endpoint, checker.client.endpoint)
	assert.Equal(t, timeout, checker.client.httpClient.Timeout)
}

func TestCheckAction_Allowed(t *testing.T) {
	// Create mock Cerbos server that returns ALLOW
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		// Verify request structure
		assert.Equal(t, "12345678900", req.Principal.ID)
		assert.Equal(t, []string{"admin"}, req.Principal.Roles)
		assert.Equal(t, "default", req.Principal.PolicyVersion)
		assert.Equal(t, "12345678900", req.Principal.Attr["cpf"])

		assert.Len(t, req.Resources, 1)
		assert.Equal(t, "proposta", req.Resources[0].Resource.Kind)
		assert.Equal(t, "resource", req.Resources[0].Resource.ID)
		assert.Equal(t, []string{"proposta:create"}, req.Resources[0].Actions)

		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"proposta:create": "EFFECT_ALLOW",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CheckAction(
		context.Background(),
		"12345678900",
		[]string{"admin"},
		"proposta",
		"proposta:create",
	)

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckAction_Denied(t *testing.T) {
	// Create mock Cerbos server that returns DENY
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"proposta:delete": "EFFECT_DENY",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CheckAction(
		context.Background(),
		"98765432100",
		[]string{"user"},
		"proposta",
		"proposta:delete",
	)

	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckAction_WithMultipleRoles(t *testing.T) {
	// Test with user having multiple roles
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		assert.Equal(t, []string{"admin", "auditor"}, req.Principal.Roles)

		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"audit:view": "EFFECT_ALLOW",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CheckAction(
		context.Background(),
		"11111111111",
		[]string{"admin", "auditor"},
		"audit",
		"audit:view",
	)

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckAction_ServerError(t *testing.T) {
	// Create server that returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CheckAction(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"resource",
		"action",
	)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "cerbos returned status 500")
}

func TestCheckAction_NetworkError(t *testing.T) {
	// Use invalid endpoint to trigger network error
	checker := NewChecker("http://invalid-endpoint-that-does-not-exist:9999", 1*time.Second)

	allowed, err := checker.CheckAction(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"resource",
		"action",
	)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "cerbos check failed")
}

func TestCheckAction_ContextCanceled(t *testing.T) {
	// Create a server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	// Create context that's already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	allowed, err := checker.CheckAction(
		ctx,
		"12345678900",
		[]string{"user"},
		"resource",
		"action",
	)

	assert.Error(t, err)
	assert.False(t, allowed)
}

func TestCheckAnyAction_OneAllowed(t *testing.T) {
	// Test when one of multiple actions is allowed
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		assert.Equal(t, []string{"proposta:read", "proposta:update", "proposta:delete"}, req.Resources[0].Actions)

		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"proposta:read":   "EFFECT_ALLOW",
						"proposta:update": "EFFECT_DENY",
						"proposta:delete": "EFFECT_DENY",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CheckAnyAction(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"proposta",
		[]string{"proposta:read", "proposta:update", "proposta:delete"},
	)

	assert.NoError(t, err)
	assert.True(t, allowed, "should be allowed because proposta:read is allowed")
}

func TestCheckAnyAction_AllAllowed(t *testing.T) {
	// Test when all actions are allowed
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"proposta:read":   "EFFECT_ALLOW",
						"proposta:update": "EFFECT_ALLOW",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CheckAnyAction(
		context.Background(),
		"12345678900",
		[]string{"admin"},
		"proposta",
		[]string{"proposta:read", "proposta:update"},
	)

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckAnyAction_NoneAllowed(t *testing.T) {
	// Test when no actions are allowed
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"proposta:delete": "EFFECT_DENY",
						"proposta:admin":  "EFFECT_DENY",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CheckAnyAction(
		context.Background(),
		"98765432100",
		[]string{"user"},
		"proposta",
		[]string{"proposta:delete", "proposta:admin"},
	)

	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCheckAnyAction_EmptyActions(t *testing.T) {
	// Test with empty actions slice
	checker := NewChecker("http://cerbos:3592", 5*time.Second)

	allowed, err := checker.CheckAnyAction(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"proposta",
		[]string{},
	)

	assert.NoError(t, err)
	assert.False(t, allowed, "empty actions should return false")
}

func TestCheckAnyAction_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CheckAnyAction(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"resource",
		[]string{"action1", "action2"},
	)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "cerbos returned status 502")
}

func TestCanAccessResource_Allowed(t *testing.T) {
	// Test resource access with specific resource attributes
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		// Verify resource-specific attributes are passed
		assert.Equal(t, "proposta-123", req.Resources[0].Resource.ID)
		assert.Equal(t, "12345678900", req.Resources[0].Resource.Attr["owner_cpf"])
		assert.Equal(t, "pending", req.Resources[0].Resource.Attr["status"])

		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"proposta:view": "EFFECT_ALLOW",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	resourceAttrs := map[string]interface{}{
		"owner_cpf": "12345678900",
		"status":    "pending",
	}

	allowed, err := checker.CanAccessResource(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"proposta",
		"proposta-123",
		"proposta:view",
		resourceAttrs,
	)

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCanAccessResource_Denied(t *testing.T) {
	// Test access denied to resource
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"proposta:delete": "EFFECT_DENY",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	resourceAttrs := map[string]interface{}{
		"owner_cpf": "98765432100",
		"status":    "approved",
	}

	allowed, err := checker.CanAccessResource(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"proposta",
		"proposta-456",
		"proposta:delete",
		resourceAttrs,
	)

	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestCanAccessResource_WithOwnership(t *testing.T) {
	// Test that ownership is properly checked via resource attributes
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		// Cerbos policy should check if principal.attr.cpf == resource.attr.owner_cpf
		principalCPF := req.Principal.Attr["cpf"].(string)
		ownerCPF := req.Resources[0].Resource.Attr["owner_cpf"].(string)

		isOwner := principalCPF == ownerCPF

		decision := "EFFECT_DENY"
		if isOwner {
			decision = "EFFECT_ALLOW"
		}

		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"proposta:update": decision,
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	// Test owner access
	resourceAttrs := map[string]interface{}{
		"owner_cpf": "12345678900",
	}

	allowed, err := checker.CanAccessResource(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"proposta",
		"proposta-123",
		"proposta:update",
		resourceAttrs,
	)

	assert.NoError(t, err)
	assert.True(t, allowed, "owner should be allowed")

	// Test non-owner access
	allowed, err = checker.CanAccessResource(
		context.Background(),
		"98765432100", // different CPF
		[]string{"user"},
		"proposta",
		"proposta-123",
		"proposta:update",
		resourceAttrs,
	)

	assert.NoError(t, err)
	assert.False(t, allowed, "non-owner should be denied")
}

func TestCanAccessResource_EmptyResourceAttrs(t *testing.T) {
	// Test with empty resource attributes (no ownership)
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		assert.Empty(t, req.Resources[0].Resource.Attr)

		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"public:read": "EFFECT_ALLOW",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CanAccessResource(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"public",
		"resource-1",
		"public:read",
		map[string]interface{}{},
	)

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCanAccessResource_NilResourceAttrs(t *testing.T) {
	// Test with nil resource attributes
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		// Should handle nil gracefully
		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"action": "EFFECT_ALLOW",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CanAccessResource(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"resource",
		"id-1",
		"action",
		nil,
	)

	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestCanAccessResource_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	allowed, err := checker.CanAccessResource(
		context.Background(),
		"12345678900",
		[]string{"user"},
		"resource",
		"id-1",
		"action",
		map[string]interface{}{},
	)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "cerbos returned status 503")
}

func TestCanAccessResource_ComplexAttributes(t *testing.T) {
	// Test with complex resource attributes
	server := mockCerbosServer(t, func(req *CheckResourcesRequest) *CheckResourcesResponse {
		// Verify complex attributes are passed correctly
		assert.Equal(t, "12345678900", req.Resources[0].Resource.Attr["owner_cpf"])
		assert.Equal(t, "approved", req.Resources[0].Resource.Attr["status"])
		assert.Equal(t, float64(100000), req.Resources[0].Resource.Attr["amount"])
		assert.Equal(t, true, req.Resources[0].Resource.Attr["verified"])

		tags, ok := req.Resources[0].Resource.Attr["tags"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, tags, 2)

		return &CheckResourcesResponse{
			RequestID: req.RequestID,
			Results: []ResourceActionResponse{
				{
					Resource: req.Resources[0].Resource,
					Actions: map[string]string{
						"proposta:approve": "EFFECT_ALLOW",
					},
				},
			},
		}
	})
	defer server.Close()

	checker := NewChecker(server.URL, 5*time.Second)

	resourceAttrs := map[string]interface{}{
		"owner_cpf": "12345678900",
		"status":    "approved",
		"amount":    100000.0,
		"verified":  true,
		"tags":      []string{"high-priority", "urgent"},
	}

	allowed, err := checker.CanAccessResource(
		context.Background(),
		"12345678900",
		[]string{"approver"},
		"proposta",
		"proposta-999",
		"proposta:approve",
		resourceAttrs,
	)

	assert.NoError(t, err)
	assert.True(t, allowed)
}
