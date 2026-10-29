package services_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// cnaeLogBuffer is a concurrency-safe log sink used to inspect what the
// service writes through the standard logger.
type cnaeLogBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *cnaeLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *cnaeLogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func captureCNAELog(t *testing.T) *cnaeLogBuffer {
	t.Helper()
	sink := &cnaeLogBuffer{}
	previous := log.Writer()
	log.SetOutput(sink)
	t.Cleanup(func() { log.SetOutput(previous) })
	return sink
}

// cnaeLogCache lets each cache operation fail independently.
type cnaeLogCache struct {
	data   map[string][]models.LegalEntity
	getErr error
	setErr error
}

func (c *cnaeLogCache) Get(_ context.Context, cpf string) ([]models.LegalEntity, error) {
	if c.getErr != nil {
		return nil, c.getErr
	}
	entities, ok := c.data[cpf]
	if !ok {
		return nil, nil
	}
	return entities, nil
}

func (c *cnaeLogCache) Set(_ context.Context, cpf string, entities []models.LegalEntity) error {
	if c.setErr != nil {
		return c.setErr
	}
	if c.data == nil {
		c.data = make(map[string][]models.LegalEntity)
	}
	c.data[cpf] = entities
	return nil
}

// Synthetic CPF carried in preferred_username of the token below.
const (
	cnaeLogSyntheticCPF = "12345678900"
	cnaeLogMaskedCPF    = "123******00"
	cnaeLogToken        = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"
)

func TestCNAEValidationService_Logs_MaskCPF(t *testing.T) {
	ownedEntity := models.LegalEntity{CNPJ: "12345678000190", CNAEFiscal: "4110700"}

	tests := []struct {
		name        string
		rmi         *MockRMIClient
		cache       *cnaeLogCache
		cnpj        string
		cnaes       []string
		ownership   bool
		wantErr     bool
		wantMarkers []string
	}{
		{
			name:        "success logs cache miss, cache set and success",
			rmi:         &MockRMIClient{entities: []models.LegalEntity{ownedEntity}},
			cache:       &cnaeLogCache{},
			cnpj:        "12345678000190",
			cnaes:       []string{"4110700"},
			wantMarkers: []string{"[CACHE_MISS]", "[CACHE_SET]", "[CNAE_VALIDATION_SUCCESS]"},
		},
		{
			name: "cache hit",
			rmi:  &MockRMIClient{},
			cache: &cnaeLogCache{data: map[string][]models.LegalEntity{
				cnaeLogSyntheticCPF: {ownedEntity},
			}},
			cnpj:        "12345678000190",
			cnaes:       []string{"4110700"},
			wantMarkers: []string{"[CACHE_HIT]"},
		},
		{
			name:        "cache get and set errors",
			rmi:         &MockRMIClient{entities: []models.LegalEntity{ownedEntity}},
			cache:       &cnaeLogCache{getErr: errors.New("redis down"), setErr: errors.New("redis down")},
			cnpj:        "12345678000190",
			cnaes:       []string{"4110700"},
			wantMarkers: []string{"[CACHE_ERROR] Failed to get", "[CACHE_ERROR] Failed to set"},
		},
		{
			name:        "rmi error",
			rmi:         &MockRMIClient{err: errors.New("rmi unavailable")},
			cache:       &cnaeLogCache{},
			cnpj:        "12345678000190",
			cnaes:       []string{"4110700"},
			wantErr:     true,
			wantMarkers: []string{"[CNAE_VALIDATION_ERROR]"},
		},
		{
			name:        "no cnpjs",
			rmi:         &MockRMIClient{entities: []models.LegalEntity{}},
			cache:       &cnaeLogCache{},
			cnpj:        "12345678000190",
			cnaes:       []string{"4110700"},
			wantErr:     true,
			wantMarkers: []string{"Reason=no_cnpjs"},
		},
		{
			name:        "cnpj not owned",
			rmi:         &MockRMIClient{entities: []models.LegalEntity{ownedEntity}},
			cache:       &cnaeLogCache{},
			cnpj:        "99999999000199",
			cnaes:       []string{"4110700"},
			wantErr:     true,
			wantMarkers: []string{"Reason=ownership"},
		},
		{
			name:        "cnae mismatch",
			rmi:         &MockRMIClient{entities: []models.LegalEntity{ownedEntity}},
			cache:       &cnaeLogCache{},
			cnpj:        "12345678000190",
			cnaes:       []string{"9999999"},
			wantErr:     true,
			wantMarkers: []string{"Reason=cnae_mismatch"},
		},
		{
			name:        "ownership check rmi error",
			rmi:         &MockRMIClient{err: errors.New("rmi unavailable")},
			cache:       &cnaeLogCache{},
			cnpj:        "12345678000190",
			ownership:   true,
			wantErr:     true,
			wantMarkers: []string{"[CNPJ_OWNERSHIP_ERROR]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := captureCNAELog(t)
			svc := services.NewCNAEValidationService(tt.rmi, tt.cache)

			var err error
			if tt.ownership {
				_, err = svc.CheckCNPJOwnership(context.Background(), cnaeLogToken, cnaeLogSyntheticCPF, tt.cnpj)
			} else {
				err = svc.ValidatePropostaForCNAE(context.Background(), cnaeLogToken, tt.cnpj, tt.cnaes)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error result: got %v, wantErr %v", err, tt.wantErr)
			}

			out := sink.String()
			for _, marker := range tt.wantMarkers {
				if !strings.Contains(out, marker) {
					t.Errorf("expected log to contain %q, got:\n%s", marker, out)
				}
			}
			if strings.Contains(out, cnaeLogSyntheticCPF) {
				t.Errorf("log must not contain the raw CPF, got:\n%s", out)
			}
			if !strings.Contains(out, cnaeLogMaskedCPF) {
				t.Errorf("expected log to contain the masked CPF %q, got:\n%s", cnaeLogMaskedCPF, out)
			}
		})
	}
}
