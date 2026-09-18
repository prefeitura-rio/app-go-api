package empregabilidade_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	handler "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// mockComportamentoAtitudesRepo implementa ComportamentoAtitudesRepositoryInterface
type mockComportamentoAtitudesRepo struct {
	err           error
	comportamento *empmodels.ComportamentoAtitudes
	list          []*empmodels.ComportamentoAtitudes
	total         int64
	lastCreatedID int64
}

func (m *mockComportamentoAtitudesRepo) CreateComportamentoAtitudes(_ context.Context, item *empmodels.ComportamentoAtitudes) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.lastCreatedID, nil
}

func (m *mockComportamentoAtitudesRepo) GetComportamentoAtitudesByID(_ context.Context, id int64) (*empmodels.ComportamentoAtitudes, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.comportamento, nil
}

func (m *mockComportamentoAtitudesRepo) UpdateComportamentoAtitudes(_ context.Context, item *empmodels.ComportamentoAtitudes) error {
	return m.err
}

func (m *mockComportamentoAtitudesRepo) DeleteComportamentoAtitudes(_ context.Context, id int64) error {
	return m.err
}

func (m *mockComportamentoAtitudesRepo) ListComportamentoAtitudes(_ context.Context, filter empmodels.ComportamentoAtitudesFilter, page, pageSize int) ([]*empmodels.ComportamentoAtitudes, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.list, m.total, nil
}

func setupComportametoAtitudesHandlerRouter(repo services.ComportamentoAtitudesRepositoryInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := services.NewComportamentoAtitudesServiceWithInterface(repo)
	h := handler.NewComportamentoAtitudesHandler(svc)

	r.GET("/comportamentos-atitudes", h.ListComportamentoAtitudes)
	r.POST("/comportamentos-atitudes", h.CreateComportamentoAtitudes)
	r.GET("/comportamentos-atitudes/:id", h.GetComportamentoAtitudesByID)
	r.PUT("/comportamentos-atitudes/:id", h.UpdateComportamentoAtitudes)
	r.DELETE("/comportamentos-atitudes/:id", h.DeleteComportamentoAtitudes)

	return r
}

func TestListComportamentoAtitudes(t *testing.T) {
	t.Run("sucesso - lista comportamentos", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{
			list:  []*empmodels.ComportamentoAtitudes{{ID: 1, Nome: "Pontualidade"}},
			total: 1,
		}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodGet, "/comportamentos-atitudes?q=Pontual&page=1&pageSize=10", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "Pontualidade")
	})

	t.Run("erro - erro interno ao listar", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{err: errors.New("database error")}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodGet, "/comportamentos-atitudes", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestCreateComportamentoAtitudes(t *testing.T) {
	t.Run("sucesso - cria comportamento", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{lastCreatedID: 10}
		router := setupComportametoAtitudesHandlerRouter(repo)

		body := map[string]string{"nome": "Trabalho em Equipe"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPost, "/comportamentos-atitudes", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)
		assert.Contains(t, resp.Body.String(), `"id":10`)
	})

	t.Run("erro - payload invalido", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodPost, "/comportamentos-atitudes", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("erro - erro interno no service", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{err: errors.New("db error")}
		router := setupComportametoAtitudesHandlerRouter(repo)

		body := map[string]string{"nome": "Liderança"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPost, "/comportamentos-atitudes", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestGetComportamentoAtitudesByID(t *testing.T) {
	t.Run("sucesso - busca por ID", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{
			comportamento: &empmodels.ComportamentoAtitudes{ID: 10, Nome: "Proatividade"},
		}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodGet, "/comportamentos-atitudes/10", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "Proatividade")
	})

	t.Run("erro - ID invalido", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodGet, "/comportamentos-atitudes/abc", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("erro - nao encontrado", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{comportamento: nil}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodGet, "/comportamentos-atitudes/99", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})
}

func TestUpdateComportamentoAtitudes(t *testing.T) {
	t.Run("sucesso - atualiza comportamento", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{}
		router := setupComportametoAtitudesHandlerRouter(repo)

		body := map[string]string{"nome": "Comunicação Assertiva"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPut, "/comportamentos-atitudes/10", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("erro - ID invalido na atualizacao", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodPut, "/comportamentos-atitudes/invalid", bytes.NewBufferString(`{"nome":"Teste"}`))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func TestDeleteComportamentoAtitudes(t *testing.T) {
	t.Run("sucesso - deleta comportamento", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodDelete, "/comportamentos-atitudes/10", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("erro - ID invalido na remocao", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodDelete, "/comportamentos-atitudes/abc", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func TestComportamentoAtitudes_CenariosDeErro(t *testing.T) {
	t.Run("erro 500 - get por id com falha no banco", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{err: errors.New("db error")}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodGet, "/comportamentos-atitudes/10", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("erro 500 - update com falha no banco", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{err: errors.New("db error")}
		router := setupComportametoAtitudesHandlerRouter(repo)

		body := map[string]string{"nome": "Comunicação Assertiva"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPut, "/comportamentos-atitudes/10", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("erro 400 - update com body invalido", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodPut, "/comportamentos-atitudes/10", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("erro 500 - delete com falha no banco", func(t *testing.T) {
		repo := &mockComportamentoAtitudesRepo{err: errors.New("db error")}
		router := setupComportametoAtitudesHandlerRouter(repo)

		req, _ := http.NewRequest(http.MethodDelete, "/comportamentos-atitudes/10", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}
