package models

type Categoria struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Nome string `json:"nome" gorm:"type:varchar(20000);not null;uniqueIndex:categorias_nome_key"`

	// Relacionamentos
	Cursos []Curso `json:"cursos,omitempty" gorm:"many2many:cursos_categorias;"`
}

func (c *Categoria) SetID(id int) {
	c.ID = id
}

func (Categoria) TableName() string {
	return "categorias"
}
