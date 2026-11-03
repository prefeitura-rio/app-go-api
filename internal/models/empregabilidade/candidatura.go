package empregabilidade

import (
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"gorm.io/gorm"
)

type StatusCandidatura string

const (
	StatusCandidaturaEnviada       StatusCandidatura = "candidatura_enviada"
	StatusCandidaturaAprovada      StatusCandidatura = "aprovada"
	StatusCandidaturaReprovada     StatusCandidatura = "reprovada"
	StatusCandidaturaVagaCongelada StatusCandidatura = "vaga_congelada"
	StatusCandidaturaDescontinuada StatusCandidatura = "vaga_descontinuada"
)

func (s StatusCandidatura) IsValid() bool {
	switch s {
	case StatusCandidaturaEnviada, StatusCandidaturaAprovada, StatusCandidaturaReprovada, StatusCandidaturaVagaCongelada, StatusCandidaturaDescontinuada:
		return true
	}
	return false
}

// OrigemCandidatura diz por onde a candidatura entrou.
type OrigemCandidatura string

const (
	// OrigemCandidaturaPortalCidadao é a candidatura enviada pelo próprio cidadão no app.
	OrigemCandidaturaPortalCidadao OrigemCandidatura = "portal_cidadao"
	// OrigemCandidaturaSecretaria é o candidato cadastrado manualmente na vaga pelo backoffice.
	OrigemCandidaturaSecretaria OrigemCandidatura = "secretaria"
	// OrigemCandidaturaBancoCurriculos é a associação feita a partir da ficha do banco de currículos.
	OrigemCandidaturaBancoCurriculos OrigemCandidatura = "banco_curriculos"
)

type RespostaInfoComplementar struct {
	IDInfo   uuid.UUID `json:"id_info"`
	Titulo   string    `json:"titulo"`
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
	StatusAnterior              *StatusCandidatura         `json:"status_anterior,omitempty" gorm:"type:varchar(100)"`
	RespostasInfoComplementares []RespostaInfoComplementar `json:"respostas_info_complementares" gorm:"type:jsonb;serializer:json"`
	CurriculoSnapshot           *CurriculoCompleto         `json:"curriculo_snapshot,omitempty" gorm:"type:jsonb;serializer:json"`
	// Origem e AssociadoPorCPF são definidos pela API, nunca pelo corpo da requisição.
	Origem          *OrigemCandidatura `json:"origem,omitempty" gorm:"type:varchar(64)"`
	AssociadoPorCPF *string            `json:"associado_por_cpf,omitempty" gorm:"type:varchar(14);column:associado_por_cpf"`
	DeletedAt       gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string" example:"2024-12-31T23:59:59Z"`
	CreatedAt       time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time          `json:"updated_at" gorm:"autoUpdateTime"`

	// Computed field (not stored in DB) - populated from CitizenSnapshot
	PersonalInfo *models.CitizenPersonalInfo `json:"personal_info,omitempty" gorm:"-"`

	// Relationships
	Vaga       *Vaga  `json:"vaga,omitempty" gorm:"foreignKey:IDVaga"`
	EtapaAtual *Etapa `json:"etapa_atual,omitempty" gorm:"foreignKey:IDEtapaAtual"`
}

func (Candidatura) TableName() string {
	return "emp_candidaturas"
}

type CandidaturaFilter struct {
	CPF     string
	VagaID  *uuid.UUID
	Status  string
	Search  string     // ILIKE em cpf, nome, email
	EtapaID *uuid.UUID // filtro por etapa atual
	// Candidato são os filtros avançados de candidatos, os mesmos do banco de currículos.
	Candidato FiltroCandidato
}
