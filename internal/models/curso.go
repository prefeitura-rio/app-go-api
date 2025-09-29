package models

import (
	"errors"
	"strings"
	"time"
)

type Modalidade string
type StatusCurso string
type FormatoAula string

const (
	ModalidadePresencial     Modalidade = "Presencial"
	ModalidadeSemipresencial Modalidade = "Semipresencial"
	ModalidadeRemoto         Modalidade = "Remoto"

	// Legacy values for backwards compatibility
	ModalidadePresencialLegacy Modalidade = "PRESENCIAL"
	ModalidadeOnline          Modalidade = "ONLINE"
	ModalidadeHibrido         Modalidade = "HIBRIDO"

	// New status values matching specification
	StatusCursoDraft     StatusCurso = "draft"
	StatusCursoOpened    StatusCurso = "opened"
	StatusCursoClosed    StatusCurso = "closed"
	StatusCursoCanceled  StatusCurso = "canceled"

	// Legacy values for backwards compatibility
	StatusCursoCriado    StatusCurso = "CRIADO"
	StatusCursoAberto    StatusCurso = "ABERTO"
	StatusCursoEncerrado StatusCurso = "ENCERRADO"

	FormatoAulaGravado FormatoAula = "GRAVADO"
	FormatoAulaAoVivo  FormatoAula = "AO_VIVO"
)

