package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type AcessibilidadeHandler struct {
	service *services.AcessibilidadeService
}

func NewAcessibilidadeHandler(service *services.AcessibilidadeService) *AcessibilidadeHandler {
	return &AcessibilidadeHandler{
		service: service,
	}
}

// @Summary      Criar acessibilidade
// @Description  Cria um novo tipo de acessibilidade
// @Tags         acessibilidades
// @Accept       json
// @Produce      json
// @Param        request  body      models.Acessibilidade  true  "Dados da acessibilidade"
// @Success      201      {object}  models.Acessibilidade
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/acessibilidades [post]
func (h *AcessibilidadeHandler) Create(c *gin.Context) {
	var acessibilidade models.Acessibilidade
	if err := c.ShouldBindJSON(&acessibilidade); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &acessibilidade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar acessibilidade: " + err.Error()})
		return
	}

	acessibilidade.ID = id
	c.JSON(http.StatusCreated, acessibilidade)
}

// @Summary      Obter acessibilidade por ID
// @Description  Retorna uma acessibilidade pelo seu ID
// @Tags         acessibilidades
// @Produce      json
// @Param        id   path      int  true  "ID da acessibilidade"
// @Success      200  {object}  models.Acessibilidade
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/acessibilidades/{id} [get]
func (h *AcessibilidadeHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	acessibilidade, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar acessibilidade: " + err.Error()})
		return
	}

	if acessibilidade == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Acessibilidade não encontrada"})
		return
	}

	c.JSON(http.StatusOK, acessibilidade)
}

// @Summary      Atualizar acessibilidade
// @Description  Atualiza os dados de uma acessibilidade existente
// @Tags         acessibilidades
// @Accept       json
// @Produce      json
// @Param        id       path      int                 true  "ID da acessibilidade"
// @Param        request  body      models.Acessibilidade  true  "Dados da acessibilidade"
// @Success      200      {object}  models.Acessibilidade
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/acessibilidades/{id} [put]
func (h *AcessibilidadeHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var acessibilidade models.Acessibilidade
	if err := c.ShouldBindJSON(&acessibilidade); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	acessibilidade.ID = id
	if err := h.service.Update(c.Request.Context(), &acessibilidade); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar acessibilidade: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, acessibilidade)
}

// @Summary      Excluir acessibilidade
// @Description  Remove uma acessibilidade pelo ID
// @Tags         acessibilidades
// @Produce      json
// @Param        id   path      int  true  "ID da acessibilidade"
// @Success      200  {object}  object
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/acessibilidades/{id} [delete]
func (h *AcessibilidadeHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir acessibilidade: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Acessibilidade excluída com sucesso"})
}

// @Summary      Listar acessibilidades
// @Description  Retorna uma lista paginada de acessibilidades
// @Tags         acessibilidades
// @Produce      json
// @Param        page     query     int  false  "Número da página (default: 1)"
// @Param        pageSize query     int  false  "Tamanho da página (default: 10)"
// @Success      200      {object}  object
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/acessibilidades [get]
func (h *AcessibilidadeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	
	if page < 1 {
		page = 1
	}
	
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	
	// Filtros
	filter := make(map[string]interface{})
	
	acessibilidades, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar acessibilidades: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": acessibilidades,
		"meta": gin.H{
			"page": page,
			"page_size": pageSize,
			"total": total,
		},
	})
}
