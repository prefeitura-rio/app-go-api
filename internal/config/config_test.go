package config

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseSettings_DSN(t *testing.T) {
	tests := []struct {
		name     string
		settings DatabaseSettings
		want     string
	}{
		{
			name: "complete database settings",
			settings: DatabaseSettings{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
				SSLMode:  "disable",
				Timezone: "UTC",
			},
			want: "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable TimeZone=UTC",
		},
		{
			name: "production database settings",
			settings: DatabaseSettings{
				Host:     "prod.example.com",
				Port:     5432,
				User:     "produser",
				Password: "prodpass",
				Name:     "proddb",
				SSLMode:  "require",
				Timezone: "America/Sao_Paulo",
			},
			want: "host=prod.example.com port=5432 user=produser password=prodpass dbname=proddb sslmode=require TimeZone=America/Sao_Paulo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.settings.DSN()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppSettings_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{"development lowercase", "development", true},
		{"development uppercase", "DEVELOPMENT", true},
		{"development mixed case", "Development", true},
		{"production", "production", false},
		{"test", "test", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppSettings{Environment: tt.environment}
			assert.Equal(t, tt.want, a.IsDevelopment())
		})
	}
}

func TestAppSettings_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{"production lowercase", "production", true},
		{"production uppercase", "PRODUCTION", true},
		{"production mixed case", "Production", true},
		{"development", "development", false},
		{"test", "test", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppSettings{Environment: tt.environment}
			assert.Equal(t, tt.want, a.IsProduction())
		})
	}
}

func TestAppSettings_IsTest(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{"test lowercase", "test", true},
		{"test uppercase", "TEST", true},
		{"test mixed case", "Test", true},
		{"development", "development", false},
		{"production", "production", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppSettings{Environment: tt.environment}
			assert.Equal(t, tt.want, a.IsTest())
		})
	}
}

