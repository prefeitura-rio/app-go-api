package apperrors

import (
	"errors"
	"net/http"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

// Kind classifies a domain error for HTTP mapping.
type Kind int

const (
	KindNotFound Kind = iota + 1
	KindValidation
	KindConflict
	KindForbidden
	KindUnavailable
	KindInternal
)

// Error is a typed domain error with a client-safe Message.
type Error struct {
	Kind    Kind
	Message string
	cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// NotFound returns a 404-mapped domain error.
func NotFound(msg string) *Error {
	return &Error{Kind: KindNotFound, Message: msg}
}

// Validation returns a 400-mapped domain error.
func Validation(msg string) *Error {
	return &Error{Kind: KindValidation, Message: msg}
}

// Conflict returns a 409-mapped domain error.
func Conflict(msg string) *Error {
	return &Error{Kind: KindConflict, Message: msg}
}

// Forbidden returns a 403-mapped domain error.
func Forbidden(msg string) *Error {
	return &Error{Kind: KindForbidden, Message: msg}
}

// Unavailable returns a 503-mapped domain error (temporary dependency failure).
func Unavailable(msg string) *Error {
	return &Error{Kind: KindUnavailable, Message: msg}
}

// Internal returns a 500-mapped domain error with a client-safe message.
func Internal(msg string) *Error {
	return &Error{Kind: KindInternal, Message: msg}
}

// Wrap attaches a cause to a domain error of the given kind.
func Wrap(kind Kind, msg string, cause error) *Error {
	return &Error{Kind: kind, Message: msg, cause: cause}
}

// IsKind reports whether err (or any wrapped error) is an *Error of the given kind.
func IsKind(err error, kind Kind) bool {
	var ae *Error
	if !errors.As(err, &ae) {
		return false
	}
	return ae.Kind == kind
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

	var ae *Error
	if errors.As(err, &ae) && ae != nil {
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
	case KindNotFound:
		return http.StatusNotFound
	case KindValidation:
		return http.StatusBadRequest
	case KindConflict:
		return http.StatusConflict
	case KindForbidden:
		return http.StatusForbidden
	case KindUnavailable:
		return http.StatusServiceUnavailable
	case KindInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
