package middlewares

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDecodeJWT(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		expectNil bool
		expectCPF string
		expectSub string
	}{
		{
			name:      "valid_jwt_with_cpf",
			token:     createTestJWT(t, map[string]interface{}{"preferred_username": "123.456.789-00", "sub": "user-123"}),
			expectNil: false,
			expectCPF: "123.456.789-00",
			expectSub: "user-123",
		},
		{
			name:      "invalid_jwt_format",
			token:     "not.a.valid.jwt.token",
			expectNil: true,
		},
		{
			name:      "jwt_with_two_parts_only",
			token:     "header.payload",
			expectNil: true,
		},
		{
			name:      "empty_token",
			token:     "",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := decodeJWT(tt.token)

			if tt.expectNil {
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tt.expectCPF, claims.PreferredUsername)
				assert.Equal(t, tt.expectSub, claims.Sub)
			}
		})
	}
}

func TestExtractUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		setupHeaders func(*httptest.ResponseRecorder, *gin.Context)
		expectCPF    string
		expectRole   string
		expectUserID string
		expectName   string
		expectEmail  string
	}{
		{
			name: "extract_from_authorization_header",
			setupHeaders: func(w *httptest.ResponseRecorder, c *gin.Context) {
				token := createTestJWT(t, map[string]interface{}{
					"preferred_username": "123.456.789-00",
					"sub":                "user-123",
					"name":               "Test User",
					"email":              "test@example.com",
					"resource_access": map[string]interface{}{
						"superapp": map[string]interface{}{
							"roles": []string{"go:admin"},
						},
					},
				})
				c.Request.Header.Set("Authorization", "Bearer "+token)
			},
			expectCPF:    "12345678900",
			expectRole:   "ADMIN",
			expectUserID: "user-123",
			expectName:   "Test User",
			expectEmail:  "test@example.com",
		},
		{
			name: "extract_from_x_auth_request_token",
			setupHeaders: func(w *httptest.ResponseRecorder, c *gin.Context) {
				token := createTestJWT(t, map[string]interface{}{
					"preferred_username": "987.654.321-00",
					"sub":                "user-456",
					"resource_access": map[string]interface{}{
						"superapp": map[string]interface{}{
							"roles": []string{"user"},
						},
					},
				})
				c.Request.Header.Set("X-Auth-Request-Token", token)
			},
			expectCPF:    "98765432100",
			expectRole:   "USER",
			expectUserID: "user-456",
		},
		{
			name: "no_auth_header",
			setupHeaders: func(w *httptest.ResponseRecorder, c *gin.Context) {
				// No headers set
			},
			expectCPF:  "",
			expectRole: "",
		},
		{
			name: "user_role_default",
			setupHeaders: func(w *httptest.ResponseRecorder, c *gin.Context) {
				token := createTestJWT(t, map[string]interface{}{
					"preferred_username": "111.222.333-44",
					"sub":                "user-789",
				})
				c.Request.Header.Set("Authorization", "Bearer "+token)
			},
			expectCPF:    "11122233344",
			expectRole:   "USER",
			expectUserID: "user-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			tt.setupHeaders(w, c)

			middleware := ExtractUserContext("")
			middleware(c)

			if tt.expectCPF != "" {
				assert.Equal(t, tt.expectCPF, GetUserCPF(c))
			}
			if tt.expectRole != "" {
				assert.Equal(t, tt.expectRole, GetUserRole(c))
			}
			if tt.expectUserID != "" {
				assert.Equal(t, tt.expectUserID, GetUserID(c))
			}
			if tt.expectName != "" {
				assert.Equal(t, tt.expectName, GetUserName(c))
			}
			if tt.expectEmail != "" {
				assert.Equal(t, tt.expectEmail, GetUserEmail(c))
			}
		})
	}
}

func TestGetUserHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test empty context
	assert.Equal(t, "", GetUserCPF(c))
	assert.Equal(t, "", GetUserRole(c))
	assert.Equal(t, "", GetUserID(c))
	assert.Equal(t, "", GetUserName(c))
	assert.Equal(t, "", GetUserEmail(c))
	assert.False(t, IsAdmin(c))
	assert.Nil(t, GetAllRoles(c))

	// Set context values
	c.Set(UserCPFKey, "12345678900")
	c.Set(UserRoleKey, "ADMIN")
	c.Set(UserIDKey, "user-123")
	c.Set(UserNameKey, "Test User")
	c.Set(UserEmailKey, "test@example.com")
	c.Set(UserRolesKey, []string{"go:admin", "user"})

	// Test populated context
	assert.Equal(t, "12345678900", GetUserCPF(c))
	assert.Equal(t, "ADMIN", GetUserRole(c))
	assert.Equal(t, "user-123", GetUserID(c))
	assert.Equal(t, "Test User", GetUserName(c))
	assert.Equal(t, "test@example.com", GetUserEmail(c))
	assert.True(t, IsAdmin(c))
	assert.True(t, HasRole(c, "go:admin"))
	assert.False(t, HasRole(c, "nonexistent"))
	assert.Equal(t, []string{"go:admin", "user"}, GetAllRoles(c))
}

func TestIsEmpregabilidadeRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		roles  []string
		expect bool
	}{
		{
			name:   "has_empregabilidade_role",
			roles:  []string{"go:empregabilidade:admin", "user"},
			expect: true,
		},
		{
			name:   "no_empregabilidade_role",
			roles:  []string{"go:admin", "user"},
			expect: false,
		},
		{
			name:   "empty_roles",
			roles:  []string{},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(UserRolesKey, tt.roles)

			assert.Equal(t, tt.expect, IsEmpregabilidadeRole(c))
		})
	}
}

func TestRequireOwnershipOrAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		ownerCPF     string
		userCPF      string
		userRole     string
		expectStatus int
		expectAbort  bool
	}{
		{
			name:         "admin_access_allowed",
			ownerCPF:     "12345678900",
			userCPF:      "98765432100",
			userRole:     "ADMIN",
			expectStatus: 0,
			expectAbort:  false,
		},
		{
			name:         "owner_access_allowed",
			ownerCPF:     "12345678900",
			userCPF:      "12345678900",
			userRole:     "USER",
			expectStatus: 0,
			expectAbort:  false,
		},
		{
			name:         "unauthorized_no_cpf",
			ownerCPF:     "12345678900",
			userCPF:      "",
			userRole:     "USER",
			expectStatus: http.StatusUnauthorized,
			expectAbort:  true,
		},
		{
			name:         "forbidden_different_user",
			ownerCPF:     "12345678900",
			userCPF:      "98765432100",
			userRole:     "USER",
			expectStatus: http.StatusForbidden,
			expectAbort:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			if tt.userCPF != "" {
				c.Set(UserCPFKey, tt.userCPF)
			}
			if tt.userRole != "" {
				c.Set(UserRoleKey, tt.userRole)
			}

			middleware := RequireOwnershipOrAdmin(tt.ownerCPF)
			middleware(c)

			if tt.expectAbort {
				assert.True(t, c.IsAborted())
				assert.Equal(t, tt.expectStatus, w.Code)
			} else {
				assert.False(t, c.IsAborted())
			}
		})
	}
}

// Helper function to create a test JWT
func createTestJWT(t *testing.T, claims map[string]interface{}) string {
	t.Helper()

	// Create header
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Create payload
	payloadJSON, _ := json.Marshal(claims)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature (fake, we're not validating it)
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))

	return headerB64 + "." + payloadB64 + "." + signature
}
