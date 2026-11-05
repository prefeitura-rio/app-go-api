package empregabilidade

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	service "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// DTOs / Structs de Requisição e Resposta

type CreateHabilidadeRequest struct {
	Nome string `json:"nome" binding:"required" example:"Desenvolvimento Go"`
}

type UpdateHabilidadeRequest struct {
	Nome string `json:"nome" binding:"required" example:"Desenvolvimento Go Avançado"`
}

type AddHabilidadeRequest struct {
	IDHabilidade int64 `json:"id_habilidade" binding:"required,gt=0" example:"10"`
}

type CreateAreaAtuacaoRequest struct {
	Nome string `json:"nome" binding:"required" example:"Tecnologia da Informação"`
}

type UpdateAreaAtuacaoRequest struct {
	Nome string `json:"nome" binding:"required" example:"Tecnologia da Informação e Comunicação"`
}

func catalogoFromHabilidades(items []*empregabilidade.Habilidade) []empregabilidade.CatalogoItem {
	out := make([]empregabilidade.CatalogoItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, empregabilidade.CatalogoItem{ID: item.ID, Nome: item.Nome})
	}
	return out
}

func catalogoFromAreas(items []*empregabilidade.AreaAtuacao) []empregabilidade.CatalogoItem {
	out := make([]empregabilidade.CatalogoItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, empregabilidade.CatalogoItem{ID: item.ID, Nome: item.Nome})
	}
	return out
}

func catalogoFromComportamentos(items []*empregabilidade.ComportamentoAtitudes) []empregabilidade.CatalogoItem {
	out := make([]empregabilidade.CatalogoItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, empregabilidade.CatalogoItem{ID: item.ID, Nome: item.Nome})
	}
	return out
}

const (
	catalogoDefaultPageSize = 20
	catalogoMaxPageSize     = 100
)

func parsePage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(catalogoDefaultPageSize)))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = catalogoDefaultPageSize
	}
	if pageSize > catalogoMaxPageSize {
		pageSize = catalogoMaxPageSize
	}
	return page, pageSize
}

// Estrutura do Handler

type HabilidadeHandler struct {
	habilidadeService *service.HabilidadeService
}

func NewHabilidadeHandler(hService *service.HabilidadeService, cService *service.CurriculoService) *HabilidadeHandler {
	return &HabilidadeHandler{
		habilidadeService: hService,
	}
}

// ==========================================
// CRUD HABILIDADES
// ==========================================

// CreateHabilidade cadastra uma nova habilidade no sistema global.
// @Summary      Criar Habilidade
// @Description  Cadastra uma nova habilidade global no sistema
// @Tags         empregabilidade-habilidades
// @Accept       json
// @Produce      json
// @Param        request  body      CreateHabilidadeRequest  true  "Dados da habilidade"
// @Success      201      {object}  map[string]int64
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
// @Failure      403      {object}  response.ErrorResponse "Sem permissão"
// @Failure      409      {object}  response.ErrorResponse "Nome já cadastrado"
// @Failure      500      {object}  response.ErrorResponse "Erro ao criar habilidade"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/habilidades [post]
func (h *HabilidadeHandler) CreateHabilidade(c *gin.Context) {
	var req CreateHabilidadeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	entity := &empregabilidade.Habilidade{
		Nome: req.Nome,
	}

	id, err := h.habilidadeService.CreateHabilidade(c.Request.Context(), entity)
	if err != nil {
		respondCatalogoError(c, err, "Erro ao criar habilidade")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetHabilidadeByID busca uma habilidade pelo seu ID (int64).
// @Summary      Buscar Habilidade por ID
// @Description  Retorna os detalhes de uma habilidade global cadastrada pelo ID
// @Tags         empregabilidade-habilidades
// @Produce      json
// @Param        id   path      int  true  "ID da Habilidade" example(10)
// @Success      200  {object}  empregabilidade.CatalogoItem
// @Failure      400  {object}  response.ErrorResponse "ID inválido"
// @Failure      404  {object}  response.ErrorResponse "Habilidade não encontrada"
// @Failure      500  {object}  response.ErrorResponse "Erro ao buscar habilidade"
// @Router       /api/v1/empregabilidade/habilidades/{id} [get]
func (h *HabilidadeHandler) GetHabilidadeByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}

	item, err := h.habilidadeService.GetHabilidadeByID(c.Request.Context(), id)
	if err != nil {
		response.RespondError(c, err, "Erro ao buscar habilidade")
		return
	}
	if item == nil {
		response.Error(c, http.StatusNotFound, "Habilidade não encontrada")
		return
	}

	c.JSON(http.StatusOK, empregabilidade.CatalogoItem{ID: item.ID, Nome: item.Nome})
}

