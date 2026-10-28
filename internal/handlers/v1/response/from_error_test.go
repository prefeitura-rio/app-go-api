package response_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
)

// leakMarkers são trechos que só aparecem em erros internos (driver, SQL, dado
// da chave). Nenhum deles pode chegar ao corpo da resposta.
var leakMarkers = []string{"SQLSTATE", "constraint", "DETAIL", "relation", "00000000000", "erro ao buscar", "dial tcp"}

func runFromError(t *testing.T, err error, internalMessage string) (int, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)

	response.FromError(c, err, internalMessage)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body), "body: %s", w.Body.String())
	for _, marker := range leakMarkers {
		assert.NotContains(t, w.Body.String(), marker, "o corpo vaza detalhe interno")
	}
	return w.Code, body["error"]
}

func TestFromError_TranslatesToHTTPStatus(t *testing.T) {
	uniqueViolation := fmt.Errorf("erro ao criar formação: %w", &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "uni_curriculo_formacao_cpf",
		Message:        `duplicate key value violates unique constraint "uni_curriculo_formacao_cpf"`,
		Detail:         "Key (cpf)=(00000000000) already exists.",
	})
	fkViolation := fmt.Errorf("erro ao criar formação: %w", &pgconn.PgError{
		Code:           "23503",
		ConstraintName: "fk_curriculo_formacao_escolaridade_id",
		Message:        `insert or update on table "curriculo_formacao" violates foreign key constraint`,
	})
	opaque := fmt.Errorf("erro ao buscar formação: %w", errors.New(`dial tcp 10.0.0.1:5432: connect: connection refused; relation "x" SQLSTATE 08006`))

	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{"domain not found", fmt.Errorf("svc: %w", apperrors.NotFound("Inscrição não encontrada")), http.StatusNotFound, "Inscrição não encontrada"},
		{"domain conflict", apperrors.Conflict("CPF já inscrito neste curso"), http.StatusConflict, "CPF já inscrito neste curso"},
		{"domain invalid", apperrors.Invalid("status inválido"), http.StatusBadRequest, "status inválido"},
		{"domain forbidden", apperrors.Forbidden("CNPJ não pertence ao seu CPF"), http.StatusForbidden, "CNPJ não pertence ao seu CPF"},
		{"domain unavailable keeps public message only", apperrors.Wrap(apperrors.KindUnavailable, "Serviço indisponível", opaque), http.StatusServiceUnavailable, "Serviço indisponível"},
		{"gorm record not found", fmt.Errorf("erro ao buscar: %w", gorm.ErrRecordNotFound), http.StatusNotFound, "Registro não encontrado"},
		{"unique violation", uniqueViolation, http.StatusConflict, "Este CPF já está cadastrado no sistema."},
		{"foreign key violation", fkViolation, http.StatusBadRequest, "A escolaridade especificada não existe. Verifique se o ID da escolaridade está correto."},
		{"unknown error is generic 500", opaque, http.StatusInternalServerError, "Erro ao buscar formação"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, message := runFromError(t, tt.err, "Erro ao buscar formação")
			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantMessage, message)
		})
	}
}

func TestFromError_NilErrorIsGeneric500(t *testing.T) {
	status, message := runFromError(t, nil, "Erro ao processar")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, "Erro ao processar", message)
}

func TestFromError_EmptyInternalMessageFallsBack(t *testing.T) {
	status, message := runFromError(t, errors.New(`pq: relation "x" does not exist`), "")
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, "Erro interno do servidor", message)
	assert.False(t, strings.Contains(message, "pq:"))
}

func TestInternalError_NeverExposesError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)

	response.InternalError(c, apperrors.NotFound("não importa"), "Erro ao salvar arquivo")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"Erro ao salvar arquivo"}`, w.Body.String())
}
