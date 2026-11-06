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

func criteriosDaResposta(t *testing.T, w *httptest.ResponseRecorder) (json.RawMessage, bool) {
	t.Helper()
	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	raw, ok := body["criterios_bloqueio"]
	return raw, ok
}

func TestVagaHandler_Create_CriteriosBloqueio(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode int
		wantJSON string
	}{
		{
			name:     "sem o campo: tudo desligado, devolve []",
			body:     `{"titulo":"Dev"}`,
			wantCode: http.StatusCreated,
			wantJSON: `[]`,
		},
		{
			name:     "critérios ligados normalizados",
			body:     `{"titulo":"Dev","criterios_bloqueio":["idioma","escolaridade","idioma"]}`,
			wantCode: http.StatusCreated,
			wantJSON: `["escolaridade","idioma"]`,
		},
		{
			name:     "chave desconhecida: 400",
			body:     `{"titulo":"Dev","criterios_bloqueio":["altura"]}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "tipo errado: 400",
			body:     `{"titulo":"Dev","criterios_bloqueio":{"idioma":true}}`,
			wantCode: http.StatusBadRequest,
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
			if tt.wantJSON == "" {
				return
			}
			raw, ok := criteriosDaResposta(t, w)
			require.True(t, ok, "criterios_bloqueio precisa vir na resposta")
			assert.JSONEq(t, tt.wantJSON, string(raw))
		})
	}
}

func TestVagaHandler_Update_CriteriosBloqueio_ChaveDesconhecida(t *testing.T) {
	id := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaEmEdicao}}
	r := setupVagaRouter(vagaRepo, &mockEmpresaRepoForVaga{}, false)

	req := httptest.NewRequest(http.MethodPut, "/vagas/"+validUUID, bodyOf(`{"titulo":"Dev","criterios_bloqueio":["altura"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func TestVagaHandler_GetByID_DevolveCriteriosBloqueio(t *testing.T) {
	id := uuid.MustParse(validUUID)

	t.Run("vaga com critérios ligados", func(t *testing.T) {
		vaga := &empmodels.Vaga{
			ID:                id,
			Status:            empmodels.StatusVagaEmEdicao,
			CriteriosBloqueio: []empmodels.CriterioBloqueio{empmodels.CriterioBloqueioEscolaridade},
		}
		r := setupVagaRouter(&mockVagaRepoH{entity: vaga}, &mockEmpresaRepoForVaga{}, false)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/vagas/"+validUUID, nil))

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		raw, ok := criteriosDaResposta(t, w)
		require.True(t, ok)
		assert.JSONEq(t, `["escolaridade"]`, string(raw))
	})

	t.Run("vaga lida do banco sem critérios devolve []", func(t *testing.T) {
		vaga := &empmodels.Vaga{ID: id, Status: empmodels.StatusVagaPublicadoAtivo}
		require.NoError(t, vaga.AfterFind(nil)) // o que o GORM faz ao ler a vaga
		r := setupVagaRouter(&mockVagaRepoH{entity: vaga}, &mockEmpresaRepoForVaga{}, false)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/public/vagas/"+validUUID, nil))

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		raw, ok := criteriosDaResposta(t, w)
		require.True(t, ok)
		assert.JSONEq(t, `[]`, string(raw))
	})
}
