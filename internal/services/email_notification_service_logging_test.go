package services

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// Dados sintéticos (DEV-586): endereço inventado num domínio reservado.
const (
	emailLogTestLocalPart = "fulana.sintetica"
	emailLogTestAddress   = emailLogTestLocalPart + "@example.com"
	emailLogTestMasked    = "***@example.com"
	emailLogTestCPF       = "11111111111"
)

// newEmailLoggingTestService returns a service whose informational logger
// writes to info and captures the standard logger (stderr in production, so
// ERROR in GKE) in stderr for the duration of the test.
func newEmailLoggingTestService(t *testing.T, enabled bool) (*EmailNotificationService, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	info := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	previous := log.Writer()
	log.SetOutput(stderr)
	t.Cleanup(func() { log.SetOutput(previous) })

	s := NewEmailNotificationService(&MockDataRelayClient{}, nil, nil, nil, enabled, "oportunidades.rio")
	s.infoLog = log.New(info, "", 0)
	return s, info, stderr
}

func emailLogTestInscricao(email string) *models.Inscricao {
	return &models.Inscricao{
		ID:    uuid.New(),
		Name:  "Fulana Sintetica",
		Email: email,
		CPF:   emailLogTestCPF,
	}
}

func emailLogTestCandidatura(email *string) *empregabilidade.Candidatura {
	nome := "Fulana Sintetica"
	return &empregabilidade.Candidatura{
		ID:    uuid.New(),
		Nome:  &nome,
		Email: email,
		CPF:   emailLogTestCPF,
		Vaga: &empregabilidade.Vaga{
			ID:          uuid.New(),
			Titulo:      "Vaga Sintetica",
			Contratante: &empregabilidade.Empresa{NomeFantasia: "Empresa Sintetica"},
		},
	}
}

type emailLogCase struct {
	name string
	// send calls one EmailNotificationService method. email is the address on
	// the enrollment or application ("" for none) and the returned string is the
	// internal ID expected in the log line.
	send func(ctx context.Context, s *EmailNotificationService, email string) string
}

func emailLogCases() []emailLogCase {
	curso := &models.Curso{ID: 1, Titulo: "Curso Sintetico", Organization: "SMTR"}

	inscricaoCase := func(name string, call func(*EmailNotificationService, context.Context, *models.Inscricao, *models.Curso) error) emailLogCase {
		return emailLogCase{name: name, send: func(ctx context.Context, s *EmailNotificationService, email string) string {
			inscricao := emailLogTestInscricao(email)
			_ = call(s, ctx, inscricao, curso)
			return inscricao.ID.String()
		}}
	}
	candidaturaCase := func(name string, call func(*EmailNotificationService, context.Context, *empregabilidade.Candidatura) error) emailLogCase {
		return emailLogCase{name: name, send: func(ctx context.Context, s *EmailNotificationService, email string) string {
			var addr *string
			if email != "" {
				addr = &email
			}
			candidatura := emailLogTestCandidatura(addr)
			_ = call(s, ctx, candidatura)
			return candidatura.ID.String()
		}}
	}

	return []emailLogCase{
		inscricaoCase("enrollment created", (*EmailNotificationService).SendEnrollmentCreatedEmail),
		inscricaoCase("enrollment approved", (*EmailNotificationService).SendEnrollmentApprovedEmail),
		inscricaoCase("enrollment rejected", (*EmailNotificationService).SendEnrollmentRejectedEmail),
		inscricaoCase("enrollment concluded", (*EmailNotificationService).SendEnrollmentConcludedEmail),
		inscricaoCase("class reminder", (*EmailNotificationService).SendEnrollmentClassReminderEmail),
		inscricaoCase("schedule changed", (*EmailNotificationService).SendScheduleChangedEmail),
		candidaturaCase("application received", (*EmailNotificationService).SendCandidaturaEnviadaEmail),
		candidaturaCase("application approved", (*EmailNotificationService).SendCandidaturaAprovadaEmail),
		candidaturaCase("application failed", (*EmailNotificationService).SendCandidaturaReprovadaEmail),
		candidaturaCase("application next stage", func(s *EmailNotificationService, ctx context.Context, c *empregabilidade.Candidatura) error {
			return s.SendCandidaturaProximaEtapaEmail(ctx, c, "Entrevista")
		}),
	}
}

