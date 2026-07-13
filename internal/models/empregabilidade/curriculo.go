package empregabilidade

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StatusFormacao string

const (
	StatusFormacaoCompleto    StatusFormacao = "Completo"
	StatusFormacaoEmAndamento StatusFormacao = "Em andamento"
	StatusFormacaoIncompleto  StatusFormacao = "Incompleto"
)

func (s StatusFormacao) IsValid() bool {
	switch s {
	case StatusFormacaoCompleto, StatusFormacaoEmAndamento, StatusFormacaoIncompleto:
		return true
	}
	return false
}

type CurriculoFormacao struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CPF             string         `json:"cpf" gorm:"type:varchar(14);not null"`
	IDEscolaridade  *uuid.UUID     `json:"id_escolaridade" gorm:"type:uuid"`
	NomeInstituicao string         `json:"nome_instituicao" gorm:"type:varchar(500)"`
	NomeCurso       string         `json:"nome_curso" gorm:"type:varchar(500)"`
	Status          StatusFormacao `json:"status" gorm:"type:varchar(50)"`
	AnoConclusao    string         `json:"ano_conclusao" gorm:"type:varchar(4)"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Escolaridade *Escolaridade `json:"escolaridade,omitempty" gorm:"foreignKey:IDEscolaridade"`
}

func (f *CurriculoFormacao) Validate() error {
	if f.Status != "" && !f.Status.IsValid() {
		return fmt.Errorf("status de formação inválido: %s", f.Status)
	}
	return nil
}

func (CurriculoFormacao) TableName() string {
	return "emp_curriculo_formacoes"
}

type CurriculoIdioma struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CPF       string    `json:"cpf" gorm:"type:varchar(14);not null"`
	IDIdioma  uuid.UUID `json:"id_idioma" gorm:"type:uuid;not null"`
	IDNivel   uuid.UUID `json:"id_nivel" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Idioma *Idioma      `json:"idioma,omitempty" gorm:"foreignKey:IDIdioma"`
	Nivel  *NivelIdioma `json:"nivel,omitempty" gorm:"foreignKey:IDNivel"`
}

func (CurriculoIdioma) TableName() string {
	return "emp_curriculo_idiomas"
}

type CurriculoCursoComplementar struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CPF             string    `json:"cpf" gorm:"type:varchar(14);not null"`
	NomeCurso       string    `json:"nome_curso" gorm:"type:varchar(500);not null"`
	NomeInstituicao string    `json:"nome_instituicao" gorm:"type:varchar(500)"`
	AnoConclusao    string    `json:"ano_conclusao" gorm:"type:varchar(4)"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CurriculoCursoComplementar) TableName() string {
	return "emp_curriculo_cursos_complementares"
}

type CurriculoExperiencia struct {
	ID                      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CPF                     string    `json:"cpf" gorm:"type:varchar(14);not null"`
	Cargo                   string    `json:"cargo" gorm:"type:varchar(500);not null"`
	Empresa                 string    `json:"empresa" gorm:"type:varchar(500);not null"`
	EhTrabalhoAtual         bool      `json:"eh_trabalho_atual" gorm:"default:false"`
	DescricaoAtividades     string    `json:"descricao_atividades" gorm:"type:text"`
	TempoExperienciaMeses   *int      `json:"tempo_experiencia_meses" gorm:"type:integer"`
	ExperienciaComprovadaCT bool      `json:"experiencia_comprovada_ct" gorm:"default:false"`
	CreatedAt               time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CurriculoExperiencia) TableName() string {
	return "emp_curriculo_experiencias"
}

type CurriculoConquista struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CPF             string     `json:"cpf" gorm:"type:varchar(14);not null"`
	IDTipoConquista *uuid.UUID `json:"id_tipo_conquista" gorm:"type:uuid"`
	Titulo          string     `json:"titulo" gorm:"type:varchar(500);not null"`
	Descricao       string     `json:"descricao" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	TipoConquista *TipoConquista `json:"tipo_conquista,omitempty" gorm:"foreignKey:IDTipoConquista"`
}

func (CurriculoConquista) TableName() string {
	return "emp_curriculo_conquistas"
}

type FormacaoAccordionRequest struct {
	Formacoes []*CurriculoFormacao `json:"formacoes"`
	Idiomas   []*CurriculoIdioma   `json:"idiomas"`
}

type ExperienciaProfissionalAccordionRequest struct {
	Experiencias       []*CurriculoExperiencia `json:"experiencias"`
	Conquistas         []*CurriculoConquista   `json:"conquistas"`
	ResumoProfissional string                  `json:"resumo_profissional"`
}

type CurriculoPerfil struct {
	CPF                string    `json:"cpf" gorm:"type:varchar(14);primaryKey"`
	ResumoProfissional string    `json:"resumo_profissional" gorm:"type:text"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CurriculoPerfil) TableName() string {
	return "emp_curriculo_perfil"
}

type CurriculoSituacaoInteresses struct {
	CPF                        string      `json:"cpf" gorm:"type:varchar(14);primaryKey"`
	IDSituacao                 *uuid.UUID  `json:"id_situacao" gorm:"type:uuid"`
	TempoProcurandoEmprego     string      `json:"tempo_procurando_emprego" gorm:"type:varchar(50)"`
	IDDisponibilidade          *uuid.UUID  `json:"id_disponibilidade" gorm:"type:uuid"`
	IDsTiposVinculoPreferencia []uuid.UUID `json:"ids_tipos_vinculo_preferencia" gorm:"type:jsonb;serializer:json"`
	CreatedAt                  time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt                  time.Time   `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Situacao        *SituacaoAtual   `json:"situacao,omitempty" gorm:"foreignKey:IDSituacao"`
	Disponibilidade *Disponibilidade `json:"disponibilidade,omitempty" gorm:"foreignKey:IDDisponibilidade"`
}

func (CurriculoSituacaoInteresses) TableName() string {
	return "emp_curriculo_situacao_interesses"
}
