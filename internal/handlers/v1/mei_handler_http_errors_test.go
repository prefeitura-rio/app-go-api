package v1_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// DEV-49: oportunidades e propostas MEI respondiam 400 com o texto de qualquer
// erro do serviço, inclusive falha de banco. Agora o erro de domínio define o
// status e a falha de banco vira 500 genérico.

var meiDBLeak = fmt.Errorf("erro ao criar: %w", &pgconn.PgError{
	Code:    "08006",
	Message: `connection failure on relation "oportunidades_mei"`,
})

func assertNoMEILeak(t *testing.T, body string) {
	t.Helper()
	for _, marker := range []string{"SQLSTATE", "relation", "oportunidades_mei", "connection failure", "erro ao criar"} {
		assert.NotContains(t, body, marker, "o corpo vaza detalhe interno")
	}
}

func TestOportunidadeMEIHandler_ErrorMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		method     string
		path       string
		setup      func(*MockOportunidadeMEIService)
		wantStatus int
		wantBody   string
	}{
		{
			name:   "Create com falha de banco",
			method: http.MethodPost,
			path:   "/oportunidades",
			setup: func(s *MockOportunidadeMEIService) {
				s.On("Create", mock.Anything, mock.Anything, false).Return(0, meiDBLeak)
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"Erro ao criar oportunidade"}`,
		},
		{
			name:   "Update de oportunidade inexistente",
			method: http.MethodPut,
			path:   "/oportunidades/1",
			setup: func(s *MockOportunidadeMEIService) {
				s.On("Update", mock.Anything, mock.Anything).Return(services.ErrOportunidadeNaoEncontrada)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"oportunidade não encontrada"}`,
		},
		{
			name:       "Publish com falha de banco",
			method:     http.MethodPost,
			path:       "/oportunidades/1/publish",
			setup:      func(s *MockOportunidadeMEIService) { s.On("Publish", mock.Anything, 1).Return(meiDBLeak) },
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"Erro ao publicar oportunidade"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockOportunidadeMEIService)
			tt.setup(svc)
			h := v1.NewOportunidadeMEIHandler(svc)
			r := gin.New()
			r.POST("/oportunidades", h.Create)
			r.PUT("/oportunidades/:id", h.Update)
			r.POST("/oportunidades/:id/publish", h.Publish)

			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(`{"titulo":"x"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())
			assert.JSONEq(t, tt.wantBody, w.Body.String())
			assertNoMEILeak(t, w.Body.String())
		})
	}
}

func TestPropostaMEIHandler_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		svc        *MockPropostaMEIService
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Create com falha de banco",
			svc:        &MockPropostaMEIService{createError: meiDBLeak},
			method:     http.MethodPost,
			path:       "/api/v1/oportunidades-mei/1/propostas",
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"Erro ao criar proposta"}`,
		},
		{
			name:       "Create duplicada",
			svc:        &MockPropostaMEIService{createError: apperrors.Conflict("já existe uma proposta desta empresa para esta oportunidade")},
			method:     http.MethodPost,
			path:       "/api/v1/oportunidades-mei/1/propostas",
			wantStatus: http.StatusConflict,
			wantBody:   `{"error":"já existe uma proposta desta empresa para esta oportunidade"}`,
		},
		{
			name:       "Create sem CNPJ vinculado",
			svc:        &MockPropostaMEIService{createError: apperrors.Forbidden("nenhum CNPJ encontrado vinculado ao seu CPF")},
			method:     http.MethodPost,
			path:       "/api/v1/oportunidades-mei/1/propostas",
			wantStatus: http.StatusForbidden,
			wantBody:   `{"error":"nenhum CNPJ encontrado vinculado ao seu CPF"}`,
		},
		{
			name:       "UpdateStatus de proposta inexistente",
			svc:        &MockPropostaMEIService{updateError: services.ErrPropostaNaoEncontrada},
			method:     http.MethodPut,
			path:       "/api/v1/oportunidades-mei/1/propostas/00000000-0000-0000-0000-000000000001/status",
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"proposta não encontrada"}`,
		},
		{
			name:       "UpdateStatus com falha de banco",
			svc:        &MockPropostaMEIService{updateError: meiDBLeak},
			method:     http.MethodPut,
			path:       "/api/v1/oportunidades-mei/1/propostas/00000000-0000-0000-0000-000000000001/status",
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"Erro ao atualizar status da proposta"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"mei_empresa_id":"12345678000190","status":"approved"}`
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")
			w := httptest.NewRecorder()
			setupRouter(tt.svc).ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())
			assert.JSONEq(t, tt.wantBody, w.Body.String())
			assertNoMEILeak(t, w.Body.String())
		})
	}
}
