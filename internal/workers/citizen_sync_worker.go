package workers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// The worker writes informational lines and expected business results to
// stdout and real failures to stderr: plain-text container logs get severity
// INFO from stdout and ERROR from stderr in GKE, so ERROR stays reserved for
// infrastructure problems (token, database, RMI unavailable).
var (
	defaultCitizenSyncInfoLog  = log.New(os.Stdout, "", log.LstdFlags)
	defaultCitizenSyncErrorLog = log.New(os.Stderr, "", log.LstdFlags)
)

// CitizenSyncWorker handles periodic synchronization of citizen data from RMI API
type CitizenSyncWorker struct {
	rmiClient           *clients.RMIClient
	citizenSnapshotRepo *repository.CitizenSnapshotRepository
	tokenManager        *auth.ServiceAccountTokenManager
	syncInterval        time.Duration
	staleThreshold      time.Duration
	batchSize           int
	infoLog             *log.Logger // nil: defaultCitizenSyncInfoLog
	errorLog            *log.Logger // nil: defaultCitizenSyncErrorLog
}

func (w *CitizenSyncWorker) infoLogger() *log.Logger {
	if w.infoLog != nil {
		return w.infoLog
	}
	return defaultCitizenSyncInfoLog
}

func (w *CitizenSyncWorker) errorLogger() *log.Logger {
	if w.errorLog != nil {
		return w.errorLog
	}
	return defaultCitizenSyncErrorLog
}

// isExpectedCitizenSyncError reports whether err is a business result the RMI
// gives for the CPF itself (citizen not found, CPF rejected as invalid) rather
// than a failure of the sync: these are logged as warnings, not errors.
func isExpectedCitizenSyncError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, clients.ErrCitizenNotFound) {
		return true
	}
	var statusErr *clients.RMIStatusError
	return errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusBadRequest
}

// NewCitizenSyncWorker creates a new citizen sync worker instance
func NewCitizenSyncWorker(
	rmiClient *clients.RMIClient,
	citizenSnapshotRepo *repository.CitizenSnapshotRepository,
	tokenManager *auth.ServiceAccountTokenManager,
	cfg *config.CitizenSyncSettings,
) *CitizenSyncWorker {
	return &CitizenSyncWorker{
		rmiClient:           rmiClient,
		citizenSnapshotRepo: citizenSnapshotRepo,
		tokenManager:        tokenManager,
		syncInterval:        cfg.SyncInterval,
		staleThreshold:      cfg.StaleThreshold,
		batchSize:           cfg.BatchSize,
	}
}

// Start begins the worker loop - runs until context is cancelled
func (w *CitizenSyncWorker) Start(ctx context.Context) error {
	w.infoLogger().Println("[CitizenSyncWorker] Starting citizen sync worker...")

	// Run initial sync immediately
	if err := w.runSync(ctx); err != nil {
		w.errorLogger().Printf("[CitizenSyncWorker] Initial sync failed: %v", err)
	}

	// Setup periodic ticker
	ticker := time.NewTicker(w.syncInterval)
	defer ticker.Stop()

	w.infoLogger().Printf("[CitizenSyncWorker] Worker started successfully. Sync interval: %v", w.syncInterval)

	for {
		select {
		case <-ctx.Done():
			w.infoLogger().Println("[CitizenSyncWorker] Context cancelled, shutting down worker...")
			return ctx.Err()
		case <-ticker.C:
			if err := w.runSync(ctx); err != nil {
				w.errorLogger().Printf("[CitizenSyncWorker] Sync cycle failed: %v", err)
			}
		}
	}
}

