package empregabilidade

import (
	"time"
)

// ComportamentoAtitudes representa a tabela 'emp_comportamento_atitudes'
type ComportamentoAtitudes struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Nome      string    `json:"nome" gorm:"type:varchar(250);unique;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ComportamentoAtitudes) TableName() string {
	return "emp_comportamento_atitudes"
}

// ComportamentoAtitudesFilter estrutura para aplicar filtros na consulta
type ComportamentoAtitudesFilter struct {
	Search string `json:"search"`
	Nome   string `json:"nome"`
	Page   int
	Limit  int
}

// CurriculoComportamentoAtitudes representa o vínculo entre o candidato (CPF) e seus Comportamentos/Atitudes
type CurriculoComportamentoAtitudes struct {
	ID                      int64                  `json:"id" gorm:"primaryKey;autoIncrement"`
	CPF                     string                 `json:"cpf" gorm:"type:char(11);not null;uniqueIndex:uk_emp_curriculo_comportamento_atitudes_cpf_comportamento"`
	IDComportamentoAtitudes int64                  `json:"id_comportamento_atitudes" gorm:"not null;uniqueIndex:uk_emp_curriculo_comportamento_atitudes_cpf_comportamento"`
	CreatedAt               time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	ComportamentoAtitudes   *ComportamentoAtitudes `json:"comportamento_atitudes,omitempty" gorm:"foreignKey:IDComportamentoAtitudes;references:ID"`
}

func (CurriculoComportamentoAtitudes) TableName() string {
	return "emp_curriculo_comportamento_atitudes"
}

// CurriculoComportamentoAtitudesFilter estrutura para aplicar filtros na consulta
type CurriculoComportamentoAtitudesFilter struct {
	Search                  string `json:"search"`
	CPF                     string `json:"cpf"`
	IDComportamentoAtitudes int64  `json:"id_comportamento_atitudes"`
	Page                    int
	Limit                   int
}
