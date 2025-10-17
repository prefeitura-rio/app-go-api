package models

import (
	"errors"
	"strings"
	"time"
)

type SituacaoCadastral string

const (
	SituacaoAtiva      SituacaoCadastral = "ATIVA"
	SituacaoSuspensa   SituacaoCadastral = "SUSPENSA"
	SituacaoInapta     SituacaoCadastral = "INAPTA"
	SituacaoBaixada    SituacaoCadastral = "BAIXADA"
	SituacaoNula       SituacaoCadastral = "NULA"
)

type MEIEmpresa struct {
	ID                int                `json:"id" gorm:"primaryKey"`
	CNPJ              string             `json:"cnpj" gorm:"type:varchar(18);uniqueIndex;not null"`
	RazaoSocial       string             `json:"razao_social" gorm:"type:varchar(255);not null"`
	Porte             string             `json:"porte" gorm:"type:varchar(100)"`
	NomeFantasia      string             `json:"nome_fantasia" gorm:"type:varchar(255)"`
	Tipo              string             `json:"tipo" gorm:"type:varchar(100)"`
	NaturezaJuridica  string             `json:"natureza_juridica" gorm:"type:varchar(100)"`
	SituacaoCadastral SituacaoCadastral  `json:"situacao_cadastral" gorm:"type:varchar(50)"`

	// Endereço
	CEP        string `json:"cep" gorm:"type:varchar(10)"`
	Logradouro string `json:"logradouro" gorm:"type:varchar(255)"`
	Numero     string `json:"numero" gorm:"type:varchar(20)"`
	Bairro     string `json:"bairro" gorm:"type:varchar(100)"`
	Cidade     string `json:"cidade" gorm:"type:varchar(100)"`
	Estado     string `json:"estado" gorm:"type:varchar(2)"`

	// Dados de contato
	Email    string `json:"email" gorm:"type:varchar(255)"`
	Whatsapp string `json:"whatsapp" gorm:"type:varchar(20)"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relacionamentos
	CNAEs     []CNAE         `json:"servicos,omitempty" gorm:"many2many:mei_empresas_cnaes;"`
	Propostas []PropostaMEI  `json:"propostas,omitempty" gorm:"foreignKey:MEIEmpresaID" swaggerignore:"true"`
}

func (MEIEmpresa) TableName() string {
	return "mei_empresas"
}

func (s SituacaoCadastral) IsValid() bool {
	validValues := []string{"ATIVA", "SUSPENSA", "INAPTA", "BAIXADA", "NULA"}
	for _, v := range validValues {
		if string(s) == v {
			return true
		}
	}
	return false
}

func (m *MEIEmpresa) Validate() error {
	if strings.TrimSpace(m.CNPJ) == "" {
		return errors.New("CNPJ é obrigatório")
	}

	if strings.TrimSpace(m.RazaoSocial) == "" {
		return errors.New("razão social é obrigatória")
	}

	return nil
}
