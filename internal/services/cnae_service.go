package services

import (
	"context"
	"errors"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type CNAEService struct {
	repo *repository.CNAERepository
}

func NewCNAEService(repo *repository.CNAERepository) *CNAEService {
	return &CNAEService{
		repo: repo,
	}
}

func (s *CNAEService) GetByCodigo(ctx context.Context, codigo string) (*models.CNAE, error) {
	return s.repo.GetByCodigo(ctx, codigo)
}

func (s *CNAEService) List(ctx context.Context, ocupacao string, page, pageSize int) ([]*models.CNAE, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, ocupacao, pageSize, offset)
}

func (s *CNAEService) ListByOcupacao(ctx context.Context, ocupacao string) ([]*models.CNAE, error) {
	return s.repo.ListByOcupacao(ctx, ocupacao)
}

func (s *CNAEService) Create(ctx context.Context, cnae *models.CNAE) error {
	// Validar se CNAE já existe
	existing, err := s.repo.GetByCodigo(ctx, cnae.Codigo)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("CNAE já existe")
	}

	return s.repo.Create(ctx, cnae)
}

func (s *CNAEService) Update(ctx context.Context, cnae *models.CNAE) error {
	return s.repo.Update(ctx, cnae)
}

func (s *CNAEService) Delete(ctx context.Context, codigo string) error {
	return s.repo.Delete(ctx, codigo)
}
