package empregabilidade

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// TamanhoMaximoTipoEtapa é o tamanho da coluna emp_etapas.tipo_etapa.
const TamanhoMaximoTipoEtapa = 64

// ErrTipoEtapaLongo indica tipo_etapa acima de TamanhoMaximoTipoEtapa caracteres.
var ErrTipoEtapaLongo = fmt.Errorf("tipo_etapa deve ter no máximo %d caracteres", TamanhoMaximoTipoEtapa)

// NormalizarTipoEtapa tira os espaços das pontas do tipo_etapa; vazio vira nil
// (o valor gravado não muda) e acima do tamanho da coluna é recusado.
func NormalizarTipoEtapa(tipo *string) (*string, error) {
	if tipo == nil {
		return nil, nil
	}
	valor := strings.TrimSpace(*tipo)
	if valor == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(valor) > TamanhoMaximoTipoEtapa {
		return nil, ErrTipoEtapaLongo
	}
	return &valor, nil
}

type Etapa struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	IDVaga    uuid.UUID `json:"id_vaga" gorm:"type:uuid;not null"`
	Titulo    string    `json:"titulo" gorm:"type:varchar(500);not null"`
	Descricao string    `json:"descricao" gorm:"type:text"`
	Ordem     int       `json:"ordem" gorm:"not null"`
	// TipoEtapa classifica a etapa (código livre definido pela gestão) para
	// comparar o funil do histórico de candidaturas entre vagas. Opcional:
	// ausente (ou vazio) numa atualização mantém o valor gravado.
	TipoEtapa *string   `json:"tipo_etapa,omitempty" gorm:"type:varchar(64)"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Vaga *Vaga `json:"vaga,omitempty" gorm:"foreignKey:IDVaga"`
}

func (Etapa) TableName() string {
	return "emp_etapas"
}