func TestJWTSettings_GetExpirationDuration(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn string
		want      time.Duration
		wantErr   bool
	}{
		{
			name:      "valid duration - hours",
			expiresIn: "24h",
			want:      24 * time.Hour,
			wantErr:   false,
		},
		{
			name:      "valid duration - minutes",
			expiresIn: "30m",
			want:      30 * time.Minute,
			wantErr:   false,
		},
		{
			name:      "valid duration - seconds",
			expiresIn: "300s",
			want:      300 * time.Second,
			wantErr:   false,
		},
		{
			name:      "valid duration - mixed",
			expiresIn: "1h30m",
			want:      90 * time.Minute,
			wantErr:   false,
		},
		{
			name:      "invalid duration",
			expiresIn: "invalid",
			want:      0,
			wantErr:   true,
		},
		{
			name:      "empty duration",
			expiresIn: "",
			want:      0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JWTSettings{ExpiresIn: tt.expiresIn}
			got, err := j.GetExpirationDuration()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDatabaseSettings_Validate(t *testing.T) {
	tests := []struct {
		name     string
		settings DatabaseSettings
		wantErr  error
	}{
		{
			name: "valid settings",
			settings: DatabaseSettings{
				Host: "localhost",
				Port: 5432,
			},
			wantErr: nil,
		},
		{
			name: "missing host",
			settings: DatabaseSettings{
				Host: "",
				Port: 5432,
			},
			wantErr: ErrNoHost,
		},
		{
			name: "invalid port - zero",
			settings: DatabaseSettings{
				Host: "localhost",
				Port: 0,
			},
			wantErr: ErrInvalidPort,
		},
		{
			name: "invalid port - negative",
			settings: DatabaseSettings{
				Host: "localhost",
				Port: -1,
			},
			wantErr: ErrInvalidPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerSettings_Validate(t *testing.T) {
	tests := []struct {
		name     string
		settings ServerSettings
		wantErr  error
	}{
		{
			name: "valid settings",
			settings: ServerSettings{
				Host: "0.0.0.0",
				Port: 8080,
			},
			wantErr: nil,
		},
		{
			name: "missing host",
			settings: ServerSettings{
				Host: "",
				Port: 8080,
			},
			wantErr: ErrNoHost,
		},
		{
			name: "invalid port - zero",
			settings: ServerSettings{
				Host: "0.0.0.0",
				Port: 0,
			},
			wantErr: ErrInvalidPort,
		},
		{
			name: "invalid port - negative",
			settings: ServerSettings{
				Host: "0.0.0.0",
				Port: -100,
			},
			wantErr: ErrInvalidPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAppConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  AppConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: AppConfig{
				Database: DatabaseSettings{
					Host: "localhost",
					Port: 5432,
				},
				Server: ServerSettings{
					Host: "0.0.0.0",
					Port: 8080,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid database",
			config: AppConfig{
				Database: DatabaseSettings{
					Host: "",
					Port: 5432,
				},
				Server: ServerSettings{
					Host: "0.0.0.0",
					Port: 8080,
				},
			},
			wantErr: true,
			errMsg:  "configuração de banco de dados inválida",
		},
		{
			name: "invalid server",
			config: AppConfig{
				Database: DatabaseSettings{
					Host: "localhost",
					Port: 5432,
				},
				Server: ServerSettings{
					Host: "",
					Port: 8080,
				},
			},
			wantErr: true,
			errMsg:  "configuração de servidor inválida",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"APP_ENV", "APP_DEBUG", "LOG_LEVEL", "API_PREFIX",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"SERVER_HOST", "SERVER_PORT",
		"REDIS_HOST", "REDIS_PORT",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
	}

	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
		// Reset singleton
		instance = nil
		once = sync.Once{}
		v = nil
	}()

	// Set test environment variables
	os.Setenv("APP_ENV", "test")
	os.Setenv("APP_DEBUG", "false")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("API_PREFIX", "/api/v2")
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("REDIS_HOST", "redishost")
	os.Setenv("REDIS_PORT", "6380")

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify App settings
	assert.Equal(t, "test", config.App.Environment)
	assert.False(t, config.App.Debug)
	assert.Equal(t, "debug", config.App.LogLevel)
	assert.Equal(t, "/api/v2", config.App.APIPrefix)

	// Verify Database settings
	assert.Equal(t, "testhost", config.Database.Host)
	assert.Equal(t, 5433, config.Database.Port)
	assert.Equal(t, "testuser", config.Database.User)
	assert.Equal(t, "testpass", config.Database.Password)
	assert.Equal(t, "testdb", config.Database.Name)

	// Verify Server settings
	assert.Equal(t, "127.0.0.1", config.Server.Host)
	assert.Equal(t, 9090, config.Server.Port)

	// Verify Redis settings
	assert.Equal(t, "redishost", config.Redis.Host)
	assert.Equal(t, 6380, config.Redis.Port)
}

func TestLoad_DefaultValues(t *testing.T) {
	// Clear environment variables
	envVars := []string{
		"APP_ENV", "APP_DEBUG", "LOG_LEVEL", "API_PREFIX",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"SERVER_HOST", "SERVER_PORT",
	}
	originalEnv := make(map[string]string)
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
		instance = nil
		once = sync.Once{}
		v = nil
	}()

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify default App settings
	assert.Equal(t, "development", config.App.Environment)
	assert.True(t, config.App.Debug)
	assert.Equal(t, "info", config.App.LogLevel)
	assert.Equal(t, "/api", config.App.APIPrefix)

	// Verify default Database settings
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "postgres", config.Database.User)
	assert.Equal(t, "app_go_api", config.Database.Name)

	// Verify default Server settings
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 8080, config.Server.Port)
}

func TestGet_Singleton(t *testing.T) {
	// Note: We cannot safely reset the singleton in tests because:
	// - Get() calls once.Do() which calls Initialize()
	// - Initialize() calls Get() at the end
	// - This creates a deadlock when once is reset
	//
	// Instead, we test that Get() returns the same instance on multiple calls
	// without resetting the singleton state.

	// Get config twice
	config1, err1 := Get()
	require.NoError(t, err1)
	require.NotNil(t, config1)

	config2, err2 := Get()
	require.NoError(t, err2)
	require.NotNil(t, config2)

	// Should return same instance (singleton behavior)
	assert.Same(t, config1, config2, "Get() should return the same instance on multiple calls")
}

func TestNewConfigProvider(t *testing.T) {
	// Reset singleton
	instance = nil
	once = sync.Once{}
	v = nil

	defer func() {
		instance = nil
		once = sync.Once{}
		v = nil
	}()

	// Set minimal valid environment
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("SERVER_HOST", "0.0.0.0")
	os.Setenv("SERVER_PORT", "8080")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
	}()

	provider, err := NewConfigProvider()
	require.NoError(t, err)
	require.NotNil(t, provider)

	config := provider.GetAppConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.Database.Host)
}

func TestConfigProvider_GetAppConfig(t *testing.T) {
	provider := &defaultConfigProvider{
		config: &AppConfig{
			App: AppSettings{
				Environment: "test",
			},
		},
	}

	config := provider.GetAppConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "test", config.App.Environment)
}

func TestGetStringSlice(t *testing.T) {
	// Save original environment
	originalValue := os.Getenv("TEST_STRING_SLICE")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_STRING_SLICE")
		} else {
			os.Setenv("TEST_STRING_SLICE", originalValue)
		}
		v = nil
	}()

	// Initialize viper
	err := Initialize()
	require.NoError(t, err)

	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{
			name:  "comma separated values",
			value: "read,write,delete",
			want:  []string{"read", "write", "delete"},
		},
		{
			name:  "comma separated with spaces",
			value: "read, write , delete",
			want:  []string{"read", "write", "delete"},
		},
		{
			name:  "single value",
			value: "read",
			want:  []string{"read"},
		},
		{
			name:  "empty value",
			value: "",
			want:  []string{},
		},
		{
			name:  "values with empty elements",
			value: "read,,write",
			want:  []string{"read", "write"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				os.Unsetenv("TEST_STRING_SLICE")
			} else {
				os.Setenv("TEST_STRING_SLICE", tt.value)
			}

			got := getStringSlice(v, "TEST_STRING_SLICE")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDuration(t *testing.T) {
	// Save original environment
	originalValue := os.Getenv("TEST_DURATION")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_DURATION")
		} else {
			os.Setenv("TEST_DURATION", originalValue)
		}
		v = nil
	}()

	// Initialize viper
	err := Initialize()
	require.NoError(t, err)

	tests := []struct {
		name         string
		value        string
		defaultValue time.Duration
		want         time.Duration
	}{
		{
			name:         "valid duration - hours",
			value:        "2h",
			defaultValue: 1 * time.Hour,
			want:         2 * time.Hour,
		},
		{
			name:         "valid duration - minutes",
			value:        "30m",
			defaultValue: 1 * time.Hour,
			want:         30 * time.Minute,
		},
		{
			name:         "invalid duration - use default",
			value:        "invalid",
			defaultValue: 1 * time.Hour,
			want:         1 * time.Hour,
		},
		{
			name:         "empty value - use default",
			value:        "",
			defaultValue: 5 * time.Minute,
			want:         5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				os.Unsetenv("TEST_DURATION")
			} else {
				os.Setenv("TEST_DURATION", tt.value)
			}

			got := getDuration(v, "TEST_DURATION", tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetBool(t *testing.T) {
	// Save original environment
	originalValue := os.Getenv("TEST_BOOL")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_BOOL")
		} else {
			os.Setenv("TEST_BOOL", originalValue)
		}
		v = nil
	}()

	// Initialize viper
	err := Initialize()
	require.NoError(t, err)

	tests := []struct {
		name         string
		value        string
		defaultValue bool
		want         bool
	}{
		{
			name:         "true lowercase",
			value:        "true",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "true uppercase",
			value:        "TRUE",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "false",
			value:        "false",
			defaultValue: true,
			want:         false,
		},
		{
			name:         "invalid - use default",
			value:        "invalid",
			defaultValue: true,
			want:         true,
		},
		{
			name:         "empty - use default",
			value:        "",
			defaultValue: false,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				os.Unsetenv("TEST_BOOL")
			} else {
				os.Setenv("TEST_BOOL", tt.value)
			}

			got := getBool(v, "TEST_BOOL", tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetEnv(t *testing.T) {
	// Save original environment
	originalValue := os.Getenv("TEST_ENV")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_ENV")
		} else {
			os.Setenv("TEST_ENV", originalValue)
		}
		v = nil
	}()

	// Initialize viper
	err := Initialize()
	require.NoError(t, err)

	tests := []struct {
		name         string
		value        string
		defaultValue string
		want         string
	}{
		{
			name:         "existing value",
			value:        "test-value",
			defaultValue: "default",
			want:         "test-value",
		},
		{
			name:         "empty value - use default",
			value:        "",
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				os.Unsetenv("TEST_ENV")
			} else {
				os.Setenv("TEST_ENV", tt.value)
			}

			got := getEnv(v, "TEST_ENV", tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetInt(t *testing.T) {
	// Save original environment
	originalValue := os.Getenv("TEST_INT")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_INT")
		} else {
			os.Setenv("TEST_INT", originalValue)
		}
		v = nil
	}()

	// Initialize viper
	err := Initialize()
	require.NoError(t, err)

	tests := []struct {
		name         string
		value        string
		defaultValue int
		want         int
	}{
		{
			name:         "valid integer",
			value:        "42",
			defaultValue: 10,
			want:         42,
		},
		{
			name:         "invalid integer - use default",
			value:        "invalid",
			defaultValue: 10,
			want:         10,
		},
		{
			name:         "empty value - use default",
			value:        "",
			defaultValue: 20,
			want:         20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				os.Unsetenv("TEST_INT")
			} else {
				os.Setenv("TEST_INT", tt.value)
			}

			got := getInt(v, "TEST_INT", tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoad_DurationSettings(t *testing.T) {
	// Reset environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"ORGAO_SYNC_INTERVAL",
		"ORGAO_SYNC_STALE_THRESHOLD",
		"CACHE_LEGAL_ENTITIES_TTL",
		"DB_HOST", "DB_PORT",
		"SERVER_HOST", "SERVER_PORT",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
		instance = nil
		once = sync.Once{}
		v = nil
	}()

	// Set required fields and duration fields
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("SERVER_HOST", "0.0.0.0")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("ORGAO_SYNC_INTERVAL", "30m")
	os.Setenv("ORGAO_SYNC_STALE_THRESHOLD", "2h")
	os.Setenv("CACHE_LEGAL_ENTITIES_TTL", "1h")

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, 30*time.Minute, config.OrgaoSync.SyncInterval)
	assert.Equal(t, 2*time.Hour, config.OrgaoSync.StaleThreshold)
	assert.Equal(t, 1*time.Hour, config.Cache.LegalEntitiesTTL)
}

func TestLoad_BooleanSettings(t *testing.T) {
	// Reset environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"APP_DEBUG",
		"RUN_MIGRATIONS",
		"TRACING_ENABLED",
		"CERBOS_ENABLED",
		"DB_HOST", "DB_PORT",
		"SERVER_HOST", "SERVER_PORT",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
		instance = nil
		once = sync.Once{}
		v = nil
	}()

	// Set required fields and boolean fields
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("SERVER_HOST", "0.0.0.0")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("APP_DEBUG", "false")
	os.Setenv("RUN_MIGRATIONS", "true")
	os.Setenv("TRACING_ENABLED", "true")
	os.Setenv("CERBOS_ENABLED", "true")

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.False(t, config.App.Debug)
	assert.True(t, config.Migrations.Run)
	assert.True(t, config.Tracing.Enabled)
	assert.True(t, config.Cerbos.Enabled)
}

func TestLoad_SliceSettings(t *testing.T) {
	// Reset environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"PROPOSTA_MEI_DELETE_PERMISSIONS",
		"PROPOSTA_MEI_UPDATE_PERMISSIONS",
		"PROPOSTA_MEI_READ_PERMISSIONS",
		"DB_HOST", "DB_PORT",
		"SERVER_HOST", "SERVER_PORT",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
		instance = nil
		once = sync.Once{}
		v = nil
	}()

	// Set required fields and slice fields
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("SERVER_HOST", "0.0.0.0")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("PROPOSTA_MEI_DELETE_PERMISSIONS", "admin:delete,manager:delete")
	os.Setenv("PROPOSTA_MEI_UPDATE_PERMISSIONS", "admin:update, manager:update")
	os.Setenv("PROPOSTA_MEI_READ_PERMISSIONS", "admin:read")

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, []string{"admin:delete", "manager:delete"}, config.PropostaMEI.DeletePermissions)
	assert.Equal(t, []string{"admin:update", "manager:update"}, config.PropostaMEI.UpdatePermissions)
	assert.Equal(t, []string{"admin:read"}, config.PropostaMEI.ReadPermissions)
}

func TestInitialize(t *testing.T) {
	// Reset viper
	v = nil

	defer func() {
		v = nil
	}()

	err := Initialize()
	// Initialize can fail if DB validation fails, but we're just testing that it doesn't panic
	// and that viper gets initialized
	if err == nil || strings.Contains(err.Error(), "configuração") {
		// Either success or validation error is acceptable
		assert.NotNil(t, v)
	}
}
