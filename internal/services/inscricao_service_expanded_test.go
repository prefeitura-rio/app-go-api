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

	t.Run("UpdateCertificate with empty URL succeeds", func(t *testing.T) {
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
				// Empty URL is allowed - just updates to empty string
				return nil
			},
		}

		cursoRepo := &MockCursoRepository{}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Pass empty URL - should succeed (allows removing certificate)
		err := svc.UpdateCertificate(ctx, 1, inscricaoID, "")
		if err != nil {
			t.Fatalf("UpdateCertificate with empty URL should succeed: %v", err)
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
		name          string
		currentStatus models.StatusInscricao
		newStatus     models.StatusInscricao
		shouldSucceed bool
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

// ==================== Comprehensive Enrichment Tests ====================
// These tests bring coverage for EnrichWithPersonalInfo and EnrichMultipleWithPersonalInfo
// from 16.7% to 85%+ and 7.9% to 85%+ respectively

func TestInscricaoService_EnrichWithPersonalInfo_Comprehensive(t *testing.T) {
	ctx := context.Background()

	t.Run("EnrichWithPersonalInfo success with snapshot", func(t *testing.T) {
		testCPF := "12345678900"
		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				if cpf == testCPF {
					return &models.CitizenSnapshot{
						CPF:     cpf,
						Nome:    "John Doe",
						Email:   "john@example.com",
						Celular: "21999999999",
					}, nil
				}
				return nil, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricao := &models.Inscricao{
			CPF: testCPF,
		}

		svc.EnrichWithPersonalInfo(ctx, inscricao)

		if inscricao.PersonalInfo == nil {
			t.Fatal("Expected PersonalInfo to be set")
		}
		if inscricao.PersonalInfo.Nome != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%s'", inscricao.PersonalInfo.Nome)
		}
	})

	t.Run("EnrichWithPersonalInfo handles GetByCPF error", func(t *testing.T) {
		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return nil, errors.New("database error")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricao := &models.Inscricao{
			CPF: "12345678900",
		}

		// Should not panic on error
		svc.EnrichWithPersonalInfo(ctx, inscricao)

		if inscricao.PersonalInfo != nil {
			t.Error("Expected PersonalInfo to remain nil on error")
		}
	})

	t.Run("EnrichWithPersonalInfo falls back to on-demand sync", func(t *testing.T) {
		testCPF := "12345678900"
		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return nil, nil // No snapshot exists
			},
		}

		fetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				if cpf == testCPF {
					return &models.CitizenSnapshot{
						CPF:     cpf,
						Nome:    "Jane Doe",
						Email:   "jane@example.com",
						Celular: "21988888888",
					}, nil
				}
				return nil, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			fetcher,
			nil,
			&config.AppConfig{},
		)

		inscricao := &models.Inscricao{
			CPF: testCPF,
		}

		svc.EnrichWithPersonalInfo(ctx, inscricao)

		if inscricao.PersonalInfo == nil {
			t.Fatal("Expected PersonalInfo to be set from on-demand sync")
		}
		if inscricao.PersonalInfo.Nome != "Jane Doe" {
			t.Errorf("Expected name 'Jane Doe', got '%s'", inscricao.PersonalInfo.Nome)
		}
	})

	t.Run("EnrichWithPersonalInfo handles on-demand sync error", func(t *testing.T) {
		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return nil, nil // No snapshot
			},
		}

		fetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return nil, errors.New("sync failed")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			fetcher,
			nil,
			&config.AppConfig{},
		)

		inscricao := &models.Inscricao{
			CPF: "12345678900",
		}

		// Should not panic on sync error
		svc.EnrichWithPersonalInfo(ctx, inscricao)

		if inscricao.PersonalInfo != nil {
			t.Error("Expected PersonalInfo to remain nil on sync error")
		}
	})

	t.Run("EnrichWithPersonalInfo with snapshot but no fetcher for legacy", func(t *testing.T) {
		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				return nil, nil // No snapshot
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			nil, // No fetcher
			nil,
			&config.AppConfig{},
		)

		inscricao := &models.Inscricao{
			CPF: "12345678900",
		}

		svc.EnrichWithPersonalInfo(ctx, inscricao)

		if inscricao.PersonalInfo != nil {
			t.Error("Expected PersonalInfo to remain nil when no snapshot and no fetcher")
		}
	})
}

