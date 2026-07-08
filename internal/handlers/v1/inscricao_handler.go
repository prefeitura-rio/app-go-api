package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/prefeitura-rio/app-go-api/internal/jobs"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type InscricaoHandler struct {
	service    services.InscricaoServiceInterface
	jobService services.JobServiceInterface
	cursoRepo  services.CursoRepositoryInterface
}

func NewInscricaoHandler(service services.InscricaoServiceInterface, jobService services.JobServiceInterface, cursoRepo services.CursoRepositoryInterface) *InscricaoHandler {
	return &InscricaoHandler{
		service:    service,
		jobService: jobService,
		cursoRepo:  cursoRepo,
	}
}

// populateRemainingVacanciesForEnrollment calculates remaining vacancies for a single enrollment's enrolled_unit
func (h *InscricaoHandler) populateRemainingVacanciesForEnrollment(c *gin.Context, inscricao *models.Inscricao) error {
	if inscricao.EnrolledUnit == nil || len(inscricao.EnrolledUnit.Schedules) == 0 {
		return nil
	}

	// Collect schedule IDs from enrolled_unit
	scheduleIDs := make([]uuid.UUID, 0, len(inscricao.EnrolledUnit.Schedules))
	for _, schedule := range inscricao.EnrolledUnit.Schedules {
		scheduleID, err := uuid.Parse(schedule.ID)
		if err != nil {
			continue
		}
		scheduleIDs = append(scheduleIDs, scheduleID)
	}

	if len(scheduleIDs) == 0 {
		return nil
	}

	// Get enrollment counts for all schedules
	enrollmentCounts, err := h.cursoRepo.CountEnrollmentsByScheduleIDs(c.Request.Context(), scheduleIDs)
	if err != nil {
		return err
	}

	// Update enrolled_unit schedules with remaining vacancies
	for i := range inscricao.EnrolledUnit.Schedules {
		schedule := &inscricao.EnrolledUnit.Schedules[i]
		scheduleID, err := uuid.Parse(schedule.ID)
		if err != nil {
			continue
		}

		enrolledCount := enrollmentCounts[scheduleID]
		remaining := schedule.Vacancies - int(enrolledCount)

		if remaining < 0 {
			remaining = 0
		}

		schedule.RemainingVacancies = remaining
	}

	return nil
}

