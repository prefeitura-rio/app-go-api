package empregabilidade

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// ErrCandidatoNaoAssociavel indica que o candidato não pode ser associado a
// nenhuma vaga: falta o currículo mínimo que o app exige para se candidatar.
var ErrCandidatoNaoAssociavel = errors.New("candidato não pode ser associado a vagas")

// CandidatoNaoAssociavelError lista o que falta no currículo do candidato.
type CandidatoNaoAssociavelError struct {
	Pendencias []string
}

func (e *CandidatoNaoAssociavelError) Error() string {
	return fmt.Sprintf("%s: %s", ErrCandidatoNaoAssociavel, strings.Join(e.Pendencias, "; "))
}

func (e *CandidatoNaoAssociavelError) Unwrap() error { return ErrCandidatoNaoAssociavel }

// CandidaturaCreatorInterface cria a candidatura pelo fluxo normal (snapshot do
// currículo, sync do cadastro), sem o e-mail de "candidatura enviada".
type CandidaturaCreatorInterface interface {
	CreateSemNotificacao(ctx context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error)
}

type VagaGetterInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error)
}

type AceiteTermosCheckerInterface interface {
	HasAcceptedTerms(ctx context.Context, cpf string) (bool, error)
}

// CandidaturasDoCPFInterface diz em quais vagas o CPF já tem candidatura ativa.
type CandidaturasDoCPFInterface interface {
	VagasComCandidatura(ctx context.Context, cpf string, idsVagas []uuid.UUID) (map[uuid.UUID]bool, error)
}

type EscolaridadeListerInterface interface {
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.Escolaridade, int, error)
}

type IdiomaListerInterface interface {
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.Idioma, int, error)
}

type NivelIdiomaListerInterface interface {
	List(ctx context.Context, filter map[string]interface{}, page, pageSize int) ([]*empregabilidade.NivelIdioma, int, error)
}

// tamanhoTabelaApoio cobre as tabelas de apoio inteiras numa página só.
const tamanhoTabelaApoio = 1000

// AssociacaoCandidatoService associa um candidato do banco de currículos a
// vagas. Valem as mesmas regras da candidatura feita pelo app: currículo
// mínimo, termos aceitos, vaga recebendo candidaturas e requisitos da vaga. As
// perguntas complementares da vaga não são respondidas: a secretaria confere
// esses critérios conversando com o candidato. O candidato não recebe o e-mail
// de "candidatura enviada"; a vaga aparece em "Minhas candidaturas" no app.
type AssociacaoCandidatoService struct {
	curriculos         BancoCurriculoRepositoryInterface
	curriculoCompleto  CurriculoCompletoGetterInterface
	candidaturasDoCPF  CandidaturasDoCPFInterface
	termos             AceiteTermosCheckerInterface
	vagas              VagaGetterInterface
	candidaturas       CandidaturaCreatorInterface
	escolaridades      EscolaridadeListerInterface
	idiomas            IdiomaListerInterface
	niveisIdioma       NivelIdiomaListerInterface
	citizenSnapshots   CitizenSnapshotRepoForCandidaturaInterface
	citizenDataFetcher CitizenDataFetcherForCandidaturaInterface // pode ser nil
	agora              func() time.Time
}

func NewAssociacaoCandidatoService(
	curriculos BancoCurriculoRepositoryInterface,
	curriculoCompleto CurriculoCompletoGetterInterface,
	candidaturasDoCPF CandidaturasDoCPFInterface,
	termos AceiteTermosCheckerInterface,
	vagas VagaGetterInterface,
	candidaturas CandidaturaCreatorInterface,
	escolaridades EscolaridadeListerInterface,
	idiomas IdiomaListerInterface,
	niveisIdioma NivelIdiomaListerInterface,
	citizenSnapshots CitizenSnapshotRepoForCandidaturaInterface,
) *AssociacaoCandidatoService {
	return &AssociacaoCandidatoService{
		curriculos:        curriculos,
		curriculoCompleto: curriculoCompleto,
		candidaturasDoCPF: candidaturasDoCPF,
		termos:            termos,
		vagas:             vagas,
		candidaturas:      candidaturas,
		escolaridades:     escolaridades,
		idiomas:           idiomas,
		niveisIdioma:      niveisIdioma,
		citizenSnapshots:  citizenSnapshots,
		agora:             time.Now,
	}
}

