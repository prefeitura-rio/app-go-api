package empregabilidade

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type VagaService struct {
	repo                      *repository.VagaRepository
	empresaRepo               *repository.EmpresaRepository
	etapaRepo                 *repository.EtapaRepository
	informacaoComplementarRepo *repository.InformacaoComplementarRepository
}

func NewVagaService(
	repo *repository.VagaRepository,
	empresaRepo *repository.EmpresaRepository,
	etapaRepo *repository.EtapaRepository,
	informacaoComplementarRepo *repository.InformacaoComplementarRepository,
) *VagaService {
	return &VagaService{
		repo:                      repo,
		empresaRepo:               empresaRepo,
		etapaRepo:                 etapaRepo,
		informacaoComplementarRepo: informacaoComplementarRepo,
	}
}

func (s *VagaService) Create(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error) {
	empresa, err := s.empresaRepo.GetByID(ctx, entity.IDContratante)
	if err != nil {
		return uuid.Nil, err
	}
	if empresa == nil {
		return uuid.Nil, errors.New("empresa contratante não encontrada")
	}

	return s.repo.Create(ctx, entity)
}

func (s *VagaService) CreateDraft(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error) {
	entity.Status = empregabilidade.StatusVagaEmEdicao
	return s.repo.Create(ctx, entity)
}

func (s *VagaService) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error) {
	vaga, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if vaga != nil {
		vaga.UpdateStatusBasedOnExpiration()
	}
	return vaga, nil
}

func (s *VagaService) Update(ctx context.Context, entity *empregabilidade.Vaga) error {
	return s.repo.Update(ctx, entity)
}

func (s *VagaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *VagaService) List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.Vaga, int, error) {
	offset := (page - 1) * pageSize
	vagas, total, err := s.repo.List(ctx, filter, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	for _, vaga := range vagas {
		vaga.UpdateStatusBasedOnExpiration()
	}

	return vagas, total, nil
}

func (s *VagaService) UpdateTiposPCD(ctx context.Context, vagaID uuid.UUID, tiposPCDIDs []uuid.UUID) error {
	return s.repo.UpdateTiposPCD(ctx, vagaID, tiposPCDIDs)
}

func (s *VagaService) ListByContratante(ctx context.Context, cnpj string, page, pageSize int) ([]*empregabilidade.Vaga, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByContratante(ctx, cnpj, pageSize, offset)
}

func (s *VagaService) ListByOrgaoParceiro(ctx context.Context, orgaoID string, page, pageSize int) ([]*empregabilidade.Vaga, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByOrgaoParceiro(ctx, orgaoID, pageSize, offset)
}

func (s *VagaService) Publish(ctx context.Context, id uuid.UUID) error {
	vaga, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if vaga == nil {
		return errors.New("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaEmEdicao && vaga.Status != empregabilidade.StatusVagaEmAprovacao {
		return errors.New("vaga não está em estado de edição ou aprovação")
	}

	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	return s.repo.Update(ctx, vaga)
}

func (s *VagaService) SendToApproval(ctx context.Context, id uuid.UUID) error {
	vaga, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if vaga == nil {
		return errors.New("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaEmEdicao {
		return errors.New("vaga não está em estado de edição")
	}

	vaga.Status = empregabilidade.StatusVagaEmAprovacao
	return s.repo.Update(ctx, vaga)
}
