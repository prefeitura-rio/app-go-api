package v1

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/authorization"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type PropostaMEIHandler struct {
	service           services.PropostaMEIServiceInterface
	cnaeValidationSvc services.CNAEValidationServiceInterface
	authChecker       *authorization.Checker
	config            *config.AppConfig
}

func NewPropostaMEIHandler(
	service services.PropostaMEIServiceInterface,
	cnaeValidationSvc services.CNAEValidationServiceInterface,
	authChecker *authorization.Checker,
	cfg *config.AppConfig,
) *PropostaMEIHandler {
	return &PropostaMEIHandler{
		service:           service,
		cnaeValidationSvc: cnaeValidationSvc,
		authChecker:       authChecker,
		config:            cfg,
	}
}

// @Summary      Criar proposta MEI
// @Description  Cria uma nova proposta MEI para uma oportunidade. Valida que o CNPJ pertence ao usuário e possui CNAE compatível.
// @Tags         propostas-mei
// @Accept       json
// @Produce      json
// @Param        Authorization   header    string              true  "Bearer token"
// @Param        id              path      int                 true  "ID da oportunidade"
// @Param        request         body      models.PropostaMEI  true  "Dados da proposta"
// @Success      201             {object}  models.PropostaMEI
// @Failure      400             {object}  models.ErrorResponse
// @Failure      401             {object}  models.ErrorResponse "Token não fornecido"
// @Failure      403             {object}  models.ErrorResponse "CNPJ não pertence ao usuário ou CNAE incompatível"
// @Failure      404             {object}  models.ErrorResponse "Oportunidade não encontrada"
// @Failure      503             {object}  models.ErrorResponse "Erro ao validar CNPJs"
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

	// Extract Authorization header
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token de autorização não fornecido"})
		return
	}

	// Pass authToken to service
	id, err := h.service.Create(c.Request.Context(), &proposta, authToken)
	if err != nil {
		// UX: Different status codes for different errors
		if strings.Contains(err.Error(), "não pertence ao seu CPF") ||
			strings.Contains(err.Error(), "não possui CNAE compatível") ||
			strings.Contains(err.Error(), "Nenhum CNPJ encontrado") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "não encontrada") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "validar seus CNPJs") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proposta.ID = id
	c.JSON(http.StatusCreated, proposta)
}

// @Summary      Obter proposta MEI por ID
// @Description  Retorna uma proposta MEI pelo seu ID. Requer que o usuário seja dono do CNPJ ou tenha permissões de leitura.
// @Tags         propostas-mei
// @Produce      json
// @Param        Authorization   header    string  true  "Bearer token"
// @Param        id              path      int     true  "ID da oportunidade"
// @Param        propostaId      path      string  true  "UUID da proposta"
// @Success      200             {object}  models.PropostaMEI
// @Failure      400             {object}  models.ErrorResponse
// @Failure      403             {object}  models.ErrorResponse "Acesso negado: não é dono do CNPJ"
// @Failure      404             {object}  models.ErrorResponse
// @Failure      500             {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id}/propostas/{propostaId} [get]
func (h *PropostaMEIHandler) GetByID(c *gin.Context) {
	propostaID, err := uuid.Parse(c.Param("propostaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID da proposta inválido"})
		return
	}

	// Get proposal from database to check ownership
	proposta, err := h.service.GetByID(c.Request.Context(), propostaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar proposta: " + err.Error()})
		return
	}

	if proposta == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proposta não encontrada"})
		return
	}

	// Check authorization: CNPJ owner OR has any of the configured read permissions
	if err := authorization.RequireCNPJOwnershipOrAnyPermission(
		c,
		h.authChecker,
		h.cnaeValidationSvc,
		proposta.MEIEmpresaID,
		"proposta_mei",
		h.config.PropostaMEI.ReadPermissions,
	); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: você não tem permissão para visualizar esta proposta"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, proposta)
}

