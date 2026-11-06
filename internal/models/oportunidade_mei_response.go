package models

// OportunidadeMEIListMeta é a paginação da listagem de oportunidades MEI.
type OportunidadeMEIListMeta struct {
	Page     int `json:"page" example:"1"`
	PageSize int `json:"page_size" example:"10"`
	Total    int `json:"total" example:"0"`
}

// OportunidadeMEIListResponse é a resposta da listagem de oportunidades MEI
// (GET /api/public/oportunidades-mei e GET /api/v1/oportunidades-mei), só para
// a documentação OpenAPI. Sem resultados, data sai como null.
type OportunidadeMEIListResponse struct {
	Data []OportunidadeMEI       `json:"data" extensions:"x-nullable"`
	Meta OportunidadeMEIListMeta `json:"meta"`
}
