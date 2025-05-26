package services

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type CategoriaService struct {
	repo *repository.CategoriaRepository
}

func NewCategoriaService(repo *repository.CategoriaRepository) *CategoriaService {
	return &CategoriaService{
		repo: repo,
	}
}

func (s *CategoriaService) Create(ctx context.Context, categoria *models.Categoria) (int, error) {
	return s.repo.Create(ctx, categoria)
}

func (s *CategoriaService) GetByID(ctx context.Context, id int) (*models.Categoria, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CategoriaService) Update(ctx context.Context, categoria *models.Categoria) error {
	return s.repo.Update(ctx, categoria)
}

func (s *CategoriaService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *CategoriaService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Categoria, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
