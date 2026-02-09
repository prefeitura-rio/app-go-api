package router

import "github.com/gin-gonic/gin"

func registerPublicRoutes(r *gin.Engine, deps *Dependencies) {
	apiPublic := r.Group("/api/public")
	{
		apiPublic.GET("/courses", deps.CourseHandler.List)
		apiPublic.GET("/courses/:courseId", deps.CourseHandler.GetByID)

		apiPublic.GET("/oportunidades-mei", deps.OportunidadeMEIHandler.List)
		apiPublic.GET("/oportunidades-mei/:id", deps.OportunidadeMEIHandler.GetByID)

		empPublic := apiPublic.Group("/empregabilidade")
		{
			empPublic.GET("/vagas", deps.EmpVagaHandler.PublicList)
			empPublic.GET("/vagas/:id", deps.EmpVagaHandler.PublicGetByID)
		}
	}
}
