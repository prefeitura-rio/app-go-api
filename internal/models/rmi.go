package models

// LegalEntity represents a legal entity (CNPJ) from the RMI API
type LegalEntity struct {
	CNPJ              string   `json:"cnpj"`
	CNAEFiscal        string   `json:"cnae_fiscal"`
	CNAESecundarias   []string `json:"cnae_secundarias"`
	RazaoSocial       string   `json:"razao_social"`
	NomeFantasia      string   `json:"nome_fantasia"`
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

// Orgao represents an orgao (department/government entity) from the RMI API
// Response from GET /v1/departments/{orgao_id}
type Orgao struct {
	ID            string `json:"id"`
	CdUA          string `json:"cd_ua"`
	SiglaUA       string `json:"sigla_ua"` // Short name/acronym
	NomeUA        string `json:"nome_ua"`  // Full name
	CdUAPai       string `json:"cd_ua_pai"`
	Nivel         int    `json:"nivel"`
	OrdemUABasica string `json:"ordem_ua_basica"`
	OrdemAbsoluta string `json:"ordem_absoluta"`
	OrdemRelativa string `json:"ordem_relativa"`
}

// LegalEntityDetails represents detailed information about a CNPJ
// Response from GET /v1/legal-entity/{cnpj}
type LegalEntityDetails struct {
	CNPJ        string `json:"cnpj"`
	RazaoSocial string `json:"razao_social"`
	Responsavel struct {
		CPF string `json:"cpf"`
	} `json:"responsavel"`
	Socios []struct {
		CPFSocio  string `json:"cpf_socio"`
		NomeSocio string `json:"nome_socio_estrangeiro"`
	} `json:"socios"`
}

// GetSocioCPFs returns all CPFs from socios list
func (l *LegalEntityDetails) GetSocioCPFs() []string {
	cpfs := []string{}
	for _, socio := range l.Socios {
		if socio.CPFSocio != "" {
			cpfs = append(cpfs, socio.CPFSocio)
		}
	}
	// Also add responsavel CPF if not already in list
	if l.Responsavel.CPF != "" {
		found := false
		for _, cpf := range cpfs {
			if cpf == l.Responsavel.CPF {
				found = true
				break
			}
		}
		if !found {
			cpfs = append(cpfs, l.Responsavel.CPF)
		}
	}
	return cpfs
}

// CitizenContactInfo represents contact information for a citizen
// Response from GET /v1/citizen/{cpf}
type CitizenContactInfo struct {
	CPF      string `json:"cpf"`
	Nome     string `json:"nome"`
	Email    CitizenEmailInfo    `json:"email"`
	Telefone CitizenTelefoneInfo `json:"telefone"`
}

// CitizenEmailInfo represents the email field structure from citizen API
type CitizenEmailInfo struct {
	Indicador  bool `json:"indicador"`
	Principal  struct {
		Valor string `json:"valor"`
	} `json:"principal"`
}

// CitizenTelefoneInfo represents the telefone field structure from citizen API
type CitizenTelefoneInfo struct {
	Indicador bool `json:"indicador"`
	Principal struct {
		DDI   string `json:"ddi"`
		DDD   string `json:"ddd"`
		Valor string `json:"valor"`
	} `json:"principal"`
}

// GetEmail returns the email value from the nested structure
func (c *CitizenContactInfo) GetEmail() string {
	if c.Email.Indicador && c.Email.Principal.Valor != "" {
		return c.Email.Principal.Valor
	}
	return ""
}

// GetCelular returns the formatted phone number
func (c *CitizenContactInfo) GetCelular() string {
	if c.Telefone.Indicador && c.Telefone.Principal.Valor != "" {
		// Return formatted as +DDI DDD Valor (e.g., +55 21 997015128)
		return c.Telefone.Principal.DDI + c.Telefone.Principal.DDD + c.Telefone.Principal.Valor
	}
	return ""
}

// CNPJOwnerInfo represents the final enriched contact information
// This is what we return to the service layer
type CNPJOwnerInfo struct {
	CNPJ                string
	EmailPessoaFisica   string
	CelularPessoaFisica string
}
