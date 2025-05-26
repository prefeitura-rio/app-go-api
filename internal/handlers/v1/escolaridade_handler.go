package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type EscolaridadeHandler struct {
	service *services.EscolaridadeService
}

func NewEscolaridadeHandler(service *services.EscolaridadeService) *EscolaridadeHandler {
	return &EscolaridadeHandler{
		service: service,
	}
}

// @Summary      Criar escolaridade
// @Description  Cria um novo nível de escolaridade
// @Tags         escolaridades
// @Accept       json
// @Produce      json
// @Param        request  body      models.Escolaridade  true  "Dados da escolaridade"
// @Success      201      {object}  models.Escolaridade
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/escolaridades [post]
func (h *EscolaridadeHandler) Create(c *gin.Context) {
	var escolaridade models.Escolaridade
	if err := c.ShouldBindJSON(&escolaridade); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados invu00e1lidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &escolaridade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar escolaridade: " + err.Error()})
		return
	}

	escolaridade.ID = id
	c.JSON(http.StatusCreated, escolaridade)
}

// @Summary      Obter escolaridade por ID
// @Description  Retorna uma escolaridade pelo seu ID
// @Tags         escolaridades
// @Produce      json
// @Param        id   path      int  true  "ID da escolaridade"
// @Success      200  {object}  models.Escolaridade
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/escolaridades/{id} [get]
func (h *EscolaridadeHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invu00e1lido"})
		return
	}

	escolaridade, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar escolaridade: " + err.Error()})
		return
	}

	if escolaridade == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Escolaridade nu00e3o encontrada"})
		return
	}

	c.JSON(http.StatusOK, escolaridade)
}

// @Summary      Atualizar escolaridade
// @Description  Atualiza os dados de uma escolaridade existente
// @Tags         escolaridades
// @Accept       json
// @Produce      json
// @Param        id       path      int                 true  "ID da escolaridade"
// @Param        request  body      models.Escolaridade  true  "Dados da escolaridade"
// @Success      200      {object}  models.Escolaridade
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/escolaridades/{id} [put]
func (h *EscolaridadeHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invu00e1lido"})
		return
	}

	var escolaridade models.Escolaridade
	if err := c.ShouldBindJSON(&escolaridade); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados invu00e1lidos: " + err.Error()})
		return
	}

	escolaridade.ID = id
	if err := h.service.Update(c.Request.Context(), &escolaridade); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar escolaridade: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, escolaridade)
}

// @Summary      Excluir escolaridade
// @Description  Remove uma escolaridade pelo ID
// @Tags         escolaridades
// @Produce      json
// @Param        id   path      int  true  "ID da escolaridade"
// @Success      200  {object}  object
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/escolaridades/{id} [delete]
func (h *EscolaridadeHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invu00e1lido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir escolaridade: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Escolaridade excluu00edda com sucesso"})
}

// @Summary      Listar escolaridades
// @Description  Retorna uma lista paginada de escolaridades
// @Tags         escolaridades
// @Produce      json
// @Param        page     query     int  false  "Nu00famero da pu00e1gina (default: 1)"
// @Param        pageSize query     int  false  "Tamanho da pu00e1gina (default: 10)"
// @Success      200      {object}  object
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/escolaridades [get]
func (h *EscolaridadeHandler) List(c *gin.Context) {
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
	
	escolaridades, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar escolaridades: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": escolaridades,
		"meta": gin.H{
			"page": page,
			"page_size": pageSize,
			"total": total,
		},
	})
}
