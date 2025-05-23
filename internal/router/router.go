package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/prefeitura-rio/app-go-api/docs" // importação dos docs gerados pelo Swag
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Middleware global
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middlewares.CorsMiddleware())

	// Configuração do Swagger
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Rota de saúde
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1
	apiV1 := r.Group("/api/v1")

	// Inicializando repositórios
	cursoRepo := repository.NewCursoRepository(db)
	empregoRepo := repository.NewEmpregoRepository(db)

	// Inicializando serviços
	cursoService := services.NewCursoService(cursoRepo)
	empregoService := services.NewEmpregoService(empregoRepo)

	// Inicializando handlers
	cursoHandler := v1.NewCursoHandler(cursoService)
	empregoHandler := v1.NewEmpregoHandler(empregoService)
	typesenseHandler, err := v1.NewTypesenseHandler()
	if err != nil {
		fmt.Printf("Erro ao inicializar o Typesense: %v\n", err)
	}

	// Rotas de cursos
	cursos := apiV1.Group("/cursos")
	{
		cursos.POST("", cursoHandler.Create)
		cursos.GET("", cursoHandler.List)
		cursos.GET("/:id", cursoHandler.GetByID)
		cursos.PUT("/:id", cursoHandler.Update)
		cursos.DELETE("/:id", cursoHandler.Delete)
	}

	// Rotas de empregos
	empregos := apiV1.Group("/empregos")
	{
		empregos.POST("", empregoHandler.Create)
		empregos.GET("", empregoHandler.List)
		empregos.GET("/:id", empregoHandler.GetByID)
		empregos.PUT("/:id", empregoHandler.Update)
		empregos.DELETE("/:id", empregoHandler.Delete)
	}

	// Rotas de Typesense
	if typesenseHandler != nil {
		apiV1.POST("/typesense/multi-search", typesenseHandler.SearchMultiCollection)
		typesenseCollections := apiV1.Group("/typesense/collections")
		typesenseCollections.POST("/:collection/documents/search", typesenseHandler.SearchDocuments)
	}

	return r
}
