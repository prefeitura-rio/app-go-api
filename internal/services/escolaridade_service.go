package services

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type EscolaridadeService struct {
	repo EscolaridadeRepositoryInterface
}

func NewEscolaridadeService(repo EscolaridadeRepositoryInterface) *EscolaridadeService {
	return &EscolaridadeService{
		repo: repo,
	}
}

func (s *EscolaridadeService) Create(ctx context.Context, escolaridade *models.Escolaridade) (int, error) {
	return s.repo.Create(ctx, escolaridade)
}

func (s *EscolaridadeService) GetByID(ctx context.Context, id int) (*models.Escolaridade, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EscolaridadeService) Update(ctx context.Context, escolaridade *models.Escolaridade) error {
	return s.repo.Update(ctx, escolaridade)
}

func (s *EscolaridadeService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *EscolaridadeService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Escolaridade, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