// UpdateHabilidade atualiza os dados de uma habilidade global.
// @Summary      Atualizar Habilidade
// @Description  Atualiza os dados de uma habilidade existente no sistema
// @Tags         empregabilidade-habilidades
// @Accept       json
// @Produce      json
// @Param        id       path      int                      true  "ID da Habilidade" example(10)
// @Param        request  body      UpdateHabilidadeRequest  true  "Dados para atualização"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
// @Failure      404      {object}  response.ErrorResponse "Habilidade não encontrada"
// @Failure      403      {object}  response.ErrorResponse "Sem permissão"
// @Failure      409      {object}  response.ErrorResponse "Nome já cadastrado"
// @Failure      500      {object}  response.ErrorResponse "Erro ao atualizar habilidade"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/habilidades/{id} [put]
func (h *HabilidadeHandler) UpdateHabilidade(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var req UpdateHabilidadeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	entity := &empregabilidade.Habilidade{
		ID:   id,
		Nome: req.Nome,
	}

	if err := h.habilidadeService.UpdateHabilidade(c.Request.Context(), entity); err != nil {
		respondCatalogoMutationError(c, err, "Habilidade não encontrada", "Erro ao atualizar habilidade")
		return
	}

	response.Success(c, http.StatusOK, "Habilidade atualizada com sucesso")
}

// DeleteHabilidade remove uma habilidade do sistema pelo ID.
// @Summary      Excluir Habilidade Global
// @Description  Remove uma habilidade global cadastrada pelo ID
// @Tags         empregabilidade-habilidades
// @Produce      json
// @Param        id   path      int  true  "ID da Habilidade" example(10)
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse "ID inválido"
// @Failure      403  {object}  response.ErrorResponse "Sem permissão"
// @Failure      404  {object}  response.ErrorResponse "Habilidade não encontrada"
// @Failure      500  {object}  response.ErrorResponse "Erro ao excluir habilidade"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/habilidades/{id} [delete]
func (h *HabilidadeHandler) DeleteHabilidade(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.habilidadeService.DeleteHabilidade(c.Request.Context(), id); err != nil {
		respondCatalogoMutationError(c, err, "Habilidade não encontrada", "Erro ao excluir habilidade")
		return
	}

	response.Success(c, http.StatusOK, "Habilidade excluída com sucesso")
}

// ListHabilidades consulta habilidades globais com paginação e filtro.
// @Summary      Listar Habilidades
// @Description  Busca a lista de habilidades cadastradas com suporte a paginação e busca textual
// @Tags         empregabilidade-habilidades
// @Produce      json
// @Param        q         query     string  false  "Termo de busca"
// @Param        page      query     int     false  "Número da página (default: 1)" default(1)
// @Param        pageSize  query     int     false  "Tamanho da página (default: 20, max: 100)" default(20)
// @Success      200       {object}  response.ListHabilidadesPaginatedResponse
// @Failure      500       {object}  response.ErrorResponse "Erro ao buscar habilidades"
// @Router       /api/v1/empregabilidade/habilidades [get]
func (h *HabilidadeHandler) ListHabilidades(c *gin.Context) {
	page, pageSize := parsePage(c)
	filter := empregabilidade.HabilidadeFilter{Search: c.Query("q")}

	habilidades, total, err := h.habilidadeService.ListHabilidades(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		response.RespondError(c, err, "Erro ao buscar habilidades")
		return
	}

	response.PaginatedJSON(c, http.StatusOK, catalogoFromHabilidades(habilidades), total, page, pageSize)
}

// ListHabilidadesDaArea lista os conhecimentos de uma área de atuação.
// @Summary      Listar conhecimentos de uma área
// @Description  Retorna os conhecimentos vinculados à área, com paginação e busca sem acento
// @Tags         empregabilidade-habilidades
// @Produce      json
// @Param        id        path      int     true   "ID da Área de Atuação" example(10)
// @Param        q         query     string  false  "Termo de busca"
// @Param        page      query     int     false  "Número da página (default: 1)" default(1)
// @Param        pageSize  query     int     false  "Tamanho da página (default: 20, max: 100)" default(20)
// @Success      200       {object}  response.ListCatalogoPaginatedResponse
// @Failure      400       {object}  response.ErrorResponse "ID inválido"
// @Failure      404       {object}  response.ErrorResponse "Área de atuação não encontrada"
// @Failure      500       {object}  response.ErrorResponse "Erro ao buscar conhecimentos"
// @Router       /api/v1/empregabilidade/habilidades/areas-atuacao/{id}/habilidades [get]
func (h *HabilidadeHandler) ListHabilidadesDaArea(c *gin.Context) {
	areaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || areaID <= 0 {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}
	page, pageSize := parsePage(c)
	items, total, err := h.habilidadeService.ListHabilidadesDaArea(
		c.Request.Context(),
		areaID,
		empregabilidade.HabilidadeFilter{Search: c.Query("q")},
		page,
		pageSize,
	)
	if err != nil {
		response.RespondError(c, err, "Erro ao buscar conhecimentos da área")
		return
	}
	if items == nil {
		items = []*empregabilidade.CatalogoItem{}
	}
	response.PaginatedJSON(c, http.StatusOK, items, total, page, pageSize)
}

