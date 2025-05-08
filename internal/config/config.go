package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config armazena todas as configurações da aplicação
type Config struct {
	AppEnv        string `mapstructure:"APP_ENV"`
	Port          string `mapstructure:"PORT"`
	Debug         bool   `mapstructure:"DEBUG"`
	APIPrefix     string `mapstructure:"API_PREFIX"`
	LogLevel      string `mapstructure:"LOG_LEVEL"`
	Database      DatabaseConfig
	JWT           JWTConfig
	Typesense     TypesenseConfig
	Swagger       SwaggerConfig
}

// DatabaseConfig armazena as configurações relacionadas ao banco de dados
type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSL_MODE"`
	TimeZone string `mapstructure:"DB_TIMEZONE"`
}

// JWTConfig armazena as configurações relacionadas a autenticação JWT
type JWTConfig struct {
	Secret     string        `mapstructure:"JWT_SECRET"`
	Expiration time.Duration `mapstructure:"JWT_EXPIRATION"`
}

// TypesenseConfig armazena as configurações relacionadas ao Typesense
type TypesenseConfig struct {
	Host     string `mapstructure:"TYPESENSE_HOST"`
	Port     int    `mapstructure:"TYPESENSE_PORT"`
	Protocol string `mapstructure:"TYPESENSE_PROTOCOL"`
	APIKey   string `mapstructure:"TYPESENSE_API_KEY"`
}

// SwaggerConfig armazena as configurações relacionadas à documentação Swagger
type SwaggerConfig struct {
	Host string `mapstructure:"SWAGGER_HOST"`
}

// GetDSN retorna o DSN formatado para conexão com o PostgreSQL
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode, d.TimeZone)
}

// GetSwaggerHost retorna o host para o Swagger (com a porta)
func (c *Config) GetSwaggerHost() string {
	if c.Swagger.Host != "" {
		return c.Swagger.Host
	}
	return fmt.Sprintf("localhost:%s", c.Port)
}

var cfg *Config

// Load carrega as configurações do ambiente para a estrutura Config
func Load() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	v := viper.New()
	
	// Configuração para leitura do arquivo .env
	v.SetConfigFile(".env")
	v.AddConfigPath(".")
	v.SetConfigType("env")
	
	// Ler variáveis de ambiente
	v.AutomaticEnv()
	
	// Definir valores padrão
	setDefaults(v)
	
	// Tentar carregar do arquivo .env
	if err := v.ReadInConfig(); err != nil {
		// Apenas logar o erro, não é fatal se o arquivo não existir
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Arquivo .env não encontrado, usando variáveis de ambiente")
		} else {
			log.Printf("Erro ao carregar .env: %v\n", err)
		}
	}
	
	// Mapear as configurações para a estrutura
	cfg = &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("erro ao decodificar configurações: %w", err)
	}
	
	// Definir configurações adicionais não mapeadas diretamente
	cfg.Database = DatabaseConfig{
		Host:     v.GetString("DB_HOST"),
		Port:     v.GetString("DB_PORT"),
		User:     v.GetString("DB_USER"),
		Password: v.GetString("DB_PASSWORD"),
		Name:     v.GetString("DB_NAME"),
		SSLMode:  v.GetString("DB_SSL_MODE"),
		TimeZone: v.GetString("DB_TIMEZONE"),
	}
	
	jwtExpiration := v.GetString("JWT_EXPIRATION")
	duration, err := time.ParseDuration(jwtExpiration)
	if err != nil {
		duration = 24 * time.Hour // Padrão de 24 horas
	}
	
	cfg.JWT = JWTConfig{
		Secret:     v.GetString("JWT_SECRET"),
		Expiration: duration,
	}

	// Configurações do Typesense
	cfg.Typesense = TypesenseConfig{
		Host:     v.GetString("TYPESENSE_HOST"),
		Port:     v.GetInt("TYPESENSE_PORT"),
		Protocol: v.GetString("TYPESENSE_PROTOCOL"),
		APIKey:   v.GetString("TYPESENSE_API_KEY"),
	}

	// Configurações do Swagger
	cfg.Swagger = SwaggerConfig{
		Host: v.GetString("SWAGGER_HOST"),
	}

	return cfg, nil
}

// Get retorna a configuração atual
func Get() *Config {
	if cfg == nil {
		var err error
		cfg, err = Load()
		if err != nil {
			log.Fatalf("Erro fatal ao carregar configurações: %v", err)
		}
	}
	return cfg
}

// IsDevelopment verifica se o ambiente é de desenvolvimento
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.AppEnv) == "development"
}

// IsProduction verifica se o ambiente é de produção
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.AppEnv) == "production"
}

// IsTest verifica se o ambiente é de teste
func (c *Config) IsTest() bool {
	return strings.ToLower(c.AppEnv) == "test"
}

// setDefaults define os valores padrão para as configurações
func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("PORT", "8081")
	v.SetDefault("DEBUG", true)
	v.SetDefault("API_PREFIX", "/api/v1")
	v.SetDefault("LOG_LEVEL", "debug")
	
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "postgres")
	v.SetDefault("DB_NAME", "app_db")
	v.SetDefault("DB_SSL_MODE", "disable")
	v.SetDefault("DB_TIMEZONE", "UTC")
	
	v.SetDefault("JWT_SECRET", "default_jwt_secret")
	v.SetDefault("JWT_EXPIRATION", "24h")

	v.SetDefault("TYPESENSE_HOST", "localhost")
	v.SetDefault("TYPESENSE_PORT", 8108)
	v.SetDefault("TYPESENSE_PROTOCOL", "http")
	v.SetDefault("TYPESENSE_API_KEY", "")
	
	v.SetDefault("SWAGGER_HOST", "")
}

// PrepareEnvForTests configura o ambiente para testes
func PrepareEnvForTests() {
	os.Setenv("APP_ENV", "test")
	os.Setenv("DB_NAME", "app_db_test")
	
	// Recarregar configurações
	cfg = nil
	Get()
}