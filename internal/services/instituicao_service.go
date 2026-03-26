package services

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

type InstituicaoService struct {
	repo InstituicaoRepositoryInterface
}

func NewInstituicaoService(repo *repository.InstituicaoRepository) *InstituicaoService {
	return &InstituicaoService{
		repo: repo,
	}
}

func NewInstituicaoServiceWithInterface(repo InstituicaoRepositoryInterface) *InstituicaoService {
	return &InstituicaoService{repo: repo}
}

func (s *InstituicaoService) Create(ctx context.Context, instituicao *models.InstituicaoEnsino) (int, error) {
	return s.repo.Create(ctx, instituicao)
}

func (s *InstituicaoService) GetByID(ctx context.Context, id int) (*models.InstituicaoEnsino, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *InstituicaoService) Update(ctx context.Context, instituicao *models.InstituicaoEnsino) error {
	return s.repo.Update(ctx, instituicao)
}

func (s *InstituicaoService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *InstituicaoService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*models.InstituicaoEnsino, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
