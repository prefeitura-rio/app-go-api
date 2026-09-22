package workers

import (
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// TemporalEmailJob is one email's exclusive temporal trigger evaluated by the worker.
// It is never reused by other emails — each communication rule owns its offset.
type TemporalEmailJob struct {
	Rule services.EmailCommunicationRule
}

// temporalClassStartJobs builds worker jobs from the communication catalog
// (only class_start temporal rules).
func temporalClassStartJobs() []TemporalEmailJob {
	rules := services.ActiveTemporalClassStartRules()
	jobs := make([]TemporalEmailJob, 0, len(rules))
	for _, rule := range rules {
		jobs = append(jobs, TemporalEmailJob{Rule: rule})
	}
	return jobs
}

// targetClassDateForOffset returns the class-start calendar day that matches
// the given offset today (OffsetDays: -1 = tomorrow, +2 = two days ago, …).
func targetClassDateForOffset(now time.Time, offsetDays int) time.Time {
	return startOfDay(now).AddDate(0, 0, -offsetDays)
}
