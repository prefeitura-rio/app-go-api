package utils

import (
	"errors"
	"strings"

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

// ParseDatabaseError analisa erros do PostgreSQL e retorna mensagens amigáveis
func ParseDatabaseError(err error) *DatabaseError {
	if err == nil {
		return nil
	}

	// Verifica se é um erro do PostgreSQL
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		// In production, hide detailed error messages
		return &DatabaseError{
			Type:    UnknownError,
			Message: "Erro interno do servidor",
		}
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
		return &DatabaseError{
			Type:    UnknownError,
			Message: "Erro interno do servidor",
		}
	}
}

func parseForeignKeyError(pqErr *pq.Error) *DatabaseError {
	constraint := pqErr.Constraint
	
	// Mapear constraints para mensagens amigáveis
	switch {
	case strings.Contains(constraint, "orgao_id"):
		return &DatabaseError{
			Type:    ForeignKeyViolation,
			Message: "O órgão especificado não existe. Verifique se o ID do órgão está correto.",
			Field:   "orgao_id",
		}
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

func parseUniqueViolationError(pqErr *pq.Error) *DatabaseError {
	constraint := pqErr.Constraint
	
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

func parseNotNullError(pqErr *pq.Error) *DatabaseError {
	column := pqErr.Column
	
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

func parseCheckViolationError(pqErr *pq.Error) *DatabaseError {
	constraint := pqErr.Constraint
	
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
