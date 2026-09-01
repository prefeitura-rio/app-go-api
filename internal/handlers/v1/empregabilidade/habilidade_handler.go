package empregabilidade

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/handlers/v1/response"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
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
	IDHabilidade int64 `json:"id_habilidade" binding:"required" example:"10"`
}

type ReplaceHabilidadesRequest struct {
	Habilidades []*empregabilidade.CurriculoHabilidade `json:"habilidades" binding:"required"`
}

type CreateAreaAtuacaoRequest struct {
	Nome string `json:"nome" binding:"required" example:"Tecnologia da Informação"`
}

type UpdateAreaAtuacaoRequest struct {
	Nome string `json:"nome" binding:"required" example:"Tecnologia da Informação e Comunicação"`
}

// Estrutura do Handler

type HabilidadeHandler struct {
	habilidadeService *service.HabilidadeService
	curriculoService  *service.CurriculoService
}

func NewHabilidadeHandler(hService *service.HabilidadeService, cService *service.CurriculoService) *HabilidadeHandler {
	return &HabilidadeHandler{
		habilidadeService: hService,
		curriculoService:  cService,
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
		response.Error(c, http.StatusInternalServerError, "Erro ao criar habilidade")
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
// @Success      200  {object}  empregabilidade.Habilidade
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
		response.Error(c, http.StatusInternalServerError, "Erro ao buscar habilidade")
		return
	}
	if item == nil {
		response.Error(c, http.StatusNotFound, "Habilidade não encontrada")
		return
	}

	c.JSON(http.StatusOK, item)
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
		response.Error(c, http.StatusInternalServerError, "Erro ao atualizar habilidade")
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
		response.Error(c, http.StatusInternalServerError, "Erro ao excluir habilidade")
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
	termo := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := empregabilidade.HabilidadeFilter{
		Search: termo,
	}

	habilidades, total, err := h.habilidadeService.ListHabilidades(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao buscar habilidades")
		return
	}

	response.PaginatedJSON(c, http.StatusOK, habilidades, total, page, pageSize)
}

// ==========================================
// CRUD DE ÁREAS DE ATUAÇÃO
// ==========================================

// CreateAreaAtuacao cadastra uma nova área de atuação.
// @Summary      Criar Área de Atuação
// @Description  Cadastra uma nova área de atuação no sistema
// @Tags         empregabilidade-areas-atuacao
// @Accept       json
// @Produce      json
// @Param        request  body      CreateAreaAtuacaoRequest  true  "Dados da área de atuação"
// @Success      201      {object}  map[string]int64
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
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
		response.Error(c, http.StatusInternalServerError, "Erro ao criar área de atuação")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetAreaAtuacaoByID busca uma área de atuação por ID.
// @Summary      Buscar Área de Atuação por ID
// @Description  Retorna os detalhes de uma área de atuação cadastrada pelo ID
// @Tags         empregabilidade-areas-atuacao
// @Produce      json
// @Param        id   path      int  true  "ID da Área de Atuação" example(10)
// @Success      200  {object}  empregabilidade.AreaAtuacao
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
		response.Error(c, http.StatusInternalServerError, "Erro ao buscar área de atuação")
		return
	}
	if item == nil {
		response.Error(c, http.StatusNotFound, "Área de atuação não encontrada")
		return
	}

	c.JSON(http.StatusOK, item)
}

// UpdateAreaAtuacao atualiza os dados de uma área de atuação.
// @Summary      Atualizar Área de Atuação
// @Description  Atualiza os dados de uma área de atuação existente
// @Tags         empregabilidade-areas-atuacao
// @Accept       json
// @Produce      json
// @Param        id       path      int                       true  "ID da Área de Atuação" example(10)
// @Param        request  body      UpdateAreaAtuacaoRequest  true  "Dados para atualização"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
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
		response.Error(c, http.StatusInternalServerError, "Erro ao atualizar área de atuação")
		return
	}

	response.Success(c, http.StatusOK, "Área de atuação atualizada com sucesso")
}

// DeleteAreaAtuacao remove uma área de atuação pelo ID.
// @Summary      Excluir Área de Atuação
// @Description  Remove uma área de atuação cadastrada pelo ID
// @Tags         empregabilidade-areas-atuacao
// @Produce      json
// @Param        id   path      int  true  "ID da Área de Atuação" example(10)
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse "ID inválido"
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
		response.Error(c, http.StatusInternalServerError, "Erro ao excluir área de atuação")
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
	termo := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := empregabilidade.AreaAtuacaoFilter{
		Search: termo,
	}

	areas, total, err := h.habilidadeService.ListAreasAtuacao(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao buscar áreas de atuação")
		return
	}

	response.PaginatedJSON(c, http.StatusOK, areas, total, page, pageSize)
}

// ==========================================
// VÍNCULO COM O CURRÍCULO DO CANDIDATO
// ==========================================

