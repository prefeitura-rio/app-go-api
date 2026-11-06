package empregabilidade

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"gorm.io/gorm"
)

type StatusVaga string

const (
	StatusVagaEmEdicao          StatusVaga = "em_edicao"
	StatusVagaEmAprovacao       StatusVaga = "em_aprovacao"
	StatusVagaPublicadoAtivo    StatusVaga = "publicado_ativo"
	StatusVagaPublicadoExpirado StatusVaga = "publicado_expirado"
	StatusVagaCongelada         StatusVaga = "vaga_congelada"
	StatusVagaDescontinuada     StatusVaga = "vaga_descontinuada"
)

func (s StatusVaga) IsValid() bool {
	switch s {
	case StatusVagaEmEdicao, StatusVagaEmAprovacao, StatusVagaPublicadoAtivo, StatusVagaPublicadoExpirado,
		StatusVagaCongelada, StatusVagaDescontinuada:
		return true
	}
	return false
}

// IsPublished reports whether the vaga has already been published (in any of the
// post-publication states). Used to gate who may edit an already-published vaga.
func (s StatusVaga) IsPublished() bool {
	switch s {
	case StatusVagaPublicadoAtivo, StatusVagaPublicadoExpirado,
		StatusVagaCongelada, StatusVagaDescontinuada:
		return true
	}
	return false
}

type AcessibilidadePCD string

const (
	AcessibilidadeParaPCD         AcessibilidadePCD = "para_pcd"
	AcessibilidadePreferencialPCD AcessibilidadePCD = "preferencial_pcd"
	AcessibilidadeExclusivoPCD    AcessibilidadePCD = "exclusivo_pcd"
)

func (a AcessibilidadePCD) IsValid() bool {
	switch a {
	case AcessibilidadeParaPCD, AcessibilidadePreferencialPCD, AcessibilidadeExclusivoPCD:
		return true
	}
	return false
}

// CriterioBloqueio identifica um critério de elegibilidade da vaga cujo
// bloqueio de candidatura pode ser ligado ou desligado. A vaga guarda só os
// critérios ligados (CriteriosBloqueio); o que não está na lista está
// desligado, que é o padrão. Critério novo (PcD, idade, experiência) entra
// como uma constante a mais e uma entrada em criteriosBloqueioConhecidos.
type CriterioBloqueio string

const (
	// CriterioBloqueioEscolaridade é a escolaridade mínima (formação).
	CriterioBloqueioEscolaridade CriterioBloqueio = "escolaridade"
	// CriterioBloqueioIdioma são os requisitos de idioma.
	CriterioBloqueioIdioma CriterioBloqueio = "idioma"
)

// criteriosBloqueioConhecidos são os critérios aceitos, na ordem em que são
// gravados e devolvidos.
var criteriosBloqueioConhecidos = []CriterioBloqueio{
	CriterioBloqueioEscolaridade,
	CriterioBloqueioIdioma,
}

func (c CriterioBloqueio) IsValid() bool {
	for _, conhecido := range criteriosBloqueioConhecidos {
		if c == conhecido {
			return true
		}
	}
	return false
}

// NormalizarCriteriosBloqueio confere as chaves, tira as repetidas e põe na
// ordem de criteriosBloqueioConhecidos. Sem nenhum critério ligado devolve
// nil, para que o desligado nunca seja gravado na vaga.
func NormalizarCriteriosBloqueio(criterios []CriterioBloqueio) ([]CriterioBloqueio, error) {
	ligados := make(map[CriterioBloqueio]bool, len(criterios))
	for _, c := range criterios {
		if !c.IsValid() {
			return nil, fmt.Errorf("critério de bloqueio desconhecido: %q", string(c))
		}
		ligados[c] = true
	}
	if len(ligados) == 0 {
		return nil, nil
	}
	normalizados := make([]CriterioBloqueio, 0, len(ligados))
	for _, c := range criteriosBloqueioConhecidos {
		if ligados[c] {
			normalizados = append(normalizados, c)
		}
	}
	return normalizados, nil
}

