package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CanGerirMEI reporta se o chamador tem um perfil que gere as oportunidades
// MEI do governo e as propostas recebidas: admin total,
// go:empregabilidade:admin ou editor de Empregabilidade (com ou sem
// curadoria). É o mesmo conjunto que gere vagas e candidaturas.
func CanGerirMEI(c *gin.Context) bool {
	return canGerirCandidaturas(c)
}

// MEIGestaoAuthorization protege a gestão das oportunidades MEI (criação,
// edição, publicação, exclusão e rascunhos) e a mudança de status das
// propostas. Libera só os perfis de CanGerirMEI; qualquer outro chamador
// (anônimo, cidadão, perfis de outros módulos) recebe 403.
//
// É aplicada sempre, sem depender de RBAC_ENABLED (DEV-478): as rotas MEI não
// têm outra checagem de autorização.
func MEIGestaoAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CanGerirMEI(c) {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para gerir oportunidades MEI"})
		c.Abort()
	}
}
