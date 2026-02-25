package empregabilidade

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
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
	BulkUpdateStatus(ctx context.Context, vagaID uuid.UUID, cpfs []string, status empregabilidade.StatusCandidatura) (empRepository.BulkUpdateResult, error)
	BulkGetByCPFs(ctx context.Context, vagaID uuid.UUID, cpfs []string) ([]*empregabilidade.Candidatura, error)
	BulkUpdateEtapa(ctx context.Context, ids []uuid.UUID, etapaID uuid.UUID) error
	BulkSaveAndUpdateStatusByVagaID(ctx context.Context, vagaID uuid.UUID, newStatus empregabilidade.StatusCandidatura) error
	BulkRestoreStatusByVagaID(ctx context.Context, vagaID uuid.UUID) error
}

type VagaRepositoryInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error)
}

type CurriculoServiceInterface interface {
	GetCurriculoCompleto(ctx context.Context, cpf string) (*empregabilidade.CurriculoCompleto, error)
}

type CitizenSnapshotRepoForCandidaturaInterface interface {
	GetByCPF(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
	GetByCPFs(ctx context.Context, cpfs []string) (map[string]*models.CitizenSnapshot, error)
}

type CitizenDataFetcherForCandidaturaInterface interface {
	SyncCitizenOnDemand(ctx context.Context, cpf string) (*models.CitizenSnapshot, error)
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
	repo                CandidaturaRepositoryInterface
	vagaRepo            VagaRepositoryInterface
	curriculoService    CurriculoServiceInterface
	citizenSnapshotRepo CitizenSnapshotRepoForCandidaturaInterface // pode ser nil
	citizenDataFetcher  CitizenDataFetcherForCandidaturaInterface  // pode ser nil
}

func NewCandidaturaService(
	repo CandidaturaRepositoryInterface,
	vagaRepo VagaRepositoryInterface,
	curriculoService CurriculoServiceInterface,
	citizenSnapshotRepo CitizenSnapshotRepoForCandidaturaInterface,
	citizenDataFetcher CitizenDataFetcherForCandidaturaInterface,
) *CandidaturaService {
	return &CandidaturaService{
		repo:                repo,
		vagaRepo:            vagaRepo,
		curriculoService:    curriculoService,
		citizenSnapshotRepo: citizenSnapshotRepo,
		citizenDataFetcher:  citizenDataFetcher,
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

type BulkUpdateStatusResult struct {
	Updated    int
	FailedCPFs []string
}

func (s *CandidaturaService) BulkUpdateStatus(ctx context.Context, vagaID uuid.UUID, cpfs []string, status empregabilidade.StatusCandidatura) (BulkUpdateStatusResult, error) {
	if !status.IsValid() {
		return BulkUpdateStatusResult{}, errors.New("status inválido")
	}
	if len(cpfs) == 0 {
		return BulkUpdateStatusResult{}, errors.New("lista de CPFs não pode ser vazia")
	}

	repoResult, err := s.repo.BulkUpdateStatus(ctx, vagaID, cpfs, status)
	if err != nil {
		return BulkUpdateStatusResult{}, err
	}

	return BulkUpdateStatusResult{
		Updated:    repoResult.Updated,
		FailedCPFs: repoResult.FailedCPFs,
	}, nil
}

type BulkUpdateEtapaResult struct {
	Updated    int
	FailedCPFs []string
}

func (s *CandidaturaService) BulkUpdateEtapa(ctx context.Context, vagaID uuid.UUID, cpfs []string, etapaID uuid.UUID) (BulkUpdateEtapaResult, error) {
	if len(cpfs) == 0 {
		return BulkUpdateEtapaResult{}, errors.New("lista de CPFs não pode ser vazia")
	}

	vaga, err := s.vagaRepo.GetByID(ctx, vagaID)
	if err != nil {
		return BulkUpdateEtapaResult{}, err
	}
	if vaga == nil {
		return BulkUpdateEtapaResult{}, errors.New("vaga não encontrada")
	}

	etapaValida := false
	for _, etapa := range vaga.Etapas {
		if etapa.ID == etapaID {
			etapaValida = true
			break
		}
	}
	if !etapaValida {
		return BulkUpdateEtapaResult{}, errors.New("etapa não pertence à vaga")
	}

	candidaturas, err := s.repo.BulkGetByCPFs(ctx, vagaID, cpfs)
	if err != nil {
		return BulkUpdateEtapaResult{}, err
	}

	foundCPFs := make(map[string]*empregabilidade.Candidatura, len(candidaturas))
	for _, c := range candidaturas {
		foundCPFs[c.CPF] = c
	}

	var result BulkUpdateEtapaResult
	var updateIDs []uuid.UUID

	// Coletar CPFs não encontrados
	for _, cpf := range cpfs {
		if _, found := foundCPFs[cpf]; !found {
			result.FailedCPFs = append(result.FailedCPFs, cpf)
		}
	}

	// Verificar que todos os candidatos encontrados estão na mesma etapa atual
	if len(candidaturas) > 0 {
		primeiraEtapa := candidaturas[0].IDEtapaAtual
		for _, c := range candidaturas {
			mesmaEtapa := false
			if primeiraEtapa == nil && c.IDEtapaAtual == nil {
				mesmaEtapa = true
			} else if primeiraEtapa != nil && c.IDEtapaAtual != nil && *primeiraEtapa == *c.IDEtapaAtual {
				mesmaEtapa = true
			}
			if !mesmaEtapa {
				return BulkUpdateEtapaResult{}, errors.New("todos os candidatos precisam estar na mesma etapa para atualização em massa")
			}
			updateIDs = append(updateIDs, c.ID)
		}
	}

	if len(updateIDs) == 0 {
		return result, nil
	}

	if err := s.repo.BulkUpdateEtapa(ctx, updateIDs, etapaID); err != nil {
		return BulkUpdateEtapaResult{}, err
	}

	result.Updated = len(updateIDs)
	return result, nil
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

// EnrichWithPersonalInfo popula PersonalInfo de uma candidatura a partir do citizen_snapshot.
func (s *CandidaturaService) EnrichWithPersonalInfo(ctx context.Context, c *empregabilidade.Candidatura) {
	if s.citizenSnapshotRepo == nil || c == nil || c.CPF == "" {
		return
	}

	snapshot, err := s.citizenSnapshotRepo.GetByCPF(ctx, c.CPF)
	if err != nil {
		fmt.Printf("[CandidaturaService] Failed to get citizen snapshot for CPF %s: %v\n", c.CPF, err)
		return
	}

	if snapshot == nil && s.citizenDataFetcher != nil {
		snapshot, err = s.citizenDataFetcher.SyncCitizenOnDemand(ctx, c.CPF)
		if err != nil {
			fmt.Printf("[CandidaturaService] On-demand sync failed for CPF %s: %v\n", c.CPF, err)
			return
		}
	}

	if snapshot != nil {
		c.PersonalInfo = snapshot.ToPersonalInfo()
	}
}

// EnrichMultipleWithPersonalInfo popula PersonalInfo de múltiplas candidaturas em batch.
func (s *CandidaturaService) EnrichMultipleWithPersonalInfo(ctx context.Context, candidaturas []*empregabilidade.Candidatura) {
	if s.citizenSnapshotRepo == nil || len(candidaturas) == 0 {
		return
	}

	// Collect unique CPFs
	cpfSet := make(map[string]struct{})
	for _, c := range candidaturas {
		if c.CPF != "" {
			cpfSet[c.CPF] = struct{}{}
		}
	}

	cpfs := make([]string, 0, len(cpfSet))
	for cpf := range cpfSet {
		cpfs = append(cpfs, cpf)
	}

	if len(cpfs) == 0 {
		return
	}

	snapshotMap, err := s.citizenSnapshotRepo.GetByCPFs(ctx, cpfs)
	if err != nil {
		fmt.Printf("[CandidaturaService] Failed to get citizen snapshots: %v\n", err)
		return
	}

	// Sync CPFs missing from cache on-demand if fetcher is available
	if s.citizenDataFetcher != nil {
		for _, cpf := range cpfs {
			if _, ok := snapshotMap[cpf]; !ok {
				snapshot, err := s.citizenDataFetcher.SyncCitizenOnDemand(ctx, cpf)
				if err != nil {
					fmt.Printf("[CandidaturaService] On-demand sync failed for CPF %s: %v\n", cpf, err)
					continue
				}
				if snapshot != nil {
					snapshotMap[cpf] = snapshot
				}
			}
		}
	}

	for _, c := range candidaturas {
		if snapshot, ok := snapshotMap[c.CPF]; ok && snapshot != nil {
			c.PersonalInfo = snapshot.ToPersonalInfo()
		}
	}
}
