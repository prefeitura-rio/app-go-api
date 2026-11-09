package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CandidaturaBloqueioListAuthorization protege a listagem dos bloqueios de
// candidatura, que traz o CPF dos cidadãos barrados. Libera admins totais,
// go:empregabilidade:admin e editores de Empregabilidade; qualquer outro
// chamador (cidadão, perfis de outros módulos) recebe 403.
//
// Diferente de VagaAuthorization, é aplicada sempre, sem depender de
// RBAC_ENABLED (como o banco de currículos): com o RBAC desligado, o vagaAuth
// vira no-op e qualquer usuário autenticado listaria os CPFs. Com o RBAC
// ligado, VagaAuthorization roda depois e exige também o escopo de
// órgão/secretaria do editor.
func CandidaturaBloqueioListAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) || HasRole(c, "go:empregabilidade:admin") || isEmpEditor(c) {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para acessar este recurso"})
		c.Abort()
	}
}
