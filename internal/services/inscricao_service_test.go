package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// MockInscricaoRepository implements InscricaoRepositoryInterface for testing
type MockInscricaoRepository struct {
	CreateFunc                func(ctx context.Context, inscricao *models.Inscricao) error
	GetByIDFunc               func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error)
	UpdateFunc                func(ctx context.Context, inscricao *models.Inscricao) error
	DeleteFunc                func(ctx context.Context, id uuid.UUID) error
	ExistsByCPFAndCursoFunc   func(ctx context.Context, cpf string, cursoID int) (bool, error)
	GetByCursoIDFunc          func(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error)
	UpdateStatusFunc          func(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error
	UpdateMultipleStatusFunc  func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error)
	GetSummaryByCursoIDFunc   func(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error)
	ListByCPFFunc             func(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error)
	UpdateCertificateFunc     func(ctx context.Context, inscricaoID uuid.UUID, certificateURL string) error
}

func (m *MockInscricaoRepository) Create(ctx context.Context, inscricao *models.Inscricao) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, inscricao)
	}
	inscricao.ID = uuid.New()
	return nil
}

func (m *MockInscricaoRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockInscricaoRepository) Update(ctx context.Context, inscricao *models.Inscricao) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, inscricao)
	}
	return nil
}

func (m *MockInscricaoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockInscricaoRepository) ExistsByCPFAndCurso(ctx context.Context, cpf string, cursoID int) (bool, error) {
	if m.ExistsByCPFAndCursoFunc != nil {
		return m.ExistsByCPFAndCursoFunc(ctx, cpf, cursoID)
	}
	return false, nil
}

func (m *MockInscricaoRepository) GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error) {
	if m.GetByCursoIDFunc != nil {
		return m.GetByCursoIDFunc(ctx, cursoID, filter, limit, offset)
	}
	return []*models.Inscricao{}, 0, nil
}

func (m *MockInscricaoRepository) UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, inscricaoID, status, reason, adminNotes)
	}
	return nil
}

func (m *MockInscricaoRepository) UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
	if m.UpdateMultipleStatusFunc != nil {
		return m.UpdateMultipleStatusFunc(ctx, inscricaoIDs, status, reason, adminNotes)
	}
	return len(inscricaoIDs), nil
}

func (m *MockInscricaoRepository) GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
	if m.GetSummaryByCursoIDFunc != nil {
		return m.GetSummaryByCursoIDFunc(ctx, cursoID)
	}
	return &models.EnrollmentSummary{}, nil
}

func (m *MockInscricaoRepository) ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
	if m.ListByCPFFunc != nil {
		return m.ListByCPFFunc(ctx, cpf, filter, offset, limit)
	}
	return []*models.Inscricao{}, 0, nil
}

func (m *MockInscricaoRepository) UpdateCertificate(ctx context.Context, inscricaoID uuid.UUID, certificateURL string) error {
	if m.UpdateCertificateFunc != nil {
		return m.UpdateCertificateFunc(ctx, inscricaoID, certificateURL)
	}
	return nil
}

// MockCitizenDataFetcher for testing
type MockCitizenDataFetcher struct {
	SyncFunc func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
}

func (m *MockCitizenDataFetcher) SyncCitizenOnDemand(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
	if m.SyncFunc != nil {
		return m.SyncFunc(ctx, cpf)
	}
	return nil, nil
}

// TestInscricaoService_Create tests the Create method
func TestInscricaoService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Create valid inscricao with auto-approve", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected status APPROVED, got %s", inscricao.Status)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, true, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Test User",
			Email:   "test@example.com",
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("Create valid inscricao without auto-approve", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoPending {
					t.Errorf("Expected status PENDING, got %s", inscricao.Status)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Test User",
			Email:   "test@example.com",
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("Create fails when curso not opened", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoDraft), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err == nil {
			t.Error("Expected error when curso not opened")
		}
	})

	t.Run("Create fails when enrollment period not started", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}

		futureDate := time.Now().Add(24 * time.Hour)
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), &futureDate, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err == nil {
			t.Error("Expected error when enrollment period not started")
		}
	})

	t.Run("Create fails when enrollment period ended", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}

		pastDate := time.Now().Add(-24 * time.Hour)
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, &pastDate, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err == nil {
			t.Error("Expected error when enrollment period ended")
		}
	})

	t.Run("Create fails when CPF already enrolled", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return true, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err == nil {
			t.Error("Expected error when CPF already enrolled")
		}
	})

	t.Run("Create with citizen data fetcher", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Email != "fetched@example.com" {
					t.Errorf("Expected email from fetcher, got %s", inscricao.Email)
				}
				if inscricao.Phone != "1234567890" {
					t.Errorf("Expected phone from fetcher, got %s", inscricao.Phone)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		citizenFetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return &models.CitizenSnapshot{
					CPF:     cpf,
					Email:   "fetched@example.com",
					Celular: "1234567890",
					Nome:    "Fetched Name",
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, citizenFetcher, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})
}

