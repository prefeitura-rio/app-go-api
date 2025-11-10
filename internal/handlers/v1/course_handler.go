package v1

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

type CourseHandler struct {
	cursoService     *services.CursoService
	inscricaoService *services.InscricaoService
	cursoRepo        *repository.CursoRepository
}

func NewCourseHandler(cursoService *services.CursoService, inscricaoService *services.InscricaoService, cursoRepo *repository.CursoRepository) *CourseHandler {
	return &CourseHandler{
		cursoService:     cursoService,
		inscricaoService: inscricaoService,
		cursoRepo:        cursoRepo,
	}
}

func transformCursoToResponse(curso *models.Curso) gin.H {
	return gin.H{
		"id":                        curso.ID,
		"title":                     curso.Titulo,
		"description":               curso.Descricao,
		"modalidade":                curso.Modalidade,
		"status":                    curso.Status,
		"organization":              curso.Organization,
		"theme":                     curso.Theme,
		"workload":                  curso.Workload,
		"target_audience":           curso.TargetAudience,
		"institutional_logo":        curso.InstitutionalLogo,
		"cover_image":               curso.CoverImage,
		"enrollment_start_date":     curso.EnrollmentStartDate,
		"enrollment_end_date":       curso.EnrollmentEndDate,
		"orgao_id":                  curso.OrgaoID,
		"instituicao_id":            curso.InstituicaoID,
		"local_realizacao":          curso.LocalRealizacao,
		"data_inicio":               curso.DataInicio,
		"data_termino":              curso.DataTermino,
		"data_limite_inscricoes":    curso.DataLimiteInscricoes,
		"numero_vagas":              curso.NumeroVagas,
		"carga_horaria":             curso.CargaHoraria,
		"turno":                     curso.Turno,
		"formato_aula":              curso.FormatoAula,
		"link_inscricao":            curso.LinkInscricao,
		"contato_duvidas":           curso.ContatoDuvidas,
		"has_certificate":           curso.HasCertificate,
		"facilitator":               curso.Facilitator,
		"objectives":                curso.Objectives,
		"expected_results":          curso.ExpectedResults,
		"program_content":           curso.ProgramContent,
		"methodology":               curso.Methodology,
		"resources_used":            curso.ResourcesUsed,
		"material_used":             curso.MaterialUsed,
		"teaching_material":         curso.TeachingMaterial,
		"pre_requisitos":            curso.PreRequisitos,
		"certificacao_oferecida":    curso.CertificacaoOferecida,
		"is_external_partner":       curso.IsExternalPartner,
		"external_partner_name":     curso.ExternalPartnerName,
		"external_partner_url":      curso.ExternalPartnerURL,
		"external_partner_logo_url": curso.ExternalPartnerLogoURL,
		"external_partner_contact":  curso.ExternalPartnerContact,
		"created_at":                curso.CreatedAt,
		"updated_at":                curso.UpdatedAt,
		"instituicao":               curso.Instituicao,
		"categorias":                curso.Categorias,
		"acessibilidades":           curso.Acessibilidades,
		"accessibility":             curso.Accessibility,
		"is_visible":                curso.IsVisible,
		"custom_fields":             curso.CustomFields,
		"locations":                 curso.LocationClasses,
		"remote_class":              curso.RemoteClass,
	}
}

// calculateRemainingVacancies calculates remaining vacancies for all schedules in a course
func (h *CourseHandler) calculateRemainingVacancies(c *gin.Context, curso *models.Curso) error {
	// Collect all schedule IDs
	var scheduleIDs []uuid.UUID
	for _, location := range curso.LocationClasses {
		for _, schedule := range location.Schedules {
			scheduleIDs = append(scheduleIDs, schedule.ID)
		}
	}

	if len(scheduleIDs) == 0 {
		return nil
	}

	// Get enrollment counts for all schedules
	enrollmentCounts, err := h.cursoRepo.CountEnrollmentsByScheduleIDs(c.Request.Context(), scheduleIDs)
	if err != nil {
		return err
	}

	// Update each schedule with remaining vacancies
	for i := range curso.LocationClasses {
		for j := range curso.LocationClasses[i].Schedules {
			schedule := &curso.LocationClasses[i].Schedules[j]
			enrolledCount := enrollmentCounts[schedule.ID]
			schedule.RemainingVacancies = schedule.Vacancies - int(enrolledCount)

			// Ensure it doesn't go negative
			if schedule.RemainingVacancies < 0 {
				schedule.RemainingVacancies = 0
			}
		}
	}

	return nil
}

// @Summary      Criar curso
// @Description  Cria um novo curso no sistema (status: "opened")
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        request  body      models.Curso  true  "Dados do curso"
// @Success      201      {object}  object
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/courses [post]
func (h *CourseHandler) Create(c *gin.Context) {
	var curso models.Curso
	if err := c.ShouldBindJSON(&curso); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	// Set status to opened for published course
	curso.Status = models.StatusCursoOpened

	id, err := h.cursoService.Create(c.Request.Context(), &curso)
	if err != nil {
		// Check if it's a validation error (not a database error)
		if strings.Contains(err.Error(), "erro de validação") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": strings.TrimPrefix(err.Error(), "erro de validação: "),
			})
			return
		}
		// Otherwise, treat as database error
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{
			"error": dbErr.GetUserFriendlyMessage(),
			"field": dbErr.Field,
		})
		return
	}

	curso.ID = id
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":         curso.ID,
			"title":      curso.Titulo,
			"status":     curso.Status,
			"created_at": curso.CreatedAt,
			"updated_at": curso.UpdatedAt,
		},
		"message": "Curso criado com sucesso",
	})
}

