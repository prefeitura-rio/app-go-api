package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type CursoHandler struct {
	service *services.CursoService
}

func NewCursoHandler(service *services.CursoService) *CursoHandler {
	return &CursoHandler{
		service: service,
	}
}

// @Summary      Criar curso
// @Description  Cria um novo curso
// @Tags         cursos
// @Accept       json
// @Produce      json
// @Param        request  body      models.Curso  true  "Dados do curso"
// @Success      201      {object}  models.Curso
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/cursos [post]
func (h *CursoHandler) Create(c *gin.Context) {
	var curso models.Curso
	if err := c.ShouldBindJSON(&curso); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &curso)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar curso: " + err.Error()})
		return
	}

	curso.ID = id
	c.JSON(http.StatusCreated, curso)
}

// @Summary      Obter curso por ID
// @Description  Retorna um curso pelo seu ID
// @Tags         cursos
// @Produce      json
// @Param        id   path      int  true  "ID do curso"
// @Success      200  {object}  models.Curso
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/cursos/{id} [get]
func (h *CursoHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	curso, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar curso: " + err.Error()})
		return
	}

	if curso == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Curso não encontrado"})
		return
	}

	c.JSON(http.StatusOK, curso)
}

// @Summary      Atualizar curso
// @Description  Atualiza os dados de um curso existente
// @Tags         cursos
// @Accept       json
// @Produce      json
// @Param        id       path      int         true  "ID do curso"
// @Param        request  body      models.Curso  true  "Dados do curso"
// @Success      200      {object}  models.Curso
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/cursos/{id} [put]
func (h *CursoHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var curso models.Curso
	if err := c.ShouldBindJSON(&curso); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	curso.ID = id
	if err := h.service.Update(c.Request.Context(), &curso); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar curso: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, curso)
}

// @Summary      Excluir curso
// @Description  Remove um curso pelo ID
// @Tags         cursos
// @Produce      json
// @Param        id   path      int  true  "ID do curso"
// @Success      200  {object}  object
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /api/v1/cursos/{id} [delete]
func (h *CursoHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir curso: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Curso excluído com sucesso"})
}

// @Summary      Listar cursos
// @Description  Retorna uma lista paginada de cursos com filtros opcionais
// @Tags         cursos
// @Produce      json
// @Param        page             query     int     false  "Número da página (default: 1)"
// @Param        pageSize         query     int     false  "Tamanho da página (default: 10)"
// @Param        orgao_id         query     int     false  "Filtrar por ID do órgão"
// @Param        instituicao_id   query     int     false  "Filtrar por ID da instituição"
// @Param        status           query     string  false  "Filtrar por status"
// @Param        modalidade       query     string  false  "Filtrar por modalidade"
// @Success      200              {object}  object
// @Failure      500              {object}  models.ErrorResponse
// @Router       /api/v1/cursos [get]
func (h *CursoHandler) List(c *gin.Context) {
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
	
	// Adicionar filtros conforme query parameters
	if orgaoID := c.Query("orgao_id"); orgaoID != "" {
		if id, err := strconv.Atoi(orgaoID); err == nil {
			filter["orgao_id"] = id
		}
	}
	
	if instituicaoID := c.Query("instituicao_id"); instituicaoID != "" {
		if id, err := strconv.Atoi(instituicaoID); err == nil {
			filter["instituicao_id"] = id
		}
	}
	
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	
	if modalidade := c.Query("modalidade"); modalidade != "" {
		filter["modalidade"] = modalidade
	}
	
	cursos, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar cursos: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": cursos,
		"meta": gin.H{
			"page": page,
			"page_size": pageSize,
			"total": total,
		},
	})
} 