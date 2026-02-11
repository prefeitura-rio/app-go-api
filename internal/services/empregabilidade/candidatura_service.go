package empregabilidade

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type CandidaturaRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Candidatura, error)
	Update(ctx context.Context, entity *empregabilidade.Candidatura) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Candidatura, int, error)
	ListByCPF(ctx context.Context, cpf string, limit, offset int) ([]*empregabilidade.Candidatura, int, error)
	ListByVaga(ctx context.Context, vagaID uuid.UUID, status string, limit, offset int) ([]*empregabilidade.Candidatura, int, error)
	CheckExistingCandidatura(ctx context.Context, cpf string, vagaID uuid.UUID) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status empregabilidade.StatusCandidatura) error
	UpdateEtapa(ctx context.Context, id uuid.UUID, etapaID uuid.UUID) error
}

type VagaRepositoryInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error)
}

type CurriculoServiceInterface interface {
	GetCurriculoCompleto(ctx context.Context, cpf string) (*empregabilidade.CurriculoCompleto, error)
}

type CandidaturaService struct {
	repo            CandidaturaRepositoryInterface
	vagaRepo        VagaRepositoryInterface
	curriculoService CurriculoServiceInterface
}

func NewCandidaturaService(
	repo CandidaturaRepositoryInterface,
	vagaRepo VagaRepositoryInterface,
	curriculoService CurriculoServiceInterface,
) *CandidaturaService {
	return &CandidaturaService{
		repo:            repo,
		vagaRepo:        vagaRepo,
		curriculoService: curriculoService,
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

	// Recalcula status baseado na data de expiração
	vaga.UpdateStatusBasedOnExpiration()

	if vaga.Status != empregabilidade.StatusVagaPublicadoAtivo {
		if vaga.Status == empregabilidade.StatusVagaPublicadoExpirado {
			return uuid.Nil, errors.New("vaga expirada, não aceita mais candidaturas")
		}
		return uuid.Nil, errors.New("vaga não está ativa para candidaturas")
	}

	exists, err := s.repo.CheckExistingCandidatura(ctx, entity.CPF, entity.IDVaga)
	if err != nil {
		return uuid.Nil, err
	}
	if exists {
		return uuid.Nil, errors.New("candidatura já existe para esta vaga")
	}

	curriculo, err := s.curriculoService.GetCurriculoCompleto(ctx, entity.CPF)
	if err == nil {
		entity.CurriculoSnapshot = curriculo
	}

	entity.Status = empregabilidade.StatusCandidaturaEnviada
	return s.repo.Create(ctx, entity)
}

func (s *CandidaturaService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Candidatura, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CandidaturaService) Update(ctx context.Context, entity *empregabilidade.Candidatura) error {
	// Fetch existing candidatura to preserve controlled fields
	existing, err := s.repo.GetByID(ctx, entity.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("candidatura não encontrada")
	}

	entity.Status = existing.Status
	entity.CPF = existing.CPF
	entity.IDVaga = existing.IDVaga
	entity.IDEtapaAtual = existing.IDEtapaAtual
	entity.CurriculoSnapshot = existing.CurriculoSnapshot

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
