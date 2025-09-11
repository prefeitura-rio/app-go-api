package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type OrgaoHandler struct {
	service *services.OrgaoService
}

func NewOrgaoHandler(service *services.OrgaoService) *OrgaoHandler {
	return &OrgaoHandler{
		service: service,
	}
}

// @Summary      Criar órgão
// @Description  Cria um novo órgão
// @Tags         orgaos
// @Accept       json
// @Produce      json
// @Param        request  body      models.Orgao  true  "Dados do órgão"
// @Success      201      {object}  models.Orgao
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/orgaos [post]
func (h *OrgaoHandler) Create(c *gin.Context) {
	var orgao models.Orgao
	if err := c.ShouldBindJSON(&orgao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &orgao)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar órgão: " + err.Error()})
		return
	}

	orgao.ID = id
	c.JSON(http.StatusCreated, orgao)
}

// @Summary      Obter órgão por ID
// @Description  Retorna um órgão pelo seu ID
// @Tags         orgaos
// @Produce      json
// @Param        id   path      int  true  "ID do órgão"
// @Success      200  {object}  models.Orgao
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/orgaos/{id} [get]
func (h *OrgaoHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	orgao, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar órgão: " + err.Error()})
		return
	}

	if orgao == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Órgão não encontrado"})
		return
	}

	c.JSON(http.StatusOK, orgao)
}

// @Summary      Atualizar órgão
// @Description  Atualiza os dados de um órgão existente
// @Tags         orgaos
// @Accept       json
// @Produce      json
// @Param        id       path      int         true  "ID do órgão"
// @Param        request  body      models.Orgao  true  "Dados do órgão"
// @Success      200      {object}  models.Orgao
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/orgaos/{id} [put]
func (h *OrgaoHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var orgao models.Orgao
	if err := c.ShouldBindJSON(&orgao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	orgao.ID = id
	if err := h.service.Update(c.Request.Context(), &orgao); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar órgão: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, orgao)
}

// @Summary      Excluir órgão
// @Description  Remove um órgão pelo ID
// @Tags         orgaos
// @Produce      json
// @Param        id   path      int  true  "ID do órgão"
// @Success      200  {object}  object
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/orgaos/{id} [delete]
func (h *OrgaoHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir órgão: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Órgão excluído com sucesso"})
}

// @Summary      Listar órgãos
// @Description  Retorna uma lista paginada de órgãos
// @Tags         orgaos
// @Produce      json
// @Param        page     query     int  false  "Número da página (default: 1)"
// @Param        pageSize query     int  false  "Tamanho da página (default: 10)"
// @Success      200      {object}  object
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/orgaos [get]
func (h *OrgaoHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	
	if page < 1 {
		page = 1
	}
	
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}
	
	// Filtros
	filter := make(map[string]interface{})
	
	orgaos, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar órgãos: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": orgaos,
		"meta": gin.H{
			"page": page,
			"page_size": pageSize,
			"total": total,
		},
	})
}
