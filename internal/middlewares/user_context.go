package middlewares

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	UserCPFKey                = "user_cpf"
	UserRoleKey               = "user_role"
	UserRolesKey              = "user_roles"
	UserGroupsKey             = "user_groups"
	UserIDKey                 = "user_id"
	UserNameKey               = "user_name"
	UserEmailKey              = "user_email"
	UserSecretariaOrgaoIDsKey = "user_secretaria_orgao_ids"
)

// JWTClaims representa as claims do JWT que nos interessam
type JWTClaims struct {
	PreferredUsername string `json:"preferred_username"` // CPF
	Sub               string `json:"sub"`                // User ID
	Name              string `json:"name"`               // Full name
	Email             string `json:"email"`              // Email
	ResourceAccess    map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"` // Roles
}

// HeimdallUserInfo representa a resposta do Heimdall para /api/v1/users/me
type HeimdallUserInfo struct {
	ID          int      `json:"id"`
	CPF         string   `json:"cpf"`
	DisplayName string   `json:"display_name"`
	Groups      []string `json:"groups"`
	Roles       []string `json:"roles"`
}

var heimdallHTTPClient = &http.Client{Timeout: 3 * time.Second}

// fetchHeimdallUserInfo consulta o Heimdall para obter as roles do usuário
func fetchHeimdallUserInfo(ctx context.Context, baseURL, authHeader string) (*HeimdallUserInfo, error) {
	// Garante que o header tem o prefixo "Bearer "
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		authHeader = "Bearer " + authHeader
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/users/me", nil)
	if err != nil {
		return nil, fmt.Errorf("heimdall: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := heimdallHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("heimdall: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("heimdall: returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("heimdall: failed to read response: %w", err)
	}

	var userInfo HeimdallUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("heimdall: failed to decode response: %w", err)
	}

	return &userInfo, nil
}

// decodeJWT decodifica o payload do JWT (sem validar assinatura - já validado pelo Istio/Keycloak)
func decodeJWT(token string) (*JWTClaims, error) {
	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)

	// Split token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil // Not a valid JWT format, skip silently
	}

	// Decode payload (second part)
	payload := parts[1]
	// Add padding if needed
	if l := len(payload) % 4; l > 0 {
		payload += strings.Repeat("=", 4-l)
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, nil // Invalid base64, skip silently
	}

	var claims JWTClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, nil // Invalid JSON, skip silently
	}

	return &claims, nil
}

// ExtractUserContext extrai informações do usuário do JWT no header Authorization.
// Quando heimdallBaseURL é fornecido, consulta o Heimdall para obter as roles reais do usuário.
// Caso contrário, lê as roles diretamente do JWT (resource_access.superapp.roles).
func ExtractUserContext(heimdallBaseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try X-Auth-Request-Token first (Istio), then Authorization header
		authHeader := c.GetHeader("X-Auth-Request-Token")
		if authHeader == "" {
			authHeader = c.GetHeader("Authorization")
		}

		if authHeader == "" {
			c.Next()
			return
		}

		// Decode JWT
		claims, err := decodeJWT(authHeader)
		if err != nil || claims == nil {
			c.Next()
			return
		}

		// Extract CPF from preferred_username (remove dots and dashes)
		if claims.PreferredUsername != "" {
			cpf := strings.NewReplacer(".", "", "-", "").Replace(claims.PreferredUsername)
			c.Set(UserCPFKey, cpf)
		}

		// Extract user ID
		if claims.Sub != "" {
			c.Set(UserIDKey, claims.Sub)
		}

		// Extract name
		if claims.Name != "" {
			c.Set(UserNameKey, claims.Name)
		}

		// Extract email
		if claims.Email != "" {
			c.Set(UserEmailKey, claims.Email)
		}

		// Determine role
		role := "USER"
		if heimdallBaseURL != "" {
			// Consulta Heimdall para obter as roles reais
			userInfo, err := fetchHeimdallUserInfo(c.Request.Context(), heimdallBaseURL, authHeader)
			if err == nil && userInfo != nil {
				c.Set(UserRolesKey, userInfo.Roles)
				c.Set(UserGroupsKey, userInfo.Groups)
				for _, r := range userInfo.Roles {
					if r == "go:admin" || r == "admin" || r == "superadmin" {
						role = "ADMIN"
						break
					}
				}
			}
			// Em caso de erro, mantém role USER (fail-safe)
		} else {
			// Fallback: lê roles diretamente do JWT (para dev local sem Heimdall configurado)
			if superappAccess, ok := claims.ResourceAccess["superapp"]; ok {
				c.Set(UserRolesKey, superappAccess.Roles)
				for _, r := range superappAccess.Roles {
					if r == "go:admin" {
						role = "ADMIN"
						break
					}
				}
			}
		}
		c.Set(UserRoleKey, role)

		c.Next()
	}
}

// GetUserCPF retorna o CPF do usuário autenticado
func GetUserCPF(c *gin.Context) string {
	if cpf, exists := c.Get(UserCPFKey); exists {
		if cpfStr, ok := cpf.(string); ok {
			return cpfStr
		}
	}
	return ""
}

