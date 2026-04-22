package authorization

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckResourcesResponse_IsAllowed(t *testing.T) {
	tests := []struct {
		name     string
		response *CheckResourcesResponse
		action   string
		want     bool
	}{
		{
			name: "action allowed",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"read": "EFFECT_ALLOW",
						},
					},
				},
			},
			action: "read",
			want:   true,
		},
		{
			name: "action denied",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"read": "EFFECT_DENY",
						},
					},
				},
			},
			action: "read",
			want:   false,
		},
		{
			name: "action not in response",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"write": "EFFECT_ALLOW",
						},
					},
				},
			},
			action: "read",
			want:   false,
		},
		{
			name:     "empty results",
			response: &CheckResourcesResponse{Results: []ResourceActionResponse{}},
			action:   "read",
			want:     false,
		},
		{
			name:     "nil results",
			response: &CheckResourcesResponse{},
			action:   "read",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.response.IsAllowed(tt.action)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckResourcesResponse_IsAnyAllowed(t *testing.T) {
	tests := []struct {
		name     string
		response *CheckResourcesResponse
		actions  []string
		want     bool
	}{
		{
			name: "one action allowed",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"read":  "EFFECT_ALLOW",
							"write": "EFFECT_DENY",
						},
					},
				},
			},
			actions: []string{"read", "write"},
			want:    true,
		},
		{
			name: "all actions denied",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"read":  "EFFECT_DENY",
							"write": "EFFECT_DENY",
						},
					},
				},
			},
			actions: []string{"read", "write"},
			want:    false,
		},
		{
			name: "no actions in response",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"delete": "EFFECT_ALLOW",
						},
					},
				},
			},
			actions: []string{"read", "write"},
			want:    false,
		},
		{
			name:     "empty results",
			response: &CheckResourcesResponse{Results: []ResourceActionResponse{}},
			actions:  []string{"read"},
			want:     false,
		},
		{
			name: "empty actions list",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"read": "EFFECT_ALLOW",
						},
					},
				},
			},
			actions: []string{},
			want:    false,
		},
		{
			name: "multiple actions with one allowed",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"read":   "EFFECT_DENY",
							"write":  "EFFECT_DENY",
							"delete": "EFFECT_ALLOW",
						},
					},
				},
			},
			actions: []string{"read", "write", "delete"},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.response.IsAnyAllowed(tt.actions)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckResourcesResponse_GetDecision(t *testing.T) {
	tests := []struct {
		name     string
		response *CheckResourcesResponse
		action   string
		want     string
	}{
		{
			name: "get allow decision",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"read": "EFFECT_ALLOW",
						},
					},
				},
			},
			action: "read",
			want:   "EFFECT_ALLOW",
		},
		{
			name: "get deny decision",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"read": "EFFECT_DENY",
						},
					},
				},
			},
			action: "read",
			want:   "EFFECT_DENY",
		},
		{
			name: "action not found",
			response: &CheckResourcesResponse{
				Results: []ResourceActionResponse{
					{
						Actions: map[string]string{
							"write": "EFFECT_ALLOW",
						},
					},
				},
			},
			action: "read",
			want:   "",
		},
		{
			name:     "empty results",
			response: &CheckResourcesResponse{Results: []ResourceActionResponse{}},
			action:   "read",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.response.GetDecision(tt.action)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateRequestID(t *testing.T) {
	// Generate two IDs sequentially
	id1 := generateRequestID()
	time.Sleep(1 * time.Millisecond) // Small delay to ensure different timestamp
	id2 := generateRequestID()

	// Should not be empty
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)

	// Should be different (time-based)
	assert.NotEqual(t, id1, id2)

	// Should match expected format (YYYYMMDDHHMMSS.microseconds)
	assert.Len(t, id1, 21) // 14 digits + dot + 6 digits = 21 characters
	assert.Contains(t, id1, ".")
}

func TestResourceInfo_Structure(t *testing.T) {
	// Test that ResourceInfo can be created with expected fields
	info := ResourceInfo{
		Kind: "document",
		ID:   "doc-123",
		Attr: map[string]interface{}{
			"owner":  "user-456",
			"public": true,
		},
	}

	assert.Equal(t, "document", info.Kind)
	assert.Equal(t, "doc-123", info.ID)
	assert.Equal(t, "user-456", info.Attr["owner"])
	assert.Equal(t, true, info.Attr["public"])
}

func TestResource_Structure(t *testing.T) {
	// Test that Resource can be created with expected fields
	resource := Resource{
		Resource: ResourceInfo{
			Kind: "file",
			ID:   "file-789",
			Attr: map[string]interface{}{},
		},
		Actions: []string{"read", "write"},
	}

	assert.Equal(t, "file", resource.Resource.Kind)
	assert.Equal(t, "file-789", resource.Resource.ID)
	assert.Equal(t, []string{"read", "write"}, resource.Actions)
}

func TestPrincipal_Structure(t *testing.T) {
	// Test that Principal can be created with expected fields
	principal := Principal{
		ID:            "12345678900",
		Roles:         []string{"admin", "user"},
		PolicyVersion: "default",
		Attr: map[string]interface{}{
			"cpf":        "12345678900",
			"department": "IT",
		},
	}

	assert.Equal(t, "12345678900", principal.ID)
	assert.Equal(t, []string{"admin", "user"}, principal.Roles)
	assert.Equal(t, "default", principal.PolicyVersion)
	assert.Equal(t, "12345678900", principal.Attr["cpf"])
	assert.Equal(t, "IT", principal.Attr["department"])
}

func TestCheckResourcesRequest_Structure(t *testing.T) {
	// Test that CheckResourcesRequest can be created with expected fields
	request := CheckResourcesRequest{
		RequestID: "test-request-id",
		Principal: Principal{
			ID:            "user-123",
			Roles:         []string{"user"},
			PolicyVersion: "default",
			Attr:          map[string]interface{}{},
		},
		Resources: []Resource{
			{
				Resource: ResourceInfo{
					Kind: "document",
					ID:   "doc-1",
					Attr: map[string]interface{}{},
				},
				Actions: []string{"read"},
			},
		},
	}

	assert.Equal(t, "test-request-id", request.RequestID)
	assert.Equal(t, "user-123", request.Principal.ID)
	assert.Len(t, request.Resources, 1)
	assert.Equal(t, "document", request.Resources[0].Resource.Kind)
}
