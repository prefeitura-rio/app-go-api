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
	CPF            string                  `json:"cpf"`
	Nome           string                  `json:"nome"`
	Email          CitizenEmailInfo        `json:"email"`
	Telefone       CitizenTelefoneInfo     `json:"telefone"`
	DataNascimento *CitizenDataNascimento  `json:"data_nascimento"`
	Endereco       *CitizenEnderecoInfo    `json:"endereco"`
	Raca           *CitizenIndicadorInfo   `json:"raca"`
	Genero         *CitizenIndicadorInfo   `json:"genero"`
	RendaFamiliar  *CitizenIndicadorInfo   `json:"renda_familiar"`
	Escolaridade   *CitizenIndicadorInfo   `json:"escolaridade"`
	Deficiencia    *CitizenDeficienciaInfo `json:"deficiencia"`
}

// CitizenEmailInfo represents the email field structure from citizen API
type CitizenEmailInfo struct {
	Indicador bool `json:"indicador"`
	Principal struct {
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

// CitizenDataNascimento represents the birth date field structure
type CitizenDataNascimento struct {
	Indicador bool `json:"indicador"`
	Principal struct {
		Valor string `json:"valor"` // Format: YYYY-MM-DD or DD/MM/YYYY
	} `json:"principal"`
}

// CitizenEnderecoInfo represents the address field structure
type CitizenEnderecoInfo struct {
	Indicador bool `json:"indicador"`
	Principal struct {
		Logradouro  string `json:"logradouro"`
		Numero      string `json:"numero"`
		Complemento string `json:"complemento"`
		Bairro      string `json:"bairro"`
		Cidade      string `json:"cidade"`
		UF          string `json:"uf"`
		CEP         string `json:"cep"`
	} `json:"principal"`
}

// CitizenIndicadorInfo represents a generic field with indicador/principal/valor structure
type CitizenIndicadorInfo struct {
	Indicador bool `json:"indicador"`
	Principal struct {
		Valor string `json:"valor"`
	} `json:"principal"`
}

// CitizenDeficienciaInfo represents the disability field structure
type CitizenDeficienciaInfo struct {
	Indicador bool `json:"indicador"`
	Principal struct {
		Valor       string   `json:"valor"`
		Descricao   string   `json:"descricao"`
		Deficiencia []string `json:"deficiencia"`
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

// GetDataNascimento returns the birth date string
func (c *CitizenContactInfo) GetDataNascimento() string {
	if c.DataNascimento != nil && c.DataNascimento.Indicador && c.DataNascimento.Principal.Valor != "" {
		return c.DataNascimento.Principal.Valor
	}
	return ""
}

// GetRaca returns the race/ethnicity value
func (c *CitizenContactInfo) GetRaca() string {
	if c.Raca != nil && c.Raca.Indicador && c.Raca.Principal.Valor != "" {
		return c.Raca.Principal.Valor
	}
	return ""
}

// GetGenero returns the gender value
func (c *CitizenContactInfo) GetGenero() string {
	if c.Genero != nil && c.Genero.Indicador && c.Genero.Principal.Valor != "" {
		return c.Genero.Principal.Valor
	}
	return ""
}

// GetRendaFamiliar returns the family income value
func (c *CitizenContactInfo) GetRendaFamiliar() string {
	if c.RendaFamiliar != nil && c.RendaFamiliar.Indicador && c.RendaFamiliar.Principal.Valor != "" {
		return c.RendaFamiliar.Principal.Valor
	}
	return ""
}

// GetEscolaridade returns the education level value
func (c *CitizenContactInfo) GetEscolaridade() string {
	if c.Escolaridade != nil && c.Escolaridade.Indicador && c.Escolaridade.Principal.Valor != "" {
		return c.Escolaridade.Principal.Valor
	}
	return ""
}

// GetDeficiencia returns the disability description
func (c *CitizenContactInfo) GetDeficiencia() string {
	if c.Deficiencia != nil && c.Deficiencia.Indicador {
		if c.Deficiencia.Principal.Descricao != "" {
			return c.Deficiencia.Principal.Descricao
		}
		if c.Deficiencia.Principal.Valor != "" {
			return c.Deficiencia.Principal.Valor
		}
	}
	return ""
}

// GetEndereco returns the address as a structured object
func (c *CitizenContactInfo) GetEndereco() *CitizenEnderecoInfo {
	if c.Endereco != nil && c.Endereco.Indicador {
		return c.Endereco
	}
	return nil
}

// CNPJOwnerInfo represents the final enriched contact information
// This is what we return to the service layer
type CNPJOwnerInfo struct {
	CNPJ                string
	EmailPessoaFisica   string
	CelularPessoaFisica string
}
