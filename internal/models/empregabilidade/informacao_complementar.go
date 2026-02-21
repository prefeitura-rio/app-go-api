package empregabilidade

import (
	"time"

	"github.com/google/uuid"
)

type TipoCampo string

const (
	TipoCampoRespostaCurta    TipoCampo = "resposta_curta"
	TipoCampoRespostaNumerica TipoCampo = "resposta_numerica"
	TipoCampoSelecaoUnica     TipoCampo = "selecao_unica"
	TipoCampoSelecaoMultipla  TipoCampo = "selecao_multipla"
)

func (t TipoCampo) IsValid() bool {
	switch t {
	case TipoCampoRespostaCurta, TipoCampoRespostaNumerica, TipoCampoSelecaoUnica, TipoCampoSelecaoMultipla:
		return true
	}
	return false
}

type InformacaoComplementar struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	IDVaga      uuid.UUID `json:"id_vaga" gorm:"type:uuid;not null"`
	Titulo      string    `json:"titulo" gorm:"type:varchar(500);not null"`
	Obrigatorio bool      `json:"obrigatorio" gorm:"default:false"`
	TipoCampo   TipoCampo `json:"tipo_campo" gorm:"type:varchar(50);not null"`
	ValorMinimo *int      `json:"valor_minimo" gorm:"type:integer"`
	ValorMaximo *int      `json:"valor_maximo" gorm:"type:integer"`
	Opcoes      []string  `json:"opcoes" gorm:"type:jsonb;serializer:json"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Vaga *Vaga `json:"vaga,omitempty" gorm:"foreignKey:IDVaga"`
}

func (InformacaoComplementar) TableName() string {
	return "emp_informacoes_complementares"
}
