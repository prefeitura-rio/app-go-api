package empregabilidade

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

func handleCandidaturaError(c *gin.Context, err error) {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "não encontrada"):
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
		return
	case strings.Contains(msg, "já existe"), strings.Contains(msg, "já se candidatou"):
		c.JSON(http.StatusConflict, gin.H{"error": msg})
		return
	case strings.Contains(msg, "expirada"),
		strings.Contains(msg, "não está ativa"),
		strings.Contains(msg, "status inválido"):
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	dbErr := utils.ParseDatabaseError(err)
	c.JSON(dbErr.GetHTTPStatusCode(), gin.H{"error": dbErr.GetUserFriendlyMessage()})
}

type CandidaturaHandler struct {
	service *services.CandidaturaService
}

func NewCandidaturaHandler(service *services.CandidaturaService) *CandidaturaHandler {
	return &CandidaturaHandler{service: service}
}

// @Summary      Criar candidatura
// @Description  Cria uma nova candidatura para uma vaga
// @Tags         empregabilidade-candidaturas
// @Accept       json
// @Produce      json
// @Param        request  body      empregabilidade.Candidatura  true  "Dados da candidatura"
// @Success      201      {object}  empregabilidade.Candidatura
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas [post]
func (h *CandidaturaHandler) Create(c *gin.Context) {
	var entity empregabilidade.Candidatura
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cpf := middlewares.GetUserCPF(c)
	if cpf == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "CPF não encontrado no token de autenticação"})
		return
	}
	entity.CPF = cpf

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleCandidaturaError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar candidaturas
// @Description  Retorna lista paginada de candidaturas
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        page      query     int     false  "Número da página (default: 1)"
// @Param        pageSize  query     int     false  "Tamanho da página (default: 10)"
// @Param        cpf       query     string  false  "Filtrar por CPF"
// @Param        vagaId    query     string  false  "Filtrar por ID da vaga"
// @Param        status    query     string  false  "Filtrar por status"
// @Success      200       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas [get]
func (h *CandidaturaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	filter := empregabilidade.CandidaturaFilter{
		CPF:    c.Query("cpf"),
		Status: c.Query("status"),
	}

	if raw := c.Query("vagaId"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID da vaga inválido"})
			return
		}
		filter.VagaID = &id
	}

	entities, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleCandidaturaError(c, err)
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

// @Summary      Buscar candidatura
// @Description  Retorna uma candidatura pelo ID
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        id   path      string  true  "ID da candidatura"
// @Success      200  {object}  empregabilidade.Candidatura
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id} [get]
func (h *CandidaturaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleCandidaturaError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Candidatura não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar candidatura
// @Description  Atualiza uma candidatura existente
// @Tags         empregabilidade-candidaturas
// @Accept       json
// @Produce      json
// @Param        id       path      string                       true  "ID da candidatura"
// @Param        request  body      empregabilidade.Candidatura  true  "Dados da candidatura"
// @Success      200      {object}  empregabilidade.Candidatura
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id} [put]
func (h *CandidaturaHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.Candidatura
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id

	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleCandidaturaError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir candidatura
// @Description  Remove uma candidatura pelo ID
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        id   path      string  true  "ID da candidatura"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id} [delete]
func (h *CandidaturaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleCandidaturaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidatura excluída com sucesso"})
}

type UpdateStatusRequest struct {
	Status empregabilidade.StatusCandidatura `json:"status"`
}

// @Summary      Atualizar status da candidatura
// @Description  Atualiza o status de uma candidatura
// @Tags         empregabilidade-candidaturas
// @Accept       json
// @Produce      json
// @Param        id       path      string               true  "ID da candidatura"
// @Param        request  body      UpdateStatusRequest  true  "Novo status"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id}/status [put]
func (h *CandidaturaHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var request UpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), id, request.Status); err != nil {
		handleCandidaturaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status atualizado com sucesso"})
}

// @Summary      Aprovar candidatura
// @Description  Aprova uma candidatura
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        id   path      string  true  "ID da candidatura"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id}/approve [put]
func (h *CandidaturaHandler) Approve(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Approve(c.Request.Context(), id); err != nil {
		handleCandidaturaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidatura aprovada com sucesso"})
}

// @Summary      Reprovar candidatura
// @Description  Reprova uma candidatura
// @Tags         empregabilidade-candidaturas
// @Produce      json
// @Param        id   path      string  true  "ID da candidatura"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id}/reject [put]
func (h *CandidaturaHandler) Reject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Reject(c.Request.Context(), id); err != nil {
		handleCandidaturaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Candidatura reprovada com sucesso"})
}

type UpdateEtapaRequest struct {
	EtapaID uuid.UUID `json:"id_etapa_atual"`
}

// @Summary      Avançar etapa da candidatura
// @Description  Atualiza a etapa atual de uma candidatura
// @Tags         empregabilidade-candidaturas
// @Accept       json
// @Produce      json
// @Param        id       path      string             true  "ID da candidatura"
// @Param        request  body      UpdateEtapaRequest true  "Nova etapa"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/candidaturas/{id}/etapa [put]
func (h *CandidaturaHandler) UpdateEtapa(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var request UpdateEtapaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateEtapa(c.Request.Context(), id, request.EtapaID); err != nil {
		handleCandidaturaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Etapa atualizada com sucesso"})
}
