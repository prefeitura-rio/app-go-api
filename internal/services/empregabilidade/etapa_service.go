package empregabilidade

import (
	"context"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// ErrEtapaNaoEncontrada indica que a etapa não existe na vaga informada
// (inexistente ou pertencente a outra vaga).
var ErrEtapaNaoEncontrada = repository.ErrEtapaNaoEncontrada

type EtapaService struct {
	repo EtapaRepositoryInterface
}

func NewEtapaService(repo *repository.EtapaRepository) *EtapaService {
	return &EtapaService{repo: repo}
}

func NewEtapaServiceWithInterface(repo EtapaRepositoryInterface) *EtapaService {
	return &EtapaService{repo: repo}
}

func (s *EtapaService) Create(ctx context.Context, entity *empregabilidade.Etapa) (uuid.UUID, error) {
	return s.repo.Create(ctx, entity)
}

func (s *EtapaService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Etapa, error) {
	return s.repo.GetByID(ctx, id)
}

// Update altera a etapa entity.ID somente se ela pertence à vaga entity.IDVaga.
// A autorização das rotas é feita sobre a vaga do path; etapa de outra vaga (ou
// inexistente) devolve ErrEtapaNaoEncontrada sem gravar nada.
func (s *EtapaService) Update(ctx context.Context, entity *empregabilidade.Etapa) error {
	if err := s.ensureEtapaDaVaga(ctx, entity.IDVaga, entity.ID); err != nil {
		return err
	}
	return s.repo.Update(ctx, entity)
}

// Delete exclui a etapa id somente se ela pertence à vaga vagaID; etapa de
// outra vaga (ou inexistente) devolve ErrEtapaNaoEncontrada sem excluir nada.
func (s *EtapaService) Delete(ctx context.Context, vagaID, id uuid.UUID) error {
	if err := s.ensureEtapaDaVaga(ctx, vagaID, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, vagaID, id)
}

func (s *EtapaService) ensureEtapaDaVaga(ctx context.Context, vagaID, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil || existing.IDVaga != vagaID {
		return ErrEtapaNaoEncontrada
	}
	return nil
}

func (s *EtapaService) ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.Etapa, error) {
	return s.repo.ListByVaga(ctx, vagaID)
}

func (s *EtapaService) DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error {
	return s.repo.DeleteByVaga(ctx, vagaID)
}
