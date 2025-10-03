package services

import (
	"context"

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
	// Validação já é feita pela constraint UNIQUE(codigo, servico) no banco
	// Não precisa validar manualmente
	return s.repo.Create(ctx, cnae)
}

func (s *CNAEService) Update(ctx context.Context, cnae *models.CNAE) error {
	return s.repo.Update(ctx, cnae)
}

func (s *CNAEService) Delete(ctx context.Context, codigo string) error {
	return s.repo.Delete(ctx, codigo)
}
