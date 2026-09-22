package services

// EmailTriggerKind distinguishes how a communication email is dispatched.
type EmailTriggerKind string

const (
	// EmailTriggerEvent fires only on a business action (create/status/etapa change).
	EmailTriggerEvent EmailTriggerKind = "event"
	// EmailTriggerTemporal fires on a calendar offset relative to an anchor date.
	EmailTriggerTemporal EmailTriggerKind = "temporal"
)

// EmailTemporalAnchor identifies the date used to evaluate temporal offsets.
type EmailTemporalAnchor string

const (
	// EmailAnchorClassStart uses the enrollment's class/turma start date.
	EmailAnchorClassStart EmailTemporalAnchor = "class_start"
)

// EmailCommunicationRule binds one email to exactly one trigger model.
// Event emails have no temporal offset. Temporal emails declare their own offset
// and are never shared across unrelated messages.
type EmailCommunicationRule struct {
	ID          string
	Module      string // "capacitacao" | "empregabilidade"
	Description string
	Kind        EmailTriggerKind

	// Temporal fields — only meaningful when Kind == EmailTriggerTemporal.
	TemporalAnchor     EmailTemporalAnchor
	TemporalOffsetDays int  // e.g. -1 = D-1, +2 = D+2 relative to the anchor
	SendEnabled        bool // false = rule documented/ready, worker must not send yet
}

// EmailCommunicationRules returns the catalog of communication emails and their
// exclusive triggers. Most are event-only; temporal offsets are per-email.
func EmailCommunicationRules() []EmailCommunicationRule {
	return []EmailCommunicationRule{
		// Capacitação — event triggers
		{
			ID:          "enrollment.created",
			Module:      "capacitacao",
			Description: "Inscrição recebida (status pendente)",
			Kind:        EmailTriggerEvent,
		},
		{
			ID:          "enrollment.approved",
			Module:      "capacitacao",
			Description: "Inscrição confirmada",
			Kind:        EmailTriggerEvent,
		},
		{
			ID:          "enrollment.rejected",
			Module:      "capacitacao",
			Description: "Reprovação de inscrição",
			Kind:        EmailTriggerEvent,
		},
		{
			ID:          "enrollment.concluded",
			Module:      "capacitacao",
			Description: "Conclusão de curso",
			Kind:        EmailTriggerEvent,
		},
		{
			ID:          "schedule.changed",
			Module:      "capacitacao",
			Description: "Troca de turma",
			Kind:        EmailTriggerEvent,
		},
		// Capacitação — temporal (own offset; not shared with other emails)
		{
			ID:                 "enrollment.class_reminder_d1",
			Module:             "capacitacao",
			Description:        "Lembrete D-1 de aula",
			Kind:               EmailTriggerTemporal,
			TemporalAnchor:     EmailAnchorClassStart,
			TemporalOffsetDays: -1,
			SendEnabled:        true,
		},
		// Empregabilidade — event triggers
		{
			ID:          "candidatura.enviada",
			Module:      "empregabilidade",
			Description: "Candidatura recebida",
			Kind:        EmailTriggerEvent,
		},
		{
			ID:          "candidatura.aprovada",
			Module:      "empregabilidade",
			Description: "Candidatura aprovada",
			Kind:        EmailTriggerEvent,
		},
		{
			ID:          "candidatura.reprovada",
			Module:      "empregabilidade",
			Description: "Candidatura reprovada",
			Kind:        EmailTriggerEvent,
		},
		{
			ID:          "candidatura.proxima_etapa",
			Module:      "empregabilidade",
			Description: "Candidatura avançou de etapa",
			Kind:        EmailTriggerEvent,
		},
	}
}

// ActiveTemporalClassStartRules returns temporal emails anchored on class start
// that the ClassReminderWorker should evaluate. Each rule carries its own offset.
func ActiveTemporalClassStartRules() []EmailCommunicationRule {
	var out []EmailCommunicationRule
	for _, rule := range EmailCommunicationRules() {
		if rule.Kind != EmailTriggerTemporal {
			continue
		}
		if rule.TemporalAnchor != EmailAnchorClassStart {
			continue
		}
		out = append(out, rule)
	}
	return out
}
