package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatusPropostaAdmin string
type StatusPropostaCidadao string

const (
	// Status para controle do admin
	StatusPropostaAdminDraft   StatusPropostaAdmin = "draft"
	StatusPropostaAdminActive  StatusPropostaAdmin = "active"
	StatusPropostaAdminExpired StatusPropostaAdmin = "expired"

	// Status para visualização do cidadão
	StatusPropostaCidadaoSubmitted StatusPropostaCidadao = "submitted"
	StatusPropostaCidadaoApproved  StatusPropostaCidadao = "approved"
	StatusPropostaCidadaoRejected  StatusPropostaCidadao = "rejected"
)

// PropostaMEI representa uma proposta de um MEI para uma oportunidade
type PropostaMEI struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" example:"550e8400-e29b-41d4-a716-446655440000"`
	OportunidadeMEIID     int       `json:"oportunidade_mei_id" gorm:"not null;index" example:"1"`
	MEIEmpresaID          string    `json:"mei_empresa_id" gorm:"type:varchar(18);not null;index" example:"12.345.678/0001-90"`
	ValorProposta         *float64  `json:"valor_proposta,omitempty" gorm:"type:decimal(10,2)" example:"1500.50"`
	PrazoExecucao         *string   `json:"prazo_execucao,omitempty" gorm:"type:varchar(255)" example:"30 dias"`
	AceitaCustosIntegrais *bool     `json:"aceita_custos_integrais,omitempty" gorm:"type:boolean" example:"true"`

	StatusAdmin   StatusPropostaAdmin   `json:"status_admin" gorm:"type:varchar(50);not null;default:'active'" enums:"draft,active,expired" example:"active"`
	StatusCidadao StatusPropostaCidadao `json:"status_cidadao" gorm:"type:varchar(50);not null;default:'submitted'" enums:"submitted,approved,rejected" example:"submitted"`

	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string" example:"2024-12-31T23:59:59Z"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime" example:"2024-01-01T10:00:00Z"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime" example:"2024-01-01T10:00:00Z"`

	// Computed fields from RMI API (not persisted to database)
	EmailPessoaFisica   string `json:"email_pessoa_fisica,omitempty" gorm:"-" example:"joao@example.com"`
	CelularPessoaFisica string `json:"celular_pessoa_fisica,omitempty" gorm:"-" example:"21987654321"`

	// Relacionamentos
	Oportunidade *OportunidadeMEI `json:"oportunidade,omitempty" gorm:"foreignKey:OportunidadeMEIID" swaggerignore:"true"`
}

func (PropostaMEI) TableName() string {
	return "propostas_mei"
}

func (s StatusPropostaAdmin) IsValid() bool {
	validValues := []string{"draft", "active", "expired"}
	for _, v := range validValues {
		if string(s) == v {
			return true
		}
	}
	return false
}

func (s StatusPropostaCidadao) IsValid() bool {
	validValues := []string{"submitted", "approved", "rejected"}
	for _, v := range validValues {
		if string(s) == v {
			return true
		}
	}
	return false
}

// Validate validates the PropostaMEI fields
func (p *PropostaMEI) Validate() error {
	if p.OportunidadeMEIID == 0 {
		return errors.New("oportunidade_mei_id é obrigatório")
	}

	if p.MEIEmpresaID == "" {
		return errors.New("mei_empresa_id (CNPJ) é obrigatório")
	}

	// Validate ValorProposta if provided (must be positive)
	if p.ValorProposta != nil && *p.ValorProposta < 0 {
		return errors.New("valor_proposta deve ser positivo")
	}

	return nil
}

// BeforeCreate hook para gerar UUID se não existir
func (p *PropostaMEI) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PropostaStatusUpdateRequest representa requisição para atualizar status de múltiplas propostas
type PropostaStatusUpdateRequest struct {
	PropostaIDs []uuid.UUID           `json:"proposta_ids" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000,660e8400-e29b-41d4-a716-446655440001"`
	Status      StatusPropostaCidadao `json:"status" binding:"required" enums:"submitted,approved,rejected" example:"approved"`
}

// PropostaStatusUpdateResponse representa resposta de atualização de status de propostas
type PropostaStatusUpdateResponse struct {
	UpdatedCount int                   `json:"updated_count" example:"2"`
	Status       StatusPropostaCidadao `json:"status" enums:"submitted,approved,rejected" example:"approved"`
	UpdatedAt    time.Time             `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}
