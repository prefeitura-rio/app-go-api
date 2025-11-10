package models

type Escolaridade struct {
	ID    int    `json:"id" gorm:"primaryKey"`
	Nivel string `json:"nivel" gorm:"type:varchar(20000);not null"`

	// Relacionamentos
	Empregos []Emprego `json:"empregos,omitempty" gorm:"foreignKey:EscolaridadeID"`
}

func (e *Escolaridade) SetID(id int) {
	e.ID = id
}

func (Escolaridade) TableName() string {
	return "escolaridades"
}
