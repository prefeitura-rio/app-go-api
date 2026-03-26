package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type TipoConquistaService struct {
	repo TipoConquistaRepositoryInterface
}

func NewTipoConquistaService(repo *repository.TipoConquistaRepository) *TipoConquistaService {
	return &TipoConquistaService{repo: repo}
}

func NewTipoConquistaServiceWithInterface(repo TipoConquistaRepositoryInterface) *TipoConquistaService {
	return &TipoConquistaService{repo: repo}
}

func (s *TipoConquistaService) Create(ctx context.Context, entity *empregabilidade.TipoConquista) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *TipoConquistaService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoConquista, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TipoConquistaService) Update(ctx context.Context, entity *empregabilidade.TipoConquista) error {
	return s.repo.Update(ctx, entity)
}

func (s *TipoConquistaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *TipoConquistaService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.TipoConquista, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}
