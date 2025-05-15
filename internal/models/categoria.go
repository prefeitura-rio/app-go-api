package models

type Categoria struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Nome string `json:"nome" gorm:"not null"`
	
	// Relacionamentos
	Cursos []Curso `json:"cursos,omitempty" gorm:"many2many:cursos_categorias;"`
}

// TableName especifica o nome da tabela para este modelo
func (Categoria) TableName() string {
	return "categorias"
} 