package apperrors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
)

func TestConstructors_ErrorMessageAndKind(t *testing.T) {
	tests := []struct {
		name string
		err  *apperrors.Error
		kind apperrors.Kind
		msg  string
	}{
		{"NotFound", apperrors.NotFound("vaga não encontrada"), apperrors.KindNotFound, "vaga não encontrada"},
		{"Validation", apperrors.Validation("campo obrigatório"), apperrors.KindValidation, "campo obrigatório"},
		{"Conflict", apperrors.Conflict("já existe"), apperrors.KindConflict, "já existe"},
		{"Forbidden", apperrors.Forbidden("acesso negado"), apperrors.KindForbidden, "acesso negado"},
		{"Unavailable", apperrors.Unavailable("tente novamente"), apperrors.KindUnavailable, "tente novamente"},
		{"Internal", apperrors.Internal("falha interna"), apperrors.KindInternal, "falha interna"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.kind, tt.err.Kind)
			assert.Equal(t, tt.msg, tt.err.Error())
			assert.Equal(t, tt.msg, tt.err.Message)
		})
	}
}

func TestWrap_UnwrapsCause(t *testing.T) {
	cause := errors.New("root cause")
	err := apperrors.Wrap(apperrors.KindNotFound, "não achou", cause)
	assert.Equal(t, "não achou", err.Error())
	assert.True(t, errors.Is(err, cause))
}

func TestMap_DomainErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"not found", apperrors.NotFound("currículo não encontrado"), http.StatusNotFound, "currículo não encontrado"},
		{"validation", apperrors.Validation("dados inválidos"), http.StatusBadRequest, "dados inválidos"},
		{"conflict", apperrors.Conflict("já se candidatou"), http.StatusConflict, "já se candidatou"},
		{"forbidden", apperrors.Forbidden("acesso negado"), http.StatusForbidden, "acesso negado"},
		{"unavailable", apperrors.Unavailable("tente novamente"), http.StatusServiceUnavailable, "tente novamente"},
		{"internal", apperrors.Internal("ops"), http.StatusInternalServerError, "ops"},
		{"wrapped domain", fmt.Errorf("ctx: %w", apperrors.NotFound("vaga não encontrada")), http.StatusNotFound, "vaga não encontrada"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg, field := apperrors.Map(tt.err)
			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantMsg, msg)
			assert.Empty(t, field)
		})
	}
}

func TestMap_GormRecordNotFound(t *testing.T) {
	status, msg, _ := apperrors.Map(gorm.ErrRecordNotFound)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, "Recurso não encontrado", msg)

	status, msg, _ = apperrors.Map(fmt.Errorf("wrap: %w", gorm.ErrRecordNotFound))
	assert.Equal(t, http.StatusNotFound, status)
	assert.Equal(t, "Recurso não encontrado", msg)
}

func TestMap_PostgresConstraints(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantInMsg  string
		wantField  string
	}{
		{
			"unique pq",
			&pq.Error{Code: pq.ErrorCode("23505"), Constraint: "users_email_key"},
			http.StatusConflict,
			"Este email já está cadastrado",
			"email",
		},
		{
			"fk pgx",
			&pgconn.PgError{Code: "23503", ConstraintName: "curso_id"},
			http.StatusBadRequest,
			"O curso especificado não existe",
			"curso_id",
		},
		{
			"wrapped unique",
			fmt.Errorf("create: %w", &pgconn.PgError{Code: "23505", ConstraintName: "emp_regimes_contratacao_descricao_key"}),
			http.StatusConflict,
			"Já existe um registro",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, msg, field := apperrors.Map(tt.err)
			assert.Equal(t, tt.wantStatus, status)
			assert.Contains(t, msg, tt.wantInMsg)
			if tt.wantField != "" {
				assert.Equal(t, tt.wantField, field)
			}
		})
	}
}

func TestMap_UnknownError_NoLeak(t *testing.T) {
	err := errors.New("pq: connection refused segredo-interno TABLE emp_vagas")
	status, msg, field := apperrors.Map(err)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, "Erro interno do servidor", msg)
	assert.Empty(t, field)
	assert.NotContains(t, msg, "segredo-interno")
	assert.NotContains(t, msg, "emp_vagas")
}

func TestMap_Nil(t *testing.T) {
	status, msg, _ := apperrors.Map(nil)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, "Erro interno do servidor", msg)
}

func TestIsKind(t *testing.T) {
	err := apperrors.NotFound("x")
	require.True(t, apperrors.IsKind(err, apperrors.KindNotFound))
	require.False(t, apperrors.IsKind(err, apperrors.KindConflict))
	require.False(t, apperrors.IsKind(errors.New("x"), apperrors.KindNotFound))
}
