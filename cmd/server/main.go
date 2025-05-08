package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/config"
)

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

	router := gin.Default()

	// Rota de verificação de saúde
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"env":    cfg.AppEnv,
		})
	})

	// Logar informações sobre o servidor
	log.Printf("Aplicação rodando em modo: %s", cfg.AppEnv)
	log.Printf("Servidor iniciado na porta %s", cfg.Port)
	
	// Iniciar o servidor
	if err := router.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
