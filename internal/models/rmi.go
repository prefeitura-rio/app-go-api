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

// LegalEntityFull represents complete information about a CNPJ from RMI API
// Response from GET /v1/legal-entity/{cnpj} with all fields
type LegalEntityFull struct {
	CNPJ            string   `json:"cnpj"`
	RazaoSocial     string   `json:"razao_social"`
	NomeFantasia    *string  `json:"nome_fantasia"`
	CapitalSocial   float64  `json:"capital_social"`
	CNAEFiscal      *string  `json:"cnae_fiscal"`
	CNAESecundarias []string `json:"cnae_secundarias"`
	Nire            *string  `json:"nire"`

	NaturezaJuridica *struct {
		ID        string `json:"id"`
		Descricao string `json:"descricao"`
	} `json:"natureza_juridica"`

	Porte *struct {
		ID        string `json:"id"`
		Descricao string `json:"descricao"`
	} `json:"porte"`

	MatrizFilial *struct {
		ID        string `json:"id"`
		Descricao string `json:"descricao"`
	} `json:"matriz_filial"`

	OrgaoRegistro *struct {
		ID        string `json:"id"`
		Descricao string `json:"descricao"`
	} `json:"orgao_registro"`

	InicioAtividadeData string `json:"inicio_atividade_data"`

	SituacaoCadastral *struct {
		ID        string `json:"id"`
		Descricao string `json:"descricao"`
		Data      string `json:"data"`
		Motivo    string `json:"motivo"`
	} `json:"situacao_cadastral"`

	SituacaoEspecial *struct {
		ID        string `json:"id"`
		Descricao string `json:"descricao"`
		Data      string `json:"data"`
	} `json:"situacao_especial"`

	EnteFederativo *struct {
		ID        string `json:"id"`
		Descricao string `json:"descricao"`
	} `json:"ente_federativo"`

	Contato *struct {
		Telefone1 string `json:"telefone_1"`
		Telefone2 string `json:"telefone_2"`
		Fax       string `json:"fax"`
		Email     string `json:"email"`
	} `json:"contato"`

	Endereco *struct {
		CEP            string `json:"cep"`
		UF             string `json:"uf"`
		Municipio      string `json:"municipio"`
		MunicipioID    string `json:"municipio_id"`
		Bairro         string `json:"bairro"`
		TipoLogradouro string `json:"tipo_logradouro"`
		Logradouro     string `json:"logradouro"`
		Numero         string `json:"numero"`
		Complemento    string `json:"complemento"`
	} `json:"endereco"`

	Contador *struct {
		PessoaFisica *struct {
			CPF  string `json:"cpf"`
			CRC  string `json:"crc"`
			Nome string `json:"nome"`
		} `json:"pessoa_fisica"`
		PessoaJuridica *struct {
			CNPJ string `json:"cnpj"`
			CRC  string `json:"crc"`
			Nome string `json:"nome"`
		} `json:"pessoa_juridica"`
	} `json:"contador"`

	Responsavel *struct {
		CPF          string `json:"cpf"`
		Qualificacao string `json:"qualificacao"`
	} `json:"responsavel"`

	TiposUnidade  []string `json:"tipos_unidade"`
	FormasAtuacao []string `json:"formas_atuacao"`

	SociosQuantidade int `json:"socios_quantidade"`
	Socios           []struct {
		CPFSocio           string `json:"cpf_socio"`
		CNPJSocio          string `json:"cnpj_socio"`
		NomeSocio          string `json:"nome_socio_estrangeiro"`
		Qualificacao       string `json:"qualificacao"`
		RepresentanteLegal *struct {
			CPF          string `json:"cpf"`
			Nome         string `json:"nome"`
			Qualificacao string `json:"qualificacao"`
		} `json:"representante_legal"`
	} `json:"socios"`
}

// LegalEntityConsultaResponse represents the response for CNPJ lookup endpoint
type LegalEntityConsultaResponse struct {
	CNPJ              string  `json:"cnpj"`
	RazaoSocial       string  `json:"razao_social"`
	NomeFantasia      *string `json:"nome_fantasia"`
	Porte             *string `json:"porte"`
	CNAEFiscal        *string `json:"cnae_fiscal"`
	SituacaoCadastral *string `json:"situacao_cadastral"`
	Endereco          *struct {
		CEP         string `json:"cep"`
		UF          string `json:"uf"`
		Municipio   string `json:"municipio"`
		Bairro      string `json:"bairro"`
		Logradouro  string `json:"logradouro"`
		Numero      string `json:"numero"`
		Complemento string `json:"complemento"`
	} `json:"endereco"`
	Contato *struct {
		Telefone string `json:"telefone"`
		Email    string `json:"email"`
	} `json:"contato"`
}

