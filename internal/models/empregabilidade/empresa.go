package empregabilidade

import (
	"time"
)

type Empresa struct {
	CNPJ         string    `json:"cnpj" gorm:"type:varchar(18);primaryKey"`
	RazaoSocial  string    `json:"razao_social" gorm:"type:varchar(500)"`
	NomeFantasia string    `json:"nome_fantasia" gorm:"type:varchar(500)"`
	Descricao    string    `json:"descricao" gorm:"type:text"`
	URLLogo      string    `json:"url_logo" gorm:"type:varchar(1000)"`
	Website      string    `json:"website" gorm:"type:varchar(1000)"`
	Setor        string    `json:"setor" gorm:"type:varchar(500)"`
	Porte        string    `json:"porte" gorm:"type:varchar(100)"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Empresa) TableName() string {
	return "emp_empresas"
}
