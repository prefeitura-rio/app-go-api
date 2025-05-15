package models

type InstituicaoEnsino struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Nome string `json:"nome" gorm:"not null"`
	
	// Relacionamentos
	Cursos []Curso `json:"cursos,omitempty" gorm:"foreignKey:InstituicaoID"`
}

// TableName especifica o nome da tabela para este modelo
func (InstituicaoEnsino) TableName() string {
	return "instituicoes_ensino"
} 