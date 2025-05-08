package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/docs"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/router"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title        API Go Prefeitura do Rio de Janeiro
// @version      1.0
// @description  API REST para a Prefeitura do Rio de Janeiro
// @contact.name   API Support
// @contact.email  support@prefeitura.rio
// @BasePath      /api/v1
func main() {
	// Carregar configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Definir o modo do Gin com base nas configurações
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Inicializar o router
	r := gin.Default()
	
	// Configurar rotas
	router.SetupRouter(r)

	// Definir host do Swagger usando as configurações
	swaggerHost := cfg.GetSwaggerHost()
	docs.SwaggerInfo.Host = swaggerHost

	// Rota para o Swagger
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Logar informações sobre o servidor
	log.Printf("Aplicação rodando em modo: %s", cfg.AppEnv)
	log.Printf("Servidor iniciado na porta %s", cfg.Port)
	log.Printf("Documentação Swagger disponível em: http://%s/docs/index.html", swaggerHost)
	
	// Iniciar o servidor
	if err := r.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
