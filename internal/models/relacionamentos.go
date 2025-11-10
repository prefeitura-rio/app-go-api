package models

// CursoCategoria representa a tabela de relacionamento many-to-many entre Curso e Categoria
type CursoCategoria struct {
	CursoID     int `json:"curso_id" gorm:"primaryKey;column:curso_id"`
	CategoriaID int `json:"categoria_id" gorm:"primaryKey;column:categoria_id"`
}

// TableName especifica o nome da tabela para este modelo
func (CursoCategoria) TableName() string {
	return "cursos_categorias"
}

// CursoAcessibilidade representa a tabela de relacionamento many-to-many entre Curso e Acessibilidade
type CursoAcessibilidade struct {
	CursoID          int `json:"curso_id" gorm:"primaryKey;column:curso_id"`
	AcessibilidadeID int `json:"acessibilidade_id" gorm:"primaryKey;column:acessibilidade_id"`
}

// TableName especifica o nome da tabela para este modelo
func (CursoAcessibilidade) TableName() string {
	return "cursos_acessibilidades"
}
