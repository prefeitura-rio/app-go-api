package empregabilidade_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// ──────────────────────────────────────────────────────────────────────────────
// Habilidades e competências no mockVagaRepoH
// ──────────────────────────────────────────────────────────────────────────────

func (m *mockVagaRepoH) GetHabilidadesCompetencias(_ context.Context, _ uuid.UUID) (*empmodels.VagaHabilidadesCompetencias, error) {
	if m.habilidades == nil {
		return &empmodels.VagaHabilidadesCompetencias{}, m.err
	}
	return m.habilidades, m.err
}

func (m *mockVagaRepoH) AreasAtuacaoExistentes(_ context.Context, ids []int64) ([]int64, error) {
	var out []int64
	for _, id := range ids {
		if _, ok := m.conhecimentosPorArea[id]; ok {
			out = append(out, id)
		}
	}
	return out, m.err
}

func (m *mockVagaRepoH) ConhecimentosDasAreas(_ context.Context, areaIDs, habilidadeIDs []int64) (map[int64]map[int64]bool, error) {
	out := map[int64]map[int64]bool{}
	for _, a := range areaIDs {
		for _, c := range m.conhecimentosPorArea[a] {
			for _, h := range habilidadeIDs {
				if c == h {
					if out[a] == nil {
						out[a] = map[int64]bool{}
					}
					out[a][h] = true
				}
			}
		}
	}
	return out, m.err
}

func (m *mockVagaRepoH) ComportamentosExistentes(_ context.Context, ids []int64) ([]int64, error) {
	return ids, m.err
}

func habilidadesGravadas() *empmodels.VagaHabilidadesCompetencias {
	return &empmodels.VagaHabilidadesCompetencias{
		AreasAtuacao: []empmodels.VagaAreaAtuacao{
			{IDAreaAtuacao: 3, Nome: "Tecnologia", Conhecimentos: []empmodels.CatalogoItem{{ID: 10, Nome: "Go"}, {ID: 11, Nome: "SQL"}}},
			{IDAreaAtuacao: 4, Nome: "Saúde", Conhecimentos: []empmodels.CatalogoItem{{ID: 20, Nome: "Primeiros socorros"}}},
		},
		ComportamentosAtitudes: []empmodels.CatalogoItem{{ID: 7, Nome: "Proatividade"}},
	}
}

func doJSON(t *testing.T, r http.Handler, method, path, body string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w, out
}

// ──────────────────────────────────────────────────────────────────────────────
// Ler
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_GetByID_DevolveAreasConhecimentosEComportamentos(t *testing.T) {
	id := uuid.New()
	repo := &mockVagaRepoH{
		entity:      &empmodels.Vaga{ID: id, Titulo: "Analista", Status: empmodels.StatusVagaEmEdicao},
		habilidades: habilidadesGravadas(),
	}
	r := setupVagaRouter(repo, &mockEmpresaRepoForVaga{}, true)

	w, body := doJSON(t, r, http.MethodGet, "/vagas/"+id.String(), "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	areasResp := body["areas_atuacao"].([]any)
	require.Len(t, areasResp, 2)
	primeira := areasResp[0].(map[string]any)
	assert.Equal(t, float64(3), primeira["id_area_atuacao"])
	assert.Equal(t, "Tecnologia", primeira["nome"])
	assert.Equal(t, []any{float64(10), float64(11)}, primeira["conhecimento_ids"])
	assert.Len(t, primeira["conhecimentos"], 2)
	assert.Equal(t, []any{float64(7)}, body["comportamento_atitudes_ids"])
	assert.Equal(t, "Proatividade", body["comportamentos_atitudes"].([]any)[0].(map[string]any)["nome"])
}

func TestVagaHandler_GetByID_VagaAntigaLeListasVazias(t *testing.T) {
	id := uuid.New()
	repo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Titulo: "Antiga", Status: empmodels.StatusVagaPublicadoAtivo}}
	r := setupVagaRouter(repo, &mockEmpresaRepoForVaga{}, true)

	w, body := doJSON(t, r, http.MethodGet, "/vagas/"+id.String(), "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, []any{}, body["areas_atuacao"])
	assert.Equal(t, []any{}, body["comportamento_atitudes_ids"])
	assert.Equal(t, []any{}, body["comportamentos_atitudes"])
}

