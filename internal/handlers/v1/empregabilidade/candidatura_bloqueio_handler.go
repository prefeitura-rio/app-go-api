package empregabilidade

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

func handleCandidaturaBloqueioError(c *gin.Context, err error) {
	dbErr := utils.ParseDatabaseError(err)
	c.JSON(dbErr.GetHTTPStatusCode(), gin.H{"error": dbErr.GetUserFriendlyMessage()})
}

type CandidaturaBloqueioHandler struct {
	service *services.CandidaturaBloqueioService
}

func NewCandidaturaBloqueioHandler(service *services.CandidaturaBloqueioService) *CandidaturaBloqueioHandler {
	return &CandidaturaBloqueioHandler{service: service}
}

// @Summary      Registrar bloqueio de candidatura
// @Description  Registra a tentativa de candidatura de um CPF a uma vaga que foi bloqueada por não atender critérios de elegibilidade
// @Tags         empregabilidade-candidatura-bloqueios
// @Accept       json
// @Produce      json
// @Param        request  body      empregabilidade.CandidaturaBloqueio  true  "Dados do bloqueio"
// @Success      201      {object}  empregabilidade.CandidaturaBloqueio
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidatura-bloqueios [post]
func (h *CandidaturaBloqueioHandler) Create(c *gin.Context) {
	var entity empregabilidade.CandidaturaBloqueio
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleCandidaturaBloqueioError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar bloqueios de candidatura (admin)
// @Description  Retorna lista paginada de bloqueios de candidatura, com filtros opcionais. Restrito a administradores.
// @Tags         empregabilidade-candidatura-bloqueios
// @Produce      json
// @Param        page      query     int     false  "Número da página (default: 1)"
// @Param        pageSize  query     int     false  "Tamanho da página (default: 10)"
// @Param        cpf       query     string  false  "Filtrar por CPF"
// @Param        id_vaga   query     string  false  "Filtrar por ID da vaga"
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidatura-bloqueios [get]
func (h *CandidaturaBloqueioHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	filter := map[string]interface{}{}
	if cpf := c.Query("cpf"); cpf != "" {
		filter["cpf"] = cpf
	}
	if raw := c.Query("id_vaga"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID da vaga inválido"})
			return
		}
		filter["id_vaga"] = id
	}

	entities, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleCandidaturaBloqueioError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": entities,
		"meta": gin.H{"page": page, "page_size": pageSize, "total": total},
	})
}
