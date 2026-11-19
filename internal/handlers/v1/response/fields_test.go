package response_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
)

// Erro de validação com campos: o corpo traz a mensagem geral e o erro de
// cada campo em "fields".
func TestRespondError_ValidationFields_DevolveOErroDeCadaCampo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/x", nil)

	response.RespondError(c, fmt.Errorf("publicar: %w", apperrors.ValidationFields("Preencha os campos obrigatórios.", map[string]string{
		"areas_atuacao":              "Informe ao menos uma área de atuação.",
		"comportamento_atitudes_ids": "Informe ao menos um comportamento ou atitude.",
	})))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var body struct {
		Error  string            `json:"error"`
		Fields map[string]string `json:"fields"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "Preencha os campos obrigatórios.", body.Error)
	assert.Equal(t, "Informe ao menos uma área de atuação.", body.Fields["areas_atuacao"])
	assert.Equal(t, "Informe ao menos um comportamento ou atitude.", body.Fields["comportamento_atitudes_ids"])
}

// Erro de domínio sem campos continua com o corpo de sempre, sem "fields".
func TestRespondError_SemFields_NaoTrazFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/x", nil)

	response.RespondError(c, apperrors.Validation("id inválido"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"id inválido"}`, w.Body.String())
}
