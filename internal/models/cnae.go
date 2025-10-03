package models

type CNAE struct {
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Codigo   string `json:"codigo" gorm:"type:varchar(20);not null;uniqueIndex:idx_codigo_servico"`
	Ocupacao string `json:"ocupacao" gorm:"type:varchar(255);not null"`
	Servico  string `json:"servico" gorm:"type:varchar(500);not null;uniqueIndex:idx_codigo_servico"`

	// Relacionamentos
	Oportunidades []OportunidadeMEI `json:"oportunidades,omitempty" gorm:"foreignKey:CNAEID" swaggerignore:"true"`
	MEIEmpresas   []MEIEmpresa      `json:"-" gorm:"many2many:mei_empresas_cnaes;joinForeignKey:cnae_id;joinReferences:mei_empresa_id" swaggerignore:"true"`
}

func (CNAE) TableName() string {
	return "cnaes"
}
