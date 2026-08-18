package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// contactMockInscricaoRepo is a minimal InscricaoRepositoryInterface implementation that
// captures the enrollment handed to Create so the test can assert what was persisted.
type contactMockInscricaoRepo struct {
	created *models.Inscricao
}

func (m *contactMockInscricaoRepo) Create(_ context.Context, inscricao *models.Inscricao) error {
	inscricao.ID = uuid.New()
	m.created = inscricao
	return nil
}
func (m *contactMockInscricaoRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Inscricao, error) {
	return nil, nil
}
func (m *contactMockInscricaoRepo) Update(_ context.Context, _ *models.Inscricao) error { return nil }
func (m *contactMockInscricaoRepo) Delete(_ context.Context, _ uuid.UUID) error         { return nil }
func (m *contactMockInscricaoRepo) ExistsByCPFAndCurso(_ context.Context, _ string, _ int) (bool, error) {
	return false, nil
}
func (m *contactMockInscricaoRepo) GetByCursoID(_ context.Context, _ int, _ map[string]interface{}, _, _ int) ([]*models.Inscricao, int, error) {
	return nil, 0, nil
}
func (m *contactMockInscricaoRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ models.StatusInscricao, _, _ string) error {
	return nil
}
func (m *contactMockInscricaoRepo) UpdateMultipleStatus(_ context.Context, _ []uuid.UUID, _ models.StatusInscricao, _, _ string) (int, error) {
	return 0, nil
}
func (m *contactMockInscricaoRepo) GetSummaryByCursoID(_ context.Context, _ int) (*models.EnrollmentSummary, error) {
	return nil, nil
}
func (m *contactMockInscricaoRepo) ListByCPF(_ context.Context, _ string, _ map[string]interface{}, _, _ int) ([]*models.Inscricao, int, error) {
	return nil, 0, nil
}
func (m *contactMockInscricaoRepo) UpdateCertificate(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

// contactMockCursoRepo is a minimal CursoRepositoryInterface implementation. Only
// ValidateForEnrollment is exercised (course closed, no auto-approve, no date window).
type contactMockCursoRepo struct{}

func (m *contactMockCursoRepo) Create(_ context.Context, _ *models.Curso) (int, error) { return 0, nil }
func (m *contactMockCursoRepo) GetByID(_ context.Context, _ int) (*models.Curso, error) {
	return nil, nil
}
func (m *contactMockCursoRepo) Update(_ context.Context, _ *models.Curso) error { return nil }
func (m *contactMockCursoRepo) UpdateStatus(_ context.Context, _ int, _ models.StatusCurso) error {
	return nil
}
func (m *contactMockCursoRepo) Delete(_ context.Context, _ int) error { return nil }
func (m *contactMockCursoRepo) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*models.Curso, int, error) {
	return nil, 0, nil
}
func (m *contactMockCursoRepo) CreateCustomFields(_ context.Context, _ []models.CustomField) error {
	return nil
}
func (m *contactMockCursoRepo) CreateRemoteClass(_ context.Context, _ *models.RemoteClass) error {
	return nil
}
func (m *contactMockCursoRepo) CreateLocationClasses(_ context.Context, _ []models.LocationClass) error {
	return nil
}
func (m *contactMockCursoRepo) ValidateForEnrollment(_ context.Context, _ int) (string, *time.Time, *time.Time, bool, error) {
	return string(models.StatusCursoClosed), nil, nil, false, nil
}
func (m *contactMockCursoRepo) CountEnrollmentsByScheduleID(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *contactMockCursoRepo) CountEnrollmentsByScheduleIDs(_ context.Context, _ []uuid.UUID) (map[uuid.UUID]int64, error) {
	return nil, nil
}
func (m *contactMockCursoRepo) GetCourseScheduleByID(_ context.Context, _ uuid.UUID) (*models.CourseSchedule, error) {
	return nil, nil
}
func (m *contactMockCursoRepo) GetRemoteScheduleByID(_ context.Context, _ uuid.UUID) (*models.RemoteSchedule, error) {
	return nil, nil
}

// contactMockCitizenFetcher returns a snapshot that DIVERGES from the sheet contact, so the
// test can prove the import does not let RMI data override what the secretaria provided.
type contactMockCitizenFetcher struct {
	snapshot *models.CitizenSnapshot
}

func (m *contactMockCitizenFetcher) SyncCitizenOnDemand(_ context.Context, _ string) (*models.CitizenSnapshot, error) {
	return m.snapshot, nil
}

// TestProcessRow_PreservesSecretariaContact covers the manual-import path end to end at the
// job level: a spreadsheet row carrying the órgão/secretaria contact must be persisted with
// that exact e-mail/phone, even when RMI (citizen snapshot) holds different values.
func TestProcessRow_PreservesSecretariaContact(t *testing.T) {
	ctx := context.Background()

	inscricaoRepo := &contactMockInscricaoRepo{}
	citizenFetcher := &contactMockCitizenFetcher{
		snapshot: &models.CitizenSnapshot{
			CPF:     "12345678900",
			Email:   "rmi@rmi.com",
			Celular: "21999999999",
			Nome:    "Nome do RMI",
		},
	}

	inscricaoService := services.NewInscricaoServiceWithInterface(
		inscricaoRepo,
		&contactMockCursoRepo{},
		nil,            // citizen snapshot repo unused here
		citizenFetcher, // divergent RMI data
		nil,            // email notifier unused
		&config.AppConfig{},
	)

	processor := NewEnrollmentImportProcessor(nil, nil, inscricaoService, nil)

	row := EnrollmentRow{
		NomeCompleto: "Maria da Secretaria",
		CPF:          "12345678900",
		Telefone:     "21988887777",
		Email:        "maria.secretaria@example.com",
	}

	id, err := processor.processRow(ctx, 1, row, []models.CustomField{}, nil, []models.LocationClass{}, false, nil)
	if err != nil {
		t.Fatalf("processRow failed: %v", err)
	}
	if id == nil {
		t.Fatal("expected an enrollment id, got nil")
	}

	created := inscricaoRepo.created
	if created == nil {
		t.Fatal("expected an enrollment to be persisted, got nil")
	}
	if created.Email != "maria.secretaria@example.com" {
		t.Errorf("expected secretaria email persisted, got %q", created.Email)
	}
	if created.Phone != "21988887777" {
		t.Errorf("expected secretaria phone persisted, got %q", created.Phone)
	}
	// The name informed on the sheet must not be replaced by the RMI name.
	if created.Name != "Maria da Secretaria" {
		t.Errorf("expected sheet name preserved, got %q", created.Name)
	}
}
