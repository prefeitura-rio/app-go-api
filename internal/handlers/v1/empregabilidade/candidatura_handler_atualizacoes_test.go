package empregabilidade_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCandidaturaHandler_ListByCPF_DevolveAtualizacoesNaoVistas(t *testing.T) {
	candRepo := &mockCandidaturaRepoH{listItems: []*empmodels.Candidatura{}, atualizacoes: 2}
	r := setupCandidaturaRouter(candRepo, &mockVagaForCandidaturaH{}, "12345678900", false)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/candidaturas/usuario/12345678900", nil))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(2), body["atualizacoes"])
}

func TestCandidaturaHandler_MarcarAtualizacoesVistas(t *testing.T) {
	tests := []struct {
		name        string
		userCPF     string
		isAdmin     bool
		pathCPF     string
		repoErr     error
		wantStatus  int
		wantMarcado string
	}{
		{
			name:        "dono do CPF marca as atualizações como vistas",
			userCPF:     "12345678900",
			pathCPF:     "12345678900",
			wantStatus:  http.StatusNoContent,
			wantMarcado: "12345678900",
		},
		{
			name:       "sem usuário identificado",
			pathCPF:    "12345678900",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "CPF de outra pessoa",
			userCPF:    "99999999999",
			pathCPF:    "12345678900",
			wantStatus: http.StatusForbidden,
		},
		{
			// Ver as atualizações é do candidato: nem admin marca no lugar dele.
			name:       "admin com CPF de outra pessoa",
			userCPF:    "99999999999",
			isAdmin:    true,
			pathCPF:    "12345678900",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "erro do banco",
			userCPF:    "12345678900",
			pathCPF:    "12345678900",
			repoErr:    errors.New("db down"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candRepo := &mockCandidaturaRepoH{err: tt.repoErr}
			r := setupCandidaturaRouter(candRepo, &mockVagaForCandidaturaH{}, tt.userCPF, tt.isAdmin)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/candidaturas/usuario/"+tt.pathCPF+"/visualizar", nil))

			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())
			assert.Equal(t, tt.wantMarcado, candRepo.marcadasCPF)
		})
	}
}