// TestInscricaoService_UpdateStatus tests the UpdateStatus method
func TestInscricaoService_UpdateStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateStatus from pending to approved", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				if status != models.StatusInscricaoApproved {
					t.Errorf("Expected status APPROVED, got %s", status)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
	})

	t.Run("UpdateStatus fails when inscricao not found", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err == nil {
			t.Error("Expected error when inscricao not found")
		}
	})

	t.Run("UpdateStatus with repository error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				return errors.New("database error")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err == nil {
			t.Error("Expected database error")
		}
	})

	t.Run("UpdateStatus with GetByID error", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, errors.New("database connection error")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err == nil {
			t.Error("Expected database error")
		}
		if !errors.Is(err, errors.New("database connection error")) && !strings.Contains(err.Error(), "erro ao verificar inscrição") {
			t.Errorf("Expected verification error, got: %v", err)
		}
	})

	t.Run("UpdateStatus status unchanged no email", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved, // Same status
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Update to same status - should not send email
		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
	})

	t.Run("UpdateStatus to rejected status", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
				if status != models.StatusInscricaoRejected {
					t.Errorf("Expected status REJECTED, got %s", status)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateStatus(ctx, inscricaoID, models.StatusInscricaoRejected, "Test reason", "Admin notes")
		if err != nil {
			t.Fatalf("UpdateStatus failed: %v", err)
		}
	})
}

// TestInscricaoService_UpdateMultipleStatus tests the UpdateMultipleStatus method
func TestInscricaoService_UpdateMultipleStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateMultipleStatus success", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count 3, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus with empty list fails", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}
		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.UpdateMultipleStatus(ctx, []uuid.UUID{}, models.StatusInscricaoApproved, "", "")
		if err == nil {
			t.Error("Expected error when list is empty")
		}
	})

	t.Run("UpdateMultipleStatus with repository error", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return 0, errors.New("bulk update failed")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err == nil {
			t.Error("Expected database error")
		}
		if count != 0 {
			t.Errorf("Expected count 0 on error, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus partial success", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				// Partial success: only 2 out of 3 updated
				return 2, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus with GetByID error during email collection", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}
		callCount := 0

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				callCount++
				if callCount == 1 {
					// First call succeeds
					return &models.Inscricao{
						ID:      id,
						CursoID: 1,
						Status:  models.StatusInscricaoPending,
					}, nil
				}
				// Second call fails
				return nil, errors.New("database error")
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Should still succeed despite GetByID error during email collection
		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus should succeed even with email collection errors: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus to rejected status", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				if status != models.StatusInscricaoRejected {
					t.Errorf("Expected status REJECTED, got %s", status)
				}
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoRejected, "Rejected reason", "Admin notes")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus status not approved or rejected", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Update to pending (not approved/rejected) - should not trigger email collection
		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoPending, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus to concluded status", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				if status != models.StatusInscricaoConcluded {
					t.Errorf("Expected status CONCLUDED, got %s", status)
				}
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoConcluded, "", "Curso concluído")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus without email service doesn't call GetByID", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		callCount := 0

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				callCount++
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		// No email notification service
		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
		// GetByID should not be called when email service is nil
		if callCount != 0 {
			t.Errorf("Expected GetByID not to be called without email service, called %d times", callCount)
		}
	})

	t.Run("UpdateMultipleStatus with zero updates", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				// Simulate zero rows updated (maybe due to concurrent update or constraint)
				return 0, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count 0, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus with all same status (no email sent)", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved, // Already approved
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return len(inscricaoIDs), nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{ID: 1, Titulo: "Test Course"}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Update to same status - should collect emails but skip sending
		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus with curso lookup failure during email", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New()}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
					Email:   "user@example.com",
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				return 1, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return nil, errors.New("curso not found")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Should still succeed even if curso lookup fails during email sending
		count, err := svc.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "", "")
		if err != nil {
			t.Fatalf("UpdateMultipleStatus should succeed even if curso lookup fails: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}
	})
}

