package empregabilidade

import "time"

// BancoCurriculoFilter filtra a listagem do banco de currículos.
type BancoCurriculoFilter struct {
	// Search faz busca parcial por nome ou nome social, ignorando acentos.
	Search string
}

// BancoCurriculoItem é uma linha da listagem do banco de currículos.
// Campo nulo significa dado não informado pelo cidadão.
type BancoCurriculoItem struct {
	CPF          string    `json:"cpf"`
	Nome         *string   `json:"nome"`
	NomeSocial   *string   `json:"nome_social"`
	Profissao    *string   `json:"profissao"`
	Escolaridade *string   `json:"escolaridade"`
	DataInclusao time.Time `json:"data_inclusao"`
}

// BancoCurriculoListMeta é a paginação da listagem do banco de currículos.
type BancoCurriculoListMeta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

// BancoCurriculoListResponse é a resposta da listagem do banco de currículos.
type BancoCurriculoListResponse struct {
	Data []*BancoCurriculoItem  `json:"data"`
	Meta BancoCurriculoListMeta `json:"meta"`
}

// BancoCurriculoDetalhe é a ficha de um currículo no banco de currículos:
// os dados pessoais vêm do cadastro do cidadão (citizen_snapshots) e o
// restante do currículo em si. Campo nulo significa dado não informado.
type BancoCurriculoDetalhe struct {
	CPF          string    `json:"cpf"`
	Nome         *string   `json:"nome"`
	NomeSocial   *string   `json:"nome_social"`
	DataInclusao time.Time `json:"data_inclusao"`
	Profissao    *string   `json:"profissao"`
	Escolaridade *string   `json:"escolaridade"`
	Bairro       *string   `json:"bairro"`
	Celular      *string   `json:"celular"`
	// Genero é a identidade de gênero autodeclarada. O sexo da base cadastral
	// não é usado como fallback.
	Genero    *string            `json:"genero"`
	Idade     *int               `json:"idade"`
	Curriculo *CurriculoCompleto `json:"curriculo"`
}
