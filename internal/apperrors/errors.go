package apperrors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

type ErrorType int

const (
	NotFound ErrorType = iota
	Validation
	Conflict
	Forbidden
	Internal
)

type AppError struct {
	Type    ErrorType
	Message string
	Field   string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewNotFound(entity string) *AppError {
	return &AppError{
		Type:    NotFound,
		Message: fmt.Sprintf("%s não encontrado(a)", entity),
	}
}

func NewNotFoundf(format string, args ...interface{}) *AppError {
	return &AppError{
		Type:    NotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

func NewValidation(message, field string) *AppError {
	return &AppError{
		Type:    Validation,
		Message: message,
		Field:   field,
	}
}

func NewValidationf(field, format string, args ...interface{}) *AppError {
	return &AppError{
		Type:    Validation,
		Message: fmt.Sprintf(format, args...),
		Field:   field,
	}
}

func NewConflict(message, field string) *AppError {
	return &AppError{
		Type:    Conflict,
		Message: message,
		Field:   field,
	}
}

func NewForbidden(message string) *AppError {
	return &AppError{
		Type:    Forbidden,
		Message: message,
	}
}

func NewInternal(message string, err error) *AppError {
	return &AppError{
		Type:    Internal,
		Message: message,
		Err:     err,
	}
}

func NewInternalf(err error, format string, args ...interface{}) *AppError {
	return &AppError{
		Type:    Internal,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

func Wrap(err error, message string) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return &AppError{
			Type:    appErr.Type,
			Message: message,
			Field:   appErr.Field,
			Err:     err,
		}
	}
	return NewInternal(message, err)
}

func HTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Type {
		case NotFound:
			return http.StatusNotFound
		case Validation:
			return http.StatusBadRequest
		case Conflict:
			return http.StatusConflict
		case Forbidden:
			return http.StatusForbidden
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

func GetField(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Field
	}
	return ""
}

func IsType(err error, t ErrorType) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == t
	}
	return false
}

func FromDatabaseError(err error) *AppError {
	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return NewInternal("Erro interno do servidor", err)
	}

	switch pqErr.Code {
	case "23503": // foreign_key_violation
		return parseForeignKeyError(pqErr)
	case "23505": // unique_violation
		return parseUniqueViolationError(pqErr)
	case "23502": // not_null_violation
		return parseNotNullError(pqErr)
	case "23514": // check_violation
		return parseCheckViolationError(pqErr)
	default:
		return NewInternal("Erro interno do servidor", err)
	}
}

func parseForeignKeyError(pqErr *pq.Error) *AppError {
	constraint := pqErr.Constraint

	fieldMessages := map[string]struct {
		field   string
		message string
	}{
		"instituicao_id":  {"instituicao_id", "A instituição de ensino especificada não existe"},
		"empresa_id":      {"empresa_id", "A empresa especificada não existe"},
		"escolaridade_id": {"escolaridade_id", "A escolaridade especificada não existe"},
		"categoria_id":    {"categoria_id", "A categoria especificada não existe"},
		"curso_id":        {"curso_id", "O curso especificado não existe"},
	}

	for key, info := range fieldMessages {
		if containsSubstring(constraint, key) {
			return NewValidation(info.message, info.field)
		}
	}

	return NewValidation("Referência inválida. Verifique se todos os IDs fornecidos existem no sistema.", "")
}

func parseUniqueViolationError(pqErr *pq.Error) *AppError {
	constraint := pqErr.Constraint

	fieldMessages := map[string]struct {
		field   string
		message string
	}{
		"email": {"email", "Este email já está cadastrado no sistema"},
		"cpf":   {"cpf", "Este CPF já está cadastrado no sistema"},
		"nome":  {"nome", "Este nome já está cadastrado no sistema"},
	}

	for key, info := range fieldMessages {
		if containsSubstring(constraint, key) {
			return NewConflict(info.message, info.field)
		}
	}

	return NewConflict("Já existe um registro com essas informações no sistema", "")
}

func parseNotNullError(pqErr *pq.Error) *AppError {
	column := pqErr.Column

	friendlyNames := map[string]string{
		"titulo":       "título",
		"nome":         "nome",
		"email":        "email",
		"cpf":          "CPF",
		"modalidade":   "modalidade",
		"status":       "status",
		"organization": "organização",
		"descricao":    "descrição",
	}

	fieldName := friendlyNames[column]
	if fieldName == "" {
		fieldName = column
	}

	return NewValidation(
		fmt.Sprintf("O campo '%s' é obrigatório e não pode estar vazio", fieldName),
		column,
	)
}

func parseCheckViolationError(pqErr *pq.Error) *AppError {
	constraint := pqErr.Constraint

	if containsSubstring(constraint, "modalidade") {
		return NewValidation("Modalidade inválida", "modalidade")
	}
	if containsSubstring(constraint, "status") {
		return NewValidation("Status inválido", "status")
	}

	return NewValidation("Valor inválido fornecido para um dos campos", "")
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
