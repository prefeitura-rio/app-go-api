package workers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// DEV-238: uma notificação de candidatura reprovada sem a vaga no payload derrubava
// o processo inteiro (panic sem recover no worker).

type recordingSender struct {
	mu   sync.Mutex
	sent []*clients.EmailRequest
}

func (s *recordingSender) SendEmail(_ context.Context, req *clients.EmailRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, req)
	return nil
}

func (s *recordingSender) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sent)
}

// panickingNotifier simula uma falha inesperada (panic) ao processar um tipo de tarefa.
type panickingNotifier struct {
	services.EmailNotifier
	mu        sync.Mutex
	enviadas  int
	reprovada int
}

func (p *panickingNotifier) SendCandidaturaReprovadaEmail(context.Context, *empregabilidade.Candidatura) error {
	p.mu.Lock()
	p.reprovada++
	p.mu.Unlock()
	var vaga *empregabilidade.Vaga
	_ = vaga.Titulo // nil pointer dereference, como no incidente
	return nil
}

func (p *panickingNotifier) SendEnrollmentCreatedEmail(context.Context, *models.Inscricao, *models.Curso) error {
	p.mu.Lock()
	p.enviadas++
	p.mu.Unlock()
	return nil
}

func (p *panickingNotifier) counts() (int, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.enviadas, p.reprovada
}

func buildCandidaturaSemVagaTask(t *testing.T, typ emailTaskType) emailTask {
	t.Helper()
	email := "pessoa.sintetica@example.com"
	payload, err := json.Marshal(candidaturaPayload{
		Candidatura: empregabilidade.Candidatura{CPF: "00000000191", Email: &email},
	})
	require.NoError(t, err)
	return emailTask{Type: typ, Payload: json.RawMessage(payload), CreatedAt: time.Now()}
}

func TestEmailWorker_dispatch_CandidaturaSemVaga_EnabledService_NoPanic(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	sender := &recordingSender{}
	worker.emailService = services.NewEmailNotificationService(sender, nil, nil, nil, true, "oportunidades.rio")

	for _, typ := range []emailTaskType{taskCandidaturaReprovada, taskCandidaturaAprovada} {
		t.Run(string(typ), func(t *testing.T) {
			task := buildCandidaturaSemVagaTask(t, typ)
			var err error
			require.NotPanics(t, func() {
				err = worker.dispatch(context.Background(), &task)
			})
			assert.NoError(t, err)
		})
	}
	assert.Equal(t, 0, sender.count(), "sem vaga o e-mail é pulado")
}

func TestEmailWorker_dispatch_PanicRecoveredAsError(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	worker.emailService = &panickingNotifier{}
	task := buildCandidaturaSemVagaTask(t, taskCandidaturaReprovada)

	var err error
	require.NotPanics(t, func() {
		err = worker.dispatch(context.Background(), &task)
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, errEmailTaskPanicked)
	assert.Contains(t, err.Error(), string(taskCandidaturaReprovada))
}

func TestEmailWorker_consume_PanickingTask_DeadLetteredAndKeepsConsuming(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	notifier := &panickingNotifier{}
	worker.emailService = notifier

	// BRPop consome da direita: a tarefa que entra em pânico é processada primeiro.
	pushTask(t, worker.redis, emailQueueKey, buildCandidaturaSemVagaTask(t, taskCandidaturaReprovada))
	pushTask(t, worker.redis, emailQueueKey, buildEnrollmentTask(t, taskEnrollmentCreated, 0))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		worker.consume(ctx)
	}()

	assert.Eventually(t, func() bool {
		enviadas, _ := notifier.counts()
		return enviadas == 1 && queueLen(worker.redis, emailDeadLetterKey) == 1
	}, 3*time.Second, 20*time.Millisecond, "o worker deve seguir consumindo depois do panic")

	assert.Equal(t, int64(0), queueLen(worker.redis, emailQueueKey), "fila principal deve estar vazia")
	assert.Equal(t, int64(0), delayedQueueLen(worker.redis), "panic é determinístico: não deve ir para retry")

	dead := peekTask(t, worker.redis, emailDeadLetterKey)
	require.NotNil(t, dead)
	assert.Equal(t, taskCandidaturaReprovada, dead.Type)
	_, reprovada := notifier.counts()
	assert.Equal(t, 1, reprovada, "a tarefa com panic não deve ser reexecutada")

	cancel()
	select {
	case <-done:
	case <-time.After(emailBRPOPTimeout + 2*time.Second):
		t.Fatal("consume não encerrou após cancelar o contexto")
	}
}
