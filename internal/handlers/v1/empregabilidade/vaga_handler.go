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

func isPublishedStatus(status empregabilidade.StatusVaga) bool {
	switch status {
	case empregabilidade.StatusVagaPublicadoAtivo,
		empregabilidade.StatusVagaPublicadoExpirado,
		empregabilidade.StatusVagaCongelada,
		empregabilidade.StatusVagaDescontinuada:
		return true
	}
	return false
}

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
// @Router       /api/v1/empregabilidade/vagas/draft [post]
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

// @Summary      Listar vagas
// @Description  Retorna lista paginada de vagas com filtros opcionais
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        page              query     int     false  "Número da página (default: 1)"
// @Param        pageSize          query     int     false  "Tamanho da página (default: 10)"
// @Param        status            query     string  false  "Filtrar por status (em_edicao, em_aprovacao, publicado_ativo, publicado_expirado, vaga_congelada, vaga_descontinuada)"
// @Param        contratante       query     string  false  "Filtrar por CNPJ do contratante"
// @Param        orgao_parceiro_id query     string  false  "Filtrar por ID do órgão parceiro"
// @Param        search            query     string  false  "Busca parcial no título da vaga"
// @Success      200               {object}  map[string]interface{}
// @Failure      500               {object}  map[string]string
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

	filter := empregabilidade.VagaFilter{
		Status:          c.Query("status"),
		Contratante:     c.Query("contratante"),
		OrgaoParceiroID: c.Query("orgao_parceiro_id"),
		Search:          c.Query("search"),
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

	existing, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleVagaError(c, err)
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	canEditPublished := middlewares.IsAdmin(c) ||
		middlewares.HasRole(c, "go:empregabilidade:admin") ||
		middlewares.HasRole(c, "go:empregabilidade:editor_sem_curadoria")
	if isPublishedStatus(existing.Status) && !canEditPublished {
		c.JSON(http.StatusForbidden, gin.H{"error": "Apenas administradores, empregabilidade:admin ou editor_sem_curadoria podem editar vagas publicadas"})
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

	updated, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
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

// @Summary      Retornar vaga para rascunho
// @Description  Retorna uma vaga em aprovação para o estado de edição (rascunho)
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/send-to-draft [put]
func (h *VagaHandler) SendToDraft(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.SendToDraft(c.Request.Context(), id); err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga retornada para rascunho com sucesso"})
}

// @Summary      Enviar vaga para aprovação
// @Description  Envia uma vaga em edição para o fluxo de aprovação
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/send-to-approval [put]
func (h *VagaHandler) SendToApproval(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.SendToApproval(c.Request.Context(), id); err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga enviada para aprovação com sucesso"})
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

// @Summary      Congelar vaga
// @Description  Congela uma vaga publicada e atualiza o status de todas as candidaturas vinculadas para vaga_congelada
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/freeze [put]
func (h *VagaHandler) Freeze(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.FreezeVaga(c.Request.Context(), id); err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga congelada com sucesso"})
}

// @Summary      Descongelar vaga
// @Description  Descongela uma vaga congelada e restaura o status anterior das candidaturas
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/unfreeze [put]
func (h *VagaHandler) Unfreeze(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.UnfreezeVaga(c.Request.Context(), id); err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga descongelada com sucesso"})
}

// @Summary      Descontinuar vaga
// @Description  Descontinua uma vaga publicada ou congelada e atualiza o status de todas as candidaturas para vaga_descontinuada
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/discontinue [put]
func (h *VagaHandler) Discontinue(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.DiscontinueVaga(c.Request.Context(), id); err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga descontinuada com sucesso"})
}

// @Summary      Reativar vaga
// @Description  Reativa uma vaga descontinuada e restaura o status anterior das candidaturas
// @Tags         empregabilidade-vagas
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/empregabilidade/vagas/{id}/reactivate [put]
func (h *VagaHandler) Reactivate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.ReactivateVaga(c.Request.Context(), id); err != nil {
		handleVagaError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vaga reativada com sucesso"})
}

// ==================== Public Endpoints ====================

// @Summary      Listar vagas públicas
// @Description  Retorna lista paginada de vagas publicadas (apenas vagas ativas)
// @Tags         empregabilidade-vagas-public
// @Produce      json
// @Param        page       query     int     false  "Número da página (default: 1)"
// @Param        pageSize   query     int     false  "Tamanho da página (default: 10)"
// @Success      200        {object}  map[string]interface{}
// @Failure      500        {object}  map[string]string
// @Router       /api/public/empregabilidade/vagas [get]
func (h *VagaHandler) PublicList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	entities, total, err := h.service.ListPublicActive(c.Request.Context(), page, pageSize)
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

// @Summary      Buscar vaga pública por slug
// @Description  Retorna uma vaga publicada pelo slug (apenas se estiver ativa). Slug no formato "{titulo-slugificado}-{8-chars-do-uuid}".
// @Tags         empregabilidade-vagas-public
// @Produce      json
// @Param        slug  path      string  true  "Slug da vaga (ex: analista-de-ti-junior-f3d23675)"
// @Success      200   {object}  empregabilidade.Vaga
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/public/empregabilidade/vagas/slug/{slug} [get]
func (h *VagaHandler) PublicGetBySlug(c *gin.Context) {
	vagaSlug := c.Param("slug")

	entity, err := h.service.GetBySlug(c.Request.Context(), vagaSlug)
	if err != nil {
		handleVagaError(c, err)
		return
	}

	if entity == nil || entity.Status != empregabilidade.StatusVagaPublicadoAtivo {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// @Summary      Buscar vaga pública
// @Description  Retorna uma vaga publicada pelo ID (apenas se estiver ativa)
// @Tags         empregabilidade-vagas-public
// @Produce      json
// @Param        id   path      string  true  "ID da vaga"
// @Success      200  {object}  empregabilidade.Vaga
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/public/empregabilidade/vagas/{id} [get]
func (h *VagaHandler) PublicGetByID(c *gin.Context) {
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

	// Return 404 if vaga doesn't exist or is not published active
	if entity == nil || entity.Status != empregabilidade.StatusVagaPublicadoAtivo {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vaga não encontrada"})
		return
	}

	c.JSON(http.StatusOK, entity)
}
