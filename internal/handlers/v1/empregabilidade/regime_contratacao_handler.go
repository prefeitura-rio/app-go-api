package empregabilidade

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

type RegimeContratacaoHandler struct {
	service *services.RegimeContratacaoService
}

func NewRegimeContratacaoHandler(service *services.RegimeContratacaoService) *RegimeContratacaoHandler {
	return &RegimeContratacaoHandler{service: service}
}

// @Summary      Criar regime de contratação
// @Description  Cria um novo regime de contratação
// @Tags         empregabilidade-regimes-contratacao
// @Accept       json
// @Produce      json
// @Param        request  body      empregabilidade.RegimeContratacao  true  "Dados do regime"
// @Success      201      {object}  empregabilidade.RegimeContratacao
// @Failure      400      {object}  map[string]string
// @Failure      409      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/regimes-contratacao [post]
func (h *RegimeContratacaoHandler) Create(c *gin.Context) {
	var entity empregabilidade.RegimeContratacao
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar regimes de contratação
// @Description  Retorna lista paginada de regimes de contratação
// @Tags         empregabilidade-regimes-contratacao
// @Produce      json
// @Param        page      query     int  false  "Número da página (default: 1)"
// @Param        pageSize  query     int  false  "Tamanho da página (default: 100)"
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/regimes-contratacao [get]
func (h *RegimeContratacaoHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	entities, total, err := h.service.List(c.Request.Context(), nil, page, pageSize)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": entities,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// @Summary      Buscar regime de contratação
// @Description  Retorna um regime de contratação pelo ID
// @Tags         empregabilidade-regimes-contratacao
// @Produce      json
// @Param        id   path      string  true  "ID do regime"
// @Success      200  {object}  empregabilidade.RegimeContratacao
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/regimes-contratacao/{id} [get]
func (h *RegimeContratacaoHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleLookupError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Regime de contratação não encontrado"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar regime de contratação
// @Description  Atualiza um regime de contratação existente
// @Tags         empregabilidade-regimes-contratacao
// @Accept       json
// @Produce      json
// @Param        id       path      string                              true  "ID do regime"
// @Param        request  body      empregabilidade.RegimeContratacao  true  "Dados do regime"
// @Success      200      {object}  empregabilidade.RegimeContratacao
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/regimes-contratacao/{id} [put]
func (h *RegimeContratacaoHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.RegimeContratacao
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id

	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir regime de contratação
// @Description  Remove um regime de contratação pelo ID
// @Tags         empregabilidade-regimes-contratacao
// @Produce      json
// @Param        id   path      string  true  "ID do regime"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/regimes-contratacao/{id} [delete]
func (h *RegimeContratacaoHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleLookupError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Regime de contratação excluído com sucesso"})
}
