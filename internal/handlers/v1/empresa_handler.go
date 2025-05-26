package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type EmpresaHandler struct {
	service *services.EmpresaService
}

func NewEmpresaHandler(service *services.EmpresaService) *EmpresaHandler {
	return &EmpresaHandler{
		service: service,
	}
}

// @Summary      Criar empresa
// @Description  Cria uma nova empresa
// @Tags         empresas
// @Accept       json
// @Produce      json
// @Param        request  body      models.Empresa  true  "Dados da empresa"
// @Success      201      {object}  models.Empresa
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/empresas [post]
func (h *EmpresaHandler) Create(c *gin.Context) {
	var empresa models.Empresa
	if err := c.ShouldBindJSON(&empresa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &empresa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar empresa: " + err.Error()})
		return
	}

	empresa.ID = id
	c.JSON(http.StatusCreated, empresa)
}

// @Summary      Obter empresa por ID
// @Description  Retorna uma empresa pelo seu ID
// @Tags         empresas
// @Produce      json
// @Param        id   path      int  true  "ID da empresa"
// @Success      200  {object}  models.Empresa
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/empresas/{id} [get]
func (h *EmpresaHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	empresa, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar empresa: " + err.Error()})
		return
	}

	if empresa == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada"})
		return
	}

	c.JSON(http.StatusOK, empresa)
}

// @Summary      Atualizar empresa
// @Description  Atualiza os dados de uma empresa existente
// @Tags         empresas
// @Accept       json
// @Produce      json
// @Param        id       path      int             true  "ID da empresa"
// @Param        request  body      models.Empresa  true  "Dados da empresa"
// @Success      200      {object}  models.Empresa
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/empresas/{id} [put]
func (h *EmpresaHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var empresa models.Empresa
	if err := c.ShouldBindJSON(&empresa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	empresa.ID = id
	if err := h.service.Update(c.Request.Context(), &empresa); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar empresa: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, empresa)
}

// @Summary      Excluir empresa
// @Description  Remove uma empresa pelo ID
// @Tags         empresas
// @Produce      json
// @Param        id   path      int  true  "ID da empresa"
// @Success      200  {object}  object
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/empresas/{id} [delete]
func (h *EmpresaHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir empresa: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Empresa excluída com sucesso"})
}

// @Summary      Listar empresas
// @Description  Retorna uma lista paginada de empresas
// @Tags         empresas
// @Produce      json
// @Param        page     query     int  false  "Número da página (default: 1)"
// @Param        pageSize query     int  false  "Tamanho da página (default: 10)"
// @Success      200      {object}  object
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/empresas [get]
func (h *EmpresaHandler) List(c *gin.Context) {
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
	
	empresas, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar empresas: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": empresas,
		"meta": gin.H{
			"page": page,
			"page_size": pageSize,
			"total": total,
		},
	})
}
