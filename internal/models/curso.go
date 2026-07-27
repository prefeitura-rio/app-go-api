package models

import (
	"errors"
	"strings"
	"time"
)

type Modalidade string
type StatusCurso string
type FormatoAula string
type CourseManagementType string

const (
	ModalidadePresencial     Modalidade = "Presencial"
	ModalidadeSemipresencial Modalidade = "Semipresencial"
	ModalidadeRemoto         Modalidade = "Remoto"

	// Legacy values for backwards compatibility
	ModalidadePresencialLegacy    Modalidade = "PRESENCIAL"
	ModalidadeOnline              Modalidade = "ONLINE"
	ModalidadeHibrido             Modalidade = "HIBRIDO"
	ModalidadeLivreFormacaoOnline Modalidade = "LIVRE_FORMACAO_ONLINE"

	// New status values matching specification
	StatusCursoDraft    StatusCurso = "draft"
	StatusCursoOpened   StatusCurso = "opened"
	StatusCursoClosed   StatusCurso = "closed"
	StatusCursoCanceled StatusCurso = "canceled"

	// Curation flow statuses
	StatusCursoInReview        StatusCurso = "in_review"
	StatusCursoNeedsChanges    StatusCurso = "needs_changes"
	StatusCursoApproved        StatusCurso = "approved"
	StatusCursoPublished       StatusCurso = "published"
	StatusCursoPendingDeletion StatusCurso = "pending_deletion"

	// Derived statuses — computed at read time from dates, never stored in DB
	StatusCursoScheduled            StatusCurso = "scheduled"
	StatusCursoAcceptingEnrollments StatusCurso = "accepting_enrollments"
	StatusCursoInProgress           StatusCurso = "in_progress"
	StatusCursoFinished             StatusCurso = "finished"

	// Legacy values for backwards compatibility
	StatusCursoCriado    StatusCurso = "CRIADO"
	StatusCursoAberto    StatusCurso = "ABERTO"
	StatusCursoEncerrado StatusCurso = "ENCERRADO"

	FormatoAulaGravado FormatoAula = "GRAVADO"
	FormatoAulaAoVivo  FormatoAula = "AO_VIVO"

	CourseManagementOwnOrg                   CourseManagementType = "OWN_ORG"
	CourseManagementExternalManagedByOrg     CourseManagementType = "EXTERNAL_MANAGED_BY_ORG"
	CourseManagementExternalManagedByPartner CourseManagementType = "EXTERNAL_MANAGED_BY_PARTNER"
)

