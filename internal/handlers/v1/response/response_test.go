package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
)

func TestResponseHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Error response", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		response.Error(c, http.StatusBadRequest, "mensagem de erro")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "mensagem de erro")
	})

	t.Run("Success response", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		response.Success(c, http.StatusOK, "sucesso")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "sucesso")
	})

	t.Run("PaginatedJSON response", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		data := []string{"item1", "item2"}
		response.PaginatedJSON(c, http.StatusOK, data, 2, 1, 10)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "item1")
	})
}

func TestRespondError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
		wantField  string
		leak       string
	}{
		{"not found", apperrors.NotFound("vaga não encontrada"), http.StatusNotFound, "vaga não encontrada", "", ""},
		{"validation", apperrors.Validation("inválido"), http.StatusBadRequest, "inválido", "", ""},
		{"conflict", apperrors.Conflict("duplicado"), http.StatusConflict, "duplicado", "", ""},
		{"forbidden", apperrors.Forbidden("negado"), http.StatusForbidden, "negado", "", ""},
		{"gorm not found", gorm.ErrRecordNotFound, http.StatusNotFound, "Recurso não encontrado", "", ""},
		{
			"unique violation",
			&pq.Error{Code: pq.ErrorCode("23505"), Constraint: "users_email_key"},
			http.StatusConflict,
			"Este email já está cadastrado no sistema.",
			"email",
			"",
		},
		{
			"unknown no leak",
			errors.New("pq: connection refused segredo-interno"),
			http.StatusInternalServerError,
			"Erro interno do servidor",
			"",
			"segredo-interno",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

			response.RespondError(c, tt.err)

			assert.Equal(t, tt.wantStatus, w.Code)
			var body map[string]any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
			assert.Equal(t, tt.wantMsg, body["error"])
			if tt.wantField != "" {
				assert.Equal(t, tt.wantField, body["field"])
			} else {
				_, hasField := body["field"]
				assert.False(t, hasField)
			}
			if tt.leak != "" {
				assert.NotContains(t, w.Body.String(), tt.leak)
			}
		})
	}
}

func TestRespondError_FallbackMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("used only for unmapped 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/habilidades", nil)

		response.RespondError(c, errors.New("pq: connection refused segredo"), "Erro ao criar habilidade")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "Erro ao criar habilidade", body["error"])
		assert.NotContains(t, w.Body.String(), "segredo")
	})

	t.Run("ignored for mapped domain errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/vagas/1", nil)

		response.RespondError(c, apperrors.NotFound("vaga não encontrada"), "Erro ao buscar vaga")

		assert.Equal(t, http.StatusNotFound, w.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "vaga não encontrada", body["error"])
	})

	t.Run("ignored for unique violation", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/habilidades", nil)

		response.RespondError(c,
			&pq.Error{Code: pq.ErrorCode("23505"), Constraint: "emp_habilidades_nome_key"},
			"Erro ao criar habilidade",
		)

		assert.Equal(t, http.StatusConflict, w.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "Este nome já está cadastrado no sistema.", body["error"])
	})
}