// TestInscricaoService_Delete tests the Delete method
func TestInscricaoService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Delete existing inscricao", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{ID: inscricaoID}, nil
			},
			DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.Delete(ctx, inscricaoID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})

	t.Run("Delete fails when inscricao not found", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.Delete(ctx, inscricaoID)
		if err == nil {
			t.Error("Expected error when inscricao not found")
		}
	})
}

// TestInscricaoService_UpdateCertificate tests the UpdateCertificate method
func TestInscricaoService_UpdateCertificate(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateCertificate for approved inscricao", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateCertificateFunc: func(ctx context.Context, id uuid.UUID, certificateURL string) error {
				if certificateURL != "https://example.com/cert.pdf" {
					t.Errorf("Expected certificate URL, got %s", certificateURL)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err != nil {
			t.Fatalf("UpdateCertificate failed: %v", err)
		}
	})

	t.Run("UpdateCertificate fails when inscricao from different curso", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 2, // Different curso ID
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err == nil {
			t.Error("Expected error when inscricao from different curso")
		}
	})

	t.Run("UpdateCertificate fails for pending inscricao", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err == nil {
			t.Error("Expected error for pending inscricao")
		}
	})
}

// TestInscricaoService_GetSummaryByCursoID tests the GetSummaryByCursoID method
func TestInscricaoService_GetSummaryByCursoID(t *testing.T) {
	ctx := context.Background()

	t.Run("GetSummaryByCursoID success", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			GetSummaryByCursoIDFunc: func(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
				return &models.EnrollmentSummary{
					Total:    100,
					Pending:  10,
					Approved: 80,
					Rejected: 10,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		summary, err := svc.GetSummaryByCursoID(ctx, 1)
		if err != nil {
			t.Fatalf("GetSummaryByCursoID failed: %v", err)
		}
		if summary.Total != 100 {
			t.Errorf("Expected total 100, got %d", summary.Total)
		}
	})

	t.Run("GetSummaryByCursoID fails when curso not found", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return "", nil, nil, false, errors.New("curso not found")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.GetSummaryByCursoID(ctx, 999)
		if err == nil {
			t.Error("Expected error when curso not found")
		}
	})
}

// TestInscricaoService_GetByID tests the GetByID method
func TestInscricaoService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("GetByID existing inscricao", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					CPF:     "12345678900",
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao, err := svc.GetByID(ctx, inscricaoID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if inscricao == nil {
			t.Fatal("Expected inscricao, got nil")
		}
		if inscricao.ID != inscricaoID {
			t.Errorf("Expected ID %s, got %s", inscricaoID, inscricao.ID)
		}
	})

	t.Run("GetByID non-existing inscricao", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao, err := svc.GetByID(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if inscricao != nil {
			t.Error("Expected nil for non-existing inscricao")
		}
	})
}

// TestInscricaoService_GetByCursoID tests the GetByCursoID method
func TestInscricaoService_GetByCursoID(t *testing.T) {
	ctx := context.Background()

	t.Run("GetByCursoID with results", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			GetByCursoIDFunc: func(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error) {
				return []*models.Inscricao{
					{ID: uuid.New(), CursoID: 1},
					{ID: uuid.New(), CursoID: 1},
				}, 2, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricoes, total, err := svc.GetByCursoID(ctx, 1, nil, 1, 10)
		if err != nil {
			t.Fatalf("GetByCursoID failed: %v", err)
		}
		if total != 2 {
			t.Errorf("Expected total 2, got %d", total)
		}
		if len(inscricoes) != 2 {
			t.Errorf("Expected 2 inscricoes, got %d", len(inscricoes))
		}
	})

	t.Run("GetByCursoID with pagination", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			GetByCursoIDFunc: func(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error) {
				if limit != 5 || offset != 10 {
					t.Errorf("Expected limit=5 offset=10, got limit=%d offset=%d", limit, offset)
				}
				return []*models.Inscricao{}, 0, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, _, err := svc.GetByCursoID(ctx, 1, nil, 3, 5)
		if err != nil {
			t.Fatalf("GetByCursoID failed: %v", err)
		}
	})
}

