package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	"github.com/prefeitura-rio/app-go-api/docs" // importação necessária para swagger
)

// DynamicSwaggerHandler retorna um handler para o Swagger UI que usa o host da requisição atual
func DynamicSwaggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obter o host da requisição atual
		host := c.Request.Host

		// Salvar o host original para restaurar depois
		originalHost := docs.SwaggerInfo.Host

		// Configurar o Swagger para usar o host atual
		docs.SwaggerInfo.Host = host

		// Preparar um handler personalizado que restaura o host original depois
		handler := ginSwagger.WrapHandler(swaggerFiles.Handler)

		// Chamar o handler do Swagger
		handler(c)

		// Restaurar o host original após a chamada
		docs.SwaggerInfo.Host = originalHost
	}
}
