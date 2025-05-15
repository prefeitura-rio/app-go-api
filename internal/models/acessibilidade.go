package models

type Acessibilidade struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Nome string `json:"nome" gorm:"not null"`
	
	// Relacionamentos
	Cursos []Curso `json:"cursos,omitempty" gorm:"many2many:cursos_acessibilidades;"`
}

// TableName especifica o nome da tabela para este modelo
func (Acessibilidade) TableName() string {
	return "acessibilidades"
} 