package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	vagaOrgaoParceiroIDsKey = "vaga_orgao_parceiro_ids"
	vagaAllowedOrgaosKey    = "vaga_allowed_orgaos"
)

func isEmpEditor(c *gin.Context) bool {
	return HasRole(c, "go:empregabilidade:editor") ||
		HasRole(c, "go:empregabilidade:editor_com_curadoria") ||
		HasRole(c, "go:empregabilidade:editor_sem_curadoria")
}

func hasVagaOrgaoScope(c *gin.Context) bool {
	return len(GetUserSecretariaOrgaoIDs(c)) > 0 || GetUserOrgaoID(c) != ""
}

// VagaAuthorization allows only total admins, go:empregabilidade:admin, or emp editors
// with secretaria/orgao scope. Secretaria alone is not enough.
func VagaAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) || HasRole(c, "go:empregabilidade:admin") {
			c.Next()
			return
		}

		if isEmpEditor(c) && hasVagaOrgaoScope(c) {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar este recurso"})
		c.Abort()
	}
}

// VagaListFilter injects orgao_parceiro_id restrictions into the context.
// Without an empregabilidade editor/admin role, injects an empty list (no results).
func VagaListFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdmin(c) && !HasRole(c, "go:empregabilidade:admin") {
			if isEmpEditor(c) {
				secretariaIDs := GetUserSecretariaOrgaoIDs(c)
				if len(secretariaIDs) > 0 {
					c.Set(vagaOrgaoParceiroIDsKey, secretariaIDs)
				} else if orgaoID := GetUserOrgaoID(c); orgaoID != "" {
					c.Set(vagaOrgaoParceiroIDsKey, []string{orgaoID})
				} else {
					c.Set(vagaOrgaoParceiroIDsKey, []string{})
				}
			} else {
				c.Set(vagaOrgaoParceiroIDsKey, []string{})
			}
		}
		c.Next()
	}
}

// GetVagaOrgaoParceiroIDs returns the orgao_parceiro_id list injected by VagaListFilter.
// Returns nil when no restriction applies (admin / empregabilidade:admin).
func GetVagaOrgaoParceiroIDs(c *gin.Context) []string {
	if v, exists := c.Get(vagaOrgaoParceiroIDsKey); exists {
		if ids, ok := v.([]string); ok {
			return ids
		}
	}
	return nil
}

// VagaOrgaoInjector injects allowed orgao IDs for write operations (POST).
// Requires emp admin bypass or an editor role with secretaria/orgao scope.
func VagaOrgaoInjector() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) || HasRole(c, "go:empregabilidade:admin") {
			c.Next()
			return
		}

		if isEmpEditor(c) {
			secretariaIDs := GetUserSecretariaOrgaoIDs(c)
			if len(secretariaIDs) > 0 {
				c.Set(vagaAllowedOrgaosKey, secretariaIDs)
				c.Next()
				return
			}
			if orgaoID := GetUserOrgaoID(c); orgaoID != "" {
				c.Set(vagaAllowedOrgaosKey, []string{orgaoID})
				c.Next()
				return
			}
			c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para criar: nenhuma secretaria associada"})
			c.Abort()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para criar: órgão não identificado"})
		c.Abort()
	}
}

// GetVagaAllowedOrgaos returns the allowed orgao IDs injected by VagaOrgaoInjector.
// Returns nil when no restriction applies (admin / empregabilidade:admin).
func GetVagaAllowedOrgaos(c *gin.Context) []string {
	if v, exists := c.Get(vagaAllowedOrgaosKey); exists {
		if ids, ok := v.([]string); ok {
			return ids
		}
	}
	return nil
}
