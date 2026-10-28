package middlewares

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CorsAllowedOriginsEnv é a variável de ambiente com a allowlist de origens
// (separadas por vírgula). Vazia por padrão: nenhuma origem cross-origin é permitida.
const CorsAllowedOriginsEnv = "CORS_ALLOWED_ORIGINS"

// ParseAllowedOrigins converte uma lista separada por vírgula em origens normalizadas.
// O wildcard "*" é ignorado deliberadamente.
func ParseAllowedOrigins(raw string) []string {
	var origins []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		if o == "" || o == "*" {
			continue
		}
		origins = append(origins, o)
	}
	return origins
}

// CorsMiddleware configura CORS usando a allowlist de CORS_ALLOWED_ORIGINS (padrão: vazia).
func CorsMiddleware() gin.HandlerFunc {
	return CorsMiddlewareWithOrigins(ParseAllowedOrigins(os.Getenv(CorsAllowedOriginsEnv)))
}

// CorsMiddlewareWithOrigins configura CORS permitindo apenas as origens informadas.
func CorsMiddlewareWithOrigins(allowed []string) gin.HandlerFunc {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		allowedSet[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		originAllowed := false
		if origin != "" {
			_, originAllowed = allowedSet[origin]
		}

		c.Writer.Header().Add("Vary", "Origin")

		if originAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