// TestInscricaoService_ListByCPF tests the ListByCPF method
func TestInscricaoService_ListByCPF(t *testing.T) {
	ctx := context.Background()

	t.Run("ListByCPF with results", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ListByCPFFunc: func(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
				if cpf != "12345678900" {
					t.Errorf("Expected CPF 12345678900, got %s", cpf)
				}
				return []*models.Inscricao{
					{ID: uuid.New(), CPF: "12345678900"},
				}, 1, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricoes, total, err := svc.ListByCPF(ctx, "12345678900", nil, 0, 10)
		if err != nil {
			t.Fatalf("ListByCPF failed: %v", err)
		}
		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if len(inscricoes) != 1 {
			t.Errorf("Expected 1 inscricao, got %d", len(inscricoes))
		}
	})

	t.Run("ListByCPF with repository error", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ListByCPFFunc: func(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
				return nil, 0, errors.New("database error")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, _, err := svc.ListByCPF(ctx, "12345678900", nil, 0, 10)
		if err == nil {
			t.Error("Expected database error")
		}
	})
}

// TestInscricaoService_UpdateInscricao tests the UpdateInscricao method
func TestInscricaoService_UpdateInscricao(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateInscricao success", func(t *testing.T) {
		inscricaoID := uuid.New()
		newName := "Updated Name"
		newEmail := "updated@example.com"

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Name:    "Original Name",
					Email:   "original@example.com",
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Name != newName {
					t.Errorf("Expected name %s, got %s", newName, inscricao.Name)
				}
				if inscricao.Email != newEmail {
					t.Errorf("Expected email %s, got %s", newEmail, inscricao.Email)
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		updateData := &models.InscricaoUpdateRequest{
			Name:  &newName,
			Email: &newEmail,
		}

		err := svc.UpdateInscricao(ctx, inscricaoID, 1, updateData)
		if err != nil {
			t.Fatalf("UpdateInscricao failed: %v", err)
		}
	})

	t.Run("UpdateInscricao fails when inscricao not found", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return nil, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		updateData := &models.InscricaoUpdateRequest{}

		err := svc.UpdateInscricao(ctx, inscricaoID, 1, updateData)
		if err == nil {
			t.Error("Expected error when inscricao not found")
		}
	})

	t.Run("UpdateInscricao fails when curso ID mismatch", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 2, // Different curso
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		updateData := &models.InscricaoUpdateRequest{}

		err := svc.UpdateInscricao(ctx, inscricaoID, 1, updateData)
		if err == nil {
			t.Error("Expected error when curso ID mismatch")
		}
	})
}

