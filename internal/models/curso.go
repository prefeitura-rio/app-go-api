package models

import (
	"time"
)

type Modalidade string
type StatusCurso string
type FormatoAula string

const (
	ModalidadePresencial Modalidade = "PRESENCIAL"
	ModalidadeOnline     Modalidade = "ONLINE"
	ModalidadeHibrido    Modalidade = "HIBRIDO"

	StatusCursoCriado    StatusCurso = "CRIADO"
	StatusCursoAberto    StatusCurso = "ABERTO"
	StatusCursoEncerrado StatusCurso = "ENCERRADO"

	FormatoAulaGravado FormatoAula = "GRAVADO"
	FormatoAulaAoVivo  FormatoAula = "AO_VIVO"
)

type Curso struct {
	ID                   int            `json:"id" gorm:"primaryKey"`
	Titulo               string         `json:"titulo" gorm:"not null"`
	Descricao            string         `json:"descricao" gorm:"type:text"`
	OrgaoID              int            `json:"orgao_id" gorm:"column:orgao_id"`
	InstituicaoID        int            `json:"instituicao_id" gorm:"column:instituicao_id"`
	Modalidade           Modalidade     `json:"modalidade" gorm:"type:modalidade_enum"`
	LocalRealizacao      string         `json:"local_realizacao" gorm:"column:local_realizacao"`
	DataInicio           time.Time      `json:"data_inicio" gorm:"column:data_inicio"`
	DataTermino          time.Time      `json:"data_termino" gorm:"column:data_termino"`
	DataLimiteInscricoes time.Time      `json:"data_limite_inscricoes" gorm:"column:data_limite_inscricoes"`
	NumeroVagas          int            `json:"numero_vagas" gorm:"column:numero_vagas"`
	CargaHoraria         int            `json:"carga_horaria" gorm:"column:carga_horaria"`
	PreRequisitos        string         `json:"pre_requisitos" gorm:"column:pre_requisitos;type:text"`
	CertificacaoOferecida bool          `json:"certificacao_oferecida" gorm:"column:certificacao_oferecida"`
	Status               StatusCurso    `json:"status" gorm:"type:status_curso_enum"`
	Turno                Turno          `json:"turno" gorm:"type:turno_enum"`
	FormatoAula          FormatoAula    `json:"formato_aula" gorm:"column:formato_aula;type:formato_aula_enum"`
	LinkInscricao        string         `json:"link_inscricao" gorm:"column:link_inscricao"`
	ContatoDuvidas       string         `json:"contato_duvidas" gorm:"column:contato_duvidas"`
	CreatedAt            time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	
	// Relacionamentos
	Orgao                *Orgao         `json:"orgao,omitempty" gorm:"foreignKey:OrgaoID"`
	Instituicao          *InstituicaoEnsino `json:"instituicao,omitempty" gorm:"foreignKey:InstituicaoID"`
	Categorias           []Categoria    `json:"categorias,omitempty" gorm:"many2many:cursos_categorias;"`
	Acessibilidades      []Acessibilidade `json:"acessibilidades,omitempty" gorm:"many2many:cursos_acessibilidades;"`
}

// TableName especifica o nome da tabela para este modelo
func (Curso) TableName() string {
	return "cursos"
} 