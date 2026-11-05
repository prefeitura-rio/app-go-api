package empregabilidade

import (
	"context"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type ComportamentoAtitudesService struct {
	repo ComportamentoAtitudesRepositoryInterface
}

func NewComportamentoAtitudesService(repo ComportamentoAtitudesRepositoryInterface) *ComportamentoAtitudesService {
	return &ComportamentoAtitudesService{repo: repo}
}

// CreateComportamentoAtitudes cria um novo comportamento/atitude e retorna o ID gerado (int64)
func (s *ComportamentoAtitudesService) CreateComportamentoAtitudes(ctx context.Context, entity *empregabilidade.ComportamentoAtitudes) (int64, error) {
	return s.repo.CreateComportamentoAtitudes(ctx, entity)
}

// GetComportamentoAtitudesByID busca um comportamento/atitude pelo seu ID
func (s *ComportamentoAtitudesService) GetComportamentoAtitudesByID(ctx context.Context, id int64) (*empregabilidade.ComportamentoAtitudes, error) {
	return s.repo.GetComportamentoAtitudesByID(ctx, id)
}

// UpdateComportamentoAtitudes atualiza um comportamento/atitude existente
func (s *ComportamentoAtitudesService) UpdateComportamentoAtitudes(ctx context.Context, entity *empregabilidade.ComportamentoAtitudes) error {
	return s.repo.UpdateComportamentoAtitudes(ctx, entity)
}

// DeleteComportamentoAtitudes remove um comportamento/atitude pelo seu ID
func (s *ComportamentoAtitudesService) DeleteComportamentoAtitudes(ctx context.Context, id int64) error {
	return s.repo.DeleteComportamentoAtitudes(ctx, id)
}

// ListComportamentoAtitudes lista comportamentos/atitudes com paginação e filtro
func (s *ComportamentoAtitudesService) ListComportamentoAtitudes(ctx context.Context, filter empregabilidade.ComportamentoAtitudesFilter, page, pageSize int) ([]*empregabilidade.ComportamentoAtitudes, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return s.repo.ListComportamentoAtitudes(ctx, filter, pageSize, offset)
}
