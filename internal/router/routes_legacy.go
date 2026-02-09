package router

import "github.com/gin-gonic/gin"

func registerLegacyRoutes(apiV1 *gin.RouterGroup, deps *Dependencies) {
	registerCRUDRoutes(apiV1.Group("/empregos"), deps.EmpregoHandler)
	registerCRUDRoutes(apiV1.Group("/acessibilidades"), deps.AcessibilidadeHandler)
	registerCRUDRoutes(apiV1.Group("/categorias"), deps.CategoriaHandler)
	registerCRUDRoutes(apiV1.Group("/empresas"), deps.EmpresaHandler)
	registerCRUDRoutes(apiV1.Group("/escolaridades"), deps.EscolaridadeHandler)
	registerCRUDRoutes(apiV1.Group("/instituicoes"), deps.InstituicaoHandler)

	// Typesense
	if deps.TypesenseHandler != nil {
		apiV1.POST("/typesense/multi-search", deps.TypesenseHandler.SearchMultiCollection)
		apiV1.POST("/typesense/cursos/search", deps.TypesenseHandler.SearchCursos)
		apiV1.POST("/typesense/empregos/search", deps.TypesenseHandler.SearchEmpregos)
		apiV1.POST("/typesense/collections/:collection/documents/search", deps.TypesenseHandler.SearchDocuments)
	}
}

type crudHandlerInterface interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	GetByID(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func registerCRUDRoutes(group *gin.RouterGroup, handler crudHandlerInterface) {
	group.POST("", handler.Create)
	group.GET("", handler.List)
	group.GET("/:id", handler.GetByID)
	group.PUT("/:id", handler.Update)
	group.DELETE("/:id", handler.Delete)
}
