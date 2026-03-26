package empregabilidade_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Edge cases for pagination in lookup handlers

func TestRegimeContratacaoHandler_List_EdgeCases(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{}
	r := setupRegimeContratacaoRouter(repo)

	t.Run("InvalidPage", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/regimes?page=0&pageSize=5", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("InvalidPageSize_TooSmall", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/regimes?page=1&pageSize=0", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("InvalidPageSize_TooLarge", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/regimes?page=1&pageSize=9999", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestModeloTrabalhoHandler_List_EdgeCases(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{}
	r := setupModeloTrabalhoRouter(repo)

	t.Run("InvalidPage", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/modelos?page=-1&pageSize=5", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})

	t.Run("InvalidPageSize", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/modelos?page=1&pageSize=2000", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestModeloTrabalhoHandler_Update_BadJSON(t *testing.T) {
	repo := &mockModeloTrabalhoRepoH{}
	r := setupModeloTrabalhoRouter(repo)
	// valid ID but bad JSON body
	req := httptest.NewRequest(http.MethodPut, "/modelos/not-a-uuid", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTipoPCDHandler_List_EdgeCases(t *testing.T) {
	repo := &mockTipoPCDRepoH{}
	r := setupTipoPCDRouter(repo)

	t.Run("InvalidPage", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/tipos-pcd?page=0", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
}

func TestIdiomaHandler_Create_Error(t *testing.T) {
	repo := &mockIdiomaRepoH{err: errTest}
	r := setupIdiomaRouter(repo)
	body := bodyOf(`{"descricao":"Inglês"}`)
	req := httptest.NewRequest(http.MethodPost, "/idiomas", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestIdiomaHandler_Update_BadJSON(t *testing.T) {
	repo := &mockIdiomaRepoH{}
	r := setupIdiomaRouter(repo)
	req := httptest.NewRequest(http.MethodPut, "/idiomas/bad-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEscolaridadeHandler_Create_Error(t *testing.T) {
	repo := &mockEscolaridadeRepoH{err: errTest}
	r := setupEscolaridadeRouter(repo)
	body := bodyOf(`{"descricao":"Superior"}`)
	req := httptest.NewRequest(http.MethodPost, "/escolaridades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestSituacaoAtualHandler_Create_Error(t *testing.T) {
	repo := &mockSituacaoAtualRepoH{err: errTest}
	r := setupSituacaoAtualRouter(repo)
	body := bodyOf(`{"descricao":"Empregado"}`)
	req := httptest.NewRequest(http.MethodPost, "/situacoes", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestDisponibilidadeHandler_Create_Error(t *testing.T) {
	repo := &mockDisponibilidadeRepoH{err: errTest}
	r := setupDisponibilidadeRouter(repo)
	body := bodyOf(`{"descricao":"Imediata"}`)
	req := httptest.NewRequest(http.MethodPost, "/disponibilidades", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestNivelIdiomaHandler_Create_Error(t *testing.T) {
	repo := &mockNivelIdiomaRepoH{err: errTest}
	r := setupNivelIdiomaRouter(repo)
	body := bodyOf(`{"descricao":"Básico"}`)
	req := httptest.NewRequest(http.MethodPost, "/niveis", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestTipoConquistaHandler_Create_Error(t *testing.T) {
	repo := &mockTipoConquistaRepoH{err: errTest}
	r := setupTipoConquistaRouter(repo)
	body := bodyOf(`{"descricao":"Premio"}`)
	req := httptest.NewRequest(http.MethodPost, "/tipos-conquista", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Error("expected error status, got 201")
	}
}

func TestRegimeContratacaoHandler_GetByID_Error(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{err: errTest}
	r := setupRegimeContratacaoRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/regimes/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestRegimeContratacaoHandler_Update_Error(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{err: errTest}
	r := setupRegimeContratacaoRouter(repo)
	body := bodyOf(`{"descricao":"PJ"}`)
	req := httptest.NewRequest(http.MethodPut, "/regimes/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}

func TestRegimeContratacaoHandler_Update_BadJSON(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{}
	r := setupRegimeContratacaoRouter(repo)
	body := bodyOf(`invalid`)
	req := httptest.NewRequest(http.MethodPut, "/regimes/"+validUUID, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegimeContratacaoHandler_Delete_Error(t *testing.T) {
	repo := &mockRegimeContratacaoRepoH{err: errTest}
	r := setupRegimeContratacaoRouter(repo)
	req := httptest.NewRequest(http.MethodDelete, "/regimes/"+validUUID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Error("expected error status")
	}
}
