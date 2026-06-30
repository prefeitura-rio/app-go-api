package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// EnrollmentAuthorization guards the standalone /enrollments routes.
// It mirrors CourseAuthorization: admin, go:cursos:casa_civil, secretaria users,
// and go:cursos:editor users with a valid orgao_id are allowed through.
// Everyone else receives 403.
func EnrollmentAuthorization() gin.HandlerFunc {
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
