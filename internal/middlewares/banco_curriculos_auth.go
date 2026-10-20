package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Roles do banco de currículos. As duas dão acesso de leitura à base. Associar
// candidatos a vagas exige também um perfil de empregabilidade.
const (
	RoleCurriculosAdmin  = "go:curriculos:admin"
	RoleCurriculosEditor = "go:curriculos:editor"
)

const bancoCurriculosOrgaosPermitidosKey = "banco_curriculos_orgaos_permitidos"

// BancoCurriculosAuthorization libera o banco de currículos para admins totais e
// para quem tem uma das roles do módulo. O acesso é liberado pessoa a pessoa,
// separado do módulo de empregabilidade. Diferente das autorizações de cursos e
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

// BancoCurriculosAssociacaoAuthorization roda depois de
// BancoCurriculosAuthorization: quem associa precisa de acesso ao banco e de um
// perfil de empregabilidade. Libera admins totais e go:empregabilidade:admin,
// sem restrição de órgão,
// e para editores de empregabilidade com secretaria ou órgão, restritos às
// vagas desses órgãos (ver GetBancoCurriculosOrgaosPermitidos). Como a
// leitura, vale mesmo com RBAC_ENABLED=false.
func BancoCurriculosAssociacaoAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAdmin(c) || HasRole(c, "go:empregabilidade:admin") {
			c.Next()
			return
		}

		if isEmpEditor(c) {
			if secretariaIDs := GetUserSecretariaOrgaoIDs(c); len(secretariaIDs) > 0 {
				c.Set(bancoCurriculosOrgaosPermitidosKey, secretariaIDs)
				c.Next()
				return
			}
			if orgaoID := GetUserOrgaoID(c); orgaoID != "" {
				c.Set(bancoCurriculosOrgaosPermitidosKey, []string{orgaoID})
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Sem permissão para associar candidatos a vagas"})
		c.Abort()
	}
}

// GetBancoCurriculosOrgaosPermitidos devolve os órgãos cujas vagas o operador
// pode associar. Nil significa sem restrição (admins).
func GetBancoCurriculosOrgaosPermitidos(c *gin.Context) []string {
	if v, exists := c.Get(bancoCurriculosOrgaosPermitidosKey); exists {
		if ids, ok := v.([]string); ok {
			return ids
		}
	}
	return nil
}
