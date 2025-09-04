package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

type CourseHandler struct {
	cursoService     *services.CursoService
	inscricaoService *services.InscricaoService
}

func NewCourseHandler(cursoService *services.CursoService, inscricaoService *services.InscricaoService) *CourseHandler {
	return &CourseHandler{
		cursoService:     cursoService,
		inscricaoService: inscricaoService,
	}
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
	
	if err := h.cursoService.Update(c.Request.Context(), &curso); err != nil {
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

	if limit < 1 || limit > 100 {
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

	cursos, total, err := h.cursoService.List(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar cursos: " + err.Error()})
		return
	}

	totalPages := (total + limit - 1) / limit

	// Transform courses to include all fields
	coursesData := make([]gin.H, len(cursos))
	for i, curso := range cursos {
		coursesData[i] = gin.H{
			// Core fields
			"id":                    curso.ID,
			"title":                 curso.Titulo,
			"description":           curso.Descricao,
			"modalidade":            curso.Modalidade,
			"status":                curso.Status,
			"organization":          curso.Organization,
			"theme":                 curso.Theme,
			"workload":              curso.Workload,
			"target_audience":       curso.TargetAudience,
			"institutional_logo":    curso.InstitutionalLogo,
			"cover_image":           curso.CoverImage,
			"enrollment_start_date": curso.EnrollmentStartDate,
			"enrollment_end_date":   curso.EnrollmentEndDate,
			
			// Legacy and additional fields
			"orgao_id":               curso.OrgaoID,
			"instituicao_id":         curso.InstituicaoID,
			"local_realizacao":       curso.LocalRealizacao,
			"data_inicio":            curso.DataInicio,
			"data_termino":           curso.DataTermino,
			"data_limite_inscricoes": curso.DataLimiteInscricoes,
			"numero_vagas":           curso.NumeroVagas,
			"carga_horaria":          curso.CargaHoraria,
			"turno":                  curso.Turno,
			"formato_aula":           curso.FormatoAula,
			"link_inscricao":         curso.LinkInscricao,
			"contato_duvidas":        curso.ContatoDuvidas,
			
			// Optional fields
			"has_certificate":        curso.HasCertificate,
			"facilitator":            curso.Facilitator,
			"objectives":             curso.Objectives,
			"expected_results":       curso.ExpectedResults,
			"program_content":        curso.ProgramContent,
			"methodology":            curso.Methodology,
			"resources_used":         curso.ResourcesUsed,
			"material_used":          curso.MaterialUsed,
			"teaching_material":      curso.TeachingMaterial,
			"pre_requisitos":         curso.PreRequisitos,
			"certificacao_oferecida": curso.CertificacaoOferecida,
			
			// Timestamps
			"created_at":             curso.CreatedAt,
			"updated_at":             curso.UpdatedAt,
			
			// Related objects
			"orgao":                  curso.Orgao,
			"instituicao":            curso.Instituicao,
			"categorias":             curso.Categorias,
			"acessibilidades":        curso.Acessibilidades,
			"custom_fields":          curso.CustomFields,
			"locations":              curso.LocationClasses,
			"remote_class":           curso.RemoteClass,
		}
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

	if limit < 1 || limit > 100 {
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

	cursos, total, err := h.cursoService.List(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar rascunhos: " + err.Error()})
		return
	}

	totalPages := (total + limit - 1) / limit

	// Transform drafts to include all fields like the main course list
	draftsData := make([]gin.H, len(cursos))
	for i, curso := range cursos {
		draftsData[i] = gin.H{
			// Core fields
			"id":                    curso.ID,
			"title":                 curso.Titulo,
			"description":           curso.Descricao,
			"modalidade":            curso.Modalidade,
			"status":                curso.Status,
			"organization":          curso.Organization,
			"theme":                 curso.Theme,
			"workload":              curso.Workload,
			"target_audience":       curso.TargetAudience,
			"institutional_logo":    curso.InstitutionalLogo,
			"cover_image":           curso.CoverImage,
			"enrollment_start_date": curso.EnrollmentStartDate,
			"enrollment_end_date":   curso.EnrollmentEndDate,
			
			// Legacy and additional fields
			"orgao_id":               curso.OrgaoID,
			"instituicao_id":         curso.InstituicaoID,
			"local_realizacao":       curso.LocalRealizacao,
			"data_inicio":            curso.DataInicio,
			"data_termino":           curso.DataTermino,
			"data_limite_inscricoes": curso.DataLimiteInscricoes,
			"numero_vagas":           curso.NumeroVagas,
			"carga_horaria":          curso.CargaHoraria,
			"turno":                  curso.Turno,
			"formato_aula":           curso.FormatoAula,
			"link_inscricao":         curso.LinkInscricao,
			"contato_duvidas":        curso.ContatoDuvidas,
			
			// Optional fields
			"has_certificate":        curso.HasCertificate,
			"facilitator":            curso.Facilitator,
			"objectives":             curso.Objectives,
			"expected_results":       curso.ExpectedResults,
			"program_content":        curso.ProgramContent,
			"methodology":            curso.Methodology,
			"resources_used":         curso.ResourcesUsed,
			"material_used":          curso.MaterialUsed,
			"teaching_material":      curso.TeachingMaterial,
			"pre_requisitos":         curso.PreRequisitos,
			"certificacao_oferecida": curso.CertificacaoOferecida,
			
			// Timestamps
			"created_at":             curso.CreatedAt,
			"updated_at":             curso.UpdatedAt,
			
			// Related objects
			"orgao":                  curso.Orgao,
			"instituicao":            curso.Instituicao,
			"categorias":             curso.Categorias,
			"acessibilidades":        curso.Acessibilidades,
			"custom_fields":          curso.CustomFields,
			"locations":              curso.LocationClasses,
			"remote_class":           curso.RemoteClass,
		}
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
// @Param        userId       path      int     true   "ID do usuário/orgão"
// @Param        page         query     int     false  "Número da página (default: 1)"
// @Param        limit        query     int     false  "Tamanho da página (default: 10)"
// @Param        status       query     string  false  "Filtrar por status"
// @Param        modalidade   query     string  false  "Filtrar por modalidade"
// @Success      200          {object}  object
// @Failure      400          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /api/v1/users/{userId}/courses [get]
func (h *CourseHandler) ListByUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do usuário inválido"})
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

	cursos, total, err := h.cursoService.List(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar cursos do usuário: " + err.Error()})
		return
	}

	totalPages := (total + limit - 1) / limit

	// Transform courses to include all fields
	coursesData := make([]gin.H, len(cursos))
	for i, curso := range cursos {
		coursesData[i] = gin.H{
			// Core fields
			"id":                    curso.ID,
			"title":                 curso.Titulo,
			"description":           curso.Descricao,
			"modalidade":            curso.Modalidade,
			"status":                curso.Status,
			"organization":          curso.Organization,
			"theme":                 curso.Theme,
			"workload":              curso.Workload,
			"target_audience":       curso.TargetAudience,
			"institutional_logo":    curso.InstitutionalLogo,
			"cover_image":           curso.CoverImage,
			"enrollment_start_date": curso.EnrollmentStartDate,
			"enrollment_end_date":   curso.EnrollmentEndDate,
			
			// Legacy and additional fields
			"orgao_id":               curso.OrgaoID,
			"instituicao_id":         curso.InstituicaoID,
			"local_realizacao":       curso.LocalRealizacao,
			"data_inicio":            curso.DataInicio,
			"data_termino":           curso.DataTermino,
			"data_limite_inscricoes": curso.DataLimiteInscricoes,
			"numero_vagas":           curso.NumeroVagas,
			"carga_horaria":          curso.CargaHoraria,
			"turno":                  curso.Turno,
			"formato_aula":           curso.FormatoAula,
			"link_inscricao":         curso.LinkInscricao,
			"contato_duvidas":        curso.ContatoDuvidas,
			
			// Optional fields
			"has_certificate":        curso.HasCertificate,
			"facilitator":            curso.Facilitator,
			"objectives":             curso.Objectives,
			"expected_results":       curso.ExpectedResults,
			"program_content":        curso.ProgramContent,
			"methodology":            curso.Methodology,
			"resources_used":         curso.ResourcesUsed,
			"material_used":          curso.MaterialUsed,
			"teaching_material":      curso.TeachingMaterial,
			"pre_requisitos":         curso.PreRequisitos,
			"certificacao_oferecida": curso.CertificacaoOferecida,
			
			// Timestamps
			"created_at":             curso.CreatedAt,
			"updated_at":             curso.UpdatedAt,
			
			// Related objects
			"orgao":                  curso.Orgao,
			"instituicao":            curso.Instituicao,
			"categorias":             curso.Categorias,
			"acessibilidades":        curso.Acessibilidades,
			"custom_fields":          curso.CustomFields,
			"locations":              curso.LocationClasses,
			"remote_class":           curso.RemoteClass,
		}
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