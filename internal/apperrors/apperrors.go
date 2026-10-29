// Package apperrors define os erros de domínio tipados da API.
//
// Um *Error carrega o tipo do problema (não encontrado, conflito, entrada
// inválida...) e a mensagem pública, que pode ser mostrada ao cliente. A causa
// interna (erro do GORM, do driver, de um serviço externo) fica em Err, só para
// o log do servidor: Error() devolve apenas a mensagem pública.
//
// Os handlers traduzem esses erros para o status HTTP com response.FromError.
package apperrors

import (
	"errors"
	"net/http"
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
)

// Error é um erro de domínio com mensagem pública.
type Error struct {
	Kind    Kind
	Message string
	Err     error
}

// Error devolve só a mensagem pública, nunca a causa interna.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap expõe a causa interna para errors.Is/As e para o log.
func (e *Error) Unwrap() error {
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
	switch e.Kind {
	case KindInvalid:
		return http.StatusBadRequest
	case KindForbidden:
		return http.StatusForbidden
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
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

// Forbidden cria um erro de operação proibida (403).
func Forbidden(message string) *Error { return New(KindForbidden, message) }

// NotFound cria um erro de recurso inexistente (404).
func NotFound(message string) *Error { return New(KindNotFound, message) }

// Conflict cria um erro de duplicidade ou estado incompatível (409).
func Conflict(message string) *Error { return New(KindConflict, message) }

// Unavailable cria um erro de dependência indisponível (503).
func Unavailable(message string) *Error { return New(KindUnavailable, message) }

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
