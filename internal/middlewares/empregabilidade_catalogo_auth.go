package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const RoleEmpregabilidadeAdmin = "go:empregabilidade:admin"

// EmpregabilidadeCatalogoAuthorization restringe alterações nos catálogos
// mestres de empregabilidade a administradores.
func EmpregabilidadeCatalogoAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) || HasRole(c, RoleEmpregabilidadeAdmin) {
			c.Next()
			return
		}

		c.JSON(
			http.StatusForbidden,
			gin.H{"error": "Sem permissão para alterar os catálogos de empregabilidade"},
		)
		c.Abort()
	}
}
