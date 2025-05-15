package services

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type CursoService struct {
	repo *repository.CursoRepository
}

func NewCursoService(repo *repository.CursoRepository) *CursoService {
	return &CursoService{
		repo: repo,
	}
}

func (s *CursoService) Create(ctx context.Context, curso *models.Curso) (int, error) {
	return s.repo.Create(ctx, curso)
}

func (s *CursoService) GetByID(ctx context.Context, id int) (*models.Curso, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CursoService) Update(ctx context.Context, curso *models.Curso) error {
	return s.repo.Update(ctx, curso)
}

func (s *CursoService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *CursoService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Curso, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
} 