// GetUserRole retorna o role do usuário (ADMIN ou USER)
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get(UserRoleKey); exists {
		if roleStr, ok := role.(string); ok {
			return roleStr
		}
	}
	return ""
}

// GetUserID retorna o ID único do usuário
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get(UserIDKey); exists {
		if userIDStr, ok := userID.(string); ok {
			return userIDStr
		}
	}
	return ""
}

// GetUserName retorna o nome completo do usuário
func GetUserName(c *gin.Context) string {
	if userName, exists := c.Get(UserNameKey); exists {
		if userNameStr, ok := userName.(string); ok {
			return userNameStr
		}
	}
	return ""
}

// GetUserEmail retorna o email do usuário
func GetUserEmail(c *gin.Context) string {
	if userEmail, exists := c.Get(UserEmailKey); exists {
		if userEmailStr, ok := userEmail.(string); ok {
			return userEmailStr
		}
	}
	return ""
}

// IsAdmin verifica se o usuário tem role ADMIN
func IsAdmin(c *gin.Context) bool {
	role := GetUserRole(c)
	return role == "ADMIN"
}

// GetAllRoles retorna todas as roles do usuário autenticado
func GetAllRoles(c *gin.Context) []string {
	if roles, exists := c.Get(UserRolesKey); exists {
		if rolesSlice, ok := roles.([]string); ok {
			return rolesSlice
		}
	}
	return nil
}

// HasRole verifica se o usuário possui uma role específica
func HasRole(c *gin.Context, role string) bool {
	for _, r := range GetAllRoles(c) {
		if r == role {
			return true
		}
	}
	return false
}

// GetUserGroups retorna todos os grupos do usuário autenticado
func GetUserGroups(c *gin.Context) []string {
	if groups, exists := c.Get(UserGroupsKey); exists {
		if groupsSlice, ok := groups.([]string); ok {
			return groupsSlice
		}
	}
	return nil
}

// GetUserOrgaoID retorna o orgao_id do usuário extraído do grupo go:orgao:{id}.
// Retorna string vazia se o usuário não pertencer a nenhum grupo de orgao.
func GetUserOrgaoID(c *gin.Context) string {
	for _, g := range GetUserGroups(c) {
		if strings.HasPrefix(g, "go:orgao:") {
			return strings.TrimPrefix(g, "go:orgao:")
		}
	}
	return ""
}

// SecretariaOrgaoResolver resolves a CPF to the list of orgao_ids the user is allowed to access
// based on the CPF-Secretaria mappings stored in app-rmi.
type SecretariaOrgaoResolver func(ctx context.Context, cpf string) ([]string, error)

// ExtractSecretariaOrgaoIDs is a middleware that resolves the user's secretaria orgao_ids
// and sets them in the gin context. It must run after ExtractUserContext (needs CPF).
// If the resolver fails or the CPF is absent, the middleware is a no-op (fail-open).
func ExtractSecretariaOrgaoIDs(resolver SecretariaOrgaoResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		cpf := GetUserCPF(c)
		if cpf == "" || resolver == nil {
			c.Next()
			return
		}

		orgaoIDs, err := resolver(c.Request.Context(), cpf)
		if err != nil {
			// fail-closed: error resolving secretaria → user sees nothing
			orgaoIDs = []string{}
		}

		c.Set(UserSecretariaOrgaoIDsKey, orgaoIDs)
		c.Next()
	}
}

// GetUserSecretariaOrgaoIDs returns the list of orgao_ids the user can access via secretaria mapping.
// Returns nil when no secretaria mapping exists for the user (middleware was skipped or no mappings found).
func GetUserSecretariaOrgaoIDs(c *gin.Context) []string {
	if v, exists := c.Get(UserSecretariaOrgaoIDsKey); exists {
		if ids, ok := v.([]string); ok {
			return ids
		}
	}
	return nil
}

// HasSecretariaFilter returns true when the user has at least one secretaria orgao_id mapped,
// meaning secretaria-level filtering should be applied.
func HasSecretariaFilter(c *gin.Context) bool {
	ids := GetUserSecretariaOrgaoIDs(c)
	return len(ids) > 0
}

// IsEmpregabilidadeRole verifica se o usuário possui qualquer role de empregabilidade (prefixo go:empregabilidade:)
func IsEmpregabilidadeRole(c *gin.Context) bool {
	for _, r := range GetAllRoles(c) {
		if strings.HasPrefix(r, "go:empregabilidade:") {
			return true
		}
	}
	return false
}

// RequireOwnershipOrAdmin middleware que verifica se o usuário é owner ou admin
func RequireOwnershipOrAdmin(ownerCPF string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userCPF := GetUserCPF(c)
		userRole := GetUserRole(c)

		// Admin tem acesso total
		if userRole == "ADMIN" {
			c.Next()
			return
		}

		// Se não tem CPF, não está autenticado
		if userCPF == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado"})
			c.Abort()
			return
		}

		// Verifica se é o próprio usuário
		if userCPF != ownerCPF {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: você só pode acessar seus próprios dados"})
			c.Abort()
			return
		}

		c.Next()
	}
}
