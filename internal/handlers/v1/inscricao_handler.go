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

type InscricaoHandler struct {
	service *services.InscricaoService
}

func NewInscricaoHandler(service *services.InscricaoService) *InscricaoHandler {
	return &InscricaoHandler{
		service: service,
	}
}

// @Summary      Criar inscrição
// @Description  Cria uma nova inscrição em um curso
// @Tags         inscricoes
// @Accept       json
// @Produce      json
// @Param        courseId path      int               true  "ID do curso"
// @Param        request  body      models.Inscricao  true  "Dados da inscrição"
// @Success      201      {object}  models.Inscricao
// @Failure      400      {object}  models.ErrorResponse
// @Failure      409      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments [post]
func (h *InscricaoHandler) Create(c *gin.Context) {
	cursoID, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do curso inválido"})
		return
	}

	var inscricao models.Inscricao
	if err := c.ShouldBindJSON(&inscricao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	// Set the course ID from URL
	inscricao.CursoID = cursoID

	if err := h.service.Create(c.Request.Context(), &inscricao); err != nil {
		if err.Error() == "CPF já inscrito neste curso" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar inscrição: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":          inscricao.ID,
			"course_id":   inscricao.CursoID,
			"status":      inscricao.Status,
			"enrolled_at": inscricao.EnrolledAt,
		},
		"message": "Inscrição criada com sucesso",
	})
}

// @Summary      Listar inscrições de um curso
// @Description  Retorna lista paginada de inscrições de um curso específico
// @Tags         inscricoes
// @Produce      json
// @Param        courseId path      int     true   "ID do curso"
// @Param        page     query     int     false  "Número da página (default: 1)"
// @Param        limit    query     int     false  "Tamanho da página (default: 20)"
// @Param        status   query     string  false  "Filtrar por status"
// @Param        search   query     string  false  "Buscar por CPF, nome ou email"
// @Success      200      {object}  object
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments [get]
func (h *InscricaoHandler) List(c *gin.Context) {
	cursoID, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do curso inválido"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Build filters
	filter := make(map[string]interface{})

	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	if search := c.Query("search"); search != "" {
		filter["search"] = search
	}

	inscricoes, total, err := h.service.GetByCursoID(c.Request.Context(), cursoID, filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar inscrições: " + err.Error()})
		return
	}

	// Get summary
	summary, err := h.service.GetSummaryByCursoID(c.Request.Context(), cursoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter resumo: " + err.Error()})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enrollments": inscricoes,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
			"summary": summary,
		},
	})
}

// @Summary      Atualizar status de múltiplas inscrições
// @Description  Atualiza o status de várias inscrições de uma vez (aprovação em lote)
// @Tags         inscricoes
// @Accept       json
// @Produce      json
// @Param        courseId path      int                                   true  "ID do curso"
// @Param        request  body      models.EnrollmentStatusUpdateRequest true  "Dados para atualização"
// @Success      200      {object}  models.EnrollmentStatusUpdateResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments/status [put]
func (h *InscricaoHandler) UpdateStatus(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do curso inválido"})
		return
	}

	var request models.EnrollmentStatusUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	updatedCount, err := h.service.UpdateMultipleStatus(
		c.Request.Context(),
		request.EnrollmentIDs,
		request.Status,
		request.Reason,
		request.AdminNotes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status: " + err.Error()})
		return
	}

	response := models.EnrollmentStatusUpdateResponse{
		UpdatedCount: updatedCount,
		Status:       request.Status,
		UpdatedAt:    time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "Status de " + strconv.Itoa(updatedCount) + " inscrições atualizado para '" + string(request.Status) + "'",
	})
}

