// Package config gerencia as configurações da aplicação
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// AppConfig contém todas as configurações da aplicação
type AppConfig struct {
	App         AppSettings
	Database    DatabaseSettings
	Server      ServerSettings
	JWT         JWTSettings
	Swagger     SwaggerSettings
	TypeSense   TypeSenseSettings
	Migrations  MigrationSettings
	RMI         RMISettings
	Redis       RedisSettings
	Tracing     TracingSettings
	OrgaoSync   OrgaoSyncSettings
	CitizenSync CitizenSyncSettings
	DataRelay   DataRelaySettings
	PrefRio     PrefRioSettings
	Cerbos      CerbosSettings
	PropostaMEI PropostaMEIPermissions
	Cache       CacheSettings
	Keycloak    KeycloakSettings
	Enrollment  EnrollmentSettings
	Heimdall    HeimdallSettings
}

// EnrollmentSettings define configurações para inscrições em cursos
type EnrollmentSettings struct {
	ScheduleChangeDeadlineHours int // Horas antes do início para permitir troca de turma (default: 72)
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
	// Connection pool settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in minutes
	ConnMaxIdleTime int // in minutes
}

// ServerSettings define configurações do servidor HTTP
type ServerSettings struct {
	Host           string
	Port           int
	RequestTimeout int // in seconds
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

// DataRelaySettings define configurações da API Data Relay (para envio de emails)
type DataRelaySettings struct {
	BaseURL string
	APIKey  string
}

// PrefRioSettings define configurações da plataforma Prefeitura Rio
type PrefRioSettings struct {
	Domain string // Domain for the Prefeitura Rio platform (e.g., "prefeitura.rio" or "1746.rio")
}

// RedisSettings define configurações do Redis
type RedisSettings struct {
	Host     string
	Port     int
	Password string
	DB       int
	// Connection pool settings
	PoolSize     int // Maximum number of socket connections
	MinIdleConns int // Minimum number of idle connections
}

// TracingSettings define configurações de observabilidade/tracing
type TracingSettings struct {
	Enabled        bool
	Endpoint       string
	ServiceName    string
	ServiceVersion string
}

// OrgaoSyncSettings define configurações do worker de sincronização de órgãos
type OrgaoSyncSettings struct {
	Enabled              bool
	SyncInterval         time.Duration // How often to run sync cycle
	StaleThreshold       time.Duration // Consider snapshot stale after this duration
	BatchSize            int           // Number of orgaos to sync per batch
	MaxRetries           int           // Maximum retries for failed syncs
	FailedRetryThreshold time.Duration // How long to wait before retrying a failed sync
}

// CitizenSyncSettings define configurações do worker de sincronização de dados de cidadão
type CitizenSyncSettings struct {
	Enabled        bool
	SyncInterval   time.Duration // How often to run sync cycle
	StaleThreshold time.Duration // Consider snapshot stale after this duration
	BatchSize      int           // Number of citizens to sync per batch
}

// CerbosSettings define configurações do Cerbos PDP para autorização
type CerbosSettings struct {
	Endpoint string
	Timeout  int
	Enabled  bool
}

// PropostaMEIPermissions define permissões configuráveis para propostas MEI
type PropostaMEIPermissions struct {
	DeletePermissions []string // Lista de ações permitidas para deletar propostas de outros
	UpdatePermissions []string // Lista de ações permitidas para atualizar propostas de outros
	ReadPermissions   []string // Lista de ações permitidas para visualizar propostas de outros
}

// CacheSettings define configurações de cache TTL
type CacheSettings struct {
	LegalEntitiesTTL time.Duration // TTL para cache de entidades legais (CNPJs)
	ReferenceDataTTL time.Duration // TTL para dados de referência (categorias, etc)
	CourseTTL        time.Duration // TTL para cache de cursos
	ContactInfoTTL   time.Duration // TTL para informações de contato de donos de CNPJ
}

// HeimdallSettings define configurações do Heimdall para verificação de roles
type HeimdallSettings struct {
	BaseURL string // URL base do Heimdall (ex: https://services.staging.app.dados.rio/heimdall-admin)
}

// KeycloakSettings define configurações do Keycloak para service account
type KeycloakSettings struct {
	URL          string // URL base do Keycloak (ex: https://auth.example.com/auth)
	Realm        string // Realm do Keycloak
	ClientID     string // Client ID do service account
	ClientSecret string // Client Secret do service account
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

// IsProduction verifica se o ambiente é de produção.
// Aceita tanto "prod" (valor real configurado no go-secrets de produção,
// via kubectl logs -n go <pod> -> "Ambiente: prod") quanto "production"
// (valor documentado/esperado), para não quebrar silenciosamente se o
// secret for normalizado no futuro.
func (a *AppSettings) IsProduction() bool {
	env := strings.ToLower(a.Environment)
	return env == "prod" || env == "production"
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

	// Note: We don't call Get() here to avoid circular dependency
	// Get() -> once.Do(Load) -> Load() -> Initialize() -> Get() would deadlock
	// The caller (Load or Get) will handle loading the config
	return nil
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
			Host:            getEnv(v, "DB_HOST", "localhost"),
			Port:            getInt(v, "DB_PORT", 5432),
			User:            getEnv(v, "DB_USER", "postgres"),
			Password:        getEnv(v, "DB_PASSWORD", "postgres"),
			Name:            getEnv(v, "DB_NAME", "app_go_api"),
			SSLMode:         getEnv(v, "DB_SSL_MODE", "disable"),
			Timezone:        getEnv(v, "DB_TIMEZONE", "UTC"),
			MaxOpenConns:    getInt(v, "DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt(v, "DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getInt(v, "DB_CONN_MAX_LIFETIME", 5),
			ConnMaxIdleTime: getInt(v, "DB_CONN_MAX_IDLE_TIME", 2),
		},
		Server: ServerSettings{
			Host:           getEnv(v, "SERVER_HOST", "0.0.0.0"),
			Port:           getInt(v, "SERVER_PORT", 8080),
			RequestTimeout: getInt(v, "SERVER_REQUEST_TIMEOUT", 30),
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
			Host:         getEnv(v, "REDIS_HOST", "localhost"),
			Port:         getInt(v, "REDIS_PORT", 6379),
			Password:     getEnv(v, "REDIS_PASSWORD", ""),
			DB:           getInt(v, "REDIS_DB", 0),
			PoolSize:     getInt(v, "REDIS_POOL_SIZE", 10),
			MinIdleConns: getInt(v, "REDIS_MIN_IDLE_CONNS", 2),
		},
		Tracing: TracingSettings{
			Enabled:        getBool(v, "TRACING_ENABLED", false),
			Endpoint:       getEnv(v, "TRACING_ENDPOINT", "localhost:4317"),
			ServiceName:    getEnv(v, "TRACING_SERVICE_NAME", "app-go-api"),
			ServiceVersion: getEnv(v, "TRACING_SERVICE_VERSION", "v1.0.0"),
		},
		OrgaoSync: OrgaoSyncSettings{
			Enabled:              getBool(v, "ORGAO_SYNC_ENABLED", true),
			SyncInterval:         getDuration(v, "ORGAO_SYNC_INTERVAL", 15*time.Minute),
			StaleThreshold:       getDuration(v, "ORGAO_SYNC_STALE_THRESHOLD", 7*24*time.Hour), // 1 week
			BatchSize:            getInt(v, "ORGAO_SYNC_BATCH_SIZE", 50),
			MaxRetries:           getInt(v, "ORGAO_SYNC_MAX_RETRIES", 3),
			FailedRetryThreshold: getDuration(v, "ORGAO_SYNC_FAILED_RETRY_THRESHOLD", 1*time.Hour),
		},
		CitizenSync: CitizenSyncSettings{
			Enabled:        getBool(v, "CITIZEN_SYNC_ENABLED", true),
			SyncInterval:   getDuration(v, "CITIZEN_SYNC_INTERVAL", 15*time.Minute),
			StaleThreshold: getDuration(v, "CITIZEN_SYNC_STALE_THRESHOLD", 15*time.Minute),
			BatchSize:      getInt(v, "CITIZEN_SYNC_BATCH_SIZE", 100),
		},
		DataRelay: DataRelaySettings{
			BaseURL: getEnv(v, "DATA_RELAY_BASE_URL", ""),
			APIKey:  getEnv(v, "DATA_RELAY_API_KEY", ""),
		},
		PrefRio: PrefRioSettings{
			Domain: getEnv(v, "PREFRIO_DOMAIN", "pref.rio"),
		},
		Cerbos: CerbosSettings{
			Endpoint: getEnv(v, "CERBOS_ENDPOINT", ""),
			Timeout:  getInt(v, "CERBOS_TIMEOUT_SECONDS", 2),
			Enabled:  getBool(v, "CERBOS_ENABLED", false),
		},
		PropostaMEI: PropostaMEIPermissions{
			DeletePermissions: getStringSlice(v, "PROPOSTA_MEI_DELETE_PERMISSIONS"),
			UpdatePermissions: getStringSlice(v, "PROPOSTA_MEI_UPDATE_PERMISSIONS"),
			ReadPermissions:   getStringSlice(v, "PROPOSTA_MEI_READ_PERMISSIONS"),
		},
		Cache: CacheSettings{
			LegalEntitiesTTL: getDuration(v, "CACHE_LEGAL_ENTITIES_TTL", 30*time.Minute),
			ReferenceDataTTL: getDuration(v, "CACHE_REFERENCE_DATA_TTL", 1*time.Hour),
			CourseTTL:        getDuration(v, "CACHE_COURSE_TTL", 5*time.Minute),
			ContactInfoTTL:   getDuration(v, "CACHE_CONTACT_INFO_TTL", 1*time.Hour),
		},
		Keycloak: KeycloakSettings{
			URL:          getEnv(v, "KEYCLOAK_URL", ""),
			Realm:        getEnv(v, "KEYCLOAK_REALM", ""),
			ClientID:     getEnv(v, "KEYCLOAK_CLIENT_ID", ""),
			ClientSecret: getEnv(v, "KEYCLOAK_CLIENT_SECRET", ""),
		},
		Enrollment: EnrollmentSettings{
			ScheduleChangeDeadlineHours: getInt(v, "ENROLLMENT_SCHEDULE_CHANGE_DEADLINE_HOURS", 48),
		},
		Heimdall: HeimdallSettings{
			BaseURL: getEnv(v, "HEIMDALL_BASE_URL", ""),
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
		strVal := v.GetString(key)
		if strVal != "" {
			// Use strconv.Atoi for proper parsing of string values
			if val, err := strconv.Atoi(strings.TrimSpace(strVal)); err == nil {
				return val
			}
			// Parse error - use default
			return defaultValue
		}
		// Empty string - try GetInt in case value is numeric type
		return v.GetInt(key)
	}
	// Fallback para os.Getenv diretamente
	if value := os.Getenv(key); value != "" {
		if val, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			return val
		}
	}
	return defaultValue
}

func getBool(v *viper.Viper, key string, defaultValue bool) bool {
	if v.IsSet(key) {
		strVal := strings.TrimSpace(v.GetString(key))
		lowerVal := strings.ToLower(strVal)
		// Only accept explicit true/false values
		if lowerVal == "true" || lowerVal == "1" {
			return true
		} else if lowerVal == "false" || lowerVal == "0" {
			return false
		}
		// Invalid boolean value - use default
		return defaultValue
	}
	// Fallback para os.Getenv diretamente
	if value := os.Getenv(key); value != "" {
		lowerVal := strings.ToLower(strings.TrimSpace(value))
		if lowerVal == "true" || lowerVal == "1" {
			return true
		} else if lowerVal == "false" || lowerVal == "0" {
			return false
		}
	}
	return defaultValue
}

func getDuration(v *viper.Viper, key string, defaultValue time.Duration) time.Duration {
	if v.IsSet(key) {
		val := v.Get(key)
		switch t := val.(type) {
		case time.Duration:
			return t
		case int:
			return time.Duration(t) * time.Second
		case int64:
			return time.Duration(t) * time.Second
		case float64:
			return time.Duration(t * float64(time.Second))
		case string:
			if d, err := time.ParseDuration(t); err == nil {
				return d
			}
		}
		// Parse error - use default
		return defaultValue
	}
	// Fallback para os.Getenv diretamente
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getStringSlice(v *viper.Viper, key string) []string {
	// Viper's GetStringSlice doesn't parse comma-separated values from env vars
	// We need to handle this manually
	if v.IsSet(key) {
		// First try to get as string and parse manually
		strVal := v.GetString(key)
		if strVal != "" {
			// Split by comma and trim spaces
			parts := strings.Split(strVal, ",")
			result := make([]string, 0, len(parts))
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					result = append(result, trimmed)
				}
			}
			return result
		}
		// Fallback to GetStringSlice for array values from config files
		return v.GetStringSlice(key)
	}
	// Fallback para os.Getenv com parse de comma-separated values
	if value := os.Getenv(key); value != "" {
		// Split by comma and trim spaces
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return []string{}
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
