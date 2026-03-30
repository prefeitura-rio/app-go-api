package middlewares

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestFetchHeimdallUserInfo_Success(t *testing.T) {
	// Create a test server that mocks Heimdall
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Return mock user info
		userInfo := HeimdallUserInfo{
			ID:          123,
			CPF:         "12345678900",
			DisplayName: "Test User",
			Groups:      []string{"test-group"},
			Roles:       []string{"go:admin", "user"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer mockServer.Close()

	userInfo, err := fetchHeimdallUserInfo(context.Background(), mockServer.URL, "Bearer test-token")

	assert.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, 123, userInfo.ID)
	assert.Equal(t, "12345678900", userInfo.CPF)
	assert.Equal(t, "Test User", userInfo.DisplayName)
	assert.Equal(t, []string{"go:admin", "user"}, userInfo.Roles)
}

func TestFetchHeimdallUserInfo_AddsBearer(t *testing.T) {
	// Create a test server that checks for Bearer prefix
	receivedAuth := ""
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		userInfo := HeimdallUserInfo{ID: 1, CPF: "123", Roles: []string{}}
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer mockServer.Close()

	// Send without Bearer prefix
	_, err := fetchHeimdallUserInfo(context.Background(), mockServer.URL, "test-token-no-bearer")

	assert.NoError(t, err)
	assert.Equal(t, "Bearer test-token-no-bearer", receivedAuth)
}

func TestFetchHeimdallUserInfo_KeepsExistingBearer(t *testing.T) {
	// Create a test server that checks for Bearer prefix
	receivedAuth := ""
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		userInfo := HeimdallUserInfo{ID: 1, CPF: "123", Roles: []string{}}
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer mockServer.Close()

	// Send with Bearer prefix already
	_, err := fetchHeimdallUserInfo(context.Background(), mockServer.URL, "Bearer existing-token")

	assert.NoError(t, err)
	assert.Equal(t, "Bearer existing-token", receivedAuth)
}

func TestFetchHeimdallUserInfo_Unauthorized(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer mockServer.Close()

	userInfo, err := fetchHeimdallUserInfo(context.Background(), mockServer.URL, "Bearer bad-token")

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.Contains(t, err.Error(), "returned status 401")
}

func TestFetchHeimdallUserInfo_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	userInfo, err := fetchHeimdallUserInfo(context.Background(), mockServer.URL, "Bearer token")

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.Contains(t, err.Error(), "returned status 500")
}

func TestFetchHeimdallUserInfo_InvalidJSON(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer mockServer.Close()

	userInfo, err := fetchHeimdallUserInfo(context.Background(), mockServer.URL, "Bearer token")

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestFetchHeimdallUserInfo_ContextCancellation(t *testing.T) {
	// Create a server with delay
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(HeimdallUserInfo{ID: 1})
	}))
	defer mockServer.Close()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	userInfo, err := fetchHeimdallUserInfo(ctx, mockServer.URL, "Bearer token")

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	assert.Contains(t, err.Error(), "request failed")
}

func TestExtractUserContext_WithHeimdall(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock Heimdall server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo := HeimdallUserInfo{
			ID:          456,
			CPF:         "98765432100",
			DisplayName: "Admin User",
			Roles:       []string{"go:admin", "superadmin"},
		}
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer mockServer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	token := createTestJWT(t, map[string]interface{}{
		"preferred_username": "987.654.321-00",
		"sub":                "user-456",
		"name":               "Test Name",
		"email":              "test@example.com",
	})
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := ExtractUserContext(mockServer.URL)
	middleware(c)

	// Should have fetched roles from Heimdall
	assert.Equal(t, "98765432100", GetUserCPF(c))
	assert.Equal(t, "ADMIN", GetUserRole(c))
	roles := GetAllRoles(c)
	assert.Contains(t, roles, "go:admin")
	assert.Contains(t, roles, "superadmin")
}

func TestExtractUserContext_HeimdallFails_FallbackToUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock Heimdall server that returns error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	token := createTestJWT(t, map[string]interface{}{
		"preferred_username": "123.456.789-00",
		"sub":                "user-123",
	})
	c.Request.Header.Set("Authorization", "Bearer "+token)

	middleware := ExtractUserContext(mockServer.URL)
	middleware(c)

	// Should fallback to USER role when Heimdall fails
	assert.Equal(t, "12345678900", GetUserCPF(c))
	assert.Equal(t, "USER", GetUserRole(c))
}

