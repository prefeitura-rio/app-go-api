package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Roles do banco de currículos. Por enquanto as duas só dão acesso de leitura;
// a separação existe para as funcionalidades futuras do módulo.
const (
	RoleCurriculosAdmin  = "go:curriculos:admin"
	RoleCurriculosEditor = "go:curriculos:editor"
)

// BancoCurriculosAuthorization libera o banco de currículos para admins totais e
// para quem tem uma das roles do módulo. Diferente das autorizações de cursos e
// vagas, é aplicada sempre, sem depender de RBAC_ENABLED: sem ela, qualquer
// usuário autenticado — inclusive cidadãos — listaria todos os currículos.
func BancoCurriculosAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) || HasRole(c, RoleCurriculosAdmin) || HasRole(c, RoleCurriculosEditor) {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar o banco de currículos"})
		c.Abort()
	}
}
