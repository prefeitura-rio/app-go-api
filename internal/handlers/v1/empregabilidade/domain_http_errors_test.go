package empregabilidade_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// INFRAVPIA-213 / domain errors: handlers must map NotFound→404, Conflict→409,
// Validation/bind→400, and never leak internal DB details on 500.

func TestVagaHandler_DomainErrors_HTTPStatusMapping(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	t.Run("not found returns 404", func(t *testing.T) {
		vagaRepo := &mockVagaRepoH{entity: nil, err: nil}
		empresaRepo := &mockEmpresaRepoForVaga{}
		r := setupVagaRouter(vagaRepo, empresaRepo, true)
		req := httptest.NewRequest(http.MethodPut, "/vagas/"+id.String()+"/publish", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "body: %s", w.Body.String())
		assert.Equal(t, "vaga não encontrada", decodeErrorMessage(t, w.Body.Bytes()))
	})

	t.Run("wrong state returns 409", func(t *testing.T) {
		vagaRepo := &mockVagaRepoH{entity: &empmodels.Vaga{
			ID:     id,
			Titulo: "Dev",
			Status: empmodels.StatusVagaPublicadoAtivo,
		}}
		empresaRepo := &mockEmpresaRepoForVaga{}
		r := setupVagaRouter(vagaRepo, empresaRepo, true)
		req := httptest.NewRequest(http.MethodPut, "/vagas/"+id.String()+"/send-to-approval", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "body: %s", w.Body.String())
		assert.Contains(t, decodeErrorMessage(t, w.Body.Bytes()), "não está em estado")
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		vagaRepo := &mockVagaRepoH{}
		empresaRepo := &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}
		r := setupVagaRouter(vagaRepo, empresaRepo, true)
		req := httptest.NewRequest(http.MethodPost, "/vagas", bodyOf(`{invalid`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unique violation returns 409 without leak", func(t *testing.T) {
		vagaRepo := &mockVagaRepoH{
			err: &pq.Error{Code: pq.ErrorCode("23505"), Constraint: "emp_vagas_pkey"},
		}
		empresaRepo := &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}
		r := setupVagaRouter(vagaRepo, empresaRepo, true)
		req := httptest.NewRequest(http.MethodPost, "/vagas", bodyOf(
			`{"titulo":"Dev Backend","tipo_contratacao":"CLT","contratante_cnpj":"12345678000100"}`,
		))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "body: %s", w.Body.String())
		msg := decodeErrorMessage(t, w.Body.Bytes())
		assert.Contains(t, msg, "Já existe um registro")
		assert.NotContains(t, w.Body.String(), "emp_vagas_pkey")
	})

	t.Run("unknown error returns 500 without leak", func(t *testing.T) {
		vagaRepo := &mockVagaRepoH{err: errors.New("pq: connection refused segredo-interno")}
		empresaRepo := &mockEmpresaRepoForVaga{entity: &empmodels.Empresa{CNPJ: "12345678000100"}}
		r := setupVagaRouter(vagaRepo, empresaRepo, true)
		req := httptest.NewRequest(http.MethodPost, "/vagas", bodyOf(
			`{"titulo":"Dev Backend","tipo_contratacao":"CLT","contratante_cnpj":"12345678000100"}`,
		))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "Erro interno do servidor", decodeErrorMessage(t, w.Body.Bytes()))
		assert.NotContains(t, w.Body.String(), "segredo-interno")
	})
}

func TestCandidaturaHandler_DomainErrors_HTTPStatusMapping(t *testing.T) {
	vagaID := uuid.MustParse(validUUID)
	vagaRepo := &mockVagaForCandidaturaH{
		entity: &empmodels.Vaga{
			ID:     vagaID,
			Status: empmodels.StatusVagaPublicadoAtivo,
		},
	}

	t.Run("duplicate candidatura returns 409", func(t *testing.T) {
		candRepo := &mockCandidaturaRepoH{exists: true}
		r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
		body := bodyOf(`{"id_vaga":"` + vagaID.String() + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/candidaturas", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "body: %s", w.Body.String())
		assert.Contains(t, decodeErrorMessage(t, w.Body.Bytes()), "já existe")
	})

	t.Run("vaga not found returns 404", func(t *testing.T) {
		candRepo := &mockCandidaturaRepoH{exists: false}
		r := setupCandidaturaRouter(candRepo, &mockVagaForCandidaturaH{entity: nil}, "12345678900", false)
		body := bodyOf(`{"id_vaga":"` + vagaID.String() + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/candidaturas", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "body: %s", w.Body.String())
		assert.Equal(t, "vaga não encontrada", decodeErrorMessage(t, w.Body.Bytes()))
	})

	t.Run("invalid payload returns 400", func(t *testing.T) {
		candRepo := &mockCandidaturaRepoH{}
		r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
		req := httptest.NewRequest(http.MethodPost, "/candidaturas", bodyOf(`{bad`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unknown error returns 500 without leak", func(t *testing.T) {
		candRepo := &mockCandidaturaRepoH{err: errors.New("pq: relation emp_candidaturas does not exist SEGREDO")}
		r := setupCandidaturaRouter(candRepo, vagaRepo, "12345678900", false)
		body := bodyOf(`{"id_vaga":"` + vagaID.String() + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/candidaturas", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "Erro interno do servidor", resp["error"])
		assert.NotContains(t, w.Body.String(), "SEGREDO")
		assert.NotContains(t, w.Body.String(), "emp_candidaturas")
	})
}
