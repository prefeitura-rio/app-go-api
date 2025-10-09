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

	// Para publicação (não rascunho), validar que CNAE e órgão existem
	if !isDraft {
		// Validar que o CNAE existe
		if oportunidade.CNAEID > 0 {
			cnae, err := s.cnaeRepo.GetByID(ctx, oportunidade.CNAEID)
			if err != nil {
				return 0, err
			}
			if cnae == nil {
				return 0, errors.New("CNAE não encontrado")
			}
		}

		// Validar que o órgão existe
		if oportunidade.OrgaoID > 0 {
			orgao, err := s.orgaoRepo.GetByID(ctx, oportunidade.OrgaoID)
			if err != nil {
				return 0, err
			}
			if orgao == nil {
				return 0, errors.New("órgão não encontrado")
			}
		}
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

	// Validar que o CNAE existe (se mudou)
	if oportunidade.CNAEID != existing.CNAEID {
		cnae, err := s.cnaeRepo.GetByID(ctx, oportunidade.CNAEID)
		if err != nil {
			return err
		}
		if cnae == nil {
			return errors.New("CNAE não encontrado")
		}
	}

	// Validar que o órgão existe (se mudou)
	if oportunidade.OrgaoID != existing.OrgaoID {
		orgao, err := s.orgaoRepo.GetByID(ctx, oportunidade.OrgaoID)
		if err != nil {
			return err
		}
		if orgao == nil {
			return errors.New("órgão não encontrado")
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
