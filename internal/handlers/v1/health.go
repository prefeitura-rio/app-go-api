package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary      Verificar saúde da API
// @Description  Retorna o status de saúde da API v1
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /health [get]
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"version": "v1",
	})
}