// AttachHabilidadeNaArea vincula um conhecimento existente a uma área.
// @Summary      Vincular conhecimento à área
// @Description  Cria o par área × conhecimento. O texto do conhecimento continua global
// @Tags         empregabilidade-habilidades
// @Accept       json
// @Produce      json
// @Param        id       path      int                   true  "ID da Área de Atuação" example(10)
// @Param        request  body      AddHabilidadeRequest  true  "ID do conhecimento"
// @Success      201      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
// @Failure      403      {object}  response.ErrorResponse "Sem permissão"
// @Failure      404      {object}  response.ErrorResponse "Área ou conhecimento não encontrado"
// @Failure      409      {object}  response.ErrorResponse "Conhecimento já vinculado"
// @Failure      500      {object}  response.ErrorResponse "Erro ao vincular conhecimento à área"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/habilidades/areas-atuacao/{id}/habilidades [post]
func (h *HabilidadeHandler) AttachHabilidadeNaArea(c *gin.Context) {
	areaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || areaID <= 0 {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}
	var req AddHabilidadeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}
	if err := h.habilidadeService.AttachHabilidadeNaArea(c.Request.Context(), areaID, req.IDHabilidade); err != nil {
		respondCatalogoError(c, err, "Erro ao vincular conhecimento à área")
		return
	}
	response.Success(c, http.StatusCreated, "Conhecimento vinculado à área de atuação")
}

// DetachHabilidadeDaArea remove um conhecimento de uma área.
// @Summary      Desvincular conhecimento da área
// @Description  Remove o par. Currículos que marcaram esse par perdem o vínculo em cascata
// @Tags         empregabilidade-habilidades
// @Produce      json
// @Param        id             path      int  true  "ID da Área de Atuação" example(10)
// @Param        habilidadeId   path      int  true  "ID do conhecimento" example(2)
// @Success      200            {object}  response.SuccessResponse
// @Failure      400            {object}  response.ErrorResponse "ID inválido"
// @Failure      403            {object}  response.ErrorResponse "Sem permissão"
// @Failure      404            {object}  response.ErrorResponse "Vínculo não encontrado"
// @Failure      500            {object}  response.ErrorResponse "Erro ao desvincular conhecimento da área"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/habilidades/areas-atuacao/{id}/habilidades/{habilidadeId} [delete]
func (h *HabilidadeHandler) DetachHabilidadeDaArea(c *gin.Context) {
	areaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || areaID <= 0 {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}
	habilidadeID, err := strconv.ParseInt(c.Param("habilidadeId"), 10, 64)
	if err != nil || habilidadeID <= 0 {
		response.Error(c, http.StatusBadRequest, "ID de conhecimento inválido")
		return
	}
	if err := h.habilidadeService.DetachHabilidadeDaArea(c.Request.Context(), areaID, habilidadeID); err != nil {
		respondCatalogoError(c, err, "Erro ao desvincular conhecimento da área")
		return
	}
	response.Success(c, http.StatusOK, "Conhecimento desvinculado da área de atuação")
}

// ==========================================
// CRUD DE ÁREAS DE ATUAÇÃO
// ==========================================

// CreateAreaAtuacao cadastra uma nova área de atuação.
// @Summary      Criar Área de Atuação
// @Description  Cadastra uma nova área de atuação no sistema
// @Tags         empregabilidade-habilidades
// @Accept       json
// @Produce      json
// @Param        request  body      CreateAreaAtuacaoRequest  true  "Dados da área de atuação"
// @Success      201      {object}  map[string]int64
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
// @Failure      403      {object}  response.ErrorResponse "Sem permissão"
// @Failure      409      {object}  response.ErrorResponse "Nome já cadastrado"
// @Failure      500      {object}  response.ErrorResponse "Erro ao criar área de atuação"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/habilidades/areas-atuacao [post]
func (h *HabilidadeHandler) CreateAreaAtuacao(c *gin.Context) {
	var req CreateAreaAtuacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	entity := &empregabilidade.AreaAtuacao{
		Nome: req.Nome,
	}

	id, err := h.habilidadeService.CreateAreaAtuacao(c.Request.Context(), entity)
	if err != nil {
		respondCatalogoError(c, err, "Erro ao criar área de atuação")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetAreaAtuacaoByID busca uma área de atuação por ID.
// @Summary      Buscar Área de Atuação por ID
// @Description  Retorna os detalhes de uma área de atuação cadastrada pelo ID
// @Tags         empregabilidade-habilidades
// @Produce      json
// @Param        id   path      int  true  "ID da Área de Atuação" example(10)
// @Success      200  {object}  empregabilidade.CatalogoItem
// @Failure      400  {object}  response.ErrorResponse "ID inválido"
// @Failure      404  {object}  response.ErrorResponse "Área de atuação não encontrada"
// @Failure      500  {object}  response.ErrorResponse "Erro ao buscar área de atuação"
// @Router       /api/v1/empregabilidade/habilidades/areas-atuacao/{id} [get]
func (h *HabilidadeHandler) GetAreaAtuacaoByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}

	item, err := h.habilidadeService.GetAreaAtuacaoByID(c.Request.Context(), id)
	if err != nil {
		response.RespondError(c, err, "Erro ao buscar área de atuação")
		return
	}
	if item == nil {
		response.Error(c, http.StatusNotFound, "Área de atuação não encontrada")
		return
	}

	c.JSON(http.StatusOK, empregabilidade.CatalogoItem{ID: item.ID, Nome: item.Nome})
}

