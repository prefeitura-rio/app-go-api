package empregabilidade

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type VagaRepoInterface interface {
	Create(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error)
	Update(ctx context.Context, entity *empregabilidade.Vaga) error
	UpdateWithAssociations(ctx context.Context, entity *empregabilidade.Vaga) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter empregabilidade.VagaFilter, limit, offset int) ([]*empregabilidade.Vaga, int, error)
	ListPublicActive(ctx context.Context, limit, offset int) ([]*empregabilidade.Vaga, int, error)
	UpdateTiposPCD(ctx context.Context, vagaID uuid.UUID, tiposPCDIDs []uuid.UUID) error
	ListByContratante(ctx context.Context, cnpj string, limit, offset int) ([]*empregabilidade.Vaga, int, error)
	ListByOrgaoParceiro(ctx context.Context, orgaoID string, limit, offset int) ([]*empregabilidade.Vaga, int, error)
}

type EmpresaRepoInterface interface {
	GetByID(ctx context.Context, cnpj string) (*empregabilidade.Empresa, error)
}

type VagaService struct {
	repo        VagaRepoInterface
	empresaRepo EmpresaRepoInterface
}

func NewVagaService(
	repo *repository.VagaRepository,
	empresaRepo *repository.EmpresaRepository,
) *VagaService {
	return &VagaService{
		repo:        repo,
		empresaRepo: empresaRepo,
	}
}

func NewVagaServiceWithInterfaces(
	repo VagaRepoInterface,
	empresaRepo EmpresaRepoInterface,
) *VagaService {
	return &VagaService{
		repo:        repo,
		empresaRepo: empresaRepo,
	}
}

// CreateDraft is an alias for Create (both create a vaga in em_edicao status).
func (s *VagaService) CreateDraft(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error) {
	return s.Create(ctx, entity)
}

func (s *VagaService) validateContratante(ctx context.Context, cnpj string) error {
	empresa, err := s.empresaRepo.GetByID(ctx, cnpj)
	if err != nil {
		return err
	}
	if empresa == nil {
		return errors.New("empresa contratante não encontrada")
	}
	return nil
}

func (s *VagaService) Create(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error) {
	if err := s.validateContratante(ctx, entity.IDContratante); err != nil {
		return uuid.Nil, err
	}

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
	existing, err := s.repo.GetByID(ctx, entity.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("vaga não encontrada")
	}

	entity.Status = existing.Status

	return s.repo.UpdateWithAssociations(ctx, entity)
}

func (s *VagaService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *VagaService) List(ctx context.Context, filter empregabilidade.VagaFilter, page, pageSize int) ([]*empregabilidade.Vaga, int, error) {
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

func (s *VagaService) ListPublicActive(ctx context.Context, page, pageSize int) ([]*empregabilidade.Vaga, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPublicActive(ctx, pageSize, offset)
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
