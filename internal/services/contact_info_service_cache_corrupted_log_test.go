package services_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// corruptedCacheLogLine returns the [CONTACT_INFO_CACHE_CORRUPTED] line of the
// captured logs, or an empty string when it was not logged.
func corruptedCacheLogLine(logs string) string {
	for _, line := range strings.Split(logs, "\n") {
		if strings.Contains(line, "[CONTACT_INFO_CACHE_CORRUPTED]") {
			return line
		}
	}
	return ""
}

func TestContactInfoService_GetCNPJOwnerContactInfo_CorruptedCache_LogsUnmarshalError(t *testing.T) {
	const corrupted = "{not-json"

	var probe models.CNPJOwnerInfo
	unmarshalErr := json.Unmarshal([]byte(corrupted), &probe)
	require.Error(t, unmarshalErr, "the cached value must not be valid JSON")

	env := newContactLogEnv(t, false)
	cacheKey := fmt.Sprintf("cnpj:owner:contact:%s", cnpjLogOwnedDigits)
	require.NoError(t, env.redis.Set(context.Background(), cacheKey, corrupted, time.Minute).Err())

	sink := captureCNPJLog(t)

	info, err := env.service.GetCNPJOwnerContactInfo(context.Background(), cnpjLogOwnedDigits)

	// Behavior is unchanged: falls back to RMI and rewrites the cache.
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "titular@example.com", info.EmailPessoaFisica)

	cached, err := env.redis.Get(context.Background(), cacheKey).Result()
	require.NoError(t, err)
	var rewritten models.CNPJOwnerInfo
	require.NoError(t, json.Unmarshal([]byte(cached), &rewritten), "cache must be rewritten with valid JSON")
	assert.Equal(t, "titular@example.com", rewritten.EmailPessoaFisica)

	// The corrupted-cache log carries the real Unmarshal error, with the CNPJ masked.
	logs := sink.String()
	line := corruptedCacheLogLine(logs)
	require.NotEmpty(t, line, "expected a [CONTACT_INFO_CACHE_CORRUPTED] log line")
	assert.Contains(t, line, "error: "+unmarshalErr.Error())
	assert.NotContains(t, line, "<nil>")
	assert.Contains(t, line, cnpjLogOwnedMasked)
	assert.NotContains(t, logs, cnpjLogOwnedDigits)
	assert.Contains(t, logs, "[CONTACT_INFO_CACHE_MISS]")
	assert.Contains(t, logs, "[CONTACT_INFO_CACHED]")
}