// TestInscricaoService_CreateManual tests the CreateManual method
func TestInscricaoService_CreateManual(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateManual bypasses enrollment period", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected status APPROVED, got %s", inscricao.Status)
				}
				if inscricao.Email != "admin@example.com" {
					t.Errorf("Expected admin-provided email, got %s", inscricao.Email)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		// Simulate expired enrollment period
		pastDate := time.Now().Add(-24 * time.Hour)
		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, &pastDate, true, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Admin Created User",
			Email:   "admin@example.com",
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
	})

	t.Run("CreateManual does not fetch citizen data", func(t *testing.T) {
		citizenDataFetcherCalled := false

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Should preserve admin-provided data
				if inscricao.Email != "admin@example.com" {
					t.Errorf("Expected admin email to be preserved, got %s", inscricao.Email)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		citizenFetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				citizenDataFetcherCalled = true
				return &models.CitizenSnapshot{
					Email: "should-not-override@example.com",
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, citizenFetcher, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Manual User",
			Email:   "admin@example.com",
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}

		if citizenDataFetcherCalled {
			t.Error("Citizen data fetcher should not be called in CreateManual")
		}
	})

	t.Run("CreateManual fails when curso not opened", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoDraft), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.CreateManual(ctx, inscricao)
		if err == nil {
			t.Error("Expected error when curso not opened")
		}
	})

	t.Run("CreateManual with schedule validation", func(t *testing.T) {
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, true, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				acceptingEnrollments := true
				return &models.Curso{
					ID: 1,
					LocationClasses: []models.LocationClass{
						{
							Schedules: []models.CourseSchedule{
								{
									ID:                   scheduleID,
									Vacancies:            10,
									AcceptingEnrollments: &acceptingEnrollments,
								},
							},
						},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 5, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			Name:       "Test User",
			Email:      "test@example.com",
			ScheduleID: &scheduleID,
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
	})

	t.Run("CreateManual with full schedule disables auto-approve", func(t *testing.T) {
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Should be pending since schedule is full
				if inscricao.Status != models.StatusInscricaoPending {
					t.Errorf("Expected status PENDING due to full schedule, got %s", inscricao.Status)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, true, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				acceptingEnrollments := true
				return &models.Curso{
					ID: 1,
					LocationClasses: []models.LocationClass{
						{
							Schedules: []models.CourseSchedule{
								{
									ID:                   scheduleID,
									Vacancies:            10,
									AcceptingEnrollments: &acceptingEnrollments,
								},
							},
						},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 10, nil // Full schedule
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			Name:       "Test User",
			Email:      "test@example.com",
			ScheduleID: &scheduleID,
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
	})

	t.Run("CreateManual fails with invalid schedule", func(t *testing.T) {
		invalidScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return &models.Curso{
					ID:              1,
					LocationClasses: []models.LocationClass{},
				}, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &invalidScheduleID,
		}

		err := svc.CreateManual(ctx, inscricao)
		if err == nil {
			t.Error("Expected error with invalid schedule ID")
		}
	})

	t.Run("CreateManual validates email with SanitizeEmail", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Valid email should be preserved
				if inscricao.Email != "test@example.com" {
					t.Errorf("Expected valid email 'test@example.com', got %s", inscricao.Email)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Test User",
			Email:   "test@example.com", // Valid email
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
	})

	t.Run("CreateManual handles invalid email", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Invalid email should be sanitized to empty string
				if inscricao.Email != "" {
					t.Errorf("Expected invalid email to be sanitized to empty string, got %s", inscricao.Email)
				}
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Test User",
			Email:   "not-a-valid-email", // Invalid email format
		}

		err := svc.CreateManual(ctx, inscricao)
		if err != nil {
			t.Fatalf("CreateManual failed: %v", err)
		}
	})
}

// MockCitizenSnapshotRepository for testing - implements CitizenSnapshotRepositoryInterface
type MockCitizenSnapshotRepository struct {
	GetByCPFFunc            func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
	GetByCPFsFunc           func(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error)
	UpsertFunc              func(ctx context.Context, snapshot *models.CitizenSnapshot) error
	BatchUpsertFunc         func(ctx context.Context, snapshots []*models.CitizenSnapshot) error
	GetStaleSnapshotsFunc   func(ctx context.Context, staleThreshold time.Duration, limit int) ([]*models.CitizenSnapshot, error)
	GetCPFsWithEnrollmentsFunc func(ctx context.Context, staleThreshold time.Duration, limit int) ([]string, error)
	DeleteFunc              func(ctx context.Context, cpf string) error
	CountFunc               func(ctx context.Context) (int64, error)
}

func (m *MockCitizenSnapshotRepository) GetByCPF(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
	if m.GetByCPFFunc != nil {
		return m.GetByCPFFunc(ctx, cpf)
	}
	return nil, nil
}

func (m *MockCitizenSnapshotRepository) GetByCPFs(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
	if m.GetByCPFsFunc != nil {
		return m.GetByCPFsFunc(ctx, cpfs)
	}
	return make(map[string]*models.CitizenSnapshot), nil
}

func (m *MockCitizenSnapshotRepository) Upsert(ctx context.Context, snapshot *models.CitizenSnapshot) error {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, snapshot)
	}
	return nil
}