// runSync executes a full sync cycle
func (w *CitizenSyncWorker) runSync(ctx context.Context) error {
	startTime := time.Now()
	w.infoLogger().Println("[CitizenSyncWorker] Starting sync cycle...")

	// Find CPFs that need syncing (have enrollments but missing/stale snapshot)
	cpfsToSync, err := w.citizenSnapshotRepo.GetCPFsWithEnrollments(ctx, w.staleThreshold, w.batchSize)
	if err != nil {
		w.errorLogger().Printf("[CitizenSyncWorker] Failed to get CPFs to sync: %v", err)
		return err
	}

	if len(cpfsToSync) == 0 {
		w.infoLogger().Println("[CitizenSyncWorker] No citizen snapshots need syncing")
		return nil
	}

	w.infoLogger().Printf("[CitizenSyncWorker] Found %d CPFs to sync", len(cpfsToSync))

	// Get service account token
	token, err := w.tokenManager.GetToken(ctx)
	if err != nil {
		w.errorLogger().Printf("[CitizenSyncWorker] Failed to get service token: %v", err)
		return err
	}

	// Sync citizens
	synced, skipped, failed := w.syncCitizens(ctx, cpfsToSync, token)

	duration := time.Since(startTime)
	w.infoLogger().Printf("[CitizenSyncWorker] Sync cycle completed in %v: %d synced, %d skipped (not found or invalid CPF), %d failed",
		duration, synced, skipped, failed)

	// Log stats
	w.logSyncStats(ctx)

	return nil
}

// syncCitizens syncs citizen data from RMI API. skipped counts the CPFs the RMI
// answered with an expected business result (see isExpectedCitizenSyncError);
// failed counts real failures.
func (w *CitizenSyncWorker) syncCitizens(ctx context.Context, cpfs []string, token string) (synced, skipped, failed int) {
	for _, cpf := range cpfs {
		err := w.syncSingleCitizen(ctx, cpf, token)
		switch {
		case err == nil:
			synced++
		case isExpectedCitizenSyncError(err):
			w.infoLogger().Printf("[CitizenSyncWorker] WARN: citizen %s not synced: %v", maskCPF(cpf), err)
			skipped++
		default:
			w.errorLogger().Printf("[CitizenSyncWorker] Failed to sync citizen %s: %v", maskCPF(cpf), err)
			failed++
		}

		// Check context cancellation between iterations
		select {
		case <-ctx.Done():
			w.infoLogger().Println("[CitizenSyncWorker] Context cancelled during sync")
			return synced, skipped, failed
		default:
		}
	}

	return synced, skipped, failed
}

// syncSingleCitizen fetches and upserts a single citizen snapshot
func (w *CitizenSyncWorker) syncSingleCitizen(ctx context.Context, cpf string, token string) error {
	// Fetch from RMI API
	citizenInfo, err := w.rmiClient.GetCitizenByCPF(ctx, token, cpf)
	if err != nil {
		return err
	}

	// Convert to snapshot
	snapshot := w.citizenInfoToSnapshot(cpf, citizenInfo)

	// Upsert snapshot
	if err := w.citizenSnapshotRepo.Upsert(ctx, snapshot); err != nil {
		return err
	}

	w.infoLogger().Printf("[CitizenSyncWorker] Synced citizen %s", maskCPF(cpf))
	return nil
}

// citizenInfoToSnapshot converts RMI CitizenContactInfo to CitizenSnapshot
func (w *CitizenSyncWorker) citizenInfoToSnapshot(cpf string, info *models.CitizenContactInfo) *models.CitizenSnapshot {
	snapshot := &models.CitizenSnapshot{
		CPF:           cpf,
		Nome:          info.Nome,
		NomeSocial:    info.NomeSocial,
		Email:         info.GetEmail(),
		Celular:       info.GetCelular(),
		Raca:          info.GetRaca(),
		Genero:        info.GetGenero(),
		RendaFamiliar: info.GetRendaFamiliar(),
		Escolaridade:  info.GetEscolaridade(),
		Deficiencia:   info.GetDeficiencia(),
	}

	// Parse data_nascimento from nascimento.data
	if dataNascimento := info.GetDataNascimento(); dataNascimento != "" {
		if t, err := parseDate(dataNascimento); err == nil {
			snapshot.DataNascimento = &t
		}
	}

	// Convert endereco
	if endereco := info.GetEndereco(); endereco != nil {
		snapshot.Endereco = &models.CitizenEndereco{
			Logradouro:     endereco.Principal.Logradouro,
			TipoLogradouro: endereco.Principal.TipoLogradouro,
			Numero:         endereco.Principal.Numero,
			Complemento:    endereco.Principal.Complemento,
			Bairro:         endereco.Principal.Bairro,
			Municipio:      endereco.Principal.Municipio,
			Estado:         endereco.Principal.Estado,
			CEP:            endereco.Principal.CEP,
		}
	}

	return snapshot
}

