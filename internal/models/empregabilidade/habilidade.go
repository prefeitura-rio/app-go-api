package empregabilidade

import (
	"time"

	"github.com/google/uuid"
)

// Habilidade representa a tabela 'emp_habilidades'
type Habilidade struct {
	ID        uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Nome      string        `json:"nome" gorm:"type:varchar(500);unique;not null"`
	CreatedAt time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	Areas     []AreaAtuacao `json:"areas,omitempty" gorm:"many2many:area_atuacao_habilidade;foreignKey:ID;joinForeignKey:id_habilidade;references:ID;joinReferences:id_area_atuacao"`
}

func (Habilidade) TableName() string {
	return "emp_habilidades"
}

// CurriculoHabilidade representa o vínculo entre o candidato (CPF) e suas Habilidades
type CurriculoHabilidade struct {
	ID           uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CPF          string      `json:"cpf" gorm:"type:char(11);not null;uniqueIndex:uk_emp_curriculo_habilidades_cpf_habilidade"`
	IDHabilidade uuid.UUID   `json:"id_habilidade" gorm:"type:uuid;not null;uniqueIndex:uk_emp_curriculo_habilidades_cpf_habilidade"`
	CreatedAt    time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	Habilidade   *Habilidade `json:"habilidade,omitempty" gorm:"foreignKey:IDHabilidade;references:ID"`
}

func (CurriculoHabilidade) TableName() string {
	return "emp_curriculo_habilidades"
}

// AreaAtuacao representa a tabela 'area_atuacao'
type AreaAtuacao struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Nome        string       `json:"nome" gorm:"type:varchar(500);unique;not null"`
	CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
	Habilidades []Habilidade `json:"habilidades,omitempty" gorm:"many2many:area_atuacao_habilidade;foreignKey:ID;joinForeignKey:id_area_atuacao;references:ID;joinReferences:id_habilidade"`
}

func (AreaAtuacao) TableName() string {
	return "area_atuacao"
}

// HabilidadeFilter estrutura para aplicar filtros na consulta
type HabilidadeFilter struct {
	Search        string    // Busca por similaridade usando trigram/unaccent ou ILIKE
	CPF           string    // Exact match
	AreaAtuacaoID uuid.UUID // Permite filtrar habilidades de uma área específica
	Page          int
	Limit         int
}

// AreaAtuacaoFilter estrutura para aplicar filtros na consulta
type AreaAtuacaoFilter struct {
	Search string // Filtro de texto por nome
	Page   int
	Limit  int
}