func (m *MockCitizenSnapshotRepository) BatchUpsert(ctx context.Context, snapshots []*models.CitizenSnapshot) error {
	if m.BatchUpsertFunc != nil {
		return m.BatchUpsertFunc(ctx, snapshots)
	}
	return nil
}

func (m *MockCitizenSnapshotRepository) GetStaleSnapshots(ctx context.Context, staleThreshold time.Duration, limit int) ([]*models.CitizenSnapshot, error) {
	if m.GetStaleSnapshotsFunc != nil {
		return m.GetStaleSnapshotsFunc(ctx, staleThreshold, limit)
	}
	return []*models.CitizenSnapshot{}, nil
}

func (m *MockCitizenSnapshotRepository) GetCPFsWithEnrollments(ctx context.Context, staleThreshold time.Duration, limit int) ([]string, error) {
	if m.GetCPFsWithEnrollmentsFunc != nil {
		return m.GetCPFsWithEnrollmentsFunc(ctx, staleThreshold, limit)
	}
	return []string{}, nil
}

func (m *MockCitizenSnapshotRepository) Delete(ctx context.Context, cpf string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, cpf)
	}
	return nil
}

func (m *MockCitizenSnapshotRepository) Count(ctx context.Context) (int64, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx)
	}
	return 0, nil
}

// TestInscricaoService_EnrichWithPersonalInfo tests the EnrichWithPersonalInfo method
func TestInscricaoService_EnrichWithPersonalInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("EnrichWithPersonalInfo returns when citizenSnapshotRepo is nil", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			nil, // nil snapshot repo
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricao := &models.Inscricao{
			CPF: "12345678900",
		}

		// Should not panic with nil repo
		svc.EnrichWithPersonalInfo(ctx, inscricao)

		if inscricao.PersonalInfo != nil {
			t.Error("Expected PersonalInfo to remain nil when repo is nil")
		}
	})

	t.Run("EnrichWithPersonalInfo handles nil inscricao", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			nil,
			nil,
			nil,
			&config.AppConfig{},
		)

		// Should not panic with nil inscricao
		svc.EnrichWithPersonalInfo(ctx, nil)
	})

	t.Run("EnrichWithPersonalInfo handles empty CPF", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			nil,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricao := &models.Inscricao{
			CPF: "",
		}

		svc.EnrichWithPersonalInfo(ctx, inscricao)

		if inscricao.PersonalInfo != nil {
			t.Error("Expected PersonalInfo to remain nil for empty CPF")
		}
	})
}

// TestInscricaoService_EnrichMultipleWithPersonalInfo tests the EnrichMultipleWithPersonalInfo method
func TestInscricaoService_EnrichMultipleWithPersonalInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("EnrichMultipleWithPersonalInfo handles empty list", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			nil,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricoes := []*models.Inscricao{}

		// Should not panic with empty list
		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)
	})

	t.Run("EnrichMultipleWithPersonalInfo returns when repo is nil", func(t *testing.T) {
		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			nil,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricoes := []*models.Inscricao{
			{CPF: "11111111111"},
		}

		// Should return early when repo is nil
		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)

		if inscricoes[0].PersonalInfo != nil {
			t.Error("Expected PersonalInfo to remain nil when repo is nil")
		}
	})
}

