package empregabilidade

import (
	"time"

	"github.com/google/uuid"
)

type Idioma struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Descricao string    `json:"descricao" gorm:"type:varchar(255);not null;unique"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Idioma) TableName() string {
	return "emp_idiomas"
}
