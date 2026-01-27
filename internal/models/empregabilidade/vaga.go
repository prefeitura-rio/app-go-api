package empregabilidade

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatusVaga string

const (
	StatusVagaEmEdicao          StatusVaga = "em_edicao"
	StatusVagaEmAprovacao       StatusVaga = "em_aprovacao"
	StatusVagaPublicadoAtivo    StatusVaga = "publicado_ativo"
	StatusVagaPublicadoExpirado StatusVaga = "publicado_expirado"
)

func (s StatusVaga) IsValid() bool {
	switch s {
	case StatusVagaEmEdicao, StatusVagaEmAprovacao, StatusVagaPublicadoAtivo, StatusVagaPublicadoExpirado:
		return true
	}
	return false
}

type Vaga struct {
	ID                   uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Titulo               string         `json:"titulo" gorm:"type:varchar(500);not null"`
	Descricao            string         `json:"descricao" gorm:"type:text;not null"`
	IDContratante        string         `json:"id_contratante" gorm:"type:varchar(18);not null"`
	IDRegimeContratacao  uuid.UUID      `json:"id_regime_contratacao" gorm:"type:uuid;not null"`
	IDModeloTrabalho     uuid.UUID      `json:"id_modelo_trabalho" gorm:"type:uuid;not null"`
	VagaPCD              bool           `json:"vaga_pcd" gorm:"default:false"`
	ValorVaga            *float64       `json:"valor_vaga" gorm:"type:decimal(12,2)"`
	Bairro               string         `json:"bairro" gorm:"type:varchar(255)"`
	DataLimite           *time.Time     `json:"data_limite" gorm:"type:timestamp with time zone"`
	Requisitos           string         `json:"requisitos" gorm:"type:text"`
	Diferenciais         string         `json:"diferenciais" gorm:"type:text"`
	Responsabilidades    string         `json:"responsabilidades" gorm:"type:text"`
	Beneficios           string         `json:"beneficios" gorm:"type:text"`
	IDOrgaoParceiro      string         `json:"id_orgao_parceiro" gorm:"type:varchar(50)"`
	Status               StatusVaga     `json:"status" gorm:"type:varchar(50);not null;default:'em_edicao'"`
	DeletedAt            gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	CreatedAt            time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Contratante       *Empresa           `json:"contratante,omitempty" gorm:"foreignKey:IDContratante;references:CNPJ"`
	RegimeContratacao *RegimeContratacao `json:"regime_contratacao,omitempty" gorm:"foreignKey:IDRegimeContratacao"`
	ModeloTrabalho    *ModeloTrabalho    `json:"modelo_trabalho,omitempty" gorm:"foreignKey:IDModeloTrabalho"`
	TiposPCD          []TipoPCD          `json:"tipos_pcd,omitempty" gorm:"many2many:emp_vagas_tipos_pcd;foreignKey:ID;joinForeignKey:id_vaga;References:ID;joinReferences:id_tipo_pcd"`
	Etapas            []Etapa            `json:"etapas,omitempty" gorm:"foreignKey:IDVaga"`
	InformacoesComplementares []InformacaoComplementar `json:"informacoes_complementares,omitempty" gorm:"foreignKey:IDVaga"`
}

func (Vaga) TableName() string {
	return "emp_vagas"
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
