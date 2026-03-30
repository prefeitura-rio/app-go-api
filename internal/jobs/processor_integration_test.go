package jobs

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupIntegrationTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	// Migrate all needed tables
	db.AutoMigrate(
		&models.Job{},
		&models.Curso{},
		&models.Inscricao{},
		&models.CustomField{},
		&models.LocationClass{},
		&models.CourseSchedule{},
		&models.RemoteClass{},
		&models.RemoteSchedule{},
	)

	return db
}

func TestProcessJob_EnrollmentImport_Integration(t *testing.T) {
	db := setupIntegrationTestDB()

	// Create repositories
	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	// Create services
	jobService := services.NewJobService(jobRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)

	// Create processor
	processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

	// Create a course
	curso := &models.Curso{
		Titulo:    "Test Course",
		Descricao: "Test Description",
		OrgaoID:   "test-orgao",
	}
	db.Create(curso)

	// Create a test CSV file
	csvContent := `Nome,CPF
João Silva,12345678901
Maria Santos,98765432100`

	tmpFile, err := os.CreateTemp("", "test-*.csv")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(csvContent)
	assert.NoError(t, err)
	tmpFile.Close()

	// Create job metadata
	metadata := models.EnrollmentImportMetadata{
		CursoID:  curso.ID,
		FileName: tmpFile.Name(),
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Create job
	job := &models.Job{
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusPending,
		Metadata: datatypes.JSON(metadataJSON),
	}
	err = jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Process job
	err = processor.ProcessJob(job.ID)

	// Job should complete (file will be parsed but enrollments may fail due to validation)
	assert.Error(t, err) // Will fail on missing required fields, but that's expected

	// Verify job was tracked and removed from active jobs
	processor.mu.RLock()
	_, exists := processor.activeJobs[job.ID]
	processor.mu.RUnlock()
	assert.False(t, exists)
}

func TestBuildScheduleMap_Integration(t *testing.T) {
	now := time.Now()
	locationID := uuid.New()
	scheduleID := uuid.New()

	locations := []models.LocationClass{
		{
			ID:           locationID,
			Address:      "Rua Teste, 123",
			Neighborhood: "Centro",
			CursoID:      1,
			Schedules: []models.CourseSchedule{
				{
					ID:             scheduleID,
					LocationID:     locationID,
					ClassTime:      "09:00-12:00",
					ClassDays:      "Segunda a Sexta",
					ClassStartDate: now,
					ClassEndDate:   now.AddDate(0, 1, 0),
					Vacancies:      30,
				},
			},
		},
	}

	scheduleMap := buildScheduleMap(locations, false, nil)

	// Test various keys
	assert.Contains(t, scheduleMap, scheduleID.String())
	assert.Contains(t, scheduleMap, locationID.String())
	assert.Contains(t, scheduleMap, "rua teste, 123")

	// Verify the mapping is correct
	mapping := scheduleMap[scheduleID.String()]
	assert.Equal(t, locationID, mapping.LocationID)
	assert.Equal(t, scheduleID, mapping.ScheduleID)
}

func TestBuildScheduleMap_RemoteClass_Integration(t *testing.T) {
	now := time.Now()
	remoteClassID := uuid.New()
	scheduleID := uuid.New()
	classTime := "14:00-17:00"
	classDays := "Terça e Quinta"

	remoteClass := &models.RemoteClass{
		ID:      remoteClassID,
		CursoID: 1,
		Schedules: []models.RemoteSchedule{
			{
				ID:             scheduleID,
				RemoteClassID:  remoteClassID,
				ClassTime:      &classTime,
				ClassDays:      &classDays,
				ClassStartDate: &now,
				ClassEndDate:   &now,
				Vacancies:      50,
			},
		},
	}

	scheduleMap := buildScheduleMap([]models.LocationClass{}, true, remoteClass)

	// Test UUID keys
	assert.Contains(t, scheduleMap, scheduleID.String())
	assert.Contains(t, scheduleMap, remoteClassID.String())

	// Test time+days composite key
	timeDaysKey := "14:00-17:00|terça e quinta"
	assert.Contains(t, scheduleMap, timeDaysKey)

	// Verify mapping
	mapping := scheduleMap[scheduleID.String()]
	assert.Equal(t, remoteClassID, mapping.LocationID)
	assert.Equal(t, scheduleID, mapping.ScheduleID)
}

func TestFindScheduleByTurma_Scenarios(t *testing.T) {
	now := time.Now()
	locationID := uuid.New()
	scheduleID := uuid.New()

	locations := []models.LocationClass{
		{
			ID:           locationID,
			Address:      "Rua das Flores, 456",
			Neighborhood: "Centro",
			CursoID:      1,
			Schedules: []models.CourseSchedule{
				{
					ID:             scheduleID,
					LocationID:     locationID,
					ClassTime:      "10:00-13:00",
					ClassDays:      "Segunda",
					ClassStartDate: now,
					ClassEndDate:   now,
					Vacancies:      20,
				},
			},
		},
	}

	scheduleMap := buildScheduleMap(locations, false, nil)

	tests := []struct {
		name        string
		turma       string
		shouldFind  bool
		expectedLoc uuid.UUID
		expectedSch uuid.UUID
	}{
		{
			name:        "UUID match",
			turma:       scheduleID.String(),
			shouldFind:  true,
			expectedLoc: locationID,
			expectedSch: scheduleID,
		},
		{
			name:        "Exact address match",
			turma:       "Rua das Flores, 456",
			shouldFind:  true,
			expectedLoc: locationID,
			expectedSch: scheduleID,
		},
		{
			name:       "Non-existent",
			turma:      "Non-existent",
			shouldFind: false,
		},
		{
			name:        "Fuzzy address match",
			turma:       "Flores",
			shouldFind:  true,
			expectedLoc: locationID,
			expectedSch: scheduleID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locID, schID, err := findScheduleByTurma(tt.turma, scheduleMap, locations, false, nil)

			if tt.shouldFind {
				assert.NoError(t, err)
				assert.NotNil(t, locID)
				assert.NotNil(t, schID)
				assert.Equal(t, tt.expectedLoc, *locID)
				assert.Equal(t, tt.expectedSch, *schID)
			} else {
				assert.Error(t, err)
				assert.Nil(t, locID)
				assert.Nil(t, schID)
			}
		})
	}
}

