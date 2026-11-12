package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ──────────────────────────────────────────────────────────────────────────────
// Catálogo e gravação simulados no MockVagaRepoForService
// ──────────────────────────────────────────────────────────────────────────────

// catalogoHabilidadesMock é o catálogo de áreas (com os conhecimentos de cada
// uma) e de comportamentos que o mock consulta.
type catalogoHabilidadesMock struct {
	areas          map[int64]string
	conhecimentos  map[int64]map[int64]string // área → conhecimento → nome
	comportamentos map[int64]string
}

// novoCatalogoHabilidadesMock: área 3 (Tecnologia) com os conhecimentos 10 e
// 11; área 4 (Saúde) com o conhecimento 20; comportamentos 7 e 8.
func novoCatalogoHabilidadesMock() catalogoHabilidadesMock {
	return catalogoHabilidadesMock{
		areas: map[int64]string{3: "Tecnologia", 4: "Saúde"},
		conhecimentos: map[int64]map[int64]string{
			3: {10: "Go", 11: "SQL"},
			4: {20: "Primeiros socorros"},
		},
		comportamentos: map[int64]string{7: "Proatividade", 8: "Trabalho em equipe"},
	}
}

func (m *MockVagaRepoForService) gravarHabilidades(id uuid.UUID, entity *empregabilidade.Vaga) {
	atual := m.habilidades[id]
	if atual == nil {
		atual = &empregabilidade.VagaHabilidadesCompetencias{}
		m.habilidades[id] = atual
	}
	if entity.AreasAtuacao != nil {
		atual.AreasAtuacao = nil
		for _, a := range *entity.AreasAtuacao {
			area := empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: a.IDAreaAtuacao, Nome: m.catalogo.areas[a.IDAreaAtuacao]}
			for _, c := range a.ConhecimentoIDs {
				area.Conhecimentos = append(area.Conhecimentos, empregabilidade.CatalogoItem{ID: c, Nome: m.catalogo.conhecimentos[a.IDAreaAtuacao][c]})
			}
			atual.AreasAtuacao = append(atual.AreasAtuacao, area)
		}
	}
	if entity.ComportamentoAtitudesIDs != nil {
		atual.ComportamentosAtitudes = nil
		for _, c := range *entity.ComportamentoAtitudesIDs {
			atual.ComportamentosAtitudes = append(atual.ComportamentosAtitudes, empregabilidade.CatalogoItem{ID: c, Nome: m.catalogo.comportamentos[c]})
		}
	}
}

func (m *MockVagaRepoForService) GetHabilidadesCompetencias(_ context.Context, vagaID uuid.UUID) (*empregabilidade.VagaHabilidadesCompetencias, error) {
	if m.habilidadesError != nil {
		return nil, m.habilidadesError
	}
	h := m.habilidades[vagaID]
	if h == nil {
		return &empregabilidade.VagaHabilidadesCompetencias{}, nil
	}
	copia := *h
	return &copia, nil
}

func (m *MockVagaRepoForService) AreasAtuacaoExistentes(_ context.Context, ids []int64) ([]int64, error) {
	if m.catalogoError != nil {
		return nil, m.catalogoError
	}
	var out []int64
	for _, id := range ids {
		if _, ok := m.catalogo.areas[id]; ok {
			out = append(out, id)
		}
	}
	return out, nil
}

func (m *MockVagaRepoForService) ConhecimentosDasAreas(_ context.Context, areaIDs, habilidadeIDs []int64) (map[int64]map[int64]bool, error) {
	if m.catalogoError != nil {
		return nil, m.catalogoError
	}
	out := map[int64]map[int64]bool{}
	for _, a := range areaIDs {
		for _, h := range habilidadeIDs {
			if _, ok := m.catalogo.conhecimentos[a][h]; ok {
				if out[a] == nil {
					out[a] = map[int64]bool{}
				}
				out[a][h] = true
			}
		}
	}
	return out, nil
}

func (m *MockVagaRepoForService) ComportamentosExistentes(_ context.Context, ids []int64) ([]int64, error) {
	if m.catalogoError != nil {
		return nil, m.catalogoError
	}
	var out []int64
	for _, id := range ids {
		if _, ok := m.catalogo.comportamentos[id]; ok {
			out = append(out, id)
		}
	}
	return out, nil
}

