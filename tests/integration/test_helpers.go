package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestDB provides database connection for integration tests
type TestDB struct {
	DB *gorm.DB
	T  *testing.T
}

// GetTestDB creates a test database connection
func GetTestDB(t *testing.T) *TestDB {
	t.Helper()
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" && os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping integration test: set RUN_INTEGRATION_TESTS=1 or DATABASE_URL to run")
		return nil
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		cfg, err := config.Load()
		if err != nil {
			t.Skipf("Skipping integration test: config load failed: %v", err)
			return nil
		}
		dsn = cfg.Database.DSN()
		if cfg.Database.Host == "" {
			t.Skip("Skipping integration test: DB_HOST not set")
			return nil
		}
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to database: %v", err)
		return nil
	}

	return &TestDB{DB: db, T: t}
}

// Cleanup cleans up test data
func (tdb *TestDB) Cleanup(ctx context.Context) {
	// Note: This is a simplified cleanup. In production, you might want more sophisticated cleanup
	// that only removes data created during tests
}

// TestServices holds all services for integration testing
type TestServices struct {
	CursoService       *services.CursoService
	InscricaoService   *services.InscricaoService
	PropostaMEIService *services.PropostaMEIService
	VagaService        *empServices.VagaService
	CandidaturaService *empServices.CandidaturaService
	CurriculoService   *empServices.CurriculoService
	EmailService       *services.EmailNotificationService
	CNAEService        *services.CNAEValidationService
	ContactInfoService *services.ContactInfoService
}

// SetupTestServices creates real service instances with real repositories
func SetupTestServices(t *testing.T, db *gorm.DB) *TestServices {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create repositories
	cursoRepo := repository.NewCursoRepository(db)
	inscricaoRepo := repository.NewInscricaoRepository(db)
	propostaMEIRepo := repository.NewPropostaMEIRepository(db)
	oportunidadeMEIRepo := repository.NewOportunidadeMEIRepository(db)
	vagaRepo := empRepository.NewVagaRepository(db)
	candidaturaRepo := empRepository.NewCandidaturaRepository(db)
	curriculoRepo := empRepository.NewCurriculoRepository(db)
	empresaRepo := empRepository.NewEmpresaRepository(db)
	citizenSnapshotRepo := repository.NewCitizenSnapshotRepository(db)

	// Create services (without external dependencies for testing)
	emailService := services.NewEmailNotificationService(nil, cfg) // Mock email client
	cnaeService := services.NewCNAEValidationService(nil)          // Mock RMI client
	contactInfoService := services.NewContactInfoService(citizenSnapshotRepo, nil)

	cursoService := services.NewCursoService(cursoRepo)
	inscricaoService := services.NewInscricaoServiceWithInterface(
		inscricaoRepo,
		cursoRepo,
		citizenSnapshotRepo,
		nil, // No citizen data fetcher for testing
		emailService,
		cfg,
	)

	propostaMEIService := services.NewPropostaMEIService(
		propostaMEIRepo,
		oportunidadeMEIRepo,
		cnaeService,
		contactInfoService,
	)

	curriculoService := empServices.NewCurriculoService(curriculoRepo)
	vagaService := empServices.NewVagaService(vagaRepo, empresaRepo, candidaturaRepo)
	candidaturaService := empServices.NewCandidaturaService(
		candidaturaRepo,
		vagaRepo,
		curriculoService,
		citizenSnapshotRepo,
		nil, // No citizen data fetcher for testing
	)

	return &TestServices{
		CursoService:       cursoService,
		InscricaoService:   inscricaoService,
		PropostaMEIService: propostaMEIService,
		VagaService:        vagaService,
		CandidaturaService: candidaturaService,
		CurriculoService:   curriculoService,
		EmailService:       emailService,
		CNAEService:        cnaeService,
		ContactInfoService: contactInfoService,
	}
}

// Test Data Builders

