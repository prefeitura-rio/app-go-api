package empregabilidade

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	repository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

type VagaRepoInterface interface {
	Create(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error)
	GetByIDPrefix(ctx context.Context, idPrefix string) (*empregabilidade.Vaga, error)
	Update(ctx context.Context, entity *empregabilidade.Vaga) error
	UpdateWithAssociations(ctx context.Context, entity *empregabilidade.Vaga) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter empregabilidade.VagaFilter, limit, offset int) ([]*empregabilidade.Vaga, int, error)
	ListPublicActive(ctx context.Context, limit, offset int) ([]*empregabilidade.Vaga, int, error)
	ListPublic(ctx context.Context, filter empregabilidade.VagaPublicFilter, limit, offset int) ([]*empregabilidade.Vaga, int, error)
	UpdateTiposPCD(ctx context.Context, vagaID uuid.UUID, tiposPCDIDs []uuid.UUID) error
	UpdateZonas(ctx context.Context, vagaID uuid.UUID, zonaIDs []uuid.UUID) error
	UpdateIdiomasRequisito(ctx context.Context, vagaID uuid.UUID, requisitos []empregabilidade.VagaIdiomaRequisito) error
	ListByContratante(ctx context.Context, cnpj string, limit, offset int) ([]*empregabilidade.Vaga, int, error)
	ListByOrgaoParceiro(ctx context.Context, orgaoID string, limit, offset int) ([]*empregabilidade.Vaga, int, error)
}

type EmpresaRepoInterface interface {
	GetByID(ctx context.Context, cnpj string) (*empregabilidade.Empresa, error)
}

type CandidaturaRepoForVagaInterface interface {
	BulkSaveAndUpdateStatusByVagaID(ctx context.Context, vagaID uuid.UUID, status empregabilidade.StatusCandidatura) error
	BulkRestoreStatusByVagaID(ctx context.Context, vagaID uuid.UUID) error
}

type VagaService struct {
	repo            VagaRepoInterface
	empresaRepo     EmpresaRepoInterface
	candidaturaRepo CandidaturaRepoForVagaInterface
}

func NewVagaService(
	repo *repository.VagaRepository,
	empresaRepo *repository.EmpresaRepository,
	candidaturaRepo *repository.CandidaturaRepository,
) *VagaService {
	return &VagaService{
		repo:            repo,
		empresaRepo:     empresaRepo,
		candidaturaRepo: candidaturaRepo,
	}
}

