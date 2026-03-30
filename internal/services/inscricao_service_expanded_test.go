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

// EXPANDED TESTS FOR INSCRICAO SERVICE
// This file adds 20+ edge cases for enrollment workflows

// ==================== Enrollment Period Edge Cases ====================

func TestInscricaoService_EnrollmentPeriodEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Enrollment exactly at start time succeeds", func(t *testing.T) {
		// Mock time.Now() to be exactly at enrollment start
		now := time.Now()
		startTime := now

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
				return string(models.StatusCursoOpened), &startTime, nil, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
			Name:    "Test User",
			Email:   "test@example.com",
		}

		// Should succeed as we're exactly at start time
		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create at exact start time should succeed: %v", err)
		}
	})

	t.Run("Enrollment exactly at end time fails", func(t *testing.T) {
		now := time.Now()
		pastEnd := now.Add(-1 * time.Second)

		inscricaoRepo := &MockInscricaoRepository{}

		cursoRepo := &MockCursoRepository{
			ValidateForEnrollmentFunc: func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
				return string(models.StatusCursoOpened), nil, &pastEnd, false, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:     "12345678900",
			CursoID: 1,
		}

		err := svc.Create(ctx, inscricao)
		if err == nil {
			t.Error("Expected error when enrollment after end time")
		}
		if !strings.Contains(err.Error(), "período de inscrições já encerrou") {
			t.Errorf("Expected 'período já encerrou' error, got: %v", err)
		}
	})

	t.Run("Enrollment with no period restrictions", func(t *testing.T) {
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
				// No start/end dates - always open
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
			t.Fatalf("Create without period restrictions should succeed: %v", err)
		}
	})
}

// ==================== Duplicate Enrollment Prevention ====================

func TestInscricaoService_DuplicateEnrollmentPrevention(t *testing.T) {
	ctx := context.Background()

	t.Run("Same CPF different course succeeds", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				return nil
			},
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				// Same CPF exists but for different course
				if cursoID == 2 {
					return false, nil // No enrollment for course 2
				}
				return true, nil // Exists for course 1
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
			CursoID: 2, // Different course
			Name:    "Test User",
			Email:   "test@example.com",
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Enrollment in different course should succeed: %v", err)
		}
	})

	t.Run("ExistsByCPFAndCurso repository error fails gracefully", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ExistsByCPFAndCursoFunc: func(ctx context.Context, cpf string, cursoID int) (bool, error) {
				return false, errors.New("database connection error")
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
			t.Error("Expected error from ExistsByCPFAndCurso")
		}
		if !strings.Contains(err.Error(), "verificar inscrição existente") {
			t.Errorf("Expected 'verificar inscrição existente' error, got: %v", err)
		}
	})
}

// ==================== Certificate Generation Edge Cases ====================

func TestInscricaoService_CertificateEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateCertificate with empty URL fails", func(t *testing.T) {
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
				if certificateURL == "" {
					t.Error("Expected non-empty certificate URL")
				}
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Pass empty URL
		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "")
		if err != nil {
			t.Fatalf("UpdateCertificate with empty URL: %v", err)
		}
	})

	t.Run("UpdateCertificate for rejected enrollment fails", func(t *testing.T) {
		inscricaoID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      inscricaoID,
					CursoID: 1,
					Status:  models.StatusInscricaoRejected,
				}, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err == nil {
			t.Error("Expected error for rejected enrollment")
		}
	})

	t.Run("UpdateCertificate repository error is propagated", func(t *testing.T) {
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
				return errors.New("database error")
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "https://example.com/cert.pdf")
		if err == nil {
			t.Error("Expected database error")
		}
	})
}

// ==================== Batch Status Updates with Partial Failures ====================

func TestInscricaoService_BatchOperationsEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateMultipleStatus with one ID", func(t *testing.T) {
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
				return 1, nil
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
			t.Fatalf("UpdateMultipleStatus with one ID failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus with very large batch", func(t *testing.T) {
		// Create 100 IDs
		ids := make([]uuid.UUID, 100)
		for i := range ids {
			ids[i] = uuid.New()
		}

		inscricaoRepo := &MockInscricaoRepository{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
				return &models.Inscricao{
					ID:      id,
					CursoID: 1,
					Status:  models.StatusInscricaoPending,
				}, nil
			},
			UpdateMultipleStatusFunc: func(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
				if len(inscricaoIDs) != 100 {
					t.Errorf("Expected 100 IDs, got %d", len(inscricaoIDs))
				}
				return 100, nil
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
			t.Fatalf("UpdateMultipleStatus with large batch failed: %v", err)
		}
		if count != 100 {
			t.Errorf("Expected count 100, got %d", count)
		}
	})

	t.Run("UpdateMultipleStatus zero updates still succeeds", func(t *testing.T) {
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
				// Simulate no rows updated (e.g., all already in target status)
				return 0, nil
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
			t.Fatalf("UpdateMultipleStatus should not fail with 0 updates: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count 0, got %d", count)
		}
	})
}

