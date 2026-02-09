package httputil

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 1000
)

func Pagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 || pageSize > MaxPageSize {
		pageSize = DefaultPageSize
	}

	return page, pageSize
}

func Offset(page, pageSize int) int {
	return (page - 1) * pageSize
}

func ParseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		BadRequest(c, "ID inválido")
		return uuid.UUID{}, false
	}
	return id, true
}

func ParseIntID(c *gin.Context, param string) (int, bool) {
	id, err := strconv.Atoi(c.Param(param))
	if err != nil {
		BadRequest(c, "ID inválido")
		return 0, false
	}
	return id, true
}