// @Summary      Criar curso como rascunho
// @Description  Cria um curso e salva como rascunho (status: "draft")
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        request  body      models.Curso  true  "Dados do curso"
// @Success      201      {object}  object
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/courses/draft [post]
func (h *CourseHandler) CreateDraft(c *gin.Context) {
	var curso models.Curso
	if err := c.ShouldBindJSON(&curso); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	// Set status to draft
	curso.Status = models.StatusCursoDraft

	id, err := h.cursoService.Create(c.Request.Context(), &curso)
	if err != nil {
		// Check if it's a validation error (not a database error)
		if strings.Contains(err.Error(), "erro de validação") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": strings.TrimPrefix(err.Error(), "erro de validação: "),
			})
			return
		}
		// Otherwise, treat as database error
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{
			"error": dbErr.GetUserFriendlyMessage(),
			"field": dbErr.Field,
		})
		return
	}

	curso.ID = id
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":         curso.ID,
			"title":      curso.Titulo,
			"status":     curso.Status,
			"created_at": curso.CreatedAt,
			"updated_at": curso.UpdatedAt,
		},
		"message": "Rascunho salvo com sucesso",
	})
}

// @Summary      Editar curso
// @Description  Atualiza um curso existente. Também usado para publicar rascunhos mudando status de "draft" para "opened"
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        courseId path      int         true  "ID do curso"
// @Param        request  body      models.Curso  true  "Dados do curso"
// @Success      200      {object}  object
// @Failure      400      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId} [put]
func (h *CourseHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do curso inválido"})
		return
	}

	// Check if course exists
	existingCurso, err := h.cursoService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar curso: " + err.Error()})
		return
	}
	if existingCurso == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Curso não encontrado"})
		return
	}

	var curso models.Curso
	if err := c.ShouldBindJSON(&curso); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: " + err.Error()})
		return
	}

	curso.ID = id
	// Preserve original created_at
	curso.CreatedAt = existingCurso.CreatedAt

	// Preserve is_visible if not provided in request
	if curso.IsVisible == nil {
		curso.IsVisible = existingCurso.IsVisible
	}

	if err := h.cursoService.Update(c.Request.Context(), &curso); err != nil {
		// Check if it's a validation error (not a database error)
		if strings.Contains(err.Error(), "erro de validação") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": strings.TrimPrefix(err.Error(), "erro de validação: "),
			})
			return
		}
		// Otherwise, treat as database error
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{
			"error": dbErr.GetUserFriendlyMessage(),
			"field": dbErr.Field,
		})
		return
	}

	var message string
	if existingCurso.Status == models.StatusCursoDraft && curso.Status == models.StatusCursoOpened {
		message = "Curso publicado com sucesso"
	} else {
		message = "Curso atualizado com sucesso"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":         curso.ID,
			"title":      curso.Titulo,
			"status":     curso.Status,
			"created_at": curso.CreatedAt,
			"updated_at": curso.UpdatedAt,
		},
		"message": message,
	})
}

// @Summary      Listar cursos criados
// @Description  Retorna lista paginada de cursos criados (status: "opened", "closed", "canceled")
// @Tags         courses
// @Produce      json
// @Param        page         query     int     false  "Número da página (default: 1)"
// @Param        limit        query     int     false  "Tamanho da página (default: 10)"
// @Param        status       query     string  false  "Filtrar por status"
// @Param        modalidade   query     string  false  "Filtrar por modalidade"
// @Param        organization query     string  false  "Filtrar por organização"
// @Param        search       query     string  false  "Buscar no título"
// @Success      200          {object}  object
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/courses [get]
func (h *CourseHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 1000 {
		limit = 10
	}

	// Build filters - exclude drafts
	filter := map[string]interface{}{
		"status NOT": models.StatusCursoDraft,
	}

	// Add additional filters
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	if modalidade := c.Query("modalidade"); modalidade != "" {
		filter["modalidade"] = modalidade
	}

	if organization := c.Query("organization"); organization != "" {
		filter["organization"] = organization
	}

	if search := c.Query("search"); search != "" {
		filter["title ILIKE"] = "%" + search + "%"
	}

	if categoriaID := c.Query("categoria_id"); categoriaID != "" {
		// Convert to int and add to filter
		if catID, err := strconv.Atoi(categoriaID); err == nil {
			filter["categoria_id"] = catID
		}
	}

	cursos, total, err := h.cursoService.List(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar cursos: " + err.Error()})
		return
	}

	// Calculate remaining vacancies for all courses
	for _, curso := range cursos {
		if err := h.calculateRemainingVacancies(c, curso); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular vagas restantes: " + err.Error()})
			return
		}
	}

	totalPages := (total + limit - 1) / limit

	coursesData := make([]gin.H, len(cursos))
	for i, curso := range cursos {
		coursesData[i] = transformCursoToResponse(curso)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"courses": coursesData,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		},
	})
}