func NewVagaServiceWithInterfaces(
	repo VagaRepoInterface,
	empresaRepo EmpresaRepoInterface,
	candidaturaRepo CandidaturaRepoForVagaInterface,
) *VagaService {
	return &VagaService{
		repo:            repo,
		empresaRepo:     empresaRepo,
		candidaturaRepo: candidaturaRepo,
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

func validateQuantidadeEstimada(v *int) error {
	if v != nil && *v <= 0 {
		return errors.New("quantidade_estimada_contratacoes deve ser um número inteiro positivo")
	}
	return nil
}

func (s *VagaService) Create(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error) {
	if err := s.validateContratante(ctx, entity.IDContratante); err != nil {
		return uuid.Nil, err
	}
	if err := validateQuantidadeEstimada(entity.QuantidadeEstimadaContratacoes); err != nil {
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

var hexPrefix12 = regexp.MustCompile(`^[0-9a-f]{12}$`)

func (s *VagaService) GetBySlug(ctx context.Context, vagaSlug string) (*empregabilidade.Vaga, error) {
	parts := strings.Split(vagaSlug, "-")
	idPrefix := parts[len(parts)-1]
	if !hexPrefix12.MatchString(idPrefix) {
		return nil, nil
	}
	vaga, err := s.repo.GetByIDPrefix(ctx, idPrefix)
	if err != nil {
		return nil, err
	}
	if vaga != nil {
		vaga.UpdateStatusBasedOnExpiration()
	}
	return vaga, nil
}

func (s *VagaService) Update(ctx context.Context, entity *empregabilidade.Vaga) error {
	if err := validateQuantidadeEstimada(entity.QuantidadeEstimadaContratacoes); err != nil {
		return err
	}

	existing, err := s.repo.GetByID(ctx, entity.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("vaga não encontrada")
	}

	entity.Status = existing.Status

	// O órgão parceiro define o escopo de autorização da vaga e nunca é
	// reatribuído por um Update: preserva sempre o valor existente para evitar
	// que um editor mova a vaga para fora da sua secretaria/órgão via body.
	entity.IDOrgaoParceiro = existing.IDOrgaoParceiro

	// Preserve required FK fields when not provided in the request body
	if entity.IDContratante == "" {
		entity.IDContratante = existing.IDContratante
	}
	if entity.IDRegimeContratacao == uuid.Nil {
		entity.IDRegimeContratacao = existing.IDRegimeContratacao
	}
	if entity.IDModeloTrabalho == uuid.Nil {
		entity.IDModeloTrabalho = existing.IDModeloTrabalho
	}

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

func (s *VagaService) ListPublic(ctx context.Context, filter empregabilidade.VagaPublicFilter, page, pageSize int) ([]*empregabilidade.Vaga, int, error) {
	offset := (page - 1) * pageSize
	vagas, total, err := s.repo.ListPublic(ctx, filter, pageSize, offset)
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

func (s *VagaService) UpdateZonas(ctx context.Context, vagaID uuid.UUID, zonaIDs []uuid.UUID) error {
	return s.repo.UpdateZonas(ctx, vagaID, zonaIDs)
}

func (s *VagaService) UpdateIdiomasRequisito(ctx context.Context, vagaID uuid.UUID, requisitos []empregabilidade.VagaIdiomaRequisito) error {
	return s.repo.UpdateIdiomasRequisito(ctx, vagaID, requisitos)
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

func (s *VagaService) SendToDraft(ctx context.Context, id uuid.UUID) error {
	vaga, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if vaga == nil {
		return errors.New("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaEmAprovacao {
		return errors.New("vaga não está em estado de aprovação")
	}

	vaga.Status = empregabilidade.StatusVagaEmEdicao
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

func (s *VagaService) FreezeVaga(ctx context.Context, id uuid.UUID) error {
	vaga, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if vaga == nil {
		return errors.New("vaga não encontrada")
	}

	vaga.UpdateStatusBasedOnExpiration()
	if vaga.Status != empregabilidade.StatusVagaPublicadoAtivo {
		return errors.New("vaga não está em estado publicado_ativo para ser congelada")
	}

	if err := s.candidaturaRepo.BulkSaveAndUpdateStatusByVagaID(ctx, id, empregabilidade.StatusCandidaturaVagaCongelada); err != nil {
		return err
	}

	vaga.Status = empregabilidade.StatusVagaCongelada
	return s.repo.Update(ctx, vaga)
}

func (s *VagaService) UnfreezeVaga(ctx context.Context, id uuid.UUID) error {
	vaga, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if vaga == nil {
		return errors.New("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaCongelada {
		return errors.New("vaga não está em estado vaga_congelada para ser descongelada")
	}

	if err := s.candidaturaRepo.BulkRestoreStatusByVagaID(ctx, id); err != nil {
		return err
	}

	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vaga.UpdateStatusBasedOnExpiration()
	return s.repo.Update(ctx, vaga)
}

func (s *VagaService) DiscontinueVaga(ctx context.Context, id uuid.UUID) error {
	vaga, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if vaga == nil {
		return errors.New("vaga não encontrada")
	}

	vaga.UpdateStatusBasedOnExpiration()
	if vaga.Status != empregabilidade.StatusVagaPublicadoAtivo &&
		vaga.Status != empregabilidade.StatusVagaPublicadoExpirado &&
		vaga.Status != empregabilidade.StatusVagaCongelada {
		return errors.New("vaga não está em estado publicado ou congelado para ser descontinuada")
	}

	if err := s.candidaturaRepo.BulkSaveAndUpdateStatusByVagaID(ctx, id, empregabilidade.StatusCandidaturaDescontinuada); err != nil {
		return err
	}

	vaga.Status = empregabilidade.StatusVagaDescontinuada
	return s.repo.Update(ctx, vaga)
}

func (s *VagaService) ReactivateVaga(ctx context.Context, id uuid.UUID) error {
	vaga, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if vaga == nil {
		return errors.New("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaDescontinuada {
		return errors.New("vaga não está em estado vaga_descontinuada para ser reativada")
	}

	if err := s.candidaturaRepo.BulkRestoreStatusByVagaID(ctx, id); err != nil {
		return err
	}

	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vaga.UpdateStatusBasedOnExpiration()
	return s.repo.Update(ctx, vaga)
}
