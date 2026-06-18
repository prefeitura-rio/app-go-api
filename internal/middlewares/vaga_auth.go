package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const vagaOrgaoParceiroIDsKey = "vaga_orgao_parceiro_ids"

func VagaAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) || HasRole(c, "go:empregabilidade:admin") {
			c.Next()
			return
		}

		secretariaIDs := GetUserSecretariaOrgaoIDs(c)
		if len(secretariaIDs) > 0 {
			c.Next()
			return
		}

		if HasRole(c, "go:empregabilidade:editor") || HasRole(c, "go:empregabilidade:editor_sem_curadoria") {
			if orgaoID := GetUserOrgaoID(c); orgaoID != "" {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar este recurso"})
		c.Abort()
	}
}

// VagaListFilter injects orgao_parceiro_id restrictions into the context.
// Admin and go:empregabilidade:admin see everything (no filter injected).
// Secretaria users are restricted to their mapped orgao IDs.
// Editor users are restricted to their single orgao ID.
func VagaListFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdmin(c) && !HasRole(c, "go:empregabilidade:admin") {
			secretariaIDs := GetUserSecretariaOrgaoIDs(c)
			if secretariaIDs != nil {
				c.Set(vagaOrgaoParceiroIDsKey, secretariaIDs)
			} else if HasRole(c, "go:empregabilidade:editor") || HasRole(c, "go:empregabilidade:editor_sem_curadoria") {
				if orgaoID := GetUserOrgaoID(c); orgaoID != "" {
					c.Set(vagaOrgaoParceiroIDsKey, []string{orgaoID})
				}
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
