package v1

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type CNAEHandler struct {
	service *services.CNAEService
}

func NewCNAEHandler(service *services.CNAEService) *CNAEHandler {
	return &CNAEHandler{
		service: service,
	}
}

// @Summary      Obter CNAE por código
// @Description  Retorna um CNAE pelo seu código
// @Tags         cnaes
// @Produce      json
// @Param        codigo   path      string  true  "Código do CNAE"
// @Success      200  {object}  models.CNAE
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/cnaes/{codigo} [get]
func (h *CNAEHandler) GetByCodigo(c *gin.Context) {
	// Wildcard params come with leading slash, strip it
	codigo := strings.TrimPrefix(c.Param("codigo"), "/")

	cnae, err := h.service.GetByCodigo(c.Request.Context(), codigo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar CNAE: " + err.Error()})
		return
	}

	if cnae == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CNAE não encontrado"})
		return
	}

	c.JSON(http.StatusOK, cnae)
}

// @Summary      Listar CNAEs
// @Description  Retorna uma lista paginada de CNAEs
// @Tags         cnaes
// @Produce      json
// @Param        page     query     int     false  "Número da página (default: 1)"
// @Param        pageSize query     int     false  "Tamanho da página (default: 100, max: 1000)"
// @Param        ocupacao query     string  false  "Filtrar por ocupação"
// @Success      200      {object}  object
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/cnaes [get]
func (h *CNAEHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))
	ocupacao := c.Query("ocupacao")

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	cnaes, total, err := h.service.List(c.Request.Context(), ocupacao, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar CNAEs: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": cnaes,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// @Summary      Criar CNAE
// @Description  Cria um novo CNAE (admin)
// @Tags         cnaes-admin
// @Accept       json
// @Produce      json
// @Param        request  body      models.CNAE  true  "Dados do CNAE"
// @Success      201      {object}  models.CNAE
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/admin/cnaes [post]
func (h *CNAEHandler) Create(c *gin.Context) {
	var cnae models.CNAE
	if err := c.ShouldBindJSON(&cnae); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	if err := h.service.Create(c.Request.Context(), &cnae); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cnae)
}

// @Summary      Atualizar CNAE
// @Description  Atualiza um CNAE existente (admin)
// @Tags         cnaes-admin
// @Accept       json
// @Produce      json
// @Param        codigo   path      string       true  "Código do CNAE"
// @Param        request  body      models.CNAE  true  "Dados do CNAE"
// @Success      200      {object}  models.CNAE
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/admin/cnaes/{codigo} [put]
func (h *CNAEHandler) Update(c *gin.Context) {
	// Wildcard params come with leading slash, strip it
	codigo := strings.TrimPrefix(c.Param("codigo"), "/")

	var cnae models.CNAE
	if err := c.ShouldBindJSON(&cnae); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	cnae.Codigo = codigo

	if err := h.service.Update(c.Request.Context(), &cnae); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cnae)
}

// @Summary      Excluir CNAE
// @Description  Remove um CNAE pelo código (admin)
// @Tags         cnaes-admin
// @Produce      json
// @Param        codigo  path      string  true  "Código do CNAE"
// @Success      200     {object}  object
// @Failure      400     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Router       /api/v1/admin/cnaes/{codigo} [delete]
func (h *CNAEHandler) Delete(c *gin.Context) {
	// Wildcard params come with leading slash, strip it
	codigo := strings.TrimPrefix(c.Param("codigo"), "/")

	if err := h.service.Delete(c.Request.Context(), codigo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir CNAE: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "CNAE excluído com sucesso"})
}
