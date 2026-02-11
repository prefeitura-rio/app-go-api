package empregabilidade

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatusCandidatura string

const (
	StatusCandidaturaEnviada        StatusCandidatura = "candidatura_enviada"
	StatusCandidaturaAprovada       StatusCandidatura = "aprovada"
	StatusCandidaturaReprovada      StatusCandidatura = "reprovada"
	StatusCandidaturaVagaCongelada  StatusCandidatura = "vaga_congelada"
	StatusCandidaturaDescontinuada  StatusCandidatura = "vaga_descontinuada"
)

func (s StatusCandidatura) IsValid() bool {
	switch s {
	case StatusCandidaturaEnviada, StatusCandidaturaAprovada, StatusCandidaturaReprovada, StatusCandidaturaVagaCongelada, StatusCandidaturaDescontinuada:
		return true
	}
	return false
}

type RespostaInfoComplementar struct {
	IDInfo   uuid.UUID `json:"id_info"`
	Resposta string    `json:"resposta"`
}

type Candidatura struct {
	ID                          uuid.UUID                  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CPF                         string                     `json:"cpf" gorm:"type:varchar(14);not null"`
	Nome                        *string                    `json:"nome,omitempty" gorm:"type:varchar(255)"`
	Email                       *string                    `json:"email,omitempty" gorm:"type:varchar(255)"`
	IDVaga                      uuid.UUID                  `json:"id_vaga" gorm:"type:uuid;not null"`
	Status                      StatusCandidatura          `json:"status" gorm:"type:varchar(100);not null;default:'candidatura_enviada'"`
	IDEtapaAtual                *uuid.UUID                 `json:"id_etapa_atual" gorm:"type:uuid"`
	RespostasInfoComplementares []RespostaInfoComplementar `json:"respostas_info_complementares" gorm:"type:jsonb;serializer:json"`
	CurriculoSnapshot           *CurriculoCompleto         `json:"curriculo_snapshot,omitempty" gorm:"type:jsonb;serializer:json"`
	DeletedAt                   gorm.DeletedAt             `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string" example:"2024-12-31T23:59:59Z"`
	CreatedAt                   time.Time                  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                   time.Time                  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Vaga       *Vaga  `json:"vaga,omitempty" gorm:"foreignKey:IDVaga"`
	EtapaAtual *Etapa `json:"etapa_atual,omitempty" gorm:"foreignKey:IDEtapaAtual"`
}

func (Candidatura) TableName() string {
	return "emp_candidaturas"
}
