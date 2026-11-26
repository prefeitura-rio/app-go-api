// Package apperrors define os erros de domínio tipados da API.
//
// Um *Error carrega o tipo do problema (não encontrado, conflito, entrada
// inválida...) e a mensagem pública, que pode ser mostrada ao cliente. A causa
// interna (erro do GORM, do driver, de um serviço externo) fica em Err, só para
// o log do servidor: Error() devolve apenas a mensagem pública.
//
// Os handlers traduzem esses erros para o status HTTP com response.FromError
// ou response.RespondError (equivalentes).
package apperrors

import (
	"errors"
	"net/http"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

// Kind classifica um erro de domínio e define o status HTTP dele.
type Kind int

const (
	// KindInvalid é entrada que viola uma regra de negócio (400).
	KindInvalid Kind = iota + 1
	// KindForbidden é operação que o usuário não pode fazer (403).
	KindForbidden
	// KindNotFound é recurso inexistente (404).
	KindNotFound
	// KindConflict é duplicidade ou estado incompatível com a operação (409).
	KindConflict
	// KindUnavailable é dependência externa indisponível (503).
	KindUnavailable
	// KindInternal é falha interna com mensagem pública estável (500).
	KindInternal

	// KindValidation is an alias of KindInvalid (same HTTP 400 mapping).
	KindValidation = KindInvalid
)

// Error é um erro de domínio com mensagem pública.
type Error struct {
	Kind    Kind
	Message string
	Err     error
	// Fields guarda, quando há, o erro de cada campo da entrada (nome do
	// campo → mensagem pública). A resposta HTTP o devolve em "fields".
	Fields map[string]string
}

// Error devolve só a mensagem pública, nunca a causa interna.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// Unwrap expõe a causa interna para errors.Is/As e para o log.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is considera iguais dois erros de domínio do mesmo tipo e mesma mensagem, o
// que permite usar valores declarados no pacote como sentinelas.
func (e *Error) Is(target error) bool {
	var t *Error
	if !errors.As(target, &t) {
		return false
	}
	return e.Kind == t.Kind && e.Message == t.Message
}

// HTTPStatus devolve o status HTTP do tipo do erro.
func (e *Error) HTTPStatus() int {
	if e == nil {
		return http.StatusInternalServerError
	}
	return kindStatus(e.Kind)
}

// New cria um erro de domínio do tipo informado.
func New(kind Kind, message string) *Error {
	return &Error{Kind: kind, Message: message}
}

// Wrap cria um erro de domínio guardando a causa interna (só para o log).
func Wrap(kind Kind, message string, cause error) *Error {
	return &Error{Kind: kind, Message: message, Err: cause}
}

// Invalid cria um erro de entrada inválida (400).
func Invalid(message string) *Error { return New(KindInvalid, message) }

// Validation is an alias of Invalid (400).
func Validation(message string) *Error { return Invalid(message) }

// ValidationFields cria um erro de entrada inválida (400) com o erro de cada
// campo, além da mensagem geral.
func ValidationFields(message string, fields map[string]string) *Error {
	err := Invalid(message)
	err.Fields = fields
	return err
}

// Forbidden cria um erro de operação proibida (403).
func Forbidden(message string) *Error { return New(KindForbidden, message) }

// NotFound cria um erro de recurso inexistente (404).
func NotFound(message string) *Error { return New(KindNotFound, message) }

// Conflict cria um erro de duplicidade ou estado incompatível (409).
func Conflict(message string) *Error { return New(KindConflict, message) }

// Unavailable cria um erro de dependência indisponível (503).
func Unavailable(message string) *Error { return New(KindUnavailable, message) }

// Internal cria um erro interno com mensagem pública estável (500).
func Internal(message string) *Error { return New(KindInternal, message) }

// As procura um erro de domínio na cadeia de err.
func As(err error) (*Error, bool) {
	var appErr *Error
	if err == nil || !errors.As(err, &appErr) {
		return nil, false
	}
	return appErr, true
}

// IsKind informa se a cadeia de err contém um erro de domínio do tipo kind.
func IsKind(err error, kind Kind) bool {
	appErr, ok := As(err)
	return ok && appErr.Kind == kind
}

const (
	// MsgInternal is the default client-safe message for unmapped 500s.
	MsgInternal       = "Erro interno do servidor"
	msgRecordNotFound = "Recurso não encontrado"
)

// Map translates err into an HTTP status, client-safe message, and optional field.
// Never returns the raw err.Error() for unknown failures.
func Map(err error) (status int, message string, field string) {
	if err == nil {
		return http.StatusInternalServerError, MsgInternal, ""
	}

	if ae, ok := As(err); ok {
		return kindStatus(ae.Kind), ae.Message, ""
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound, msgRecordNotFound, ""
	}

	dbErr := utils.ParseDatabaseError(err)
	if dbErr != nil && dbErr.Type != utils.UnknownError {
		return dbErr.GetHTTPStatusCode(), dbErr.GetUserFriendlyMessage(), dbErr.Field
	}

	return http.StatusInternalServerError, MsgInternal, ""
}

func kindStatus(kind Kind) int {
	switch kind {
	case KindInvalid: // also KindValidation
		return http.StatusBadRequest
	case KindForbidden:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindUnavailable:
		return http.StatusServiceUnavailable
	case KindInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
