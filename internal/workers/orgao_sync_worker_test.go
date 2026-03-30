package workers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate test tables (skip OportunidadeMEI as it uses PostgreSQL arrays)
	err = db.AutoMigrate(&models.Curso{}, &models.Emprego{})
	require.NoError(t, err)

	// Manually create a simple version of oportunidades_mei for testing
	err = db.Exec(`
		CREATE TABLE oportunidades_mei (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			titulo VARCHAR(255),
			orgao_id VARCHAR(100),
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	return db
}

func TestNewOrgaoSyncWorker(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(
		db,
		nil, // rmiClient
		nil, // redisClient
		nil, // orgaoSnapshotRepo
		nil, // cursoRepo
		nil, // empregoRepo
		nil, // oportunidadeMEIRepo
		cfg,
	)

	assert.NotNil(t, worker)
	assert.Equal(t, 10*time.Minute, worker.syncInterval)
	assert.Equal(t, 24*time.Hour, worker.staleThreshold)
	assert.Equal(t, 1*time.Hour, worker.failedRetryThreshold)
	assert.Equal(t, 50, worker.batchSize)
	assert.Equal(t, 3, worker.maxRetries)
}

func TestOrgaoSyncWorker_discoverOrgaoIDs_Empty(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	ctx := context.Background()
	orgaoIDs, err := worker.discoverOrgaoIDs(ctx)

	assert.NoError(t, err)
	assert.Empty(t, orgaoIDs)
}

func TestOrgaoSyncWorker_discoverOrgaoIDs_FromCursos(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	// Insert test data
	require.NoError(t, db.Create(&models.Curso{OrgaoID: "orgao-1", Titulo: "Curso A"}).Error)
	require.NoError(t, db.Create(&models.Curso{OrgaoID: "orgao-2", Titulo: "Curso B"}).Error)
	require.NoError(t, db.Create(&models.Curso{OrgaoID: "orgao-1", Titulo: "Curso C"}).Error) // Duplicate

	ctx := context.Background()
	orgaoIDs, err := worker.discoverOrgaoIDs(ctx)

	assert.NoError(t, err)
	assert.Len(t, orgaoIDs, 2) // Should deduplicate
	assert.Contains(t, orgaoIDs, "orgao-1")
	assert.Contains(t, orgaoIDs, "orgao-2")
}

func TestOrgaoSyncWorker_discoverOrgaoIDs_FromEmpregos(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	// Insert test data
	db.Create(&models.Emprego{OrgaoID: "orgao-emp-1", Titulo: "Job 1"})
	db.Create(&models.Emprego{OrgaoID: "orgao-emp-2", Titulo: "Job 2"})

	ctx := context.Background()
	orgaoIDs, err := worker.discoverOrgaoIDs(ctx)

	assert.NoError(t, err)
	assert.Len(t, orgaoIDs, 2)
	assert.Contains(t, orgaoIDs, "orgao-emp-1")
	assert.Contains(t, orgaoIDs, "orgao-emp-2")
}

func TestOrgaoSyncWorker_discoverOrgaoIDs_FromOportunidadesMEI(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	// Insert test data using raw SQL
	db.Exec("INSERT INTO oportunidades_mei (titulo, orgao_id) VALUES (?, ?)", "MEI 1", "orgao-mei-1")
	db.Exec("INSERT INTO oportunidades_mei (titulo, orgao_id) VALUES (?, ?)", "MEI 2", "orgao-mei-2")

	ctx := context.Background()
	orgaoIDs, err := worker.discoverOrgaoIDs(ctx)

	assert.NoError(t, err)
	assert.Len(t, orgaoIDs, 2)
	assert.Contains(t, orgaoIDs, "orgao-mei-1")
	assert.Contains(t, orgaoIDs, "orgao-mei-2")
}

func TestOrgaoSyncWorker_discoverOrgaoIDs_AcrossAllTables(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	// Insert data across all tables
	db.Create(&models.Curso{OrgaoID: "orgao-1", Titulo: "Curso"})
	db.Create(&models.Emprego{OrgaoID: "orgao-2", Titulo: "Emprego"})
	db.Exec("INSERT INTO oportunidades_mei (titulo, orgao_id) VALUES (?, ?)", "MEI", "orgao-3")
	db.Create(&models.Curso{OrgaoID: "orgao-1", Titulo: "Curso 2"}) // Duplicate across tables

	ctx := context.Background()
	orgaoIDs, err := worker.discoverOrgaoIDs(ctx)

	assert.NoError(t, err)
	assert.Len(t, orgaoIDs, 3) // Should deduplicate across tables
	assert.Contains(t, orgaoIDs, "orgao-1")
	assert.Contains(t, orgaoIDs, "orgao-2")
	assert.Contains(t, orgaoIDs, "orgao-3")
}

func TestOrgaoSyncWorker_discoverOrgaoIDs_IgnoresEmpty(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	// Insert data with empty orgao_id
	db.Create(&models.Curso{OrgaoID: "", Titulo: "No Orgao"})
	db.Create(&models.Curso{OrgaoID: "orgao-valid", Titulo: "Valid Orgao"})

	ctx := context.Background()
	orgaoIDs, err := worker.discoverOrgaoIDs(ctx)

	assert.NoError(t, err)
	assert.Len(t, orgaoIDs, 1) // Should skip empty IDs
	assert.Contains(t, orgaoIDs, "orgao-valid")
}

func TestOrgaoSyncWorker_tryAcquireLock_NoRedis(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	ctx := context.Background()
	token, acquired, err := worker.tryAcquireLock(ctx)

	assert.NoError(t, err)
	assert.True(t, acquired) // Should proceed without lock when Redis is nil
	assert.Empty(t, token)
}

func TestOrgaoSyncWorker_releaseLock_NoRedis(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	ctx := context.Background()
	// Should not panic
	worker.releaseLock(ctx, "test-token")
	worker.releaseLock(ctx, "")
}

func TestOrgaoSyncWorker_releaseLock_EmptyToken(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	ctx := context.Background()
	// Should handle empty token gracefully
	worker.releaseLock(ctx, "")
}

func TestStringPtr(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "non-empty string",
			input: "test",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "string with spaces",
			input: "  test  ",
		},
		{
			name:  "unicode characters",
			input: "Teste açẽntõs",
		},
		{
			name:  "special characters",
			input: "!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringPtr(tt.input)
			assert.NotNil(t, got, "stringPtr should never return nil")
			assert.Equal(t, tt.input, *got, "stringPtr should preserve the input value")
		})
	}
}

func TestStringPtrAddressUniqueness(t *testing.T) {
	str1 := stringPtr("test")
	str2 := stringPtr("test")

	assert.Equal(t, *str1, *str2, "Values should be equal")
	assert.NotSame(t, str1, str2, "Pointers should be different")
}

func TestStringPtrMultipleCalls(t *testing.T) {
	// Call stringPtr multiple times with same value
	results := make([]*string, 10)
	for i := 0; i < 10; i++ {
		results[i] = stringPtr("value")
	}

	// All values should be equal
	for i := 1; i < 10; i++ {
		assert.Equal(t, *results[0], *results[i])
	}

	// But all pointers should be different
	for i := 0; i < 10; i++ {
		for j := i + 1; j < 10; j++ {
			assert.NotSame(t, results[i], results[j])
		}
	}
}

func TestStringPtrCanModifyValue(t *testing.T) {
	ptr := stringPtr("original")
	assert.Equal(t, "original", *ptr)

	*ptr = "modified"
	assert.Equal(t, "modified", *ptr)
}

func TestOrgaoSyncWorker_ConfigurationValues(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name string
		cfg  *config.OrgaoSyncSettings
	}{
		{
			name: "default configuration",
			cfg: &config.OrgaoSyncSettings{
				SyncInterval:         10 * time.Minute,
				StaleThreshold:       24 * time.Hour,
				FailedRetryThreshold: 1 * time.Hour,
				BatchSize:            50,
				MaxRetries:           3,
			},
		},
		{
			name: "custom configuration",
			cfg: &config.OrgaoSyncSettings{
				SyncInterval:         5 * time.Minute,
				StaleThreshold:       12 * time.Hour,
				FailedRetryThreshold: 30 * time.Minute,
				BatchSize:            100,
				MaxRetries:           5,
			},
		},
		{
			name: "high frequency configuration",
			cfg: &config.OrgaoSyncSettings{
				SyncInterval:         1 * time.Minute,
				StaleThreshold:       1 * time.Hour,
				FailedRetryThreshold: 5 * time.Minute,
				BatchSize:            10,
				MaxRetries:           1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, tt.cfg)

			assert.Equal(t, tt.cfg.SyncInterval, worker.syncInterval)
			assert.Equal(t, tt.cfg.StaleThreshold, worker.staleThreshold)
			assert.Equal(t, tt.cfg.FailedRetryThreshold, worker.failedRetryThreshold)
			assert.Equal(t, tt.cfg.BatchSize, worker.batchSize)
			assert.Equal(t, tt.cfg.MaxRetries, worker.maxRetries)
		})
	}
}

func TestOrgaoSyncLockKey(t *testing.T) {
	// Verify the lock key constant
	assert.Equal(t, "orgao_sync:lock", orgaoSyncLockKey)
}

func TestOrgaoSyncWorker_discoverOrgaoIDs_LargeDataset(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	// Create many records
	for i := 1; i <= 100; i++ {
		orgaoID := "orgao-" + string(rune(i%10+'0')) // Create 10 unique orgaos
		db.Create(&models.Curso{OrgaoID: orgaoID, Titulo: "Curso"})
	}

	ctx := context.Background()
	orgaoIDs, err := worker.discoverOrgaoIDs(ctx)

	assert.NoError(t, err)
	assert.LessOrEqual(t, len(orgaoIDs), 10) // Should deduplicate
}

func TestOrgaoSyncWorker_discoverOrgaoIDs_EmptyOrgaoID(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	// Create records with valid and empty orgao IDs
	db.Create(&models.Curso{Titulo: "Course without orgao"}) // OrgaoID will be empty
	db.Create(&models.Curso{OrgaoID: "valid-orgao", Titulo: "Course with orgao"})

	ctx := context.Background()
	orgaoIDs, err := worker.discoverOrgaoIDs(ctx)

	assert.NoError(t, err)
	assert.Len(t, orgaoIDs, 1)
	assert.Contains(t, orgaoIDs, "valid-orgao")
	assert.NotContains(t, orgaoIDs, "")
}

func TestOrgaoSyncWorker_MultipleTableDeduplication(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.OrgaoSyncSettings{
		SyncInterval:         10 * time.Minute,
		StaleThreshold:       24 * time.Hour,
		FailedRetryThreshold: 1 * time.Hour,
		BatchSize:            50,
		MaxRetries:           3,
	}

	worker := NewOrgaoSyncWorker(db, nil, nil, nil, nil, nil, nil, cfg)

	// Same orgao ID across multiple tables
	sharedOrgaoID := "shared-orgao"

	db.Create(&models.Curso{OrgaoID: sharedOrgaoID, Titulo: "Curso 1"})
	db.Create(&models.Curso{OrgaoID: sharedOrgaoID, Titulo: "Curso 2"})
	db.Create(&models.Emprego{OrgaoID: sharedOrgaoID, Titulo: "Job 1"})
	db.Exec("INSERT INTO oportunidades_mei (titulo, orgao_id) VALUES (?, ?)", "MEI 1", sharedOrgaoID)

	// Unique orgao IDs
	db.Create(&models.Curso{OrgaoID: "curso-only", Titulo: "Unique Curso"})
	db.Create(&models.Emprego{OrgaoID: "emprego-only", Titulo: "Unique Job"})

	ctx := context.Background()
	orgaoIDs, err := worker.discoverOrgaoIDs(ctx)

	assert.NoError(t, err)
	assert.Len(t, orgaoIDs, 3)
	assert.Contains(t, orgaoIDs, sharedOrgaoID)
	assert.Contains(t, orgaoIDs, "curso-only")
	assert.Contains(t, orgaoIDs, "emprego-only")
}
