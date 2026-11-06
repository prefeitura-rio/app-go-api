package models

// Schemas de resposta dos endpoints de curso, só para a documentação OpenAPI
// (swag). Descrevem o JSON que a API devolve: as listagens montam cada item com
// transformCursoToResponse (internal/handlers/v1/course_handler.go) e o detalhe
// serializa models.Curso dentro de { success, data }. Um teste do pacote de
// handlers confere que as chaves deste schema são as que a resposta traz.

// CursoResponse é um curso como a API o devolve: os campos de Curso mais os
// relacionamentos carregados. Os relacionamentos ausentes saem como null.
type CursoResponse struct {
	Curso

	Instituicao     *InstituicaoEnsino `json:"instituicao" extensions:"x-nullable"`
	Categorias      []Categoria        `json:"categorias" extensions:"x-nullable"`
	Acessibilidades []Acessibilidade   `json:"acessibilidades" extensions:"x-nullable"`
	CustomFields    []CustomField      `json:"custom_fields" extensions:"x-nullable"`
	LocationClasses []LocationClass    `json:"locations" extensions:"x-nullable"`
	RemoteClass     *RemoteClass       `json:"remote_class" extensions:"x-nullable"`
}

// CursoPagination é a paginação das listagens de cursos.
type CursoPagination struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"10"`
	Total      int `json:"total" example:"394"`
	TotalPages int `json:"total_pages" example:"40"`
}

// CursoListData é o conteúdo de data nas listagens de cursos.
type CursoListData struct {
	Courses    []CursoResponse `json:"courses"`
	Pagination CursoPagination `json:"pagination"`
}

// CursoListResponse é a resposta das listagens de cursos
// (GET /api/public/courses e GET /api/v1/courses).
type CursoListResponse struct {
	Success bool          `json:"success" example:"true"`
	Data    CursoListData `json:"data"`
}

// CursoDetailResponse é a resposta do detalhe de um curso
// (GET /api/public/courses/{courseId} e GET /api/v1/courses/{courseId}).
type CursoDetailResponse struct {
	Success bool          `json:"success" example:"true"`
	Data    CursoResponse `json:"data"`
}
