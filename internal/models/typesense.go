package models

// Campos padrões para todos os esquemas de coleção Typesense
type CollectionField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Facet    bool   `json:"facet,omitempty"`
	Index    bool   `json:"index,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

// Requisição para criar uma coleção
type CreateCollectionRequest struct {
	Name                   string           `json:"name"`
	Fields                 []CollectionField `json:"fields"`
	DefaultSortingField    string           `json:"default_sorting_field,omitempty"`
	EnableNestedFields     bool             `json:"enable_nested_fields,omitempty"`
	TokenSeparators        []string         `json:"token_separators,omitempty"`
	SymbolsToIndex         []string         `json:"symbols_to_index,omitempty"`
	SymbolsToIndexSeparate []string         `json:"symbols_to_index_separate,omitempty"`
}

// Resposta para criação de coleção
type CollectionResponse struct {
	Name                   string           `json:"name"`
	Fields                 []CollectionField `json:"fields"`
	DefaultSortingField    string           `json:"default_sorting_field,omitempty"`
	EnableNestedFields     bool             `json:"enable_nested_fields,omitempty"`
	TokenSeparators        []string         `json:"token_separators,omitempty"`
	SymbolsToIndex         []string         `json:"symbols_to_index,omitempty"`
	SymbolsToIndexSeparate []string         `json:"symbols_to_index_separate,omitempty"`
	CreatedAt              int64            `json:"created_at"`
	NumDocuments           int              `json:"num_documents"`
}

// Documento genérico do Typesense
type Document map[string]interface{}

// Requisição para upsert de documento
type UpsertDocumentRequest struct {
	Document      Document `json:"document"`
	ActionOnMatch string   `json:"action_on_match,omitempty"` // "update", "create", "upsert"
}

// Resposta com documentos
type SearchDocumentsResponse struct {
	Found      int                    `json:"found"`
	Hits       []map[string]interface{} `json:"hits"`
	Page       int                    `json:"page"`
	RequestParams map[string]interface{} `json:"request_params"`
}

// Parâmetros para busca de documentos
type SearchParameters struct {
	Q               string   `json:"q"`
	QueryBy         string   `json:"query_by"`
	FilterBy        string   `json:"filter_by,omitempty"`
	SortBy          string   `json:"sort_by,omitempty"`
	Page            int      `json:"page,omitempty"`
	PerPage         int      `json:"per_page,omitempty"`
	FacetBy         string   `json:"facet_by,omitempty"`
	MaxFacetValues  int      `json:"max_facet_values,omitempty"`
	FacetQuery      string   `json:"facet_query,omitempty"`
	IncludeFields   string   `json:"include_fields,omitempty"`
	ExcludeFields   string   `json:"exclude_fields,omitempty"`
	HighlightFields string   `json:"highlight_fields,omitempty"`
	HighlightFullFields string `json:"highlight_full_fields,omitempty"`
	Prefix          bool     `json:"prefix,omitempty"`
	NumTypos        int      `json:"num_typos,omitempty"`
}

// Resposta de erro
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
} 