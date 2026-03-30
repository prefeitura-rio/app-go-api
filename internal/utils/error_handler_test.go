package utils

import (
	"errors"
	"strings"
	"testing"

	"github.com/lib/pq"
)

func TestParseDatabaseError_Nil(t *testing.T) {
	out := ParseDatabaseError(nil)
	if out != nil {
		t.Errorf("ParseDatabaseError(nil): expected nil, got %+v", out)
	}
}

func TestParseDatabaseError_NonPQError(t *testing.T) {
	err := errors.New("generic error")
	out := ParseDatabaseError(err)
	if out == nil {
		t.Fatal("ParseDatabaseError: expected non-nil")
	}
	if out.Type != UnknownError {
		t.Errorf("expected UnknownError, got %d", out.Type)
	}
	if out.Message != "Erro interno do servidor" {
		t.Errorf("unexpected Message: %s", out.Message)
	}
}

func TestParseDatabaseError_ForeignKeyViolation(t *testing.T) {
	tests := []struct {
		constraint string
		wantField  string
		wantMsg    string
	}{
		{"fk_instituicao_id", "instituicao_id", "A instituição de ensino especificada não existe"},
		{"cursos_empresa_id_fkey", "empresa_id", "A empresa especificada não existe"},
		{"escolaridade_id", "escolaridade_id", "A escolaridade especificada não existe"},
		{"categoria_id", "categoria_id", "A categoria especificada não existe"},
		{"curso_id", "curso_id", "O curso especificado não existe"},
		{"other_fkey", "", "Referência inválida"},
	}
	for _, tt := range tests {
		t.Run(tt.constraint, func(t *testing.T) {
			err := &pq.Error{Code: pq.ErrorCode("23503"), Constraint: tt.constraint}
			out := ParseDatabaseError(err)
			if out == nil {
				t.Fatal("expected non-nil")
			}
			if out.Type != ForeignKeyViolation {
				t.Errorf("expected ForeignKeyViolation, got %d", out.Type)
			}
			if tt.wantField != "" && out.Field != tt.wantField {
				t.Errorf("Field: got %s, want %s", out.Field, tt.wantField)
			}
			if tt.wantMsg != "" && out.Message != "" && out.Message[:len(tt.wantMsg)] != tt.wantMsg {
				t.Errorf("Message: got %s, want prefix %s", out.Message, tt.wantMsg)
			}
		})
	}
}

func TestParseDatabaseError_UniqueViolation(t *testing.T) {
	tests := []struct {
		constraint string
		wantField  string
		wantMsg    string
	}{
		{"users_email_key", "email", "Este email já está cadastrado"},
		{"cpf_unique", "cpf", "Este CPF já está cadastrado"},
		{"categorias_nome_key", "nome", "Este nome já está cadastrado"},
		{"other_unique", "", "Já existe um registro"},
	}
	for _, tt := range tests {
		t.Run(tt.constraint, func(t *testing.T) {
			err := &pq.Error{Code: pq.ErrorCode("23505"), Constraint: tt.constraint}
			out := ParseDatabaseError(err)
			if out == nil {
				t.Fatal("expected non-nil")
			}
			if out.Type != UniqueViolation {
				t.Errorf("expected UniqueViolation, got %d", out.Type)
			}
			if tt.wantField != "" && out.Field != tt.wantField {
				t.Errorf("Field: got %s, want %s", out.Field, tt.wantField)
			}
			if tt.wantMsg != "" && out.Message != "" && out.Message[:len(tt.wantMsg)] != tt.wantMsg {
				t.Errorf("Message: got %s, want prefix %s", out.Message, tt.wantMsg)
			}
		})
	}
}

func TestParseDatabaseError_NotNullViolation(t *testing.T) {
	err := &pq.Error{Code: pq.ErrorCode("23502"), Column: "nome"}
	out := ParseDatabaseError(err)
	if out == nil {
		t.Fatal("expected non-nil")
	}
	if out.Type != NotNullViolation {
		t.Errorf("expected NotNullViolation, got %d", out.Type)
	}
	if out.Message == "" || out.Message[:len("O campo '")] != "O campo '" {
		t.Errorf("unexpected Message: %s", out.Message)
	}
}

