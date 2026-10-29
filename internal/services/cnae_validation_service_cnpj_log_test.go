package services_test

import (
	"bytes"
	"context"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// cnpjLogSink is a concurrency-safe log sink used to inspect what the
// service writes through the standard logger.
type cnpjLogSink struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *cnpjLogSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *cnpjLogSink) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func captureCNPJLog(t *testing.T) *cnpjLogSink {
	t.Helper()
	sink := &cnpjLogSink{}
	previous := log.Writer()
	log.SetOutput(sink)
	t.Cleanup(func() { log.SetOutput(previous) })
	return sink
}

// Synthetic CNPJs (never real data). The token carries a synthetic CPF in
// preferred_username.
const (
	cnpjLogToken          = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJwcmVmZXJyZWRfdXNlcm5hbWUiOiIxMjM0NTY3ODkwMCJ9.fake"
	cnpjLogOwnedDigits    = "11222333000181"
	cnpjLogOwnedFormatted = "11.222.333/0001-81"
	cnpjLogOwnedRoot      = "11222333"
	cnpjLogOwnedMasked    = "11**********81"
	cnpjLogOtherDigits    = "99888777000166"
	cnpjLogOtherRoot      = "99888777"
	cnpjLogOtherMasked    = "99**********66"
)

func TestCNAEValidationService_Logs_MaskCNPJ(t *testing.T) {
	ownedEntity := models.LegalEntity{CNPJ: cnpjLogOwnedDigits, CNAEFiscal: "4110700"}

	tests := []struct {
		name       string
		entities   []models.LegalEntity
		cnpj       string
		cnaes      []string
		wantErr    bool
		wantMarker string
		wantMasked string
		forbidden  []string
	}{
		{
			name:       "success",
			entities:   []models.LegalEntity{ownedEntity},
			cnpj:       cnpjLogOwnedDigits,
			cnaes:      []string{"4110700"},
			wantMarker: "[CNAE_VALIDATION_SUCCESS]",
			wantMasked: cnpjLogOwnedMasked,
			forbidden:  []string{cnpjLogOwnedDigits, cnpjLogOwnedFormatted, cnpjLogOwnedRoot},
		},
		{
			name:       "success with formatted cnpj",
			entities:   []models.LegalEntity{ownedEntity},
			cnpj:       cnpjLogOwnedFormatted,
			cnaes:      []string{"4110700"},
			wantMarker: "[CNAE_VALIDATION_SUCCESS]",
			wantMasked: cnpjLogOwnedMasked,
			forbidden:  []string{cnpjLogOwnedDigits, cnpjLogOwnedFormatted, "11.222.333"},
		},
		{
			name:       "no cnpjs",
			entities:   []models.LegalEntity{},
			cnpj:       cnpjLogOwnedDigits,
			cnaes:      []string{"4110700"},
			wantErr:    true,
			wantMarker: "Reason=no_cnpjs",
			wantMasked: cnpjLogOwnedMasked,
			forbidden:  []string{cnpjLogOwnedDigits, cnpjLogOwnedRoot},
		},
		{
			name:       "cnpj not owned",
			entities:   []models.LegalEntity{ownedEntity},
			cnpj:       cnpjLogOtherDigits,
			cnaes:      []string{"4110700"},
			wantErr:    true,
			wantMarker: "Reason=ownership",
			wantMasked: cnpjLogOtherMasked,
			forbidden:  []string{cnpjLogOtherDigits, cnpjLogOtherRoot},
		},
		{
			name:       "cnae mismatch",
			entities:   []models.LegalEntity{ownedEntity},
			cnpj:       cnpjLogOwnedDigits,
			cnaes:      []string{"9999999"},
			wantErr:    true,
			wantMarker: "Reason=cnae_mismatch",
			wantMasked: cnpjLogOwnedMasked,
			forbidden:  []string{cnpjLogOwnedDigits, cnpjLogOwnedRoot},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := captureCNPJLog(t)
			svc := services.NewCNAEValidationService(&MockRMIClient{entities: tt.entities}, NewMockCache())

			err := svc.ValidatePropostaForCNAE(context.Background(), cnpjLogToken, tt.cnpj, tt.cnaes)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error result: got %v, wantErr %v", err, tt.wantErr)
			}

			out := sink.String()
			if !strings.Contains(out, tt.wantMarker) {
				t.Errorf("expected log to contain %q, got:\n%s", tt.wantMarker, out)
			}
			if !strings.Contains(out, "CNPJ="+tt.wantMasked) {
				t.Errorf("expected log to contain the masked CNPJ %q, got:\n%s", tt.wantMasked, out)
			}
			for _, raw := range tt.forbidden {
				if strings.Contains(out, raw) {
					t.Errorf("log must not contain %q, got:\n%s", raw, out)
				}
			}
		})
	}
}