// @Summary      Atualizar status de inscrição individual
// @Description  Atualiza o status de uma inscrição específica
// @Tags         inscricoes
// @Accept       json
// @Produce      json
// @Param        courseId     path      int     true  "ID do curso"
// @Param        enrollmentId path      string  true  "UUID da inscrição"
// @Param        request      body      object  true  "Dados para atualização: {\"status\": \"approved\", \"reason\": \"...\", \"admin_notes\": \"...\"}"
// @Success      200          {object}  object
// @Failure      400          {object}  models.ErrorResponse
// @Failure      404          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments/{enrollmentId}/status [put]
func (h *InscricaoHandler) UpdateIndividualStatus(c *gin.Context) {
	_, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do curso inválido"})
		return
	}

	enrollmentID, err := uuid.Parse(c.Param("enrollmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da inscrição inválido"})
		return
	}

	var request struct {
		Status     models.StatusInscricao `json:"status" binding:"required"`
		Reason     string                 `json:"reason"`
		AdminNotes string                 `json:"admin_notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), enrollmentID, request.Status, request.Reason, request.AdminNotes); err != nil {
		if err.Error() == "inscrição não encontrada" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enrollment_id": enrollmentID,
			"status":        request.Status,
			"reason":        request.Reason,
			"updated_at":    time.Now(),
		},
		"message": "Status da inscrição atualizado para '" + string(request.Status) + "'",
	})
}

// @Summary      Obter inscrição por ID
// @Description  Retorna dados completos de uma inscrição específica
// @Tags         inscricoes
// @Produce      json
// @Param        courseId     path      int     true  "ID do curso"
// @Param        enrollmentId path      string  true  "UUID da inscrição"
// @Success      200          {object}  models.Inscricao
// @Failure      400          {object}  models.ErrorResponse
// @Failure      403          {object}  models.ErrorResponse
// @Failure      404          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments/{enrollmentId} [get]
func (h *InscricaoHandler) GetByID(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("enrollmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da inscrição inválido"})
		return
	}

	inscricao, err := h.service.GetByID(c.Request.Context(), enrollmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar inscrição: " + err.Error()})
		return
	}

	if inscricao == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inscrição não encontrada"})
		return
	}

	// Verificar se o usuário tem permissão para ver esta inscrição
	userCPF := c.GetString("user_cpf")
	userRole := c.GetString("user_role")
	
	// Admin pode ver todas as inscrições
	// Usuário comum só pode ver suas próprias inscrições
	if userRole != "ADMIN" && userCPF != "" && inscricao.CPF != userCPF {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: você só pode visualizar suas próprias inscrições"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inscricao,
	})
}

// @Summary      Atualizar certificado de inscrição
// @Description  Adiciona ou atualiza a URL do certificado de uma inscrição
// @Tags         inscricoes
// @Accept       json
// @Produce      json
// @Param        courseId     path      int                            true  "ID do curso"
// @Param        enrollmentId path      string                         true  "UUID da inscrição"
// @Param        request      body      models.CertificateUpdateRequest true  "URL do certificado"
// @Success      200          {object}  models.CertificateUpdateResponse
// @Failure      400          {object}  models.ErrorResponse
// @Failure      403          {object}  models.ErrorResponse
// @Failure      404          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments/{enrollmentId}/certificate [put]
func (h *InscricaoHandler) UpdateCertificate(c *gin.Context) {
	cursoID, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do curso inválido"})
		return
	}

	enrollmentID, err := uuid.Parse(c.Param("enrollmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da inscrição inválido"})
		return
	}

	var request models.CertificateUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	if err := h.service.UpdateCertificate(c.Request.Context(), cursoID, enrollmentID, request.CertificateURL); err != nil {
		if err.Error() == "inscrição não encontrada" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "inscrição não pertence ao curso especificado" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "certificado só pode ser atribuído a inscrições aprovadas ou concluídas" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar certificado: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.CertificateUpdateResponse{
		Message:        "Certificado atualizado com sucesso",
		CertificateURL: request.CertificateURL,
	})
}

// @Summary      Excluir inscrição
// @Description  Remove uma inscrição específica
// @Tags         inscricoes
// @Produce      json
// @Param        courseId     path      int     true  "ID do curso"
// @Param        enrollmentId path      string  true  "UUID da inscrição"
// @Success      200          {object}  object
// @Failure      400          {object}  models.ErrorResponse
// @Failure      404          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments/{enrollmentId} [delete]
func (h *InscricaoHandler) Delete(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("enrollmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da inscrição inválido"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), enrollmentID); err != nil {
		if err.Error() == "inscrição não encontrada" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir inscrição: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Inscrição excluída com sucesso",
	})
}

// @Summary      Listar inscrições de um usuário específico
// @Description  Retorna lista paginada de inscrições realizadas por um usuário (identificado pelo CPF)
// @Tags         enrollments
// @Produce      json
// @Param        cpf          path      string  true   "CPF do usuário"
// @Param        page         query     int     false  "Número da página (default: 1)"
// @Param        limit        query     int     false  "Tamanho da página (default: 10)"
// @Param        status       query     string  false  "Filtrar por status"
// @Success      200          {object}  object
// @Failure      400          {object}  models.ErrorResponse
// @Failure      403          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/enrollments/user/{cpf} [get]
func (h *InscricaoHandler) ListByUser(c *gin.Context) {
	cpf := c.Param("cpf")
	if cpf == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CPF é obrigatório"})
		return
	}

	// Verificar se o usuário tem permissão para ver estas inscrições
	userCPF := c.GetString("user_cpf")
	userRole := c.GetString("user_role")
	
	// Admin pode ver inscrições de qualquer pessoa
	// Usuário comum só pode ver suas próprias inscrições
	if userRole != "ADMIN" && userCPF != "" && cpf != userCPF {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: você só pode visualizar suas próprias inscrições"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build filters for user's enrollments
	filter := map[string]interface{}{
		"cpf": cpf,
	}

	// Add status filter if provided
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	inscricoes, total, err := h.service.ListByCPF(c.Request.Context(), cpf, filter, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar inscrições do usuário: " + err.Error()})
		return
	}

	totalPages := (total + limit - 1) / limit

	// Transform enrollments to include course information
	enrollmentsData := make([]gin.H, len(inscricoes))
	for i, inscricao := range inscricoes {
		enrollmentsData[i] = gin.H{
			"id":                inscricao.ID,
			"course_id":         inscricao.CursoID,
			"cpf":               inscricao.CPF,
			"name":              inscricao.Name,
			"email":             inscricao.Email,
			"phone":             inscricao.Phone,
			"status":            inscricao.Status,
			"custom_fields":     inscricao.CustomFieldsData,
			"admin_notes":       inscricao.AdminNotes,
			"reason":            inscricao.Reason,
			"enrolled_at":       inscricao.EnrolledAt,
			"updated_at":        inscricao.UpdatedAt,
			"curso":             inscricao.Curso,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enrollments": enrollmentsData,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		},
	})
}