// @Summary      Atualizar proposta MEI
// @Description  Atualiza os dados de uma proposta MEI existente. Requer que o usuário seja o dono da proposta ou tenha uma das permissões configuradas.
// @Tags         propostas-mei
// @Accept       json
// @Produce      json
// @Param        id              path      int     true  "ID da oportunidade"
// @Param        propostaId      path      string  true  "UUID da proposta"
// @Param        request         body      object  true  "Dados para atualização: {\"valor_proposta\": 1500.00}"
// @Success      200             {object}  models.PropostaMEI
// @Failure      400             {object}  models.ErrorResponse
// @Failure      403             {object}  models.ErrorResponse "Acesso negado"
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

	// Get proposal from database to check ownership
	proposta, err := h.service.GetByID(c.Request.Context(), propostaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar proposta: " + err.Error()})
		return
	}

	if proposta == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proposta não encontrada"})
		return
	}

	// Check authorization: CNPJ owner OR has any of the configured permissions
	if err := authorization.RequireCNPJOwnershipOrAnyPermission(
		c,
		h.authChecker,
		h.cnaeValidationSvc,
		proposta.MEIEmpresaID,
		"proposta_mei",
		h.config.PropostaMEI.UpdatePermissions,
	); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: você não tem permissão para atualizar esta proposta"})
		c.Abort()
		return
	}

	var req struct {
		ValorProposta         *float64 `json:"valor_proposta"`
		PrazoExecucao         *string  `json:"prazo_execucao"`
		AceitaCustosIntegrais *bool    `json:"aceita_custos_integrais"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	if err := h.service.UpdateProposta(c.Request.Context(), propostaID, oportunidadeID, req.ValorProposta, req.PrazoExecucao, req.AceitaCustosIntegrais); err != nil {
		if err.Error() == "proposta não encontrada" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proposta, _ = h.service.GetByID(c.Request.Context(), propostaID)
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
// @Description  Remove uma proposta MEI pelo ID (soft delete). Requer que o usuário seja o dono da proposta ou tenha uma das permissões configuradas.
// @Tags         propostas-mei
// @Produce      json
// @Param        id              path      int     true  "ID da oportunidade"
// @Param        propostaId      path      string  true  "UUID da proposta"
// @Success      200             {object}  object
// @Failure      400             {object}  models.ErrorResponse
// @Failure      403             {object}  models.ErrorResponse "Acesso negado"
// @Failure      404             {object}  models.ErrorResponse
// @Failure      500             {object}  models.ErrorResponse
// @Router       /api/v1/oportunidades-mei/{id}/propostas/{propostaId} [delete]
func (h *PropostaMEIHandler) Delete(c *gin.Context) {
	propostaID, err := uuid.Parse(c.Param("propostaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UUID da proposta inválido"})
		return
	}

	// Get proposal from database to check ownership
	proposta, err := h.service.GetByID(c.Request.Context(), propostaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar proposta: " + err.Error()})
		return
	}

	if proposta == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proposta não encontrada"})
		return
	}

	// Check authorization: CNPJ owner OR has any of the configured permissions
	if err := authorization.RequireCNPJOwnershipOrAnyPermission(
		c,
		h.authChecker,
		h.cnaeValidationSvc,
		proposta.MEIEmpresaID,
		"proposta_mei",
		h.config.PropostaMEI.DeletePermissions,
	); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: você não tem permissão para excluir esta proposta"})
		c.Abort()
		return
	}

	// Proceed with deletion
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
// @Param        meiEmpresaId  query     string  true   "CNPJ da MEI empresa"
// @Param        page          query     int     false  "Número da página (default: 1)"
// @Param        pageSize      query     int     false  "Tamanho da página (default: 10, max: 1000)"
// @Success      200           {object}  object
// @Failure      400           {object}  models.ErrorResponse
// @Failure      500           {object}  models.ErrorResponse
// @Router       /api/v1/propostas-mei/por-empresa [get]
func (h *PropostaMEIHandler) ListByMEIEmpresa(c *gin.Context) {
	meiEmpresaID := c.Query("meiEmpresaId")
	if meiEmpresaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CNPJ da MEI empresa não fornecido"})
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
