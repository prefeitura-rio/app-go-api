package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/handlers/common"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type CategoriaHandler struct {
	*common.CRUDHandler[models.Categoria]
	service *services.CategoriaService
}

func NewCategoriaHandler(service *services.CategoriaService) *CategoriaHandler {
	return &CategoriaHandler{
		CRUDHandler: common.NewCRUDHandler[models.Categoria](service, "Categoria"),
		service:     service,
	}
}

// WithCache adds cache support to the handler
func (h *CategoriaHandler) WithCache(cache *cache.ReferenceDataCache) *CategoriaHandler {
	h.CRUDHandler.WithCache(cache)
	return h
}

// @Summary      Criar categoria
// @Description  Cria uma nova categoria de curso
// @Tags         categorias
// @Accept       json
// @Produce      json
// @Param        request  body      models.Categoria  true  "Dados da categoria"
// @Success      201      {object}  models.Categoria
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/categorias [post]
func (h *CategoriaHandler) Create(c *gin.Context) {
	h.CRUDHandler.Create(c)
}

// @Summary      Listar categorias
// @Description  Retorna lista paginada de categorias. Use onlyWithCourses=true para filtrar apenas categorias com cursos publicados e com inscrições abertas ou vencidas há no máximo N dias.
// @Tags         categorias
// @Produce      json
// @Param        page            query     int   false  "Número da página (default: 1)"
// @Param        pageSize        query     int   false  "Tamanho da página (default: 10)"
// @Param        onlyWithCourses query     bool  false  "Filtrar apenas categorias com cursos publicados"
// @Param        daysTolerance   query     int   false  "Dias de tolerância após vencimento da inscrição (default: 30)"
// @Success      200             {object}  object{data=[]models.Categoria,meta=object{page=int,page_size=int,total=int}}
// @Failure      500             {object}  models.ErrorResponse
// @Router       /api/v1/categorias [get]
func (h *CategoriaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	filter := make(map[string]interface{})

	if c.Query("onlyWithCourses") == "true" {
		daysTolerance := 30
		if dt := c.Query("daysTolerance"); dt != "" {
			if parsed, err := strconv.Atoi(dt); err == nil && parsed >= 0 {
				daysTolerance = parsed
			}
		}
		filter["days_tolerance"] = daysTolerance
	}

	categorias, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar categorias: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categorias,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// @Summary      Buscar categoria
// @Description  Retorna uma categoria pelo ID
// @Tags         categorias
// @Produce      json
// @Param        id   path      int  true  "ID da categoria"
// @Success      200  {object}  models.Categoria
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/categorias/{id} [get]
func (h *CategoriaHandler) GetByID(c *gin.Context) {
	h.CRUDHandler.GetByID(c)
}

// @Summary      Atualizar categoria
// @Description  Atualiza uma categoria existente
// @Tags         categorias
// @Accept       json
// @Produce      json
// @Param        id       path      int               true  "ID da categoria"
// @Param        request  body      models.Categoria  true  "Dados da categoria"
// @Success      200      {object}  models.Categoria
// @Failure      400      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/categorias/{id} [put]
func (h *CategoriaHandler) Update(c *gin.Context) {
	h.CRUDHandler.Update(c)
}

// @Summary      Excluir categoria
// @Description  Remove uma categoria pelo ID
// @Tags         categorias
// @Produce      json
// @Param        id   path      int  true  "ID da categoria"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/categorias/{id} [delete]
func (h *CategoriaHandler) Delete(c *gin.Context) {
	h.CRUDHandler.Delete(c)
}
