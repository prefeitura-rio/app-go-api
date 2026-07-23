package v1

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/cache"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/prefeitura-rio/app-go-api/internal/utils"
)

type CourseHandler struct {
	cursoService     services.CursoServiceInterface
	inscricaoService services.InscricaoServiceInterface
	cursoRepo        services.CursoRepositoryInterface
	courseCache      *cache.CourseCache
}

func NewCourseHandler(cursoService services.CursoServiceInterface, inscricaoService services.InscricaoServiceInterface, cursoRepo services.CursoRepositoryInterface) *CourseHandler {
	return &CourseHandler{
		cursoService:     cursoService,
		inscricaoService: inscricaoService,
		cursoRepo:        cursoRepo,
		courseCache:      nil,
	}
}

// WithCache adds cache support to the handler
func (h *CourseHandler) WithCache(cache *cache.CourseCache) *CourseHandler {
	h.courseCache = cache
	return h
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
		"formacao_link":             curso.FormacaoLink,
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
		"course_management_type":    curso.CourseManagementType,
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
		"auto_approve_enrollments":  curso.AutoApproveEnrollments,
		"custom_fields":             curso.CustomFields,
		"locations":                 curso.LocationClasses,
		"remote_class":              curso.RemoteClass,
	}
}

// calculateRemainingVacancies calculates remaining vacancies for all schedules in a course
func (h *CourseHandler) calculateRemainingVacancies(c *gin.Context, curso *models.Curso) error {
	// Collect all schedule IDs (both location and remote)
	var scheduleIDs []uuid.UUID

	// Collect location schedule IDs
	for _, location := range curso.LocationClasses {
		for _, schedule := range location.Schedules {
			scheduleIDs = append(scheduleIDs, schedule.ID)
		}
	}

	// Collect remote schedule IDs
	if curso.RemoteClass != nil {
		for _, schedule := range curso.RemoteClass.Schedules {
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

	// Update location schedules with remaining vacancies
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

	// Update remote schedules with remaining vacancies
	if curso.RemoteClass != nil {
		for j := range curso.RemoteClass.Schedules {
			schedule := &curso.RemoteClass.Schedules[j]
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

// calculateRemainingVacanciesForCourses calculates remaining vacancies for multiple courses
// in a single batched query to eliminate N+1 query problem in list endpoints
func (h *CourseHandler) calculateRemainingVacanciesForCourses(c *gin.Context, cursos []*models.Curso) error {
	if len(cursos) == 0 {
		return nil
	}

	// Collect all schedule IDs from all courses (both location and remote)
	var allScheduleIDs []uuid.UUID
	for _, curso := range cursos {
		// Collect location schedule IDs
		for _, location := range curso.LocationClasses {
			for _, schedule := range location.Schedules {
				allScheduleIDs = append(allScheduleIDs, schedule.ID)
			}
		}

		// Collect remote schedule IDs
		if curso.RemoteClass != nil {
			for _, schedule := range curso.RemoteClass.Schedules {
				allScheduleIDs = append(allScheduleIDs, schedule.ID)
			}
		}
	}

	// If no schedules at all, nothing to do
	if len(allScheduleIDs) == 0 {
		return nil
	}

	// Get enrollment counts for ALL schedules in a SINGLE query
	enrollmentCounts, err := h.cursoRepo.CountEnrollmentsByScheduleIDs(c.Request.Context(), allScheduleIDs)
	if err != nil {
		return err
	}

	// Apply counts to each course's schedules
	for _, curso := range cursos {
		// Update location schedules
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

		// Update remote schedules
		if curso.RemoteClass != nil {
			for j := range curso.RemoteClass.Schedules {
				schedule := &curso.RemoteClass.Schedules[j]
				enrolledCount := enrollmentCounts[schedule.ID]
				schedule.RemainingVacancies = schedule.Vacancies - int(enrolledCount)

				// Ensure it doesn't go negative
				if schedule.RemainingVacancies < 0 {
					schedule.RemainingVacancies = 0
				}
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

	// A criação de curso é restrita a admin / go:cursos:casa_civil pelo
	// CourseOrgaoInjector (dev/test) e, em produção, pelo gateway Heimdall — e
	// esses papéis podem criar em qualquer órgão. Não há, portanto, filtro de
	// órgão adicional a aplicar aqui.
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

	// Invalidate course list cache after create (if caching is enabled)
	if h.courseCache != nil {
		_ = h.courseCache.InvalidateAll(c.Request.Context())
	}

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

	// A criação de curso é restrita a admin / go:cursos:casa_civil pelo
	// CourseOrgaoInjector (dev/test) e, em produção, pelo gateway Heimdall — e
	// esses papéis podem criar em qualquer órgão. Não há, portanto, filtro de
	// órgão adicional a aplicar aqui.
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

	if h.courseCache != nil {
		_ = h.courseCache.InvalidateAll(c.Request.Context())
	}

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

	// Preserve auto_approve_enrollments if not provided in request
	if curso.AutoApproveEnrollments == nil {
		curso.AutoApproveEnrollments = existingCurso.AutoApproveEnrollments
	}

	// Preserve accepting_enrollments for existing schedules if not provided in request
	existingScheduleValues := make(map[string]*bool)
	for _, loc := range existingCurso.LocationClasses {
		for _, sched := range loc.Schedules {
			existingScheduleValues[sched.ID.String()] = sched.AcceptingEnrollments
		}
	}
	if existingCurso.RemoteClass != nil {
		for _, sched := range existingCurso.RemoteClass.Schedules {
			existingScheduleValues[sched.ID.String()] = sched.AcceptingEnrollments
		}
	}
	zeroUUID := "00000000-0000-0000-0000-000000000000"
	for i := range curso.LocationClasses {
		for j := range curso.LocationClasses[i].Schedules {
			sid := curso.LocationClasses[i].Schedules[j].ID.String()
			if curso.LocationClasses[i].Schedules[j].AcceptingEnrollments == nil && sid != zeroUUID {
				if existing, ok := existingScheduleValues[sid]; ok {
					curso.LocationClasses[i].Schedules[j].AcceptingEnrollments = existing
				}
			}
		}
	}
	if curso.RemoteClass != nil {
		for j := range curso.RemoteClass.Schedules {
			sid := curso.RemoteClass.Schedules[j].ID.String()
			if curso.RemoteClass.Schedules[j].AcceptingEnrollments == nil && sid != zeroUUID {
				if existing, ok := existingScheduleValues[sid]; ok {
					curso.RemoteClass.Schedules[j].AcceptingEnrollments = existing
				}
			}
		}
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

	// Invalidate course list cache after update (if caching is enabled)
	if h.courseCache != nil {
		_ = h.courseCache.InvalidateAll(c.Request.Context())
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

func parseStatusFilter(param string) []string {
	validStatuses := map[string]bool{
		"draft": true, "opened": true, "closed": true, "canceled": true,
		"in_review": true, "needs_changes": true, "approved": true,
		"published": true, "pending_deletion": true,
		"CRIADO": true, "ABERTO": true, "ENCERRADO": true,
		"scheduled": true, "accepting_enrollments": true,
		"in_progress": true, "finished": true,
	}

	seen := map[string]bool{}
	var result []string
	for _, s := range strings.Split(param, ",") {
		s = strings.TrimSpace(s)
		if validStatuses[s] && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// @Summary      Listar cursos criados
// @Description  Retorna lista paginada de cursos públicos. Exclui rascunhos e cursos marcados como não visíveis publicamente (is_visible=false); estes continuam acessíveis por link direto via GET /api/public/courses/{id}. Com status: filtra pelos status informados (CSV). Derived statuses (scheduled, accepting_enrollments, in_progress, finished) são mapeados para published.
// @Tags         courses
// @Produce      json
// @Param        page              query     int     false  "Número da página (default: 1)"
// @Param        limit             query     int     false  "Tamanho da página (default: 10)"
// @Param        status            query     string  false  "Filtrar por status (CSV, ex: draft,published,in_review)"
// @Param        modalidade        query     string  false  "Filtrar por modalidade"
// @Param        organization      query     string  false  "Filtrar por organização (provedor do curso)"
// @Param        search            query     string  false  "Buscar no título"
// @Param        categoria_id      query     int     false  "Filtrar por categoria"
// @Param        acessibilidade_id query     int     false  "Filtrar por acessibilidade"
// @Param        neighborhood_zone query     string  false  "Filtrar por zona do bairro"
// @Param        sort              query     string  false  "Ordenação opcional. 'availability': cursos com inscrição disponível primeiro; os sem inscrição (status final, prazo vencido ou vagas esgotadas — inclusive accepting_enrollments sem vagas) vão para o final. A ordenação é aplicada antes da paginação. Omitido = ordem padrão (id desc)."  Enums(availability)
// @Success      200               {object}  object
// @Failure      500               {object}  models.ErrorResponse
// @Router       /api/public/courses [get]
func (h *CourseHandler) ListPublic(c *gin.Context) {
	h.list(c, true)
}

// @Summary      Listar cursos criados
// @Description  Retorna lista paginada de cursos. Sem status: exclui rascunhos. Com status: filtra pelos status informados (CSV). Derived statuses (scheduled, accepting_enrollments, in_progress, finished) são mapeados para published.
// @Tags         courses
// @Produce      json
// @Param        page              query     int     false  "Número da página (default: 1)"
// @Param        limit             query     int     false  "Tamanho da página (default: 10)"
// @Param        status            query     string  false  "Filtrar por status (CSV, ex: draft,published,in_review)"
// @Param        modalidade        query     string  false  "Filtrar por modalidade"
// @Param        organization      query     string  false  "Filtrar por organização (provedor do curso)"
// @Param        search            query     string  false  "Buscar no título"
// @Param        categoria_id      query     int     false  "Filtrar por categoria"
// @Param        acessibilidade_id query     int     false  "Filtrar por acessibilidade"
// @Param        neighborhood_zone query     string  false  "Filtrar por zona do bairro"
// @Param        sort              query     string  false  "Ordenação opcional. 'availability': cursos com inscrição disponível primeiro; os sem inscrição (status final, prazo vencido ou vagas esgotadas — inclusive accepting_enrollments sem vagas) vão para o final. A ordenação é aplicada antes da paginação. Omitido = ordem padrão (id desc)."  Enums(availability)
// @Success      200               {object}  object
// @Failure      500               {object}  models.ErrorResponse
// @Router       /api/v1/courses [get]
func (h *CourseHandler) List(c *gin.Context) {
	h.list(c, false)
}

// list backs both the admin (/api/v1/courses) and public (/api/public/courses)
// listings. When publicOnly is true, courses flagged as not publicly visible
// (is_visible=false) are excluded so they are reachable only by direct link; the
// admin listing always includes them.
func (h *CourseHandler) list(c *gin.Context, publicOnly bool) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 1000 {
		limit = 10
	}

	filter := map[string]interface{}{}

	for k, v := range middlewares.GetCourseFilters(c) {
		filter[k] = v
	}

	if _, noResults := filter["_no_results"]; noResults {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"courses": []interface{}{},
				"pagination": gin.H{
					"page":        page,
					"limit":       limit,
					"total":       0,
					"total_pages": 0,
				},
			},
		})
		return
	}

	if statusParam := c.Query("status"); statusParam != "" {
		if statuses := parseStatusFilter(statusParam); len(statuses) > 0 {
			filter["status IN"] = statuses
		} else {
			filter["status NOT"] = models.StatusCursoDraft
		}
	} else {
		filter["status NOT"] = models.StatusCursoDraft
	}

	// Public listing hides courses marked as not publicly visible. They remain
	// reachable by direct link (GetByIDPublic) and in the admin listing.
	if publicOnly {
		filter["is_visible NOT_FALSE"] = true
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
		if catID, err := strconv.Atoi(categoriaID); err == nil {
			filter["categoria_id"] = catID
		}
	}

	if acessibilidadeID := c.Query("acessibilidade_id"); acessibilidadeID != "" {
		if accID, err := strconv.Atoi(acessibilidadeID); err == nil {
			filter["acessibilidade_id"] = accID
		}
	}

	if neighborhoodZone := c.Query("neighborhood_zone"); neighborhoodZone != "" {
		filter["neighborhood_zone"] = neighborhoodZone
	}

	// Optional ordering: put courses a citizen can still enroll in first.
	sortByAvailability := strings.EqualFold(strings.TrimSpace(c.Query("sort")), sortAvailability)

	// The cache key must vary by ordering so the sorted and unsorted variants of the
	// same filter/page do not collide. The sort marker is kept out of `filter` (which
	// is passed to the repository query) to avoid it being treated as a column filter.
	filterHash := cache.HashFilter(cacheFilter(filter, sortByAvailability), page, limit)

	// Try cache first (if caching is enabled)
	if h.courseCache != nil {
		cachedData, err := h.courseCache.GetList(c.Request.Context(), filterHash)
		if err == nil && cachedData != nil {
			var cachedResponse gin.H
			if err := json.Unmarshal(cachedData, &cachedResponse); err == nil {
				c.JSON(http.StatusOK, cachedResponse)
				return
			}
			// If unmarshal fails, fall through to database query
		}
	}

	var cursos []*models.Curso
	var total int

	if sortByAvailability {
		var err error
		cursos, total, err = h.listSortedByAvailability(c, filter, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar cursos: " + err.Error()})
			return
		}
	} else {
		var err error
		cursos, total, err = h.cursoService.List(c.Request.Context(), filter, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar cursos: " + err.Error()})
			return
		}

		// Calculate remaining vacancies for all courses in a single batched query
		if err := h.calculateRemainingVacanciesForCourses(c, cursos); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular vagas restantes: " + err.Error()})
			return
		}
	}

	totalPages := (total + limit - 1) / limit

	coursesData := make([]gin.H, len(cursos))
	for i, curso := range cursos {
		coursesData[i] = transformCursoToResponse(curso)
	}

	response := gin.H{
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
	}

	// Cache the result (if caching is enabled)
	if h.courseCache != nil {
		_ = h.courseCache.SetList(c.Request.Context(), filterHash, response)
	}

	c.JSON(http.StatusOK, response)
}

// sortAvailability is the value of the `sort` query param that orders courses by
// enrollment availability (available first, sold-out/closed last).
const sortAvailability = "availability"

// cacheFilter returns a copy of filter augmented with the ordering marker when
// availability sorting is active, so cached responses key on the ordering too. The
// original filter is never mutated (it is used verbatim by the repository query).
func cacheFilter(filter map[string]interface{}, sortByAvailability bool) map[string]interface{} {
	if !sortByAvailability {
		return filter
	}
	out := make(map[string]interface{}, len(filter)+1)
	for k, v := range filter {
		out[k] = v
	}
	out["_sort"] = sortAvailability
	return out
}

// listSortedByAvailability fetches every course matching filter, orders those a
// citizen can still enroll in ahead of the rest, and returns the requested page.
// Ordering must happen before pagination, so the full result set is materialized
// and sorted in memory; within each group the repository's id-desc order is kept.
func (h *CourseHandler) listSortedByAvailability(c *gin.Context, filter map[string]interface{}, page, limit int) ([]*models.Curso, int, error) {
	// limit <= 0 tells the repository to skip the SQL LIMIT and return every match.
	allCursos, total, err := h.cursoRepo.List(c.Request.Context(), filter, -1, 0)
	if err != nil {
		return nil, 0, err
	}

	// Vacancies feed the availability check, so they must be computed before sorting.
	if err := h.calculateRemainingVacanciesForCourses(c, allCursos); err != nil {
		return nil, 0, err
	}

	now := time.Now()
	type rankedCurso struct {
		curso     *models.Curso
		available bool
	}
	ranked := make([]rankedCurso, len(allCursos))
	for i, curso := range allCursos {
		ranked[i] = rankedCurso{curso: curso, available: curso.IsEnrollmentAvailable(now)}
	}
	// Stable so the repository's id-desc order is preserved within each group.
	sort.SliceStable(ranked, func(a, b int) bool {
		return ranked[a].available && !ranked[b].available
	})

	start := (page - 1) * limit
	if start > len(ranked) {
		start = len(ranked)
	}
	end := start + limit
	if end > len(ranked) {
		end = len(ranked)
	}

	pageCursos := make([]*models.Curso, 0, end-start)
	for _, r := range ranked[start:end] {
		// Apply the derived status for the response, mirroring CursoService.List.
		r.curso.Status = r.curso.DeriveStatus(now)
		pageCursos = append(pageCursos, r.curso)
	}
	return pageCursos, total, nil
}

// @Summary      Listar cursos rascunho
// @Description  Retorna lista paginada de cursos salvos como rascunho (status: "draft")
// @Tags         courses
// @Produce      json
// @Param        page              query     int     false  "Número da página (default: 1)"
// @Param        limit             query     int     false  "Tamanho da página (default: 10)"
// @Param        organization      query     string  false  "Filtrar por organização (provedor do curso)"
// @Param        search            query     string  false  "Buscar no título"
// @Param        modalidade        query     string  false  "Filtrar por modalidade"
// @Param        categoria_id      query     int     false  "Filtrar por categoria"
// @Param        acessibilidade_id query     int     false  "Filtrar por acessibilidade"
// @Param        neighborhood_zone query     string  false  "Filtrar por zona do bairro"
// @Success      200               {object}  object
// @Failure      500               {object}  models.ErrorResponse
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

	filter := map[string]interface{}{
		"status": models.StatusCursoDraft,
	}

	for k, v := range middlewares.GetCourseFilters(c) {
		filter[k] = v
	}

	if _, noResults := filter["_no_results"]; noResults {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"courses":    []interface{}{},
				"pagination": gin.H{"page": page, "limit": limit, "total": 0, "total_pages": 0},
			},
		})
		return
	}

	if organization := c.Query("organization"); organization != "" {
		filter["organization"] = organization
	}

	if search := c.Query("search"); search != "" {
		filter["title ILIKE"] = "%" + search + "%"
	}

	if modalidade := c.Query("modalidade"); modalidade != "" {
		filter["modalidade"] = modalidade
	}

	if categoriaID := c.Query("categoria_id"); categoriaID != "" {
		if catID, err := strconv.Atoi(categoriaID); err == nil {
			filter["categoria_id"] = catID
		}
	}

	if acessibilidadeID := c.Query("acessibilidade_id"); acessibilidadeID != "" {
		if accID, err := strconv.Atoi(acessibilidadeID); err == nil {
			filter["acessibilidade_id"] = accID
		}
	}

	if neighborhoodZone := c.Query("neighborhood_zone"); neighborhoodZone != "" {
		filter["neighborhood_zone"] = neighborhoodZone
	}

	// Try cache first (if caching is enabled)
	if h.courseCache != nil {
		filterHash := cache.HashFilter(filter, page, limit)
		cachedData, err := h.courseCache.GetList(c.Request.Context(), filterHash)
		if err == nil && cachedData != nil {
			var cachedResponse gin.H
			if err := json.Unmarshal(cachedData, &cachedResponse); err == nil {
				c.JSON(http.StatusOK, cachedResponse)
				return
			}
			// If unmarshal fails, fall through to database query
		}
	}

	cursos, total, err := h.cursoService.List(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar rascunhos: " + err.Error()})
		return
	}

	// Calculate remaining vacancies for all courses in a single batched query
	if err := h.calculateRemainingVacanciesForCourses(c, cursos); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular vagas restantes: " + err.Error()})
		return
	}

	totalPages := (total + limit - 1) / limit

	draftsData := make([]gin.H, len(cursos))
	for i, curso := range cursos {
		draftsData[i] = transformCursoToResponse(curso)
	}

	response := gin.H{
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
	}

	// Cache the result (if caching is enabled)
	if h.courseCache != nil {
		filterHash := cache.HashFilter(filter, page, limit)
		_ = h.courseCache.SetList(c.Request.Context(), filterHash, response)
	}

	c.JSON(http.StatusOK, response)
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
// @Router       /api/public/courses/{courseId} [get]
func (h *CourseHandler) GetByIDPublic(c *gin.Context) {
	h.getByID(c, true)
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
	h.getByID(c, false)
}

// getByID busca um curso por ID. Quando publicOnly é true (rota /api/public),
// cursos em rascunho não são expostos — retornam 404 —, mantendo o mesmo
// contrato da listagem pública, que já esconde rascunhos (status NOT draft).
func (h *CourseHandler) getByID(c *gin.Context, publicOnly bool) {
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

	if publicOnly && curso.Status.Normalize() == models.StatusCursoDraft {
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

	if err := h.cursoService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir curso: " + err.Error()})
		return
	}

	// Invalidate course list cache after delete (if caching is enabled)
	if h.courseCache != nil {
		_ = h.courseCache.InvalidateAll(c.Request.Context())
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
// @Param        userId            path      string  true   "ID do usuário/orgão (external reference)"
// @Param        page              query     int     false  "Número da página (default: 1)"
// @Param        limit             query     int     false  "Tamanho da página (default: 10)"
// @Param        status            query     string  false  "Filtrar por status"
// @Param        modalidade        query     string  false  "Filtrar por modalidade"
// @Param        categoria_id      query     int     false  "Filtrar por categoria"
// @Param        acessibilidade_id query     int     false  "Filtrar por acessibilidade"
// @Param        neighborhood_zone query     string  false  "Filtrar por zona do bairro"
// @Success      200               {object}  object
// @Failure      400               {object}  models.ErrorResponse
// @Failure      500               {object}  models.ErrorResponse
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
		if catID, err := strconv.Atoi(categoriaID); err == nil {
			filter["categoria_id"] = catID
		}
	}

	if acessibilidadeID := c.Query("acessibilidade_id"); acessibilidadeID != "" {
		if accID, err := strconv.Atoi(acessibilidadeID); err == nil {
			filter["acessibilidade_id"] = accID
		}
	}

	if neighborhoodZone := c.Query("neighborhood_zone"); neighborhoodZone != "" {
		filter["neighborhood_zone"] = neighborhoodZone
	}

	// Try cache first (if caching is enabled)
	if h.courseCache != nil {
		filterHash := cache.HashFilter(filter, page, limit)
		cachedData, err := h.courseCache.GetList(c.Request.Context(), filterHash)
		if err == nil && cachedData != nil {
			var cachedResponse gin.H
			if err := json.Unmarshal(cachedData, &cachedResponse); err == nil {
				c.JSON(http.StatusOK, cachedResponse)
				return
			}
			// If unmarshal fails, fall through to database query
		}
	}

	cursos, total, err := h.cursoService.List(c.Request.Context(), filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar cursos do usuário: " + err.Error()})
		return
	}

	// Calculate remaining vacancies for all courses in a single batched query
	if err := h.calculateRemainingVacanciesForCourses(c, cursos); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao calcular vagas restantes: " + err.Error()})
		return
	}

	totalPages := (total + limit - 1) / limit

	coursesData := make([]gin.H, len(cursos))
	for i, curso := range cursos {
		coursesData[i] = transformCursoToResponse(curso)
	}

	response := gin.H{
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
	}

	// Cache the result (if caching is enabled)
	if h.courseCache != nil {
		filterHash := cache.HashFilter(filter, page, limit)
		_ = h.courseCache.SetList(c.Request.Context(), filterHash, response)
	}

	c.JSON(http.StatusOK, response)
}

func handleCursoTransitionError(c *gin.Context, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "não encontrado"):
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
	case strings.Contains(msg, "não está em estado"):
		c.JSON(http.StatusConflict, gin.H{"error": msg})
	case strings.Contains(msg, "curso incompleto"):
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	default:
		dbErr := utils.ParseDatabaseError(err)
		c.JSON(dbErr.GetHTTPStatusCode(), gin.H{"error": dbErr.GetUserFriendlyMessage()})
	}
}

func (h *CourseHandler) SendToReview(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.cursoService.SendToReview(c.Request.Context(), id); err != nil {
		handleCursoTransitionError(c, err)
		return
	}
	if h.courseCache != nil {
		_ = h.courseCache.InvalidateAll(c.Request.Context())
	}
	c.JSON(http.StatusOK, gin.H{"message": "Curso enviado para revisão com sucesso"})
}

func (h *CourseHandler) Approve(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.cursoService.Approve(c.Request.Context(), id); err != nil {
		handleCursoTransitionError(c, err)
		return
	}
	if h.courseCache != nil {
		_ = h.courseCache.InvalidateAll(c.Request.Context())
	}
	c.JSON(http.StatusOK, gin.H{"message": "Curso aprovado e publicado com sucesso"})
}

func (h *CourseHandler) RequestChanges(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.cursoService.RequestChanges(c.Request.Context(), id); err != nil {
		handleCursoTransitionError(c, err)
		return
	}
	if h.courseCache != nil {
		_ = h.courseCache.InvalidateAll(c.Request.Context())
	}
	c.JSON(http.StatusOK, gin.H{"message": "Solicitação de edição enviada com sucesso"})
}

func (h *CourseHandler) RequestDeletion(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("courseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.cursoService.RequestDeletion(c.Request.Context(), id); err != nil {
		handleCursoTransitionError(c, err)
		return
	}
	if h.courseCache != nil {
		_ = h.courseCache.InvalidateAll(c.Request.Context())
	}
	c.JSON(http.StatusOK, gin.H{"message": "Solicitação de exclusão enviada com sucesso"})
}
