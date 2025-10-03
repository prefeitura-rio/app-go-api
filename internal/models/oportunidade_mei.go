package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type StatusOportunidadeMEI string
type FormaPagamento string

const (
	StatusOportunidadeDraft   StatusOportunidadeMEI = "draft"
	StatusOportunidadeActive  StatusOportunidadeMEI = "active"
	StatusOportunidadeExpired StatusOportunidadeMEI = "expired"

	FormaPagamentoCheque  FormaPagamento = "CHEQUE"
	FormaPagamentoDinheiro FormaPagamento = "DINHEIRO"
	FormaPagamentoCartao  FormaPagamento = "CARTAO"
	FormaPagamentoPix     FormaPagamento = "PIX"
	FormaPagamentoTransferencia FormaPagamento = "TRANSFERENCIA"
)

type OportunidadeMEI struct {
	ID               int                    `json:"id" gorm:"primaryKey"`
	Titulo           string                 `json:"titulo" gorm:"type:varchar(255);not null"`
	DescricaoServico string                 `json:"descricao_servico" gorm:"type:text;not null"`
	OutrasInformacoes string                `json:"outras_informacoes" gorm:"type:text"`

	// Relacionamentos
	OrgaoID int `json:"orgao_id" gorm:"not null"`
	CNAEID  int `json:"cnae_id" gorm:"not null"`

	// Local de execução
	Logradouro  string `json:"logradouro" gorm:"type:varchar(255);not null"`
	Numero      string `json:"numero" gorm:"type:varchar(20);not null"`
	Complemento string `json:"complemento" gorm:"type:varchar(100)"`
	Bairro      string `json:"bairro" gorm:"type:varchar(100);not null"`
	Cidade      string `json:"cidade" gorm:"type:varchar(100);not null"`
	Estado      string `json:"estado" gorm:"type:varchar(2);not null"`

	// Pagamento e prazos
	FormaPagamento      FormaPagamento `json:"forma_pagamento" gorm:"type:varchar(50);not null"`
	PrazoPagamento      string         `json:"prazo_pagamento" gorm:"type:varchar(100)"`
	DataLimiteExecucao  *time.Time     `json:"data_limite_execucao" gorm:"type:timestamp with time zone"`
	DataExpiracao       *time.Time     `json:"data_expiracao" gorm:"type:timestamp with time zone;not null"`

	// Anexo
	AnexoURL string `json:"anexo_url" gorm:"type:varchar(500)"`

	// Status e controle
	Status    StatusOportunidadeMEI `json:"status" gorm:"type:varchar(50);not null;default:'draft'"`
	DeletedAt gorm.DeletedAt        `json:"deleted_at,omitempty" gorm:"index"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relacionamentos expandidos
	Orgao     *Orgao        `json:"orgao,omitempty" gorm:"foreignKey:OrgaoID"`
	CNAE      *CNAE         `json:"cnae,omitempty" gorm:"foreignKey:CNAEID"`
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
	validValues := []string{"CHEQUE", "DINHEIRO", "CARTAO", "PIX", "TRANSFERENCIA"}
	for _, v := range validValues {
		if string(f) == v {
			return true
		}
	}
	return false
}

func (o *OportunidadeMEI) Validate() error {
	if strings.TrimSpace(o.Titulo) == "" {
		return errors.New("título é obrigatório")
	}

	if strings.TrimSpace(o.DescricaoServico) == "" {
		return errors.New("descrição do serviço é obrigatória")
	}

	if o.OrgaoID == 0 {
		return errors.New("órgão demandante é obrigatório")
	}

	if o.CNAEID == 0 {
		return errors.New("CNAE é obrigatório")
	}

	if strings.TrimSpace(o.Logradouro) == "" {
		return errors.New("logradouro é obrigatório")
	}

	if strings.TrimSpace(o.Cidade) == "" {
		return errors.New("cidade é obrigatória")
	}

	if !o.Status.IsValid() {
		return errors.New("status inválido")
	}

	if !o.FormaPagamento.IsValid() {
		return errors.New("forma de pagamento inválida")
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
