package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type OportunidadeMEIHandler struct {
	service *services.OportunidadeMEIService
}

func NewOportunidadeMEIHandler(service *services.OportunidadeMEIService) *OportunidadeMEIHandler {
	return &OportunidadeMEIHandler{
		service: service,
	}
}

// @Summary      Criar oportunidade MEI
// @Description  Cria uma nova oportunidade MEI publicada
// @Tags         oportunidades-mei
// @Accept       json
// @Produce      json
// @Param        request  body      models.OportunidadeMEI  true  "Dados da oportunidade"
// @Success      201      {object}  models.OportunidadeMEI
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei [post]
func (h *OportunidadeMEIHandler) Create(c *gin.Context) {
	var oportunidade models.OportunidadeMEI
	if err := c.ShouldBindJSON(&oportunidade); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &oportunidade, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oportunidade.ID = id
	c.JSON(http.StatusCreated, oportunidade)
}

// @Summary      Criar oportunidade MEI como rascunho
// @Description  Cria uma nova oportunidade MEI em status de rascunho
// @Tags         oportunidades-mei
// @Accept       json
// @Produce      json
// @Param        request  body      models.OportunidadeMEI  true  "Dados da oportunidade"
// @Success      201      {object}  models.OportunidadeMEI
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/draft [post]
func (h *OportunidadeMEIHandler) CreateDraft(c *gin.Context) {
	var oportunidade models.OportunidadeMEI
	if err := c.ShouldBindJSON(&oportunidade); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &oportunidade, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oportunidade.ID = id
	c.JSON(http.StatusCreated, oportunidade)
}

// @Summary      Obter oportunidade MEI por ID
// @Description  Retorna uma oportunidade MEI pelo seu ID
// @Tags         oportunidades-mei
// @Produce      json
// @Param        id   path      int  true  "ID da oportunidade"
// @Success      200  {object}  models.OportunidadeMEI
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id} [get]
func (h *OportunidadeMEIHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	oportunidade, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar oportunidade: " + err.Error()})
		return
	}

	if oportunidade == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Oportunidade não encontrada"})
		return
	}

	c.JSON(http.StatusOK, oportunidade)
}

// @Summary      Atualizar oportunidade MEI
// @Description  Atualiza os dados de uma oportunidade MEI existente mantendo seu status atual
// @Tags         oportunidades-mei
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "ID da oportunidade"
// @Param        request  body      models.OportunidadeMEI  true  "Dados da oportunidade"
// @Success      200      {object}  models.OportunidadeMEI
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id} [put]
func (h *OportunidadeMEIHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var oportunidade models.OportunidadeMEI
	if err := c.ShouldBindJSON(&oportunidade); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	oportunidade.ID = id
	if err := h.service.Update(c.Request.Context(), &oportunidade); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, oportunidade)
}

// @Summary      Excluir oportunidade MEI
// @Description  Remove uma oportunidade MEI pelo ID (soft delete)
// @Tags         oportunidades-mei
// @Produce      json
// @Param        id   path      int  true  "ID da oportunidade"
// @Success      200  {object}  object
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id} [delete]
func (h *OportunidadeMEIHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir oportunidade: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Oportunidade excluída com sucesso"})
}

// @Summary      Publicar oportunidade MEI
// @Description  Publica uma oportunidade MEI que estava em rascunho
// @Tags         oportunidades-mei
// @Produce      json
// @Param        id   path      int  true  "ID da oportunidade"
// @Success      200  {object}  models.OportunidadeMEI
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id}/publish [put]
func (h *OportunidadeMEIHandler) Publish(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Publish(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oportunidade, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar oportunidade: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, oportunidade)
}

// @Summary      Listar oportunidades MEI
// @Description  Retorna uma lista paginada de oportunidades MEI ativas
// @Tags         oportunidades-mei
// @Produce      json
// @Param        page      query     int     false  "Número da página (default: 1)"
// @Param        pageSize  query     int     false  "Tamanho da página (default: 10, max: 1000)"
// @Param        orgaoId   query     int     false  "Filtrar por órgão"
// @Param        status    query     string  false  "Filtrar por status (draft, active, expired)"
// @Param        titulo    query     string  false  "Buscar por título (case-insensitive)"
// @Success      200       {object}  object
// @Failure      500       {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei [get]
func (h *OportunidadeMEIHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	orgaoID := c.Query("orgaoId")
	status := c.Query("status")
	titulo := c.Query("titulo")

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	var oportunidades []*models.OportunidadeMEI
	var total int
	var err error

	// Construir filtros
	filters := make(map[string]interface{})

	if orgaoID != "" {
		filters["orgao_id"] = orgaoID
	}

	if status != "" {
		switch status {
		case "draft":
			filters["status"] = models.StatusOportunidadeDraft
		case "active":
			filters["status"] = models.StatusOportunidadeActive
		case "expired":
			filters["status"] = models.StatusOportunidadeExpired
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Status inválido"})
			return
		}
	} else {
		// Se não especificou status, listar apenas ativas por padrão
		filters["status"] = models.StatusOportunidadeActive
	}

	oportunidades, total, err = h.service.List(c.Request.Context(), filters, titulo, page, pageSize)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar oportunidades: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": oportunidades,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// @Summary      Listar rascunhos de oportunidades MEI
// @Description  Retorna uma lista paginada de oportunidades MEI em rascunho
// @Tags         oportunidades-mei
// @Produce      json
// @Param        page     query     int  false  "Número da página (default: 1)"
// @Param        pageSize query     int  false  "Tamanho da página (default: 10, max: 1000)"
// @Success      200      {object}  object
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/drafts [get]
func (h *OportunidadeMEIHandler) ListDrafts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	oportunidades, total, err := h.service.ListDrafts(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar rascunhos: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": oportunidades,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}