// populateRemainingVacanciesForEnrollments calculates remaining vacancies for multiple enrollments in a single batched query
func (h *InscricaoHandler) populateRemainingVacanciesForEnrollments(c *gin.Context, inscricoes []*models.Inscricao) error {
	if len(inscricoes) == 0 {
		return nil
	}

	// Collect all schedule IDs from all enrolled_units
	var allScheduleIDs []uuid.UUID
	for _, inscricao := range inscricoes {
		if inscricao.EnrolledUnit == nil {
			continue
		}
		for _, schedule := range inscricao.EnrolledUnit.Schedules {
			scheduleID, err := uuid.Parse(schedule.ID)
			if err != nil {
				continue
			}
			allScheduleIDs = append(allScheduleIDs, scheduleID)
		}
	}

	if len(allScheduleIDs) == 0 {
		return nil
	}

	// Single query to count all enrollments
	enrollmentCounts, err := h.cursoRepo.CountEnrollmentsByScheduleIDs(c.Request.Context(), allScheduleIDs)
	if err != nil {
		return err
	}

	// Apply counts to each enrollment
	for _, inscricao := range inscricoes {
		if inscricao.EnrolledUnit == nil {
			continue
		}
		for i := range inscricao.EnrolledUnit.Schedules {
			schedule := &inscricao.EnrolledUnit.Schedules[i]
			scheduleID, err := uuid.Parse(schedule.ID)
			if err != nil {
				continue
			}

			enrolledCount := enrollmentCounts[scheduleID]
			remaining := schedule.Vacancies - int(enrolledCount)

			if remaining < 0 {
				remaining = 0
			}

			schedule.RemainingVacancies = remaining
		}
	}

	return nil
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
			"id":            inscricao.ID,
			"course_id":     inscricao.CursoID,
			"status":        inscricao.Status,
			"enrolled_unit": inscricao.EnrolledUnit,
			"enrolled_at":   inscricao.EnrolledAt,
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

	if limit < 1 || limit > 1000 {
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

	// Calculate remaining vacancies for all enrollments in a single batched query
	if err := h.populateRemainingVacanciesForEnrollments(c, inscricoes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular vagas restantes: " + err.Error()})
		return
	}

	// Enrich enrollments with personal info from citizen snapshots
	h.service.EnrichMultipleWithPersonalInfo(c.Request.Context(), inscricoes)

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

// @Summary      Atualizar inscrição
// @Description  Atualiza os dados de uma inscrição existente (exceto CPF, curso e status)
// @Tags         inscricoes
// @Accept       json
// @Produce      json
// @Param        courseId     path      int                           true  "ID do curso"
// @Param        enrollmentId path      string                        true  "UUID da inscrição"
// @Param        request      body      models.InscricaoUpdateRequest true  "Dados para atualização"
// @Success      200          {object}  models.Inscricao
// @Failure      400          {object}  models.ErrorResponse
// @Failure      404          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments/{enrollmentId} [put]
func (h *InscricaoHandler) Update(c *gin.Context) {
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

	var updateData models.InscricaoUpdateRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	if err := h.service.UpdateInscricao(c.Request.Context(), enrollmentID, cursoID, &updateData); err != nil {
		if err.Error() == "inscrição não encontrada" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inscricao, _ := h.service.GetByID(c.Request.Context(), enrollmentID)
	c.JSON(http.StatusOK, inscricao)
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

	// Calculate remaining vacancies for enrolled_unit schedules
	if err := h.populateRemainingVacanciesForEnrollment(c, inscricao); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular vagas restantes: " + err.Error()})
		return
	}

	// Enrich enrollment with personal info from citizen snapshot
	h.service.EnrichWithPersonalInfo(c.Request.Context(), inscricao)

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
// @Description  Remove permanentemente uma inscrição específica. Apenas administradores podem executar esta ação.
// @Tags         inscricoes
// @Produce      json
// @Param        courseId     path      int     true  "ID do curso"
// @Param        enrollmentId path      string  true  "UUID da inscrição"
// @Success      200          {object}  object
// @Failure      400          {object}  models.ErrorResponse
// @Failure      403          {object}  models.ErrorResponse
// @Failure      404          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments/{enrollmentId} [delete]
func (h *InscricaoHandler) Delete(c *gin.Context) {
	userRole := c.GetString("user_role")
	if userRole != "ADMIN" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado: apenas administradores podem excluir inscrições"})
		return
	}

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

	if err := h.service.Delete(c.Request.Context(), enrollmentID, cursoID); err != nil {
		switch err.Error() {
		case "inscrição não encontrada":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "inscrição não pertence ao curso especificado":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir inscrição: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Inscrição excluída com sucesso",
	})
}

// @Summary      Criar inscrição manual
// @Description  Cria uma inscrição manual no curso (para uso do admin). Aplica mesmas validações do endpoint regular.
// @Tags         inscricoes
// @Accept       json
// @Produce      json
// @Param        courseId path      int               true  "ID do curso"
// @Param        request  body      models.Inscricao  true  "Dados da inscrição (nome, cpf, idade, telefone, email, endereço, bairro)"
// @Success      201      {object}  models.Inscricao
// @Failure      400      {object}  models.ErrorResponse
// @Failure      409      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments/manual [post]
func (h *InscricaoHandler) CreateManual(c *gin.Context) {
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

	if err := h.service.CreateManual(c.Request.Context(), &inscricao); err != nil {
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
			"id":            inscricao.ID,
			"course_id":     inscricao.CursoID,
			"cpf":           inscricao.CPF,
			"name":          inscricao.Name,
			"email":         inscricao.Email,
			"phone":         inscricao.Phone,
			"age":           inscricao.Age,
			"address":       inscricao.Address,
			"neighborhood":  inscricao.Neighborhood,
			"status":        inscricao.Status,
			"enrolled_unit": inscricao.EnrolledUnit,
			"enrolled_at":   inscricao.EnrolledAt,
		},
		"message": "Inscrição manual criada com sucesso",
	})
}