// AddHabilidadeAoCurriculo vincula uma nova habilidade ao currículo do usuário.
// @Summary      Adicionar habilidade ao currículo
// @Description  Vincula uma habilidade específica ao currículo do usuário autenticado via JWT
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        request  body      AddHabilidadeRequest  true  "ID da habilidade a ser vinculada"
// @Success      201      {object}  empregabilidade.CurriculoHabilidade
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
// @Failure      401      {object}  response.ErrorResponse "Usuário não autenticado"
// @Failure      500      {object}  response.ErrorResponse "Erro ao adicionar habilidade ao currículo"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/curriculo/habilidades [post]
func (h *HabilidadeHandler) AddHabilidadeAoCurriculo(c *gin.Context) {
	cpf := middlewares.GetUserCPF(c)
	if cpf == "" {
		response.Error(c, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	var req AddHabilidadeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	vinculo := &empregabilidade.CurriculoHabilidade{
		CPF:          cpf,
		IDHabilidade: req.IDHabilidade,
	}

	if err := h.habilidadeService.AddHabilidadeAoCurriculo(c.Request.Context(), vinculo); err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao adicionar habilidade ao currículo")
		return
	}

	c.JSON(http.StatusCreated, vinculo)
}

// ListHabilidadesDoCurriculo busca as habilidades associadas ao candidato logado.
// @Summary      Listar habilidades do currículo
// @Description  Retorna as habilidades vinculadas ao currículo do usuário autenticado via JWT
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Success      200  {array}   empregabilidade.CurriculoHabilidade
// @Failure      401  {object}  response.ErrorResponse "Usuário não autenticado"
// @Failure      500  {object}  response.ErrorResponse "Erro ao buscar habilidades do currículo"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/curriculo/habilidades [get]
func (h *HabilidadeHandler) ListHabilidadesDoCurriculo(c *gin.Context) {
	cpf := middlewares.GetUserCPF(c)
	if cpf == "" {
		response.Error(c, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	habilidades, err := h.habilidadeService.ListHabilidadesPorCPF(c.Request.Context(), cpf)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao buscar habilidades do currículo")
		return
	}

	c.JSON(http.StatusOK, habilidades)
}

// DeleteHabilidadeDoCurriculo remove uma habilidade do currículo.
// @Summary      Remover habilidade do currículo
// @Description  Remove o vínculo de uma habilidade do currículo garantindo ownership pelo CPF do usuário
// @Tags         empregabilidade-curriculo
// @Produce      json
// @Param        id   path      int  true  "ID do vínculo da habilidade" example(10)
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse "ID inválido"
// @Failure      401  {object}  response.ErrorResponse "Usuário não autenticado"
// @Failure      403  {object}  response.ErrorResponse "Acesso negado: o recurso não pertence ao usuário"
// @Failure      500  {object}  response.ErrorResponse "Erro ao remover habilidade do currículo"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/curriculo/habilidades/{id} [delete]
func (h *HabilidadeHandler) DeleteHabilidadeDoCurriculo(c *gin.Context) {
	cpf := middlewares.GetUserCPF(c)
	if cpf == "" {
		response.Error(c, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	idParam := c.Param("id")
	vinculoID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}

	habilidadesDoUsuario, err := h.habilidadeService.ListHabilidadesPorCPF(c.Request.Context(), cpf)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao validar permissões")
		return
	}

	var pertenceAoUsuario bool
	for _, item := range habilidadesDoUsuario {
		if item.ID == vinculoID {
			pertenceAoUsuario = true
			break
		}
	}

	if !pertenceAoUsuario {
		response.Error(c, http.StatusForbidden, "Acesso negado: o recurso não pertence ao usuário")
		return
	}

	if err := h.habilidadeService.DeleteHabilidade(c.Request.Context(), vinculoID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao remover habilidade do currículo")
		return
	}

	response.Success(c, http.StatusOK, "Habilidade removida com sucesso")
}

// ReplaceAllHabilidades substitui a lista inteira de habilidades no accordion do currículo.
// @Summary      Substituir todas as habilidades (Accordion)
// @Description  Substitui em lote a lista completa de habilidades cadastradas no currículo do usuário
// @Tags         empregabilidade-curriculo
// @Accept       json
// @Produce      json
// @Param        request  body      ReplaceHabilidadesRequest  true  "Lista completa de habilidades"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
// @Failure      401      {object}  response.ErrorResponse "Usuário não autenticado"
// @Failure      500      {object}  response.ErrorResponse "Erro ao atualizar accordion de habilidades"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/curriculo/accordion/habilidades [put]
func (h *HabilidadeHandler) ReplaceAllHabilidades(c *gin.Context) {
	cpf := middlewares.GetUserCPF(c)
	if cpf == "" {
		response.Error(c, http.StatusUnauthorized, "Usuário não autenticado")
		return
	}

	var req ReplaceHabilidadesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	for _, item := range req.Habilidades {
		item.CPF = cpf
	}

	if err := h.curriculoService.ReplaceAllHabilidadesByCPF(c.Request.Context(), cpf, req.Habilidades); err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao atualizar accordion de habilidades")
		return
	}

	response.Success(c, http.StatusOK, "Habilidades atualizadas com sucesso")
}
