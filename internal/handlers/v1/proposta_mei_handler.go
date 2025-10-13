package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type PropostaMEIHandler struct {
	service *services.PropostaMEIService
}

func NewPropostaMEIHandler(service *services.PropostaMEIService) *PropostaMEIHandler {
	return &PropostaMEIHandler{
		service: service,
	}
}

// @Summary      Criar proposta MEI
// @Description  Cria uma nova proposta MEI para uma oportunidade
// @Tags         propostas-mei
// @Accept       json
// @Produce      json
// @Param        id              path      int                 true  "ID da oportunidade"
// @Param        request         body      models.PropostaMEI  true  "Dados da proposta"
// @Success      201             {object}  models.PropostaMEI
// @Failure      400             {object}  models.ErrorResponse
// @Failure      500             {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id}/propostas [post]
func (h *PropostaMEIHandler) Create(c *gin.Context) {
	oportunidadeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da oportunidade inválido"})
		return
	}

	var proposta models.PropostaMEI
	if err := c.ShouldBindJSON(&proposta); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	proposta.OportunidadeMEIID = oportunidadeID

	id, err := h.service.Create(c.Request.Context(), &proposta)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proposta.ID = id
	c.JSON(http.StatusCreated, proposta)
}

// @Summary      Obter proposta MEI por ID
// @Description  Retorna uma proposta MEI pelo seu ID
// @Tags         propostas-mei
// @Produce      json
// @Param        id              path      int     true  "ID da oportunidade"
// @Param        propostaId      path      string  true  "UUID da proposta"
// @Success      200             {object}  models.PropostaMEI
// @Failure      400             {object}  models.ErrorResponse
// @Failure      404             {object}  models.ErrorResponse
// @Failure      500             {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id}/propostas/{propostaId} [get]
func (h *PropostaMEIHandler) GetByID(c *gin.Context) {
	propostaID, err := uuid.Parse(c.Param("propostaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID da proposta inválido"})
		return
	}

	proposta, err := h.service.GetByID(c.Request.Context(), propostaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar proposta: " + err.Error()})
		return
	}

	if proposta == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proposta não encontrada"})
		return
	}

	c.JSON(http.StatusOK, proposta)
}

