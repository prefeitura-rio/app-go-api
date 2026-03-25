package services

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type EmpresaService struct {
	repo EmpresaRepositoryInterface
}

func NewEmpresaService(repo *repository.EmpresaRepository) *EmpresaService {
	return &EmpresaService{
		repo: repo,
	}
}

func NewEmpresaServiceWithInterface(repo EmpresaRepositoryInterface) *EmpresaService {
	return &EmpresaService{repo: repo}
}

func (s *EmpresaService) Create(ctx context.Context, empresa *models.Empresa) (int, error) {
	return s.repo.Create(ctx, empresa)
}

func (s *EmpresaService) GetByID(ctx context.Context, id int) (*models.Empresa, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EmpresaService) Update(ctx context.Context, empresa *models.Empresa) error {
	return s.repo.Update(ctx, empresa)
}

func (s *EmpresaService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *EmpresaService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.Empresa, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
