package empregabilidade

import (
	"time"

	"github.com/google/uuid"
)

type Etapa struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	IDVaga    uuid.UUID `json:"id_vaga" gorm:"type:uuid;not null"`
	Titulo    string    `json:"titulo" gorm:"type:varchar(500);not null"`
	Descricao string    `json:"descricao" gorm:"type:text"`
	Ordem     int       `json:"ordem" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Vaga *Vaga `json:"vaga,omitempty" gorm:"foreignKey:IDVaga"`
}

func (Etapa) TableName() string {
	return "emp_etapas"
}
