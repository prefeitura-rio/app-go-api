package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type mockEscolaridadeRepoForHandler struct {
	createID  int
	entity    *models.Escolaridade
	listItems []*models.Escolaridade
	listTotal int
}

func (m *mockEscolaridadeRepoForHandler) Create(_ context.Context, _ *models.Escolaridade) (int, error) {
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockEscolaridadeRepoForHandler) GetByID(_ context.Context, id int) (*models.Escolaridade, error) {
	return m.entity, nil
}

func (m *mockEscolaridadeRepoForHandler) Update(_ context.Context, _ *models.Escolaridade) error {
	return nil
}

func (m *mockEscolaridadeRepoForHandler) Delete(_ context.Context, _ int) error {
	return nil
}

func (m *mockEscolaridadeRepoForHandler) List(_ context.Context, _ map[string]interface{}, _, _ int) ([]*models.Escolaridade, int, error) {
	if m.listItems == nil {
		return []*models.Escolaridade{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func setupEscolaridadeRouter(repo services.EscolaridadeRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewEscolaridadeService(repo)
	h := v1.NewEscolaridadeHandler(svc)
	r.POST("/api/v1/escolaridades", h.Create)
	r.GET("/api/v1/escolaridades", h.List)
	r.GET("/api/v1/escolaridades/:id", h.GetByID)
	r.PUT("/api/v1/escolaridades/:id", h.Update)
	r.DELETE("/api/v1/escolaridades/:id", h.Delete)
	return r
}

func TestEscolaridadeHandler_CreateAndGet(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{}
	router := setupEscolaridadeRouter(repo)
	body := []byte(`{"nivel":"Superior"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/escolaridades", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Create: expected 201, got %d", w.Code)
	}

	repo.entity = &models.Escolaridade{ID: 1, Nivel: "Superior"}
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/escolaridades/1", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("GetByID: expected 200, got %d", w2.Code)
	}
	var out models.Escolaridade
	if err := json.Unmarshal(w2.Body.Bytes(), &out); err != nil {
		t.Fatalf("GetByID unmarshal: %v", err)
	}
	if out.Nivel != "Superior" {
		t.Errorf("GetByID: expected Nivel Superior, got %s", out.Nivel)
	}
}

func TestEscolaridadeHandler_List(t *testing.T) {
	repo := &mockEscolaridadeRepoForHandler{
		listItems: []*models.Escolaridade{{ID: 1, Nivel: "Médio"}},
		listTotal: 1,
	}
	router := setupEscolaridadeRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/escolaridades", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("List: expected 200, got %d", w.Code)
	}
}
