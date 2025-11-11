package models

type InstituicaoEnsino struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Nome string `json:"nome" gorm:"type:varchar(20000);not null"`

	// Relacionamentos
	Cursos []Curso `json:"cursos,omitempty" gorm:"foreignKey:InstituicaoID"`
}

func (i *InstituicaoEnsino) SetID(id int) {
	i.ID = id
}

func (InstituicaoEnsino) TableName() string {
	return "instituicoes_ensino"
}