// @Summary      Listar cursos rascunho
// @Description  Retorna lista paginada de cursos salvos como rascunho (status: "draft")
// @Tags         courses
// @Produce      json
// @Param        page         query     int     false  "Número da página (default: 1)"
// @Param        limit        query     int     false  "Tamanho da página (default: 10)"
// @Param        organization query     string  false  "Filtrar por organização"
// @Param        search       query     string  false  "Buscar no título"
// @Success      200          {object}  object
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/courses/drafts [get]
func (h *CourseHandler) ListDrafts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 1000 {
		limit = 10
	}

	// Build filters - only drafts
	filter := map[string]interface{}{
		"status": models.StatusCursoDraft,
	}

	if organization := c.Query("organization"); organization != "" {
		filter["organization"] = organization
	}

	if search := c.Query("search"); search != "" {
		filter["title ILIKE"] = "%" + search + "%"
	}

	if categoriaID := c.Query("categoria_id"); categoriaID != "" {
		// Convert to int and add to filter
		if catID, err := strconv.Atoi(categoriaID); err == nil {
			filter["categoria_id"] = catID
		}
	}

	cursos, total, err := h.cursoService.List(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar rascunhos: " + err.Error()})
		return
	}

	// Calculate remaining vacancies for all courses
	for _, curso := range cursos {
		if err := h.calculateRemainingVacancies(c, curso); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular vagas restantes: " + err.Error()})
			return
		}
	}

	totalPages := (total + limit - 1) / limit

	draftsData := make([]gin.H, len(cursos))
	for i, curso := range cursos {
		draftsData[i] = transformCursoToResponse(curso)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"drafts": draftsData,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		},
	})
}

// @Summary      Buscar curso específico
// @Description  Retorna dados completos de um curso específico
// @Tags         courses
// @Produce      json
// @Param        courseId path      int  true  "ID do curso"
// @Success      200      {object}  models.Curso
// @Failure      400      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId} [get]
func (h *CourseHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do curso inválido"})
		return
	}

	curso, err := h.cursoService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar curso: " + err.Error()})
		return
	}

	if curso == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Curso não encontrado"})
		return
	}

	// Calculate remaining vacancies for all schedules
	if err := h.calculateRemainingVacancies(c, curso); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular vagas restantes: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    curso,
	})
}

// @Summary      Excluir curso
// @Description  Remove um curso pelo ID
// @Tags         courses
// @Produce      json
// @Param        courseId path      int  true  "ID do curso"
// @Success      200      {object}  object
// @Failure      400      {object}  models.ErrorResponse
// @Failure      404      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /api/v1/courses/{courseId} [delete]
func (h *CourseHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do curso inválido"})
		return
	}

	// Check if course exists
	curso, err := h.cursoService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar curso: " + err.Error()})
		return
	}
	if curso == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Curso não encontrado"})
		return
	}

	if err := h.cursoService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir curso: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Curso excluído com sucesso",
	})
}

// @Summary      Listar cursos de um usuário específico
// @Description  Retorna lista paginada de cursos criados por um usuário/organização
// @Tags         courses
// @Produce      json
// @Param        userId       path      string  true   "ID do usuário/orgão (external reference)"
// @Param        page         query     int     false  "Número da página (default: 1)"
// @Param        limit        query     int     false  "Tamanho da página (default: 10)"
// @Param        status       query     string  false  "Filtrar por status"
// @Param        modalidade   query     string  false  "Filtrar por modalidade"
// @Success      200          {object}  object
// @Failure      400          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/users/{userId}/courses [get]
func (h *CourseHandler) ListByUser(c *gin.Context) {
	userID := c.Param("userId")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 1000 {
		limit = 10
	}

	// Build filters for user's courses
	filter := map[string]interface{}{
		"orgao_id": userID,
	}

	// Add additional filters
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}

	if modalidade := c.Query("modalidade"); modalidade != "" {
		filter["modalidade"] = modalidade
	}

	if categoriaID := c.Query("categoria_id"); categoriaID != "" {
		// Convert to int and add to filter
		if catID, err := strconv.Atoi(categoriaID); err == nil {
			filter["categoria_id"] = catID
		}
	}

	cursos, total, err := h.cursoService.List(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar cursos do usuário: " + err.Error()})
		return
	}

	// Calculate remaining vacancies for all courses
	for _, curso := range cursos {
		if err := h.calculateRemainingVacancies(c, curso); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular vagas restantes: " + err.Error()})
			return
		}
	}

	totalPages := (total + limit - 1) / limit

	coursesData := make([]gin.H, len(cursos))
	for i, curso := range cursos {
		coursesData[i] = transformCursoToResponse(curso)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"courses": coursesData,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		},
	})
}
