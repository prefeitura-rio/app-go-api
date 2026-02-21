package empregabilidade

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type EmpresaService struct {
	repo *repository.EmpresaRepository
}

func NewEmpresaService(repo *repository.EmpresaRepository) *EmpresaService {
	return &EmpresaService{repo: repo}
}

func (s *EmpresaService) Create(ctx context.Context, entity *empregabilidade.Empresa) (string, error) {
	return s.repo.Create(ctx, entity)
}

func (s *EmpresaService) GetByID(ctx context.Context, cnpj string) (*empregabilidade.Empresa, error) {
	return s.repo.GetByID(ctx, cnpj)
}

func (s *EmpresaService) Update(ctx context.Context, entity *empregabilidade.Empresa) error {
	return s.repo.Update(ctx, entity)
}

func (s *EmpresaService) Delete(ctx context.Context, cnpj string) error {
	return s.repo.Delete(ctx, cnpj)
}

func (s *EmpresaService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.Empresa, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}

func (s *EmpresaService) Upsert(ctx context.Context, entity *empregabilidade.Empresa) error {
	return s.repo.Upsert(ctx, entity)
}