// seedHabilidadesCompletas grava na vaga uma área com conhecimento e um
// comportamento: o mínimo para publicar ou enviar para aprovação.
func seedHabilidadesCompletas(m *MockVagaRepoForService, vagaID uuid.UUID) {
	m.habilidades[vagaID] = &empregabilidade.VagaHabilidadesCompetencias{
		AreasAtuacao: []empregabilidade.VagaAreaAtuacao{
			{IDAreaAtuacao: 3, Nome: "Tecnologia", Conhecimentos: []empregabilidade.CatalogoItem{{ID: 10, Nome: "Go"}}},
		},
		ComportamentosAtitudes: []empregabilidade.CatalogoItem{{ID: 7, Nome: "Proatividade"}},
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

func novoServicoHabilidades() (*services.VagaService, *MockVagaRepoForService) {
	repo := NewMockVagaRepoForService()
	empresas := NewMockEmpresaRepoForService()
	empresas.empresas["12345678000190"] = &empregabilidade.Empresa{CNPJ: "12345678000190"}
	return services.NewVagaServiceWithInterfaces(repo, empresas, nil), repo
}

func vagaNovaComHabilidades(areas *[]empregabilidade.VagaAreaAtuacao, comportamentos *[]int64) *empregabilidade.Vaga {
	return &empregabilidade.Vaga{
		Titulo:                   "Analista",
		IDContratante:            "12345678000190",
		AreasAtuacao:             areas,
		ComportamentoAtitudesIDs: comportamentos,
	}
}

func areas(a ...empregabilidade.VagaAreaAtuacao) *[]empregabilidade.VagaAreaAtuacao { return &a }
func ids(v ...int64) *[]int64                                                       { return &v }

// requireFieldError confere que err é 400 com o erro do campo informado.
func requireFieldError(t *testing.T, err error, campo, trecho string) {
	t.Helper()
	require.Error(t, err)
	appErr, ok := apperrors.As(err)
	require.True(t, ok, "esperava erro de domínio, veio %T: %v", err, err)
	assert.Equal(t, apperrors.KindValidation, appErr.Kind)
	require.Contains(t, appErr.Fields, campo, "fields: %v", appErr.Fields)
	assert.Contains(t, appErr.Fields[campo], trecho)
}

// ──────────────────────────────────────────────────────────────────────────────
// Criar
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaService_Create_GravaEDevolveAreasConhecimentosEComportamentos(t *testing.T) {
	svc, repo := novoServicoHabilidades()
	vaga := vagaNovaComHabilidades(
		areas(
			empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 3, ConhecimentoIDs: []int64{10, 11, 10}},
			empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 4, ConhecimentoIDs: []int64{20}},
		),
		ids(7, 8, 7),
	)

	id, err := svc.Create(context.Background(), vaga)
	require.NoError(t, err)

	gravada := repo.habilidades[id]
	require.Len(t, gravada.AreasAtuacao, 2)
	assert.Equal(t, []empregabilidade.CatalogoItem{{ID: 10, Nome: "Go"}, {ID: 11, Nome: "SQL"}}, gravada.AreasAtuacao[0].Conhecimentos, "conhecimento repetido na mesma área vira um só")
	assert.Equal(t, []empregabilidade.CatalogoItem{{ID: 7, Nome: "Proatividade"}, {ID: 8, Nome: "Trabalho em equipe"}}, gravada.ComportamentosAtitudes, "comportamento repetido vira um só")

	// A resposta do POST traz os nomes, como a leitura.
	require.NotNil(t, vaga.AreasAtuacao)
	assert.Equal(t, "Tecnologia", (*vaga.AreasAtuacao)[0].Nome)
	assert.Equal(t, []empregabilidade.CatalogoItem{{ID: 10, Nome: "Go"}, {ID: 11, Nome: "SQL"}}, (*vaga.AreasAtuacao)[0].Conhecimentos)
	require.NotNil(t, vaga.ComportamentosAtitudes)
	assert.Equal(t, "Proatividade", (*vaga.ComportamentosAtitudes)[0].Nome)
}

// Rascunho salva sem as listas: nada é exigido e a resposta traz listas vazias.
func TestVagaService_Create_RascunhoSemListasSalva(t *testing.T) {
	svc, repo := novoServicoHabilidades()
	vaga := vagaNovaComHabilidades(nil, nil)

	id, err := svc.Create(context.Background(), vaga)
	require.NoError(t, err)
	assert.Empty(t, repo.habilidades[id].AreasAtuacao, "nada gravado")
	assert.Empty(t, repo.habilidades[id].ComportamentosAtitudes, "nada gravado")
	require.NotNil(t, vaga.AreasAtuacao)
	assert.Empty(t, *vaga.AreasAtuacao)
	require.NotNil(t, vaga.ComportamentoAtitudesIDs)
	assert.Empty(t, *vaga.ComportamentoAtitudesIDs)
}

