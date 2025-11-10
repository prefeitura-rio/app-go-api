package models

import (
	"fmt"
	"strings"
)

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

// CursoSearchParameters contém parâmetros específicos para busca de cursos
type CursoSearchParameters struct {
	Q               string `json:"q"`                                // Termo de busca
	OrgaoID         string `json:"orgao_id,omitempty"`              // ID do órgão (external reference)
	InstituicaoID   int    `json:"instituicao_id,omitempty"`        // ID da instituição
	Status          string `json:"status,omitempty"`                // Status do curso
	Modalidade      string `json:"modalidade,omitempty"`            // Modalidade
	Turno           string `json:"turno,omitempty"`                 // Turno
	FormatoAula     string `json:"formato_aula,omitempty"`          // Formato da aula
	Page            int    `json:"page,omitempty"`                  // Página
	PerPage         int    `json:"per_page,omitempty"`              // Itens por página
	SortBy          string `json:"sort_by,omitempty"`               // Campo para ordenação
	IncludeFields   string `json:"include_fields,omitempty"`        // Campos a incluir
	ExcludeFields   string `json:"exclude_fields,omitempty"`        // Campos a excluir
}

// ToSearchParameters converte CursoSearchParameters para SearchParameters
func (c *CursoSearchParameters) ToSearchParameters() SearchParameters {
	// Campos do record para busca textual
	queryBy := "record.titulo,record.descricao,record.pre_requisitos"
	
	// Construir filtros
	var filters []string
	
	// Filtrar apenas registros com action = 'read'
	filters = append(filters, "action:=read")

	if c.OrgaoID != "" {
		filters = append(filters, fmt.Sprintf("record.orgao_id:=%s", c.OrgaoID))
	}
	
	if c.InstituicaoID > 0 {
		filters = append(filters, fmt.Sprintf("record.instituicao_id:=%d", c.InstituicaoID))
	}
	
	if c.Status != "" {
		filters = append(filters, fmt.Sprintf("record.status:=%s", c.Status))
	}
	
	if c.Modalidade != "" {
		filters = append(filters, fmt.Sprintf("record.modalidade:=%s", c.Modalidade))
	}
	
	if c.Turno != "" {
		filters = append(filters, fmt.Sprintf("record.turno:=%s", c.Turno))
	}
	
	if c.FormatoAula != "" {
		filters = append(filters, fmt.Sprintf("record.formato_aula:=%s", c.FormatoAula))
	}
	
	filterBy := ""
	if len(filters) > 0 {
		filterBy = strings.Join(filters, " && ")
	}
	
	// Ordenação padrão por data de criação mais recente
	sortBy := c.SortBy
	if sortBy == "" {
		sortBy = "record.created_at:desc"
	}
	
	return SearchParameters{
		Q:             c.Q,
		QueryBy:       queryBy,
		FilterBy:      filterBy,
		SortBy:        sortBy,
		Page:          c.Page,
		PerPage:       c.PerPage,
		IncludeFields: c.IncludeFields,
		ExcludeFields: c.ExcludeFields,
		Prefix:        true,
		NumTypos:      2,
	}
}

// EmpregoSearchParameters contém parâmetros específicos para busca de empregos
type EmpregoSearchParameters struct {
	Q                  string `json:"q"`                                   // Termo de busca
	OrgaoID            string `json:"orgao_id,omitempty"`                 // ID do órgão (external reference)
	EmpresaID          int    `json:"empresa_id,omitempty"`               // ID da empresa
	EscolaridadeID     int    `json:"escolaridade_id,omitempty"`          // ID da escolaridade
	Status             string `json:"status,omitempty"`                   // Status do emprego
	TipoContratacao    string `json:"tipo_contratacao,omitempty"`         // Tipo de contratação
	JornadaTrabalho    string `json:"jornada_trabalho,omitempty"`         // Jornada de trabalho
	Turno              string `json:"turno,omitempty"`                    // Turno
	SalarioMin         int    `json:"salario_min,omitempty"`              // Salário mínimo
	SalarioMax         int    `json:"salario_max,omitempty"`              // Salário máximo
	Page               int    `json:"page,omitempty"`                     // Página
	PerPage            int    `json:"per_page,omitempty"`                 // Itens por página
	SortBy             string `json:"sort_by,omitempty"`                  // Campo para ordenação
	IncludeFields      string `json:"include_fields,omitempty"`           // Campos a incluir
	ExcludeFields      string `json:"exclude_fields,omitempty"`           // Campos a excluir
}

// ToSearchParameters converte EmpregoSearchParameters para SearchParameters
func (e *EmpregoSearchParameters) ToSearchParameters() SearchParameters {
	// Campos do record para busca textual
	queryBy := "record.titulo,record.descricao,record.pre_requisitos,record.beneficios"
	
	// Construir filtros
	var filters []string
	
	// Filtrar apenas registros com action = 'read'
	filters = append(filters, "action:=read")

	if e.OrgaoID != "" {
		filters = append(filters, fmt.Sprintf("record.orgao_id:=%s", e.OrgaoID))
	}
	
	if e.EmpresaID > 0 {
		filters = append(filters, fmt.Sprintf("record.empresa_id:=%d", e.EmpresaID))
	}
	
	if e.EscolaridadeID > 0 {
		filters = append(filters, fmt.Sprintf("record.escolaridade_id:=%d", e.EscolaridadeID))
	}
	
	if e.Status != "" {
		filters = append(filters, fmt.Sprintf("record.status:=%s", e.Status))
	}
	
	if e.TipoContratacao != "" {
		filters = append(filters, fmt.Sprintf("record.tipo_contratacao:=%s", e.TipoContratacao))
	}
	
	if e.JornadaTrabalho != "" {
		filters = append(filters, fmt.Sprintf("record.jornada_trabalho:=%s", e.JornadaTrabalho))
	}
	
	if e.Turno != "" {
		filters = append(filters, fmt.Sprintf("record.turno:=%s", e.Turno))
	}
	
	// Filtros por faixa salarial
	if e.SalarioMin > 0 {
		filters = append(filters, fmt.Sprintf("record.salario_min:>=%d", e.SalarioMin))
	}
	
	if e.SalarioMax > 0 {
		filters = append(filters, fmt.Sprintf("record.salario_max:<=%d", e.SalarioMax))
	}
	
	filterBy := ""
	if len(filters) > 0 {
		filterBy = strings.Join(filters, " && ")
	}
	
	// Ordenação padrão por data de criação mais recente
	sortBy := e.SortBy
	if sortBy == "" {
		sortBy = "record.created_at:desc"
	}
	
	return SearchParameters{
		Q:             e.Q,
		QueryBy:       queryBy,
		FilterBy:      filterBy,
		SortBy:        sortBy,
		Page:          e.Page,
		PerPage:       e.PerPage,
		IncludeFields: e.IncludeFields,
		ExcludeFields: e.ExcludeFields,
		Prefix:        true,
		NumTypos:      2,
	}
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