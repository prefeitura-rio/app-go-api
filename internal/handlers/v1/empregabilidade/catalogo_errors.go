package empregabilidade

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

// respondCatalogoError responde erros dos handlers do catálogo (habilidades,
// áreas de atuação, comportamentos/atitudes) no formato response.ErrorResponse.
// Violações de constraint do banco viram erro do cliente (unicidade → 409;
// FK/not-null/check → 400), como em handleLookupError; qualquer outro erro
// continua 500 com a mensagem genérica da operação, sem expor o detalhe interno.
func respondCatalogoError(c *gin.Context, err error, internalMessage string) {
	dbErr := utils.ParseDatabaseError(err)
	if dbErr == nil || dbErr.Type == utils.UnknownError {
		response.Error(c, http.StatusInternalServerError, internalMessage)
		return
	}
	response.Error(c, dbErr.GetHTTPStatusCode(), dbErr.GetUserFriendlyMessage())
}

// respondCatalogoMutationError responde erros de PUT/DELETE por ID no catálogo:
// registro inexistente (gorm.ErrRecordNotFound, devolvido pelo repositório
// quando nenhuma linha é afetada) vira 404 com notFoundMessage, a mesma do GET
// por ID; os demais erros seguem respondCatalogoError.
func respondCatalogoMutationError(c *gin.Context, err error, notFoundMessage, internalMessage string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, notFoundMessage)
		return
	}
	respondCatalogoError(c, err, internalMessage)
}
