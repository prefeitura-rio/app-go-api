package empregabilidade

import (
	"time"

	"github.com/google/uuid"
)

type RegimeContratacao struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Descricao string    `json:"descricao" gorm:"type:varchar(255);not null;unique"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (RegimeContratacao) TableName() string {
	return "emp_regimes_contratacao"
}
