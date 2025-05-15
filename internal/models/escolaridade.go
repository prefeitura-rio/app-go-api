package models

type Escolaridade struct {
	ID    int    `json:"id" gorm:"primaryKey"`
	Nivel string `json:"nivel" gorm:"not null"`
	
	// Relacionamentos
	Empregos []Emprego `json:"empregos,omitempty" gorm:"foreignKey:EscolaridadeID"`
}

// TableName especifica o nome da tabela para este modelo
func (Escolaridade) TableName() string {
	return "escolaridades"
} 