// @Summary      Atualizar proposta MEI
// @Description  Atualiza os dados de uma proposta MEI existente
// @Tags         propostas-mei
// @Accept       json
// @Produce      json
// @Param        id              path      int     true  "ID da oportunidade"
// @Param        propostaId      path      string  true  "UUID da proposta"
// @Param        request         body      object  true  "Dados para atualização: {\"valor_proposta\": 1500.00}"
// @Success      200             {object}  models.PropostaMEI
// @Failure      400             {object}  models.ErrorResponse
// @Failure      404             {object}  models.ErrorResponse
// @Failure      500             {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id}/propostas/{propostaId} [put]
func (h *PropostaMEIHandler) Update(c *gin.Context) {
	oportunidadeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da oportunidade inválido"})
		return
	}

	propostaID, err := uuid.Parse(c.Param("propostaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID da proposta inválido"})
		return
	}

	var req struct {
		ValorProposta *float64 `json:"valor_proposta"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	if err := h.service.UpdateProposta(c.Request.Context(), propostaID, oportunidadeID, req.ValorProposta); err != nil {
		if err.Error() == "proposta não encontrada" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proposta, _ := h.service.GetByID(c.Request.Context(), propostaID)
	c.JSON(http.StatusOK, proposta)
}

// @Summary      Atualizar status de múltiplas propostas MEI
// @Description  Atualiza o status de várias propostas MEI de uma vez (aprovação em lote)
// @Tags         propostas-mei
// @Accept       json
// @Produce      json
// @Param        id              path      int                                     true  "ID da oportunidade"
// @Param        request         body      models.PropostaStatusUpdateRequest      true  "Dados para atualização"
// @Success      200             {object}  models.PropostaStatusUpdateResponse
// @Failure      400             {object}  models.ErrorResponse
// @Failure      500             {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id}/propostas/status [put]
func (h *PropostaMEIHandler) UpdateStatusBulk(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da oportunidade inválido"})
		return
	}

	var request models.PropostaStatusUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	updatedCount, err := h.service.UpdateMultipleStatus(
		c.Request.Context(),
		request.PropostaIDs,
		request.Status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status: " + err.Error()})
		return
	}

	response := models.PropostaStatusUpdateResponse{
		UpdatedCount: updatedCount,
		Status:       request.Status,
		UpdatedAt:    time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Status de " + strconv.Itoa(updatedCount) + " propostas atualizado para '" + string(request.Status) + "'",
	})
}

// @Summary      Atualizar status da proposta MEI
// @Description  Atualiza o status de uma proposta MEI (aprovar ou rejeitar)
// @Tags         propostas-mei
// @Accept       json
// @Produce      json
// @Param        id              path      int     true  "ID da oportunidade"
// @Param        propostaId      path      string  true  "UUID da proposta"
// @Param        request         body      object  true  "Status: {\"status\": \"approved\" | \"rejected\"}"
// @Success      200             {object}  models.PropostaMEI
// @Failure      400             {object}  models.ErrorResponse
// @Failure      500             {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id}/propostas/{propostaId}/status [put]
func (h *PropostaMEIHandler) UpdateStatus(c *gin.Context) {
	propostaID, err := uuid.Parse(c.Param("propostaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID da proposta inválido"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	var status models.StatusPropostaCidadao
	switch req.Status {
	case "approved":
		status = models.StatusPropostaCidadaoApproved
	case "rejected":
		status = models.StatusPropostaCidadaoRejected
	case "submitted":
		status = models.StatusPropostaCidadaoSubmitted
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status inválido. Use: approved, rejected ou submitted"})
		return
	}

	if err := h.service.UpdateStatusCidadao(c.Request.Context(), propostaID, status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proposta, _ := h.service.GetByID(c.Request.Context(), propostaID)
	c.JSON(http.StatusOK, proposta)
}

// @Summary      Excluir proposta MEI
// @Description  Remove uma proposta MEI pelo ID (soft delete)
// @Tags         propostas-mei
// @Produce      json
// @Param        id              path      int     true  "ID da oportunidade"
// @Param        propostaId      path      string  true  "UUID da proposta"
// @Success      200             {object}  object
// @Failure      400             {object}  models.ErrorResponse
// @Failure      500             {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id}/propostas/{propostaId} [delete]
func (h *PropostaMEIHandler) Delete(c *gin.Context) {
	propostaID, err := uuid.Parse(c.Param("propostaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID da proposta inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), propostaID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir proposta: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proposta excluída com sucesso"})
}

// @Summary      Listar propostas MEI por oportunidade
// @Description  Retorna uma lista paginada de propostas MEI para uma oportunidade
// @Tags         propostas-mei
// @Produce      json
// @Param        id              path      int     true   "ID da oportunidade"
// @Param        page            query     int     false  "Número da página (default: 1)"
// @Param        pageSize        query     int     false  "Tamanho da página (default: 10, max: 1000)"
// @Param        nomeEmpresa     query     string  false  "Buscar por nome da empresa (case-insensitive)"
// @Param        cnpj            query     string  false  "Buscar por CNPJ"
// @Param        status          query     string  false  "Filtrar por status (submitted, approved, rejected)"
// @Success      200             {object}  object
// @Failure      500             {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id}/propostas [get]
func (h *PropostaMEIHandler) List(c *gin.Context) {
	oportunidadeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da oportunidade inválido"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	nomeEmpresa := c.Query("nomeEmpresa")
	cnpj := c.Query("cnpj")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	propostas, total, err := h.service.ListByOportunidade(c.Request.Context(), oportunidadeID, nomeEmpresa, cnpj, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar propostas: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": propostas,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// @Summary      Listar propostas MEI por MEI empresa
// @Description  Retorna uma lista paginada de propostas MEI de uma empresa
// @Tags         propostas-mei
// @Produce      json
// @Param        meiEmpresaId  query     int  true   "ID da MEI empresa"
// @Param        page          query     int  false  "Número da página (default: 1)"
// @Param        pageSize      query     int  false  "Tamanho da página (default: 10, max: 1000)"
// @Success      200           {object}  object
// @Failure      400           {object}  models.ErrorResponse
// @Failure      500           {object}  models.ErrorResponse
// @Router       /api/v1/propostas-mei/por-empresa [get]
func (h *PropostaMEIHandler) ListByMEIEmpresa(c *gin.Context) {
	meiEmpresaID, err := strconv.Atoi(c.Query("meiEmpresaId"))
	if err != nil || meiEmpresaID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da MEI empresa inválido ou não fornecido"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 1000 {
		pageSize = 10
	}

	propostas, total, err := h.service.ListByMEIEmpresa(c.Request.Context(), meiEmpresaID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar propostas: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": propostas,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}
