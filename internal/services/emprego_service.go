package services

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type EmpregoService struct {
	repo *repository.EmpregoRepository
}

func NewEmpregoService(repo *repository.EmpregoRepository) *EmpregoService {
	return &EmpregoService{
		repo: repo,
	}
}

func (s *EmpregoService) Create(ctx context.Context, emprego *models.Emprego) (int, error) {
	return s.repo.Create(ctx, emprego)
}

func (s *EmpregoService) GetByID(ctx context.Context, id int) (*models.Emprego, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EmpregoService) Update(ctx context.Context, emprego *models.Emprego) error {
	return s.repo.Update(ctx, emprego)
}

func (s *EmpregoService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *EmpregoService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Emprego, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
} 