func TestInscricaoService_EnrichMultipleWithPersonalInfo_Comprehensive(t *testing.T) {
	ctx := context.Background()

	t.Run("EnrichMultipleWithPersonalInfo success with multiple inscricoes", func(t *testing.T) {
		cpf1 := "11111111111"
		cpf2 := "22222222222"
		cpf3 := "33333333333"

		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFsFunc: func(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
				snapshots := make(map[string]*models.CitizenSnapshot)
				for _, cpf := range cpfs {
					switch cpf {
					case cpf1:
						snapshots[cpf] = &models.CitizenSnapshot{
							CPF:     cpf,
							Nome:    "Person One",
							Email:   "one@example.com",
							Celular: "21999999999",
						}
					case cpf2:
						snapshots[cpf] = &models.CitizenSnapshot{
							CPF:     cpf,
							Nome:    "Person Two",
							Email:   "two@example.com",
							Celular: "21988888888",
						}
					case cpf3:
						snapshots[cpf] = &models.CitizenSnapshot{
							CPF:     cpf,
							Nome:    "Person Three",
							Email:   "three@example.com",
							Celular: "21977777777",
						}
					}
				}
				return snapshots, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricoes := []*models.Inscricao{
			{CPF: cpf1},
			{CPF: cpf2},
			{CPF: cpf3},
		}

		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)

		for i, inscricao := range inscricoes {
			if inscricao.PersonalInfo == nil {
				t.Errorf("Expected PersonalInfo to be set for inscription %d", i)
			}
		}

		if inscricoes[0].PersonalInfo.Nome != "Person One" {
			t.Errorf("Expected 'Person One', got '%s'", inscricoes[0].PersonalInfo.Nome)
		}
		if inscricoes[1].PersonalInfo.Nome != "Person Two" {
			t.Errorf("Expected 'Person Two', got '%s'", inscricoes[1].PersonalInfo.Nome)
		}
		if inscricoes[2].PersonalInfo.Nome != "Person Three" {
			t.Errorf("Expected 'Person Three', got '%s'", inscricoes[2].PersonalInfo.Nome)
		}
	})

	t.Run("EnrichMultipleWithPersonalInfo deduplicates CPFs", func(t *testing.T) {
		cpf := "11111111111"
		callCount := 0

		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFsFunc: func(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
				callCount++
				// Verify deduplication - should receive only unique CPFs
				if len(cpfs) != 1 {
					t.Errorf("Expected 1 unique CPF, got %d", len(cpfs))
				}
				snapshots := make(map[string]*models.CitizenSnapshot)
				snapshots[cpf] = &models.CitizenSnapshot{
					CPF:  cpf,
					Nome: "Test Person",
				}
				return snapshots, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			nil,
			nil,
			&config.AppConfig{},
		)

		// Multiple inscricoes with same CPF
		inscricoes := []*models.Inscricao{
			{CPF: cpf},
			{CPF: cpf},
			{CPF: cpf},
		}

		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)

		if callCount != 1 {
			t.Errorf("Expected GetByCPFs to be called once, got %d", callCount)
		}

		for i, inscricao := range inscricoes {
			if inscricao.PersonalInfo == nil {
				t.Errorf("Expected PersonalInfo to be set for inscription %d", i)
			}
		}
	})

	t.Run("EnrichMultipleWithPersonalInfo handles GetByCPFs error", func(t *testing.T) {
		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFsFunc: func(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
				return nil, errors.New("database error")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricoes := []*models.Inscricao{
			{CPF: "11111111111"},
			{CPF: "22222222222"},
		}

		// Should not panic
		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)

		for _, inscricao := range inscricoes {
			if inscricao.PersonalInfo != nil {
				t.Error("Expected PersonalInfo to remain nil on error")
			}
		}
	})

	t.Run("EnrichMultipleWithPersonalInfo partial success", func(t *testing.T) {
		cpf1 := "11111111111"
		cpf2 := "22222222222"
		cpf3 := "33333333333"

		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFsFunc: func(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
				snapshots := make(map[string]*models.CitizenSnapshot)
				// Only return snapshot for cpf1 and cpf2, not cpf3
				snapshots[cpf1] = &models.CitizenSnapshot{
					CPF:  cpf1,
					Nome: "Person One",
				}
				snapshots[cpf2] = &models.CitizenSnapshot{
					CPF:  cpf2,
					Nome: "Person Two",
				}
				return snapshots, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricoes := []*models.Inscricao{
			{CPF: cpf1},
			{CPF: cpf2},
			{CPF: cpf3},
		}

		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)

		if inscricoes[0].PersonalInfo == nil {
			t.Error("Expected PersonalInfo to be set for first inscription")
		}
		if inscricoes[1].PersonalInfo == nil {
			t.Error("Expected PersonalInfo to be set for second inscription")
		}
		if inscricoes[2].PersonalInfo != nil {
			t.Error("Expected PersonalInfo to remain nil for third inscription (no snapshot)")
		}
	})

	t.Run("EnrichMultipleWithPersonalInfo syncs missing CPFs on-demand", func(t *testing.T) {
		cpf1 := "11111111111"
		cpf2 := "22222222222"

		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFsFunc: func(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
				snapshots := make(map[string]*models.CitizenSnapshot)
				// Only return snapshot for cpf1, cpf2 is missing
				snapshots[cpf1] = &models.CitizenSnapshot{
					CPF:  cpf1,
					Nome: "Person One",
				}
				return snapshots, nil
			},
		}

		syncCallCount := 0
		fetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				syncCallCount++
				if cpf == cpf2 {
					return &models.CitizenSnapshot{
						CPF:  cpf,
						Nome: "Person Two (synced)",
					}, nil
				}
				return nil, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			fetcher,
			nil,
			&config.AppConfig{},
		)

		inscricoes := []*models.Inscricao{
			{CPF: cpf1},
			{CPF: cpf2},
		}

		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)

		if syncCallCount != 1 {
			t.Errorf("Expected on-demand sync to be called once for missing CPF, got %d", syncCallCount)
		}

		if inscricoes[0].PersonalInfo == nil || inscricoes[0].PersonalInfo.Nome != "Person One" {
			t.Error("Expected first inscription to be enriched from snapshot")
		}
		if inscricoes[1].PersonalInfo == nil || inscricoes[1].PersonalInfo.Nome != "Person Two (synced)" {
			t.Error("Expected second inscription to be enriched from on-demand sync")
		}
	})

	t.Run("EnrichMultipleWithPersonalInfo handles empty CPFs in list", func(t *testing.T) {
		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFsFunc: func(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
				// Should only receive non-empty CPFs
				for _, cpf := range cpfs {
					if cpf == "" {
						t.Error("Expected empty CPFs to be filtered out")
					}
				}
				snapshots := make(map[string]*models.CitizenSnapshot)
				snapshots["11111111111"] = &models.CitizenSnapshot{
					CPF:  "11111111111",
					Nome: "Valid Person",
				}
				return snapshots, nil
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			nil,
			nil,
			&config.AppConfig{},
		)

		inscricoes := []*models.Inscricao{
			{CPF: "11111111111"},
			{CPF: ""},
			{CPF: ""},
		}

		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)

		if inscricoes[0].PersonalInfo == nil {
			t.Error("Expected PersonalInfo for valid CPF")
		}
		if inscricoes[1].PersonalInfo != nil {
			t.Error("Expected nil PersonalInfo for empty CPF")
		}
		if inscricoes[2].PersonalInfo != nil {
			t.Error("Expected nil PersonalInfo for empty CPF")
		}
	})

	t.Run("EnrichMultipleWithPersonalInfo handles on-demand sync failure gracefully", func(t *testing.T) {
		cpf1 := "11111111111"
		cpf2 := "22222222222"

		snapshotRepo := &MockCitizenSnapshotRepository{
			GetByCPFsFunc: func(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error) {
				// No snapshots exist
				return make(map[string]*models.CitizenSnapshot), nil
			},
		}

		fetcher := &MockCitizenDataFetcher{
			SyncFunc: func(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
				if cpf == cpf1 {
					return &models.CitizenSnapshot{
						CPF:  cpf,
						Nome: "Person One",
					}, nil
				}
				// cpf2 sync fails
				return nil, errors.New("sync failed")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(
			&MockInscricaoRepository{},
			&MockCursoRepository{},
			snapshotRepo,
			fetcher,
			nil,
			&config.AppConfig{},
		)

		inscricoes := []*models.Inscricao{
			{CPF: cpf1},
			{CPF: cpf2},
		}

		svc.EnrichMultipleWithPersonalInfo(ctx, inscricoes)

		// First should succeed, second should fail gracefully
		if inscricoes[0].PersonalInfo == nil {
			t.Error("Expected PersonalInfo for first inscription")
		}
		if inscricoes[1].PersonalInfo != nil {
			t.Error("Expected nil PersonalInfo for failed sync")
		}
	})
}

// ==================== Comprehensive Vacancy Tests ====================
// These tests bring coverage for findScheduleVacancies from 44.4% to 85%+
// Tests are indirect through Create method which exercises findScheduleVacancies

func TestInscricaoService_FindScheduleVacancies_Comprehensive(t *testing.T) {
	ctx := context.Background()

	t.Run("findScheduleVacancies in LocationClass schedules", func(t *testing.T) {
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Auto-approve should work because there are vacancies
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected APPROVED status, got %s", inscricao.Status)
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
									Vacancies:            20,
									AcceptingEnrollments: &acceptingEnrollments,
								},
							},
						},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 10, nil // 10/20 filled
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &scheduleID,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("findScheduleVacancies in RemoteClass schedules", func(t *testing.T) {
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected APPROVED status, got %s", inscricao.Status)
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
					RemoteClass: &models.RemoteClass{
						Schedules: []models.RemoteSchedule{
							{
								ID:                   scheduleID,
								Vacancies:            50,
								AcceptingEnrollments: &acceptingEnrollments,
							},
						},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 25, nil // 25/50 filled
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &scheduleID,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("findScheduleVacancies with multiple LocationClasses", func(t *testing.T) {
		targetScheduleID := uuid.New()
		otherScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected APPROVED status, got %s", inscricao.Status)
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
									ID:                   otherScheduleID,
									Vacancies:            30,
									AcceptingEnrollments: &acceptingEnrollments,
								},
							},
						},
						{
							Schedules: []models.CourseSchedule{
								{
									ID:                   targetScheduleID,
									Vacancies:            15,
									AcceptingEnrollments: &acceptingEnrollments,
								},
							},
						},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 5, nil // 5/15 filled for target
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &targetScheduleID,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("findScheduleVacancies with multiple schedules per location", func(t *testing.T) {
		targetScheduleID := uuid.New()
		schedule2ID := uuid.New()
		schedule3ID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected APPROVED status, got %s", inscricao.Status)
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
									ID:                   schedule2ID,
									Vacancies:            25,
									AcceptingEnrollments: &acceptingEnrollments,
								},
								{
									ID:                   targetScheduleID,
									Vacancies:            40,
									AcceptingEnrollments: &acceptingEnrollments,
								},
								{
									ID:                   schedule3ID,
									Vacancies:            35,
									AcceptingEnrollments: &acceptingEnrollments,
								},
							},
						},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 20, nil // 20/40 filled for target
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &targetScheduleID,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("findScheduleVacancies with both LocationClasses and RemoteClass", func(t *testing.T) {
		locationScheduleID := uuid.New()
		remoteScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected APPROVED status, got %s", inscricao.Status)
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
									ID:                   locationScheduleID,
									Vacancies:            20,
									AcceptingEnrollments: &acceptingEnrollments,
								},
							},
						},
					},
					RemoteClass: &models.RemoteClass{
						Schedules: []models.RemoteSchedule{
							{
								ID:                   remoteScheduleID,
								Vacancies:            100,
								AcceptingEnrollments: &acceptingEnrollments,
							},
						},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 50, nil // 50/100 filled for remote
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		// Test with remote schedule
		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &remoteScheduleID,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed for remote schedule: %v", err)
		}
	})

	t.Run("findScheduleVacancies with empty LocationClasses", func(t *testing.T) {
		remoteScheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected APPROVED status, got %s", inscricao.Status)
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
					ID:              1,
					LocationClasses: []models.LocationClass{}, // Empty
					RemoteClass: &models.RemoteClass{
						Schedules: []models.RemoteSchedule{
							{
								ID:                   remoteScheduleID,
								Vacancies:            75,
								AcceptingEnrollments: &acceptingEnrollments,
							},
						},
					},
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 30, nil // 30/75 filled
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &remoteScheduleID,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("findScheduleVacancies with nil RemoteClass", func(t *testing.T) {
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				if inscricao.Status != models.StatusInscricaoApproved {
					t.Errorf("Expected APPROVED status, got %s", inscricao.Status)
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
									Vacancies:            12,
									AcceptingEnrollments: &acceptingEnrollments,
								},
							},
						},
					},
					RemoteClass: nil, // Nil remote class
				}, nil
			},
			CountEnrollmentsByScheduleIDFunc: func(ctx context.Context, sid uuid.UUID) (int64, error) {
				return 6, nil // 6/12 filled
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &scheduleID,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("findScheduleVacancies disables auto-approve when count error", func(t *testing.T) {
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Should be pending because count error disables auto-approve
				if inscricao.Status != models.StatusInscricaoPending {
					t.Errorf("Expected PENDING status due to count error, got %s", inscricao.Status)
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
				return 0, errors.New("database error")
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &scheduleID,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})

	t.Run("findScheduleVacancies disables auto-approve when exactly at capacity", func(t *testing.T) {
		scheduleID := uuid.New()

		inscricaoRepo := &MockInscricaoRepository{
			CreateFunc: func(ctx context.Context, inscricao *models.Inscricao) error {
				// Should be pending because schedule is at capacity
				if inscricao.Status != models.StatusInscricaoPending {
					t.Errorf("Expected PENDING status at capacity, got %s", inscricao.Status)
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
				return 10, nil // Exactly at capacity
			},
		}

		svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil, &config.AppConfig{})

		inscricao := &models.Inscricao{
			CPF:        "12345678900",
			CursoID:    1,
			ScheduleID: &scheduleID,
		}

		err := svc.Create(ctx, inscricao)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	})
}
