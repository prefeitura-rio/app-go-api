package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// AppConfig contém todas as configurações da aplicação
type AppConfig struct {
	App        AppSettings
	Database   DatabaseSettings
	Server     ServerSettings
	JWT        JWTSettings
	Swagger    SwaggerSettings
	TypeSense  TypeSenseSettings
	Migrations MigrationSettings
}

// AppSettings define configurações gerais da aplicação
type AppSettings struct {
	Environment string
	Debug       bool
	LogLevel    string
	APIPrefix   string
}

// DatabaseSettings define configurações do banco de dados
type DatabaseSettings struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	Timezone string
}

// ServerSettings define configurações do servidor HTTP
type ServerSettings struct {
	Host string
	Port int
}

// JWTSettings define configurações de autenticação
type JWTSettings struct {
	Secret    string
	ExpiresIn string
}

// SwaggerSettings define configurações da documentação Swagger
type SwaggerSettings struct {
	Host string
}

// TypeSenseSettings define configurações do TypeSense
type TypeSenseSettings struct {
	Protocol string
	Host     string
	Port     int
	APIKey   string
}

// MigrationSettings define configurações para migrações
type MigrationSettings struct {
	Run bool
}

// Singleton instance
var (
	config        *AppConfig
	ErrNoHost     = errors.New("host não pode estar vazio")
	ErrInvalidPort = errors.New("porta deve ser maior que zero")
)

// Construtor de DSN para conexão PostgreSQL
func (db *DatabaseSettings) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode, db.Timezone,
	)
}

// IsDevelopment verifica se o ambiente é de desenvolvimento
func (a *AppSettings) IsDevelopment() bool {
	return strings.ToLower(a.Environment) == "development"
}

// IsProduction verifica se o ambiente é de produção
func (a *AppSettings) IsProduction() bool {
	return strings.ToLower(a.Environment) == "production"
}

// IsTest verifica se o ambiente é de teste
func (a *AppSettings) IsTest() bool {
	return strings.ToLower(a.Environment) == "test"
}

// Validate valida as configurações do banco de dados
func (db *DatabaseSettings) Validate() error {
	if db.Host == "" {
		return ErrNoHost
	}
	if db.Port <= 0 {
		return ErrInvalidPort
	}
	return nil
}

// Validate valida as configurações do servidor
func (s *ServerSettings) Validate() error {
	if s.Host == "" {
		return ErrNoHost
	}
	if s.Port <= 0 {
		return ErrInvalidPort
	}
	return nil
}

// Validate valida todas as configurações
func (c *AppConfig) Validate() error {
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("configuração de banco de dados inválida: %w", err)
	}
	
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("configuração de servidor inválida: %w", err)
	}
	
	return nil
}

// Load carrega configurações de variáveis de ambiente e arquivo .env
func Load() (*AppConfig, error) {
	// Inicializa o Viper
	v := viper.New()
	v.AutomaticEnv()
	
	// Configura leitura de arquivo .env
	v.SetConfigType("env")
	v.SetConfigName(".env")
	v.AddConfigPath(".")
	
	// Ler arquivo .env (ignorar erro se não existir)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Aviso: erro ao ler arquivo .env: %v", err)
		}
	}
	
	// Carregar configurações
	appConfig := &AppConfig{
		App: AppSettings{
			Environment: getEnv(v, "APP_ENV", "development"),
			Debug:       getBool(v, "DEBUG", true),
			LogLevel:    getEnv(v, "LOG_LEVEL", "debug"),
			APIPrefix:   getEnv(v, "API_PREFIX", "/api/v1"),
		},
		Database: DatabaseSettings{
			Host:     getEnv(v, "DB_HOST", "localhost"),
			Port:     getInt(v, "DB_PORT", 5432),
			User:     getEnv(v, "DB_USER", "postgres"),
			Password: getEnv(v, "DB_PASSWORD", "postgres"),
			Name:     getEnv(v, "DB_NAME", "app_go_api"),
			SSLMode:  getEnv(v, "DB_SSL_MODE", "disable"),
			Timezone: getEnv(v, "DB_TIMEZONE", "UTC"),
		},
		Server: ServerSettings{
			Host: getEnv(v, "SERVER_HOST", "0.0.0.0"),
			Port: getInt(v, "SERVER_PORT", 8080),
		},
		JWT: JWTSettings{
			Secret:    getEnv(v, "JWT_SECRET", "sua_chave_secreta_aqui"),
			ExpiresIn: getEnv(v, "JWT_EXPIRATION", "24h"),
		},
		Swagger: SwaggerSettings{
			Host: getEnv(v, "SWAGGER_HOST", "localhost:8080"),
		},
		TypeSense: TypeSenseSettings{
			Protocol: getEnv(v, "TYPESENSE_PROTOCOL", "http"),
			Host:     getEnv(v, "TYPESENSE_HOST", "localhost"),
			Port:     getInt(v, "TYPESENSE_PORT", 8108),
			APIKey:   getEnv(v, "TYPESENSE_API_KEY", ""),
		},
		Migrations: MigrationSettings{
			Run: getBool(v, "RUN_MIGRATIONS", false),
		},
	}
	
	// Validar configurações
	if err := appConfig.Validate(); err != nil {
		return nil, err
	}
	
	// Exibir configurações se em modo debug
	if appConfig.App.Debug {
		logConfig(appConfig)
	}
	
	return appConfig, nil
}

// Get retorna a instância singleton da configuração
func Get() *AppConfig {
	if config == nil {
		var err error
		config, err = Load()
		if err != nil {
			panic(fmt.Sprintf("erro fatal ao carregar configurações: %v", err))
		}
	}
	return config
}

// Funções auxiliares para obter valores do Viper com fallback

func getEnv(v *viper.Viper, key, defaultValue string) string {
	if value := v.GetString(key); value != "" {
		return value
	}
	// Fallback para os.Getenv diretamente, para casos onde viper falha
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getInt(v *viper.Viper, key string, defaultValue int) int {
	if v.IsSet(key) {
		return v.GetInt(key)
	}
	// Fallback para os.Getenv diretamente
	if value := os.Getenv(key); value != "" {
		if val, err := fmt.Sscanf(value, "%d", new(int)); err == nil && val > 0 {
			return val
		}
	}
	return defaultValue
}

func getBool(v *viper.Viper, key string, defaultValue bool) bool {
	if v.IsSet(key) {
		return v.GetBool(key)
	}
	// Fallback para os.Getenv diretamente
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

// Função de logging para depuração
func logConfig(config *AppConfig) {
	log.Println("===== Configurações Carregadas =====")
	log.Printf("Ambiente: %s (Debug: %v)", config.App.Environment, config.App.Debug)
	log.Printf("Servidor: %s:%d", config.Server.Host, config.Server.Port)
	log.Printf("Banco de Dados: %s:%d/%s (User: %s, SSL: %s)", 
		config.Database.Host, config.Database.Port, config.Database.Name, 
		config.Database.User, config.Database.SSLMode)
	log.Println("====================================")
} 