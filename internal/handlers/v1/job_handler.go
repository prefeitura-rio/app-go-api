package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type JobHandler struct {
	service services.JobServiceInterface
}

func NewJobHandler(service services.JobServiceInterface) *JobHandler {
	return &JobHandler{
		service: service,
	}
}

// @Summary      Obter status de job
// @Description  Retorna o status e progresso de um job em execução ou concluído
// @Tags         jobs
// @Produce      json
// @Param        jobId path      string  true  "ID do job (UUID)"
// @Success      200   {object}  models.Job
// @Failure      400   {object}  models.ErrorResponse
// @Failure      404   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /api/v1/jobs/{jobId}/status [get]
func (h *JobHandler) GetStatus(c *gin.Context) {
	jobIDStr := c.Param("jobId")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do job inválido"})
		return
	}

	job, err := h.service.GetByID(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar job: " + err.Error()})
		return
	}

	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job não encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    job,
	})
}
