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
	repo     *repository.InscricaoRepository
	cursoRepo *repository.CursoRepository
}

func NewInscricaoService(repo *repository.InscricaoRepository, cursoRepo *repository.CursoRepository) *InscricaoService {
	return &InscricaoService{
		repo:     repo,
		cursoRepo: cursoRepo,
	}
}

func (s *InscricaoService) Create(ctx context.Context, inscricao *models.Inscricao) error {
	// Validate course exists
	curso, err := s.cursoRepo.GetByID(ctx, inscricao.CursoID)
	if err != nil {
		return fmt.Errorf("erro ao verificar curso: %w", err)
	}
	if curso == nil {
		return fmt.Errorf("curso não encontrado")
	}
	
	// Check if enrollment is open
	if curso.Status != models.StatusCursoOpened {
		return fmt.Errorf("curso não está aberto para inscrições")
	}
	
	// Check enrollment dates
	now := time.Now()
	if curso.EnrollmentStartDate != nil && now.Before(*curso.EnrollmentStartDate) {
		return fmt.Errorf("período de inscrições ainda não iniciou")
	}
	if curso.EnrollmentEndDate != nil && now.After(*curso.EnrollmentEndDate) {
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
	// Validate course exists
	curso, err := s.cursoRepo.GetByID(ctx, cursoID)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar curso: %w", err)
	}
	if curso == nil {
		return nil, fmt.Errorf("curso não encontrado")
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