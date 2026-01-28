package empregabilidade

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

type EtapaHandler struct {
	service *services.EtapaService
}

func NewEtapaHandler(service *services.EtapaService) *EtapaHandler {
	return &EtapaHandler{service: service}
}

// @Summary      Criar etapa
// @Description  Cria uma nova etapa para uma vaga
// @Tags         empregabilidade-etapas
// @Accept       json
// @Produce      json
// @Param        vagaId   path      string                 true  "ID da vaga"
// @Param        request  body      empregabilidade.Etapa  true  "Dados da etapa"
// @Success      201      {object}  empregabilidade.Etapa
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{vagaId}/etapas [post]
func (h *EtapaHandler) Create(c *gin.Context) {
	vagaID, err := uuid.Parse(c.Param("vagaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da vaga inválido"})
		return
	}

	var entity empregabilidade.Etapa
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.IDVaga = vagaID

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar etapas da vaga
// @Description  Retorna todas as etapas de uma vaga
// @Tags         empregabilidade-etapas
// @Produce      json
// @Param        vagaId   path      string  true  "ID da vaga"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{vagaId}/etapas [get]
func (h *EtapaHandler) ListByVaga(c *gin.Context) {
	vagaID, err := uuid.Parse(c.Param("vagaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da vaga inválido"})
		return
	}

	entities, err := h.service.ListByVaga(c.Request.Context(), vagaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entities})
}

// @Summary      Buscar etapa
// @Description  Retorna uma etapa pelo ID
// @Tags         empregabilidade-etapas
// @Produce      json
// @Param        vagaId  path      string  true  "ID da vaga"
// @Param        id      path      string  true  "ID da etapa"
// @Success      200     {object}  empregabilidade.Etapa
// @Failure      400     {object}  map[string]string
// @Failure      404     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{vagaId}/etapas/{id} [get]
func (h *EtapaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Etapa não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar etapa
// @Description  Atualiza uma etapa existente
// @Tags         empregabilidade-etapas
// @Accept       json
// @Produce      json
// @Param        vagaId   path      string                 true  "ID da vaga"
// @Param        id       path      string                 true  "ID da etapa"
// @Param        request  body      empregabilidade.Etapa  true  "Dados da etapa"
// @Success      200      {object}  empregabilidade.Etapa
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{vagaId}/etapas/{id} [put]
func (h *EtapaHandler) Update(c *gin.Context) {
	vagaID, err := uuid.Parse(c.Param("vagaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da vaga inválido"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.Etapa
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id
	entity.IDVaga = vagaID

	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir etapa
// @Description  Remove uma etapa pelo ID
// @Tags         empregabilidade-etapas
// @Produce      json
// @Param        vagaId  path      string  true  "ID da vaga"
// @Param        id      path      string  true  "ID da etapa"
// @Success      200     {object}  map[string]string
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{vagaId}/etapas/{id} [delete]
func (h *EtapaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Etapa excluída com sucesso"})
}