func TestVagaService_Create_RecusaEntradaInvalidaDasHabilidades(t *testing.T) {
	tests := []struct {
		name           string
		areas          *[]empregabilidade.VagaAreaAtuacao
		comportamentos *[]int64
		campo          string
		trecho         string
	}{
		{
			name:   "área repetida na vaga",
			areas:  areas(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 3}, empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 3, ConhecimentoIDs: []int64{10}}),
			campo:  "areas_atuacao",
			trecho: "repetida",
		},
		{
			name:   "conhecimento de outra área",
			areas:  areas(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 3, ConhecimentoIDs: []int64{10, 20}}),
			campo:  "areas_atuacao.conhecimento_ids",
			trecho: "não pertence à área de atuação",
		},
		{
			name:   "conhecimento que não existe",
			areas:  areas(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 3, ConhecimentoIDs: []int64{999}}),
			campo:  "areas_atuacao.conhecimento_ids",
			trecho: "999",
		},
		{
			name:   "área que não existe",
			areas:  areas(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 99}),
			campo:  "areas_atuacao",
			trecho: "não encontrada",
		},
		{
			name:   "id de área inválido",
			areas:  areas(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 0}),
			campo:  "areas_atuacao",
			trecho: "id_area_atuacao inválido",
		},
		{
			name:   "id de conhecimento inválido",
			areas:  areas(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 3, ConhecimentoIDs: []int64{-1}}),
			campo:  "areas_atuacao.conhecimento_ids",
			trecho: "inválido",
		},
		{
			name:           "comportamento que não existe",
			comportamentos: ids(7, 70),
			campo:          "comportamento_atitudes_ids",
			trecho:         "70",
		},
		{
			name:           "id de comportamento inválido",
			comportamentos: ids(0),
			campo:          "comportamento_atitudes_ids",
			trecho:         "inválido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := novoServicoHabilidades()
			_, err := svc.Create(context.Background(), vagaNovaComHabilidades(tt.areas, tt.comportamentos))
			requireFieldError(t, err, tt.campo, tt.trecho)
			assert.Empty(t, repo.vagas, "nada pode ser gravado")
		})
	}
}

// Listas grandes demais são recusadas antes de consultar o banco (o IN da
// validação tem limite de parâmetros no Postgres).
func TestVagaService_Create_RecusaListasAcimaDoLimite(t *testing.T) {
	muitasAreas := make([]empregabilidade.VagaAreaAtuacao, services.MaxAreasAtuacaoPorVaga+1)
	for i := range muitasAreas {
		muitasAreas[i] = empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: int64(i + 1)}
	}
	muitosConhecimentos := make([]int64, services.MaxConhecimentosPorArea+1)
	for i := range muitosConhecimentos {
		muitosConhecimentos[i] = int64(i + 1)
	}
	muitosComportamentos := make([]int64, services.MaxComportamentosPorVaga+1)
	for i := range muitosComportamentos {
		muitosComportamentos[i] = int64(i + 1)
	}

	tests := []struct {
		name           string
		areas          *[]empregabilidade.VagaAreaAtuacao
		comportamentos *[]int64
		campo          string
	}{
		{"áreas demais", &muitasAreas, nil, "areas_atuacao"},
		{"conhecimentos demais numa área", areas(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 3, ConhecimentoIDs: muitosConhecimentos}), nil, "areas_atuacao.conhecimento_ids"},
		{"comportamentos demais", nil, &muitosComportamentos, "comportamento_atitudes_ids"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := novoServicoHabilidades()
			repo.catalogoError = errors.New("o catálogo não deveria ser consultado")
			_, err := svc.Create(context.Background(), vagaNovaComHabilidades(tt.areas, tt.comportamentos))
			requireFieldError(t, err, tt.campo, "no máximo")
			assert.Empty(t, repo.vagas)
		})
	}
}

func TestVagaService_Create_ErroDoCatalogoPropaga(t *testing.T) {
	svc, repo := novoServicoHabilidades()
	repo.catalogoError = errors.New("db fora")

	_, err := svc.Create(context.Background(), vagaNovaComHabilidades(areas(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 3}), nil))
	require.Error(t, err)
	assert.Empty(t, repo.vagas)
}

// ──────────────────────────────────────────────────────────────────────────────
// Editar
// ──────────────────────────────────────────────────────────────────────────────

