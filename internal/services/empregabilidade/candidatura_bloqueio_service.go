package empregabilidade

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// ErrBloqueioVagaNaoEncontrada indica que o id_vaga do bloqueio não corresponde
// a nenhuma vaga cadastrada.
var ErrBloqueioVagaNaoEncontrada = errors.New("vaga não encontrada")

// VagaExistenciaRepositoryInterface confere se uma vaga existe sem carregá-la.
type VagaExistenciaRepositoryInterface interface {
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}

type CandidaturaBloqueioService struct {
	repo     CandidaturaBloqueioRepositoryInterface
	vagaRepo VagaExistenciaRepositoryInterface
}

func NewCandidaturaBloqueioService(repo *repository.CandidaturaBloqueioRepository, vagaRepo *repository.VagaRepository) *CandidaturaBloqueioService {
	return &CandidaturaBloqueioService{repo: repo, vagaRepo: vagaRepo}
}

func NewCandidaturaBloqueioServiceWithInterface(repo CandidaturaBloqueioRepositoryInterface, vagaRepo VagaExistenciaRepositoryInterface) *CandidaturaBloqueioService {
	return &CandidaturaBloqueioService{repo: repo, vagaRepo: vagaRepo}
}

// Create registra o bloqueio depois de conferir que a vaga existe. A vaga não
// precisa estar ativa: ela pode ter expirado durante a tentativa do candidato.
func (s *CandidaturaBloqueioService) Create(ctx context.Context, entity *empregabilidade.CandidaturaBloqueio) (uuid.UUID, error) {
	exists, err := s.vagaRepo.ExistsByID(ctx, entity.IDVaga)
	if err != nil {
		return uuid.Nil, err
	}
	if !exists {
		return uuid.Nil, ErrBloqueioVagaNaoEncontrada
	}
	return s.repo.Create(ctx, entity)
}

// List devolve os bloqueios paginados. orgaoParceiroIDs é o escopo de órgão do
// editor (nil não restringe; lista vazia não devolve nada).
func (s *CandidaturaBloqueioService) List(ctx context.Context, filter map[string]interface{}, orgaoParceiroIDs []string, page, pageSize int) ([]*empregabilidade.CandidaturaBloqueio, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, orgaoParceiroIDs, pageSize, offset)
}