type Curso struct {
	ID                   int            `json:"id" gorm:"primaryKey"`
	
	// Core fields (always required)
	Titulo               string         `json:"title" gorm:"not null"`
	Descricao            string         `json:"description" gorm:"type:text"`
	EnrollmentStartDate  *time.Time     `json:"enrollment_start_date" gorm:"type:timestamp with time zone;column:enrollment_start_date"`
	EnrollmentEndDate    *time.Time     `json:"enrollment_end_date" gorm:"type:timestamp with time zone;column:enrollment_end_date"`
	Organization         string         `json:"organization" gorm:"type:varchar(255)"`
	Modalidade           Modalidade     `json:"modalidade" gorm:"type:varchar(50)"`
	Theme                string         `json:"theme" gorm:"type:varchar(100)"`
	Workload             string         `json:"workload" gorm:"type:varchar(50)"`
	TargetAudience       string         `json:"target_audience" gorm:"type:varchar(600);column:target_audience"`
	InstitutionalLogo    string         `json:"institutional_logo" gorm:"type:varchar(500);column:institutional_logo"`
	CoverImage           string         `json:"cover_image" gorm:"type:varchar(500);column:cover_image"`
	Status               StatusCurso    `json:"status" gorm:"type:varchar(50)"`

	// Legacy and additional fields
	OrgaoID              int            `json:"orgao_id" gorm:"column:orgao_id"`
	InstituicaoID        int            `json:"instituicao_id" gorm:"column:instituicao_id"`
	LocalRealizacao      string         `json:"local_realizacao" gorm:"column:local_realizacao"`
	DataInicio           *time.Time     `json:"data_inicio" gorm:"column:data_inicio"`
	DataTermino          *time.Time     `json:"data_termino" gorm:"column:data_termino"`
	DataLimiteInscricoes *time.Time     `json:"data_limite_inscricoes" gorm:"column:data_limite_inscricoes"`
	NumeroVagas          int            `json:"numero_vagas" gorm:"column:numero_vagas"`
	CargaHoraria         int            `json:"carga_horaria" gorm:"column:carga_horaria"`
	Turno                Turno          `json:"turno" gorm:"type:varchar(50)"`
	FormatoAula          FormatoAula    `json:"formato_aula" gorm:"column:formato_aula;type:varchar(50)"`
	LinkInscricao        string         `json:"link_inscricao" gorm:"column:link_inscricao"`
	ContatoDuvidas       string         `json:"contato_duvidas" gorm:"column:contato_duvidas"`

	// Optional fields matching specification
	HasCertificate       bool           `json:"has_certificate" gorm:"column:has_certificate;default:false"`
	Facilitator          string         `json:"facilitator,omitempty" gorm:"type:varchar(255)"`
	Objectives           string         `json:"objectives,omitempty" gorm:"type:text"`
	ExpectedResults      string         `json:"expected_results,omitempty" gorm:"type:text;column:expected_results"`
	ProgramContent       string         `json:"program_content,omitempty" gorm:"type:text;column:program_content"`
	Methodology          string         `json:"methodology,omitempty" gorm:"type:text"`
	ResourcesUsed        string         `json:"resources_used,omitempty" gorm:"type:text;column:resources_used"`
	MaterialUsed         string         `json:"material_used,omitempty" gorm:"type:text;column:material_used"`
	TeachingMaterial     string         `json:"teaching_material,omitempty" gorm:"type:text;column:teaching_material"`

	// Unified field for prerequisites - using single JSON tag
	PreRequisitos         string         `json:"pre_requisitos" gorm:"column:pre_requisitos;type:text"`
	CertificacaoOferecida bool          `json:"certificacao_oferecida" gorm:"column:certificacao_oferecida"`

	// External partner fields
	IsExternalPartner       *bool   `json:"is_external_partner,omitempty" gorm:"column:is_external_partner"`
	ExternalPartnerName     string  `json:"external_partner_name,omitempty" gorm:"column:external_partner_name;type:varchar(255)"`
	ExternalPartnerURL      string  `json:"external_partner_url,omitempty" gorm:"column:external_partner_url;type:varchar(500)"`
	ExternalPartnerLogoURL  string  `json:"external_partner_logo_url,omitempty" gorm:"column:external_partner_logo_url;type:varchar(500)"`
	ExternalPartnerContact  string  `json:"external_partner_contact,omitempty" gorm:"column:external_partner_contact;type:varchar(255)"`

	CreatedAt            time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relacionamentos
	Orgao                *Orgao           `json:"orgao,omitempty" gorm:"foreignKey:OrgaoID" swaggerignore:"true"`
	Instituicao          *InstituicaoEnsino `json:"instituicao,omitempty" gorm:"foreignKey:InstituicaoID" swaggerignore:"true"`
	Categorias           []Categoria      `json:"categorias,omitempty" gorm:"many2many:cursos_categorias;" swaggerignore:"true"`
	Acessibilidades      []Acessibilidade `json:"acessibilidades,omitempty" gorm:"many2many:cursos_acessibilidades;" swaggerignore:"true"`

	// New relationships for enrollment system
	CustomFields         []CustomField    `json:"custom_fields" gorm:"foreignKey:CursoID" swaggerignore:"true"`
	Inscricoes           []Inscricao      `json:"enrollments,omitempty" gorm:"foreignKey:CursoID" swaggerignore:"true"`
	LocationClasses      []LocationClass  `json:"locations,omitempty" gorm:"foreignKey:CursoID" swaggerignore:"true"`
	RemoteClass          *RemoteClass     `json:"remote_class" gorm:"foreignKey:CursoID" swaggerignore:"true"`
}

// TableName especifica o nome da tabela para este modelo
func (Curso) TableName() string {
	return "cursos"
}

// Validation methods
func (m Modalidade) IsValid() bool {
	validValues := []string{
		"Presencial", "Semipresencial", "Remoto",
		"PRESENCIAL", "ONLINE", "HIBRIDO",
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
		"CRIADO", "ABERTO", "ENCERRADO",
	}
	for _, v := range validValues {
		if string(s) == v {
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

// Normalize methods for compatibility
func (s StatusCurso) Normalize() StatusCurso {
	switch strings.ToUpper(string(s)) {
	case "CRIADO":
		return StatusCursoDraft
	case "ABERTO":
		return StatusCursoOpened
	case "ENCERRADO":
		return StatusCursoClosed
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