// CreateTestCurso creates a test course with minimal required fields
func CreateTestCurso(t *testing.T) *models.Curso {
	t.Helper()
	autoApprove := false
	return &models.Curso{
		Titulo:                 "Test Course - " + uuid.New().String()[:8],
		Organization:           "Test Org",
		Status:                 models.StatusCursoOpened,
		Modalidade:             models.ModalidadePresencial,
		NumeroVagas:            100,
		CargaHoraria:           40,
		AutoApproveEnrollments: &autoApprove,
		LocationClasses: []models.LocationClass{
			{
				Address:      "Test Address, 123",
				Neighborhood: "Test Neighborhood",
				Schedules: []models.CourseSchedule{
					{
						ID:             uuid.New(),
						Vacancies:      50,
						ClassTime:      "08:00-12:00",
						ClassDays:      "Segunda a Sexta",
						ClassStartDate: time.Now().Add(48 * time.Hour),
						ClassEndDate:   time.Now().Add(30 * 24 * time.Hour),
					},
				},
			},
		},
	}
}

// CreateTestInscricao creates a test enrollment
func CreateTestInscricao(cursoID int, cpf string) *models.Inscricao {
	return &models.Inscricao{
		CursoID: cursoID,
		CPF:     cpf,
		Name:    "Test User",
		Email:   "test@example.com",
		Phone:   "21999999999",
		Status:  models.StatusInscricaoPending,
	}
}

// CreateTestVaga creates a test job posting
func CreateTestVaga(t *testing.T, cnpjContratante string) *empregabilidade.Vaga {
	t.Helper()
	return &empregabilidade.Vaga{
		IDContratante:          cnpjContratante,
		Titulo:                 "Test Vaga - " + uuid.New().String()[:8],
		Descricao:              "Test description",
		IDRegimeContratacao:    uuid.New(), // Should be valid regime ID in real tests
		IDModeloTrabalho:       uuid.New(), // Should be valid modelo ID in real tests
		Remuneracao:            3000.00,
		Status:                 empregabilidade.StatusVagaEmEdicao,
		DataExpiracao:          time.Now().Add(30 * 24 * time.Hour),
		QuantidadeVagas:        10,
		RequisitosObrigatorios: "Test requirements",
	}
}

// CreateTestCandidatura creates a test job application
func CreateTestCandidatura(cpf string, vagaID uuid.UUID) *empregabilidade.Candidatura {
	return &empregabilidade.Candidatura{
		CPF:    cpf,
		IDVaga: vagaID,
		Status: empregabilidade.StatusCandidaturaEnviada,
	}
}

// CreateTestOportunidadeMEI creates a test MEI opportunity
func CreateTestOportunidadeMEI(t *testing.T) *models.OportunidadeMEI {
	t.Helper()
	return &models.OportunidadeMEI{
		NomeServico:      "Test Service - " + uuid.New().String()[:8],
		Descricao:        "Test description",
		Status:           models.StatusOportunidadeActive,
		ValorEstimado:    5000.00,
		CNAEIDs:          []string{"6201-5/00"}, // Example CNAE
		OrgaoSolicitante: "Test Orgao",
	}
}

// CreateTestPropostaMEI creates a test MEI proposal
func CreateTestPropostaMEI(oportunidadeID int, cnpj string) *models.PropostaMEI {
	valorProposta := 5000.00
	prazo := "30 dias"
	aceita := true
	return &models.PropostaMEI{
		OportunidadeMEIID:     oportunidadeID,
		MEIEmpresaID:          cnpj,
		ValorProposta:         &valorProposta,
		PrazoExecucao:         &prazo,
		AceitaCustosIntegrais: &aceita,
		StatusAdmin:           models.StatusPropostaAdminActive,
		StatusCidadao:         models.StatusPropostaCidadaoSubmitted,
	}
}

// Helper functions for assertions

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

// AssertNotNil fails the test if value is nil
func AssertNotNil(t *testing.T, value interface{}) {
	t.Helper()
	if value == nil {
		t.Fatal("Expected non-nil value")
	}
}

// GenerateUniqueCPF generates a unique CPF for testing
func GenerateUniqueCPF() string {
	return uuid.New().String()[:11]
}

// GenerateUniqueCNPJ generates a unique CNPJ for testing
func GenerateUniqueCNPJ() string {
	return uuid.New().String()[:14]
}
