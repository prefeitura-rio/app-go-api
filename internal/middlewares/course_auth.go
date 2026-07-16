package middlewares

import (
	"context"
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	courseLoadedCursoKey   = "course_loaded_curso"
	courseAllowedOrgaosKey = "course_allowed_orgaos"
)

func isCursosEditor(c *gin.Context) bool {
	return HasRole(c, "go:cursos:editor")
}

func hasCourseOrgaoScope(c *gin.Context) bool {
	return len(GetUserSecretariaOrgaoIDs(c)) > 0 || GetUserOrgaoID(c) != ""
}

// CourseAuthorization allows only total admins, go:cursos:casa_civil, or go:cursos:editor
// with secretaria/orgao scope. Secretaria alone is not enough.
func CourseAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) || HasRole(c, "go:cursos:casa_civil") {
			c.Next()
			return
		}

		if isCursosEditor(c) && hasCourseOrgaoScope(c) {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar este recurso"})
		c.Abort()
	}
}

// CourseListFilter scopes list queries. Without a cursos role, results are empty
// (secretaria alone does not grant access).
func CourseListFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		filters := make(map[string]interface{})
		if !IsAdmin(c) && !HasRole(c, "go:cursos:casa_civil") {
			if isCursosEditor(c) {
				secretariaIDs := GetUserSecretariaOrgaoIDs(c)
				if len(secretariaIDs) > 0 {
					filters["orgao_id IN"] = secretariaIDs
				} else if orgaoID := GetUserOrgaoID(c); orgaoID != "" {
					filters["orgao_id"] = orgaoID
				} else {
					filters["_no_results"] = true
				}
			} else {
				filters["_no_results"] = true
			}
		}
		c.Set("course_filters", filters)
		c.Next()
	}
}

func GetCourseFilters(c *gin.Context) map[string]interface{} {
	if filters, exists := c.Get("course_filters"); exists {
		if f, ok := filters.(map[string]interface{}); ok {
			return f
		}
	}
	return make(map[string]interface{})
}

// CourseOrgaoInjector: only total admins / casa_civil may create courses.
// Editors and secretaria-only users are denied.
func CourseOrgaoInjector() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isCursosEditor(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Editores não têm permissão para criar cursos"})
			c.Abort()
			return
		}

		if IsAdmin(c) || HasRole(c, "go:cursos:casa_civil") {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para criar: órgão não identificado"})
		c.Abort()
	}
}

func GetUserAllowedOrgaos(c *gin.Context) []string {
	if v, exists := c.Get(courseAllowedOrgaosKey); exists {
		if id, ok := v.([]string); ok {
			return id
		}
	}
	return nil
}

type courseLoaderFunc func(ctx context.Context, id int) (orgaoID string, found bool, err error)

func CourseOwnershipCheck(loader courseLoaderFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("courseId"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID do curso inválido"})
			c.Abort()
			return
		}

		orgaoID, found, err := loader(c.Request.Context(), id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar curso: " + err.Error()})
			c.Abort()
			return
		}

		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "Curso não encontrado"})
			c.Abort()
			return
		}

		c.Set(courseLoadedCursoKey, orgaoID)

		if IsAdmin(c) || HasRole(c, "go:cursos:casa_civil") {
			c.Next()
			return
		}

		if isCursosEditor(c) {
			secretariaIDs := GetUserSecretariaOrgaoIDs(c)
			if len(secretariaIDs) > 0 {
				if slices.Contains(secretariaIDs, orgaoID) {
					c.Next()
					return
				}
				c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: curso não pertence à sua secretaria"})
				c.Abort()
				return
			}

			userOrgao := GetUserOrgaoID(c)
			if userOrgao != "" && userOrgao == orgaoID {
				c.Next()
				return
			}
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: curso não pertence à sua secretaria"})
			c.Abort()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar este recurso"})
		c.Abort()
	}
}

func GetLoadedCursoOrgaoID(c *gin.Context) string {
	if v, exists := c.Get(courseLoadedCursoKey); exists {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
