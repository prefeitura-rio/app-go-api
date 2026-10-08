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

type CreateComportamentoAtitudesRequest struct {
	Nome string `json:"nome" binding:"required" example:"Trabalho em Equipe"`
}

type UpdateComportamentoAtitudesRequest struct {
	Nome string `json:"nome" binding:"required" example:"Trabalho em Equipe e Colaboração"`
}

type AddComportamentoAtitudesRequest struct {
	IDComportamentoAtitudes int64 `json:"id_comportamento_atitudes" binding:"required" example:"10"`
}

type ReplaceComportamentoAtitudesRequest struct {
	Comportamentos []*empregabilidade.CurriculoComportamentoAtitudes `json:"comportamentos" binding:"required"`
}

// Estrutura do Handler

type ComportamentoAtitudesHandler struct {
	comportamentoAtitudesService *service.ComportamentoAtitudesService
}

func NewComportamentoAtitudesHandler(caService *service.ComportamentoAtitudesService) *ComportamentoAtitudesHandler {
	return &ComportamentoAtitudesHandler{
		comportamentoAtitudesService: caService,
	}
}

// ==========================================
// CRUD COMPORTAMENTOS E ATITUDES
// ==========================================

// CreateComportamentoAtitudes cadastra um novo comportamento/atitude no sistema global.
// @Summary      Criar Comportamento e Atitude
// @Description  Cadastra um novo comportamento e atitude global no sistema
// @Tags         empregabilidade-comportamentos-atitudes
// @Accept       json
// @Produce      json
// @Param        request  body      CreateComportamentoAtitudesRequest  true  "Dados do comportamento/atitude"
// @Success      201      {object}  map[string]int64
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
// @Failure      500      {object}  response.ErrorResponse "Erro ao criar comportamento/atitude"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/comportamentos-atitudes [post]
func (h *ComportamentoAtitudesHandler) CreateComportamentoAtitudes(c *gin.Context) {
	var req CreateComportamentoAtitudesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	entity := &empregabilidade.ComportamentoAtitudes{
		Nome: req.Nome,
	}

	id, err := h.comportamentoAtitudesService.CreateComportamentoAtitudes(c.Request.Context(), entity)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao criar comportamento/atitude")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetComportamentoAtitudesByID busca um comportamento/atitude pelo seu ID (int64).
// @Summary      Buscar Comportamento e Atitude por ID
// @Description  Retorna os detalhes de um comportamento/atitude global cadastrado pelo ID
// @Tags         empregabilidade-comportamentos-atitudes
// @Produce      json
// @Param        id   path      int  true  "ID do Comportamento/Atitude" example(10)
// @Success      200  {object}  empregabilidade.ComportamentoAtitudes
// @Failure      400  {object}  response.ErrorResponse "ID inválido"
// @Failure      404  {object}  response.ErrorResponse "Comportamento/atitude não encontrado"
// @Failure      500  {object}  response.ErrorResponse "Erro ao buscar comportamento/atitude"
// @Router       /api/v1/empregabilidade/comportamentos-atitudes/{id} [get]
func (h *ComportamentoAtitudesHandler) GetComportamentoAtitudesByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}

	item, err := h.comportamentoAtitudesService.GetComportamentoAtitudesByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao buscar comportamento/atitude")
		return
	}
	if item == nil {
		response.Error(c, http.StatusNotFound, "Comportamento/atitude não encontrado")
		return
	}

	c.JSON(http.StatusOK, item)
}

// UpdateComportamentoAtitudes atualiza os dados de um comportamento/atitude global.
// @Summary      Atualizar Comportamento e Atitude
// @Description  Atualiza os dados de um comportamento/atitude existente no sistema
// @Tags         empregabilidade-comportamentos-atitudes
// @Accept       json
// @Produce      json
// @Param        id       path      int                                 true  "ID do Comportamento/Atitude" example(10)
// @Param        request  body      UpdateComportamentoAtitudesRequest  true  "Dados para atualização"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.ErrorResponse "Dados inválidos"
// @Failure      500      {object}  response.ErrorResponse "Erro ao atualizar comportamento/atitude"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/comportamentos-atitudes/{id} [put]
func (h *ComportamentoAtitudesHandler) UpdateComportamentoAtitudes(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}

	var req UpdateComportamentoAtitudesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Dados inválidos: "+err.Error())
		return
	}

	entity := &empregabilidade.ComportamentoAtitudes{
		ID:   id,
		Nome: req.Nome,
	}

	if err := h.comportamentoAtitudesService.UpdateComportamentoAtitudes(c.Request.Context(), entity); err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao atualizar comportamento/atitude")
		return
	}

	response.Success(c, http.StatusOK, "Comportamento/atitude atualizado com sucesso")
}

// DeleteComportamentoAtitudes remove um comportamento/atitude do sistema pelo ID.
// @Summary      Excluir Comportamento e Atitude Global
// @Description  Remove um comportamento/atitude global cadastrado pelo ID
// @Tags         empregabilidade-comportamentos-atitudes
// @Produce      json
// @Param        id   path      int  true  "ID do Comportamento/Atitude" example(10)
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.ErrorResponse "ID inválido"
// @Failure      500  {object}  response.ErrorResponse "Erro ao excluir comportamento/atitude"
// @Security     BearerAuth
// @Router       /api/v1/empregabilidade/comportamentos-atitudes/{id} [delete]
func (h *ComportamentoAtitudesHandler) DeleteComportamentoAtitudes(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID inválido")
		return
	}

	if err := h.comportamentoAtitudesService.DeleteComportamentoAtitudes(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao excluir comportamento/atitude")
		return
	}

	response.Success(c, http.StatusOK, "Comportamento/atitude excluído com sucesso")
}

// ListComportamentoAtitudes consulta comportamentos/atitudes globais com paginação e filtro.
// @Summary      Listar Comportamentos e Atitudes
// @Description  Busca a lista de comportamentos e atitudes cadastrados com suporte a paginação e busca textual
// @Tags         empregabilidade-comportamentos-atitudes
// @Produce      json
// @Param        q         query     string  false  "Termo de busca"
// @Param        page      query     int     false  "Número da página (default: 1)" default(1)
// @Param        pageSize  query     int     false  "Tamanho da página (default: 20, max: 100)" default(20)
// @Success      200       {object}  response.ListComportamentoAtitudesPaginatedResponse
// @Failure      500       {object}  response.ErrorResponse "Erro ao buscar comportamentos/atitudes"
// @Router       /api/v1/empregabilidade/comportamentos-atitudes [get]
func (h *ComportamentoAtitudesHandler) ListComportamentoAtitudes(c *gin.Context) {
	termo := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := empregabilidade.ComportamentoAtitudesFilter{
		Search: termo,
	}

	comportamentos, total, err := h.comportamentoAtitudesService.ListComportamentoAtitudes(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Erro ao buscar comportamentos/atitudes")
		return
	}

	response.PaginatedJSON(c, http.StatusOK, comportamentos, total, page, pageSize)
}
