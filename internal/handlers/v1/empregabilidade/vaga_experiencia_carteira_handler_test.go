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

func TestVagaHandler_Create_ExperienciaCarteira(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		wantCode      int
		wantCampo     string
		wantCriterios string
		wantErro      string
	}{
		{name: "sem exigência devolve null", body: `{"titulo":"Dev"}`, wantCode: http.StatusCreated, wantCampo: `null`, wantCriterios: `[]`},
		{
			name:          "bloqueio ligado com mínimo de 6 meses",
			body:          `{"titulo":"Dev","experiencia_carteira":"minimo_6_meses","criterios_bloqueio":["experiencia_carteira"]}`,
			wantCode:      http.StatusCreated,
			wantCampo:     `"minimo_6_meses"`,
			wantCriterios: `["experiencia_carteira"]`,
		},
		{
			name:          "bloqueio ligado com qualquer tempo",
			body:          `{"titulo":"Dev","experiencia_carteira":"qualquer_tempo","criterios_bloqueio":["idioma","experiencia_carteira"]}`,
			wantCode:      http.StatusCreated,
			wantCampo:     `"qualquer_tempo"`,
			wantCriterios: `["idioma","experiencia_carteira"]`,
		},
		{
			name:     "bloqueio ligado sem opção: 400",
			body:     `{"titulo":"Dev","criterios_bloqueio":["experiencia_carteira"]}`,
			wantCode: http.StatusBadRequest,
			wantErro: empmodels.MsgExperienciaCarteiraObrigatoria,
		},
		{
			name:     "opção desconhecida: 400",
			body:     `{"titulo":"Dev","experiencia_carteira":"12_meses"}`,
			wantCode: http.StatusBadRequest,
			wantErro: "experiencia_carteira inválida",
		},
		{name: "opção com tipo errado: 400", body: `{"titulo":"Dev","experiencia_carteira":6}`, wantCode: http.StatusBadRequest},
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
			raw, ok := campoDaResposta(t, w, "experiencia_carteira")
			require.True(t, ok, "experiencia_carteira precisa vir na resposta")
			assert.JSONEq(t, tt.wantCampo, string(raw))
			raw, ok = campoDaResposta(t, w, "criterios_bloqueio")
			require.True(t, ok)
			assert.JSONEq(t, tt.wantCriterios, string(raw))
		})
	}
}

func TestVagaHandler_Update_ExperienciaCarteira(t *testing.T) {
	id := uuid.MustParse(validUUID)
	seisMeses := empmodels.ExperienciaCarteiraMinimo6Meses
	tests := []struct {
		name     string
		atual    *empmodels.Vaga
		body     string
		wantCode int
		wantErro string
	}{
		{
			name:     "liga o bloqueio sem opção em lugar nenhum: 400",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao},
			body:     `{"titulo":"Dev","criterios_bloqueio":["experiencia_carteira"]}`,
			wantCode: http.StatusBadRequest,
			wantErro: empmodels.MsgExperienciaCarteiraObrigatoria,
		},
		{
			name:     "liga o bloqueio com a opção gravada",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao, ExperienciaCarteira: &seisMeses},
			body:     `{"titulo":"Dev","criterios_bloqueio":["experiencia_carteira"]}`,
			wantCode: http.StatusOK,
		},
		{
			name:     "cliente sem o campo mantém a opção do bloqueio ligado",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao, ExperienciaCarteira: &seisMeses, CriteriosBloqueio: []empmodels.CriterioBloqueio{empmodels.CriterioBloqueioExperienciaCarteira}},
			body:     `{"titulo":"Dev"}`,
			wantCode: http.StatusOK,
		},
		{
			name:     "remove a opção com o bloqueio ligado: 400",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao, ExperienciaCarteira: &seisMeses, CriteriosBloqueio: []empmodels.CriterioBloqueio{empmodels.CriterioBloqueioExperienciaCarteira}},
			body:     `{"titulo":"Dev","experiencia_carteira":""}`,
			wantCode: http.StatusBadRequest,
			wantErro: empmodels.MsgExperienciaCarteiraObrigatoria,
		},
		{
			name:     "opção desconhecida: 400",
			atual:    &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao},
			body:     `{"titulo":"Dev","experiencia_carteira":"sempre"}`,
			wantCode: http.StatusBadRequest,
			wantErro: "experiencia_carteira inválida",
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

func TestVagaHandler_GetByID_DevolveExperienciaCarteira(t *testing.T) {
	id := uuid.MustParse(validUUID)
	seisMeses := empmodels.ExperienciaCarteiraMinimo6Meses
	vaga := &empmodels.Vaga{
		ID:                  id,
		Status:              empmodels.StatusVagaPublicadoAtivo,
		ExperienciaCarteira: &seisMeses,
		CriteriosBloqueio:   []empmodels.CriterioBloqueio{empmodels.CriterioBloqueioExperienciaCarteira},
	}
	r := setupVagaRouter(&mockVagaRepoH{entity: vaga}, &mockEmpresaRepoForVaga{}, false)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	raw, ok := campoDaResposta(t, w, "experiencia_carteira")
	require.True(t, ok)
	assert.JSONEq(t, `"minimo_6_meses"`, string(raw))
	raw, ok = campoDaResposta(t, w, "criterios_bloqueio")
	require.True(t, ok)
	assert.JSONEq(t, `["experiencia_carteira"]`, string(raw))
}

// O superapp registra a candidatura barrada pela experiência em carteira com o
// slug "experiencia_carteira" (o mesmo de criterios_bloqueio), e a listagem
// devolve o slug como gravado.
func TestCandidaturaBloqueioHandler_SlugExperienciaCarteira(t *testing.T) {
	repo := &mockCandidaturaBloqueioRepoH{}
	r := setupCandidaturaBloqueioRouter(repo)

	w := postBloqueio(r, `{"id_vaga":"`+validUUID+`","criterios_nao_atendidos":["escolaridade","experiencia_carteira"]}`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	require.NotNil(t, repo.created)
	assert.Equal(t, []string{"escolaridade", string(empmodels.CriterioBloqueioExperienciaCarteira)}, repo.created.CriteriosNaoAtendidos)

	raw, ok := campoDaResposta(t, w, "criterios_nao_atendidos")
	require.True(t, ok)
	assert.JSONEq(t, `["escolaridade","experiencia_carteira"]`, string(raw))
}
