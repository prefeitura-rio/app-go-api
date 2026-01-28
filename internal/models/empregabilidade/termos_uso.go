package empregabilidade

import (
	"time"
)

type TermosUso struct {
	CPF         string     `json:"cpf" gorm:"type:varchar(14);primaryKey"`
	UserConsent bool       `json:"user_consent" gorm:"default:false"`
	AceitoEm    *time.Time `json:"aceito_em" gorm:"type:timestamp with time zone"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (TermosUso) TableName() string {
	return "emp_termos_uso"
}
