package empregabilidade

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

func handleVagaError(c *gin.Context, err error) {
	msg := err.Error()

	// Primeiro verifica erros de domínio conhecidos
	switch {
	case strings.Contains(msg, "não encontrada"):
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
		return
	case strings.Contains(msg, "não está em estado"):
		c.JSON(http.StatusConflict, gin.H{"error": msg})
		return
	case strings.Contains(msg, "contratante não encontrada"):
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Para outros erros, usa ParseDatabaseError para normalizar
	dbErr := utils.ParseDatabaseError(err)
	c.JSON(dbErr.GetHTTPStatusCode(), gin.H{"error": dbErr.GetUserFriendlyMessage()})
}

type VagaHandler struct {
	service *services.VagaService
}

func NewVagaHandler(service *services.VagaService) *VagaHandler {
	return &VagaHandler{service: service}
}

// @Summary      Criar vaga
// @Description  Cria uma nova vaga
// @Tags         empregabilidade-vagas
// @Accept       json
// @Produce      json
// @Param        request  body      empregabilidade.Vaga  true  "Dados da vaga"
// @Success      201      {object}  empregabilidade.Vaga
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas [post]
func (h *VagaHandler) Create(c *gin.Context) {
	var entity empregabilidade.Vaga
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		handleVagaError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Criar rascunho de vaga
// @Description  Cria um rascunho de vaga
// @Tags         empregabilidade-vagas
// @Accept       json
// @Produce      json
// @Param        request  body      empregabilidade.Vaga  true  "Dados da vaga"
// @Success      201      {object}  empregabilidade.Vaga
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/draft [post]
func (h *VagaHandler) CreateDraft(c *gin.Context) {
	var entity empregabilidade.Vaga
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.service.CreateDraft(c.Request.Context(), &entity)
	if err != nil {
		handleVagaError(c, err)
		return
	}

	entity.ID = id
	c.JSON(http.StatusCreated, entity)
}

// @Summary      Listar vagas
// @Description  Retorna lista paginada de vagas
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        page       query     int     false  "Número da página (default: 1)"
// @Param        pageSize   query     int     false  "Tamanho da página (default: 10)"
// @Param        status     query     string  false  "Filtrar por status"
// @Param        contratante query    string  false  "Filtrar por CNPJ do contratante"
// @Success      200        {object}  map[string]interface{}
// @Failure      500        {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas [get]
func (h *VagaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	filter := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	if contratante := c.Query("contratante"); contratante != "" {
		filter["id_contratante"] = contratante
	}

	entities, total, err := h.service.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleVagaError(c, err)
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

// @Summary      Buscar vaga
// @Description  Retorna uma vaga pelo ID
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  empregabilidade.Vaga
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id} [get]
func (h *VagaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleVagaError(c, err)
		return
	}

	if entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Atualizar vaga
// @Description  Atualiza uma vaga existente
// @Tags         empregabilidade-vagas
// @Accept       json
// @Produce      json
// @Param        id       path      string                true  "ID da vaga"
// @Param        request  body      empregabilidade.Vaga  true  "Dados da vaga"
// @Success      200      {object}  empregabilidade.Vaga
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id} [put]
func (h *VagaHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var entity empregabilidade.Vaga
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = id

	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Excluir vaga
// @Description  Remove uma vaga pelo ID
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id} [delete]
func (h *VagaHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga excluída com sucesso"})
}

// @Summary      Publicar vaga
// @Description  Publica uma vaga em edição
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/publish [put]
func (h *VagaHandler) Publish(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.Publish(c.Request.Context(), id); err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga publicada com sucesso"})
}

type UpdateTiposPCDRequest struct {
	TiposPCDIDs []uuid.UUID `json:"tipos_pcd_ids"`
}

// @Summary      Atualizar tipos PCD da vaga
// @Description  Atualiza os tipos de PCD aceitos na vaga
// @Tags         empregabilidade-vagas
// @Accept       json
// @Produce      json
// @Param        id       path      string                 true  "ID da vaga"
// @Param        request  body      UpdateTiposPCDRequest  true  "IDs dos tipos PCD"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/tipos-pcd [put]
func (h *VagaHandler) UpdateTiposPCD(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var request UpdateTiposPCDRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateTiposPCD(c.Request.Context(), id, request.TiposPCDIDs); err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tipos PCD atualizados com sucesso"})
}
