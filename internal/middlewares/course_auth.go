package middlewares

import (
	"context"
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	courseLoadedCursoKey = "course_loaded_curso"
)

func isCursosEditor(c *gin.Context) bool {
	return HasRole(c, "go:cursos:editor")
}

func hasCourseOrgaoScope(c *gin.Context) bool {
	return len(GetUserSecretariaOrgaoIDs(c)) > 0 || GetUserOrgaoID(c) != ""
}

// courseOrgaoInScope reports whether orgaoID is within the caller's authorized
// scope, considering both the CPF→secretaria mapping and a single go:orgao role.
// This mirrors hasCourseOrgaoScope (either source grants access) so that a caller
// authorized by CourseAuthorization is never denied by CourseOwnershipCheck for a
// course in the other source's scope.
func courseOrgaoInScope(c *gin.Context, orgaoID string) bool {
	if slices.Contains(GetUserSecretariaOrgaoIDs(c), orgaoID) {
		return true
	}
	userOrgao := GetUserOrgaoID(c)
	return userOrgao != "" && userOrgao == orgaoID
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
			if courseOrgaoInScope(c, orgaoID) {
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

const enrollmentListSelfAccessKey = "enrollment_list_self_access"

func digitsOnly(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			b = append(b, s[i])
		}
	}
	return string(b)
}

// IsEnrollmentListSelfAccess reports whether CourseEnrollmentListAccess allowed
// the request because search CPF matches the JWT CPF (citizen self-read).
func IsEnrollmentListSelfAccess(c *gin.Context) bool {
	v, ok := c.Get(enrollmentListSelfAccessKey)
	if !ok {
		return false
	}
	flag, _ := v.(bool)
	return flag
}

// CourseEnrollmentListAccess authorizes GET /courses/:courseId/enrollments.
//
// Allows:
//   - admin / go:cursos:casa_civil / go:cursos:editor with org ownership (same as
//     CourseAuthorization + CourseOwnershipCheck)
//   - citizen self-read when ?search= matches the JWT CPF (digits-only compare)
//
// Self-read sets IsEnrollmentListSelfAccess so the handler can omit the
// course-wide enrollment summary (search already scopes the list).
func CourseEnrollmentListAccess(loader courseLoaderFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		userCPF := GetUserCPF(c)
		search := c.Query("search")
		if userCPF != "" && search != "" && digitsOnly(search) == digitsOnly(userCPF) {
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
			c.Set(enrollmentListSelfAccessKey, true)
			c.Next()
			return
		}

		if IsAdmin(c) || HasRole(c, "go:cursos:casa_civil") {
			CourseOwnershipCheck(loader)(c)
			return
		}

		if isCursosEditor(c) && hasCourseOrgaoScope(c) {
			CourseOwnershipCheck(loader)(c)
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar este recurso"})
		c.Abort()
	}
}
