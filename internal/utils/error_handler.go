package utils

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

// DatabaseError representa diferentes tipos de erros de banco de dados
type DatabaseError struct {
	Type    DatabaseErrorType
	Message string
	Field   string
	Value   string
}

type DatabaseErrorType int

const (
	ForeignKeyViolation DatabaseErrorType = iota
	UniqueViolation
	NotNullViolation
	CheckViolation
	UnknownError
)

// pgErrorFields normaliza erros do lib/pq e do pgx (usado pelo GORM postgres driver).
type pgErrorFields struct {
	Code       string
	Constraint string
	Column     string
}

func extractPGErrorFields(err error) (pgErrorFields, bool) {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pgErrorFields{
			Code:       string(pqErr.Code),
			Constraint: pqErr.Constraint,
			Column:     pqErr.Column,
		}, true
	}

	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) {
		return pgErrorFields{
			Code:       pgxErr.Code,
			Constraint: pgxErr.ConstraintName,
			Column:     pgxErr.ColumnName,
		}, true
	}

	return pgErrorFields{}, false
}

// ParseDatabaseError analisa erros do PostgreSQL e retorna mensagens amigáveis
func ParseDatabaseError(err error) *DatabaseError {
	if err == nil {
		return nil
	}

	fields, ok := extractPGErrorFields(err)
	if !ok {
		// In production, hide detailed error messages
		return &DatabaseError{
			Type:    UnknownError,
			Message: "Erro interno do servidor",
		}
	}

	switch fields.Code {
	case "23503": // foreign_key_violation
		return parseForeignKeyError(fields.Constraint)
	case "23505": // unique_violation
		return parseUniqueViolationError(fields.Constraint)
	case "23502": // not_null_violation
		return parseNotNullError(fields.Column)
	case "23514": // check_violation
		return parseCheckViolationError(fields.Constraint)
	default:
		return &DatabaseError{
			Type:    UnknownError,
			Message: "Erro interno do servidor",
		}
	}
}

func parseForeignKeyError(constraint string) *DatabaseError {
	// Mapear constraints para mensagens amigáveis
	switch {
	case strings.Contains(constraint, "instituicao_id"):
		return &DatabaseError{
			Type:    ForeignKeyViolation,
			Message: "A instituição de ensino especificada não existe. Verifique se o ID da instituição está correto.",
			Field:   "instituicao_id",
		}
	case strings.Contains(constraint, "empresa_id"):
		return &DatabaseError{
			Type:    ForeignKeyViolation,
			Message: "A empresa especificada não existe. Verifique se o ID da empresa está correto.",
			Field:   "empresa_id",
		}
	case strings.Contains(constraint, "escolaridade_id"):
		return &DatabaseError{
			Type:    ForeignKeyViolation,
			Message: "A escolaridade especificada não existe. Verifique se o ID da escolaridade está correto.",
			Field:   "escolaridade_id",
		}
	case strings.Contains(constraint, "categoria_id"):
		return &DatabaseError{
			Type:    ForeignKeyViolation,
			Message: "A categoria especificada não existe. Verifique se o ID da categoria está correto.",
			Field:   "categoria_id",
		}
	case strings.Contains(constraint, "curso_id"):
		return &DatabaseError{
			Type:    ForeignKeyViolation,
			Message: "O curso especificado não existe. Verifique se o ID do curso está correto.",
			Field:   "curso_id",
		}
	default:
		return &DatabaseError{
			Type:    ForeignKeyViolation,
			Message: "Referência inválida. Verifique se todos os IDs fornecidos existem no sistema.",
		}
	}
}

func parseUniqueViolationError(constraint string) *DatabaseError {
	switch {
	case strings.Contains(constraint, "email"):
		return &DatabaseError{
			Type:    UniqueViolation,
			Message: "Este email já está cadastrado no sistema.",
			Field:   "email",
		}
	case strings.Contains(constraint, "cpf"):
		return &DatabaseError{
			Type:    UniqueViolation,
			Message: "Este CPF já está cadastrado no sistema.",
			Field:   "cpf",
		}
	case strings.Contains(constraint, "nome"):
		return &DatabaseError{
			Type:    UniqueViolation,
			Message: "Este nome já está cadastrado no sistema.",
			Field:   "nome",
		}
	default:
		return &DatabaseError{
			Type:    UniqueViolation,
			Message: "Já existe um registro com essas informações no sistema.",
		}
	}
}

func parseNotNullError(column string) *DatabaseError {
	// Mapear colunas para nomes amigáveis
	fieldNames := map[string]string{
		"titulo":       "título",
		"nome":         "nome",
		"email":        "email",
		"cpf":          "CPF",
		"modalidade":   "modalidade",
		"status":       "status",
		"organization": "organização",
		"descricao":    "descrição",
	}

	fieldName := fieldNames[column]
	if fieldName == "" {
		fieldName = column
	}

	return &DatabaseError{
		Type:    NotNullViolation,
		Message: "O campo '" + fieldName + "' é obrigatório e não pode estar vazio.",
		Field:   column,
	}
}

func parseCheckViolationError(constraint string) *DatabaseError {
	switch {
	case strings.Contains(constraint, "modalidade"):
		return &DatabaseError{
			Type:    CheckViolation,
			Message: "Modalidade inválida. Use: 'Presencial', 'Semipresencial', 'Remoto', 'PRESENCIAL', 'ONLINE' ou 'HIBRIDO'.",
			Field:   "modalidade",
		}
	case strings.Contains(constraint, "status"):
		return &DatabaseError{
			Type:    CheckViolation,
			Message: "Status inválido. Use: 'draft', 'opened', 'closed', 'canceled', 'CRIADO', 'ABERTO' ou 'ENCERRADO'.",
			Field:   "status",
		}
	default:
		return &DatabaseError{
			Type:    CheckViolation,
			Message: "Valor inválido fornecido para um dos campos.",
		}
	}
}

// GetUserFriendlyMessage retorna uma mensagem amigável baseada no tipo de erro
func (de *DatabaseError) GetUserFriendlyMessage() string {
	if de == nil {
		return "Erro interno do servidor"
	}
	return de.Message
}

// GetHTTPStatusCode retorna o código HTTP apropriado para o tipo de erro
func (de *DatabaseError) GetHTTPStatusCode() int {
	if de == nil {
		return 500
	}

	switch de.Type {
	case ForeignKeyViolation, NotNullViolation, CheckViolation:
		return 400 // Bad Request
	case UniqueViolation:
		return 409 // Conflict
	default:
		return 500 // Internal Server Error
	}
}
