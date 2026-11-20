package empregabilidade

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Classificações de tipo_etapa (PREF-687): comparam o funil do histórico de
// candidaturas entre vagas cujas etapas têm títulos livres diferentes.
const (
	TipoEtapaAnaliseCurriculo = "analise_curriculo"
	TipoEtapaEntrevista       = "entrevista"
	TipoEtapaTeste            = "teste"
	TipoEtapaProposta         = "proposta"
	TipoEtapaOutro            = "outro"
)

// TiposEtapa são os valores aceitos em emp_etapas.tipo_etapa, na ordem exibida.
var TiposEtapa = []string{
	TipoEtapaAnaliseCurriculo,
	TipoEtapaEntrevista,
	TipoEtapaTeste,
	TipoEtapaProposta,
	TipoEtapaOutro,
}

// ErrTipoEtapaInvalido indica tipo_etapa fora de TiposEtapa.
var ErrTipoEtapaInvalido = fmt.Errorf("tipo_etapa deve ser um de: %s", strings.Join(TiposEtapa, ", "))

// RegraTipoEtapa associa um padrão do título da etapa a um tipo.
type RegraTipoEtapa struct {
	Tipo   string
	Padrao string
}

// RegrasTipoEtapaPorTitulo classificam o título livre, na ordem: a primeira que
// casa vence; nenhuma, TipoEtapaOutro. Cada padrão casa, sem diferenciar
// maiúsculas, no começo de uma palavra do título ("Prova" casa, "Aprovação"
// não). Os padrões são os mesmos da migration
// 20261120160000_classifica_tipo_etapa_emp_etapas.sql, que os aplica às etapas
// existentes com ~* '\m(<padrão>)' no PostgreSQL, por isso as letras
// acentuadas aparecem nas duas caixas.
var RegrasTipoEtapaPorTitulo = []RegraTipoEtapa{
	{TipoEtapaAnaliseCurriculo, "curr[iíÍ]cul|triagem"},
	{TipoEtapaEntrevista, "entrevista"},
	{TipoEtapaTeste, "teste|prova|avalia|din[aâÂ]mica"},
	{TipoEtapaProposta, "proposta|oferta|contrata|admiss"},
}

var regexesTipoEtapa = compilarRegrasTipoEtapa()

func compilarRegrasTipoEtapa() []*regexp.Regexp {
	regexes := make([]*regexp.Regexp, len(RegrasTipoEtapaPorTitulo))
	for i, regra := range RegrasTipoEtapaPorTitulo {
		// Começo de palavra como o \m do PostgreSQL: o início do título ou um
		// caractere que não é letra, dígito nem "_" (o \b do Go só vê ASCII).
		regexes[i] = regexp.MustCompile(`(?i)(?:^|[^\p{L}\p{N}_])(?:` + regra.Padrao + `)`)
	}
	return regexes
}

// InferirTipoEtapa classifica a etapa pelo título livre (RegrasTipoEtapaPorTitulo).
func InferirTipoEtapa(titulo string) string {
	for i, re := range regexesTipoEtapa {
		if re.MatchString(titulo) {
			return RegrasTipoEtapaPorTitulo[i].Tipo
		}
	}
	return TipoEtapaOutro
}

// NormalizarTipoEtapa tira os espaços das pontas e passa o tipo_etapa para
// minúsculas; vazio vira nil (o valor gravado não muda) e valor fora de
// TiposEtapa é recusado com ErrTipoEtapaInvalido.
func NormalizarTipoEtapa(tipo *string) (*string, error) {
	if tipo == nil {
		return nil, nil
	}
	valor := strings.ToLower(strings.TrimSpace(*tipo))
	if valor == "" {
		return nil, nil
	}
	if !slices.Contains(TiposEtapa, valor) {
		return nil, ErrTipoEtapaInvalido
	}
	return &valor, nil
}

type Etapa struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	IDVaga    uuid.UUID `json:"id_vaga" gorm:"type:uuid;not null"`
	Titulo    string    `json:"titulo" gorm:"type:varchar(500);not null"`
	Descricao string    `json:"descricao" gorm:"type:text"`
	Ordem     int       `json:"ordem" gorm:"not null"`
	// TipoEtapa classifica a etapa (analise_curriculo, entrevista, teste,
	// proposta ou outro) para comparar o funil do histórico de candidaturas
	// entre vagas; o título continua livre. Ausente (ou vazio) na criação é
	// inferido do título; numa atualização, mantém o valor gravado.
	TipoEtapa *string   `json:"tipo_etapa,omitempty" gorm:"type:varchar(64)" enums:"analise_curriculo,entrevista,teste,proposta,outro"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Vaga *Vaga `json:"vaga,omitempty" gorm:"foreignKey:IDVaga"`
}

func (Etapa) TableName() string {
	return "emp_etapas"
}

// BeforeCreate infere o tipo_etapa do título quando a etapa nova chega sem ele.
func (e *Etapa) BeforeCreate(_ *gorm.DB) error {
	if e.TipoEtapa == nil {
		tipo := InferirTipoEtapa(e.Titulo)
		e.TipoEtapa = &tipo
	}
	return nil
}
