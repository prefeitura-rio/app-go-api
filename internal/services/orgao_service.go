package services

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type OrgaoService struct {
	repo *repository.OrgaoRepository
}

func NewOrgaoService(repo *repository.OrgaoRepository) *OrgaoService {
	return &OrgaoService{
		repo: repo,
	}
}

func (s *OrgaoService) Create(ctx context.Context, orgao *models.Orgao) (int, error) {
	return s.repo.Create(ctx, orgao)
}

func (s *OrgaoService) GetByID(ctx context.Context, id int) (*models.Orgao, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *OrgaoService) Update(ctx context.Context, orgao *models.Orgao) error {
	return s.repo.Update(ctx, orgao)
}

func (s *OrgaoService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *OrgaoService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Orgao, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
