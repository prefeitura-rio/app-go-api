package clients

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sony/gobreaker"
)

// ErrRMICircuitOpen is returned when the RMI circuit breaker rejects a request
// without calling the RMI.
var ErrRMICircuitOpen = errors.New("RMI circuit breaker open")

const (
	rmiMaxRetries = 2

	defaultRMIClientName           = "api"
	defaultRetryBaseDelay          = 100 * time.Millisecond
	defaultRetryMaxDelay           = time.Second
	defaultBreakerFailureThreshold = 5
	defaultBreakerOpenTimeout      = 30 * time.Second
	defaultBreakerInterval         = time.Minute
	defaultSecretariaCacheEntries  = 10000

	DefaultSecretariaCacheTTL = 2 * time.Minute
	MinSecretariaCacheTTL     = time.Minute
	MaxSecretariaCacheTTL     = 5 * time.Minute
)

var (
	rmiCircuitState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rmi_client_circuit_breaker_state",
		Help: "Estado do circuit breaker do cliente RMI: 0 = closed, 1 = half-open, 2 = open.",
	}, []string{"client"})
	rmiCircuitTransitionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rmi_client_circuit_breaker_transitions_total",
		Help: "Transições de estado do circuit breaker do cliente RMI.",
	}, []string{"client", "from", "to"})
	rmiCircuitRejectionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rmi_client_circuit_breaker_rejections_total",
		Help: "Requisições ao RMI recusadas pelo circuit breaker aberto.",
	}, []string{"client"})
	rmiRetriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rmi_client_retries_total",
		Help: "Novas tentativas de requisições ao RMI, por motivo.",
	}, []string{"client", "reason"})
	rmiSecretariaCacheTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rmi_client_secretaria_cache_total",
		Help: "Consultas ao cache de secretarias por CPF do cliente RMI (hit/miss).",
	}, []string{"client", "result"})
)

// RMIClientOption customizes an RMIClient.
type RMIClientOption func(*rmiClientOptions)

type rmiClientOptions struct {
	name               string
	retryBaseDelay     time.Duration
	retryMaxDelay      time.Duration
	breakerThreshold   uint32
	breakerOpenTimeout time.Duration
	secretariaCacheTTL time.Duration
}

// WithName sets the client label used in the Prometheus metrics.
func WithName(name string) RMIClientOption {
	return func(o *rmiClientOptions) {
		if name != "" {
			o.name = name
		}
	}
}

// WithRetryBackoff sets the base and maximum delay of the exponential backoff.
func WithRetryBackoff(base, maxDelay time.Duration) RMIClientOption {
	return func(o *rmiClientOptions) {
		if base > 0 {
			o.retryBaseDelay = base
		}
		if maxDelay >= o.retryBaseDelay {
			o.retryMaxDelay = maxDelay
		}
	}
}

// WithCircuitBreaker sets how many consecutive failures open the circuit and
// how long it stays open before letting one probe request through.
func WithCircuitBreaker(failureThreshold int, openTimeout time.Duration) RMIClientOption {
	return func(o *rmiClientOptions) {
		if failureThreshold > 0 {
			o.breakerThreshold = uint32(failureThreshold)
		}
		if openTimeout > 0 {
			o.breakerOpenTimeout = openTimeout
		}
	}
}

// WithSecretariaCacheTTL sets the TTL of the per-CPF secretaria cache, clamped to
// [MinSecretariaCacheTTL, MaxSecretariaCacheTTL]; zero or negative keeps the default.
func WithSecretariaCacheTTL(ttl time.Duration) RMIClientOption {
	return func(o *rmiClientOptions) {
		switch {
		case ttl <= 0:
			o.secretariaCacheTTL = DefaultSecretariaCacheTTL
		case ttl < MinSecretariaCacheTTL:
			o.secretariaCacheTTL = MinSecretariaCacheTTL
		case ttl > MaxSecretariaCacheTTL:
			o.secretariaCacheTTL = MaxSecretariaCacheTTL
		default:
			o.secretariaCacheTTL = ttl
		}
	}
}

func defaultRMIClientOptions() rmiClientOptions {
	return rmiClientOptions{
		name:               defaultRMIClientName,
		retryBaseDelay:     defaultRetryBaseDelay,
		retryMaxDelay:      defaultRetryMaxDelay,
		breakerThreshold:   defaultBreakerFailureThreshold,
		breakerOpenTimeout: defaultBreakerOpenTimeout,
		secretariaCacheTTL: DefaultSecretariaCacheTTL,
	}
}

func newRMICircuitBreaker(o rmiClientOptions) *gobreaker.CircuitBreaker {
	threshold := o.breakerThreshold
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        o.name,
		MaxRequests: 1,
		Interval:    defaultBreakerInterval,
		Timeout:     o.breakerOpenTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= threshold
		},
		IsSuccessful: func(err error) bool {
			// A caller giving up (client disconnect) says nothing about RMI health.
			return err == nil || errors.Is(err, context.Canceled)
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			rmiCircuitState.WithLabelValues(name).Set(circuitStateValue(to))
			rmiCircuitTransitionsTotal.WithLabelValues(name, from.String(), to.String()).Inc()
			log.Printf("[RMIClient] circuit breaker %q: %s -> %s", name, from, to)
		},
	})
}

