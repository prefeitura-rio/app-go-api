package empregabilidade

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TipoEventoCandidatura diz que transição da candidatura o evento registra.
type TipoEventoCandidatura string

const (
	// EventoCandidaturaCriada é o primeiro evento de uma candidatura nova.
	EventoCandidaturaCriada TipoEventoCandidatura = "candidatura_criada"
	// EventoMudancaEtapa é a ida para uma etapa de ordem maior (ou a primeira etapa).
	EventoMudancaEtapa TipoEventoCandidatura = "mudanca_etapa"
	// EventoVoltaEtapa é a volta para uma etapa de ordem menor.
	EventoVoltaEtapa TipoEventoCandidatura = "volta_etapa"
	// EventoCandidaturaAprovada é a mudança de status para aprovada.
	EventoCandidaturaAprovada TipoEventoCandidatura = "candidatura_aprovada"
	// EventoCandidaturaReprovada é a mudança de status para reprovada; o ator
	// diz se foi manual (operador) ou automática (automação).
	EventoCandidaturaReprovada TipoEventoCandidatura = "candidatura_reprovada"
	// EventoCandidaturaCancelada é a exclusão da candidatura (desistência do
	// candidato ou remoção pelo operador).
	EventoCandidaturaCancelada TipoEventoCandidatura = "candidatura_cancelada"
	// EventoStatusAlterado é qualquer outra mudança de status (reabertura,
	// vaga congelada, descontinuada, descongelada ou reativada).
	EventoStatusAlterado TipoEventoCandidatura = "status_alterado"
)

// AtorEventoCandidatura diz quem provocou a transição. Nunca identifica a
// pessoa: o histórico não guarda dado pessoal.
type AtorEventoCandidatura string

const (
	AtorCandidato AtorEventoCandidatura = "candidato"
	AtorOperador  AtorEventoCandidatura = "operador"
	AtorAutomacao AtorEventoCandidatura = "automacao"
)

// Motivos (códigos, nunca texto livre) das transições feitas pela automação
// quando a vaga muda de estado.
const (
	MotivoVagaCongelada     = "vaga_congelada"
	MotivoVagaDescongelada  = "vaga_descongelada"
	MotivoVagaDescontinuada = "vaga_descontinuada"
	MotivoVagaReativada     = "vaga_reativada"
)

// CandidaturaHistorico é um evento do histórico da candidatura. A tabela é só
// de inclusão (o banco recusa UPDATE, DELETE e TRUNCATE) e guarda só ids:
// nada de CPF, nome ou e-mail. Em eventos de status, a etapa de antes e a de
// depois são a etapa atual; em eventos de etapa, os dois status são o atual.
type CandidaturaHistorico struct {
	ID              int64                 `json:"id" gorm:"primaryKey;autoIncrement"`
	IDCandidatura   uuid.UUID             `json:"id_candidatura" gorm:"type:uuid;not null"`
	IDVaga          uuid.UUID             `json:"id_vaga" gorm:"type:uuid;not null"`
	TipoEvento      TipoEventoCandidatura `json:"tipo_evento" gorm:"type:varchar(64);not null"`
	Ator            AtorEventoCandidatura `json:"ator" gorm:"type:varchar(32);not null"`
	StatusAnterior  *StatusCandidatura    `json:"status_anterior,omitempty" gorm:"type:varchar(100)"`
	StatusNovo      StatusCandidatura     `json:"status_novo" gorm:"type:varchar(100);not null"`
	IDEtapaAnterior *uuid.UUID            `json:"id_etapa_anterior,omitempty" gorm:"type:uuid"`
	IDEtapaNova     *uuid.UUID            `json:"id_etapa_nova,omitempty" gorm:"type:uuid"`
	Motivo          *string               `json:"motivo,omitempty" gorm:"type:varchar(100)"`
	CreatedAt       time.Time             `json:"created_at" gorm:"autoCreateTime"`
}

func (CandidaturaHistorico) TableName() string {
	return "emp_candidatura_historico"
}

// OrigemEventoCandidatura é quem provoca as transições de uma operação e, na
// automação, por quê. Vai no contexto até o repositório, que grava o evento na
// mesma transação da mudança.
type OrigemEventoCandidatura struct {
	Ator   AtorEventoCandidatura
	Motivo string
}

type origemEventoCandidaturaKey struct{}

// ComOrigemEvento devolve um contexto que atribui as transições a origem.
func ComOrigemEvento(ctx context.Context, origem OrigemEventoCandidatura) context.Context {
	return context.WithValue(ctx, origemEventoCandidaturaKey{}, origem)
}

// OrigemEventoDoContexto devolve a origem gravada no contexto, se houver.
func OrigemEventoDoContexto(ctx context.Context) (OrigemEventoCandidatura, bool) {
	origem, ok := ctx.Value(origemEventoCandidaturaKey{}).(OrigemEventoCandidatura)
	return origem, ok && origem.Ator != ""
}

// TipoEventoPorStatus é o tipo do evento de uma mudança de status para novo.
func TipoEventoPorStatus(novo StatusCandidatura) TipoEventoCandidatura {
	switch novo {
	case StatusCandidaturaAprovada:
		return EventoCandidaturaAprovada
	case StatusCandidaturaReprovada:
		return EventoCandidaturaReprovada
	default:
		return EventoStatusAlterado
	}
}
