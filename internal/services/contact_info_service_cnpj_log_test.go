package services_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// contactLogRMIHandler answers the RMI endpoints used by ContactInfoService
// with synthetic data. When notFound is true, the legal entity lookup returns
// 404 so the service fails.
func contactLogRMIHandler(t *testing.T, notFound bool) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/v1/legal-entity/"):
			if notFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			entity := models.LegalEntityDetails{CNPJ: cnpjLogOwnedDigits, RazaoSocial: "Empresa Sintetica"}
			entity.Socios = []struct {
				CPFSocio  string `json:"cpf_socio"`
				NomeSocio string `json:"nome_socio_estrangeiro"`
			}{{CPFSocio: "12345678900", NomeSocio: "Titular Sintetico"}}
			_ = json.NewEncoder(w).Encode(entity)
		case strings.Contains(r.URL.Path, "/v1/citizen/"):
			citizen := models.CitizenContactInfo{CPF: "12345678900", Nome: "Titular Sintetico"}
			citizen.Email.Indicador = true
			citizen.Email.Principal.Valor = "titular@example.com"
			_ = json.NewEncoder(w).Encode(citizen)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

type contactLogEnv struct {
	service *services.ContactInfoService
	redis   *redis.Client
	mr      *miniredis.Miniredis
}

func newContactLogEnv(t *testing.T, notFound bool) *contactLogEnv {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	server := httptest.NewServer(contactLogRMIHandler(t, notFound))
	t.Cleanup(server.Close)

	cfg := &config.AppConfig{Cache: config.CacheSettings{ContactInfoTTL: 5 * time.Minute}}
	service := services.NewContactInfoService(
		clients.NewRMIClient(server.URL, 5*time.Second),
		&mockTokenManager{token: "test-token"},
		redisClient,
		cfg,
	)
	return &contactLogEnv{service: service, redis: redisClient, mr: mr}
}

func assertCNPJMaskedInLog(t *testing.T, logs string, markers []string) {
	t.Helper()
	for _, marker := range markers {
		assert.Contains(t, logs, marker, "expected log marker %s", marker)
	}
	assert.Contains(t, logs, cnpjLogOwnedMasked, "masked CNPJ must appear in the logs")
	for _, forbidden := range []string{cnpjLogOwnedDigits, cnpjLogOwnedFormatted, cnpjLogOwnedRoot} {
		assert.NotContains(t, logs, forbidden, "CNPJ must not appear unmasked in the logs")
	}
}

func TestContactInfoService_Logs_MaskCNPJ(t *testing.T) {
	tests := []struct {
		name     string
		cnpj     string
		notFound bool
		prepare  func(t *testing.T, env *contactLogEnv, cnpj string)
		run      func(t *testing.T, env *contactLogEnv, cnpj string)
		markers  []string
	}{
		{
			name: "cache hit",
			cnpj: cnpjLogOwnedDigits,
			prepare: func(t *testing.T, env *contactLogEnv, cnpj string) {
				data, err := json.Marshal(&models.CNPJOwnerInfo{CNPJ: cnpj, EmailPessoaFisica: "titular@example.com"})
				require.NoError(t, err)
				require.NoError(t, env.redis.Set(context.Background(), fmt.Sprintf("cnpj:owner:contact:%s", cnpj), data, time.Minute).Err())
			},
			run: func(t *testing.T, env *contactLogEnv, cnpj string) {
				_, err := env.service.GetCNPJOwnerContactInfo(context.Background(), cnpj)
				require.NoError(t, err)
			},
			markers: []string{"[CONTACT_INFO_CACHE_HIT]"},
		},
		{
			name: "cache miss then cached",
			cnpj: cnpjLogOwnedDigits,
			run: func(t *testing.T, env *contactLogEnv, cnpj string) {
				_, err := env.service.GetCNPJOwnerContactInfo(context.Background(), cnpj)
				require.NoError(t, err)
			},
			markers: []string{"[CONTACT_INFO_CACHE_MISS]", "[CONTACT_INFO_CACHED]"},
		},
		{
			name: "cache miss with formatted cnpj",
			cnpj: cnpjLogOwnedFormatted,
			run: func(t *testing.T, env *contactLogEnv, cnpj string) {
				_, err := env.service.GetCNPJOwnerContactInfo(context.Background(), cnpj)
				require.NoError(t, err)
			},
			markers: []string{"[CONTACT_INFO_CACHE_MISS]", "[CONTACT_INFO_CACHED]"},
		},
		{
			name: "corrupted cache",
			cnpj: cnpjLogOwnedDigits,
			prepare: func(t *testing.T, env *contactLogEnv, cnpj string) {
				require.NoError(t, env.redis.Set(context.Background(), fmt.Sprintf("cnpj:owner:contact:%s", cnpj), "{not-json", time.Minute).Err())
			},
			run: func(t *testing.T, env *contactLogEnv, cnpj string) {
				_, err := env.service.GetCNPJOwnerContactInfo(context.Background(), cnpj)
				require.NoError(t, err)
			},
			markers: []string{"[CONTACT_INFO_CACHE_CORRUPTED]"},
		},
		{
			name: "redis unavailable",
			cnpj: cnpjLogOwnedDigits,
			prepare: func(t *testing.T, env *contactLogEnv, cnpj string) {
				env.mr.Close()
			},
			run: func(t *testing.T, env *contactLogEnv, cnpj string) {
				_, err := env.service.GetCNPJOwnerContactInfo(context.Background(), cnpj)
				require.NoError(t, err)
			},
			markers: []string{"[CONTACT_INFO_CACHE_ERROR]", "[CONTACT_INFO_CACHE_SET_ERROR]"},
		},
		{
			name:     "batch lookup error",
			cnpj:     cnpjLogOwnedDigits,
			notFound: true,
			run: func(t *testing.T, env *contactLogEnv, cnpj string) {
				results := env.service.GetMultipleCNPJOwnerContactInfo(context.Background(), []string{cnpj})
				assert.Empty(t, results)
			},
			markers: []string{"[CONTACT_INFO_ERROR] CNPJ:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := newContactLogEnv(t, tt.notFound)
			if tt.prepare != nil {
				tt.prepare(t, env, tt.cnpj)
			}
			sink := captureCNPJLog(t)

			tt.run(t, env, tt.cnpj)

			assertCNPJMaskedInLog(t, sink.String(), tt.markers)
		})
	}
}

func TestPropostaMEIService_GetByID_ContactInfoErrorLog_MaskCNPJ(t *testing.T) {
	env := newContactLogEnv(t, true)

	repo := NewMockPropostaRepo()
	id := uuid.New()
	repo.propostas[id] = &models.PropostaMEI{ID: id, MEIEmpresaID: cnpjLogOwnedDigits}
	service := services.NewPropostaMEIService(repo, NewMockOportunidadeRepo(), &MockCNAEValidation{}, env.service)

	sink := captureCNPJLog(t)

	proposta, err := service.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, proposta)
	assert.Empty(t, proposta.EmailPessoaFisica)

	assertCNPJMaskedInLog(t, sink.String(), []string{"[CONTACT_INFO_ERROR] Failed to fetch contact info for CNPJ"})
}
