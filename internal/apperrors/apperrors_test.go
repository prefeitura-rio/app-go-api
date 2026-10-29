package apperrors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
)

func TestConstructors_KindAndMessage(t *testing.T) {
	tests := []struct {
		name       string
		err        *apperrors.Error
		wantKind   apperrors.Kind
		wantStatus int
	}{
		{"invalid", apperrors.Invalid("valor inválido"), apperrors.KindInvalid, http.StatusBadRequest},
		{"forbidden", apperrors.Forbidden("acesso negado"), apperrors.KindForbidden, http.StatusForbidden},
		{"not found", apperrors.NotFound("curso não encontrado"), apperrors.KindNotFound, http.StatusNotFound},
		{"conflict", apperrors.Conflict("já existe"), apperrors.KindConflict, http.StatusConflict},
		{"unavailable", apperrors.Unavailable("serviço fora"), apperrors.KindUnavailable, http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantKind, tt.err.Kind)
			assert.Equal(t, tt.wantStatus, tt.err.HTTPStatus())
			assert.Equal(t, tt.err.Message, tt.err.Error())
		})
	}
}

func TestError_MessageIsPublicOnly(t *testing.T) {
	cause := errors.New(`ERROR: relation "cursos" does not exist (SQLSTATE 42P01)`)
	err := apperrors.Wrap(apperrors.KindUnavailable, "não foi possível validar seus CNPJs", cause)

	assert.Equal(t, "não foi possível validar seus CNPJs", err.Error())
	assert.NotContains(t, err.Error(), "SQLSTATE")
	assert.ErrorIs(t, err, cause, "a causa continua disponível para o log do servidor")
}

func TestAs_FindsWrappedDomainError(t *testing.T) {
	domain := apperrors.NotFound("inscrição não encontrada")
	wrapped := fmt.Errorf("camada de serviço: %w", domain)

	got, ok := apperrors.As(wrapped)
	assert.True(t, ok)
	assert.Same(t, domain, got)

	_, ok = apperrors.As(errors.New("erro qualquer"))
	assert.False(t, ok)

	_, ok = apperrors.As(nil)
	assert.False(t, ok)
}

func TestIsKind(t *testing.T) {
	err := fmt.Errorf("x: %w", apperrors.Conflict("CPF já inscrito neste curso"))

	assert.True(t, apperrors.IsKind(err, apperrors.KindConflict))
	assert.False(t, apperrors.IsKind(err, apperrors.KindNotFound))
	assert.False(t, apperrors.IsKind(errors.New("CPF já inscrito neste curso"), apperrors.KindConflict))
}

func TestErrorsIs_ComparesKindAndMessage(t *testing.T) {
	sentinel := apperrors.NotFound("proposta não encontrada")
	same := apperrors.NotFound("proposta não encontrada")
	other := apperrors.NotFound("oportunidade não encontrada")

	assert.ErrorIs(t, fmt.Errorf("w: %w", same), sentinel)
	assert.NotErrorIs(t, other, sentinel)
	assert.NotErrorIs(t, apperrors.Conflict("proposta não encontrada"), sentinel)
}

func TestHTTPStatus_UnknownKindIs500(t *testing.T) {
	err := &apperrors.Error{Kind: apperrors.Kind(99), Message: "x"}
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus())
}
