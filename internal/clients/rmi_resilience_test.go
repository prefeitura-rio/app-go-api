package clients

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fastRetry keeps retry backoff negligible in tests.
var fastRetry = WithRetryBackoff(time.Millisecond, 2*time.Millisecond)

var clientNameSeq atomic.Int64

// uniqueClientName keeps metric label sets independent across tests and -count runs.
func uniqueClientName(t *testing.T, prefix string) string {
	return fmt.Sprintf("%s%s-%d", prefix, t.Name(), clientNameSeq.Add(1))
}

func newCountingServer(t *testing.T, handler func(call int32, w http.ResponseWriter, r *http.Request)) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		handler(n, w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func writeSecretarias(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

func TestRMIClient_Retry_TransientStatusThenSuccess(t *testing.T) {
	for _, status := range []int{http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			name := uniqueClientName(t, "retry-")
			srv, calls := newCountingServer(t, func(call int32, w http.ResponseWriter, _ *http.Request) {
				if call == 1 {
					w.WriteHeader(status)
					return
				}
				writeSecretarias(w, `{"cpf":"12345678900","cd_uas":["1000"]}`)
			})
			client := NewRMIClient(srv.URL, 5*time.Second, WithName(name), fastRetry)

			cdUAs, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")

			require.NoError(t, err)
			assert.Equal(t, []string{"1000"}, cdUAs)
			assert.Equal(t, int32(2), calls.Load())
			assert.Equal(t, float64(1), testutil.ToFloat64(rmiRetriesTotal.WithLabelValues(name, retryReasonForStatus(status))))
		})
	}
}

func TestRMIClient_Retry_AtMostTwoRetries(t *testing.T) {
	srv, calls := newCountingServer(t, func(_ int32, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(uniqueClientName(t, "retry-max-")), fastRetry)

	_, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 503")
	assert.Equal(t, int32(3), calls.Load(), "1 attempt + 2 retries")
}

func TestRMIClient_Retry_NotOnNonTransientStatus(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv, calls := newCountingServer(t, func(_ int32, w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			})
			client := NewRMIClient(srv.URL, 5*time.Second, WithName(uniqueClientName(t, "noretry-")), fastRetry)

			_, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")

			require.Error(t, err)
			assert.Equal(t, int32(1), calls.Load())
		})
	}
}

func TestRMIClient_Retry_NetworkError(t *testing.T) {
	srv, calls := newCountingServer(t, func(call int32, w http.ResponseWriter, _ *http.Request) {
		if call <= 2 {
			hj, ok := w.(http.Hijacker)
			require.True(t, ok)
			conn, _, err := hj.Hijack()
			require.NoError(t, err)
			_ = conn.Close()
			return
		}
		writeSecretarias(w, `{"cpf":"12345678900","cd_uas":["2000"]}`)
	})
	name := uniqueClientName(t, "retry-net-")
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(name), fastRetry)

	cdUAs, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")

	require.NoError(t, err)
	assert.Equal(t, []string{"2000"}, cdUAs)
	assert.Equal(t, int32(3), calls.Load())
	assert.Equal(t, float64(2), testutil.ToFloat64(rmiRetriesTotal.WithLabelValues(name, "network_error")))
}

func TestRMIClient_NetworkError_DoesNotLeakCPFInError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	client := NewRMIClient("http://"+addr, time.Second, WithName(uniqueClientName(t, "leak-")), fastRetry)
	_, err = client.GetCPFSecretarias(context.Background(), "token", "12345678900")

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "12345678900")
}

func TestRMIClient_Retry_StopsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	srv, calls := newCountingServer(t, func(_ int32, w http.ResponseWriter, _ *http.Request) {
		cancel()
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(uniqueClientName(t, "retry-ctx-")), WithRetryBackoff(200*time.Millisecond, time.Second))

	_, err := client.GetCPFSecretarias(ctx, "token", "12345678900")

	require.Error(t, err)
	assert.Equal(t, int32(1), calls.Load())
}

func TestRMIClient_BackoffDelay_ExponentialWithJitter(t *testing.T) {
	client := NewRMIClient("http://x", time.Second, WithRetryBackoff(100*time.Millisecond, 300*time.Millisecond))
	for i := 0; i < 50; i++ {
		d0 := client.backoffDelay(0)
		assert.GreaterOrEqual(t, d0, 50*time.Millisecond)
		assert.LessOrEqual(t, d0, 100*time.Millisecond)

		d1 := client.backoffDelay(1)
		assert.GreaterOrEqual(t, d1, 100*time.Millisecond)
		assert.LessOrEqual(t, d1, 200*time.Millisecond)

		d5 := client.backoffDelay(5)
		assert.GreaterOrEqual(t, d5, 150*time.Millisecond)
		assert.LessOrEqual(t, d5, 300*time.Millisecond, "capped at max delay")
	}
}

