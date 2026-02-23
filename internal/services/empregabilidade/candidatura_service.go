package empregabilidade

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type CandidaturaRepositoryInterface interface {
	Create(ctx context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Candidatura, error)
	Update(ctx context.Context, entity *empregabilidade.Candidatura) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter empregabilidade.CandidaturaFilter, limit, offset int) ([]*empregabilidade.Candidatura, int, error)
	CheckExistingCandidatura(ctx context.Context, cpf string, vagaID uuid.UUID) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status empregabilidade.StatusCandidatura) error
	UpdateEtapa(ctx context.Context, id uuid.UUID, etapaID uuid.UUID) error
	BulkUpdateStatus(ctx context.Context, vagaID uuid.UUID, cpfs []string, status empregabilidade.StatusCandidatura) (repository.BulkUpdateResult, error)
}

type VagaRepositoryInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error)
}

type CurriculoServiceInterface interface {
	GetCurriculoCompleto(ctx context.Context, cpf string) (*empregabilidade.CurriculoCompleto, error)
}

var validStatusTransitions = map[empregabilidade.StatusCandidatura][]empregabilidade.StatusCandidatura{
	empregabilidade.StatusCandidaturaEnviada: {
		empregabilidade.StatusCandidaturaAprovada,
		empregabilidade.StatusCandidaturaReprovada,
		empregabilidade.StatusCandidaturaVagaCongelada,
		empregabilidade.StatusCandidaturaDescontinuada,
	},
	empregabilidade.StatusCandidaturaAprovada: {
		empregabilidade.StatusCandidaturaEnviada,
		empregabilidade.StatusCandidaturaReprovada,
		empregabilidade.StatusCandidaturaVagaCongelada,
		empregabilidade.StatusCandidaturaDescontinuada,
	},
	empregabilidade.StatusCandidaturaReprovada: {
		empregabilidade.StatusCandidaturaEnviada,
		empregabilidade.StatusCandidaturaAprovada,
		empregabilidade.StatusCandidaturaVagaCongelada,
		empregabilidade.StatusCandidaturaDescontinuada,
	},
	empregabilidade.StatusCandidaturaVagaCongelada: {
		empregabilidade.StatusCandidaturaEnviada,
		empregabilidade.StatusCandidaturaAprovada,
		empregabilidade.StatusCandidaturaReprovada,
		empregabilidade.StatusCandidaturaDescontinuada,
	},
	empregabilidade.StatusCandidaturaDescontinuada: {},
}

func canTransitionCandidatura(from, to empregabilidade.StatusCandidatura) bool {
	allowed, ok := validStatusTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

type CandidaturaService struct {
	repo             CandidaturaRepositoryInterface
	vagaRepo         VagaRepositoryInterface
	curriculoService CurriculoServiceInterface
}

func NewCandidaturaService(
	repo CandidaturaRepositoryInterface,
	vagaRepo VagaRepositoryInterface,
	curriculoService CurriculoServiceInterface,
) *CandidaturaService {
	return &CandidaturaService{
		repo:             repo,
		vagaRepo:         vagaRepo,
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

func (s *CandidaturaService) List(ctx context.Context, filter empregabilidade.CandidaturaFilter, page, pageSize int) ([]*empregabilidade.Candidatura, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filter, pageSize, offset)
}

func (s *CandidaturaService) UpdateStatus(ctx context.Context, id uuid.UUID, status empregabilidade.StatusCandidatura) error {
	if !status.IsValid() {
		return errors.New("status inválido")
	}
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("candidatura não encontrada")
	}
	if !canTransitionCandidatura(existing.Status, status) {
		return fmt.Errorf("transição de status inválida: %s → %s", existing.Status, status)
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *CandidaturaService) UpdateEtapa(ctx context.Context, id uuid.UUID, etapaID uuid.UUID) error {
	candidatura, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if candidatura == nil {
		return errors.New("candidatura não encontrada")
	}

	vaga, err := s.vagaRepo.GetByID(ctx, candidatura.IDVaga)
	if err != nil {
		return err
	}
	if vaga == nil {
		return errors.New("vaga não encontrada")
	}

	etapaValida := false
	for _, etapa := range vaga.Etapas {
		if etapa.ID == etapaID {
			etapaValida = true
			break
		}
	}
	if !etapaValida {
		return errors.New("etapa não pertence à vaga desta candidatura")
	}

	return s.repo.UpdateEtapa(ctx, id, etapaID)
}

func (s *CandidaturaService) Approve(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("candidatura não encontrada")
	}
	if !canTransitionCandidatura(existing.Status, empregabilidade.StatusCandidaturaAprovada) {
		return fmt.Errorf("candidatura não pode ser aprovada no status atual: %s", existing.Status)
	}
	return s.repo.UpdateStatus(ctx, id, empregabilidade.StatusCandidaturaAprovada)
}

type BulkUpdateStatusRequest struct {
	CPFs   []string                        `json:"cpfs"`
	VagaID uuid.UUID                       `json:"vaga_id"`
	Status empregabilidade.StatusCandidatura `json:"status"`
}

type BulkUpdateStatusResult struct {
	Updated    int      `json:"updated"`
	FailedCPFs []string `json:"failed_cpfs"`
}

func (s *CandidaturaService) BulkUpdateStatus(ctx context.Context, req BulkUpdateStatusRequest) (BulkUpdateStatusResult, error) {
	if !req.Status.IsValid() {
		return BulkUpdateStatusResult{}, errors.New("status inválido")
	}
	if len(req.CPFs) == 0 {
		return BulkUpdateStatusResult{}, errors.New("lista de CPFs não pode ser vazia")
	}

	repoResult, err := s.repo.BulkUpdateStatus(ctx, req.VagaID, req.CPFs, req.Status)
	if err != nil {
		return BulkUpdateStatusResult{}, err
	}

	return BulkUpdateStatusResult{
		Updated:    repoResult.Updated,
		FailedCPFs: repoResult.FailedCPFs,
	}, nil
}

func (s *CandidaturaService) Reject(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("candidatura não encontrada")
	}
	if !canTransitionCandidatura(existing.Status, empregabilidade.StatusCandidaturaReprovada) {
		return fmt.Errorf("candidatura não pode ser reprovada no status atual: %s", existing.Status)
	}
	return s.repo.UpdateStatus(ctx, id, empregabilidade.StatusCandidaturaReprovada)
}
