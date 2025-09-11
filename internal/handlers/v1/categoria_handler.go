package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type CategoriaHandler struct {
	service *services.CategoriaService
}

func NewCategoriaHandler(service *services.CategoriaService) *CategoriaHandler {
	return &CategoriaHandler{
		service: service,
	}
}

// @Summary      Criar categoria
// @Description  Cria uma nova categoria
// @Tags         categorias
// @Accept       json
// @Produce      json
// @Param        request  body      models.Categoria  true  "Dados da categoria"
// @Success      201      {object}  models.Categoria
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/categorias [post]
func (h *CategoriaHandler) Create(c *gin.Context) {
	var categoria models.Categoria
	if err := c.ShouldBindJSON(&categoria); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados invu00e1lidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &categoria)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar categoria: " + err.Error()})
		return
	}

	categoria.ID = id
	c.JSON(http.StatusCreated, categoria)
}

// @Summary      Obter categoria por ID
// @Description  Retorna uma categoria pelo seu ID
// @Tags         categorias
// @Produce      json
// @Param        id   path      int  true  "ID da categoria"
// @Success      200  {object}  models.Categoria
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/categorias/{id} [get]
func (h *CategoriaHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invu00e1lido"})
		return
	}

	categoria, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar categoria: " + err.Error()})
		return
	}

	if categoria == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Categoria nu00e3o encontrada"})
		return
	}

	c.JSON(http.StatusOK, categoria)
}

// @Summary      Atualizar categoria
// @Description  Atualiza os dados de uma categoria existente
// @Tags         categorias
// @Accept       json
// @Produce      json
// @Param        id       path      int             true  "ID da categoria"
// @Param        request  body      models.Categoria  true  "Dados da categoria"
// @Success      200      {object}  models.Categoria
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/categorias/{id} [put]
func (h *CategoriaHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invu00e1lido"})
		return
	}

	var categoria models.Categoria
	if err := c.ShouldBindJSON(&categoria); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados invu00e1lidos: " + err.Error()})
		return
	}

	categoria.ID = id
	if err := h.service.Update(c.Request.Context(), &categoria); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar categoria: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, categoria)
}

// @Summary      Excluir categoria
// @Description  Remove uma categoria pelo ID
// @Tags         categorias
// @Produce      json
// @Param        id   path      int  true  "ID da categoria"
// @Success      200  {object}  object
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/categorias/{id} [delete]
func (h *CategoriaHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invu00e1lido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir categoria: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Categoria excluu00edda com sucesso"})
}

// @Summary      Listar categorias
// @Description  Retorna uma lista paginada de categorias
// @Tags         categorias
// @Produce      json
// @Param        page     query     int  false  "Nu00famero da pu00e1gina (default: 1)"
// @Param        pageSize query     int  false  "Tamanho da pu00e1gina (default: 10)"
// @Success      200      {object}  object
// @Failure      500      {object}  models.ErrorResponse
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
	
	// Filtros
	filter := make(map[string]interface{})
	
	categorias, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar categorias: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": categorias,
		"meta": gin.H{
			"page": page,
			"page_size": pageSize,
			"total": total,
		},
	})
}