func assertNoEmailServiceLineOnStderr(t *testing.T, stderr *bytes.Buffer) {
	t.Helper()
	if strings.Contains(stderr.String(), "[EmailNotificationService]") {
		t.Errorf("expected no EmailNotificationService line on stderr (ERROR in GKE), got: %q", stderr.String())
	}
}

func assertNoCitizenEmail(t *testing.T, out string) {
	t.Helper()
	if strings.Contains(out, emailLogTestLocalPart) {
		t.Errorf("expected the citizen's e-mail to be absent from the logs, got: %q", out)
	}
}

func TestEmailNotificationService_SentLog_InfoWithoutCitizenEmail(t *testing.T) {
	for _, tc := range emailLogCases() {
		t.Run(tc.name, func(t *testing.T) {
			s, info, stderr := newEmailLoggingTestService(t, true)

			id := tc.send(context.Background(), s, emailLogTestAddress)

			assertNoEmailServiceLineOnStderr(t, stderr)
			assertNoCitizenEmail(t, info.String()+stderr.String())

			out := info.String()
			if !strings.Contains(out, "[EmailNotificationService] Sent") {
				t.Fatalf("expected the sent event on the info logger, got: %q", out)
			}
			if !strings.Contains(out, emailLogTestMasked) {
				t.Errorf("expected the masked domain %q in the log, got: %q", emailLogTestMasked, out)
			}
			if !strings.Contains(out, id) {
				t.Errorf("expected the internal ID %q in the log, got: %q", id, out)
			}
		})
	}
}

func TestEmailNotificationService_DisabledLog_InfoWithoutCitizenEmail(t *testing.T) {
	for _, tc := range emailLogCases() {
		t.Run(tc.name, func(t *testing.T) {
			s, info, stderr := newEmailLoggingTestService(t, false)

			tc.send(context.Background(), s, emailLogTestAddress)

			assertNoEmailServiceLineOnStderr(t, stderr)
			assertNoCitizenEmail(t, info.String()+stderr.String())
			if !strings.Contains(info.String(), "disabled") {
				t.Errorf("expected the disabled notice on the info logger, got: %q", info.String())
			}
		})
	}
}

func TestEmailNotificationService_NoAddressLog_InfoNotError(t *testing.T) {
	for _, tc := range emailLogCases() {
		t.Run(tc.name, func(t *testing.T) {
			s, info, stderr := newEmailLoggingTestService(t, true)

			id := tc.send(context.Background(), s, "")

			assertNoEmailServiceLineOnStderr(t, stderr)
			out := info.String()
			if !strings.Contains(out, "No email address") {
				t.Fatalf("expected the missing address notice on the info logger, got: %q", out)
			}
			if !strings.Contains(out, id) {
				t.Errorf("expected the internal ID %q in the log, got: %q", id, out)
			}
			if strings.Contains(out, emailLogTestCPF) {
				t.Errorf("expected the CPF to be masked in the log, got: %q", out)
			}
		})
	}
}

func TestEmailNotificationService_NilInfoLogger_UsesDefault(t *testing.T) {
	s := NewEmailNotificationService(&MockDataRelayClient{}, nil, nil, nil, true, "oportunidades.rio")
	if s.infoLogger() != defaultEmailNotificationInfoLog {
		t.Error("expected the default stdout logger when infoLog is nil")
	}
	custom := log.New(&bytes.Buffer{}, "", 0)
	s.infoLog = custom
	if s.infoLogger() != custom {
		t.Error("expected the configured logger when infoLog is set")
	}
}
