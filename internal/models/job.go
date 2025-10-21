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

type Job struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Type         JobType        `json:"type" gorm:"type:varchar(50);not null"`
	Status       JobStatus      `json:"status" gorm:"type:varchar(50);default:'pending';not null"`
	Progress     int            `json:"progress" gorm:"default:0"`
	TotalRecords int            `json:"total_records" gorm:"column:total_records;default:0"`
	SuccessCount int            `json:"success_count" gorm:"column:success_count;default:0"`
	ErrorCount   int            `json:"error_count" gorm:"column:error_count;default:0"`
	Errors       datatypes.JSON `json:"errors,omitempty" gorm:"type:jsonb" swaggertype:"array,object"`
	Result       datatypes.JSON `json:"result,omitempty" gorm:"type:jsonb" swaggertype:"object"`
	Metadata     datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb" swaggertype:"object"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty" gorm:"column:completed_at"`
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
	SuccessCount   int   `json:"success_count"`
	ErrorCount     int   `json:"error_count"`
	DuplicateCount int   `json:"duplicate_count"`
	EnrollmentIDs  []uuid.UUID `json:"enrollment_ids,omitempty"`
}
