package apperrors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestError_Error_DoesNotExposeCause(t *testing.T) {
	err := Wrap(KindInternal, "Erro ao criar curso", errors.New("sql: connection refused"))
	assert.Equal(t, "Erro ao criar curso", err.Error())
	assert.NotContains(t, err.Error(), "sql")
	assert.NotContains(t, err.Error(), "connection")
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := Wrap(KindNotFound, "não encontrado", cause)
	assert.Equal(t, cause, errors.Unwrap(err))
	assert.True(t, errors.Is(err, cause))
}

func TestError_Is_Sentinel(t *testing.T) {
	sentinel := Conflict("já existe")
	assert.True(t, errors.Is(Conflict("já existe"), sentinel))
	assert.False(t, errors.Is(Conflict("outro"), sentinel))
	assert.False(t, errors.Is(NotFound("já existe"), sentinel))
}

func TestConstructors_HTTPStatus(t *testing.T) {
	tests := []struct {
		name   string
		err    *Error
		status int
	}{
		{"invalid", Invalid("bad"), http.StatusBadRequest},
		{"validation alias", Validation("bad"), http.StatusBadRequest},
		{"forbidden", Forbidden("no"), http.StatusForbidden},
		{"not found", NotFound("gone"), http.StatusNotFound},
		{"conflict", Conflict("dup"), http.StatusConflict},
		{"unavailable", Unavailable("down"), http.StatusServiceUnavailable},
		{"internal", Internal("boom"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.status, tt.err.HTTPStatus())
			assert.Equal(t, tt.status, kindStatus(tt.err.Kind))
		})
	}
}

func TestValidation_EqualsInvalid(t *testing.T) {
	assert.Equal(t, KindInvalid, KindValidation)
	assert.True(t, IsKind(Validation("x"), KindInvalid))
	assert.True(t, IsKind(Invalid("x"), KindValidation))
}

func TestAs_And_IsKind(t *testing.T) {
	err := fmt.Errorf("wrap: %w", NotFound("curso não encontrado"))
	ae, ok := As(err)
	require.True(t, ok)
	assert.Equal(t, KindNotFound, ae.Kind)
	assert.True(t, IsKind(err, KindNotFound))
	assert.False(t, IsKind(err, KindConflict))
	assert.False(t, IsKind(errors.New("plain"), KindNotFound))
}

func TestMap_TypedErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{"nil", nil, http.StatusInternalServerError, MsgInternal},
		{"invalid", Invalid("campo inválido"), http.StatusBadRequest, "campo inválido"},
		{"validation", Validation("regra"), http.StatusBadRequest, "regra"},
		{"forbidden", Forbidden("sem permissão"), http.StatusForbidden, "sem permissão"},
		{"not found", NotFound("não existe"), http.StatusNotFound, "não existe"},
		{"conflict", Conflict("duplicado"), http.StatusConflict, "duplicado"},
		{"unavailable", Unavailable("fora"), http.StatusServiceUnavailable, "fora"},
		{"internal", Internal("Erro ao listar"), http.StatusInternalServerError, "Erro ao listar"},
		{"gorm not found", gorm.ErrRecordNotFound, http.StatusNotFound, msgRecordNotFound},
		{"wrapped typed", fmt.Errorf("x: %w", Conflict("dup")), http.StatusConflict, "dup"},
		{"unknown", errors.New("sql: secret details"), http.StatusInternalServerError, MsgInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, message, field := Map(tt.err)
			assert.Equal(t, tt.status, status)
			assert.Equal(t, tt.message, message)
			assert.Empty(t, field)
			if tt.err != nil && !errors.Is(tt.err, gorm.ErrRecordNotFound) {
				if _, ok := As(tt.err); !ok {
					assert.NotContains(t, message, "sql")
					assert.NotEqual(t, tt.err.Error(), message)
				}
			}
		})
	}
}

func TestMap_PostgresUniqueViolation(t *testing.T) {
	pqErr := &pq.Error{Code: "23505", Constraint: "idx_unique_email"}
	status, message, field := Map(pqErr)
	assert.Equal(t, http.StatusConflict, status)
	assert.NotEmpty(t, message)
	_ = field
}

func TestMap_PostgresFKViolation(t *testing.T) {
	pqErr := &pq.Error{Code: "23503"}
	status, message, _ := Map(pqErr)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.NotEmpty(t, message)
}

func TestValidationFields_GuardaOErroDeCadaCampo(t *testing.T) {
	fields := map[string]string{"areas_atuacao": "Informe ao menos uma área de atuação."}
	err := ValidationFields("Preencha os campos obrigatórios.", fields)

	assert.Equal(t, "Preencha os campos obrigatórios.", err.Error())
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus())
	assert.Equal(t, fields, err.Fields)

	status, message, _ := Map(fmt.Errorf("publicar vaga: %w", err))
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "Preencha os campos obrigatórios.", message)
}
