package workers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// --- test helpers ---

func setupEmailWorkerTest(t *testing.T) (*EmailWorker, *miniredis.Miniredis, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err, "failed to start miniredis")

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	emailSvc := services.NewEmailNotificationService(nil, nil, nil, nil, false, "")
	worker := NewEmailWorker(redisClient, emailSvc)

	cleanup := func() {
		redisClient.Close()
		mr.Close()
	}
	return worker, mr, cleanup
}

// peekTask unmarshals the task at the right end of key (the next BRPop candidate).
func peekTask(t *testing.T, client *redis.Client, key string) *emailTask {
	t.Helper()
	items, err := client.LRange(context.Background(), key, -1, -1).Result()
	require.NoError(t, err)
	if len(items) == 0 {
		return nil
	}
	var task emailTask
	require.NoError(t, json.Unmarshal([]byte(items[0]), &task))
	return &task
}

func queueLen(client *redis.Client, key string) int64 {
	n, _ := client.LLen(context.Background(), key).Result()
	return n
}

func delayedQueueLen(client *redis.Client) int64 {
	n, _ := client.ZCard(context.Background(), emailDelayedQueueKey).Result()
	return n
}

// buildEnrollmentTask creates a ready-to-push emailTask with an enrollment payload.
func buildEnrollmentTask(t *testing.T, typ emailTaskType, attempts int) emailTask {
	t.Helper()
	payload, err := json.Marshal(enrollmentPayload{
		Inscricao: models.Inscricao{Email: "user@example.com"},
		Curso:     models.Curso{ID: 1, Titulo: "Curso Test"},
	})
	require.NoError(t, err)
	return emailTask{Type: typ, Payload: json.RawMessage(payload), Attempts: attempts, CreatedAt: time.Now()}
}

// buildCandidaturaTask creates a ready-to-push emailTask with a candidatura payload.
func buildCandidaturaTask(t *testing.T, typ emailTaskType, attempts int) emailTask {
	t.Helper()
	payload, err := json.Marshal(candidaturaPayload{
		Candidatura: empregabilidade.Candidatura{CPF: "12345678901"},
	})
	require.NoError(t, err)
	return emailTask{Type: typ, Payload: json.RawMessage(payload), Attempts: attempts, CreatedAt: time.Now()}
}

func pushTask(t *testing.T, client *redis.Client, key string, task emailTask) {
	t.Helper()
	b, err := json.Marshal(task)
	require.NoError(t, err)
	require.NoError(t, client.LPush(context.Background(), key, b).Err())
}

// --- construction ---

func TestNewEmailWorker(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	assert.NotNil(t, worker)
	assert.NotNil(t, worker.redis)
	assert.NotNil(t, worker.emailService)
}

// --- Send* methods enqueue to Redis ---

func TestEmailWorker_SendEnrollmentCreatedEmail_Enqueues(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	err := worker.SendEnrollmentCreatedEmail(context.Background(),
		&models.Inscricao{Email: "a@b.com"},
		&models.Curso{ID: 1})
	require.NoError(t, err)

	assert.Equal(t, int64(1), queueLen(worker.redis, emailQueueKey))

	task := peekTask(t, worker.redis, emailQueueKey)
	require.NotNil(t, task)
	assert.Equal(t, taskEnrollmentCreated, task.Type)
	assert.Equal(t, 0, task.Attempts)
}

func TestEmailWorker_SendEnrollmentApprovedEmail_Enqueues(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	err := worker.SendEnrollmentApprovedEmail(context.Background(),
		&models.Inscricao{Email: "a@b.com"},
		&models.Curso{ID: 1})
	require.NoError(t, err)

	task := peekTask(t, worker.redis, emailQueueKey)
	require.NotNil(t, task)
	assert.Equal(t, taskEnrollmentApproved, task.Type)
}

func TestEmailWorker_SendEnrollmentRejectedEmail_Enqueues(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	err := worker.SendEnrollmentRejectedEmail(context.Background(),
		&models.Inscricao{Email: "a@b.com"},
		&models.Curso{ID: 1})
	require.NoError(t, err)

	task := peekTask(t, worker.redis, emailQueueKey)
	require.NotNil(t, task)
	assert.Equal(t, taskEnrollmentRejected, task.Type)
}