// ToConsultaResponse converts LegalEntityFull to a simplified response for the frontend
func (l *LegalEntityFull) ToConsultaResponse() *LegalEntityConsultaResponse {
	resp := &LegalEntityConsultaResponse{
		CNPJ:         l.CNPJ,
		RazaoSocial:  l.RazaoSocial,
		NomeFantasia: l.NomeFantasia,
		CNAEFiscal:   l.CNAEFiscal,
	}

	if l.Porte != nil {
		resp.Porte = &l.Porte.Descricao
	}

	if l.SituacaoCadastral != nil {
		resp.SituacaoCadastral = &l.SituacaoCadastral.Descricao
	}

	if l.Endereco != nil {
		resp.Endereco = &struct {
			CEP         string `json:"cep"`
			UF          string `json:"uf"`
			Municipio   string `json:"municipio"`
			Bairro      string `json:"bairro"`
			Logradouro  string `json:"logradouro"`
			Numero      string `json:"numero"`
			Complemento string `json:"complemento"`
		}{
			CEP:         l.Endereco.CEP,
			UF:          l.Endereco.UF,
			Municipio:   l.Endereco.Municipio,
			Bairro:      l.Endereco.Bairro,
			Logradouro:  l.Endereco.Logradouro,
			Numero:      l.Endereco.Numero,
			Complemento: l.Endereco.Complemento,
		}
	}

	if l.Contato != nil {
		telefone := l.Contato.Telefone1
		if telefone == "" {
			telefone = l.Contato.Telefone2
		}
		resp.Contato = &struct {
			Telefone string `json:"telefone"`
			Email    string `json:"email"`
		}{
			Telefone: telefone,
			Email:    l.Contato.Email,
		}
	}

	return resp
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
	CPF           string               `json:"cpf"`
	Nome          string               `json:"nome"`
	NomeSocial    string               `json:"nome_social"`
	Email         CitizenEmailInfo     `json:"email"`
	Telefone      CitizenTelefoneInfo  `json:"telefone"`
	Nascimento    *CitizenNascimento   `json:"nascimento"`
	Endereco      *CitizenEnderecoInfo `json:"endereco"`
	Raca          string               `json:"raca"`
	Genero        string               `json:"genero"`
	Sexo          string               `json:"sexo"`
	RendaFamiliar string               `json:"renda_familiar"`
	Escolaridade  string               `json:"escolaridade"`
	Deficiencia   string               `json:"deficiencia"`
}

// CitizenEmailInfo represents the email field structure from citizen API
type CitizenEmailInfo struct {
	Indicador bool `json:"indicador"`
	Principal struct {
		Valor     string `json:"valor"`
		Origem    string `json:"origem"`
		Sistema   string `json:"sistema"`
		UpdatedAt string `json:"updated_at"`
	} `json:"principal"`
}

// CitizenTelefoneInfo represents the telefone field structure from citizen API
type CitizenTelefoneInfo struct {
	Indicador bool `json:"indicador"`
	Principal struct {
		DDI       string `json:"ddi"`
		DDD       string `json:"ddd"`
		Valor     string `json:"valor"`
		Origem    string `json:"origem"`
		Sistema   string `json:"sistema"`
		UpdatedAt string `json:"updated_at"`
	} `json:"principal"`
}

// CitizenNascimento represents the birth information from citizen API
type CitizenNascimento struct {
	Data        string `json:"data"`
	Municipio   string `json:"municipio"`
	MunicipioID string `json:"municipio_id"`
	UF          string `json:"uf"`
	Pais        string `json:"pais"`
	PaisID      string `json:"pais_id"`
}

// CitizenEnderecoInfo represents the address field structure
type CitizenEnderecoInfo struct {
	Indicador bool `json:"indicador"`
	Principal struct {
		Logradouro     string `json:"logradouro"`
		TipoLogradouro string `json:"tipo_logradouro"`
		Numero         string `json:"numero"`
		Complemento    string `json:"complemento"`
		Bairro         string `json:"bairro"`
		Municipio      string `json:"municipio"`
		Estado         string `json:"estado"`
		CEP            string `json:"cep"`
		Origem         string `json:"origem"`
		Sistema        string `json:"sistema"`
		UpdatedAt      string `json:"updated_at"`
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
		// Return formatted as DDI+DDD+Valor (e.g., 5521997015128)
		return c.Telefone.Principal.DDI + c.Telefone.Principal.DDD + c.Telefone.Principal.Valor
	}
	return ""
}

// GetDataNascimento returns the birth date string
func (c *CitizenContactInfo) GetDataNascimento() string {
	if c.Nascimento != nil && c.Nascimento.Data != "" {
		return c.Nascimento.Data
	}
	return ""
}

// GetRaca returns the race/ethnicity value
func (c *CitizenContactInfo) GetRaca() string {
	return c.Raca
}

// GetGenero returns the gender value (prefers genero, falls back to sexo)
func (c *CitizenContactInfo) GetGenero() string {
	if c.Genero != "" {
		return c.Genero
	}
	return c.Sexo
}

// GetRendaFamiliar returns the family income value
func (c *CitizenContactInfo) GetRendaFamiliar() string {
	return c.RendaFamiliar
}

// GetEscolaridade returns the education level value
func (c *CitizenContactInfo) GetEscolaridade() string {
	return c.Escolaridade
}

// GetDeficiencia returns the disability description
func (c *CitizenContactInfo) GetDeficiencia() string {
	return c.Deficiencia
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