func TestExtractUserContext_WithoutHeimdall_UsesJWTRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	token := createTestJWT(t, map[string]interface{}{
		"preferred_username": "111.222.333-44",
		"sub":                "user-789",
		"resource_access": map[string]interface{}{
			"superapp": map[string]interface{}{
				"roles": []string{"go:admin", "custom-role"},
			},
		},
	})
	c.Request.Header.Set("Authorization", "Bearer "+token)

	// No Heimdall URL provided
	middleware := ExtractUserContext("")
	middleware(c)

	// Should use JWT roles
	assert.Equal(t, "11122233344", GetUserCPF(c))
	assert.Equal(t, "ADMIN", GetUserRole(c))
	roles := GetAllRoles(c)
	assert.Contains(t, roles, "go:admin")
	assert.Contains(t, roles, "custom-role")
}

func TestExtractUserContext_AdminRoleVariants(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		roles        []string
		expectedRole string
	}{
		{
			name:         "go:admin",
			roles:        []string{"go:admin", "user"},
			expectedRole: "ADMIN",
		},
		{
			name:         "admin",
			roles:        []string{"admin", "user"},
			expectedRole: "ADMIN",
		},
		{
			name:         "superadmin",
			roles:        []string{"superadmin"},
			expectedRole: "ADMIN",
		},
		{
			name:         "no admin role",
			roles:        []string{"user", "viewer"},
			expectedRole: "USER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock Heimdall server
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userInfo := HeimdallUserInfo{
					ID:    1,
					CPF:   "12345678900",
					Roles: tt.roles,
				}
				json.NewEncoder(w).Encode(userInfo)
			}))
			defer mockServer.Close()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			token := createTestJWT(t, map[string]interface{}{
				"preferred_username": "123.456.789-00",
				"sub":                "user-1",
			})
			c.Request.Header.Set("Authorization", "Bearer "+token)

			middleware := ExtractUserContext(mockServer.URL)
			middleware(c)

			assert.Equal(t, tt.expectedRole, GetUserRole(c))
		})
	}
}

func TestDecodeJWT_MalformedJSON(t *testing.T) {
	// Create a JWT with invalid JSON in payload
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{invalid-json`))
	signature := base64.RawURLEncoding.EncodeToString([]byte("sig"))

	token := header + "." + payload + "." + signature

	claims, err := decodeJWT(token)

	assert.Nil(t, err) // Should return nil without error
	assert.Nil(t, claims)
}

func TestDecodeJWT_InvalidBase64(t *testing.T) {
	// Create a JWT with invalid base64 in payload
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`))
	payload := "!!!invalid-base64!!!"
	signature := base64.RawURLEncoding.EncodeToString([]byte("sig"))

	token := header + "." + payload + "." + signature

	claims, err := decodeJWT(token)

	assert.Nil(t, err) // Should return nil without error
	assert.Nil(t, claims)
}

func TestDecodeJWT_WithBearerPrefix(t *testing.T) {
	token := createTestJWT(t, map[string]interface{}{
		"preferred_username": "12345678900",
		"sub":                "user-1",
	})

	// Test with Bearer prefix
	claims, err := decodeJWT("Bearer " + token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "12345678900", claims.PreferredUsername)
}

func TestGetUserHelpers_WrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set wrong types
	c.Set(UserCPFKey, 123) // int instead of string
	c.Set(UserRoleKey, 456)
	c.Set(UserIDKey, true)
	c.Set(UserNameKey, []string{"test"})
	c.Set(UserEmailKey, map[string]string{})
	c.Set(UserRolesKey, "not-a-slice")

	// Should return empty strings when type is wrong
	assert.Equal(t, "", GetUserCPF(c))
	assert.Equal(t, "", GetUserRole(c))
	assert.Equal(t, "", GetUserID(c))
	assert.Equal(t, "", GetUserName(c))
	assert.Equal(t, "", GetUserEmail(c))
	assert.Nil(t, GetAllRoles(c))
	assert.False(t, IsAdmin(c))
}

func TestHasRole_NilRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Don't set any roles
	assert.False(t, HasRole(c, "any-role"))
}

func TestIsEmpregabilidadeRole_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		roles  []string
		expect bool
	}{
		{
			name:   "exact prefix only",
			roles:  []string{"go:empregabilidade:"},
			expect: true,
		},
		{
			name:   "prefix in middle",
			roles:  []string{"user", "go:empregabilidade:viewer", "admin"},
			expect: true,
		},
		{
			name:   "similar but not matching",
			roles:  []string{"go:empregabilidad:", "empregabilidade"},
			expect: false,
		},
		{
			name:   "nil roles",
			roles:  nil,
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.roles != nil {
				c.Set(UserRolesKey, tt.roles)
			}

			assert.Equal(t, tt.expect, IsEmpregabilidadeRole(c))
		})
	}
}