// Rotas públicas (consumidas pelo SuperApp) não devolvem os campos.
func TestVagaHandler_RotasPublicas_NaoDevolvemHabilidades(t *testing.T) {
	id := uuid.New()
	vaga := &empmodels.Vaga{ID: id, Titulo: "Analista", Status: empmodels.StatusVagaPublicadoAtivo}
	vaga.Slug = empmodels.BuildVagaSlug(vaga.Titulo, id)
	repo := &mockVagaRepoH{
		entity:      vaga,
		listItems:   []*empmodels.Vaga{vaga},
		listTotal:   1,
		habilidades: habilidadesGravadas(),
	}
	r := setupVagaRouter(repo, &mockEmpresaRepoForVaga{}, false)

	for _, path := range []string{"/public/vagas/" + id.String(), "/public/vagas/slug/" + vaga.Slug, "/public/vagas", "/vagas"} {
		w, _ := doJSON(t, r, http.MethodGet, path, "")
		require.Equal(t, http.StatusOK, w.Code, "%s: %s", path, w.Body.String())
		for _, campo := range []string{"areas_atuacao", "comportamento_atitudes_ids", "comportamentos_atitudes", "conhecimentos"} {
			assert.NotContains(t, w.Body.String(), `"`+campo+`"`, "%s não deveria trazer %s", path, campo)
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Criar, duplicar e editar
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_Create_GravaEDevolveHabilidades(t *testing.T) {
	repo := &mockVagaRepoH{conhecimentosPorArea: map[int64][]int64{3: {10, 11}}}
	r := setupVagaRouter(repo, &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}, true)

	w, body := doJSON(t, r, http.MethodPost, "/vagas", `{"titulo":"Analista","id_contratante":"12345678000100",
		"areas_atuacao":[{"id_area_atuacao":3,"conhecimento_ids":[10,11]}],"comportamento_atitudes_ids":[7]}`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	require.NotNil(t, repo.created.AreasAtuacao)
	assert.Equal(t, []empmodels.VagaAreaAtuacao{{IDAreaAtuacao: 3, ConhecimentoIDs: []int64{10, 11}}}, *repo.created.AreasAtuacao)
	assert.Equal(t, []int64{7}, *repo.created.ComportamentoAtitudesIDs)
	assert.Contains(t, body, "areas_atuacao")
	assert.Contains(t, body, "comportamento_atitudes_ids")
}

func TestVagaHandler_Create_AreaRepetidaOuConhecimentoDeOutraArea_400PorCampo(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		campo string
	}{
		{"área repetida", `{"titulo":"A","id_contratante":"12345678000100","areas_atuacao":[{"id_area_atuacao":3},{"id_area_atuacao":3}]}`, "areas_atuacao"},
		{"conhecimento de outra área", `{"titulo":"A","id_contratante":"12345678000100","areas_atuacao":[{"id_area_atuacao":3,"conhecimento_ids":[20]}]}`, "areas_atuacao.conhecimento_ids"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVagaRepoH{conhecimentosPorArea: map[int64][]int64{3: {10}, 4: {20}}}
			r := setupVagaRouter(repo, &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}, true)

			w, body := doJSON(t, r, http.MethodPost, "/vagas", tt.body)
			require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			assert.NotEmpty(t, body["error"])
			fields, ok := body["fields"].(map[string]any)
			require.True(t, ok, "esperava fields: %s", w.Body.String())
			assert.NotEmpty(t, fields[tt.campo])
			assert.Nil(t, repo.created, "a vaga não pode ser criada")
		})
	}
}

