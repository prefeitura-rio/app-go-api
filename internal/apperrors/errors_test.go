package apperrors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/lib/pq"
)

func TestNewNotFound(t *testing.T) {
	err := NewNotFound("Curso")
	if err.Type != NotFound {
		t.Errorf("expected NotFound, got %d", err.Type)
	}
	if err.Error() != "Curso não encontrado(a)" {
		t.Errorf("unexpected message: %s", err.Error())
	}
}

func TestNewValidation(t *testing.T) {
	err := NewValidation("campo obrigatório", "titulo")
	if err.Type != Validation {
		t.Errorf("expected Validation, got %d", err.Type)
	}
	if err.Field != "titulo" {
		t.Errorf("expected field 'titulo', got '%s'", err.Field)
	}
}

func TestNewConflict(t *testing.T) {
	err := NewConflict("já existe", "email")
	if err.Type != Conflict {
		t.Errorf("expected Conflict, got %d", err.Type)
	}
}

func TestNewInternal(t *testing.T) {
	cause := errors.New("db connection failed")
	err := NewInternal("erro interno", cause)
	if err.Type != Internal {
		t.Errorf("expected Internal, got %d", err.Type)
	}
	if !errors.Is(err, cause) {
		t.Error("expected Unwrap to return cause")
	}
}

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{NewNotFound("x"), http.StatusNotFound},
		{NewValidation("x", ""), http.StatusBadRequest},
		{NewConflict("x", ""), http.StatusConflict},
		{NewForbidden("x"), http.StatusForbidden},
		{NewInternal("x", nil), http.StatusInternalServerError},
		{errors.New("unknown"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		status := HTTPStatus(tt.err)
		if status != tt.status {
			t.Errorf("HTTPStatus(%v) = %d, want %d", tt.err, status, tt.status)
		}
	}
}

func TestIsType(t *testing.T) {
	if !IsType(NewNotFound("x"), NotFound) {
		t.Error("expected IsType NotFound to be true")
	}
	if IsType(NewNotFound("x"), Validation) {
		t.Error("expected IsType Validation to be false for NotFound error")
	}
	if IsType(errors.New("plain"), NotFound) {
		t.Error("expected IsType to be false for non-AppError")
	}
}

func TestWrap(t *testing.T) {
	original := NewValidation("campo inválido", "nome")
	wrapped := Wrap(original, "erro ao processar")
	if wrapped.Type != Validation {
		t.Errorf("expected Validation type preserved, got %d", wrapped.Type)
	}

	plain := errors.New("generic")
	wrappedPlain := Wrap(plain, "contexto")
	if wrappedPlain.Type != Internal {
		t.Errorf("expected Internal for non-AppError, got %d", wrappedPlain.Type)
	}
}

func TestFromDatabaseError_UniqueViolation(t *testing.T) {
	pqErr := &pq.Error{Code: "23505", Constraint: "idx_unique_email"}
	appErr := FromDatabaseError(pqErr)
	if appErr.Type != Conflict {
		t.Errorf("expected Conflict, got %d", appErr.Type)
	}
}

func TestFromDatabaseError_ForeignKeyViolation(t *testing.T) {
	pqErr := &pq.Error{Code: "23503", Constraint: "fk_curso_id"}
	appErr := FromDatabaseError(pqErr)
	if appErr.Type != Validation {
		t.Errorf("expected Validation, got %d", appErr.Type)
	}
}

func TestFromDatabaseError_NotNullViolation(t *testing.T) {
	pqErr := &pq.Error{Code: "23502", Column: "titulo"}
	appErr := FromDatabaseError(pqErr)
	if appErr.Type != Validation {
		t.Errorf("expected Validation, got %d", appErr.Type)
	}
	if appErr.Field != "titulo" {
		t.Errorf("expected field 'titulo', got '%s'", appErr.Field)
	}
}

func TestFromDatabaseError_Nil(t *testing.T) {
	appErr := FromDatabaseError(nil)
	if appErr != nil {
		t.Error("expected nil for nil error")
	}
}

func TestFromDatabaseError_NonPQ(t *testing.T) {
	appErr := FromDatabaseError(errors.New("generic"))
	if appErr.Type != Internal {
		t.Errorf("expected Internal, got %d", appErr.Type)
	}
}
