package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// Dados sintéticos (DEV-148): CPFs de teste e um nome inventado.
const (
	syncTestCPFOK          = "11111111111"
	syncTestCPFNotFound    = "22222222222"
	syncTestCPFInvalid     = "33333333333"
	syncTestCPFUnavailable = "44444444444"
	syncTestNome           = "Fulana Sintetica de Teste"
)

func newCitizenSyncTestWorker(t *testing.T) (*CitizenSyncWorker, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/"+syncTestCPFOK):
			_ = json.NewEncoder(w).Encode(models.CitizenContactInfo{CPF: syncTestCPFOK, Nome: syncTestNome})
		case strings.HasSuffix(r.URL.Path, "/"+syncTestCPFNotFound):
			w.WriteHeader(http.StatusNotFound)
		case strings.HasSuffix(r.URL.Path, "/"+syncTestCPFInvalid):
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"Invalid CPF format"}`))
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	t.Cleanup(srv.Close)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.CitizenSnapshot{}))

	rmi := clients.NewRMIClient(srv.URL, 5*time.Second,
		clients.WithName(fmt.Sprintf("citizen-sync-%s-%d", t.Name(), time.Now().UnixNano())),
		clients.WithRetryBackoff(time.Millisecond, 2*time.Millisecond))

	var infoBuf, errorBuf bytes.Buffer
	w := &CitizenSyncWorker{
		rmiClient:           rmi,
		citizenSnapshotRepo: repository.NewCitizenSnapshotRepository(db),
		infoLog:             log.New(&infoBuf, "", 0),
		errorLog:            log.New(&errorBuf, "", 0),
	}
	return w, &infoBuf, &errorBuf
}

func TestCitizenSyncWorker_SyncCitizens_DoesNotLogCitizenName(t *testing.T) {
	w, infoBuf, errorBuf := newCitizenSyncTestWorker(t)

	synced, skipped, failed := w.syncCitizens(context.Background(), []string{syncTestCPFOK}, "token")

	assert.Equal(t, 1, synced)
	assert.Equal(t, 0, skipped)
	assert.Equal(t, 0, failed)
	assert.Contains(t, infoBuf.String(), "Synced citizen "+maskCPF(syncTestCPFOK))
	for _, out := range []string{infoBuf.String(), errorBuf.String()} {
		assert.NotContains(t, out, syncTestNome)
		assert.NotContains(t, out, "Fulana")
		assert.NotContains(t, out, syncTestCPFOK)
	}
}

func TestCitizenSyncWorker_SyncCitizens_ExpectedBusinessResultsAreNotErrors(t *testing.T) {
	w, infoBuf, errorBuf := newCitizenSyncTestWorker(t)

	synced, skipped, failed := w.syncCitizens(context.Background(),
		[]string{syncTestCPFNotFound, syncTestCPFInvalid}, "token")

	assert.Equal(t, 0, synced)
	assert.Equal(t, 2, skipped)
	assert.Equal(t, 0, failed)
	assert.Empty(t, errorBuf.String(), "expected business results must not reach the ERROR stream")

	info := infoBuf.String()
	assert.Contains(t, info, "WARN")
	assert.Contains(t, info, maskCPF(syncTestCPFNotFound))
	assert.Contains(t, info, "citizen not found")
	assert.Contains(t, info, maskCPF(syncTestCPFInvalid))
	assert.Contains(t, info, "status 400")
	assert.NotContains(t, info, syncTestCPFNotFound)
	assert.NotContains(t, info, syncTestCPFInvalid)
}

func TestCitizenSyncWorker_SyncCitizens_RMIUnavailableIsError(t *testing.T) {
	w, infoBuf, errorBuf := newCitizenSyncTestWorker(t)

	synced, skipped, failed := w.syncCitizens(context.Background(), []string{syncTestCPFUnavailable}, "token")

	assert.Equal(t, 0, synced)
	assert.Equal(t, 0, skipped)
	assert.Equal(t, 1, failed)
	assert.Contains(t, errorBuf.String(), "Failed to sync citizen "+maskCPF(syncTestCPFUnavailable))
	assert.Contains(t, errorBuf.String(), "status 503")
	assert.NotContains(t, errorBuf.String(), syncTestCPFUnavailable)
	assert.NotContains(t, infoBuf.String(), maskCPF(syncTestCPFUnavailable))
}

func TestIsExpectedCitizenSyncError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"cidadão inexistente", clients.ErrCitizenNotFound, true},
		{"cidadão inexistente embrulhado", fmt.Errorf("sync: %w", clients.ErrCitizenNotFound), true},
		{"CPF inválido (400)", &clients.RMIStatusError{StatusCode: http.StatusBadRequest}, true},
		{"RMI fora (500)", &clients.RMIStatusError{StatusCode: http.StatusInternalServerError}, false},
		{"RMI fora (503)", &clients.RMIStatusError{StatusCode: http.StatusServiceUnavailable}, false},
		{"acesso negado (403)", &clients.RMIStatusError{StatusCode: http.StatusForbidden}, false},
		{"rede", errors.New("failed to execute citizen request: dial tcp: connection refused"), false},
		{"texto parecido sem tipo", errors.New("citizen not found"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isExpectedCitizenSyncError(tt.err))
		})
	}
}

func TestCitizenSyncWorker_DefaultLoggers_WhenNotSet(t *testing.T) {
	w := &CitizenSyncWorker{}

	assert.Same(t, defaultCitizenSyncInfoLog, w.infoLogger())
	assert.Same(t, defaultCitizenSyncErrorLog, w.errorLogger())
}

func TestCitizenSyncWorker_SyncCitizenForced_StaleFallbackSeverity(t *testing.T) {
	tests := []struct {
		name      string
		cpf       string
		wantInfo  bool
		wantError bool
	}{
		{"cidadão inexistente vira aviso", syncTestCPFNotFound, true, false},
		{"RMI fora continua erro", syncTestCPFUnavailable, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, infoBuf, errorBuf := newCitizenSyncTestWorker(t)

			keycloak := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(rw).Encode(auth.TokenResponse{AccessToken: "token-sintetico", ExpiresIn: 300})
			}))
			t.Cleanup(keycloak.Close)
			w.tokenManager = auth.NewServiceAccountTokenManager(keycloak.URL, "realm", "client", "secret")

			require.NoError(t, w.citizenSnapshotRepo.Upsert(context.Background(),
				&models.CitizenSnapshot{CPF: tt.cpf, Nome: syncTestNome}))

			snapshot, err := w.SyncCitizenForced(context.Background(), tt.cpf)
			require.NoError(t, err)
			require.NotNil(t, snapshot, "the stale snapshot must still be returned")

			marker := "stale snapshot"
			assert.Equal(t, tt.wantInfo, strings.Contains(infoBuf.String(), marker), "stdout: %q", infoBuf.String())
			assert.Equal(t, tt.wantError, strings.Contains(errorBuf.String(), marker), "stderr: %q", errorBuf.String())
			for _, out := range []string{infoBuf.String(), errorBuf.String()} {
				assert.NotContains(t, out, syncTestNome)
				assert.NotContains(t, out, tt.cpf)
			}
		})
	}
}