// StaleThreshold exposes the age after which a snapshot is considered stale, so
// callers holding a snapshot can decide whether a refresh is needed without
// paying for an extra lookup inside SyncCitizenOnDemand.
func (w *CitizenSyncWorker) StaleThreshold() time.Duration {
	return w.staleThreshold
}

// SyncCitizenOnDemand syncs a single citizen on-demand, respecting staleThreshold.
func (w *CitizenSyncWorker) SyncCitizenOnDemand(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
	return w.syncCitizen(ctx, cpf, false)
}

// SyncCitizenForced syncs a single citizen on-demand, always querying RMI directly (bypassing staleThreshold).
// Used on write/application operations (such as candidatura and inscricao creation) where the citizen has
// just confirmed their data, ensuring the freshest snapshot is always persisted.
func (w *CitizenSyncWorker) SyncCitizenForced(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
	return w.syncCitizen(ctx, cpf, true)
}

func (w *CitizenSyncWorker) syncCitizen(ctx context.Context, cpf string, force bool) (*models.CitizenSnapshot, error) {
	// Check if we already have a fresh snapshot
	existing, err := w.citizenSnapshotRepo.GetByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}

	if !force && existing != nil {
		staleTime := time.Now().Add(-w.staleThreshold)
		if existing.LastSyncedAt.After(staleTime) {
			return existing, nil
		}
	}

	// Fetch fresh data from RMI
	token, err := w.tokenManager.GetToken(ctx)
	if err != nil {
		// If token fails but we have existing data, return it
		if existing != nil {
			w.errorLogger().Printf("[CitizenSyncWorker] Token failed, returning stale snapshot for %s: %v", maskCPF(cpf), err)
			return existing, nil
		}
		return nil, err
	}

	citizenInfo, err := w.rmiClient.GetCitizenByCPF(ctx, token, cpf)
	if err != nil {
		// If RMI fails but we have existing data, return it
		if existing != nil {
			if isExpectedCitizenSyncError(err) {
				w.infoLogger().Printf("[CitizenSyncWorker] WARN: RMI did not return citizen %s, returning stale snapshot: %v", maskCPF(cpf), err)
			} else {
				w.errorLogger().Printf("[CitizenSyncWorker] RMI failed, returning stale snapshot for %s: %v", maskCPF(cpf), err)
			}
			return existing, nil
		}
		return nil, err
	}

	// Convert to snapshot
	snapshot := w.citizenInfoToSnapshot(cpf, citizenInfo)

	if existing != nil {
		oldHash := existing.ComputeDataHash()
		newHash := snapshot.ComputeDataHash()
		if oldHash != newHash && len(oldHash) >= 8 && len(newHash) >= 8 {
			w.infoLogger().Printf("[CitizenSyncWorker] Citizen %s data updated in RMI (hash changed: %s -> %s)", maskCPF(cpf), oldHash[:8], newHash[:8])
		}
	}

	// Upsert snapshot
	if err := w.citizenSnapshotRepo.Upsert(ctx, snapshot); err != nil {
		return nil, err
	}

	w.infoLogger().Printf("[CitizenSyncWorker] On-demand sync completed for citizen %s", maskCPF(cpf))
	return snapshot, nil
}

// logSyncStats logs current sync statistics
func (w *CitizenSyncWorker) logSyncStats(ctx context.Context) {
	count, err := w.citizenSnapshotRepo.Count(ctx)
	if err != nil {
		w.errorLogger().Printf("[CitizenSyncWorker] Failed to get stats: %v", err)
		return
	}
	w.infoLogger().Printf("[CitizenSyncWorker] Total citizen snapshots: %d", count)
}

// parseDate tries to parse date in common formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, nil
}

// maskCPF masks CPF for logging (shows only first 3 and last 2 digits)
func maskCPF(cpf string) string {
	cpf = strings.ReplaceAll(cpf, ".", "")
	cpf = strings.ReplaceAll(cpf, "-", "")
	if len(cpf) < 5 {
		return "***"
	}
	return cpf[:3] + "******" + cpf[len(cpf)-2:]
}