func TestParseDatabaseError_CheckViolation(t *testing.T) {
	t.Run("modalidade", func(t *testing.T) {
		err := &pq.Error{Code: pq.ErrorCode("23514"), Constraint: "chk_modalidade"}
		out := ParseDatabaseError(err)
		if out == nil {
			t.Fatal("expected non-nil")
		}
		if out.Type != CheckViolation || out.Field != "modalidade" {
			t.Errorf("got Type=%d Field=%s", out.Type, out.Field)
		}
	})
	t.Run("status", func(t *testing.T) {
		err := &pq.Error{Code: pq.ErrorCode("23514"), Constraint: "status_check"}
		out := ParseDatabaseError(err)
		if out == nil {
			t.Fatal("expected non-nil")
		}
		if out.Field != "status" {
			t.Errorf("Field: got %s", out.Field)
		}
	})
}

func TestDatabaseError_GetUserFriendlyMessage(t *testing.T) {
	if (*DatabaseError)(nil).GetUserFriendlyMessage() != "Erro interno do servidor" {
		t.Error("nil receiver should return default message")
	}
	de := &DatabaseError{Message: "custom"}
	if de.GetUserFriendlyMessage() != "custom" {
		t.Errorf("got %s", de.GetUserFriendlyMessage())
	}
}

func TestDatabaseError_GetHTTPStatusCode(t *testing.T) {
	if (*DatabaseError)(nil).GetHTTPStatusCode() != 500 {
		t.Error("nil receiver should return 500")
	}
	tests := []struct {
		typ  DatabaseErrorType
		want int
	}{
		{ForeignKeyViolation, 400},
		{UniqueViolation, 409},
		{NotNullViolation, 400},
		{CheckViolation, 400},
		{UnknownError, 500},
	}
	for _, tt := range tests {
		de := &DatabaseError{Type: tt.typ}
		if got := de.GetHTTPStatusCode(); got != tt.want {
			t.Errorf("Type %d: got %d, want %d", tt.typ, got, tt.want)
		}
	}
}

func TestParseNotNullError_AllFieldMappings(t *testing.T) {
	tests := []struct {
		column   string
		wantName string
	}{
		{"titulo", "título"},
		{"nome", "nome"},
		{"email", "email"},
		{"cpf", "CPF"},
		{"modalidade", "modalidade"},
		{"status", "status"},
		{"organization", "organização"},
		{"descricao", "descrição"},
		{"unknown_field", "unknown_field"},
	}
	for _, tt := range tests {
		t.Run(tt.column, func(t *testing.T) {
			err := &pq.Error{Code: pq.ErrorCode("23502"), Column: tt.column}
			out := ParseDatabaseError(err)
			if out == nil {
				t.Fatal("expected non-nil")
			}
			if out.Type != NotNullViolation {
				t.Errorf("expected NotNullViolation, got %d", out.Type)
			}
			if !strings.Contains(out.Message, tt.wantName) {
				t.Errorf("Message %q should contain %q", out.Message, tt.wantName)
			}
		})
	}
}

func TestParseCheckViolationError_AllConstraints(t *testing.T) {
	tests := []struct {
		constraint string
		wantField  string
		wantInMsg  string
	}{
		{"chk_modalidade_valid", "modalidade", "Modalidade inválida"},
		{"check_status", "status", "Status inválido"},
		{"other_check", "", "Valor inválido"},
	}
	for _, tt := range tests {
		t.Run(tt.constraint, func(t *testing.T) {
			err := &pq.Error{Code: pq.ErrorCode("23514"), Constraint: tt.constraint}
			out := ParseDatabaseError(err)
			if out == nil {
				t.Fatal("expected non-nil")
			}
			if out.Type != CheckViolation {
				t.Errorf("expected CheckViolation, got %d", out.Type)
			}
			if tt.wantField != "" && out.Field != tt.wantField {
				t.Errorf("Field: got %s, want %s", out.Field, tt.wantField)
			}
			if !strings.Contains(out.Message, tt.wantInMsg) {
				t.Errorf("Message %q should contain %q", out.Message, tt.wantInMsg)
			}
		})
	}
}

func TestParseDatabaseError_UnknownCode(t *testing.T) {
	// Test an unknown PostgreSQL error code
	err := &pq.Error{Code: pq.ErrorCode("99999"), Message: "Some unknown error"}
	out := ParseDatabaseError(err)
	if out == nil {
		t.Fatal("expected non-nil")
	}
	if out.Type != UnknownError {
		t.Errorf("expected UnknownError, got %d", out.Type)
	}
	if out.Message != "Erro interno do servidor" {
		t.Errorf("unexpected Message: %s", out.Message)
	}
}
