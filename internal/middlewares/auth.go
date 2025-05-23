package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/config"
)

var (
	ErrMissingAuthHeader = errors.New("cabeçalho de autorização ausente")
	ErrInvalidAuthFormat = errors.New("formato de autorização inválido")
	ErrInvalidToken     = errors.New("token inválido")
)

// AuthMiddleware verifica se o token de API fornecido é válido
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtém a configuração da aplicação
		cfg, err := config.Get()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "erro ao obter configurações",
			})
			return
		}

		// Verifica se há uma configuração de token definida
		if cfg.App.APIToken == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "autenticação não configurada corretamente",
			})
			return
		}

		// Obtém o header de autorização
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": ErrMissingAuthHeader.Error(),
			})
			return
		}

		// Verifica o formato do header Bearer
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": ErrInvalidAuthFormat.Error(),
			})
			return
		}

		// Valida o token
		token := parts[1]
		if token != cfg.App.APIToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": ErrInvalidToken.Error(),
			})
			return
		}

		c.Next()
	}
}
