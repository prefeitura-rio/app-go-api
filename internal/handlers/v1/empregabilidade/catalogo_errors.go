package empregabilidade

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
)

// respondCatalogoError maps DB constraint errors via RespondError; unmapped
// failures use internalMessage as the stable 500 body (never err.Error()).
func respondCatalogoError(c *gin.Context, err error, internalMessage string) {
	response.RespondError(c, err, internalMessage)
}

// respondCatalogoMutationError maps gorm.ErrRecordNotFound to 404 with a
// resource-specific message; remaining errors follow respondCatalogoError.
func respondCatalogoMutationError(c *gin.Context, err error, notFoundMessage, internalMessage string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, notFoundMessage)
		return
	}
	respondCatalogoError(c, err, internalMessage)
}