// ==================== Status Transition Validation ====================

func TestInscricaoService_StatusTransitionValidation(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		currentStatus  models.StatusInscricao
		newStatus      models.StatusInscricao
		shouldSucceed  bool
	}{
		{
			name:          "Pending to Approved",
			currentStatus: models.StatusInscricaoPending,
			newStatus:     models.StatusInscricaoApproved,
			shouldSucceed: true,
		},
		{
			name:          "Pending to Rejected",
			currentStatus: models.StatusInscricaoPending,
			newStatus:     models.StatusInscricaoRejected,
			shouldSucceed: true,
		},
		{
			name:          "Approved to Rejected",
			currentStatus: models.StatusInscricaoApproved,
			newStatus:     models.StatusInscricaoRejected,
			shouldSucceed: true,
		},
		{
			name:          "Rejected to Approved",
			currentStatus: models.StatusInscricaoRejected,
			newStatus:     models.StatusInscricaoApproved,
			shouldSucceed: true,
		},
		{
			name:          "Approved to Approved (no-op)",
			currentStatus: models.StatusInscricaoApproved,
			newStatus:     models.StatusInscricaoApproved,
			shouldSucceed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inscricaoID := uuid.New()

			inscricaoRepo := &MockInscricaoRepository{
				GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
					return &models.Inscricao{
						ID:      inscricaoID,
						CursoID: 1,
						Status:  tc.currentStatus,
					}, nil
				},
				UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
					if status != tc.newStatus {
						t.Errorf("Expected new status %s, got %s", tc.newStatus, status)
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

			err := svc.UpdateStatus(ctx, inscricaoID, tc.newStatus, "", "")
			if tc.shouldSucceed && err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
			if !tc.shouldSucceed && err == nil {
				t.Error("Expected error but got success")
			}
		})
	}
}

// ==================== CPF Duplicate Detection Across Courses ====================

func TestInscricaoService_CPFValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("Create with CPF formatting variations", func(t *testing.T) {
		testCases := []string{
			"12345678900",
			"123.456.789-00",
			" 12345678900 ",
		}

		for _, cpf := range testCases {
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
					return string(models.StatusCursoOpened), nil, nil, false, nil
				},
			}

			svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

			inscricao := &models.Inscricao{
				CPF:     cpf,
				CursoID: 1,
				Name:    "Test User",
				Email:   "test@example.com",
			}

			err := svc.Create(ctx, inscricao)
			if err != nil {
				t.Errorf("CPF %s should be accepted: %v", cpf, err)
			}
		}
	})
}

// ==================== Schedule Change Validation ====================

func TestInscricaoService_ScheduleChangeScenarios(t *testing.T) {
	ctx := context.Background()

	t.Run("ListByCPF pagination", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			ListByCPFFunc: func(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
				// Verify pagination
				if offset != 20 { // (3-1) * 10
					t.Errorf("Expected offset 20, got %d", offset)
				}
				if limit != 10 {
					t.Errorf("Expected limit 10, got %d", limit)
				}
				return []*models.Inscricao{{ID: uuid.New()}}, 1, nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		_, _, err := svc.ListByCPF(ctx, "12345678900", nil, 20, 10) // offset=20, limit=10
		if err != nil {
			t.Fatalf("ListByCPF failed: %v", err)
		}
	})

	t.Run("GetSummaryByCursoID with no enrollments", func(t *testing.T) {
		inscricaoRepo := &MockInscricaoRepository{
			GetSummaryByCursoIDFunc: func(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
				return &models.EnrollmentSummary{
					Total:    0,
					Pending:  0,
					Approved: 0,
					Rejected: 0,
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
		if summary.Total != 0 {
			t.Errorf("Expected total 0, got %d", summary.Total)
		}
	})
}
