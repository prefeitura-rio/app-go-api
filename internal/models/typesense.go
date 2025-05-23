package models

// Documento genérico do Typesense
type Document map[string]interface{}

// Resposta com documentos
type SearchDocumentsResponse struct {
	Found         int                       `json:"found"`
	Hits          []map[string]interface{} `json:"hits"`
	Page          int                       `json:"page"`
	RequestParams map[string]interface{}   `json:"request_params"`
}

// Parâmetros para busca de documentos
type SearchParameters struct {
	Q                   string `json:"q"`
	QueryBy             string `json:"query_by"`
	FilterBy            string `json:"filter_by,omitempty"`
	SortBy              string `json:"sort_by,omitempty"`
	Page                int    `json:"page,omitempty"`
	PerPage             int    `json:"per_page,omitempty"`
	FacetBy             string `json:"facet_by,omitempty"`
	MaxFacetValues      int    `json:"max_facet_values,omitempty"`
	FacetQuery          string `json:"facet_query,omitempty"`
	IncludeFields       string `json:"include_fields,omitempty"`
	ExcludeFields       string `json:"exclude_fields,omitempty"`
	HighlightFields     string `json:"highlight_fields,omitempty"`
	HighlightFullFields string `json:"highlight_full_fields,omitempty"`
	Prefix              bool   `json:"prefix,omitempty"`
	NumTypos            int    `json:"num_typos,omitempty"`
}

// Resposta de erro
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// MultiCollectionSearchParameters contém parâmetros para busca em múltiplas coleções
type MultiCollectionSearchParameters struct {
	Collections []string        `json:"collections"` // Lista de nomes das coleções para buscar
	Params      SearchParameters `json:"params"`     // Parâmetros de busca
}

// MultiCollectionSearchResponse contém resultados de busca em múltiplas coleções
type MultiCollectionSearchResponse struct {
	TotalFound int                               `json:"total_found"`
	Results    map[string]SearchDocumentsResponse `json:"results"`      // Mapa de resultados por coleção
}