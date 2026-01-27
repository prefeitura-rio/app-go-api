package empregabilidade

import (
	"time"
)

type Onboarding struct {
	CPF                         string    `json:"cpf" gorm:"type:varchar(14);primaryKey"`
	IsEmpregabilidadeFirstLogin bool      `json:"is_empregabilidade_first_login" gorm:"default:true"`
	CreatedAt                   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Onboarding) TableName() string {
	return "emp_onboarding"
}
