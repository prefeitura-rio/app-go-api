package middlewares

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	courseLoadedCursoKey   = "course_loaded_curso"
	courseAllowedOrgaosKey = "course_allowed_orgaos"
)

func CourseAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) || HasRole(c, "go:cursos:casa_civil") {
			c.Next()
			return
		}

		secretariaIDs := GetUserSecretariaOrgaoIDs(c)
		if len(secretariaIDs) > 0 {
			c.Next()
			return
		}

		if HasRole(c, "go:cursos:editor") {
			if orgaoID := GetUserOrgaoID(c); orgaoID != "" {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar este recurso"})
		c.Abort()
	}
}

func CourseListFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		filters := make(map[string]interface{})
		if !IsAdmin(c) && !HasRole(c, "go:cursos:casa_civil") {
			secretariaIDs := GetUserSecretariaOrgaoIDs(c)
			if len(secretariaIDs) > 0 {
				filters["orgao_id IN"] = secretariaIDs
			} else if HasRole(c, "go:cursos:editor") {
				if orgaoID := GetUserOrgaoID(c); orgaoID != "" {
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

func CourseOrgaoInjector() gin.HandlerFunc {
	return func(c *gin.Context) {
		if HasRole(c, "go:cursos:editor") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Editores não têm permissão para criar cursos"})
			c.Abort()
			return
		}

		if IsAdmin(c) || HasRole(c, "go:cursos:casa_civil") {
			c.Next()
			return
		}

		secretariaIDs := GetUserSecretariaOrgaoIDs(c)
		if secretariaIDs != nil {
			if len(secretariaIDs) == 0 {
				c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para criar: nenhuma secretaria associada"})
				c.Abort()
				return
			}
			c.Set(courseAllowedOrgaosKey, secretariaIDs)
			c.Next()
			return
		}

		if HasRole(c, "go:cursos:editor") {
			if orgaoID := GetUserOrgaoID(c); orgaoID != "" {
				c.Set(courseAllowedOrgaosKey, []string{orgaoID})
				c.Next()
				return
			}
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

		secretariaIDs := GetUserSecretariaOrgaoIDs(c)

		if secretariaIDs != nil {
			for _, sid := range secretariaIDs {
				if orgaoID == sid {
					c.Next()
					return
				}
			}
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: curso não pertence à sua secretaria"})
			c.Abort()
			return
		}

		if HasRole(c, "go:cursos:editor") {
			userOrgao := GetUserOrgaoID(c)

			if userOrgao != "" && userOrgao == orgaoID {
				c.Next()
				return
			}
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: curso não pertence à sua secretaria"})
			c.Abort()
			return
		}

		c.Next()
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
