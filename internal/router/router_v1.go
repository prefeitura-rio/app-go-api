package router

import (
	"log"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
)

// SetupRouter configura todas as rotas da API
func SetupRouter(r *gin.Engine) {
	// Rota de saúde na raiz
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Grupo de rotas da API v1
	apiV1 := r.Group("/api/v1")
	{
		// Endpoint de health check da v1
		apiV1.GET("/health", v1.HealthHandler)

		// Criar uma instância do handler do Typesense
		typesenseHandler, err := v1.NewTypesenseHandler()
		if err != nil {
			log.Printf("AVISO: Não foi possível inicializar o serviço Typesense: %v", err)
		} else {
			// Rotas para coleções do Typesense
			collections := apiV1.Group("/typesense/collections")
			{
				collections.POST("", typesenseHandler.CreateCollection)
				collections.GET("", typesenseHandler.ListCollections)
				collections.GET("/:name", typesenseHandler.GetCollection)
				collections.DELETE("/:name", typesenseHandler.DeleteCollection)
			}

			// Rotas para documentos do Typesense
			documents := apiV1.Group("/typesense/collections/:collection/documents")
			{
				documents.POST("", typesenseHandler.UpsertDocument)
				documents.GET("/:id", typesenseHandler.GetDocument)
				documents.DELETE("/:id", typesenseHandler.DeleteDocument)
				documents.POST("/search", typesenseHandler.SearchDocuments)
				documents.POST("/import", typesenseHandler.ImportDocuments)
			}
		}
	}
}