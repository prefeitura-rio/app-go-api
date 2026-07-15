package empregabilidade

import (
	"time"

	"github.com/google/uuid"
)

// VagaIdiomaRequisito representa um requisito de idioma + nível mínimo exigido
// por uma vaga. É uma entidade própria (não um many2many simples) porque carrega
// a coluna adicional id_nivel_minimo, análoga a CurriculoIdioma no lado do currículo.
type VagaIdiomaRequisito struct {
	IDVaga        uuid.UUID `json:"id_vaga" gorm:"type:uuid;primaryKey"`
	IDIdioma      uuid.UUID `json:"id_idioma" gorm:"type:uuid;primaryKey"`
	IDNivelMinimo uuid.UUID `json:"id_nivel_minimo" gorm:"type:uuid;not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Vaga        *Vaga        `json:"vaga,omitempty" gorm:"foreignKey:IDVaga"`
	Idioma      *Idioma      `json:"idioma,omitempty" gorm:"foreignKey:IDIdioma"`
	NivelMinimo *NivelIdioma `json:"nivel_minimo,omitempty" gorm:"foreignKey:IDNivelMinimo"`
}

func (VagaIdiomaRequisito) TableName() string {
	return "emp_vagas_idiomas_requisitos"
}
