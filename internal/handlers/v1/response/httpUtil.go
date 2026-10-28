package response

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// DTOs / Contratos Genéricos

type ErrorResponse struct {
	Error string `json:"error" example:"Mensagem de erro aqui"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"Operação realizada com sucesso"`
}

type Paginated[T any] struct {
	Data     []T   `json:"data"`
	Total    int64 `json:"total" example:"42"`
	Page     int   `json:"page" example:"1"`
	PageSize int   `json:"pageSize" example:"20"`
}

type ListHabilidadesPaginatedResponse = Paginated[empregabilidade.Habilidade]

type ListAreasAtuacaoPaginatedResponse = Paginated[empregabilidade.AreaAtuacao]

type ListComportamentoAtitudesPaginatedResponse = Paginated[empregabilidade.ComportamentoAtitudes]

// Helpers HTTP

// Error envia uma resposta JSON padronizada para erros.
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error: message,
	})
}

// ErrorWithField is the additive JSON shape used when a DB constraint maps to a field.
type ErrorWithField struct {
	Error string `json:"error" example:"Mensagem de erro aqui"`
	Field string `json:"field,omitempty" example:"email"`
}

// RespondError logs the full error (with trace_id when available) and writes a
// stable client-facing JSON body via apperrors.Map. Never exposes internal details.
//
// An optional fallbackMessage is used only for unmapped 500s (replacing
// apperrors.MsgInternal). Domain errors and mapped DB constraints keep their
// own client-safe messages; fallback is ignored for 4xx.
func RespondError(c *gin.Context, err error, fallbackMessage ...string) {
	if err == nil {
		return
	}

	traceID := ""
	if span := trace.SpanFromContext(c.Request.Context()); span != nil {
		sc := span.SpanContext()
		if sc.IsValid() {
			traceID = sc.TraceID().String()
		}
	}

	method, path := "", ""
	if c.Request != nil {
		method = c.Request.Method
		path = c.Request.URL.Path
	}
	log.Printf("request error method=%s path=%s trace_id=%s err=%v", method, path, traceID, err)

	status, message, field := apperrors.Map(err)
	if status == http.StatusInternalServerError &&
		message == apperrors.MsgInternal &&
		len(fallbackMessage) > 0 &&
		fallbackMessage[0] != "" {
		message = fallbackMessage[0]
	}
	if field != "" {
		c.JSON(status, ErrorWithField{Error: message, Field: field})
		return
	}
	Error(c, status, message)
}

// Success envia uma resposta JSON padronizada para mensagens de sucesso.
func Success(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, SuccessResponse{
		Message: message,
	})
}

// PaginatedJSON envia uma resposta JSON padronizada com estrutura de paginação.
func PaginatedJSON[T any](c *gin.Context, statusCode int, data []T, total int64, page, pageSize int) {
	if data == nil {
		data = make([]T, 0)
	}

	c.JSON(statusCode, Paginated[T]{
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
