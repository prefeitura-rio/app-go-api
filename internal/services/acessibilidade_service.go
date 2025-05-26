package services

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type AcessibilidadeService struct {
	repo *repository.AcessibilidadeRepository
}

func NewAcessibilidadeService(repo *repository.AcessibilidadeRepository) *AcessibilidadeService {
	return &AcessibilidadeService{
		repo: repo,
	}
}

func (s *AcessibilidadeService) Create(ctx context.Context, acessibilidade *models.Acessibilidade) (int, error) {
	return s.repo.Create(ctx, acessibilidade)
}

func (s *AcessibilidadeService) GetByID(ctx context.Context, id int) (*models.Acessibilidade, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AcessibilidadeService) Update(ctx context.Context, acessibilidade *models.Acessibilidade) error {
	return s.repo.Update(ctx, acessibilidade)
}

func (s *AcessibilidadeService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *AcessibilidadeService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Acessibilidade, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
