package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"

	_ "github.com/prefeitura-rio/app-go-api/docs"
)

func SetupRouter(db *gorm.DB, cfg *config.AppConfig) *gin.Engine {
	r := gin.Default()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	if cfg.Tracing.Enabled {
		r.Use(otelgin.Middleware(cfg.Tracing.ServiceName))
	}

	r.Use(middlewares.TimeoutMiddleware(time.Duration(cfg.Server.RequestTimeout) * time.Second))
	r.Use(middlewares.CorsMiddleware())

	r.GET("/docs/*any", middlewares.DynamicSwaggerHandler())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	deps := InitDependencies(db, cfg)

	apiV1 := r.Group("/api/v1")
	apiV1.Use(middlewares.ExtractUserContext())

	registerCourseRoutes(apiV1, deps)
	registerLegacyRoutes(apiV1, deps)
	registerMEIRoutes(apiV1, deps)
	registerEmpregabilidadeRoutes(apiV1, deps)
	registerPublicRoutes(r, deps)

	return r
}
