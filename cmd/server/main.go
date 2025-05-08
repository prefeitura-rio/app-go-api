package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Arquivo .env não encontrado")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	log.Printf("Servidor iniciado na porta %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
