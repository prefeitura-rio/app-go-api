package models

// LegalEntity represents a legal entity (CNPJ) from the RMI API
type LegalEntity struct {
	CNPJ             string   `json:"cnpj"`
	CNAEFiscal       string   `json:"cnae_fiscal"`
	CNAESecundarias  []string `json:"cnae_secundarias"`
	RazaoSocial      string   `json:"razao_social"`
	NomeFantasia     string   `json:"nome_fantasia"`
	SituacaoCadastral struct {
		ID        string `json:"id"`
		Descricao string `json:"descricao"`
		Data      string `json:"data"`
	} `json:"situacao_cadastral"`
	// Other fields can be added as needed
}

// LegalEntitiesResponse represents the paginated response from RMI API
type LegalEntitiesResponse struct {
	Data       []LegalEntity      `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginationResponse represents pagination metadata from RMI API
type PaginationResponse struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// GetAllCNAEs returns all CNAEs for this legal entity (fiscal + secundarias)
func (l *LegalEntity) GetAllCNAEs() []string {
	cnaes := []string{}

	// Add fiscal CNAE if not empty
	if l.CNAEFiscal != "" {
		cnaes = append(cnaes, l.CNAEFiscal)
	}

	// Add secondary CNAEs
	if l.CNAESecundarias != nil {
		cnaes = append(cnaes, l.CNAESecundarias...)
	}

	return cnaes
}