// TestInscricaoService_ChangeSchedule tests the ChangeSchedule method
func TestInscricaoService_ChangeSchedule(t *testing.T) {
	ctx := context.Background()

	t.Run("ChangeSchedule success", func(t *testing.T) {
		inscricaoID := uuid.New()
		newScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.ScheduleID == nil || *inscricao.ScheduleID != newScheduleID {
					t.Error("Expected schedule ID to be updated")
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{
					newScheduleID: 5, // 5 enrolled, has space
				}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error) {
				acceptingEnrollments := true
				return &models.CourseSchedule{
					ID:                   scheduleID,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		futureDate := time.Now().Add(72 * time.Hour) // 3 days from now
		request := &models.ScheduleChangeRequest{
			ScheduleID: &newScheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				ID: "unit-1",
				Schedules: []models.EnrolledUnitSchedule{
					{
						ID:             newScheduleID.String(),
						ClassStartDate: futureDate.Format(time.RFC3339),
					},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{
			Enrollment: config.EnrollmentSettings{
				ScheduleChangeDeadlineHours: 48,
			},
		})

		result, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err != nil {
			t.Fatalf("ChangeSchedule failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected inscricao result")
		}
	})

	t.Run("ChangeSchedule fails when user CPF doesn't match", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "99999999999", request)
		if err == nil {
			t.Error("Expected error when CPF doesn't match")
		}
	})

	t.Run("ChangeSchedule fails for cancelled enrollment", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoCancelled,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error for cancelled enrollment")
		}
	})

	t.Run("ChangeSchedule fails within deadline", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		// Class starts in 24 hours, but deadline is 48 hours
		nearFutureDate := time.Now().Add(24 * time.Hour)
		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: nearFutureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{
			Enrollment: config.EnrollmentSettings{
				ScheduleChangeDeadlineHours: 48,
			},
		})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when changing schedule within deadline")
		}
	})

	t.Run("ChangeSchedule fails when schedule is full", func(t *testing.T) {
		inscricaoID := uuid.New()
		newScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{
					newScheduleID: 10, // Full
				}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error) {
				acceptingEnrollments := true
				return &models.CourseSchedule{
					ID:                   scheduleID,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &newScheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{
			Enrollment: config.EnrollmentSettings{
				ScheduleChangeDeadlineHours: 48,
			},
		})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when schedule is full")
		}
	})

	t.Run("ChangeSchedule with remote schedule", func(t *testing.T) {
		inscricaoID := uuid.New()
		remoteScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{
					remoteScheduleID: 3,
				}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error) {
				return nil, nil // Not a course schedule
			},
			GetRemoteScheduleByIDFunc: func(ctx context.Context, scheduleID uuid.UUID) (*models.RemoteSchedule, error) {
				acceptingEnrollments := true
				return &models.RemoteSchedule{
					ID:                   scheduleID,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &remoteScheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{
			Enrollment: config.EnrollmentSettings{
				ScheduleChangeDeadlineHours: 48,
			},
		})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err != nil {
			t.Fatalf("ChangeSchedule with remote schedule failed: %v", err)
		}
	})

	t.Run("ChangeSchedule fails when schedule not accepting enrollments", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 0}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				acceptingEnrollments := false
				return &models.CourseSchedule{
					ID:                   sid,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{
			Enrollment: config.EnrollmentSettings{
				ScheduleChangeDeadlineHours: 48,
			},
		})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error when schedule not accepting enrollments")
		}
	})

	t.Run("ChangeSchedule fails with invalid date format", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		request := &models.ScheduleChangeRequest{
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: "invalid-date-format"},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err == nil {
			t.Error("Expected error with invalid date format")
		}
	})

	t.Run("ChangeSchedule uses default deadline when config missing", func(t *testing.T) {
		inscricaoID := uuid.New()
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CPF:     "12345678900",
					CursoID: 1,
					Status:  models.StatusInscricaoApproved,
				}, nil
			},
			UpdateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{
			CountEnrollmentsByScheduleIDsFunc: func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
				return map[uuid.UUID]int64{scheduleID: 0}, nil
			},
			GetCourseScheduleByIDFunc: func(ctx context.Context, sid uuid.UUID) (*models.CourseSchedule, error) {
				acceptingEnrollments := true
				return &models.CourseSchedule{
					ID:                   sid,
					Vacancies:            10,
					AcceptingEnrollments: &acceptingEnrollments,
				}, nil
			},
		}

		// 72 hours should be enough for default 48 hour deadline
		futureDate := time.Now().Add(72 * time.Hour)
		request := &models.ScheduleChangeRequest{
			ScheduleID: &scheduleID,
			EnrolledUnit: &models.EnrolledUnit{
				Schedules: []models.EnrolledUnitSchedule{
					{ClassStartDate: futureDate.Format(time.RFC3339)},
				},
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, nil)

		_, err := svc.ChangeSchedule(ctx, inscricaoID, "12345678900", request)
		if err != nil {
			t.Fatalf("ChangeSchedule should use default 48h deadline: %v", err)
		}
	})
}
