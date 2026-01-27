package empregabilidade

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type CandidaturaService struct {
	repo     *repository.CandidaturaRepository
	vagaRepo *repository.VagaRepository
}

func NewCandidaturaService(
	repo *repository.CandidaturaRepository,
	vagaRepo *repository.VagaRepository,
) *CandidaturaService {
	return &CandidaturaService{
		repo:     repo,
		vagaRepo: vagaRepo,
	}
}

func (s *CandidaturaService) Create(ctx context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error) {
	vaga, err := s.vagaRepo.GetByID(ctx, entity.IDVaga)
	if err != nil {
		return uuid.Nil, err
	}
	if vaga == nil {
		return uuid.Nil, errors.New("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaPublicadoAtivo {
		return uuid.Nil, errors.New("vaga não está ativa para candidaturas")
	}

	exists, err := s.repo.CheckExistingCandidatura(ctx, entity.CPF, entity.IDVaga)
	if err != nil {
		return uuid.Nil, err
	}
	if exists {
		return uuid.Nil, errors.New("candidatura já existe para esta vaga")
	}

	entity.Status = empregabilidade.StatusCandidaturaEnviada
	return s.repo.Create(ctx, entity)
}

func (s *CandidaturaService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Candidatura, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CandidaturaService) Update(ctx context.Context, entity *empregabilidade.Candidatura) error {
	return s.repo.Update(ctx, entity)
}

func (s *CandidaturaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *CandidaturaService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.Candidatura, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}

func (s *CandidaturaService) ListByCPF(ctx context.Context, cpf string, page, pageSize int) ([]*empregabilidade.Candidatura, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByCPF(ctx, cpf, pageSize, offset)
}

func (s *CandidaturaService) ListByVaga(ctx context.Context, vagaID uuid.UUID, status string, page, pageSize int) ([]*empregabilidade.Candidatura, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByVaga(ctx, vagaID, status, pageSize, offset)
}

func (s *CandidaturaService) UpdateStatus(ctx context.Context, id uuid.UUID, status empregabilidade.StatusCandidatura) error {
	if !status.IsValid() {
		return errors.New("status inválido")
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *CandidaturaService) UpdateEtapa(ctx context.Context, id uuid.UUID, etapaID uuid.UUID) error {
	return s.repo.UpdateEtapa(ctx, id, etapaID)
}

func (s *CandidaturaService) Approve(ctx context.Context, id uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, id, empregabilidade.StatusCandidaturaAprovada)
}

func (s *CandidaturaService) Reject(ctx context.Context, id uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, id, empregabilidade.StatusCandidaturaReprovada)
}
