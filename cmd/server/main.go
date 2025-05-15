package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/router"
)

// @title API Go
// @version 1.0
// @description API de serviços para aplicativos da Prefeitura do Rio
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://iplan.rio
// @contact.email frederico.zolio@prefeitura.rio

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /
// @schemes http https
func main() {
	// Imprimir todas as variáveis de ambiente para debug
	if os.Getenv("DEBUG") == "true" {
		log.Println("Variáveis de ambiente disponíveis:")
		for _, env := range os.Environ() {
			log.Println(env)
		}
	}

	// Carrega configurações
	log.Println("Carregando configurações...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Conecta ao banco de dados usando GORM
	dsn := cfg.Database.DSN()
	
	// Configura o logger do GORM de acordo com o ambiente
	var logMode logger.LogLevel
	if cfg.App.IsDevelopment() {
		logMode = logger.Info
	} else {
		logMode = logger.Silent
	}
	
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	}
	
	log.Println("Tentando conectar ao banco de dados...")
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	log.Println("Conexão com o banco de dados estabelecida com sucesso!")

	// Auto-migração do GORM (opcional)
	if cfg.Migrations.Run {
		log.Println("Iniciando auto-migração...")
		err = db.AutoMigrate(
			&models.Curso{},
			&models.Emprego{},
			&models.Categoria{},
			&models.Acessibilidade{},
			&models.Orgao{},
			&models.InstituicaoEnsino{},
			&models.Empresa{},
			&models.Escolaridade{},
			&models.CursoCategoria{},
			&models.CursoAcessibilidade{},
		)
		if err != nil {
			log.Fatalf("Erro ao executar auto-migração: %v", err)
		}
		log.Println("Auto-migração concluída com sucesso!")
	}

	// Configura o router
	r := router.SetupRouter(db)

	// Inicia o servidor
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Servidor iniciado em %s", addr)
	
	if err := r.Run(addr); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
