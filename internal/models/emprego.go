package models

import (
	"time"
)

type TipoContratacao string
type StatusEmprego string
type JornadaTrabalho string

const (
	TipoContratacaoCLT         TipoContratacao = "CLT"
	TipoContratacaoEstagio     TipoContratacao = "ESTAGIO"
	TipoContratacaoJovemAprendiz TipoContratacao = "JOVEM_APRENDIZ"
	TipoContratacaoMEI         TipoContratacao = "MEI"
	TipoContratacaoPJ          TipoContratacao = "PJ"

	StatusEmpregoCriado    StatusEmprego = "CRIADO"
	StatusEmpregoAberto    StatusEmprego = "ABERTO"
	StatusEmpregoEmAnalise StatusEmprego = "EM_ANALISE"
	StatusEmpregoEncerrado StatusEmprego = "ENCERRADO"

	JornadaIntegral    JornadaTrabalho = "INTEGRAL"
	JornadaMeioPeriodo JornadaTrabalho = "MEIO_PERIODO"
	JornadaFlexivel    JornadaTrabalho = "FLEXIVEL"
	JornadaEstagio     JornadaTrabalho = "ESTAGIO"
	JornadaHomeOffice  JornadaTrabalho = "HOME_OFFICE"
)

type Emprego struct {
	ID                   int             `json:"id" gorm:"primaryKey"`
	Titulo               string          `json:"titulo" gorm:"not null"`
	Descricao            string          `json:"descricao" gorm:"type:text"`
	OrgaoID              int             `json:"orgao_id" gorm:"column:orgao_id"`
	EmpresaID            int             `json:"empresa_id" gorm:"column:empresa_id"`
	TipoContratacao      TipoContratacao `json:"tipo_contratacao" gorm:"type:tipo_contratacao_enum"`
	NumeroVagas          int             `json:"numero_vagas" gorm:"column:numero_vagas"`
	Latitude             float64         `json:"latitude" gorm:"type:decimal(10,8)"`
	Longitude            float64         `json:"longitude" gorm:"type:decimal(11,8)"`
	SalarioMin           int             `json:"salario_min" gorm:"column:salario_min"`
	SalarioMax           int             `json:"salario_max" gorm:"column:salario_max"`
	Beneficios           string          `json:"beneficios" gorm:"type:text"`
	PreRequisitos        string          `json:"pre_requisitos" gorm:"column:pre_requisitos;type:text"`
	DataInicioPrevista   time.Time       `json:"data_inicio_prevista" gorm:"column:data_inicio_prevista"`
	DataLimiteCandidatura time.Time      `json:"data_limite_candidatura" gorm:"column:data_limite_candidatura"`
	Status               StatusEmprego   `json:"status" gorm:"type:status_emprego_enum"`
	EscolaridadeID       int             `json:"escolaridade_id" gorm:"column:escolaridade_id"`
	JornadaTrabalho      JornadaTrabalho `json:"jornada_trabalho" gorm:"column:jornada_trabalho;type:jornada_trabalho_enum"`
	Turno                Turno           `json:"turno" gorm:"type:turno_enum"`
	ContatoDuvidas       string          `json:"contato_duvidas" gorm:"column:contato_duvidas"`
	CreatedAt            time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relacionamentos
	Orgao                *Orgao          `json:"orgao,omitempty" gorm:"foreignKey:OrgaoID"`
	Empresa              *Empresa        `json:"empresa,omitempty" gorm:"foreignKey:EmpresaID"`
	Escolaridade         *Escolaridade   `json:"escolaridade,omitempty" gorm:"foreignKey:EscolaridadeID"`
}

// TableName especifica o nome da tabela para este modelo
func (Emprego) TableName() string {
	return "empregos"
} 