func vagaGravada(repo *MockVagaRepoForService, status empregabilidade.StatusVaga) uuid.UUID {
	id := uuid.New()
	repo.vagas[id] = &empregabilidade.Vaga{ID: id, Titulo: "Analista", IDContratante: "12345678000190", Status: status}
	return id
}

func TestVagaService_Update_ListasEnviadasSubstituemOConjunto(t *testing.T) {
	svc, repo := novoServicoHabilidades()
	id := vagaGravada(repo, empregabilidade.StatusVagaEmEdicao)
	seedHabilidadesCompletas(repo, id)

	err := svc.Update(context.Background(), &empregabilidade.Vaga{
		ID:                       id,
		Titulo:                   "Analista",
		AreasAtuacao:             areas(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 4, ConhecimentoIDs: []int64{20}}),
		ComportamentoAtitudesIDs: ids(8),
	})
	require.NoError(t, err)

	arg := repo.updateWithAssocArg
	require.NotNil(t, arg.AreasAtuacao)
	assert.Equal(t, []empregabilidade.VagaAreaAtuacao{{IDAreaAtuacao: 4, ConhecimentoIDs: []int64{20}}}, *arg.AreasAtuacao)
	assert.Equal(t, []int64{8}, *arg.ComportamentoAtitudesIDs)

	gravado, err := repo.GetHabilidadesCompetencias(context.Background(), id)
	require.NoError(t, err)
	require.Len(t, gravado.AreasAtuacao, 1)
	assert.Equal(t, int64(4), gravado.AreasAtuacao[0].IDAreaAtuacao)
}

// Sem as listas no corpo, o PUT não mexe no que está gravado.
func TestVagaService_Update_SemListasMantemOGravado(t *testing.T) {
	svc, repo := novoServicoHabilidades()
	id := vagaGravada(repo, empregabilidade.StatusVagaEmEdicao)

	require.NoError(t, svc.Update(context.Background(), &empregabilidade.Vaga{ID: id, Titulo: "Analista"}))
	assert.Nil(t, repo.updateWithAssocArg.AreasAtuacao)
	assert.Nil(t, repo.updateWithAssocArg.ComportamentoAtitudesIDs)
}

func TestVagaService_Update_ConhecimentoDeOutraAreaNaoGrava(t *testing.T) {
	svc, repo := novoServicoHabilidades()
	id := vagaGravada(repo, empregabilidade.StatusVagaEmEdicao)

	err := svc.Update(context.Background(), &empregabilidade.Vaga{
		ID:           id,
		AreasAtuacao: areas(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 4, ConhecimentoIDs: []int64{10}}),
	})
	requireFieldError(t, err, "areas_atuacao.conhecimento_ids", "não pertence à área de atuação")
	assert.Nil(t, repo.updateWithAssocArg, "a vaga não pode ser gravada")
}

// ──────────────────────────────────────────────────────────────────────────────
// Ler
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaService_GetByIDGestao_DevolveAsListas(t *testing.T) {
	svc, repo := novoServicoHabilidades()
	id := vagaGravada(repo, empregabilidade.StatusVagaEmEdicao)
	seedHabilidadesCompletas(repo, id)

	vaga, err := svc.GetByIDGestao(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, vaga.AreasAtuacao)
	assert.Equal(t, int64(3), (*vaga.AreasAtuacao)[0].IDAreaAtuacao)
	assert.Equal(t, []int64{10}, (*vaga.AreasAtuacao)[0].ConhecimentoIDs)
	assert.Equal(t, []int64{7}, *vaga.ComportamentoAtitudesIDs)
}

func TestVagaService_GetByIDGestao_VagaAntigaLeVazio(t *testing.T) {
	svc, repo := novoServicoHabilidades()
	id := vagaGravada(repo, empregabilidade.StatusVagaPublicadoAtivo)

	vaga, err := svc.GetByIDGestao(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, vaga.AreasAtuacao)
	assert.Empty(t, *vaga.AreasAtuacao)
	assert.Empty(t, *vaga.ComportamentoAtitudesIDs)
	assert.Empty(t, *vaga.ComportamentosAtitudes)
}

