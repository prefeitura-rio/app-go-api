package models

type Empresa struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Nome string `json:"nome" gorm:"type:varchar(20000);not null"`

	// Relacionamentos
	Empregos []Emprego `json:"empregos,omitempty" gorm:"foreignKey:EmpresaID"`
}

func (e *Empresa) SetID(id int) {
	e.ID = id
}

func (Empresa) TableName() string {
	return "empresas"
} 