type Curso struct {
	ID int `json:"id" gorm:"primaryKey"`

	// Core fields (always required)
	Titulo              string      `json:"title" gorm:"type:varchar(20000);not null"`
	Descricao           string      `json:"description" gorm:"type:text"`
	EnrollmentStartDate *time.Time  `json:"enrollment_start_date" gorm:"type:timestamp with time zone;column:enrollment_start_date" example:"2026-08-01T00:00:00-03:00"`
	EnrollmentEndDate   *time.Time  `json:"enrollment_end_date" gorm:"type:timestamp with time zone;column:enrollment_end_date" example:"2026-08-23T23:59:59-03:00"`
	Organization        string      `json:"organization" gorm:"type:varchar(20000)"`
	Modalidade          Modalidade  `json:"modalidade" gorm:"type:varchar(50)"`
	Theme               string      `json:"theme" gorm:"type:varchar(20000)"`
	Workload            string      `json:"workload" gorm:"type:varchar(20000)"`
	TargetAudience      string      `json:"target_audience" gorm:"type:varchar(20000);column:target_audience"`
	InstitutionalLogo   string      `json:"institutional_logo" gorm:"type:varchar(20000);column:institutional_logo"`
	CoverImage          string      `json:"cover_image" gorm:"type:varchar(20000);column:cover_image"`
	Status              StatusCurso `json:"status" gorm:"type:varchar(50)"`

	// Legacy and additional fields (optional)
	OrgaoID              string      `json:"orgao_id,omitempty" gorm:"type:varchar(100);column:orgao_id"`
	InstituicaoID        *int        `json:"instituicao_id,omitempty" gorm:"column:instituicao_id"`
	LocalRealizacao      string      `json:"local_realizacao" gorm:"type:varchar(20000);column:local_realizacao"`
	DataInicio           *time.Time  `json:"data_inicio" gorm:"column:data_inicio"`
	DataTermino          *time.Time  `json:"data_termino" gorm:"column:data_termino"`
	DataLimiteInscricoes *time.Time  `json:"data_limite_inscricoes" gorm:"column:data_limite_inscricoes"`
	NumeroVagas          int         `json:"numero_vagas" gorm:"column:numero_vagas"`
	CargaHoraria         int         `json:"carga_horaria" gorm:"column:carga_horaria"`
	Turno                Turno       `json:"turno" gorm:"type:varchar(50)"`
	FormatoAula          FormatoAula `json:"formato_aula" gorm:"column:formato_aula;type:varchar(50)"`
	FormacaoLink         string      `json:"formacao_link,omitempty" gorm:"column:formacao_link;type:varchar(20000)"`
	LinkInscricao        string      `json:"link_inscricao" gorm:"type:varchar(20000);column:link_inscricao"`
	ContatoDuvidas       string      `json:"contato_duvidas" gorm:"type:varchar(20000);column:contato_duvidas"`

	// Optional fields matching specification
	HasCertificate   bool   `json:"has_certificate" gorm:"column:has_certificate;default:false"`
	Facilitator      string `json:"facilitator,omitempty" gorm:"type:varchar(20000)"`
	Objectives       string `json:"objectives,omitempty" gorm:"type:text"`
	ExpectedResults  string `json:"expected_results,omitempty" gorm:"type:text;column:expected_results"`
	ProgramContent   string `json:"program_content,omitempty" gorm:"type:text;column:program_content"`
	Methodology      string `json:"methodology,omitempty" gorm:"type:text"`
	ResourcesUsed    string `json:"resources_used,omitempty" gorm:"type:text;column:resources_used"`
	MaterialUsed     string `json:"material_used,omitempty" gorm:"type:text;column:material_used"`
	TeachingMaterial string `json:"teaching_material,omitempty" gorm:"type:text;column:teaching_material"`

	// Unified field for prerequisites - using single JSON tag
	PreRequisitos         string `json:"pre_requisitos" gorm:"column:pre_requisitos;type:text"`
	CertificacaoOferecida bool   `json:"certificacao_oferecida" gorm:"column:certificacao_oferecida"`

	// External partner fields
	IsExternalPartner      *bool                `json:"is_external_partner,omitempty" gorm:"column:is_external_partner"`
	CourseManagementType   CourseManagementType `json:"course_management_type" gorm:"column:course_management_type;type:varchar(50);default:'OWN_ORG'"`
	ExternalPartnerName    string               `json:"external_partner_name,omitempty" gorm:"column:external_partner_name;type:varchar(20000)"`
	ExternalPartnerURL     string               `json:"external_partner_url,omitempty" gorm:"column:external_partner_url;type:varchar(20000)"`
	ExternalPartnerLogoURL string               `json:"external_partner_logo_url,omitempty" gorm:"column:external_partner_logo_url;type:varchar(20000)"`
	ExternalPartnerContact string               `json:"external_partner_contact,omitempty" gorm:"column:external_partner_contact;type:varchar(20000)"`

	// Accessibility field - free text field for frontend
	Accessibility string `json:"accessibility,omitempty" gorm:"column:accessibility;type:varchar(20000)"`

	// Visibility field - controls if course appears in public listings (default: true)
	// Used for in-person courses that require manual enrollment and should not appear in public course lists
	IsVisible *bool `json:"is_visible,omitempty" gorm:"column:is_visible;default:true" example:"true"`

	// Auto-approve enrollments - when true, new enrollments are automatically approved instead of pending
	AutoApproveEnrollments *bool `json:"auto_approve_enrollments,omitempty" gorm:"column:auto_approve_enrollments;default:false" example:"false"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relacionamentos
	Instituicao     *InstituicaoEnsino `json:"instituicao,omitempty" gorm:"foreignKey:InstituicaoID" swaggerignore:"true"`
	Categorias      []Categoria        `json:"categorias,omitempty" gorm:"many2many:cursos_categorias;" swaggerignore:"true"`
	Acessibilidades []Acessibilidade   `json:"acessibilidades,omitempty" gorm:"many2many:cursos_acessibilidades;" swaggerignore:"true"`

	// New relationships for enrollment system
	CustomFields    []CustomField   `json:"custom_fields" gorm:"foreignKey:CursoID" swaggerignore:"true"`
	Inscricoes      []Inscricao     `json:"enrollments,omitempty" gorm:"foreignKey:CursoID" swaggerignore:"true"`
	LocationClasses []LocationClass `json:"locations,omitempty" gorm:"foreignKey:CursoID" swaggerignore:"true"`
	RemoteClass     *RemoteClass    `json:"remote_class" gorm:"foreignKey:CursoID" swaggerignore:"true"`
}

// TableName especifica o nome da tabela para este modelo
func (Curso) TableName() string {
	return "cursos"
}

// Validation methods
func (m Modalidade) IsValid() bool {
	validValues := []string{
		"Presencial", "Semipresencial", "Remoto",
		"PRESENCIAL", "ONLINE", "HIBRIDO", "LIVRE_FORMACAO_ONLINE",
	}
	for _, v := range validValues {
		if string(m) == v {
			return true
		}
	}
	return false
}

func (s StatusCurso) IsValid() bool {
	validValues := []string{
		"draft", "opened", "closed", "canceled",
		"in_review", "needs_changes", "approved", "published", "pending_deletion",
		// derived statuses — returned by the API, normalized to published on write
		"scheduled", "accepting_enrollments", "in_progress", "finished",
		"CRIADO", "ABERTO", "ENCERRADO",
	}
	for _, v := range validValues {
		if string(s) == v {
			return true
		}
	}
	return false
}

// CanTransitionTo returns true if transitioning from the current status to `to` is allowed.
// Both statuses are normalized before comparison to handle legacy values (ABERTO, ENCERRADO, etc).
func (s StatusCurso) CanTransitionTo(to StatusCurso) bool {
	from := s.Normalize()
	target := to.Normalize()

	allowed := map[StatusCurso][]StatusCurso{
		StatusCursoDraft:        {StatusCursoInReview, StatusCursoPendingDeletion},
		StatusCursoNeedsChanges: {StatusCursoInReview, StatusCursoPendingDeletion},
		StatusCursoInReview:     {StatusCursoApproved, StatusCursoPublished, StatusCursoNeedsChanges, StatusCursoPendingDeletion},
		StatusCursoApproved:     {StatusCursoPublished, StatusCursoPendingDeletion},
		StatusCursoPublished:    {StatusCursoNeedsChanges, StatusCursoPendingDeletion, StatusCursoClosed, StatusCursoCanceled},
		StatusCursoClosed:       {StatusCursoPendingDeletion},
		StatusCursoCanceled:     {StatusCursoPendingDeletion},
	}
	for _, t := range allowed[from] {
		if t == target {
			return true
		}
	}
	return false
}

func (f FormatoAula) IsValid() bool {
	validValues := []string{"GRAVADO", "AO_VIVO", "PRESENCIAL", ""}
	for _, v := range validValues {
		if string(f) == v {
			return true
		}
	}
	return false
}

func (c CourseManagementType) IsValid() bool {
	validValues := []string{
		string(CourseManagementOwnOrg),
		string(CourseManagementExternalManagedByOrg),
		string(CourseManagementExternalManagedByPartner),
		"",
	}
	for _, v := range validValues {
		if string(c) == v {
			return true
		}
	}
	return false
}

// Normalize methods for compatibility
func (s StatusCurso) Normalize() StatusCurso {
	switch strings.ToUpper(string(s)) {
	case "CRIADO":
		return StatusCursoDraft
	case "ABERTO", "OPENED":
		return StatusCursoPublished
	case "ENCERRADO":
		return StatusCursoClosed
	// derived statuses are aliases of published — normalize back so they are never persisted
	case "SCHEDULED", "ACCEPTING_ENROLLMENTS", "IN_PROGRESS", "FINISHED":
		return StatusCursoPublished
	default:
		return s
	}
}

func (m Modalidade) Normalize() Modalidade {
	switch strings.ToUpper(string(m)) {
	case "PRESENCIAL":
		return ModalidadePresencial
	case "ONLINE":
		return ModalidadeRemoto
	case "HIBRIDO":
		return ModalidadeSemipresencial
	default:
		return m
	}
}

func (f FormatoAula) Normalize() FormatoAula {
	normalized := strings.ToUpper(strings.TrimSpace(string(f)))
	if normalized == "" {
		return ""
	}
	return FormatoAula(normalized)
}

// DeriveStatus computes the effective status for a published course based on enrollment and class
// dates. Returns the derived status without mutating the receiver. Non-published courses return
// their current status unchanged.
func (c *Curso) DeriveStatus(now time.Time) StatusCurso {
	if c.Status.Normalize() != StatusCursoPublished {
		return c.Status
	}

	if c.EnrollmentStartDate != nil && now.Before(*c.EnrollmentStartDate) {
		return StatusCursoScheduled
	}

	if c.EnrollmentStartDate != nil && c.EnrollmentEndDate != nil &&
		!now.Before(*c.EnrollmentStartDate) && !now.After(*c.EnrollmentEndDate) {
		return StatusCursoAcceptingEnrollments
	}

	if c.Modalidade == ModalidadeLivreFormacaoOnline {
		if c.EnrollmentEndDate != nil && now.After(*c.EnrollmentEndDate) {
			return StatusCursoInProgress
		}
		return c.Status
	}

	minStart, maxEnd, hasSchedules := collectClassDateRange(c)
	if !hasSchedules {
		return c.Status
	}
	if !now.Before(minStart) && !now.After(maxEnd) {
		return StatusCursoInProgress
	}
	if now.After(maxEnd) {
		return StatusCursoFinished
	}
	return c.Status
}

// IsEnrollmentAvailable reports whether a citizen can still enroll in this course
// right now. It mirrors the frontend getCourseEnrollmentInfo(...).canEnroll logic
// (with no user enrollment) so that API-side ordering matches the enrollment button
// state shown in the portal: a course is available only when its (derived) status
// allows new enrollments, the enrollment window is open, the classes have not ended
// and at least one turma still has vacancies.
//
// RemainingVacancies must already be populated on the schedules (see the handler's
// vacancy calculation); otherwise every turma looks sold out.
func (c *Curso) IsEnrollmentAvailable(now time.Time) bool {
	status := c.DeriveStatus(now)

	// Derived terminal / not-yet-open states are never enrollable.
	switch status {
	case StatusCursoFinished, StatusCursoScheduled:
		return false
	}

	// Stored terminal / non-published states (also covers legacy via Normalize).
	switch status.Normalize() {
	case StatusCursoClosed, StatusCursoCanceled, StatusCursoDraft,
		StatusCursoInReview, StatusCursoNeedsChanges, StatusCursoApproved,
		StatusCursoPendingDeletion:
		return false
	}

	// When the backend has not explicitly derived accepting_enrollments, apply the
	// same date-based checks the frontend does before looking at vacancies.
	if status != StatusCursoAcceptingEnrollments {
		if c.EnrollmentStartDate != nil && now.Before(*c.EnrollmentStartDate) {
			return false // inscrições ainda não abriram
		}
		if end, ok := latestClassEndDate(c); ok && now.After(end) {
			return false // curso encerrado
		}
		if c.EnrollmentEndDate != nil && now.After(*c.EnrollmentEndDate) {
			return false // inscrições encerradas
		}
	}

	// Vacancy check, gated by modalidade exactly like the frontend (hibrido and
	// LIVRE_FORMACAO_ONLINE are intentionally not blocked by vacancies).
	switch strings.ToLower(string(c.Modalidade)) {
	case "online", "remoto":
		if !hasAvailableRemoteVacancies(c) {
			return false // vagas encerradas
		}
	case "presencial", "semipresencial":
		if !hasAvailableInPersonVacancies(c) {
			return false // vagas encerradas
		}
	}

	return true
}

// latestClassEndDate returns the latest class end date across all turmas (remote and
// location schedules), mirroring the frontend getLatestClassEndDate helper.
func latestClassEndDate(c *Curso) (latest time.Time, has bool) {
	apply := func(end time.Time) {
		if !has || end.After(latest) {
			latest, has = end, true
		}
	}
	for _, loc := range c.LocationClasses {
		for _, s := range loc.Schedules {
			apply(s.ClassEndDate)
		}
	}
	if c.RemoteClass != nil {
		for _, s := range c.RemoteClass.Schedules {
			if s.ClassEndDate != nil {
				apply(*s.ClassEndDate)
			}
		}
	}
	return
}

// hasAvailableRemoteVacancies mirrors the negation of the frontend
// hasNoAvailableOnlineClasses for online/remoto courses.
func hasAvailableRemoteVacancies(c *Curso) bool {
	if c.RemoteClass != nil && len(c.RemoteClass.Schedules) > 0 {
		for _, s := range c.RemoteClass.Schedules {
			if s.RemainingVacancies > 0 {
				return true
			}
		}
		return false
	}
	// A remote class with no turmas has no vacancies; no remote class at all means
	// this check does not apply (the frontend does not block enrollment on it).
	return c.RemoteClass == nil
}

// hasAvailableInPersonVacancies mirrors the negation of the frontend
// hasNoAvailableInPersonClasses for presencial/semipresencial courses.
func hasAvailableInPersonVacancies(c *Curso) bool {
	for _, loc := range c.LocationClasses {
		for _, s := range loc.Schedules {
			if s.RemainingVacancies > 0 {
				return true
			}
		}
	}
	return false
}

func collectClassDateRange(c *Curso) (minStart, maxEnd time.Time, has bool) {
	apply := func(start, end time.Time) {
		if !has {
			minStart, maxEnd, has = start, end, true
			return
		}
		if start.Before(minStart) {
			minStart = start
		}
		if end.After(maxEnd) {
			maxEnd = end
		}
	}

	for _, loc := range c.LocationClasses {
		for _, s := range loc.Schedules {
			apply(s.ClassStartDate, s.ClassEndDate)
		}
	}
	if c.RemoteClass != nil {
		for _, s := range c.RemoteClass.Schedules {
			if s.ClassStartDate != nil && s.ClassEndDate != nil {
				apply(*s.ClassStartDate, *s.ClassEndDate)
			}
		}
	}
	return
}

// collectEnrollmentDateRange returns the earliest opening and latest closing
// enrollment dates across all turmas (schedules) of the course.
func collectEnrollmentDateRange(c *Curso) (minStart, maxEnd time.Time, has bool) {
	apply := func(start, end *time.Time) {
		if start == nil || end == nil {
			return
		}
		if !has {
			minStart, maxEnd, has = *start, *end, true
			return
		}
		if start.Before(minStart) {
			minStart = *start
		}
		if end.After(maxEnd) {
			maxEnd = *end
		}
	}

	for _, loc := range c.LocationClasses {
		for _, s := range loc.Schedules {
			apply(s.EnrollmentStartDate, s.EnrollmentEndDate)
		}
	}
	if c.RemoteClass != nil {
		for _, s := range c.RemoteClass.Schedules {
			apply(s.EnrollmentStartDate, s.EnrollmentEndDate)
		}
	}
	return
}

// ApplyDerivedEnrollmentPeriod recomputes the course-level enrollment window from
// its turmas (earliest opening, latest closing) and stores it on the course. This
// keeps the denormalized course period in sync so existing course-level reads,
// status filters and enrollment-window checks keep working unchanged.
// No-op for LIVRE_FORMACAO_ONLINE, which has no turmas and keeps its own dates.
func (c *Curso) ApplyDerivedEnrollmentPeriod() {
	if c.Modalidade == ModalidadeLivreFormacaoOnline {
		return
	}
	minStart, maxEnd, has := collectEnrollmentDateRange(c)
	if !has {
		return
	}
	c.EnrollmentStartDate = &minStart
	c.EnrollmentEndDate = &maxEnd
}

// Validate validates a course instance
func (c *Curso) Validate() error {
	if strings.TrimSpace(c.Titulo) == "" {
		return errors.New("título é obrigatório")
	}

	if !c.Modalidade.IsValid() {
		return errors.New("modalidade inválida")
	}

	if !c.Status.IsValid() {
		return errors.New("status inválido")
	}

	if c.FormatoAula != "" && !c.FormatoAula.IsValid() {
		return errors.New("formato de aula inválido")
	}

	return nil
}