func circuitStateValue(s gobreaker.State) float64 {
	switch s {
	case gobreaker.StateHalfOpen:
		return 1
	case gobreaker.StateOpen:
		return 2
	default:
		return 0
	}
}

// upstreamStatusError marks a 5xx response so the circuit breaker counts it as a
// failure; the response itself is still handed back to the caller.
type upstreamStatusError struct{ status int }

func (e *upstreamStatusError) Error() string {
	return fmt.Sprintf("RMI API returned status %d", e.status)
}

// do executes an idempotent request against the RMI through the circuit breaker,
// retrying up to rmiMaxRetries times on network errors and 502/503/504.
func (c *RMIClient) do(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		resp, err := c.attempt(req)
		reason, retry := c.retryReason(req, resp, err)
		if !retry || attempt >= rmiMaxRetries {
			return resp, err
		}
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		rmiRetriesTotal.WithLabelValues(c.name, reason).Inc()

		timer := time.NewTimer(c.backoffDelay(attempt))
		select {
		case <-req.Context().Done():
			timer.Stop()
			if err == nil {
				err = &upstreamStatusError{status: resp.StatusCode}
			}
			return nil, fmt.Errorf("%w (retry aborted: %v)", err, req.Context().Err())
		case <-timer.C:
		}
	}
}

func (c *RMIClient) attempt(req *http.Request) (*http.Response, error) {
	result, err := c.breaker.Execute(func() (interface{}, error) {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, redactURLError(err)
		}
		if resp.StatusCode >= http.StatusInternalServerError {
			return resp, &upstreamStatusError{status: resp.StatusCode}
		}
		return resp, nil
	})
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		rmiCircuitRejectionsTotal.WithLabelValues(c.name).Inc()
		return nil, fmt.Errorf("%w: %v", ErrRMICircuitOpen, err)
	}
	var statusErr *upstreamStatusError
	if err != nil && !errors.As(err, &statusErr) {
		return nil, err
	}
	resp, _ := result.(*http.Response)
	return resp, nil
}

func (c *RMIClient) retryReason(req *http.Request, resp *http.Response, err error) (string, bool) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return "", false
	}
	if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
		return "", false
	}
	if req.Context().Err() != nil {
		return "", false
	}
	if err != nil {
		if errors.Is(err, ErrRMICircuitOpen) {
			return "", false
		}
		// Timeouts are not retried: retrying a slow RMI would multiply the latency
		// the breaker is meant to cut.
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return "", false
		}
		return "network_error", true
	}
	if resp == nil {
		return "", false
	}
	switch resp.StatusCode {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return retryReasonForStatus(resp.StatusCode), true
	default:
		return "", false
	}
}

func retryReasonForStatus(status int) string {
	return fmt.Sprintf("status_%d", status)
}

// backoffDelay returns the wait before retry number attempt+1: exponential
// (base * 2^attempt, capped at the max delay) with "equal jitter", i.e. a random
// value in [d/2, d].
func (c *RMIClient) backoffDelay(attempt int) time.Duration {
	d := c.retryBaseDelay
	for i := 0; i < attempt && d < c.retryMaxDelay; i++ {
		d *= 2
	}
	if d > c.retryMaxDelay {
		d = c.retryMaxDelay
	}
	half := d / 2
	return half + time.Duration(rand.Int64N(int64(d-half)+1))
}

var longDigitRun = regexp.MustCompile(`\d{11,}`)

// redactURLError strips CPF/CNPJ-like digit runs from the URL embedded in
// *url.Error so request failures can be logged without personal data.
func redactURLError(err error) error {
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return err
	}
	return &url.Error{Op: urlErr.Op, URL: longDigitRun.ReplaceAllString(urlErr.URL, "[redacted]"), Err: urlErr.Err}
}

type secretariaCacheEntry struct {
	cdUAs     []string
	expiresAt time.Time
}

// secretariaCache is a small in-memory TTL cache of CPF -> cd_uas.
type secretariaCache struct {
	mu         sync.Mutex
	ttl        time.Duration
	maxEntries int
	now        func() time.Time
	entries    map[string]secretariaCacheEntry
}

func newSecretariaCache(ttl time.Duration, maxEntries int) *secretariaCache {
	return &secretariaCache{
		ttl:        ttl,
		maxEntries: maxEntries,
		now:        time.Now,
		entries:    make(map[string]secretariaCacheEntry),
	}
}

func (s *secretariaCache) get(cpf string) ([]string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[cpf]
	if !ok {
		return nil, false
	}
	if !s.now().Before(entry.expiresAt) {
		delete(s.entries, cpf)
		return nil, false
	}
	return append([]string{}, entry.cdUAs...), true
}

func (s *secretariaCache) set(cpf string, cdUAs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if _, exists := s.entries[cpf]; !exists && len(s.entries) >= s.maxEntries {
		for k, e := range s.entries {
			if !now.Before(e.expiresAt) {
				delete(s.entries, k)
			}
		}
		if len(s.entries) >= s.maxEntries {
			s.entries = make(map[string]secretariaCacheEntry)
		}
	}
	s.entries[cpf] = secretariaCacheEntry{cdUAs: append([]string{}, cdUAs...), expiresAt: now.Add(s.ttl)}
}