// Duplicar = ler a vaga e enviar ao POST o corpo sem o id da vaga (como faz o
// botão de duplicar do portal-interno): as listas que o GET devolve, com os
// nomes, são aceitas de volta e gravam as mesmas áreas, conhecimentos e
// comportamentos.
func TestVagaHandler_Duplicar_GetDepoisPostGravaOMesmoConjunto(t *testing.T) {
	id := uuid.New()
	repo := &mockVagaRepoH{
		entity:               &empmodels.Vaga{ID: id, Titulo: "Analista", IDContratante: "12345678000100", Status: empmodels.StatusVagaPublicadoAtivo},
		habilidades:          habilidadesGravadas(),
		conhecimentosPorArea: map[int64][]int64{3: {10, 11}, 4: {20}},
	}
	r := setupVagaRouter(repo, &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}, true)

	w, lida := doJSON(t, r, http.MethodGet, "/vagas/"+id.String(), "")
	require.Equal(t, http.StatusOK, w.Code)
	delete(lida, "id")
	corpo, err := json.Marshal(lida)
	require.NoError(t, err)

	w2, _ := doJSON(t, r, http.MethodPost, "/vagas", string(corpo))
	require.Equal(t, http.StatusCreated, w2.Code, w2.Body.String())
	require.NotNil(t, repo.created.AreasAtuacao)
	assert.Equal(t, []empmodels.VagaAreaAtuacao{
		{IDAreaAtuacao: 3, ConhecimentoIDs: []int64{10, 11}},
		{IDAreaAtuacao: 4, ConhecimentoIDs: []int64{20}},
	}, *repo.created.AreasAtuacao)
	assert.Equal(t, []int64{7}, *repo.created.ComportamentoAtitudesIDs)
}

func TestVagaHandler_Update_DevolveHabilidadesGravadas(t *testing.T) {
	id := uuid.New()
	repo := &mockVagaRepoH{
		entity:               &empmodels.Vaga{ID: id, Titulo: "Analista", Status: empmodels.StatusVagaEmEdicao},
		habilidades:          habilidadesGravadas(),
		conhecimentosPorArea: map[int64][]int64{3: {10, 11}, 4: {20}},
	}
	r := setupVagaRouter(repo, &mockEmpresaRepoForVaga{}, true)

	w, body := doJSON(t, r, http.MethodPut, "/vagas/"+id.String(), `{"titulo":"Analista","areas_atuacao":[{"id_area_atuacao":4,"conhecimento_ids":[20]}]}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Len(t, body["areas_atuacao"], 2, "a resposta é a leitura do que está gravado")
}

// ──────────────────────────────────────────────────────────────────────────────
// Publicar e enviar para aprovação
// ──────────────────────────────────────────────────────────────────────────────

func TestVagaHandler_PublicarSemHabilidades_400ComErroPorCampo(t *testing.T) {
	for _, rota := range []string{"publish", "send-to-approval"} {
		t.Run(rota, func(t *testing.T) {
			id := uuid.New()
			repo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Titulo: "Analista", Status: empmodels.StatusVagaEmEdicao}}
			r := setupVagaRouter(repo, &mockEmpresaRepoForVaga{}, true)

			w, body := doJSON(t, r, http.MethodPut, "/vagas/"+id.String()+"/"+rota, "")
			require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
			fields, ok := body["fields"].(map[string]any)
			require.True(t, ok, w.Body.String())
			assert.Contains(t, fields["areas_atuacao"], "Informe ao menos uma área de atuação")
			assert.Contains(t, fields["comportamento_atitudes_ids"], "Informe ao menos um comportamento ou atitude")
		})
	}
}

func TestVagaHandler_PublicarComHabilidades_200(t *testing.T) {
	for _, rota := range []string{"publish", "send-to-approval"} {
		t.Run(rota, func(t *testing.T) {
			id := uuid.New()
			repo := &mockVagaRepoH{
				entity:      &empmodels.Vaga{ID: id, Titulo: "Analista", Status: empmodels.StatusVagaEmEdicao},
				habilidades: habilidadesGravadas(),
			}
			r := setupVagaRouter(repo, &mockEmpresaRepoForVaga{}, true)

			w, _ := doJSON(t, r, http.MethodPut, "/vagas/"+id.String()+"/"+rota, "")
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		})
	}
}
