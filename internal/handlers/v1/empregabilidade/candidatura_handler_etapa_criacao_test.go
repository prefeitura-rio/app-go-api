package empregabilidade_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// Etapas sintéticas: uma da vaga da candidatura, outra de uma vaga diferente.
const (
	etapaDaVagaUUID   = "6f1c2d3e-4a5b-4c6d-8e7f-000000000001"
	etapaOutraVagaUID = "6f1c2d3e-4a5b-4c6d-8e7f-000000000002"
)

func vagaAtivaComEtapa() *empmodels.Vaga {
	vagaID := uuid.MustParse(validUUID)
	return &empmodels.Vaga{
		ID:     vagaID,
		Status: empmodels.StatusVagaPublicadoAtivo,
		Etapas: []empmodels.Etapa{{ID: uuid.MustParse(etapaDaVagaUUID), IDVaga: vagaID, Titulo: "Análise de currículo", Ordem: 1}},
	}
}

func postCandidatura(t *testing.T, cpf string, admin bool, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := setupCandidaturaRouter(&mockCandidaturaRepoH{}, &mockVagaForCandidaturaH{entity: vagaAtivaComEtapa()}, cpf, admin)
	req := httptest.NewRequest(http.MethodPost, "/candidaturas", bodyOf(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// O cidadão não define etapa nem status anterior: os dois vêm do fluxo
// autorizado de etapa/status. O corpo forjado é descartado e a candidatura
// nasce sem eles (DEV-767).
func TestCandidaturaHandler_Create_Cidadao_DescartaEtapaEStatusAnterior(t *testing.T) {
	for _, etapa := range []string{etapaOutraVagaUID, etapaDaVagaUUID} {
		t.Run(etapa, func(t *testing.T) {
			w := postCandidatura(t, "12345678900", false,
				`{"id_vaga":"`+validUUID+`","id_etapa_atual":"`+etapa+`","status_anterior":"aprovada"}`)

			if w.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
			}
			var criada empmodels.Candidatura
			if err := json.Unmarshal(w.Body.Bytes(), &criada); err != nil {
				t.Fatal(err)
			}
			if criada.IDEtapaAtual != nil {
				t.Errorf("id_etapa_atual deveria ficar vazio, veio %s", *criada.IDEtapaAtual)
			}
			if criada.StatusAnterior != nil {
				t.Errorf("status_anterior deveria ficar vazio, veio %s", *criada.StatusAnterior)
			}
			if criada.Status != empmodels.StatusCandidaturaEnviada {
				t.Errorf("status = %s, want %s", criada.Status, empmodels.StatusCandidaturaEnviada)
			}
		})
	}
}

// O admin que cadastra a candidatura pode já pô-la numa etapa, desde que seja
// da própria vaga; etapa de outra vaga é 400. O status anterior nunca vem do corpo.
func TestCandidaturaHandler_Create_Admin_ValidaEtapaDaVaga(t *testing.T) {
	t.Run("etapa de outra vaga", func(t *testing.T) {
		w := postCandidatura(t, "admin-user", true,
			`{"cpf":"12345678900","id_vaga":"`+validUUID+`","id_etapa_atual":"`+etapaOutraVagaUID+`"}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("etapa da vaga", func(t *testing.T) {
		w := postCandidatura(t, "admin-user", true,
			`{"cpf":"12345678900","id_vaga":"`+validUUID+`","id_etapa_atual":"`+etapaDaVagaUUID+`","status_anterior":"aprovada"}`)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var criada empmodels.Candidatura
		if err := json.Unmarshal(w.Body.Bytes(), &criada); err != nil {
			t.Fatal(err)
		}
		if criada.IDEtapaAtual == nil || criada.IDEtapaAtual.String() != etapaDaVagaUUID {
			t.Errorf("id_etapa_atual = %v, want %s", criada.IDEtapaAtual, etapaDaVagaUUID)
		}
		if criada.StatusAnterior != nil {
			t.Errorf("status_anterior deveria ficar vazio, veio %s", *criada.StatusAnterior)
		}
	})
}
