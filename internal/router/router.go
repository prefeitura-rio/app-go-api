package router

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/jobs"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/wire"
	"github.com/prefeitura-rio/app-go-api/internal/workers"

	_ "github.com/prefeitura-rio/app-go-api/docs"
)

// SetupRouter initializes the Gin engine with all routes using Wire-managed dependencies.
// ctx controls the lifetime of all background workers started here.
func SetupRouter(ctx context.Context, cfg *config.AppConfig) (*gin.Engine, error) {
	app, err := wire.InitializeApplication(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize application container: %w", err)
	}

	r := gin.Default()
	r.Use(gin.Recovery(), gin.Logger())

	if cfg.Tracing.Enabled {
		r.Use(otelgin.Middleware(cfg.Tracing.ServiceName))
	}
	r.Use(middlewares.TimeoutMiddleware(time.Duration(cfg.Server.RequestTimeout) * time.Second))
	r.Use(middlewares.CorsMiddleware())

	r.GET("/docs/*any", middlewares.DynamicSwaggerHandler())
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// Citizen sync worker (optional, requires Keycloak)
	if cfg.CitizenSync.Enabled && app.TokenManager != nil {
		w := workers.NewCitizenSyncWorker(app.RMIClient, app.CitizenSnapshotRepo, app.TokenManager, &cfg.CitizenSync)
		app.InscricaoService.SetCitizenDataFetcher(w)
		app.EmpCandidaturaService.SetCitizenDataFetcher(w)
		go func() {
			if err := w.Start(ctx); err != nil && err != context.Canceled {
				log.Printf("[Router] Citizen sync worker stopped: %v", err)
			}
		}()
		log.Println("[Router] Citizen sync worker initialized and started successfully")
	} else if cfg.CitizenSync.Enabled {
		log.Println("[Router] Citizen sync worker disabled - Keycloak not configured")
	} else {
		log.Println("[Router] Citizen sync worker disabled (CITIZEN_SYNC_ENABLED=false)")
	}

	go func() {
		if err := app.EmailWorker.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("[Router] Email worker stopped: %v", err)
		}
	}()
	log.Println("[Router] Email worker started")

	jobs.InitializeJobProcessor(app.DB, app.JobRepo, app.InscricaoRepo, app.CursoRepo)

	apiV1 := r.Group("/api/v1")
	apiV1.Use(middlewares.ExtractUserContext(cfg.Heimdall.BaseURL))

	// Resolve secretaria orgao_ids via CPF-Secretaria mapping in app-rmi.
	// Only enabled when both TokenManager and RMIClient are available.
	if app.TokenManager != nil {
		resolver := func(ctx context.Context, cpf string) ([]string, error) {
			token, err := app.TokenManager.GetToken(ctx)
			if err != nil {
				return nil, err
			}
			cdUAs, err := app.RMIClient.GetCPFSecretarias(ctx, token, cpf)
			if err != nil || len(cdUAs) == 0 {
				return []string{}, nil
			}
			return app.OrgaoSnapshotRepo.GetOrgaoIDsByCdUAs(ctx, cdUAs)
		}
		apiV1.Use(middlewares.ExtractSecretariaOrgaoIDs(resolver))
	}

	apiPublic := r.Group("/api/public")

	registerCoreRoutes(apiV1, apiPublic, app)
	registerEmpregabilidadeRoutes(apiV1, apiPublic, app)

	return r, nil
}
