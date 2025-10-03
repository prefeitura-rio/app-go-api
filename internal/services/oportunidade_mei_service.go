package services

import (
	"context"
	"errors"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type OportunidadeMEIService struct {
	repo     *repository.OportunidadeMEIRepository
	cnaeRepo *repository.CNAERepository
	orgaoRepo *repository.OrgaoRepository
}

func NewOportunidadeMEIService(
	repo *repository.OportunidadeMEIRepository,
	cnaeRepo *repository.CNAERepository,
	orgaoRepo *repository.OrgaoRepository,
) *OportunidadeMEIService {
	return &OportunidadeMEIService{
		repo:     repo,
		cnaeRepo: cnaeRepo,
		orgaoRepo: orgaoRepo,
	}
}

func (s *OportunidadeMEIService) Create(ctx context.Context, oportunidade *models.OportunidadeMEI, isDraft bool) (int, error) {
	if err := oportunidade.Validate(); err != nil {
		return 0, err
	}

	// Validar que o CNAE existe
	cnae, err := s.cnaeRepo.GetByCodigo(ctx, oportunidade.CNAECodigo)
	if err != nil {
		return 0, err
	}
	if cnae == nil {
		return 0, errors.New("CNAE não encontrado")
	}

	// Validar que o órgão existe
	orgao, err := s.orgaoRepo.GetByID(ctx, oportunidade.OrgaoID)
	if err != nil {
		return 0, err
	}
	if orgao == nil {
		return 0, errors.New("órgão não encontrado")
	}

	if isDraft {
		oportunidade.Status = models.StatusOportunidadeDraft
	} else {
		oportunidade.Status = models.StatusOportunidadeActive
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
	if err := oportunidade.Validate(); err != nil {
		return err
	}

	// Se estava em draft e está sendo atualizado, publicar
	existing, err := s.repo.GetByID(ctx, oportunidade.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("oportunidade não encontrada")
	}

	if existing.Status == models.StatusOportunidadeDraft {
		oportunidade.Status = models.StatusOportunidadeActive
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

	oportunidade.Status = models.StatusOportunidadeActive
	oportunidade.UpdateStatusBasedOnExpiration()

	return s.repo.Update(ctx, oportunidade)
}

func (s *OportunidadeMEIService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *OportunidadeMEIService) List(ctx context.Context, filters map[string]interface{}, titulo string, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filters, titulo, pageSize, offset)
}

func (s *OportunidadeMEIService) ListByStatus(ctx context.Context, status models.StatusOportunidadeMEI, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	offset := (page - 1) * pageSize

	// Atualizar oportunidades expiradas antes de listar
	if err := s.repo.UpdateExpiredOpportunities(ctx); err != nil {
		// Log error mas não falha a operação
	}

	return s.repo.ListByStatus(ctx, status, pageSize, offset)
}

func (s *OportunidadeMEIService) ListDrafts(ctx context.Context, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	return s.ListByStatus(ctx, models.StatusOportunidadeDraft, page, pageSize)
}

func (s *OportunidadeMEIService) ListActive(ctx context.Context, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	return s.ListByStatus(ctx, models.StatusOportunidadeActive, page, pageSize)
}

func (s *OportunidadeMEIService) ListByOrgao(ctx context.Context, orgaoID int, page, pageSize int) ([]*models.OportunidadeMEI, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByOrgao(ctx, orgaoID, pageSize, offset)
}

func (s *OportunidadeMEIService) UpdateExpiredOpportunities(ctx context.Context) error {
	return s.repo.UpdateExpiredOpportunities(ctx)
}