// UpdateAreaAtuacao atualiza os dados de uma área de atuação.
// @Summary      Atualizar Área de Atuação
// @Description  Atualiza os dados de uma área de atuação existente
// @Tags         empregabilidade-habilidades
// @Accept       json
// @Produce      json
// @Param        id       path      int                       true  "ID da Área de Atuação" example(10)
// @Param        request  body      UpdateAreaAtuacaoRequest  true  "Dados para atualização"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
// @Failure      403      {object}  response.ErrorResponse "Sem permissão"
// @Failure      404      {object}  response.ErrorResponse "Área de atuação não encontrada"
// @Failure      409      {object}  response.ErrorResponse "Nome já cadastrado"
// @Failure      500      {object}  response.ErrorResponse "Erro ao atualizar área de atuação"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/habilidades/areas-atuacao/{id} [put]
func (h *HabilidadeHandler) UpdateAreaAtuacao(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var req UpdateAreaAtuacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	entity := &empregabilidade.AreaAtuacao{
		ID:   id,
		Nome: req.Nome,
	}

	if err := h.habilidadeService.UpdateAreaAtuacao(c.Request.Context(), entity); err != nil {
		respondCatalogoMutationError(c, err, "Área de atuação não encontrada", "Erro ao atualizar área de atuação")
		return
	}

	response.Success(c, http.StatusOK, "Área de atuação atualizada com sucesso")
}

// DeleteAreaAtuacao remove uma área de atuação pelo ID.
// @Summary      Excluir Área de Atuação
// @Description  Remove uma área de atuação cadastrada pelo ID
// @Tags         empregabilidade-habilidades
// @Produce      json
// @Param        id   path      int  true  "ID da Área de Atuação" example(10)
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse "ID inválido"
// @Failure      403  {object}  response.ErrorResponse "Sem permissão"
// @Failure      404  {object}  response.ErrorResponse "Área de atuação não encontrada"
// @Failure      500  {object}  response.ErrorResponse "Erro ao excluir área de atuação"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/habilidades/areas-atuacao/{id} [delete]
func (h *HabilidadeHandler) DeleteAreaAtuacao(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.habilidadeService.DeleteAreaAtuacao(c.Request.Context(), id); err != nil {
		respondCatalogoMutationError(c, err, "Área de atuação não encontrada", "Erro ao excluir área de atuação")
		return
	}

	response.Success(c, http.StatusOK, "Área de atuação excluída com sucesso")
}

// ListAreasAtuacao lista as áreas de atuação associadas às habilidades com paginação.
// @Summary      Listar Áreas de Atuação
// @Description  Retorna uma lista paginada das áreas de atuação vinculadas às habilidades
// @Tags         empregabilidade-habilidades
// @Produce      json
// @Param        q         query     string  false  "Termo de busca"
// @Param        page      query     int     false  "Número da página (default: 1)" default(1)
// @Param        pageSize  query     int     false  "Tamanho da página (default: 20, max: 100)" default(20)
// @Success      200       {object}  response.ListAreasAtuacaoPaginatedResponse
// @Failure      500       {object}  response.ErrorResponse "Erro ao buscar áreas de atuação"
// @Router       /api/v1/empregabilidade/habilidades/areas-atuacao [get]
func (h *HabilidadeHandler) ListAreasAtuacao(c *gin.Context) {
	page, pageSize := parsePage(c)
	filter := empregabilidade.AreaAtuacaoFilter{Search: c.Query("q")}

	areas, total, err := h.habilidadeService.ListAreasAtuacao(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		response.RespondError(c, err, "Erro ao buscar áreas de atuação")
		return
	}

	response.PaginatedJSON(c, http.StatusOK, catalogoFromAreas(areas), total, page, pageSize)
}
