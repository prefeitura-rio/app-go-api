package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailCommunicationRules_EventEmailsHaveNoTemporalOffsetSemantics(t *testing.T) {
	rules := EmailCommunicationRules()
	require.NotEmpty(t, rules)

	eventCount := 0
	temporalCount := 0
	for _, rule := range rules {
		switch rule.Kind {
		case EmailTriggerEvent:
			eventCount++
			assert.Empty(t, rule.TemporalAnchor)
		case EmailTriggerTemporal:
			temporalCount++
			assert.NotEmpty(t, rule.TemporalAnchor)
		default:
			t.Fatalf("unknown trigger kind: %s", rule.Kind)
		}
	}

	assert.Greater(t, eventCount, temporalCount, "most emails should be event-driven")
	assert.Equal(t, 1, temporalCount, "only Lembrete D-1 is temporal today")
}

func TestActiveTemporalClassStartRules_OnlyD1(t *testing.T) {
	rules := ActiveTemporalClassStartRules()
	require.Len(t, rules, 1)
	assert.Equal(t, "enrollment.class_reminder_d1", rules[0].ID)
	assert.Equal(t, -1, rules[0].TemporalOffsetDays)
	assert.True(t, rules[0].SendEnabled)
}

func TestEmailCommunicationRules_UniqueIDs(t *testing.T) {
	seen := map[string]struct{}{}
	for _, rule := range EmailCommunicationRules() {
		if _, ok := seen[rule.ID]; ok {
			t.Fatalf("duplicate email rule ID: %s", rule.ID)
		}
		seen[rule.ID] = struct{}{}
	}
}
