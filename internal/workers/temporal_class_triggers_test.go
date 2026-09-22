package workers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/services"
)

func TestTemporalClassStartJobs_DoNotIncludeEventEmails(t *testing.T) {
	jobs := temporalClassStartJobs()
	for _, job := range jobs {
		assert.Equal(t, services.EmailTriggerTemporal, job.Rule.Kind)
		assert.Equal(t, services.EmailAnchorClassStart, job.Rule.TemporalAnchor)
	}
}

func TestTemporalClassStartJobs_EachEmailOwnsOffset(t *testing.T) {
	jobs := temporalClassStartJobs()
	require.NotEmpty(t, jobs)

	seen := map[string]struct{}{}
	for _, job := range jobs {
		_, dup := seen[job.Rule.ID]
		assert.False(t, dup, "email rule ID must be unique: %s", job.Rule.ID)
		seen[job.Rule.ID] = struct{}{}
	}
}
