package empregabilidade

import (
	"time"

	"github.com/google/uuid"
)

// CandidaturaBloqueio registra a tentativa de candidatura de um CPF a uma vaga
// que foi bloqueada por não atender a um ou mais critérios de elegibilidade
// configurados na vaga. Não há Candidatura associada nesse fluxo (o candidato
// nunca chega a se candidatar de fato), por isso IDVaga não possui FK.
type CandidaturaBloqueio struct {
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CPF                   string    `json:"cpf" gorm:"type:varchar(14);not null"`
	IDVaga                uuid.UUID `json:"id_vaga" gorm:"type:uuid;not null"`
	CriteriosNaoAtendidos []string  `json:"criterios_nao_atendidos" gorm:"type:jsonb;serializer:json"`
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (CandidaturaBloqueio) TableName() string {
	return "emp_candidatura_bloqueios"
}
