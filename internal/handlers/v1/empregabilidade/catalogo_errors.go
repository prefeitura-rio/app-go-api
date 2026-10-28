package empregabilidade

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
)

// respondCatalogoMutationError maps gorm.ErrRecordNotFound to 404 with a
// resource-specific message; all other errors go through response.RespondError.
func respondCatalogoMutationError(c *gin.Context, err error, notFoundMessage string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, notFoundMessage)
		return
	}
	response.RespondError(c, err)
}
