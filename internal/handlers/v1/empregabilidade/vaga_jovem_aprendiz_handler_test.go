package empregabilidade_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVagaHandler_Create_JovemAprendiz(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		wantCode      int
		wantCampo     string
		wantCriterios string
		wantIdade     string
		wantErro      string
	}{
		{name: "sem o campo nasce desligado", body: `{"titulo":"Dev"}`, wantCode: http.StatusCreated, wantCampo: `false`, wantCriterios: `[]`},
		{
			name:          "Jovem Aprendiz com bloqueio ligado",
			body:          `{"titulo":"Dev","jovem_aprendiz":true,"criterios_bloqueio":["jovem_aprendiz"]}`,
			wantCode:      http.StatusCreated,
			wantCampo:     `true`,
			wantCriterios: `["jovem_aprendiz"]`,
		},
		{
			name:          "Maiores de 18 continua em idade_minima",
			body:          `{"titulo":"Dev","idade_minima":18}`,
			wantCode:      http.StatusCreated,
			wantCampo:     `false`,
			wantCriterios: `[]`,
			wantIdade:     `18`,
		},
		{
			name:     "bloqueio ligado sem a opção: 400",
			body:     `{"titulo":"Dev","criterios_bloqueio":["jovem_aprendiz"]}`,
			wantCode: http.StatusBadRequest,
			wantErro: empmodels.MsgJovemAprendizObrigatorio,
		},
		{name: "número livre no lugar da opção: 400", body: `{"titulo":"Dev","jovem_aprendiz":14}`, wantCode: http.StatusBadRequest},
		{name: "texto no lugar da opção: 400", body: `{"titulo":"Dev","jovem_aprendiz":"sim"}`, wantCode: http.StatusBadRequest},
		{
			name:     "idade_minima como chave de bloqueio: 400",
			body:     `{"titulo":"Dev","idade_minima":18,"criterios_bloqueio":["idade_minima"]}`,
			wantCode: http.StatusBadRequest,
			wantErro: "idade_minima",
		},
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
			if tt.wantCampo == "" {
				return
			}
			raw, ok := campoDaResposta(t, w, "jovem_aprendiz")
			require.True(t, ok, "jovem_aprendiz precisa vir na resposta")
			assert.JSONEq(t, tt.wantCampo, string(raw))
			raw, ok = campoDaResposta(t, w, "criterios_bloqueio")
			require.True(t, ok)
			assert.JSONEq(t, tt.wantCriterios, string(raw))
			if tt.wantIdade != "" {
				raw, ok = campoDaResposta(t, w, "idade_minima")
				require.True(t, ok)
				assert.JSONEq(t, tt.wantIdade, string(raw))
			}
		})
	}
}

func TestVagaHandler_Update_JovemAprendiz(t *testing.T) {
	id := uuid.MustParse(validUUID)
	marcado, desmarcado := true, false
	ligado := []empmodels.CriterioBloqueio{empmodels.CriterioBloqueioJovemAprendiz}
	tests := []struct {
		name     string
		atual    *empmodels.Vaga
		body     string
		wantCode int
		wantErro string
	}{
		{
			name:     "liga o bloqueio sem a opção: 400",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao, JovemAprendiz: &desmarcado},
			body:     `{"titulo":"Dev","criterios_bloqueio":["jovem_aprendiz"]}`,
			wantCode: http.StatusBadRequest,
			wantErro: empmodels.MsgJovemAprendizObrigatorio,
		},
		{
			name:     "liga o bloqueio com a opção gravada",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao, JovemAprendiz: &marcado},
			body:     `{"titulo":"Dev","criterios_bloqueio":["jovem_aprendiz"]}`,
			wantCode: http.StatusOK,
		},
		{
			name:     "cliente sem o campo mantém a opção do bloqueio ligado",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao, JovemAprendiz: &marcado, CriteriosBloqueio: ligado},
			body:     `{"titulo":"Dev"}`,
			wantCode: http.StatusOK,
		},
		{
			name:     "desmarca a opção com o bloqueio ligado: 400",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao, JovemAprendiz: &marcado, CriteriosBloqueio: ligado},
			body:     `{"titulo":"Dev","jovem_aprendiz":false}`,
			wantCode: http.StatusBadRequest,
			wantErro: empmodels.MsgJovemAprendizObrigatorio,
		},
		{
			name:     "número livre: 400",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao},
			body:     `{"titulo":"Dev","jovem_aprendiz":24}`,
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
			if tt.wantErro != "" {
				assert.Contains(t, w.Body.String(), tt.wantErro)
			}
		})
	}
}

func TestVagaHandler_GetByID_DevolveJovemAprendiz(t *testing.T) {
	id := uuid.MustParse(validUUID)
	marcado := true
	vaga := &empmodels.Vaga{
		ID:                id,
		Status:            empmodels.StatusVagaPublicadoAtivo,
		JovemAprendiz:     &marcado,
		CriteriosBloqueio: []empmodels.CriterioBloqueio{empmodels.CriterioBloqueioJovemAprendiz},
	}
	r := setupVagaRouter(&mockVagaRepoH{entity: vaga}, &mockEmpresaRepoForVaga{}, false)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	raw, ok := campoDaResposta(t, w, "jovem_aprendiz")
	require.True(t, ok)
	assert.JSONEq(t, `true`, string(raw))
	raw, ok = campoDaResposta(t, w, "criterios_bloqueio")
	require.True(t, ok)
	assert.JSONEq(t, `["jovem_aprendiz"]`, string(raw))
}

// O superapp registra a candidatura barrada pela faixa de Jovem Aprendiz com o
// slug "jovem_aprendiz" (o mesmo de criterios_bloqueio); "Maiores de 18"
// continua com "idade_minima". A listagem devolve os slugs como gravados.
func TestCandidaturaBloqueioHandler_SlugJovemAprendiz(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouter(repo)

	w := postBloqueio(r, `{"id_vaga":"`+validUUID+`","criterios_nao_atendidos":["idade_minima","jovem_aprendiz"]}`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	require.NotNil(t, repo.created)
	assert.Equal(t, []string{"idade_minima", string(empmodels.CriterioBloqueioJovemAprendiz)}, repo.created.CriteriosNaoAtendidos)

	raw, ok := campoDaResposta(t, w, "criterios_nao_atendidos")
	require.True(t, ok)
	assert.JSONEq(t, `["idade_minima","jovem_aprendiz"]`, string(raw))
}
