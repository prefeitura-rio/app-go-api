package empregabilidade

import (
	"context"
	"log"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
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

	// Habilidades e competências da vaga (DEV-524).
	GetHabilidadesCompetencias(ctx context.Context, vagaID uuid.UUID) (*empregabilidade.VagaHabilidadesCompetencias, error)
	AreasAtuacaoExistentes(ctx context.Context, ids []int64) ([]int64, error)
	ConhecimentosDasAreas(ctx context.Context, areaIDs, habilidadeIDs []int64) (map[int64]map[int64]bool, error)
	ComportamentosExistentes(ctx context.Context, ids []int64) ([]int64, error)
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
		return apperrors.Validation("empresa contratante não encontrada")
	}
	return nil
}

func validateQuantidadeEstimada(v *int) error {
	if v != nil && *v <= 0 {
		return apperrors.Validation("quantidade_estimada_contratacoes deve ser um número inteiro positivo")
	}
	return nil
}

// normalizarCriteriosBloqueio valida os critérios de bloqueio recebidos e os
// deixa no formato gravado (nil quando nenhum está ligado).
func normalizarCriteriosBloqueio(criterios []empregabilidade.CriterioBloqueio) ([]empregabilidade.CriterioBloqueio, error) {
	normalizados, err := empregabilidade.NormalizarCriteriosBloqueio(criterios)
	if err != nil {
		return nil, apperrors.Validation(err.Error())
	}
	return normalizados, nil
}

// validarExperienciaCarteira confere o tempo de experiência em carteira
// exigido: "" vale como sem exigência, valor desconhecido é recusado e o
// bloqueio experiencia_carteira ligado exige o tempo. criterios são os
// critérios de bloqueio já normalizados que a vaga terá depois da escrita.
func validarExperienciaCarteira(entity *empregabilidade.Vaga, criterios []empregabilidade.CriterioBloqueio) error {
	if entity.ExperienciaCarteira != nil && *entity.ExperienciaCarteira == "" {
		entity.ExperienciaCarteira = nil
	}
	if entity.ExperienciaCarteira != nil && !entity.ExperienciaCarteira.IsValid() {
		return apperrors.Validation("experiencia_carteira inválida: deve ser qualquer_tempo ou minimo_6_meses")
	}
	if entity.ExperienciaCarteira == nil && slices.Contains(criterios, empregabilidade.CriterioBloqueioExperienciaCarteira) {
		return apperrors.Validation(empregabilidade.MsgExperienciaCarteiraObrigatoria)
	}
	return nil
}

// validarJovemAprendiz deixa a opção Jovem Aprendiz sempre definida (sem ela,
// desligada), recusa o bloqueio jovem_aprendiz ligado sem a opção marcada e
// recusa Jovem Aprendiz ativo junto de Maiores de 18 (idade_minima): a vaga
// tem um critério de idade por vez. criterios são os critérios de bloqueio já
// normalizados que a vaga terá depois da escrita.
func validarJovemAprendiz(entity *empregabilidade.Vaga, criterios []empregabilidade.CriterioBloqueio) error {
	if entity.JovemAprendiz == nil {
		desligado := false
		entity.JovemAprendiz = &desligado
	}
	bloqueioLigado := slices.Contains(criterios, empregabilidade.CriterioBloqueioJovemAprendiz)
	if !*entity.JovemAprendiz && bloqueioLigado {
		return apperrors.Validation(empregabilidade.MsgJovemAprendizObrigatorio)
	}
	if *entity.JovemAprendiz && bloqueioLigado && entity.IdadeMinima != nil {
		return apperrors.Validation(empregabilidade.MsgIdadeCriterioUnico)
	}
	return nil
}

// resolverOpcaoPCD valida a opção PcD e deixa opcao_pcd e o campo antigo
// acessibilidade_pcd coerentes. Com opcao_pcd enviada, ela vale e o campo
// antigo é gravado a partir dela; sem ela (cliente antigo), a opção sai do
// acessibilidade_pcd recebido pelo de/para. qtdTipos é quantos tipos de
// deficiência a vaga terá depois da escrita: a exclusiva exige ao menos um.
func resolverOpcaoPCD(entity *empregabilidade.Vaga, qtdTipos int) error {
	if entity.OpcaoPCD == "" {
		if entity.AcessibilidadePCD != nil && !entity.AcessibilidadePCD.IsValid() {
			return apperrors.Validation("acessibilidade_pcd inválido: deve ser para_pcd, preferencial_pcd ou exclusivo_pcd")
		}
		entity.OpcaoPCD = empregabilidade.OpcaoPCDDeAcessibilidade(entity.AcessibilidadePCD)
	} else {
		if !entity.OpcaoPCD.IsValid() {
			return apperrors.Validation("opcao_pcd inválida: deve ser aberta, preferencial ou exclusiva")
		}
		entity.AcessibilidadePCD = entity.OpcaoPCD.AcessibilidadeLegada()
	}
	if entity.OpcaoPCD == empregabilidade.OpcaoPCDExclusiva && qtdTipos == 0 {
		return apperrors.Validation(empregabilidade.MsgTipoPCDObrigatorio)
	}
	return nil
}

