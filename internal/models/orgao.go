package models

type Orgao struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Nome string `json:"nome" gorm:"type:varchar(20000);not null"`

	// Relacionamentos
	Cursos  []Curso  `json:"cursos,omitempty" gorm:"foreignKey:OrgaoID"`
	Empregos []Emprego `json:"empregos,omitempty" gorm:"foreignKey:OrgaoID"`
}

func (o *Orgao) SetID(id int) {
	o.ID = id
}

func (Orgao) TableName() string {
	return "orgaos"
} 