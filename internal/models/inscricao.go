package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type StatusInscricao string

const (
	StatusInscricaoPending   StatusInscricao = "pending"
	StatusInscricaoApproved  StatusInscricao = "approved"
	StatusInscricaoRejected  StatusInscricao = "rejected"
	StatusInscricaoCancelled StatusInscricao = "cancelled"
)

type Inscricao struct {
	ID               uuid.UUID                  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CursoID          int                        `json:"course_id" gorm:"column:curso_id;not null"`
	CPF              string                     `json:"cpf" gorm:"type:varchar(11);not null"`
	Name             string                     `json:"name" gorm:"type:varchar(255);not null"`
	Email            string                     `json:"email" gorm:"type:varchar(255);not null"`
	Phone            string                     `json:"phone" gorm:"type:varchar(20)"`
	Status           StatusInscricao            `json:"status" gorm:"type:status_inscricao_enum;default:'pending'"`
	CustomFieldsData datatypes.JSON            `json:"custom_fields" gorm:"type:jsonb;column:custom_fields_data"`
	AdminNotes       string                     `json:"admin_notes,omitempty" gorm:"type:text;column:admin_notes"`
	Reason           string                     `json:"reason,omitempty" gorm:"type:text"`
	EnrolledAt       time.Time                  `json:"enrolled_at" gorm:"column:enrolled_at;autoCreateTime"`
	UpdatedAt        time.Time                  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`

	// Relacionamentos
	Curso *Curso `json:"curso,omitempty" gorm:"foreignKey:CursoID"`
}

func (Inscricao) TableName() string {
	return "inscricoes"
}

type CustomField struct {
	ID         uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CursoID    int                    `json:"curso_id" gorm:"column:curso_id;not null"`
	Title      string                 `json:"title" gorm:"type:varchar(255);not null"`
	FieldType  string                 `json:"field_type" gorm:"type:varchar(50);default:'text'"`
	Required   bool                   `json:"required" gorm:"default:false"`
	Options    datatypes.JSON         `json:"options,omitempty" gorm:"type:jsonb"`
	CreatedAt  time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time              `json:"updated_at" gorm:"autoUpdateTime"`

	// Relacionamentos
	Curso *Curso `json:"curso,omitempty" gorm:"foreignKey:CursoID"`
}

func (CustomField) TableName() string {
	return "custom_fields"
}

type LocationClass struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CursoID        int       `json:"curso_id" gorm:"column:curso_id;not null"`
	Address        string    `json:"address" gorm:"type:varchar(500);not null"`
	Neighborhood   string    `json:"neighborhood" gorm:"type:varchar(100);not null"`
	Vacancies      int       `json:"vacancies" gorm:"not null"`
	ClassStartDate time.Time `json:"class_start_date" gorm:"type:timestamp with time zone;column:class_start_date;not null"`
	ClassEndDate   time.Time `json:"class_end_date" gorm:"type:timestamp with time zone;column:class_end_date;not null"`
	ClassTime      string    `json:"class_time" gorm:"type:varchar(50);column:class_time;not null"`
	ClassDays      string    `json:"class_days" gorm:"type:varchar(200);column:class_days;not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relacionamentos
	Curso *Curso `json:"curso,omitempty" gorm:"foreignKey:CursoID"`
}

func (LocationClass) TableName() string {
	return "location_classes"
}

type RemoteClass struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CursoID        int       `json:"curso_id" gorm:"column:curso_id;not null"`
	Vacancies      int       `json:"vacancies" gorm:"not null"`
	ClassStartDate time.Time `json:"class_start_date" gorm:"type:timestamp with time zone;column:class_start_date;not null"`
	ClassEndDate   time.Time `json:"class_end_date" gorm:"type:timestamp with time zone;column:class_end_date;not null"`
	ClassTime      string    `json:"class_time" gorm:"type:varchar(50);column:class_time;not null"`
	ClassDays      string    `json:"class_days" gorm:"type:varchar(200);column:class_days;not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relacionamentos
	Curso *Curso `json:"curso,omitempty" gorm:"foreignKey:CursoID"`
}

func (RemoteClass) TableName() string {
	return "remote_classes"
}

type EnrollmentSummary struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Approved  int `json:"approved"`
	Rejected  int `json:"rejected"`
	Cancelled int `json:"cancelled"`
}

type EnrollmentStatusUpdateRequest struct {
	EnrollmentIDs []uuid.UUID     `json:"enrollment_ids" binding:"required"`
	Status        StatusInscricao `json:"status" binding:"required"`
	Reason        string          `json:"reason"`
	AdminNotes    string          `json:"admin_notes"`
}

type EnrollmentStatusUpdateResponse struct {
	UpdatedCount int             `json:"updated_count"`
	Status       StatusInscricao `json:"status"`
	UpdatedAt    time.Time       `json:"updated_at"`
}