package empregabilidade_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// repoCapturaOrigemH guarda a origem do evento que chega ao repositório na
// exclusão (DEV-532).
type repoCapturaOrigemH struct {
	mockCandidaturaRepoH
	origem    empmodels.OrigemEventoCandidatura
	temOrigem bool
}

func (m *repoCapturaOrigemH) Delete(ctx context.Context, _ uuid.UUID) error {
	m.origem, m.temOrigem = empmodels.OrigemEventoDoContexto(ctx)
	return m.err
}

func TestCandidaturaHandler_Delete_OrigemDoCancelamento(t *testing.T) {
	id := uuid.MustParse(validUUID)

	t.Run("o próprio candidato cancela", func(t *testing.T) {
		repo := &repoCapturaOrigemH{mockCandidaturaRepoH: mockCandidaturaRepoH{entity: &empmodels.Candidatura{ID: id, CPF: "12345678900"}}}
		r := setupCandidaturaRouter(repo, &mockVagaForCandidaturaH{}, "12345678900", false)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/candidaturas/"+validUUID, nil))

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !repo.temOrigem || repo.origem.Ator != empmodels.AtorCandidato {
			t.Errorf("expected ator candidato, got %+v (presente=%v)", repo.origem, repo.temOrigem)
		}
	})

	t.Run("o operador de empregabilidade exclui a candidatura de outra pessoa", func(t *testing.T) {
		repo := &repoCapturaOrigemH{mockCandidaturaRepoH: mockCandidaturaRepoH{entity: &empmodels.Candidatura{ID: id, CPF: "12345678900"}}}
		r := setupCandidaturaRouterWithRoles(repo, &mockVagaForCandidaturaH{}, "99999999999", []string{"go:empregabilidade:admin"})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/candidaturas/"+validUUID, nil))

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if !repo.temOrigem || repo.origem.Ator != empmodels.AtorOperador {
			t.Errorf("expected ator operador, got %+v (presente=%v)", repo.origem, repo.temOrigem)
		}
	})
}
