package v1

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

type EmpregoHandler struct {
	service *services.EmpregoService
}

func NewEmpregoHandler(service *services.EmpregoService) *EmpregoHandler {
	return &EmpregoHandler{
		service: service,
	}
}

// @Summary      Criar emprego
// @Description  Cria uma nova oportunidade de emprego
// @Tags         empregos
// @Accept       json
// @Produce      json
// @Param        request  body      models.Emprego  true  "Dados do emprego"
// @Success      201      {object}  models.Emprego
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/empregos [post]
func (h *EmpregoHandler) Create(c *gin.Context) {
	var emprego models.Emprego
	if err := c.ShouldBindJSON(&emprego); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &emprego)
	if err != nil {
		// Check if it's a validation error (simple check for common validation terms)
		errMsg := err.Error()
		if strings.Contains(errMsg, "obrigatório") || strings.Contains(errMsg, "inválid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		// Otherwise, treat as database error
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{
			"error": dbErr.GetUserFriendlyMessage(),
			"field": dbErr.Field,
		})
		return
	}

	emprego.ID = id
	c.JSON(http.StatusCreated, emprego)
}

// @Summary      Obter emprego por ID
// @Description  Retorna um emprego pelo seu ID
// @Tags         empregos
// @Produce      json
// @Param        id   path      int  true  "ID do emprego"
// @Success      200  {object}  models.Emprego
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/empregos/{id} [get]
func (h *EmpregoHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	emprego, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar emprego: " + err.Error()})
		return
	}

	if emprego == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Emprego não encontrado"})
		return
	}

	c.JSON(http.StatusOK, emprego)
}

// @Summary      Atualizar emprego
// @Description  Atualiza os dados de um emprego existente
// @Tags         empregos
// @Accept       json
// @Produce      json
// @Param        id       path      int         true  "ID do emprego"
// @Param        request  body      models.Emprego  true  "Dados do emprego"
// @Success      200      {object}  models.Emprego
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/empregos/{id} [put]
func (h *EmpregoHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var emprego models.Emprego
	if err := c.ShouldBindJSON(&emprego); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	emprego.ID = id
	if err := h.service.Update(c.Request.Context(), &emprego); err != nil {
		// Check if it's a validation error
		errMsg := err.Error()
		if strings.Contains(errMsg, "obrigatório") || strings.Contains(errMsg, "inválid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		// Otherwise, treat as database error
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{
			"error": dbErr.GetUserFriendlyMessage(),
			"field": dbErr.Field,
		})
		return
	}

	c.JSON(http.StatusOK, emprego)
}

// @Summary      Excluir emprego
// @Description  Remove um emprego pelo ID
// @Tags         empregos
// @Produce      json
// @Param        id   path      int  true  "ID do emprego"
// @Success      200  {object}  object
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/empregos/{id} [delete]
func (h *EmpregoHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Check if emprego exists first
	emprego, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar emprego: " + err.Error()})
		return
	}
	if emprego == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Emprego não encontrado"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir emprego: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Emprego excluído com sucesso"})
}

// @Summary      Listar empregos
// @Description  Retorna uma lista paginada de empregos com filtros opcionais
// @Tags         empregos
// @Produce      json
// @Param        page               query     int     false  "Número da página (default: 1)"
// @Param        pageSize           query     int     false  "Tamanho da página (default: 10)"
// @Param        orgao_id           query     string  false  "Filtrar por ID do órgão (external reference)"
// @Param        empresa_id         query     int     false  "Filtrar por ID da empresa"
// @Param        escolaridade_id    query     int     false  "Filtrar por ID da escolaridade"
// @Param        status             query     string  false  "Filtrar por status"
// @Param        tipo_contratacao   query     string  false  "Filtrar por tipo de contratação"
// @Param        jornada_trabalho   query     string  false  "Filtrar por jornada de trabalho"
// @Success      200                {object}  object
// @Failure      500                {object}  models.ErrorResponse
// @Router       /api/v1/empregos [get]
func (h *EmpregoHandler) List(c *gin.Context) {
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
	
	// Adicionar filtros conforme query parameters
	if orgaoID := c.Query("orgao_id"); orgaoID != "" {
		filter["orgao_id"] = orgaoID
	}
	
	if empresaID := c.Query("empresa_id"); empresaID != "" {
		if id, err := strconv.Atoi(empresaID); err == nil {
			filter["empresa_id"] = id
		}
	}
	
	if escolaridadeID := c.Query("escolaridade_id"); escolaridadeID != "" {
		if id, err := strconv.Atoi(escolaridadeID); err == nil {
			filter["escolaridade_id"] = id
		}
	}
	
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	
	if tipoContratacao := c.Query("tipo_contratacao"); tipoContratacao != "" {
		filter["tipo_contratacao"] = tipoContratacao
	}
	
	if jornadaTrabalho := c.Query("jornada_trabalho"); jornadaTrabalho != "" {
		filter["jornada_trabalho"] = jornadaTrabalho
	}
	
	empregos, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar empregos: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": empregos,
		"meta": gin.H{
			"page": page,
			"page_size": pageSize,
			"total": total,
		},
	})
} 