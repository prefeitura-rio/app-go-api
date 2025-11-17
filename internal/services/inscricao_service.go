package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type InscricaoService struct {
	repo      *repository.InscricaoRepository
	cursoRepo *repository.CursoRepository
}

func NewInscricaoService(repo *repository.InscricaoRepository, cursoRepo *repository.CursoRepository) *InscricaoService {
	return &InscricaoService{
		repo:      repo,
		cursoRepo: cursoRepo,
	}
}

func (s *InscricaoService) Create(ctx context.Context, inscricao *models.Inscricao) error {
	// Validate course exists and can accept enrollments (lightweight query)
	status, enrollmentStart, enrollmentEnd, err := s.cursoRepo.ValidateForEnrollment(ctx, inscricao.CursoID)
	if err != nil {
		return err
	}

	// Check if enrollment is open
	if models.StatusCurso(status) != models.StatusCursoOpened {
		return fmt.Errorf("curso não está aberto para inscrições")
	}

	// Check enrollment dates
	now := time.Now()
	if enrollmentStart != nil && now.Before(*enrollmentStart) {
		return fmt.Errorf("período de inscrições ainda não iniciou")
	}
	if enrollmentEnd != nil && now.After(*enrollmentEnd) {
		return fmt.Errorf("período de inscrições já encerrou")
	}

	// Check if CPF is already enrolled
	exists, err := s.repo.ExistsByCPFAndCurso(ctx, inscricao.CPF, inscricao.CursoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição existente: %w", err)
	}
	if exists {
		return fmt.Errorf("CPF já inscrito neste curso")
	}

	// Validate schedule_id if provided
	// Only load full course with schedules when schedule validation is needed
	if inscricao.ScheduleID != nil {
		curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
		if err != nil {
			return fmt.Errorf("erro ao verificar curso para validação de schedule: %w", err)
		}
		if curso == nil {
			return fmt.Errorf("curso não encontrado")
		}

		if err := s.validateScheduleID(ctx, *inscricao.ScheduleID, curso); err != nil {
			return err
		}
	}

	// Set default values
	inscricao.Status = models.StatusInscricaoPending
	inscricao.EnrolledAt = time.Now()
	inscricao.UpdatedAt = time.Now()

	return s.repo.Create(ctx, inscricao)
}

func (s *InscricaoService) GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *InscricaoService) GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, page, pageSize int) ([]*models.Inscricao, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetByCursoID(ctx, cursoID, filter, pageSize, offset)
}

func (s *InscricaoService) UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
	// Validate enrollment exists
	inscricao, err := s.repo.GetByID(ctx, inscricaoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição: %w", err)
	}
	if inscricao == nil {
		return fmt.Errorf("inscrição não encontrada")
	}

	return s.repo.UpdateStatus(ctx, inscricaoID, status, reason, adminNotes)
}

func (s *InscricaoService) UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
	// Validate at least one ID provided
	if len(inscricaoIDs) == 0 {
		return 0, fmt.Errorf("nenhuma inscrição selecionada")
	}

	return s.repo.UpdateMultipleStatus(ctx, inscricaoIDs, status, reason, adminNotes)
}

func (s *InscricaoService) GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
	// Validate course exists (lightweight query - only checks existence)
	_, _, _, err := s.cursoRepo.ValidateForEnrollment(ctx, cursoID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetSummaryByCursoID(ctx, cursoID)
}

func (s *InscricaoService) Delete(ctx context.Context, id uuid.UUID) error {
	// Validate enrollment exists
	inscricao, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição: %w", err)
	}
	if inscricao == nil {
		return fmt.Errorf("inscrição não encontrada")
	}

	return s.repo.Delete(ctx, id)
}

func (s *InscricaoService) ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
	return s.repo.ListByCPF(ctx, cpf, filter, offset, limit)
}

func (s *InscricaoService) UpdateCertificate(ctx context.Context, cursoID int, inscricaoID uuid.UUID, certificateURL string) error {
	// Verify enrollment exists and belongs to the course
	inscricao, err := s.repo.GetByID(ctx, inscricaoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição: %w", err)
	}
	if inscricao == nil {
		return fmt.Errorf("inscrição não encontrada")
	}
	if inscricao.CursoID != cursoID {
		return fmt.Errorf("inscrição não pertence ao curso especificado")
	}

	// Only allow certificate for approved or concluded enrollments
	if inscricao.Status != models.StatusInscricaoApproved && inscricao.Status != models.StatusInscricaoConcluded {
		return fmt.Errorf("certificado só pode ser atribuído a inscrições aprovadas ou concluídas")
	}

	// Update certificate URL
	return s.repo.UpdateCertificate(ctx, inscricaoID, certificateURL)
}

func (s *InscricaoService) UpdateInscricao(ctx context.Context, id uuid.UUID, cursoID int, updateData *models.InscricaoUpdateRequest) error {
	// Buscar inscrição existente
	inscricao, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao verificar inscrição: %w", err)
	}
	if inscricao == nil {
		return fmt.Errorf("inscrição não encontrada")
	}

	// Validar que a inscrição pertence ao curso especificado
	if inscricao.CursoID != cursoID {
		return fmt.Errorf("inscrição não pertence ao curso especificado")
	}

	// Atualizar apenas os campos permitidos
	if updateData.Name != nil {
		inscricao.Name = *updateData.Name
	}
	if updateData.Email != nil {
		inscricao.Email = *updateData.Email
	}
	if updateData.Phone != nil {
		inscricao.Phone = *updateData.Phone
	}
	if updateData.CustomFieldsData != nil {
		inscricao.CustomFieldsData = updateData.CustomFieldsData
	}
	if updateData.AdminNotes != nil {
		inscricao.AdminNotes = *updateData.AdminNotes
	}
	if updateData.EnrolledUnit != nil {
		inscricao.EnrolledUnit = updateData.EnrolledUnit
	}

	inscricao.UpdatedAt = time.Now()

	return s.repo.Update(ctx, inscricao)
}

// validateScheduleID validates that the schedule exists and belongs to the course
func (s *InscricaoService) validateScheduleID(ctx context.Context, scheduleID uuid.UUID, curso *models.Curso) error {
	// Check all locations and schedules in the course
	for _, location := range curso.LocationClasses {
		for _, schedule := range location.Schedules {
			if schedule.ID == scheduleID {
				// Schedule found and belongs to this course
				return nil
			}
		}
	}

	// Check remote class schedules
	if curso.RemoteClass != nil {
		for _, schedule := range curso.RemoteClass.Schedules {
			if schedule.ID == scheduleID {
				// Schedule found and belongs to this course
				return nil
			}
		}
	}

	return fmt.Errorf("schedule_id fornecido não pertence a este curso")
}