func TestEmailWorker_SendScheduleChangedEmail_Enqueues(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	err := worker.SendScheduleChangedEmail(context.Background(),
		&models.Inscricao{Email: "a@b.com"},
		&models.Curso{ID: 1})
	require.NoError(t, err)

	task := peekTask(t, worker.redis, emailQueueKey)
	require.NotNil(t, task)
	assert.Equal(t, taskScheduleChanged, task.Type)
}

func TestEmailWorker_SendCandidaturaEnviadaEmail_Enqueues(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	err := worker.SendCandidaturaEnviadaEmail(context.Background(),
		&empregabilidade.Candidatura{CPF: "12345678901"})
	require.NoError(t, err)

	task := peekTask(t, worker.redis, emailQueueKey)
	require.NotNil(t, task)
	assert.Equal(t, taskCandidaturaEnviada, task.Type)
}

func TestEmailWorker_SendCandidaturaAprovadaEmail_Enqueues(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	err := worker.SendCandidaturaAprovadaEmail(context.Background(),
		&empregabilidade.Candidatura{CPF: "12345678901"})
	require.NoError(t, err)

	task := peekTask(t, worker.redis, emailQueueKey)
	require.NotNil(t, task)
	assert.Equal(t, taskCandidaturaAprovada, task.Type)
}

func TestEmailWorker_SendCandidaturaReprovadaEmail_Enqueues(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	err := worker.SendCandidaturaReprovadaEmail(context.Background(),
		&empregabilidade.Candidatura{CPF: "12345678901"})
	require.NoError(t, err)

	task := peekTask(t, worker.redis, emailQueueKey)
	require.NotNil(t, task)
	assert.Equal(t, taskCandidaturaReprovada, task.Type)
}

func TestEmailWorker_Send_MultipleItemsQueueUp(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	ctx := context.Background()
	inscricao := &models.Inscricao{Email: "a@b.com"}
	curso := &models.Curso{ID: 1}

	require.NoError(t, worker.SendEnrollmentCreatedEmail(ctx, inscricao, curso))
	require.NoError(t, worker.SendEnrollmentApprovedEmail(ctx, inscricao, curso))
	require.NoError(t, worker.SendEnrollmentRejectedEmail(ctx, inscricao, curso))

	assert.Equal(t, int64(3), queueLen(worker.redis, emailQueueKey))
}

// --- dispatch ---

func TestEmailWorker_dispatch_EnrollmentTasks(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	ctx := context.Background()
	types := []emailTaskType{
		taskEnrollmentCreated,
		taskEnrollmentApproved,
		taskEnrollmentRejected,
		taskScheduleChanged,
	}

	for _, typ := range types {
		t.Run(string(typ), func(t *testing.T) {
			task := buildEnrollmentTask(t, typ, 0)
			err := worker.dispatch(ctx, &task)
			assert.NoError(t, err)
		})
	}
}

func TestEmailWorker_dispatch_CandidaturaTasks(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	ctx := context.Background()
	types := []emailTaskType{
		taskCandidaturaEnviada,
		taskCandidaturaAprovada,
		taskCandidaturaReprovada,
	}

	for _, typ := range types {
		t.Run(string(typ), func(t *testing.T) {
			task := buildCandidaturaTask(t, typ, 0)
			err := worker.dispatch(ctx, &task)
			assert.NoError(t, err)
		})
	}
}

