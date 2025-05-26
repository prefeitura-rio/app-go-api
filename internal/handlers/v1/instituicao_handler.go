package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type InstituicaoHandler struct {
	service *services.InstituicaoService
}

func NewInstituicaoHandler(service *services.InstituicaoService) *InstituicaoHandler {
	return &InstituicaoHandler{
		service: service,
	}
}

// @Summary      Criar instituição de ensino
// @Description  Cria uma nova instituição de ensino
// @Tags         instituicoes
// @Accept       json
// @Produce      json
// @Param        request  body      models.InstituicaoEnsino  true  "Dados da instituição"
// @Success      201      {object}  models.InstituicaoEnsino
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/instituicoes [post]
func (h *InstituicaoHandler) Create(c *gin.Context) {
	var instituicao models.InstituicaoEnsino
	if err := c.ShouldBindJSON(&instituicao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &instituicao)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar instituição: " + err.Error()})
		return
	}

	instituicao.ID = id
	c.JSON(http.StatusCreated, instituicao)
}

// @Summary      Obter instituição de ensino por ID
// @Description  Retorna uma instituição pelo seu ID
// @Tags         instituicoes
// @Produce      json
// @Param        id   path      int  true  "ID da instituição"
// @Success      200  {object}  models.InstituicaoEnsino
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/instituicoes/{id} [get]
func (h *InstituicaoHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	instituicao, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar instituição: " + err.Error()})
		return
	}

	if instituicao == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instituição não encontrada"})
		return
	}

	c.JSON(http.StatusOK, instituicao)
}

// @Summary      Atualizar instituição de ensino
// @Description  Atualiza os dados de uma instituição existente
// @Tags         instituicoes
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "ID da instituição"
// @Param        request  body      models.InstituicaoEnsino  true  "Dados da instituição"
// @Success      200      {object}  models.InstituicaoEnsino
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/instituicoes/{id} [put]
func (h *InstituicaoHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var instituicao models.InstituicaoEnsino
	if err := c.ShouldBindJSON(&instituicao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	instituicao.ID = id
	if err := h.service.Update(c.Request.Context(), &instituicao); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar instituição: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, instituicao)
}

// @Summary      Excluir instituição de ensino
// @Description  Remove uma instituição pelo ID
// @Tags         instituicoes
// @Produce      json
// @Param        id   path      int  true  "ID da instituição"
// @Success      200  {object}  object
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/instituicoes/{id} [delete]
func (h *InstituicaoHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir instituição: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instituição excluída com sucesso"})
}

// @Summary      Listar instituições de ensino
// @Description  Retorna uma lista paginada de instituições
// @Tags         instituicoes
// @Produce      json
// @Param        page     query     int  false  "Número da página (default: 1)"
// @Param        pageSize query     int  false  "Tamanho da página (default: 10)"
// @Success      200      {object}  object
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/instituicoes [get]
func (h *InstituicaoHandler) List(c *gin.Context) {
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
	
	instituicoes, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar instituições: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": instituicoes,
		"meta": gin.H{
			"page": page,
			"page_size": pageSize,
			"total": total,
		},
	})
}
