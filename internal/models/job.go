package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type JobStatus string
type JobType string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"

	JobTypeEnrollmentImport JobType = "enrollment_import"
)

// Job representa um trabalho assíncrono (ex: importação de inscrições)
type Job struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`                     // ID único do job
	Type         JobType        `json:"type" gorm:"type:varchar(50);not null" swaggertype:"string" example:"enrollment_import"`             // Tipo de job (enrollment_import, etc)
	Status       JobStatus      `json:"status" gorm:"type:varchar(50);default:'pending';not null" swaggertype:"string" example:"completed"` // Status: pending, processing, completed, failed
	Progress     int            `json:"progress" gorm:"default:0" example:"100"`                                       // Progresso em porcentagem (0-100)
	TotalRecords int            `json:"total_records" gorm:"column:total_records;default:0" example:"10"`              // Total de registros a processar
	SuccessCount int            `json:"success_count" gorm:"column:success_count;default:0" example:"8"`               // Quantidade de registros processados com sucesso
	ErrorCount   int            `json:"error_count" gorm:"column:error_count;default:0" example:"2"`                   // Quantidade de registros com erro
	Errors       datatypes.JSON `json:"errors,omitempty" gorm:"type:jsonb" swaggertype:"array,object"`                 // Lista de erros detalhados por linha
	Result       datatypes.JSON `json:"result,omitempty" gorm:"type:jsonb" swaggertype:"object"`                       // Resultado final do processamento
	Metadata     datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb" swaggertype:"object"`                     // Metadata adicional (curso_id, file_name, etc)
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`                                              // Data/hora de criação
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`                                              // Data/hora da última atualização
	CompletedAt  *time.Time     `json:"completed_at,omitempty" gorm:"column:completed_at"`                             // Data/hora de conclusão
}

func (Job) TableName() string {
	return "jobs"
}

type JobError struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

type EnrollmentImportMetadata struct {
	CursoID  int    `json:"curso_id"`
	FileName string `json:"file_name"`
}

type EnrollmentImportResult struct {
	SuccessCount   int         `json:"success_count"`
	ErrorCount     int         `json:"error_count"`
	DuplicateCount int         `json:"duplicate_count"`
	EnrollmentIDs  []uuid.UUID `json:"enrollment_ids,omitempty"`
}
