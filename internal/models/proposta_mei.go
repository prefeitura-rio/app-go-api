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

type PropostaMEI struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	OportunidadeMEIID int       `json:"oportunidade_mei_id" gorm:"not null;index"`
	MEIEmpresaID      string    `json:"mei_empresa_id" gorm:"type:varchar(18);not null;index"`
	ValorProposta     *float64  `json:"valor_proposta,omitempty" gorm:"type:decimal(10,2)"`

	StatusAdmin   StatusPropostaAdmin   `json:"status_admin" gorm:"type:varchar(50);not null;default:'active'"`
	StatusCidadao StatusPropostaCidadao `json:"status_cidadao" gorm:"type:varchar(50);not null;default:'submitted'"`

	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	// Relacionamentos
	Oportunidade *OportunidadeMEI `json:"oportunidade,omitempty" gorm:"foreignKey:OportunidadeMEIID"`
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

type PropostaStatusUpdateRequest struct {
	PropostaIDs []uuid.UUID           `json:"proposta_ids" binding:"required"`
	Status      StatusPropostaCidadao `json:"status" binding:"required"`
}

type PropostaStatusUpdateResponse struct {
	UpdatedCount int                   `json:"updated_count"`
	Status       StatusPropostaCidadao `json:"status"`
	UpdatedAt    time.Time             `json:"updated_at"`
}
