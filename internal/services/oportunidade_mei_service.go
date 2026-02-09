package services

import (
	"context"
	"errors"

	"github.com/prefeitura-rio/app-go-api/internal/logger"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type OportunidadeMEIService struct {
	repo OportunidadeMEIRepositoryInterface
}

func NewOportunidadeMEIService(
	repo OportunidadeMEIRepositoryInterface,
) *OportunidadeMEIService {
	return &OportunidadeMEIService{
		repo: repo,
	}
}

func (s *OportunidadeMEIService) Create(ctx context.Context, oportunidade *models.OportunidadeMEI, isDraft bool) (int, error) {
	// Setar status ANTES da validação
	if isDraft {
		oportunidade.Status = models.StatusOportunidadeDraft
	} else {
		oportunidade.Status = models.StatusOportunidadeActive
	}

	// Agora validar com status já definido
	if err := oportunidade.Validate(); err != nil {
		return 0, err
	}

	return s.repo.Create(ctx, oportunidade)
}

func (s *OportunidadeMEIService) GetByID(ctx context.Context, id int) (*models.OportunidadeMEI, error) {
	oportunidade, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if oportunidade != nil {
		oportunidade.UpdateStatusBasedOnExpiration()
	}

	return oportunidade, nil
}

func (s *OportunidadeMEIService) Update(ctx context.Context, oportunidade *models.OportunidadeMEI) error {
	// Buscar a oportunidade existente
	existing, err := s.repo.GetByID(ctx, oportunidade.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("oportunidade não encontrada")
	}

	// Preservar o status existente ao atualizar (não publica automaticamente)
	oportunidade.Status = existing.Status

	// Validar de acordo com o status
	if oportunidade.Status == models.StatusOportunidadeDraft {
		// Para drafts, validação mínima
		if err := oportunidade.Validate(); err != nil {
			return err
		}
	} else {
		// Para status publicado, validação completa
		if err := oportunidade.ValidateForPublish(); err != nil {
			return err
		}
	}

	oportunidade.UpdateStatusBasedOnExpiration()

	return s.repo.Update(ctx, oportunidade)
}

func (s *OportunidadeMEIService) Publish(ctx context.Context, id int) error {
	oportunidade, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if oportunidade == nil {
		return errors.New("oportunidade não encontrada")
	}

	// Validar completamente antes de publicar
	if err := oportunidade.ValidateForPublish(); err != nil {
		return errors.New("não é possível publicar: " + err.Error())
	}

	oportunidade.Status = models.StatusOportunidadeActive
	oportunidade.UpdateStatusBasedOnExpiration()

	return s.repo.Update(ctx, oportunidade)
}

func (s *OportunidadeMEIService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *OportunidadeMEIService) List(ctx context.Context, filters map[string]interface{}, titulo string, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	// Extract status filter if present, we'll handle it after updating expiration
	var requestedStatus models.StatusOportunidadeMEI
	var hasStatusFilter bool
	if statusFilter, ok := filters["status"]; ok {
		requestedStatus = statusFilter.(models.StatusOportunidadeMEI)
		hasStatusFilter = true
		// Remove status from database filters - we'll filter after updating status
		delete(filters, "status")
	}

	// Get all opportunities matching other filters (without status filter)
	offset := (page - 1) * pageSize
	oportunidades, _, err := s.repo.List(ctx, filters, titulo, pageSize*10, 0) // Get more to filter
	if err != nil {
		return nil, 0, err
	}

	// Update status based on expiration for each opportunity
	for _, oportunidade := range oportunidades {
		oportunidade.UpdateStatusBasedOnExpiration()
	}

	// Filter by status if requested (after updating statuses)
	if hasStatusFilter {
		filtered := make([]*models.OportunidadeMEI, 0)
		for _, oportunidade := range oportunidades {
			if oportunidade.Status == requestedStatus {
				filtered = append(filtered, oportunidade)
			}
		}
		oportunidades = filtered
	}

	// Apply pagination to filtered results
	total := len(oportunidades)
	start := offset
	end := offset + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	if start < end {
		oportunidades = oportunidades[start:end]
	} else {
		oportunidades = []*models.OportunidadeMEI{}
	}

	return oportunidades, total, nil
}

func (s *OportunidadeMEIService) ListByStatus(ctx context.Context, status models.StatusOportunidadeMEI, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	offset := (page - 1) * pageSize

	// Atualizar oportunidades expiradas antes de listar
	if err := s.repo.UpdateExpiredOpportunities(ctx); err != nil {
		// Log error mas não falha a operação
		logger.Error("failed to update expired opportunities", "error", err)
	}

	return s.repo.ListByStatus(ctx, status, pageSize, offset)
}

func (s *OportunidadeMEIService) ListDrafts(ctx context.Context, orgaoID, titulo string, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	filters := map[string]interface{}{
		"status": models.StatusOportunidadeDraft,
	}

	if orgaoID != "" {
		filters["orgao_id"] = orgaoID
	}

	return s.List(ctx, filters, titulo, page, pageSize)
}

func (s *OportunidadeMEIService) ListActive(ctx context.Context, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	return s.ListByStatus(ctx, models.StatusOportunidadeActive, page, pageSize)
}

func (s *OportunidadeMEIService) ListByOrgao(ctx context.Context, orgaoID string, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByOrgao(ctx, orgaoID, pageSize, offset)
}

func (s *OportunidadeMEIService) UpdateExpiredOpportunities(ctx context.Context) error {
	return s.repo.UpdateExpiredOpportunities(ctx)
}