func TestRMIClient_CircuitBreaker_OpensAfterConsecutiveFailures(t *testing.T) {
	srv, calls := newCountingServer(t, func(_ int32, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	name := uniqueClientName(t, "cb-open-")
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(name), fastRetry, WithCircuitBreaker(2, time.Minute))

	assert.Equal(t, float64(0), testutil.ToFloat64(rmiCircuitState.WithLabelValues(name)), "starts closed")

	for i := 0; i < 2; i++ {
		_, err := client.GetOrgao(context.Background(), "1000")
		require.Error(t, err)
	}
	assert.Equal(t, int32(2), calls.Load())
	assert.Equal(t, float64(2), testutil.ToFloat64(rmiCircuitState.WithLabelValues(name)), "open")

	_, err := client.GetOrgao(context.Background(), "1000")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrRMICircuitOpen), "got %v", err)
	assert.Equal(t, int32(2), calls.Load(), "open circuit must not reach the RMI")
	assert.Equal(t, float64(1), testutil.ToFloat64(rmiCircuitRejectionsTotal.WithLabelValues(name)))
	assert.Equal(t, float64(1), testutil.ToFloat64(rmiCircuitTransitionsTotal.WithLabelValues(name, "closed", "open")))
}

func TestRMIClient_CircuitBreaker_OpenCircuitIsNotRetried(t *testing.T) {
	srv, calls := newCountingServer(t, func(_ int32, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(uniqueClientName(t, "cb-noretry-")), fastRetry, WithCircuitBreaker(1, time.Minute))

	_, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrRMICircuitOpen), "got %v", err)
	assert.Equal(t, int32(1), calls.Load(), "retries stop as soon as the circuit opens")
}

func TestRMIClient_CircuitBreaker_HalfOpenThenClosesOnSuccess(t *testing.T) {
	var healthy atomic.Bool
	srv, _ := newCountingServer(t, func(_ int32, w http.ResponseWriter, _ *http.Request) {
		if !healthy.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeSecretarias(w, `{"cpf":"12345678900","cd_uas":[]}`)
	})
	name := uniqueClientName(t, "cb-close-")
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(name), fastRetry, WithCircuitBreaker(1, 50*time.Millisecond))

	_, err := client.GetOrgao(context.Background(), "1000")
	require.Error(t, err)
	assert.Equal(t, float64(2), testutil.ToFloat64(rmiCircuitState.WithLabelValues(name)))

	healthy.Store(true)
	time.Sleep(80 * time.Millisecond)

	_, err = client.GetCPFSecretarias(context.Background(), "token", "12345678900")
	require.NoError(t, err)
	assert.Equal(t, float64(0), testutil.ToFloat64(rmiCircuitState.WithLabelValues(name)))
	assert.Equal(t, float64(1), testutil.ToFloat64(rmiCircuitTransitionsTotal.WithLabelValues(name, "open", "half-open")))
	assert.Equal(t, float64(1), testutil.ToFloat64(rmiCircuitTransitionsTotal.WithLabelValues(name, "half-open", "closed")))
}

func TestRMIClient_CircuitBreaker_ClientErrorsDoNotTrip(t *testing.T) {
	srv, calls := newCountingServer(t, func(_ int32, w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	name := uniqueClientName(t, "cb-4xx-")
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(name), fastRetry, WithCircuitBreaker(2, time.Minute))

	for i := 0; i < 5; i++ {
		_, _ = client.GetOrgao(context.Background(), "1000")
	}

	assert.Equal(t, int32(5), calls.Load())
	assert.Equal(t, float64(0), testutil.ToFloat64(rmiCircuitState.WithLabelValues(name)))
}

func TestRMIClient_SecretariaCache_HitMissAndExpiry(t *testing.T) {
	srv, calls := newCountingServer(t, func(_ int32, w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/99999999999") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		writeSecretarias(w, `{"cpf":"12345678900","cd_uas":["1000","2000"]}`)
	})
	name := uniqueClientName(t, "cache-")
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(name), fastRetry, WithSecretariaCacheTTL(2*time.Minute))
	now := time.Now()
	client.secretariaCache.now = func() time.Time { return now }

	first, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")
	require.NoError(t, err)
	first[0] = "mutated"

	second, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")
	require.NoError(t, err)
	assert.Equal(t, []string{"1000", "2000"}, second, "cached value must not be shared with callers")
	assert.Equal(t, int32(1), calls.Load(), "second call served from cache")

	empty, err := client.GetCPFSecretarias(context.Background(), "token", "99999999999")
	require.NoError(t, err)
	assert.Empty(t, empty)
	_, _ = client.GetCPFSecretarias(context.Background(), "token", "99999999999")
	assert.Equal(t, int32(2), calls.Load(), "empty (404) result is cached per CPF")

	now = now.Add(2*time.Minute + time.Second)
	_, err = client.GetCPFSecretarias(context.Background(), "token", "12345678900")
	require.NoError(t, err)
	assert.Equal(t, int32(3), calls.Load(), "expired entry goes back to the RMI")

	assert.Equal(t, float64(2), testutil.ToFloat64(rmiSecretariaCacheTotal.WithLabelValues(name, "hit")))
	assert.Equal(t, float64(3), testutil.ToFloat64(rmiSecretariaCacheTotal.WithLabelValues(name, "miss")))
}

