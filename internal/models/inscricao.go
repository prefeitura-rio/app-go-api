package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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
	StatusInscricaoConcluded StatusInscricao = "concluded"
)

type Inscricao struct {
	ID               uuid.UUID                  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CursoID          int                        `json:"course_id" gorm:"column:curso_id;not null"`
	CPF              string                     `json:"cpf" gorm:"type:varchar(11);not null"`
	Name             string                     `json:"name" gorm:"type:varchar(255);not null"`
	Email            string                     `json:"email" gorm:"type:varchar(255);not null"`
	Phone            string                     `json:"phone" gorm:"type:varchar(20)"`

	// Additional enrollment fields for manual/bulk import
	Age              int                        `json:"age,omitempty" gorm:"column:age" example:"25"`                                         // Idade do inscrito
	Address          string                     `json:"address,omitempty" gorm:"type:varchar(500);column:address" example:"Rua das Flores, 123"` // Endereço completo
	Neighborhood     string                     `json:"neighborhood,omitempty" gorm:"type:varchar(100);column:neighborhood" example:"Centro"`     // Bairro

	Status           StatusInscricao            `json:"status" gorm:"type:status_inscricao_enum;default:'pending'"`
	CustomFieldsData datatypes.JSON            `json:"custom_fields" gorm:"type:jsonb;column:custom_fields_data" swaggertype:"object"`
	AdminNotes       string                     `json:"admin_notes,omitempty" gorm:"type:text;column:admin_notes"`
	Reason           string                     `json:"reason,omitempty" gorm:"type:text"`
	CertificateURL   string                     `json:"certificate_url,omitempty" gorm:"type:varchar(500);column:certificate_url"`
	EnrolledUnit     *EnrolledUnit              `json:"enrolled_unit,omitempty" gorm:"type:jsonb;column:enrolled_unit"`
	EnrolledAt       time.Time                  `json:"enrolled_at" gorm:"column:enrolled_at;autoCreateTime"`
	UpdatedAt        time.Time                  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	ConcludedAt      *time.Time                 `json:"concluded_at,omitempty" gorm:"column:concluded_at"`

	// Relacionamentos
	Curso *Curso `json:"curso,omitempty" gorm:"foreignKey:CursoID"`
}

func (Inscricao) TableName() string {
	return "inscricoes"
}

type EnrolledUnit struct {
	ID             string    `json:"id"`
	CursoID        int       `json:"curso_id"`
	Address        string    `json:"address"`
	Neighborhood   string    `json:"neighborhood"`
	Vacancies      int       `json:"vacancies"`
	ClassStartDate string    `json:"class_start_date"`
	ClassEndDate   string    `json:"class_end_date"`
	ClassTime      string    `json:"class_time"`
	ClassDays      string    `json:"class_days"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

// Value implements the driver.Valuer interface for EnrolledUnit
func (e EnrolledUnit) Value() (driver.Value, error) {
	if e.ID == "" {
		return nil, nil
	}
	return json.Marshal(e)
}

// Scan implements the sql.Scanner interface for EnrolledUnit
func (e *EnrolledUnit) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan EnrolledUnit: value is not []byte")
	}

	if len(bytes) == 0 {
		return nil
	}

	return json.Unmarshal(bytes, e)
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
	Concluded int `json:"concluded"`
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

type InscricaoUpdateRequest struct {
	Name             *string         `json:"name"`
	Email            *string         `json:"email"`
	Phone            *string         `json:"phone"`
	Age              *int            `json:"age"`
	Address          *string         `json:"address"`
	Neighborhood     *string         `json:"neighborhood"`
	CustomFieldsData datatypes.JSON  `json:"custom_fields" swaggertype:"object"`
	AdminNotes       *string         `json:"admin_notes"`
	EnrolledUnit     *EnrolledUnit   `json:"enrolled_unit"`
}