package response

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

const (
	defaultInternalMessage = "Erro interno do servidor"
	defaultNotFoundMessage = "Registro não encontrado"
)

// FromError traduz err para a resposta HTTP, sem expor detalhe interno:
//
//   - erro de domínio (apperrors.Error): status do tipo e a mensagem pública;
//   - gorm.ErrRecordNotFound: 404;
//   - violação de constraint do PostgreSQL: 409 (unicidade) ou 400 (FK, not
//     null, check), com a mensagem amigável de utils.ParseDatabaseError;
//   - qualquer outro erro: 500 com internalMessage, e o erro original só no log.
func FromError(c *gin.Context, err error, internalMessage string) {
	if appErr, ok := apperrors.As(err); ok {
		if appErr.Err != nil {
			logError(c, appErr.Message, appErr.Err)
		}
		Error(c, appErr.HTTPStatus(), appErr.Message)
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		Error(c, http.StatusNotFound, defaultNotFoundMessage)
		return
	}

	if dbErr := utils.ParseDatabaseError(err); dbErr != nil && dbErr.Type != utils.UnknownError {
		Error(c, dbErr.GetHTTPStatusCode(), dbErr.GetUserFriendlyMessage())
		return
	}

	InternalError(c, err, internalMessage)
}

// InternalError responde 500 com internalMessage e registra err no log do
// servidor. Use quando o erro é sempre falha do servidor (arquivo, fila, banco).
func InternalError(c *gin.Context, err error, internalMessage string) {
	if internalMessage == "" {
		internalMessage = defaultInternalMessage
	}
	logError(c, internalMessage, err)
	Error(c, http.StatusInternalServerError, internalMessage)
}

func logError(c *gin.Context, message string, err error) {
	// FullPath é o molde da rota (/curriculo/:cpf), não a URL com o CPF.
	method, path := "", c.FullPath()
	if c.Request != nil {
		method = c.Request.Method
	}
	log.Printf("[http-error] %s %s: %s: %v", method, path, message, err)
}
