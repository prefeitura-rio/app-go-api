package models

import (
	"database/sql/driver"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type StatusOportunidadeMEI string
type FormaPagamento string

// StringArray is a custom type for PostgreSQL string arrays
type StringArray []string

// Scan implements the sql.Scanner interface for StringArray
func (a *StringArray) Scan(value interface{}) error {
	return pq.Array(a).Scan(value)
}

// Value implements the driver.Valuer interface for StringArray
func (a StringArray) Value() (driver.Value, error) {
	return pq.Array(a).Value()
}

const (
	StatusOportunidadeDraft   StatusOportunidadeMEI = "draft"
	StatusOportunidadeActive  StatusOportunidadeMEI = "active"
	StatusOportunidadeExpired StatusOportunidadeMEI = "expired"

	FormaPagamentoCheque        FormaPagamento = "CHEQUE"
	FormaPagamentoDinheiro      FormaPagamento = "DINHEIRO"
	FormaPagamentoCartao        FormaPagamento = "CARTAO"
	FormaPagamentoPix           FormaPagamento = "PIX"
	FormaPagamentoTransferencia FormaPagamento = "TRANSFERENCIA"
)

// OportunidadeMEI representa uma oportunidade de negócio para MEIs
type OportunidadeMEI struct {
	ID                int    `json:"id" gorm:"primaryKey" example:"1"`
	Titulo            string `json:"titulo" gorm:"type:varchar(255);not null" example:"Manutenção de Jardins da Praça XV"`
	DescricaoServico  string `json:"descricao_servico" gorm:"type:text;not null" example:"Serviço de poda, jardinagem e manutenção das áreas verdes"`
	OutrasInformacoes string `json:"outras_informacoes,omitempty" gorm:"type:text" example:"Trabalho a ser realizado nos finais de semana"`

	// Relacionamentos
	OrgaoID string      `json:"orgao_id" gorm:"type:varchar(100);not null" example:"SECONSERVA"`
	CNAEIDs StringArray `json:"cnae_ids" gorm:"type:varchar(20)[];not null;default:'{}'" swaggertype:"array,string" example:"8130300,8121400"`

	// Local de execução
	Logradouro  string `json:"logradouro" gorm:"type:varchar(255);not null" example:"Praça XV de Novembro"`
	Numero      string `json:"numero" gorm:"type:varchar(20);not null" example:"S/N"`
	Complemento string `json:"complemento,omitempty" gorm:"type:varchar(100)" example:"Próximo ao Paço Imperial"`
	Bairro      string `json:"bairro" gorm:"type:varchar(100);not null" example:"Centro"`
	Cidade      string `json:"cidade" gorm:"type:varchar(100);not null" example:"Rio de Janeiro"`
	Estado      string `json:"estado" gorm:"type:varchar(2);not null" example:"RJ"`

	// Pagamento e prazos
	FormaPagamento     *string    `json:"forma_pagamento,omitempty" gorm:"type:varchar(50)" swaggertype:"string" enums:"CHEQUE,DINHEIRO,CARTAO,PIX,TRANSFERENCIA," example:"PIX"`
	PrazoPagamento     string     `json:"prazo_pagamento,omitempty" gorm:"type:varchar(100)" example:"30 dias após conclusão"`
	DataLimiteExecucao *time.Time `json:"data_limite_execucao,omitempty" gorm:"type:timestamp with time zone" example:"2024-12-31T23:59:59Z"`
	DataExpiracao      *time.Time `json:"data_expiracao" gorm:"type:timestamp with time zone;not null" example:"2024-12-31T23:59:59Z"`

	// Imagens
	CoverImage    string   `json:"cover_image,omitempty" gorm:"type:varchar(500)" example:"https://example.com/imagem.jpg"`
	GalleryImages []string `json:"gallery_images,omitempty" gorm:"type:jsonb;serializer:json" swaggertype:"array,string" example:"https://example.com/img1.jpg,https://example.com/img2.jpg"`

	// Status e controle
	Status    StatusOportunidadeMEI `json:"status" gorm:"type:varchar(50);not null;default:'draft'" enums:"draft,active,expired" example:"active"`
	DeletedAt gorm.DeletedAt        `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string" example:"2024-12-31T23:59:59Z"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" example:"2024-01-01T10:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime" example:"2024-01-01T10:00:00Z"`

	// Relacionamentos expandidos
	Propostas []PropostaMEI `json:"propostas,omitempty" gorm:"foreignKey:OportunidadeMEIID" swaggerignore:"true"`
}

func (OportunidadeMEI) TableName() string {
	return "oportunidades_mei"
}

func (s StatusOportunidadeMEI) IsValid() bool {
	validValues := []string{"draft", "active", "expired"}
	for _, v := range validValues {
		if string(s) == v {
			return true
		}
	}
	return false
}

func (f FormaPagamento) IsValid() bool {
	// String vazia é válida (campo opcional)
	if string(f) == "" {
		return true
	}

	validValues := []string{"CHEQUE", "DINHEIRO", "CARTAO", "PIX", "TRANSFERENCIA"}
	for _, v := range validValues {
		if string(f) == v {
			return true
		}
	}
	return false
}

// Validate realiza validação básica, permitindo campos vazios para rascunhos
func (o *OportunidadeMEI) Validate() error {
	// Validação de status sempre é obrigatória
	if !o.Status.IsValid() {
		return errors.New("status inválido")
	}

	// Se for rascunho, permite campos vazios (validação mínima)
	if o.Status == StatusOportunidadeDraft {
		// Apenas valida FormaPagamento se tiver valor
		if o.FormaPagamento != nil && *o.FormaPagamento != "" {
			if !FormaPagamento(*o.FormaPagamento).IsValid() {
				return errors.New("forma de pagamento inválida")
			}
		}
		// Para rascunhos, não exige campos obrigatórios
		return nil
	}

	// Para status diferente de draft, faz validação completa
	return o.ValidateForPublish()
}

// ValidateForPublish realiza validação completa necessária para publicação
func (o *OportunidadeMEI) ValidateForPublish() error {
	if strings.TrimSpace(o.Titulo) == "" {
		return errors.New("título é obrigatório")
	}

	if strings.TrimSpace(o.DescricaoServico) == "" {
		return errors.New("descrição do serviço é obrigatória")
	}

	if strings.TrimSpace(o.OrgaoID) == "" {
		return errors.New("órgão demandante é obrigatório")
	}

	if len(o.CNAEIDs) == 0 {
		return errors.New("pelo menos um CNAE é obrigatório")
	}

	// Validate that all CNAEs are non-empty
	for i, cnae := range o.CNAEIDs {
		if strings.TrimSpace(cnae) == "" {
			return errors.New("CNAE na posição " + string(rune(i)) + " está vazio")
		}
	}

	if strings.TrimSpace(o.Logradouro) == "" {
		return errors.New("logradouro é obrigatório")
	}

	if strings.TrimSpace(o.Numero) == "" {
		return errors.New("número é obrigatório")
	}

	if strings.TrimSpace(o.Bairro) == "" {
		return errors.New("bairro é obrigatório")
	}

	if strings.TrimSpace(o.Cidade) == "" {
		return errors.New("cidade é obrigatória")
	}

	if strings.TrimSpace(o.Estado) == "" {
		return errors.New("estado é obrigatório")
	}

	// FormaPagamento é opcional, só valida se tiver valor
	if o.FormaPagamento != nil && *o.FormaPagamento != "" {
		if !FormaPagamento(*o.FormaPagamento).IsValid() {
			return errors.New("forma de pagamento inválida")
		}
	}

	if o.DataExpiracao == nil {
		return errors.New("data de expiração é obrigatória")
	}

	return nil
}

// UpdateStatusBasedOnExpiration atualiza o status baseado na data de expiração
func (o *OportunidadeMEI) UpdateStatusBasedOnExpiration() {
	if o.Status == StatusOportunidadeActive && o.DataExpiracao != nil {
		if time.Now().After(*o.DataExpiracao) {
			o.Status = StatusOportunidadeExpired
		}
	}
}