func TestRMIClient_SecretariaCache_ErrorsAreNotCached(t *testing.T) {
	srv, calls := newCountingServer(t, func(call int32, w http.ResponseWriter, _ *http.Request) {
		if call == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeSecretarias(w, `{"cpf":"12345678900","cd_uas":["1000"]}`)
	})
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(uniqueClientName(t, "cache-err-")), fastRetry)

	_, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")
	require.Error(t, err)

	cdUAs, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")
	require.NoError(t, err)
	assert.Equal(t, []string{"1000"}, cdUAs)
	assert.Equal(t, int32(2), calls.Load())
}

func TestRMIClient_SecretariaCacheTTL_ClampedToOneToFiveMinutes(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want time.Duration
	}{
		{"zero uses default", 0, DefaultSecretariaCacheTTL},
		{"negative uses default", -time.Second, DefaultSecretariaCacheTTL},
		{"below min", 10 * time.Second, MinSecretariaCacheTTL},
		{"within range", 3 * time.Minute, 3 * time.Minute},
		{"above max", time.Hour, MaxSecretariaCacheTTL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewRMIClient("http://x", time.Second, WithSecretariaCacheTTL(tt.in))
			assert.Equal(t, tt.want, client.secretariaCache.ttl)
		})
	}
	assert.Equal(t, DefaultSecretariaCacheTTL, NewRMIClient("http://x", time.Second).secretariaCache.ttl)
}

func TestRMIClient_SecretariaCache_BoundedSize(t *testing.T) {
	c := newSecretariaCache(time.Minute, 2)
	c.set("a", []string{"1"})
	c.set("b", []string{"2"})
	c.set("c", []string{"3"})

	assert.LessOrEqual(t, len(c.entries), 2)
	got, ok := c.get("c")
	assert.True(t, ok)
	assert.Equal(t, []string{"3"}, got)
}

func TestRMIClient_MetricsExposedByPromHandler(t *testing.T) {
	srv, _ := newCountingServer(t, func(call int32, w http.ResponseWriter, _ *http.Request) {
		if call == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writeSecretarias(w, `{"cpf":"12345678900","cd_uas":["1000"]}`)
	})
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(uniqueClientName(t, "metrics-")), fastRetry)
	_, err := client.GetCPFSecretarias(context.Background(), "token", "12345678900")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := rec.Body.String()
	assert.Equal(t, http.StatusOK, rec.Code)
	for _, metric := range []string{
		"rmi_client_circuit_breaker_state",
		"rmi_client_retries_total",
		"rmi_client_circuit_breaker_rejections_total",
		"rmi_client_secretaria_cache_total",
	} {
		assert.Contains(t, body, metric)
	}
}

func TestRMIClient_CircuitBreaker_CancelledHalfOpenProbeDoesNotClose(t *testing.T) {
	var healthy atomic.Bool
	release := make(chan struct{})
	srv, _ := newCountingServer(t, func(_ int32, w http.ResponseWriter, r *http.Request) {
		if !healthy.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		select {
		case <-r.Context().Done():
		case <-release:
		}
		w.WriteHeader(http.StatusInternalServerError)
	})
	t.Cleanup(func() { close(release) })
	name := uniqueClientName(t, "cb-cancel-")
	client := NewRMIClient(srv.URL, 5*time.Second, WithName(name), fastRetry, WithCircuitBreaker(1, 50*time.Millisecond))

	_, err := client.GetOrgao(context.Background(), "1000")
	require.Error(t, err)
	require.Equal(t, float64(2), testutil.ToFloat64(rmiCircuitState.WithLabelValues(name)))

	healthy.Store(true)
	time.Sleep(80 * time.Millisecond)

	cancelCtx, cancelProbe := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancelProbe()
	}()
	_, err = client.GetOrgao(cancelCtx, "1000")
	require.Error(t, err)

	assert.NotEqual(t, float64(0), testutil.ToFloat64(rmiCircuitState.WithLabelValues(name)),
		"a probe cancelled by the caller must not close the circuit")
}