// SetCitizenDataFetcher liga o sync do cadastro depois da construção, quando o
// worker é iniciado no router.
func (s *AssociacaoCandidatoService) SetCitizenDataFetcher(fetcher CitizenDataFetcherForCandidaturaInterface) {
	s.citizenDataFetcher = fetcher
}

// AssociarParams descreve uma associação pedida pelo operador.
type AssociarParams struct {
	CPF         string
	IDsVagas    []uuid.UUID
	OperadorCPF string
	// OrgaosPermitidos limita as vagas aos órgãos do operador. Nil libera
	// todos os órgãos (admins).
	OrgaosPermitidos []string
}

// Associar tenta associar o candidato a cada vaga e devolve o resultado de
// cada uma, na ordem pedida. Retorna (nil, nil) quando o CPF não tem currículo
// e CandidatoNaoAssociavelError quando falta o currículo mínimo.
func (s *AssociacaoCandidatoService) Associar(ctx context.Context, params AssociarParams) ([]*empregabilidade.AssociacaoVaga, error) {
	raiz, err := s.curriculos.GetCurriculoByCPF(ctx, params.CPF)
	if err != nil {
		return nil, err
	}
	if raiz == nil {
		return nil, nil
	}

	curriculo, err := s.curriculoCompleto.GetCurriculoCompleto(ctx, params.CPF)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar currículo: %w", err)
	}
	if err := s.validarCurriculoMinimo(ctx, params.CPF, curriculo); err != nil {
		return nil, err
	}

	idsVagas := semRepetir(params.IDsVagas)
	jaParticipa, err := s.candidaturasDoCPF.VagasComCandidatura(ctx, params.CPF, idsVagas)
	if err != nil {
		return nil, err
	}
	refs, err := s.carregarReferencias(ctx)
	if err != nil {
		return nil, err
	}

	cadastro := s.carregarCadastro(ctx, params.CPF)
	candidato := candidatoElegibilidade{curriculo: curriculo}
	if cadastro != nil {
		candidato.dataNascimento = cadastro.DataNascimento
		candidato.escolaridadeCadastro = cadastro.Escolaridade
	}

	resultados := make([]*empregabilidade.AssociacaoVaga, 0, len(idsVagas))
	for _, idVaga := range idsVagas {
		resultado := s.associarVaga(ctx, params, idVaga, jaParticipa[idVaga], candidato, cadastro, refs)
		resultados = append(resultados, resultado)
	}
	return resultados, nil
}

