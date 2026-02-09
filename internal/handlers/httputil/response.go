package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

func Deleted(c *gin.Context, entityName string) {
	c.JSON(http.StatusOK, gin.H{"message": entityName + " excluído(a) com sucesso"})
}

func Paginated(c *gin.Context, data interface{}, page, pageSize, total int) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

func Error(c *gin.Context, err error) {
	status := apperrors.HTTPStatus(err)
	resp := gin.H{"error": err.Error()}
	if field := apperrors.GetField(err); field != "" {
		resp["field"] = field
	}
	c.JSON(status, resp)
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}
