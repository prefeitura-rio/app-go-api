package models

type Empresa struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Nome string `json:"nome" gorm:"not null"`
	
	// Relacionamentos
	Empregos []Emprego `json:"empregos,omitempty" gorm:"foreignKey:EmpresaID"`
}

// TableName especifica o nome da tabela para este modelo
func (Empresa) TableName() string {
	return "empresas"
} 