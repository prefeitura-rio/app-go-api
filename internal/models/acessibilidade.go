package models

type Acessibilidade struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Nome string `json:"nome" gorm:"type:varchar(20000);not null"`

	// Relacionamentos
	Cursos []Curso `json:"cursos,omitempty" gorm:"many2many:cursos_acessibilidades;"`
}

func (a *Acessibilidade) SetID(id int) {
	a.ID = id
}

func (Acessibilidade) TableName() string {
	return "acessibilidades"
} 