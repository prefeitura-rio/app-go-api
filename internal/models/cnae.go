package models

type CNAE struct {
	Codigo   string `json:"codigo" gorm:"primaryKey;type:varchar(20)"`
	Ocupacao string `json:"ocupacao" gorm:"type:varchar(255);not null"`
	Servico  string `json:"servico" gorm:"type:varchar(500);not null"`

	// Relacionamentos
	Oportunidades []OportunidadeMEI `json:"oportunidades,omitempty" gorm:"foreignKey:CNAECodigo" swaggerignore:"true"`
	MEIEmpresas   []MEIEmpresa      `json:"-" gorm:"many2many:mei_empresas_cnaes;" swaggerignore:"true"`
}

func (CNAE) TableName() string {
	return "cnaes"
}
