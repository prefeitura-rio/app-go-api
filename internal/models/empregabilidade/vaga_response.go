package empregabilidade

// VagaPublicListMeta é a paginação da listagem pública de vagas.
type VagaPublicListMeta struct {
	Page     int `json:"page" example:"1"`
	PageSize int `json:"page_size" example:"10"`
	Total    int `json:"total" example:"109"`
}

// VagaPublicListResponse é a resposta da listagem pública de vagas
// (GET /api/public/empregabilidade/vagas), só para a documentação OpenAPI.
type VagaPublicListResponse struct {
	Data []Vaga             `json:"data" extensions:"x-nullable"`
	Meta VagaPublicListMeta `json:"meta"`
}