func (s *AssociacaoCandidatoService) associarVaga(
	ctx context.Context,
	params AssociarParams,
	idVaga uuid.UUID,
	jaParticipa bool,
	candidato candidatoElegibilidade,
	cadastro *models.CitizenSnapshot,
	refs referenciasElegibilidade,
) *empregabilidade.AssociacaoVaga {
	resultado := &empregabilidade.AssociacaoVaga{IDVaga: idVaga}

	vaga, err := s.vagas.GetByID(ctx, idVaga)
	if err != nil {
		log.Printf("[AssociacaoCandidatoService] falha ao carregar vaga %s: %v", idVaga, err)
		resultado.Resultado = empregabilidade.AssociacaoErro
		resultado.Motivos = []string{"Não foi possível carregar a vaga"}
		return resultado
	}
	if vaga == nil {
		resultado.Resultado = empregabilidade.AssociacaoVagaIndisponivel
		resultado.Motivos = []string{"Vaga não encontrada"}
		return resultado
	}
	resultado.Titulo = vaga.Titulo

	if params.OrgaosPermitidos != nil && !slices.Contains(params.OrgaosPermitidos, vaga.IDOrgaoParceiro) {
		resultado.Resultado = empregabilidade.AssociacaoForaDoEscopo
		resultado.Motivos = []string{"A vaga não pertence à sua secretaria"}
		return resultado
	}

	vaga.UpdateStatusBasedOnExpiration()
	if vaga.Status != empregabilidade.StatusVagaPublicadoAtivo {
		resultado.Resultado = empregabilidade.AssociacaoVagaIndisponivel
		resultado.Motivos = []string{"A vaga não está recebendo candidaturas"}
		return resultado
	}

	if jaParticipa {
		resultado.Resultado = empregabilidade.AssociacaoJaParticipa
		return resultado
	}

	if faltas := requisitosNaoAtendidos(vaga, candidato, refs, s.agora()); len(faltas) > 0 {
		resultado.Resultado = empregabilidade.AssociacaoNaoElegivel
		resultado.Motivos = faltas
		return resultado
	}

	origem := empregabilidade.OrigemCandidaturaBancoCurriculos
	operador := params.OperadorCPF
	candidatura := &empregabilidade.Candidatura{
		CPF:             params.CPF,
		IDVaga:          idVaga,
		Origem:          &origem,
		AssociadoPorCPF: &operador,
	}
	if cadastro != nil {
		candidatura.Nome = naoVazio(cadastro.Nome)
		candidatura.Email = naoVazio(cadastro.Email)
	}

	id, err := s.candidaturas.CreateSemNotificacao(ctx, candidatura)
	if err != nil {
		// Outra associação pode ter criado a candidatura entre a checagem e aqui.
		if strings.Contains(err.Error(), "já existe") {
			resultado.Resultado = empregabilidade.AssociacaoJaParticipa
			return resultado
		}
		log.Printf("[AssociacaoCandidatoService] falha ao associar à vaga %s: %v", idVaga, err)
		resultado.Resultado = empregabilidade.AssociacaoErro
		resultado.Motivos = []string{"Não foi possível associar o candidato a esta vaga"}
		return resultado
	}

	resultado.Resultado = empregabilidade.AssociacaoRealizada
	resultado.IDCandidatura = &id
	return resultado
}

// validarCurriculoMinimo aplica o mínimo que o app exige para se candidatar
// (pendenciasCurriculoMinimo).
func (s *AssociacaoCandidatoService) validarCurriculoMinimo(ctx context.Context, cpf string, curriculo *empregabilidade.CurriculoCompleto) error {
	aceitos, err := s.termos.HasAcceptedTerms(ctx, cpf)
	if err != nil {
		return fmt.Errorf("erro ao verificar termos de uso: %w", err)
	}
	if pendencias := pendenciasCurriculoMinimo(curriculo, aceitos); len(pendencias) > 0 {
		return &CandidatoNaoAssociavelError{Pendencias: pendencias}
	}
	return nil
}

func (s *AssociacaoCandidatoService) carregarReferencias(ctx context.Context) (referenciasElegibilidade, error) {
	var refs referenciasElegibilidade
	var err error
	if refs.escolaridades, _, err = s.escolaridades.List(ctx, nil, 1, tamanhoTabelaApoio); err != nil {
		return refs, fmt.Errorf("erro ao carregar escolaridades: %w", err)
	}
	if refs.idiomas, _, err = s.idiomas.List(ctx, nil, 1, tamanhoTabelaApoio); err != nil {
		return refs, fmt.Errorf("erro ao carregar idiomas: %w", err)
	}
	if refs.niveisIdioma, _, err = s.niveisIdioma.List(ctx, nil, 1, tamanhoTabelaApoio); err != nil {
		return refs, fmt.Errorf("erro ao carregar níveis de idioma: %w", err)
	}
	return refs, nil
}

// carregarCadastro busca os dados pessoais do cidadão. Sem cadastro, a idade
// e a escolaridade do cadastro ficam desconhecidas e os requisitos que
// dependem delas não são atendidos.
func (s *AssociacaoCandidatoService) carregarCadastro(ctx context.Context, cpf string) *models.CitizenSnapshot {
	return carregarCadastroCidadao(ctx, s.citizenSnapshots, s.citizenDataFetcher, cpf)
}

func semRepetir(ids []uuid.UUID) []uuid.UUID {
	vistos := make(map[uuid.UUID]bool, len(ids))
	unicos := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if !vistos[id] {
			vistos[id] = true
			unicos = append(unicos, id)
		}
	}
	return unicos
}
