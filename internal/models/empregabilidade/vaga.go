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

// AcessibilidadePCD é o campo antigo "Acessibilidade PcD", descontinuado
// (DEV-407): a opção PcD da vaga é OpcaoPCD. A coluna continua preenchida a
// partir de OpcaoPCD para quem ainda lê ou filtra por ela.
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

// OpcaoPCD diz se a vaga é aberta a todos, preferencial ou exclusiva para
// pessoas com deficiência (história PREF-632). Os tipos de deficiência aceitos
// ficam em Vaga.TiposPCD e o bloqueio de candidatura, em CriteriosBloqueio
// (chave CriterioBloqueioPCD).
type OpcaoPCD string

const (
	OpcaoPCDAberta       OpcaoPCD = "aberta"
	OpcaoPCDPreferencial OpcaoPCD = "preferencial"
	OpcaoPCDExclusiva    OpcaoPCD = "exclusiva"
)

// MsgTipoPCDObrigatorio é o erro da vaga exclusiva sem tipo de deficiência.
const MsgTipoPCDObrigatorio = "Selecione ao menos um tipo de deficiência."

func (o OpcaoPCD) IsValid() bool {
	switch o {
	case OpcaoPCDAberta, OpcaoPCDPreferencial, OpcaoPCDExclusiva:
		return true
	}
	return false
}

// OpcaoPCDDeAcessibilidade é o de/para do campo antigo: sem acessibilidade
// ("Vaga PcD" desmarcada) é aberta; preferencial_pcd é preferencial;
// exclusivo_pcd é exclusiva; para_pcd ("Vaga PcD" marcada sem modalidade) é
// preferencial, que não bloqueia ninguém. A migração das vagas antigas usa a
// mesma regra.
func OpcaoPCDDeAcessibilidade(a *AcessibilidadePCD) OpcaoPCD {
	if a == nil {
		return OpcaoPCDAberta
	}
	if *a == AcessibilidadeExclusivoPCD {
		return OpcaoPCDExclusiva
	}
	return OpcaoPCDPreferencial
}

// AcessibilidadeLegada é o valor do campo antigo que corresponde à opção,
// gravado junto para os clientes que ainda o leem (nil para aberta).
func (o OpcaoPCD) AcessibilidadeLegada() *AcessibilidadePCD {
	var a AcessibilidadePCD
	switch o {
	case OpcaoPCDPreferencial:
		a = AcessibilidadePreferencialPCD
	case OpcaoPCDExclusiva:
		a = AcessibilidadeExclusivoPCD
	default:
		return nil
	}
	return &a
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
	// CriterioBloqueioPCD é o critério PcD (vaga exclusiva). "pcd" é também o
	// slug do critério em CandidaturaBloqueio.CriteriosNaoAtendidos.
	CriterioBloqueioPCD CriterioBloqueio = "pcd"
	// CriterioBloqueioExperienciaCarteira é a experiência comprovada em
	// carteira (DEV-424), com o tempo exigido em Vaga.ExperienciaCarteira.
	// "experiencia_carteira" é também o slug do critério em
	// CandidaturaBloqueio.CriteriosNaoAtendidos.
	CriterioBloqueioExperienciaCarteira CriterioBloqueio = "experiencia_carteira"
)

