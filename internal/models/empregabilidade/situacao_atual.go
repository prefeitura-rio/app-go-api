package empregabilidade

import (
	"time"

	"github.com/google/uuid"
)

type SituacaoAtual struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Descricao string    `json:"descricao" gorm:"type:varchar(255);not null;unique"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SituacaoAtual) TableName() string {
	return "emp_situacoes_atual"
}

// SituacoesProcurandoEmprego são as situações atuais (seed de
// emp_situacoes_atual) em que o currículo mínimo exige o tempo procurando
// emprego. Vale para a candidatura, a associação pelo banco e a listagem do
// banco de currículos.
var SituacoesProcurandoEmprego = []string{"Desempregado(a)", "Buscando primeiro emprego"}