func TestParseCSV_Integration(t *testing.T) {
	db := setupIntegrationTestDB()
	processor := NewEnrollmentImportProcessor(db, nil, nil, nil)

	tests := []struct {
		name        string
		csvContent  string
		wantErr     bool
		wantRows    int
		errContains string
	}{
		{
			name: "Valid CSV",
			csvContent: `Nome,CPF,Idade
João Silva,12345678901,30
Maria Santos,98765432100,25`,
			wantErr:  false,
			wantRows: 2,
		},
		{
			name: "Missing CPF column",
			csvContent: `Nome,Idade
João Silva,30`,
			wantErr:     true,
			errContains: "cpf",
		},
		{
			name: "Missing Nome column",
			csvContent: `CPF,Idade
12345678901,30`,
			wantErr:     true,
			errContains: "nome",
		},
		{
			name: "With custom fields",
			csvContent: `Nome,CPF,Campo Extra
João,12345678901,Valor1`,
			wantErr:  false,
			wantRows: 1,
		},
		{
			name: "With Turma field",
			csvContent: `Nome,CPF,Turma
João,12345678901,Turma A`,
			wantErr:  false,
			wantRows: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, _ := os.CreateTemp("", "test-*.csv")
			defer os.Remove(tmpFile.Name())
			tmpFile.WriteString(tt.csvContent)
			tmpFile.Close()

			rows, err := processor.parseCSV(tmpFile.Name(), map[string]string{})

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRows, len(rows))
			}
		})
	}
}

func TestCancelJob_Integration(t *testing.T) {
	processor := NewJobProcessor(nil, nil, nil, nil)

	jobID := uuid.New()

	// Job not running
	err := processor.CancelJob(jobID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "não está em execução")

	// Add job to active jobs
	cancelled := false
	processor.mu.Lock()
	processor.activeJobs[jobID] = func() {
		cancelled = true
	}
	processor.mu.Unlock()

	// Cancel should work
	err = processor.CancelJob(jobID)
	assert.NoError(t, err)
	assert.True(t, cancelled)
}

func TestStartJob_Integration(t *testing.T) {
	db := setupIntegrationTestDB()

	jobRepo := repository.NewJobRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	cursoRepo := repository.NewCursoRepository(db)

	jobService := services.NewJobService(jobRepo)
	inscricaoService := services.NewInscricaoService(inscricaoRepo, cursoRepo, nil, nil, nil, nil)
	cursoService := services.NewCursoService(cursoRepo)

	processor := NewJobProcessor(db, jobService, inscricaoService, cursoService)

	// Create a job
	job := &models.Job{
		Type:     models.JobTypeEnrollmentImport,
		Status:   models.JobStatusPending,
		Metadata: datatypes.JSON([]byte(`{"curso_id": 1, "file_name": "/nonexistent"}`)),
	}
	err := jobService.Create(context.Background(), job)
	assert.NoError(t, err)

	// Start job in background
	processor.StartJob(job.ID)

	// Give it time to process
	time.Sleep(100 * time.Millisecond)

	// Job should have been processed (even if it failed)
	processor.mu.RLock()
	_, exists := processor.activeJobs[job.ID]
	processor.mu.RUnlock()

	assert.False(t, exists, "Job should be removed from active jobs after completion")
}
