package workers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

const (
	classReminderLockKey       = "email:class_reminder:lock"
	classReminderSentKeyPrefix = "email:temporal:"
	classReminderDefaultTTL    = 48 * time.Hour
)

// ClassReminderEnrollmentLister lists approved enrollments whose class starts in a time window.
type ClassReminderEnrollmentLister interface {
	ListApprovedStartingBetween(ctx context.Context, start, end time.Time) ([]*models.Inscricao, error)
}

// ClassReminderWorker evaluates temporal emails anchored on class start.
// Each email has its own offset from the communication catalog; event-only emails
// are never processed here.
type ClassReminderWorker struct {
	inscricaoRepo ClassReminderEnrollmentLister
	emailNotifier services.EmailNotifier
	redisClient   *redis.Client
	enabled       bool
	syncInterval  time.Duration
	jobs          []TemporalEmailJob
	now           func() time.Time
}

// NewClassReminderWorker creates a worker for per-email temporal class-start rules.
func NewClassReminderWorker(
	inscricaoRepo ClassReminderEnrollmentLister,
	emailNotifier services.EmailNotifier,
	redisClient *redis.Client,
	cfg *config.ClassReminderSettings,
) *ClassReminderWorker {
	return &ClassReminderWorker{
		inscricaoRepo: inscricaoRepo,
		emailNotifier: emailNotifier,
		redisClient:   redisClient,
		enabled:       cfg.Enabled,
		syncInterval:  cfg.SyncInterval,
		jobs:          temporalClassStartJobs(),
		now:           time.Now,
	}
}

func (w *ClassReminderWorker) Start(ctx context.Context) error {
	if !w.enabled {
		log.Println("[ClassReminderWorker] Disabled, not starting")
		return nil
	}

	log.Printf("[ClassReminderWorker] Starting. Interval: %v, temporal emails: %d", w.syncInterval, len(w.jobs))

	if err := w.runCycle(ctx); err != nil {
		log.Printf("[ClassReminderWorker] Initial cycle failed: %v", err)
	}

	ticker := time.NewTicker(w.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[ClassReminderWorker] Context cancelled, shutting down")
			return ctx.Err()
		case <-ticker.C:
			if err := w.runCycle(ctx); err != nil {
				log.Printf("[ClassReminderWorker] Cycle failed: %v", err)
			}
		}
	}
}

func (w *ClassReminderWorker) runCycle(ctx context.Context) error {
	token, acquired, err := w.tryAcquireLock(ctx)
	if err != nil {
		return fmt.Errorf("lock error: %w", err)
	}
	if !acquired {
		log.Println("[ClassReminderWorker] Another instance holds the lock, skipping")
		return nil
	}
	defer w.releaseLock(ctx, token)

	now := w.now()
	for _, job := range w.jobs {
		if err := w.processJob(ctx, now, job); err != nil {
			log.Printf("[ClassReminderWorker] Email %s failed: %v", job.Rule.ID, err)
		}
	}
	return nil
}

func (w *ClassReminderWorker) processJob(ctx context.Context, now time.Time, job TemporalEmailJob) error {
	rule := job.Rule
	targetDay := targetClassDateForOffset(now, rule.TemporalOffsetDays)
	windowEnd := targetDay.Add(24 * time.Hour)

	inscricoes, err := w.inscricaoRepo.ListApprovedStartingBetween(ctx, targetDay, windowEnd)
	if err != nil {
		return err
	}

	log.Printf("[ClassReminderWorker] Email %s (offset=%+d): found %d enrollments with class on %s (send_enabled=%v)",
		rule.ID, rule.TemporalOffsetDays, len(inscricoes), targetDay.Format("02/01/2006"), rule.SendEnabled)

	if !rule.SendEnabled {
		return nil
	}

	sent := 0
	for _, inscricao := range inscricoes {
		if inscricao.Curso == nil {
			log.Printf("[ClassReminderWorker] Enrollment %s has no course loaded, skipping", inscricao.ID)
			continue
		}

		if !w.markAsPendingSend(ctx, rule, inscricao.ID, targetDay) {
			continue
		}

		if err := w.sendForRule(ctx, rule, inscricao); err != nil {
			log.Printf("[ClassReminderWorker] Failed to enqueue %s for enrollment %s: %v", rule.ID, inscricao.ID, err)
			w.clearSentMark(ctx, rule, inscricao.ID, targetDay)
			continue
		}
		sent++
	}

	log.Printf("[ClassReminderWorker] Email %s: enqueued %d emails", rule.ID, sent)
	return nil
}

func (w *ClassReminderWorker) sendForRule(ctx context.Context, rule services.EmailCommunicationRule, inscricao *models.Inscricao) error {
	switch rule.ID {
	case "enrollment.class_reminder_d1":
		return w.emailNotifier.SendEnrollmentClassReminderEmail(ctx, inscricao, inscricao.Curso)
	default:
		return fmt.Errorf("no email sender configured for temporal rule %s", rule.ID)
	}
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func (w *ClassReminderWorker) sentKey(rule services.EmailCommunicationRule, inscricaoID uuid.UUID, day time.Time) string {
	return fmt.Sprintf("%s%s:%s:%s", classReminderSentKeyPrefix, rule.ID, inscricaoID.String(), day.Format("2006-01-02"))
}

func (w *ClassReminderWorker) markAsPendingSend(ctx context.Context, rule services.EmailCommunicationRule, inscricaoID uuid.UUID, day time.Time) bool {
	if w.redisClient == nil {
		return true
	}
	ok, err := w.redisClient.SetNX(ctx, w.sentKey(rule, inscricaoID, day), "1", classReminderDefaultTTL).Result()
	if err != nil {
		log.Printf("[ClassReminderWorker] Redis SetNX error for %s/%s: %v", rule.ID, inscricaoID, err)
		return true
	}
	return ok
}

func (w *ClassReminderWorker) clearSentMark(ctx context.Context, rule services.EmailCommunicationRule, inscricaoID uuid.UUID, day time.Time) {
	if w.redisClient == nil {
		return
	}
	if err := w.redisClient.Del(ctx, w.sentKey(rule, inscricaoID, day)).Err(); err != nil {
		log.Printf("[ClassReminderWorker] Failed to clear sent mark for %s/%s: %v", rule.ID, inscricaoID, err)
	}
}

func (w *ClassReminderWorker) tryAcquireLock(ctx context.Context) (string, bool, error) {
	if w.redisClient == nil {
		return "", true, nil
	}
	token := uuid.New().String()
	acquired, err := w.redisClient.SetNX(ctx, classReminderLockKey, token, w.syncInterval).Result()
	if err != nil {
		return "", false, err
	}
	if !acquired {
		return "", false, nil
	}
	return token, true, nil
}

func (w *ClassReminderWorker) releaseLock(ctx context.Context, token string) {
	if w.redisClient == nil || token == "" {
		return
	}
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)
	if err := script.Run(ctx, w.redisClient, []string{classReminderLockKey}, token).Err(); err != nil && err != redis.Nil {
		log.Printf("[ClassReminderWorker] Failed to release lock: %v", err)
	}
}
