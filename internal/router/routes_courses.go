package router

import "github.com/gin-gonic/gin"

func registerCourseRoutes(apiV1 *gin.RouterGroup, deps *Dependencies) {
	courses := apiV1.Group("/courses")
	{
		courses.POST("", deps.CourseHandler.Create)
		courses.POST("/draft", deps.CourseHandler.CreateDraft)
		courses.PUT("/:courseId", deps.CourseHandler.Update)
		courses.GET("", deps.CourseHandler.List)
		courses.GET("/drafts", deps.CourseHandler.ListDrafts)
		courses.GET("/:courseId", deps.CourseHandler.GetByID)
		courses.DELETE("/:courseId", deps.CourseHandler.Delete)

		// Enrollment endpoints
		courses.POST("/:courseId/enrollments", deps.InscricaoHandler.Create)
		courses.POST("/:courseId/enrollments/manual", deps.InscricaoHandler.CreateManual)
		courses.POST("/:courseId/enrollments/import", deps.InscricaoHandler.Import)
		courses.GET("/:courseId/enrollments", deps.InscricaoHandler.List)
		courses.PUT("/:courseId/enrollments/status", deps.InscricaoHandler.UpdateStatus)
		courses.PUT("/:courseId/enrollments/:enrollmentId", deps.InscricaoHandler.Update)
		courses.PUT("/:courseId/enrollments/:enrollmentId/status", deps.InscricaoHandler.UpdateIndividualStatus)
		courses.GET("/:courseId/enrollments/:enrollmentId", deps.InscricaoHandler.GetByID)
		courses.PUT("/:courseId/enrollments/:enrollmentId/certificate", deps.InscricaoHandler.UpdateCertificate)
		courses.DELETE("/:courseId/enrollments/:enrollmentId", deps.InscricaoHandler.Delete)
	}

	// Job status
	jobsGroup := apiV1.Group("/jobs")
	{
		jobsGroup.GET("/:jobId/status", deps.JobHandler.GetStatus)
	}

	// User courses
	users := apiV1.Group("/users")
	{
		users.GET("/:userId/courses", deps.CourseHandler.ListByUser)
	}

	// Enrollments by CPF
	enrollments := apiV1.Group("/enrollments")
	{
		enrollments.GET("/user/:cpf", deps.InscricaoHandler.ListByUser)
		enrollments.PUT("/:enrollmentId/schedule", deps.InscricaoHandler.ChangeSchedule)
	}
}
