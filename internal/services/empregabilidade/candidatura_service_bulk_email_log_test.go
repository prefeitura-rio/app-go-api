package empregabilidade_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// bulkEmailLogBuffer is a concurrency-safe log sink: the bulk e-mail is sent
// from a goroutine, so the test reads it while the service may still write.
type bulkEmailLogBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *bulkEmailLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *bulkEmailLogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// failingBulkEmailNotifier fails every approved/rejected candidatura e-mail.
type failingBulkEmailNotifier struct {
	*mockEmpEmailNotifier
}

func (f *failingBulkEmailNotifier) SendCandidaturaAprovadaEmail(_ context.Context, _ *empregabilidade.Candidatura) error {
	return errors.New("data relay unavailable")
}

func (f *failingBulkEmailNotifier) SendCandidaturaReprovadaEmail(_ context.Context, _ *empregabilidade.Candidatura) error {
	return errors.New("data relay unavailable")
}

func TestCandidaturaService_BulkUpdateStatus_EmailFailureLog_MasksCPF(t *testing.T) {
	statuses := []empregabilidade.StatusCandidatura{
		empregabilidade.StatusCandidaturaAprovada,
		empregabilidade.StatusCandidaturaReprovada,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			sink := &bulkEmailLogBuffer{}
			previous := log.Writer()
			log.SetOutput(sink)
			t.Cleanup(func() { log.SetOutput(previous) })

			const syntheticCPF = "98765432100"
			const maskedCPF = "987******00"

			vagaID := uuid.New()
			repo := NewMockCandidaturaRepo()
			repo.candidaturas[uuid.New()] = &empregabilidade.Candidatura{
				CPF:    syntheticCPF,
				IDVaga: vagaID,
				Status: empregabilidade.StatusCandidaturaEnviada,
			}
			svc := newSvcWithDisabledEmail(repo, NewMockVagaRepo())
			svc.SetEmailNotifier(&failingBulkEmailNotifier{newMockEmpEmailNotifier()})

			result, err := svc.BulkUpdateStatus(context.Background(), vagaID, []string{syntheticCPF}, status)
			if err != nil {
				t.Fatalf("BulkUpdateStatus failed: %v", err)
			}
			if result.Updated != 1 {
				t.Fatalf("expected 1 update, got %d", result.Updated)
			}

			const marker = "falha ao enviar email bulk"
			deadline := time.Now().Add(2 * time.Second)
			for !strings.Contains(sink.String(), marker) && time.Now().Before(deadline) {
				time.Sleep(10 * time.Millisecond)
			}

			out := sink.String()
			if !strings.Contains(out, marker) {
				t.Fatalf("expected bulk e-mail failure to be logged, got:\n%s", out)
			}
			if strings.Contains(out, syntheticCPF) {
				t.Errorf("log must not contain the raw CPF, got:\n%s", out)
			}
			if !strings.Contains(out, maskedCPF) {
				t.Errorf("expected log to contain the masked CPF %q, got:\n%s", maskedCPF, out)
			}
		})
	}
}
