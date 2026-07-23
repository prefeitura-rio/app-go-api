package middlewares

import (
	"context"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// vagaOrgaoInScope reports whether orgaoID is within the caller's authorized
// scope, considering both the CPF→secretaria mapping and a single go:orgao role.
// This mirrors hasVagaOrgaoScope (either source grants access) so that a caller
// authorized by VagaAuthorization is never denied by VagaOwnershipCheck for a
// vaga in the other source's scope.
func vagaOrgaoInScope(c *gin.Context, orgaoID string) bool {
	if slices.Contains(GetUserSecretariaOrgaoIDs(c), orgaoID) {
		return true
	}
	userOrgao := GetUserOrgaoID(c)
	return userOrgao != "" && userOrgao == orgaoID
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
//
// admin / go:empregabilidade:admin: sem restrição (nada injetado, veem tudo).
// editor com escopo (secretaria ou órgão): restrito aos seus orgaos.
// Qualquer outro caso — inclusive quem tem apenas secretaria mapeada mas NENHUMA
// role de empregabilidade — recebe lista vazia ([]string{}), resultando em zero
// resultados no handler. Isso é intencional: secretaria sozinha não libera acesso
// (mesma regra da VagaAuthorization); listar tudo aqui vazaria vagas de todos os
// órgãos para quem não deveria enxergá-las.
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

// VagaOwnershipInfo carries the vaga fields VagaOwnershipCheck needs.
type VagaOwnershipInfo struct {
	OrgaoParceiroID string
}

// VagaLoaderFunc loads ownership info for a vaga. Returns nil when the vaga does
// not exist.
type VagaLoaderFunc func(ctx context.Context, id uuid.UUID) (*VagaOwnershipInfo, error)

// VagaOwnershipCheck enforces per-vaga org ownership on mutation routes.
//
// admin / go:empregabilidade:admin bypass every check. Any other caller must be
// an empregabilidade editor whose secretaria/orgao scope contains the vaga's
// orgao_parceiro.
//
// A regra de "quem pode editar vaga já publicada" NÃO fica aqui: ela depende do
// status da vaga (que o gateway Heimdall não conhece) e por isso é aplicada no
// handler (VagaHandler.Update), ativa em qualquer ambiente. Este middleware,
// como os demais de autorização, roda só em development/test (no-op em prod —
// ver routes_emp.go / noOpHandler).
func VagaOwnershipCheck(loader VagaLoaderFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) || HasRole(c, "go:empregabilidade:admin") {
			c.Next()
			return
		}

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
			c.Abort()
			return
		}

		info, err := loader(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar vaga"})
			c.Abort()
			return
		}
		if info == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
			c.Abort()
			return
		}

		if !isEmpEditor(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar este recurso"})
			c.Abort()
			return
		}

		if !vagaOrgaoInScope(c, info.OrgaoParceiroID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: vaga não pertence à sua secretaria"})
			c.Abort()
			return
		}

		c.Next()
	}
}
