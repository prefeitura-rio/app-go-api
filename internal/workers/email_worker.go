package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

const (
	emailQueueKey          = "email:queue"
	emailDelayedQueueKey   = "email:queue:delayed"
	emailDeadLetterKey     = "email:queue:failed"
	emailMaxRetries        = 5
	emailBRPOPTimeout      = 5 * time.Second
	emailWorkerCount       = 2
	emailRetryPollInterval = 1 * time.Second
)

type emailTaskType string

const (
	taskEnrollmentCreated    emailTaskType = "enrollment.created"
	taskEnrollmentApproved   emailTaskType = "enrollment.approved"
	taskEnrollmentRejected   emailTaskType = "enrollment.rejected"
	taskScheduleChanged      emailTaskType = "schedule.changed"
	taskCandidaturaEnviada   emailTaskType = "candidatura.enviada"
	taskCandidaturaAprovada  emailTaskType = "candidatura.aprovada"
	taskCandidaturaReprovada emailTaskType = "candidatura.reprovada"
)

type emailTask struct {
	Type        emailTaskType   `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	Attempts    int             `json:"attempts"`
	CreatedAt   time.Time       `json:"created_at"`
	NextRetryAt time.Time       `json:"next_retry_at,omitempty"`
}

type enrollmentPayload struct {
	Inscricao models.Inscricao `json:"inscricao"`
	Curso     models.Curso     `json:"curso"`
}

type candidaturaPayload struct {
	Candidatura empregabilidade.Candidatura `json:"candidatura"`
}

// EmailWorker processes email notifications asynchronously via a Redis List queue.
type EmailWorker struct {
	redis        *redis.Client
	emailService *services.EmailNotificationService
	retryDelay   func(attempts int) time.Duration // nil = use default exponential backoff
}

var _ services.EmailNotifier = (*EmailWorker)(nil)

func NewEmailWorker(redisClient *redis.Client, emailService *services.EmailNotificationService) *EmailWorker {
	return &EmailWorker{
		redis:        redisClient,
		emailService: emailService,
	}
}

func (w *EmailWorker) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	// Start a goroutine to move delayed tasks to the main queue
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.processDelayedTasks(ctx)
	}()

	// Start worker consumers
	for range emailWorkerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.consume(ctx)
		}()
	}

	wg.Wait()
	return nil
}

// processDelayedTasks moves tasks from the delayed queue to the main queue when they're ready
func (w *EmailWorker) processDelayedTasks(ctx context.Context) {
	ticker := time.NewTicker(emailRetryPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.moveReadyDelayedTasks(ctx)
		}
	}
}

func (w *EmailWorker) moveReadyDelayedTasks(ctx context.Context) {
	now := float64(time.Now().Unix())

	// Get all tasks with score <= current time
	tasks, err := w.redis.ZRangeByScore(ctx, emailDelayedQueueKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    fmt.Sprintf("%f", now),
		Offset: 0,
		Count:  100, // Process in batches
	}).Result()

	if err != nil {
		log.Printf("[EmailWorker] failed to get delayed tasks: %v", err)
		return
	}

	for _, taskJSON := range tasks {
		// Remove from delayed queue
		if err := w.redis.ZRem(ctx, emailDelayedQueueKey, taskJSON).Err(); err != nil {
			log.Printf("[EmailWorker] failed to remove task from delayed queue: %v", err)
			continue
		}

		// Push to main queue
		if err := w.redis.LPush(ctx, emailQueueKey, taskJSON).Err(); err != nil {
			log.Printf("[EmailWorker] failed to move task to main queue: %v", err)
			// Re-add to delayed queue if push fails
			w.redis.ZAdd(ctx, emailDelayedQueueKey, redis.Z{
				Score:  float64(time.Now().Add(5 * time.Second).Unix()),
				Member: taskJSON,
			})
		}
	}
}

func (w *EmailWorker) consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Try to move delayed tasks before each pop (alternative to separate goroutine)
		// w.moveReadyDelayedTasks(ctx) // Uncomment if you want sync processing

		result, err := w.redis.BRPop(ctx, emailBRPOPTimeout, emailQueueKey).Result()

		if err == redis.Nil {
			continue
		}

		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[EmailWorker] BRPOP error: %v", err)
			continue
		}

		if len(result) < 2 {
			continue
		}

		var task emailTask

		if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
			log.Printf("[EmailWorker] failed to unmarshal task: %v", err)
			continue
		}

		if err := w.dispatch(ctx, &task); err != nil {
			task.Attempts++

			if task.Attempts >= emailMaxRetries {
				log.Printf("[EmailWorker] task %s exhausted retries, moving to dead-letter: %v", task.Type, err)
				w.pushToKey(ctx, emailDeadLetterKey, &task)
			} else {
				// Calculate backoff delay
				delay := w.getRetryDelay(task.Attempts)
				task.NextRetryAt = time.Now().Add(delay)

				log.Printf("[EmailWorker] task %s failed (attempt %d/%d), retrying in %v: %v",
					task.Type, task.Attempts, emailMaxRetries, delay, err)

				// Add to delayed queue instead of immediate requeue
				w.addToDelayedQueue(ctx, &task)
			}
		}
	}
}

// getRetryDelay returns exponential backoff delay for the given attempt number.
func (w *EmailWorker) getRetryDelay(attempts int) time.Duration {
	if w.retryDelay != nil {
		return w.retryDelay(attempts)
	}
	switch attempts {
	case 1:
		return 30 * time.Second
	case 2:
		return 1 * time.Minute
	case 3:
		return 2 * time.Minute
	case 4:
		return 4 * time.Minute
	default:
		return 8 * time.Minute
	}
}

func (w *EmailWorker) addToDelayedQueue(ctx context.Context, task *emailTask) {
	taskBytes, err := json.Marshal(task)

	if err != nil {
		log.Printf("[EmailWorker] failed to marshal task for delayed queue: %v", err)
		return
	}

	score := float64(task.NextRetryAt.Unix())

	if err := w.redis.ZAdd(ctx, emailDelayedQueueKey, redis.Z{
		Score:  score,
		Member: taskBytes,
	}).Err(); err != nil {
		log.Printf("[EmailWorker] failed to add task to delayed queue: %v", err)
		// Fallback: immediate requeue
		w.redis.LPush(ctx, emailQueueKey, taskBytes)
	}
}

func (w *EmailWorker) dispatch(ctx context.Context, task *emailTask) error {
	switch task.Type {
	case taskEnrollmentCreated:
		var p enrollmentPayload

		if err := json.Unmarshal(task.Payload, &p); err != nil {
			return err
		}

		return w.emailService.SendEnrollmentCreatedEmail(ctx, &p.Inscricao, &p.Curso)

	case taskEnrollmentApproved:
		var p enrollmentPayload

		if err := json.Unmarshal(task.Payload, &p); err != nil {
			return err
		}

		return w.emailService.SendEnrollmentApprovedEmail(ctx, &p.Inscricao, &p.Curso)

	case taskEnrollmentRejected:
		var p enrollmentPayload

		if err := json.Unmarshal(task.Payload, &p); err != nil {
			return err
		}

		return w.emailService.SendEnrollmentRejectedEmail(ctx, &p.Inscricao, &p.Curso)

	case taskScheduleChanged:
		var p enrollmentPayload

		if err := json.Unmarshal(task.Payload, &p); err != nil {
			return err
		}

		return w.emailService.SendScheduleChangedEmail(ctx, &p.Inscricao, &p.Curso)

	case taskCandidaturaEnviada:
		var p candidaturaPayload

		if err := json.Unmarshal(task.Payload, &p); err != nil {
			return err
		}

		return w.emailService.SendCandidaturaEnviadaEmail(ctx, &p.Candidatura)

	case taskCandidaturaAprovada:
		var p candidaturaPayload

		if err := json.Unmarshal(task.Payload, &p); err != nil {
			return err
		}

		return w.emailService.SendCandidaturaAprovadaEmail(ctx, &p.Candidatura)

	case taskCandidaturaReprovada:
		var p candidaturaPayload

		if err := json.Unmarshal(task.Payload, &p); err != nil {
			return err
		}

		return w.emailService.SendCandidaturaReprovadaEmail(ctx, &p.Candidatura)

	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

func (w *EmailWorker) enqueue(ctx context.Context, taskType emailTaskType, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	task := emailTask{
		Type:      taskType,
		Payload:   json.RawMessage(payloadBytes),
		Attempts:  0,
		CreatedAt: time.Now(),
	}

	taskBytes, err := json.Marshal(task)

	if err != nil {
		return fmt.Errorf("failed to marshal email task: %w", err)
	}

	return w.redis.LPush(ctx, emailQueueKey, taskBytes).Err()
}

func (w *EmailWorker) pushToKey(ctx context.Context, key string, task *emailTask) {
	taskBytes, err := json.Marshal(task)

	if err != nil {
		log.Printf("[EmailWorker] failed to marshal task for push: %v", err)
		return
	}

	if err := w.redis.LPush(ctx, key, taskBytes).Err(); err != nil {
		log.Printf("[EmailWorker] failed to push task to %s: %v", key, err)
	}
}

func (w *EmailWorker) SendEnrollmentCreatedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	return w.enqueue(ctx, taskEnrollmentCreated, enrollmentPayload{Inscricao: *inscricao, Curso: *curso})
}

func (w *EmailWorker) SendEnrollmentApprovedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	return w.enqueue(ctx, taskEnrollmentApproved, enrollmentPayload{Inscricao: *inscricao, Curso: *curso})
}

func (w *EmailWorker) SendEnrollmentRejectedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	return w.enqueue(ctx, taskEnrollmentRejected, enrollmentPayload{Inscricao: *inscricao, Curso: *curso})
}

func (w *EmailWorker) SendScheduleChangedEmail(ctx context.Context, inscricao *models.Inscricao, curso *models.Curso) error {
	return w.enqueue(ctx, taskScheduleChanged, enrollmentPayload{Inscricao: *inscricao, Curso: *curso})
}

func (w *EmailWorker) SendCandidaturaEnviadaEmail(ctx context.Context, candidatura *empregabilidade.Candidatura) error {
	return w.enqueue(ctx, taskCandidaturaEnviada, candidaturaPayload{Candidatura: *candidatura})
}

func (w *EmailWorker) SendCandidaturaAprovadaEmail(ctx context.Context, candidatura *empregabilidade.Candidatura) error {
	return w.enqueue(ctx, taskCandidaturaAprovada, candidaturaPayload{Candidatura: *candidatura})
}

func (w *EmailWorker) SendCandidaturaReprovadaEmail(ctx context.Context, candidatura *empregabilidade.Candidatura) error {
	return w.enqueue(ctx, taskCandidaturaReprovada, candidaturaPayload{Candidatura: *candidatura})
}
