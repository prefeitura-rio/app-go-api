package empregabilidade

import (
	"time"
)

// AreaAtuacaoHabilidade representa a tabela pivô 'area_atuacao_habilidade'
type AreaAtuacaoHabilidade struct {
	ID            int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	IDHabilidade  int64 `json:"id_habilidade" gorm:"column:id_habilidade;not null;uniqueIndex:uk_habilidade_area"`
	IDAreaAtuacao int64 `json:"id_area_atuacao" gorm:"column:id_area_atuacao;not null;uniqueIndex:uk_habilidade_area"`

	Habilidade  *Habilidade  `json:"habilidade,omitempty" gorm:"foreignKey:IDHabilidade;references:ID"`
	AreaAtuacao *AreaAtuacao `json:"area_atuacao,omitempty" gorm:"foreignKey:IDAreaAtuacao;references:ID"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (AreaAtuacaoHabilidade) TableName() string {
	return "area_atuacao_habilidade"
}

// Habilidade representa a tabela 'emp_habilidades'
type Habilidade struct {
	ID        int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	Nome      string        `json:"nome" gorm:"type:varchar(250);unique;not null"`
	CreatedAt time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	Areas     []AreaAtuacao `json:"areas,omitempty" gorm:"many2many:area_atuacao_habilidade;foreignKey:ID;joinForeignKey:id_habilidade;references:ID;joinReferences:id_area_atuacao"`
}

func (Habilidade) TableName() string {
	return "emp_habilidades"
}

// AreaAtuacao representa a tabela 'emp_areas_atuacao'
type AreaAtuacao struct {
	ID          int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	Nome        string       `json:"nome" gorm:"type:varchar(250);unique;not null"`
	CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	Habilidades []Habilidade `json:"habilidades,omitempty" gorm:"many2many:area_atuacao_habilidade;foreignKey:ID;joinForeignKey:id_area_atuacao;references:ID;joinReferences:id_habilidade"`
}

func (AreaAtuacao) TableName() string {
	return "emp_areas_atuacao"
}

// HabilidadeFilter estrutura para aplicar filtros na consulta
type HabilidadeFilter struct {
	Search        string `json:"search"`
	CPF           string `json:"cpf"`
	AreaAtuacaoID int64  `json:"id_area_atuacao"`
	Page          int
	Limit         int
}

// AreaAtuacaoFilter estrutura para aplicar filtros na consulta
type AreaAtuacaoFilter struct {
	Search       string `json:"search"`
	HabilidadeID int64  `json:"id_habilidade"`
	Page         int
	Limit        int
}