func TestEmailWorker_dispatch_UnknownType(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	task := emailTask{Type: "unknown.type", Payload: json.RawMessage(`{}`)}
	err := worker.dispatch(context.Background(), &task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown task type")
}

func TestEmailWorker_dispatch_InvalidEnrollmentPayload(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	task := emailTask{
		Type:    taskEnrollmentCreated,
		Payload: json.RawMessage(`not valid json`),
	}
	err := worker.dispatch(context.Background(), &task)
	assert.Error(t, err)
}

func TestEmailWorker_dispatch_InvalidCandidaturaPayload(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	task := emailTask{
		Type:    taskCandidaturaEnviada,
		Payload: json.RawMessage(`not valid json`),
	}
	err := worker.dispatch(context.Background(), &task)
	assert.Error(t, err)
}

// --- consume ---

func TestEmailWorker_consume_SuccessfulTask_EmptiesQueue(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	pushTask(t, worker.redis, emailQueueKey, buildEnrollmentTask(t, taskEnrollmentCreated, 0))

	ctx, cancel := context.WithCancel(context.Background())
	go worker.consume(ctx)

	assert.Eventually(t, func() bool {
		return queueLen(worker.redis, emailQueueKey) == 0
	}, 2*time.Second, 20*time.Millisecond, "task should be consumed from queue")

	cancel()

	// No dead-letter entries for a successful dispatch.
	assert.Equal(t, int64(0), queueLen(worker.redis, emailDeadLetterKey))
}

// TestEmailWorker_consume_FailedTask_MovesToDelayedQueue verifies that a single
// dispatch failure moves the task to the delayed queue (not dead-letter).
func TestEmailWorker_consume_FailedTask_MovesToDelayedQueue(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	pushTask(t, worker.redis, emailQueueKey, emailTask{
		Type:    "bad.type",
		Payload: json.RawMessage(`{}`),
	})

	ctx, cancel := context.WithCancel(context.Background())
	go worker.consume(ctx)

	assert.Eventually(t, func() bool {
		return delayedQueueLen(worker.redis) == 1
	}, 2*time.Second, 20*time.Millisecond, "failed task should be moved to delayed queue")

	cancel()

	assert.Equal(t, int64(0), queueLen(worker.redis, emailQueueKey), "main queue should be empty")
	assert.Equal(t, int64(0), queueLen(worker.redis, emailDeadLetterKey), "should not be in dead-letter yet")
}

// TestEmailWorker_consume_FullRetryLifecycle_DeadLettered uses zero retry delay so all
// emailMaxRetries attempts complete quickly and the task lands in the dead-letter queue.
func TestEmailWorker_consume_FullRetryLifecycle_DeadLettered(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	// Zero retry delay so tasks re-enter the main queue immediately.
	worker.retryDelay = func(int) time.Duration { return 0 }

	pushTask(t, worker.redis, emailQueueKey, emailTask{
		Type:    "bad.type",
		Payload: json.RawMessage(`{}`),
	})

	ctx, cancel := context.WithCancel(context.Background())
	// Start the full worker (consume + processDelayedTasks).
	go worker.Start(ctx) //nolint:errcheck

	assert.Eventually(t, func() bool {
		return queueLen(worker.redis, emailDeadLetterKey) == 1
	}, 15*time.Second, 50*time.Millisecond, "task should be dead-lettered after exhausting retries")

	cancel()

	assert.Equal(t, int64(0), queueLen(worker.redis, emailQueueKey), "main queue should be empty")

	task := peekTask(t, worker.redis, emailDeadLetterKey)
	require.NotNil(t, task)
	assert.Equal(t, emailMaxRetries, task.Attempts, "dead-letter task should record all attempts")
}

// TestEmailWorker_consume_ExhaustedRetries_ImmediateDeadLetter starts with
// Attempts=emailMaxRetries-1 so a single failure is enough to dead-letter the task.
func TestEmailWorker_consume_ExhaustedRetries_ImmediateDeadLetter(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	pushTask(t, worker.redis, emailQueueKey, emailTask{
		Type:     "bad.type",
		Payload:  json.RawMessage(`{}`),
		Attempts: emailMaxRetries - 1,
	})

	ctx, cancel := context.WithCancel(context.Background())
	go worker.consume(ctx)

	assert.Eventually(t, func() bool {
		return queueLen(worker.redis, emailDeadLetterKey) == 1
	}, 2*time.Second, 20*time.Millisecond, "task should be immediately dead-lettered")

	cancel()

	assert.Equal(t, int64(0), queueLen(worker.redis, emailQueueKey), "main queue should be empty")

	task := peekTask(t, worker.redis, emailDeadLetterKey)
	require.NotNil(t, task)
	assert.Equal(t, emailMaxRetries, task.Attempts)
}

func TestEmailWorker_consume_InvalidJSON_Skipped(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	// Push raw garbage — unmarshal will fail, task is dropped (not requeued).
	require.NoError(t,
		worker.redis.LPush(context.Background(), emailQueueKey, "not-valid-json").Err())

	ctx, cancel := context.WithCancel(context.Background())
	go worker.consume(ctx)

	assert.Eventually(t, func() bool {
		return queueLen(worker.redis, emailQueueKey) == 0
	}, 2*time.Second, 20*time.Millisecond, "invalid JSON task should be dropped")

	cancel()
	assert.Equal(t, int64(0), queueLen(worker.redis, emailDeadLetterKey))
}

func TestEmailWorker_consume_StopsOnContextCancel(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		worker.consume(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// consume exited cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("consume did not stop after context cancellation")
	}
}

// --- Start ---

func TestEmailWorker_Start_StopsOnContextCancel(t *testing.T) {
	worker, _, cleanup := setupEmailWorkerTest(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- worker.Start(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}
