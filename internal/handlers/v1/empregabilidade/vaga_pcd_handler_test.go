package empregabilidade_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tipoPCDJSON = `[{"id":"6f1c2d3e-4b5a-4c6d-8e7f-9a0b1c2d3e4f"}]`

func campoDaResposta(t *testing.T, w *httptest.ResponseRecorder, campo string) (json.RawMessage, bool) {
	t.Helper()
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	raw, ok := body[campo]
	return raw, ok
}

func TestVagaHandler_Create_OpcaoPCD(t *testing.T) {
	tests := []struct {
		name               string
		body               string
		wantCode           int
		wantOpcao          string
		wantAcessibilidade string
		wantErro           string
	}{
		{name: "sem opção é aberta", body: `{"titulo":"Dev"}`, wantCode: http.StatusCreated, wantOpcao: `"aberta"`},
		{name: "preferencial", body: `{"titulo":"Dev","opcao_pcd":"preferencial"}`, wantCode: http.StatusCreated, wantOpcao: `"preferencial"`, wantAcessibilidade: `"preferencial_pcd"`},
		{name: "exclusiva com tipo", body: `{"titulo":"Dev","opcao_pcd":"exclusiva","tipos_pcd":` + tipoPCDJSON + `}`, wantCode: http.StatusCreated, wantOpcao: `"exclusiva"`, wantAcessibilidade: `"exclusivo_pcd"`},
		{name: "exclusiva sem tipo: 400", body: `{"titulo":"Dev","opcao_pcd":"exclusiva"}`, wantCode: http.StatusBadRequest, wantErro: empmodels.MsgTipoPCDObrigatorio},
		{name: "exclusiva com tipos vazios: 400", body: `{"titulo":"Dev","opcao_pcd":"exclusiva","tipos_pcd":[]}`, wantCode: http.StatusBadRequest, wantErro: empmodels.MsgTipoPCDObrigatorio},
		{name: "opção desconhecida: 400", body: `{"titulo":"Dev","opcao_pcd":"exclusivo_pcd"}`, wantCode: http.StatusBadRequest, wantErro: "opcao_pcd inválida"},
		{name: "opção com tipo errado: 400", body: `{"titulo":"Dev","opcao_pcd":true}`, wantCode: http.StatusBadRequest},
		{name: "cliente antigo com exclusivo_pcd e tipo", body: `{"titulo":"Dev","acessibilidade_pcd":"exclusivo_pcd","tipos_pcd":` + tipoPCDJSON + `}`, wantCode: http.StatusCreated, wantOpcao: `"exclusiva"`, wantAcessibilidade: `"exclusivo_pcd"`},
		{name: "bloqueio PcD ligado", body: `{"titulo":"Dev","opcao_pcd":"exclusiva","tipos_pcd":` + tipoPCDJSON + `,"criterios_bloqueio":["pcd"]}`, wantCode: http.StatusCreated, wantOpcao: `"exclusiva"`, wantAcessibilidade: `"exclusivo_pcd"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			empresaRepo := &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}
			r := setupVagaRouter(&mockVagaRepoH{}, empresaRepo, false)
			req := httptest.NewRequest(http.MethodPost, "/vagas", bodyOf(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, tt.wantCode, w.Code, w.Body.String())
			if tt.wantErro != "" {
				assert.Contains(t, w.Body.String(), tt.wantErro)
			}
			if tt.wantOpcao == "" {
				return
			}
			raw, ok := campoDaResposta(t, w, "opcao_pcd")
			require.True(t, ok, "opcao_pcd precisa vir na resposta")
			assert.JSONEq(t, tt.wantOpcao, string(raw))
			rawAcess, ok := campoDaResposta(t, w, "acessibilidade_pcd")
			if tt.wantAcessibilidade == "" {
				assert.False(t, ok, "aberta não tem acessibilidade_pcd")
				return
			}
			require.True(t, ok)
			assert.JSONEq(t, tt.wantAcessibilidade, string(rawAcess))
		})
	}
}

func TestVagaHandler_Create_CriterioBloqueioPCD(t *testing.T) {
	empresaRepo := &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}
	r := setupVagaRouter(&mockVagaRepoH{}, empresaRepo, false)
	body := `{"titulo":"Dev","opcao_pcd":"exclusiva","tipos_pcd":` + tipoPCDJSON + `,"criterios_bloqueio":["pcd","idioma"]}`
	req := httptest.NewRequest(http.MethodPost, "/vagas", bodyOf(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	raw, ok := campoDaResposta(t, w, "criterios_bloqueio")
	require.True(t, ok)
	assert.JSONEq(t, `["idioma","pcd"]`, string(raw))
}

func TestVagaHandler_Update_OpcaoPCD(t *testing.T) {
	id := uuid.MustParse(validUUID)
	tests := []struct {
		name     string
		atual    *empmodels.Vaga
		body     string
		wantCode int
	}{
		{
			name:     "exclusiva sem tipo em lugar nenhum: 400",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao, OpcaoPCD: empmodels.OpcaoPCDAberta},
			body:     `{"titulo":"Dev","opcao_pcd":"exclusiva"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "exclusiva com tipos gravados",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao, TiposPCD: []empmodels.TipoPCD{{ID: uuid.New()}}},
			body:     `{"titulo":"Dev","opcao_pcd":"exclusiva"}`,
			wantCode: http.StatusOK,
		},
		{
			name:     "opção desconhecida: 400",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao},
			body:     `{"titulo":"Dev","opcao_pcd":"todas"}`,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupVagaRouter(&mockVagaRepoH{entity: tt.atual}, &mockEmpresaRepoForVaga{}, false)
			req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, bodyOf(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code, w.Body.String())
			if tt.wantCode == http.StatusBadRequest && tt.atual.OpcaoPCD == empmodels.OpcaoPCDAberta {
				assert.Contains(t, w.Body.String(), empmodels.MsgTipoPCDObrigatorio)
			}
		})
	}
}