// criteriosBloqueioConhecidos são os critérios aceitos, na ordem em que são
// gravados e devolvidos.
var criteriosBloqueioConhecidos = []CriterioBloqueio{
	CriterioBloqueioEscolaridade,
	CriterioBloqueioIdioma,
	CriterioBloqueioPCD,
	CriterioBloqueioExperienciaCarteira,
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

// ExperienciaCarteira é o tempo de experiência comprovada em carteira que a
// vaga exige (DEV-424, história PREF-633): qualquer tempo ou o mínimo fixo de
// 6 meses. O bloqueio de candidatura é a chave
// CriterioBloqueioExperienciaCarteira de Vaga.CriteriosBloqueio.
type ExperienciaCarteira string

const (
	// ExperienciaCarteiraQualquerTempo pede ao menos uma experiência
	// comprovada em carteira, de qualquer duração.
	ExperienciaCarteiraQualquerTempo ExperienciaCarteira = "qualquer_tempo"
	// ExperienciaCarteiraMinimo6Meses pede MesesMinimosExperienciaCarteira
	// somados entre as experiências comprovadas em carteira.
	ExperienciaCarteiraMinimo6Meses ExperienciaCarteira = "minimo_6_meses"
)

// MesesMinimosExperienciaCarteira é o prazo fixo da opção "mínimo de 6 meses".
const MesesMinimosExperienciaCarteira = 6

// MsgExperienciaCarteiraObrigatoria é o erro do bloqueio de experiência em
// carteira ligado sem o tempo exigido.
const MsgExperienciaCarteiraObrigatoria = "Selecione o tempo de experiência em carteira exigido (qualquer_tempo ou minimo_6_meses)."

func (e ExperienciaCarteira) IsValid() bool {
	return e == ExperienciaCarteiraQualquerTempo || e == ExperienciaCarteiraMinimo6Meses
}

// Atende diz se as experiências do currículo cumprem a exigência. Só contam
// as comprovadas em carteira (experiencia_comprovada_ct): qualquer tempo pede
// ao menos uma; o mínimo de 6 meses soma tempo_experiencia_meses entre elas
// (sem tempo informado, ou negativo, conta zero). É a mesma regra do filtro
// carteira do banco de currículos (sim / 6_meses).
func (e ExperienciaCarteira) Atende(experiencias []*CurriculoExperiencia) bool {
	algumaEmCarteira := false
	totalMeses := 0
	for _, exp := range experiencias {
		if exp == nil || !exp.ExperienciaComprovadaCT {
			continue
		}
		algumaEmCarteira = true
		if exp.TempoExperienciaMeses != nil && *exp.TempoExperienciaMeses > 0 {
			totalMeses += *exp.TempoExperienciaMeses
		}
	}
	switch e {
	case ExperienciaCarteiraQualquerTempo:
		return algumaEmCarteira
	case ExperienciaCarteiraMinimo6Meses:
		return totalMeses >= MesesMinimosExperienciaCarteira
	}
	return false
}

// Requisito é o texto do requisito, como aparece na lista do que o candidato
// não atende.
func (e ExperienciaCarteira) Requisito() string {
	if e == ExperienciaCarteiraMinimo6Meses {
		return fmt.Sprintf("Experiência comprovada em carteira — mínimo de %d meses", MesesMinimosExperienciaCarteira)
	}
	return "Experiência comprovada em carteira"
}

type Vaga struct {
	ID                             uuid.UUID          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Titulo                         string             `json:"titulo" gorm:"type:varchar(500);not null"`
	Descricao                      string             `json:"descricao" gorm:"type:text;not null"`
	IDContratante                  string             `json:"id_contratante" gorm:"type:varchar(18);not null"`
	IDRegimeContratacao            uuid.UUID          `json:"id_regime_contratacao" gorm:"type:uuid;not null"`
	IDModeloTrabalho               uuid.UUID          `json:"id_modelo_trabalho" gorm:"type:uuid;not null"`
	AcessibilidadePCD              *AcessibilidadePCD `json:"acessibilidade_pcd,omitempty" gorm:"type:acessibilidade_pcd_enum"`
	OpcaoPCD                       OpcaoPCD           `json:"opcao_pcd" gorm:"type:varchar(20);not null"`
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
	// ExperienciaCarteira é o tempo de experiência em carteira exigido
	// (qualquer_tempo ou minimo_6_meses); null = sem exigência. Sempre
	// devolvido. Com a chave experiencia_carteira em criterios_bloqueio é
	// obrigatório. No PUT, o campo ausente (ou null) mantém o valor gravado e
	// "" remove a exigência.
	ExperienciaCarteira *ExperienciaCarteira `json:"experiencia_carteira" gorm:"type:varchar(20)" swaggertype:"string" enums:"qualquer_tempo,minimo_6_meses" extensions:"x-nullable"`
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
	if v.OpcaoPCD == "" {
		v.OpcaoPCD = OpcaoPCDAberta
	}
	return nil
}

func (v *Vaga) BeforeCreate(_ *gorm.DB) error {
	if v.OpcaoPCD == "" {
		v.OpcaoPCD = OpcaoPCDAberta
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