func TestVagaService_GetByIDGestao_InexistenteDevolveNil(t *testing.T) {
	svc, _ := novoServicoHabilidades()
	vaga, err := svc.GetByIDGestao(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Nil(t, vaga)
}

func TestVagaService_GetByIDGestao_ErroDoBancoPropaga(t *testing.T) {
	svc, repo := novoServicoHabilidades()
	id := vagaGravada(repo, empregabilidade.StatusVagaEmEdicao)
	repo.habilidadesError = errors.New("db fora")

	_, err := svc.GetByIDGestao(context.Background(), id)
	require.Error(t, err)
}

// A leitura usada pelas rotas públicas não carrega os campos.
func TestVagaService_GetByID_NaoCarregaHabilidades(t *testing.T) {
	svc, repo := novoServicoHabilidades()
	id := vagaGravada(repo, empregabilidade.StatusVagaPublicadoAtivo)
	seedHabilidadesCompletas(repo, id)

	vaga, err := svc.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Nil(t, vaga.AreasAtuacao)
	assert.Nil(t, vaga.ComportamentoAtitudesIDs)
	assert.Nil(t, vaga.ComportamentosAtitudes)
}

// ──────────────────────────────────────────────────────────────────────────────
// Publicar e enviar para aprovação
// ──────────────────────────────────────────────────────────────────────────────

type transicao struct {
	name   string
	status empregabilidade.StatusVaga
	run    func(*services.VagaService, uuid.UUID) error
}

func transicoesQueExigemHabilidades() []transicao {
	return []transicao{
		{"publicar", empregabilidade.StatusVagaEmEdicao, func(s *services.VagaService, id uuid.UUID) error { return s.Publish(context.Background(), id) }},
		{"publicar da aprovação", empregabilidade.StatusVagaEmAprovacao, func(s *services.VagaService, id uuid.UUID) error { return s.Publish(context.Background(), id) }},
		{"enviar para aprovação", empregabilidade.StatusVagaEmEdicao, func(s *services.VagaService, id uuid.UUID) error {
			return s.SendToApproval(context.Background(), id)
		}},
	}
}

func TestVagaService_TransicoesSemHabilidades_ErroPorCampo(t *testing.T) {
	for _, tr := range transicoesQueExigemHabilidades() {
		t.Run(tr.name, func(t *testing.T) {
			svc, repo := novoServicoHabilidades()
			id := vagaGravada(repo, tr.status)

			err := tr.run(svc, id)
			requireFieldError(t, err, "areas_atuacao", "Informe ao menos uma área de atuação")
			requireFieldError(t, err, "comportamento_atitudes_ids", "Informe ao menos um comportamento ou atitude")
			assert.Equal(t, tr.status, repo.vagas[id].Status, "o status não pode mudar")
		})
	}
}

func TestVagaService_TransicoesComAreaSemConhecimento_ErroNoCampoDosConhecimentos(t *testing.T) {
	for _, tr := range transicoesQueExigemHabilidades() {
		t.Run(tr.name, func(t *testing.T) {
			svc, repo := novoServicoHabilidades()
			id := vagaGravada(repo, tr.status)
			repo.habilidades[id] = &empregabilidade.VagaHabilidadesCompetencias{
				AreasAtuacao: []empregabilidade.VagaAreaAtuacao{
					{IDAreaAtuacao: 3, Nome: "Tecnologia", Conhecimentos: []empregabilidade.CatalogoItem{{ID: 10, Nome: "Go"}}},
					{IDAreaAtuacao: 4, Nome: "Saúde"},
				},
				ComportamentosAtitudes: []empregabilidade.CatalogoItem{{ID: 7, Nome: "Proatividade"}},
			}

			err := tr.run(svc, id)
			requireFieldError(t, err, "areas_atuacao.conhecimento_ids", "Saúde")
			appErr, _ := apperrors.As(err)
			assert.NotContains(t, appErr.Fields, "areas_atuacao")
			assert.NotContains(t, appErr.Fields, "comportamento_atitudes_ids")
			assert.Equal(t, tr.status, repo.vagas[id].Status)
		})
	}
}

func TestVagaService_TransicoesComHabilidades_Seguem(t *testing.T) {
	for _, tr := range transicoesQueExigemHabilidades() {
		t.Run(tr.name, func(t *testing.T) {
			svc, repo := novoServicoHabilidades()
			id := vagaGravada(repo, tr.status)
			seedHabilidadesCompletas(repo, id)

			require.NoError(t, tr.run(svc, id))
			assert.NotEqual(t, tr.status, repo.vagas[id].Status)
		})
	}
}

func TestVagaService_TransicoesErroAoLerHabilidades_NaoMudaStatus(t *testing.T) {
	for _, tr := range transicoesQueExigemHabilidades() {
		t.Run(tr.name, func(t *testing.T) {
			svc, repo := novoServicoHabilidades()
			id := vagaGravada(repo, tr.status)
			repo.habilidadesError = errors.New("db fora")

			require.Error(t, tr.run(svc, id))
			assert.Equal(t, tr.status, repo.vagas[id].Status)
		})
	}
}
