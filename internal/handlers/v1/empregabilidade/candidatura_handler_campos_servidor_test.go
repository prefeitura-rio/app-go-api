package empregabilidade_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// candRepoCapturaCriacao guarda a candidatura que chegou ao repositório.
type candRepoCapturaCriacao struct {
	*mockCandidaturaRepoH
	recebida *empmodels.Candidatura
}

func (m *candRepoCapturaCriacao) Create(_ context.Context, entity *empmodels.Candidatura) (uuid.UUID, error) {
	copia := *entity
	m.recebida = &copia
	return uuid.New(), nil
}

// id e deleted_at enviados na criação são descartados: o cliente não escolhe o
// id da candidatura nem a cria já removida, seja cidadão ou admin (DEV-775).
func TestCandidaturaHandler_Create_DescartaCamposDoServidor(t *testing.T) {
	const idForjado = "6f1c2d3e-4a5b-4c6d-8e7f-0000000000ff"
	for _, tc := range []struct {
		name  string
		cpf   string
		admin bool
		body  string
	}{
		{"cidadao", "12345678900", false, `{"id":"` + idForjado + `","id_vaga":"` + validUUID + `","deleted_at":"2001-01-01T00:00:00Z","created_at":"2001-01-01T00:00:00Z","updated_at":"2001-01-01T00:00:00Z"}`},
		{"admin", "admin-user", true, `{"id":"` + idForjado + `","cpf":"12345678900","id_vaga":"` + validUUID + `","deleted_at":"2001-01-01T00:00:00Z","created_at":"2001-01-01T00:00:00Z","updated_at":"2001-01-01T00:00:00Z"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &candRepoCapturaCriacao{mockCandidaturaRepoH: &mockCandidaturaRepoH{}}
			r := setupCandidaturaRouter(repo, &mockVagaForCandidaturaH{entity: vagaAtivaComEtapa()}, tc.cpf, tc.admin)
			req := httptest.NewRequest(http.MethodPost, "/candidaturas", bodyOf(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
			}
			if repo.recebida == nil {
				t.Fatal("o repositório não recebeu a candidatura")
			}
			if repo.recebida.ID != uuid.Nil {
				t.Errorf("o id do corpo chegou ao repositório: %s", repo.recebida.ID)
			}
			if repo.recebida.DeletedAt.Valid {
				t.Errorf("deleted_at do corpo chegou ao repositório: %v", repo.recebida.DeletedAt.Time)
			}
			if !repo.recebida.CreatedAt.IsZero() || !repo.recebida.UpdatedAt.IsZero() {
				t.Errorf("created_at/updated_at do corpo chegaram ao repositório: %v / %v", repo.recebida.CreatedAt, repo.recebida.UpdatedAt)
			}

			var resp map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp["id"] == idForjado {
				t.Errorf("a resposta devolveu o id forjado")
			}
			if v, ok := resp["deleted_at"]; ok && v != nil {
				t.Errorf("a resposta devolveu deleted_at = %v", v)
			}
		})
	}
}