func (s *VagaService) Create(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error) {
	criterios, err := normalizarCriteriosBloqueio(entity.CriteriosBloqueio)
	if err != nil {
		return uuid.Nil, err
	}
	if err := validarExperienciaCarteira(entity, criterios); err != nil {
		return uuid.Nil, err
	}
	if err := validarJovemAprendiz(entity, criterios); err != nil {
		return uuid.Nil, err
	}
	if err := resolverOpcaoPCD(entity, len(entity.TiposPCD)); err != nil {
		return uuid.Nil, err
	}
	if err := s.validateContratante(ctx, entity.IDContratante); err != nil {
		return uuid.Nil, err
	}
	if err := validateQuantidadeEstimada(entity.QuantidadeEstimadaContratacoes); err != nil {
		return uuid.Nil, err
	}
	if err := normalizarTiposEtapa(entity.Etapas); err != nil {
		return uuid.Nil, err
	}
	if err := s.validarHabilidadesCompetencias(ctx, entity); err != nil {
		return uuid.Nil, err
	}

	entity.Status = empregabilidade.StatusVagaEmEdicao
	entity.CriteriosBloqueio = criterios

	id, err := s.repo.Create(ctx, entity)
	if err != nil {
		return uuid.Nil, err
	}
	// A resposta do POST devolve o estado como as leituras: [] sem critério ligado.
	if entity.CriteriosBloqueio == nil {
		entity.CriteriosBloqueio = []empregabilidade.CriterioBloqueio{}
	}
	// E as habilidades e competências como a leitura da gestão, com os nomes.
	// A vaga já foi criada: se a releitura falhar, a resposta leva o que veio
	// no corpo (sem os nomes) em vez de um erro que levaria a criar de novo.
	if err := s.CarregarHabilidadesCompetencias(ctx, entity); err != nil {
		log.Printf("vaga %s criada, mas a releitura das habilidades falhou: %v", id, err)
		entity.SetHabilidadesCompetencias(habilidadesDoCorpo(entity))
	}
	return id, nil
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

// GetByIDGestao é a leitura da vaga pela gestão: GetByID mais as áreas de
// atuação (com os conhecimentos), os comportamentos e atitudes. As rotas
// públicas usam GetByID, que não os carrega.
func (s *VagaService) GetByIDGestao(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error) {
	vaga, err := s.GetByID(ctx, id)
	if err != nil || vaga == nil {
		return vaga, err
	}
	if err := s.CarregarHabilidadesCompetencias(ctx, vaga); err != nil {
		return nil, err
	}
	return vaga, nil
}

// CarregarHabilidadesCompetencias põe na vaga o que está gravado para ela.
func (s *VagaService) CarregarHabilidadesCompetencias(ctx context.Context, vaga *empregabilidade.Vaga) error {
	habilidades, err := s.repo.GetHabilidadesCompetencias(ctx, vaga.ID)
	if err != nil {
		return err
	}
	vaga.SetHabilidadesCompetencias(habilidades)
	return nil
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
	if err := normalizarTiposEtapa(entity.Etapas); err != nil {
		return err
	}

	existing, err := s.repo.GetByID(ctx, entity.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperrors.NotFound("vaga não encontrada")
	}

	entity.Status = existing.Status

	// Preserve required FK fields when not provided in the request body.
	// id_orgao_parceiro preenchido no body é persistido; vazio mantém o atual.
	// A restrição de escopo (editor só nas próprias secretarias) fica no handler.
	if entity.IDOrgaoParceiro == "" {
		entity.IDOrgaoParceiro = existing.IDOrgaoParceiro
	}
	if entity.IDContratante == "" {
		entity.IDContratante = existing.IDContratante
	}
	if entity.IDRegimeContratacao == uuid.Nil {
		entity.IDRegimeContratacao = existing.IDRegimeContratacao
	}
	if entity.IDModeloTrabalho == uuid.Nil {
		entity.IDModeloTrabalho = existing.IDModeloTrabalho
	}
	// criterios_bloqueio ausente no body mantém o estado gravado; [] desliga todos.
	if entity.CriteriosBloqueio == nil {
		entity.CriteriosBloqueio = existing.CriteriosBloqueio
	}
	criterios, err := normalizarCriteriosBloqueio(entity.CriteriosBloqueio)
	if err != nil {
		return err
	}
	entity.CriteriosBloqueio = criterios
	// experiencia_carteira ausente no body mantém o valor gravado; "" remove.
	if entity.ExperienciaCarteira == nil {
		entity.ExperienciaCarteira = existing.ExperienciaCarteira
	}
	if err := validarExperienciaCarteira(entity, criterios); err != nil {
		return err
	}
	// jovem_aprendiz ausente no body mantém o valor gravado.
	if entity.JovemAprendiz == nil {
		entity.JovemAprendiz = existing.JovemAprendiz
	}
	if err := validarJovemAprendiz(entity, criterios); err != nil {
		return err
	}
	// tipos_pcd ausente no body mantém os tipos gravados; [] remove todos.
	tipos := entity.TiposPCD
	if tipos == nil {
		tipos = existing.TiposPCD
	}
	if err := resolverOpcaoPCD(entity, len(tipos)); err != nil {
		return err
	}
	// areas_atuacao e comportamento_atitudes_ids ausentes (nil) mantêm o
	// gravado; a lista enviada substitui o conjunto.
	if err := s.validarHabilidadesCompetencias(ctx, entity); err != nil {
		return err
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
	if len(tiposPCDIDs) == 0 {
		vaga, err := s.repo.GetByID(ctx, vagaID)
		if err != nil {
			return err
		}
		if vaga != nil && vaga.OpcaoPCD == empregabilidade.OpcaoPCDExclusiva {
			return apperrors.Validation(empregabilidade.MsgTipoPCDObrigatorio)
		}
	}
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
		return apperrors.NotFound("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaEmEdicao && vaga.Status != empregabilidade.StatusVagaEmAprovacao {
		return apperrors.Conflict("vaga não está em estado de edição ou aprovação")
	}
	if err := s.exigirHabilidadesCompetencias(ctx, id, "publicar"); err != nil {
		return err
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
		return apperrors.NotFound("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaEmAprovacao {
		return apperrors.Conflict("vaga não está em estado de aprovação")
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
		return apperrors.NotFound("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaEmEdicao {
		return apperrors.Conflict("vaga não está em estado de edição")
	}
	if err := s.exigirHabilidadesCompetencias(ctx, id, "enviar para aprovação"); err != nil {
		return err
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
		return apperrors.NotFound("vaga não encontrada")
	}

	vaga.UpdateStatusBasedOnExpiration()
	if vaga.Status != empregabilidade.StatusVagaPublicadoAtivo {
		return apperrors.Conflict("vaga não está em estado publicado_ativo para ser congelada")
	}

	if err := s.candidaturaRepo.BulkSaveAndUpdateStatusByVagaID(comOrigemAutomacao(ctx, empregabilidade.MotivoVagaCongelada), id, empregabilidade.StatusCandidaturaVagaCongelada); err != nil {
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
		return apperrors.NotFound("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaCongelada {
		return apperrors.Conflict("vaga não está em estado vaga_congelada para ser descongelada")
	}

	if err := s.candidaturaRepo.BulkRestoreStatusByVagaID(comOrigemAutomacao(ctx, empregabilidade.MotivoVagaDescongelada), id); err != nil {
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
		return apperrors.NotFound("vaga não encontrada")
	}

	vaga.UpdateStatusBasedOnExpiration()
	if vaga.Status != empregabilidade.StatusVagaPublicadoAtivo &&
		vaga.Status != empregabilidade.StatusVagaPublicadoExpirado &&
		vaga.Status != empregabilidade.StatusVagaCongelada {
		return apperrors.Conflict("vaga não está em estado publicado ou congelado para ser descontinuada")
	}

	if err := s.candidaturaRepo.BulkSaveAndUpdateStatusByVagaID(comOrigemAutomacao(ctx, empregabilidade.MotivoVagaDescontinuada), id, empregabilidade.StatusCandidaturaDescontinuada); err != nil {
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
		return apperrors.NotFound("vaga não encontrada")
	}

	if vaga.Status != empregabilidade.StatusVagaDescontinuada {
		return apperrors.Conflict("vaga não está em estado vaga_descontinuada para ser reativada")
	}

	if err := s.candidaturaRepo.BulkRestoreStatusByVagaID(comOrigemAutomacao(ctx, empregabilidade.MotivoVagaReativada), id); err != nil {
		return err
	}

	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vaga.UpdateStatusBasedOnExpiration()
	return s.repo.Update(ctx, vaga)
}

// comOrigemAutomacao atribui à automação, com o motivo, as transições que a
// mudança de estado da vaga provoca nas candidaturas (histórico, DEV-532).
func comOrigemAutomacao(ctx context.Context, motivo string) context.Context {
	return empregabilidade.ComOrigemEvento(ctx, empregabilidade.OrigemEventoCandidatura{
		Ator:   empregabilidade.AtorAutomacao,
		Motivo: motivo,
	})
}

// normalizarTiposEtapa ajusta o tipo_etapa das etapas enviadas com a vaga.
func normalizarTiposEtapa(etapas []empregabilidade.Etapa) error {
	for i := range etapas {
		if err := normalizarTipoEtapa(&etapas[i]); err != nil {
			return err
		}
	}
	return nil
}
