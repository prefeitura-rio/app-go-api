package empregabilidade

import (
	"time"

	"github.com/google/uuid"
)

// BancoCurriculoFilter filtra a listagem do banco de currículos.
type BancoCurriculoFilter struct {
	// Search faz busca parcial por nome ou nome social, ignorando acentos.
	Search string
	// Candidato são os filtros avançados (gênero, idade, experiência etc.).
	Candidato FiltroCandidato
	// SomenteDisponiveis deixa de fora quem se ocultou do banco de currículos no
	// app. O service sempre liga; o recrutador não escolhe.
	SomenteDisponiveis bool
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
	Genero *string `json:"genero"`
	Idade  *int    `json:"idade"`
	Email  *string `json:"email"`
	Raca   *string `json:"raca"`
	// Deficiencia é o que o cidadão declarou sobre deficiência (PCD).
	Deficiencia *string `json:"deficiencia"`
	// DataAtualizacao é o último salvamento do currículo pelo cidadão. É nula
	// quando nenhuma seção foi preenchida.
	DataAtualizacao *time.Time         `json:"data_atualizacao"`
	Curriculo       *CurriculoCompleto `json:"curriculo"`
}

// BancoCurriculoCandidaturaFilter filtra o histórico de candidaturas da ficha.
type BancoCurriculoCandidaturaFilter struct {
	// Search faz busca parcial pelo título da vaga, ignorando acentos.
	Search string
}

// BancoCurriculoCandidatura é uma linha do histórico de candidaturas da ficha:
// as vagas a que o cidadão se candidatou ou foi associado.
type BancoCurriculoCandidatura struct {
	ID uuid.UUID `json:"id"`
	// Status é a situação do candidato na vaga.
	Status StatusCandidatura `json:"status"`
	// DataCandidatura é quando a candidatura foi criada.
	DataCandidatura time.Time `json:"data_candidatura"`
	// Origem é por onde a candidatura entrou; nula nas candidaturas antigas.
	Origem *OrigemCandidatura `json:"origem"`
	// Vaga traz o status já considerando a data limite (publicado_expirado).
	Vaga *Vaga `json:"vaga"`
}

// BancoCurriculoCandidaturaListResponse é a resposta do histórico de candidaturas.
type BancoCurriculoCandidaturaListResponse struct {
	Data []*BancoCurriculoCandidatura `json:"data"`
	Meta BancoCurriculoListMeta       `json:"meta"`
}

// ResultadoAssociacao é o desfecho da associação do candidato a uma vaga.
type ResultadoAssociacao string

const (
	// AssociacaoRealizada: a candidatura foi criada.
	AssociacaoRealizada ResultadoAssociacao = "associado"
	// AssociacaoJaParticipa: o candidato já está no processo seletivo da vaga.
	AssociacaoJaParticipa ResultadoAssociacao = "ja_participa"
	// AssociacaoVagaIndisponivel: a vaga não existe ou não está recebendo candidaturas.
	AssociacaoVagaIndisponivel ResultadoAssociacao = "vaga_indisponivel"
	// AssociacaoForaDoEscopo: a vaga é de um órgão fora da secretaria do operador.
	AssociacaoForaDoEscopo ResultadoAssociacao = "fora_do_escopo"
	// AssociacaoNaoElegivel: o candidato não atende os requisitos da vaga.
	AssociacaoNaoElegivel ResultadoAssociacao = "nao_elegivel"
	// AssociacaoErro: falha inesperada ao associar.
	AssociacaoErro ResultadoAssociacao = "erro"
)

// AssociarVagasRequest pede a associação do candidato a uma ou mais vagas.
type AssociarVagasRequest struct {
	IDsVagas []uuid.UUID `json:"ids_vagas" binding:"required,min=1,max=50"`
}

// AssociacaoVaga é o resultado da associação a uma das vagas pedidas.
type AssociacaoVaga struct {
	IDVaga    uuid.UUID           `json:"id_vaga"`
	Titulo    string              `json:"titulo,omitempty"`
	Resultado ResultadoAssociacao `json:"resultado"`
	// Motivos explica por que a associação não aconteceu (requisitos não
	// atendidos, vaga indisponível etc.).
	Motivos       []string   `json:"motivos,omitempty"`
	IDCandidatura *uuid.UUID `json:"id_candidatura,omitempty"`
}

// AssociacaoRecusadaResponse é a resposta 422: o currículo não tem o mínimo
// para se candidatar, e Pendencias diz o que falta.
type AssociacaoRecusadaResponse struct {
	Error      string   `json:"error"`
	Pendencias []string `json:"pendencias"`
}

// AssociarVagasResponse traz o resultado de cada vaga, na ordem pedida.
type AssociarVagasResponse struct {
	Resultados []*AssociacaoVaga `json:"resultados"`
}
