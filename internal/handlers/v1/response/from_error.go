package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
)

// FromError traduz err para a resposta HTTP sem expor detalhe interno.
// É um wrapper de RespondError: mesma mapeação (apperrors, GORM, PG) e o
// internalMessage só substitui a mensagem genérica em 500s não mapeados.
func FromError(c *gin.Context, err error, internalMessage string) {
	if err == nil {
		InternalError(c, err, internalMessage)
		return
	}
	RespondError(c, err, internalMessage)
}

// InternalError responde 500 com internalMessage e registra err no log do
// servidor. Use quando o erro é sempre falha do servidor (arquivo, fila, banco).
func InternalError(c *gin.Context, err error, internalMessage string) {
	if internalMessage == "" {
		internalMessage = apperrors.MsgInternal
	}
	logRespondError(c, err)
	Error(c, http.StatusInternalServerError, internalMessage)
}