type Vaga struct {
	ID                             uuid.UUID          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Titulo                         string             `json:"titulo" gorm:"type:varchar(500);not null"`
	Descricao                      string             `json:"descricao" gorm:"type:text;not null"`
	IDContratante                  string             `json:"id_contratante" gorm:"type:varchar(18);not null"`
	IDRegimeContratacao            uuid.UUID          `json:"id_regime_contratacao" gorm:"type:uuid;not null"`
	IDModeloTrabalho               uuid.UUID          `json:"id_modelo_trabalho" gorm:"type:uuid;not null"`
	AcessibilidadePCD              *AcessibilidadePCD `json:"acessibilidade_pcd,omitempty" gorm:"type:acessibilidade_pcd_enum"`
	ValorVaga                      *float64           `json:"valor_vaga" gorm:"type:decimal(12,2)" extensions:"x-nullable"`
	QuantidadeEstimadaContratacoes *int               `json:"quantidade_estimada_contratacoes,omitempty" gorm:"type:integer"`
	Bairro                         string             `json:"bairro" gorm:"type:varchar(255)"`
	DataLimite                     *time.Time         `json:"data_limite" gorm:"type:timestamp with time zone" extensions:"x-nullable"`
	Requisitos                     string             `json:"requisitos" gorm:"type:text"`
	Diferenciais                   string             `json:"diferenciais" gorm:"type:text"`
	Responsabilidades              string             `json:"responsabilidades" gorm:"type:text"`
	Beneficios                     string             `json:"beneficios" gorm:"type:text"`
	IDOrgaoParceiro                string             `json:"id_orgao_parceiro" gorm:"type:varchar(50)"`
	Status                         StatusVaga         `json:"status" gorm:"type:varchar(50);not null;default:'em_edicao'"`
	Slug                           string             `json:"slug" gorm:"-"`
	DeletedAt                      gorm.DeletedAt     `json:"deleted_at,omitempty" gorm:"index" swaggertype:"string" example:"2024-12-31T23:59:59Z" extensions:"x-nullable"`
	CreatedAt                      time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                      time.Time          `json:"updated_at" gorm:"autoUpdateTime"`

	// Critérios de elegibilidade
	IdadeMinima                *int       `json:"idade_minima,omitempty" gorm:"type:integer"`
	IdadeMaxima                *int       `json:"idade_maxima,omitempty" gorm:"type:integer"`
	BairrosElegibilidade       []string   `json:"bairros_elegibilidade,omitempty" gorm:"type:jsonb;serializer:json"`
	IDEscolaridadeMinima       *uuid.UUID `json:"id_escolaridade_minima,omitempty" gorm:"type:uuid"`
	AreasFormacaoElegibilidade []string   `json:"areas_formacao_elegibilidade,omitempty" gorm:"type:jsonb;serializer:json"`
	// CriteriosBloqueio lista os critérios com bloqueio de candidatura ligado;
	// os ausentes estão desligados. Sempre devolvido ([] quando nenhum). No
	// PUT, o campo ausente mantém o estado gravado e [] desliga todos.
	CriteriosBloqueio []CriterioBloqueio `json:"criterios_bloqueio" gorm:"type:jsonb;serializer:json"`

	// Relationships
	Contratante               *Empresa                 `json:"contratante,omitempty" gorm:"foreignKey:IDContratante;references:CNPJ"`
	RegimeContratacao         *RegimeContratacao       `json:"regime_contratacao,omitempty" gorm:"foreignKey:IDRegimeContratacao"`
	ModeloTrabalho            *ModeloTrabalho          `json:"modelo_trabalho,omitempty" gorm:"foreignKey:IDModeloTrabalho"`
	OrgaoParceiro             *models.OrgaoSnapshot    `json:"orgao_parceiro,omitempty" gorm:"foreignKey:IDOrgaoParceiro;references:OrgaoID"`
	TiposPCD                  []TipoPCD                `json:"tipos_pcd,omitempty" gorm:"many2many:emp_vagas_tipos_pcd;foreignKey:ID;joinForeignKey:id_vaga;References:ID;joinReferences:id_tipo_pcd"`
	Zonas                     []Zona                   `json:"zonas,omitempty" gorm:"many2many:emp_vagas_zonas;foreignKey:ID;joinForeignKey:id_vaga;References:ID;joinReferences:id_zona"`
	IdiomasRequisito          []VagaIdiomaRequisito    `json:"idiomas_requisito,omitempty" gorm:"foreignKey:IDVaga"`
	Etapas                    []Etapa                  `json:"etapas,omitempty" gorm:"foreignKey:IDVaga"`
	InformacoesComplementares []InformacaoComplementar `json:"informacoes_complementares,omitempty" gorm:"foreignKey:IDVaga"`
	EscolaridadeMinima        *Escolaridade            `json:"escolaridade_minima,omitempty" gorm:"foreignKey:IDEscolaridadeMinima"`
}

func (Vaga) TableName() string {
	return "emp_vagas"
}

func BuildVagaSlug(titulo string, id uuid.UUID) string {
	return slug.Make(titulo) + "-" + strings.ReplaceAll(id.String(), "-", "")[:12]
}

func (v *Vaga) AfterFind(_ *gorm.DB) error {
	v.Slug = BuildVagaSlug(v.Titulo, v.ID)
	if v.CriteriosBloqueio == nil {
		v.CriteriosBloqueio = []CriterioBloqueio{}
	}
	return nil
}

type VagaFilter struct {
	Status              string
	Contratante         string
	OrgaoParceiroID     string
	OrgaoParceiroIDs    []string
	Search              string
	IDRegimeContratacao []string
	IDModeloTrabalho    []string
	AcessibilidadePCD   []string
	Bairro              []string
	DataPublicacao      DataPublicacaoRange
}

type DataPublicacaoRange string

const (
	DataPublicacaoHoje         DataPublicacaoRange = "hoje"
	DataPublicacaoUltimaSemana DataPublicacaoRange = "ultima_semana"
	DataPublicacaoUltimoMes    DataPublicacaoRange = "ultimo_mes"
)

func (d DataPublicacaoRange) IsValid() bool {
	switch d {
	case DataPublicacaoHoje, DataPublicacaoUltimaSemana, DataPublicacaoUltimoMes:
		return true
	}
	return false
}

type VagaPublicFilter struct {
	Status              string
	DataPublicacao      DataPublicacaoRange
	IDRegimeContratacao []string
	IDModeloTrabalho    []string
	Contratante         []string
	AcessibilidadePCD   []string
	Bairro              []string
}

func (v *Vaga) UpdateStatusBasedOnExpiration() {
	if v.DataLimite == nil {
		return
	}

	now := time.Now()

	if now.After(*v.DataLimite) {
		if v.Status == StatusVagaPublicadoAtivo {
			v.Status = StatusVagaPublicadoExpirado
		}
	} else {
		if v.Status == StatusVagaPublicadoExpirado {
			v.Status = StatusVagaPublicadoAtivo
		}
	}
}