func TestVagaHandler_UpdateTiposPCD_ExclusivaSemTipo(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vaga := &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao, OpcaoPCD: empmodels.OpcaoPCDExclusiva}
	r := setupVagaRouter(&mockVagaRepoH{entity: vaga}, &mockEmpresaRepoForVaga{}, true)

	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID+"/tipos-pcd", bodyOf(`{"tipos_pcd_ids":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), empmodels.MsgTipoPCDObrigatorio)
}

func TestVagaHandler_GetByID_DevolveOpcaoPCD(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vaga := &empmodels.Vaga{
		ID:                id,
		Status:            empmodels.StatusVagaPublicadoAtivo,
		OpcaoPCD:          empmodels.OpcaoPCDExclusiva,
		TiposPCD:          []empmodels.TipoPCD{{ID: uuid.New(), Descricao: "Física"}},
		CriteriosBloqueio: []empmodels.CriterioBloqueio{empmodels.CriterioBloqueioPCD},
	}
	r := setupVagaRouter(&mockVagaRepoH{entity: vaga}, &mockEmpresaRepoForVaga{}, false)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	raw, ok := campoDaResposta(t, w, "opcao_pcd")
	require.True(t, ok)
	assert.JSONEq(t, `"exclusiva"`, string(raw))
	raw, ok = campoDaResposta(t, w, "criterios_bloqueio")
	require.True(t, ok)
	assert.JSONEq(t, `["pcd"]`, string(raw))
	raw, ok = campoDaResposta(t, w, "tipos_pcd")
	require.True(t, ok)
	assert.Contains(t, string(raw), "Física")
}

// O superapp registra a candidatura barrada pelo critério PcD com o slug "pcd"
// (o mesmo de criterios_bloqueio), e a listagem devolve o slug como gravado.
func TestCandidaturaBloqueioHandler_SlugPCD(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouter(repo)

	w := postBloqueio(r, `{"id_vaga":"`+validUUID+`","criterios_nao_atendidos":["pcd"]}`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	require.NotNil(t, repo.created)
	assert.Equal(t, []string{string(empmodels.CriterioBloqueioPCD)}, repo.created.CriteriosNaoAtendidos)

	raw, ok := campoDaResposta(t, w, "criterios_nao_atendidos")
	require.True(t, ok)
	assert.JSONEq(t, `["pcd"]`, string(raw))
}