// @Summary      Importar inscrições via CSV/XLSX
// @Description  Faz upload de arquivo CSV ou XLSX para importar inscrições em lote. Processamento assíncrono.
// @Tags         inscricoes
// @Accept       multipart/form-data
// @Produce      json
// @Param        courseId path      int    true  "ID do curso"
// @Param        file     formData  file   true  "Arquivo CSV ou XLSX com as inscrições"
// @Success      202      {object}  object "Job ID para acompanhar o progresso"
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId}/enrollments/import [post]
func (h *InscricaoHandler) Import(c *gin.Context) {
	cursoID, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do curso inválido"})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não fornecido ou inválido: " + err.Error()})
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Error closing uploaded file: %v\n", err)
		}
	}()

	// Validate file size (max 10MB)
	maxSize := int64(10 << 20) // 10MB
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo muito grande. Tamanho máximo: 10MB"})
		return
	}

	// Validate file extension
	fileName := header.Filename
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext != ".csv" && ext != ".xlsx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de arquivo inválido. Aceito: CSV ou XLSX"})
		return
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "enrollment-import-*"+ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar arquivo temporário: " + err.Error()})
		return
	}
	defer func() {
		if err := tempFile.Close(); err != nil {
			fmt.Printf("Error closing temp file: %v\n", err)
		}
	}()

	// Copy uploaded file to temp file
	if _, err := io.Copy(tempFile, file); err != nil {
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			fmt.Printf("Error removing temp file: %v\n", removeErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo: " + err.Error()})
		return
	}

	// Create job metadata
	metadata := models.EnrollmentImportMetadata{
		CursoID:  cursoID,
		FileName: tempFile.Name(),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			fmt.Printf("Error removing temp file: %v\n", removeErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar metadata do job: " + err.Error()})
		return
	}

	// Create job
	job := &models.Job{
		Type:     models.JobTypeEnrollmentImport,
		Metadata: datatypes.JSON(metadataJSON),
	}

	if err := h.jobService.Create(c.Request.Context(), job); err != nil {
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			fmt.Printf("Error removing temp file: %v\n", removeErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar job: " + err.Error()})
		return
	}

	// Start processing in background
	if jobs.GlobalJobProcessor != nil {
		jobs.GlobalJobProcessor.StartJob(job.ID)
	} else {
		if err := os.Remove(tempFile.Name()); err != nil {
			fmt.Printf("Error removing temp file: %v\n", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Processador de jobs não inicializado"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"data": gin.H{
			"job_id":   job.ID,
			"status":   job.Status,
			"type":     job.Type,
			"filename": fileName,
		},
		"message": fmt.Sprintf("Importação iniciada. Use GET /api/v1/jobs/%s/status para acompanhar o progresso", job.ID),
	})
}

// @Summary      Listar inscrições de um usuário específico
// @Description  Retorna lista paginada de inscrições realizadas por um usuário (identificado pelo CPF), incluindo certificate_url
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

	if limit < 1 || limit > 1000 {
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

	// Calculate remaining vacancies for all enrollments in a single batched query
	if err := h.populateRemainingVacanciesForEnrollments(c, inscricoes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular vagas restantes: " + err.Error()})
		return
	}

	// Enrich enrollments with personal info from citizen snapshots
	h.service.EnrichMultipleWithPersonalInfo(c.Request.Context(), inscricoes)

	totalPages := (total + limit - 1) / limit

	// Transform enrollments to include course information and personal_info
	enrollmentsData := make([]gin.H, len(inscricoes))
	for i, inscricao := range inscricoes {
		enrollmentsData[i] = gin.H{
			"id":              inscricao.ID,
			"course_id":       inscricao.CursoID,
			"cpf":             inscricao.CPF,
			"name":            inscricao.Name,
			"email":           inscricao.Email,
			"phone":           inscricao.Phone,
			"status":          inscricao.Status,
			"custom_fields":   inscricao.CustomFieldsData,
			"admin_notes":     inscricao.AdminNotes,
			"reason":          inscricao.Reason,
			"certificate_url": inscricao.CertificateURL,
			"enrolled_unit":   inscricao.EnrolledUnit,
			"enrolled_at":     inscricao.EnrolledAt,
			"updated_at":      inscricao.UpdatedAt,
			"curso":           inscricao.Curso,
			"personal_info":   inscricao.PersonalInfo,
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

// @Summary      Trocar turma/horário de inscrição
// @Description  Permite ao cidadão trocar de turma em uma inscrição existente. Deve ser feito com pelo menos 72h de antecedência do início da aula. Se a inscrição estava aprovada, volta para status pendente.
// @Tags         enrollments
// @Accept       json
// @Produce      json
// @Param        enrollmentId path      string                         true  "UUID da inscrição"
// @Param        request      body      models.ScheduleChangeRequest   true  "Dados da nova turma"
// @Success      200          {object}  models.Inscricao
// @Failure      400          {object}  models.ErrorResponse
// @Failure      403          {object}  models.ErrorResponse
// @Failure      404          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/enrollments/{enrollmentId}/schedule [put]
func (h *InscricaoHandler) ChangeSchedule(c *gin.Context) {
	enrollmentID, err := uuid.Parse(c.Param("enrollmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da inscrição inválido"})
		return
	}

	// Get user CPF from token
	userCPF := c.GetString("user_cpf")
	if userCPF == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "CPF do usuário não encontrado no token"})
		return
	}

	var request models.ScheduleChangeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	inscricao, err := h.service.ChangeSchedule(c.Request.Context(), enrollmentID, userCPF, &request)
	if err != nil {
		errMsg := err.Error()
		switch {
		case errMsg == "inscrição não encontrada":
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		case errMsg == "você só pode alterar suas próprias inscrições":
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		case strings.Contains(errMsg, "não é possível trocar de turma"):
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		case strings.Contains(errMsg, "não há vagas"):
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao trocar de turma: " + errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inscricao,
		"message": "Turma alterada com sucesso",
	})
}
