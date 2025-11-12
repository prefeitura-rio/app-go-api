// Package config gerencia as configurações da aplicação
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

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
	RMI        RMISettings
	Redis      RedisSettings
}

// AppSettings define configurações gerais da aplicação
type AppSettings struct {
	Environment string
	Debug       bool
	LogLevel    string
	APIPrefix   string
	APIToken    string
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

// GetExpirationDuration retorna a duração de expiração do token
func (j *JWTSettings) GetExpirationDuration() (time.Duration, error) {
	return time.ParseDuration(j.ExpiresIn)
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

// RMISettings define configurações da API RMI (para validação de CNPJs/CNAEs)
type RMISettings struct {
	BaseURL string
}

// RedisSettings define configurações do Redis
type RedisSettings struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Erros comuns de validação
var (
	ErrNoHost      = errors.New("host não pode estar vazio")
	ErrInvalidPort = errors.New("porta deve ser maior que zero")
)

// Singleton instance com proteção para concorrência
var (
	instance *AppConfig
	once     sync.Once
	mu       sync.RWMutex
	v        *viper.Viper
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

// ConfigProvider define uma interface para obter configurações
type ConfigProvider interface {
	GetAppConfig() *AppConfig
	Reload() error
}

// defaultConfigProvider implementa ConfigProvider
type defaultConfigProvider struct {
	config *AppConfig
	viper  *viper.Viper
}

// NewConfigProvider cria uma nova instância de provedor de configuração
func NewConfigProvider() (ConfigProvider, error) {
	config, err := Load()
	if err != nil {
		return nil, err
	}

	return &defaultConfigProvider{
		config: config,
		viper:  v,
	}, nil
}

// GetAppConfig retorna a configuração atual
func (p *defaultConfigProvider) GetAppConfig() *AppConfig {
	return p.config
}

// Reload recarrega as configurações
func (p *defaultConfigProvider) Reload() error {
	if err := p.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("erro ao recarregar configurações: %w", err)
		}
	}

	config, err := Load()
	if err != nil {
		return err
	}

	p.config = config
	return nil
}

// Initialize inicializa o viper e carrega as configurações iniciais
func Initialize() error {
	v = viper.New()
	v.AutomaticEnv()

	// Configura leitura de arquivo .env
	v.SetConfigType("env")
	v.SetConfigName(".env")
	v.AddConfigPath(".")

	// Configuração de observação de alterações no arquivo
	v.WatchConfig()

	// Ler arquivo .env (ignorar erro se não existir)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Aviso: erro ao ler arquivo .env: %v", err)
		}
	}

	// Carregar configuração inicial
	_, err := Get()
	return err
}

// Load carrega configurações de variáveis de ambiente e arquivo .env
func Load() (*AppConfig, error) {
	if v == nil {
		if err := Initialize(); err != nil {
			return nil, err
		}
	}

	// Carregar configurações
	appConfig := &AppConfig{
		App: AppSettings{
			Environment: getEnv(v, "APP_ENV", "development"),
			Debug:       getBool(v, "APP_DEBUG", true),
			LogLevel:    getEnv(v, "LOG_LEVEL", "info"),
			APIPrefix:   getEnv(v, "API_PREFIX", "/api"),
			APIToken:    getEnv(v, "API_TOKEN", ""),
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
		RMI: RMISettings{
			BaseURL: getEnv(v, "RMI_BASE_URL", ""),
		},
		Redis: RedisSettings{
			Host:     getEnv(v, "REDIS_HOST", "localhost"),
			Port:     getInt(v, "REDIS_PORT", 6379),
			Password: getEnv(v, "REDIS_PASSWORD", ""),
			DB:       getInt(v, "REDIS_DB", 0),
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

// Get retorna a instância singleton da configuração de forma thread-safe
func Get() (*AppConfig, error) {
	once.Do(func() {
		cfg, err := Load()
		if err != nil {
			log.Printf("Erro ao carregar configurações: %v", err)
			return
		}
		instance = cfg
	})

	if instance == nil {
		return nil, errors.New("falha ao inicializar configurações")
	}

	mu.RLock()
	defer mu.RUnlock()
	return instance, nil
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

// logConfig imprime as configurações atuais para depuração
func logConfig(config *AppConfig) {
	log.Println("=== Configurações Carregadas ====")
	log.Printf("Ambiente: %s", config.App.Environment)
	log.Printf("Debug: %v", config.App.Debug)
	log.Printf("Log Level: %s", config.App.LogLevel)
	log.Printf("API Prefix: %s", config.App.APIPrefix)
	log.Printf("DB: %s@%s:%d/%s", config.Database.User, config.Database.Host, config.Database.Port, config.Database.Name)
	log.Printf("Server: %s:%d", config.Server.Host, config.Server.Port)
	log.Printf("Migrations: %v", config.Migrations.Run)
	log.Println("====================================")
}
