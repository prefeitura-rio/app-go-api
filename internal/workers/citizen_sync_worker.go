package workers

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// CitizenSyncWorker handles periodic synchronization of citizen data from RMI API
type CitizenSyncWorker struct {
	rmiClient           *clients.RMIClient
	citizenSnapshotRepo *repository.CitizenSnapshotRepository
	tokenManager        *auth.ServiceAccountTokenManager
	syncInterval        time.Duration
	staleThreshold      time.Duration
	batchSize           int
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
	log.Println("[CitizenSyncWorker] Starting citizen sync worker...")

	// Run initial sync immediately
	if err := w.runSync(ctx); err != nil {
		log.Printf("[CitizenSyncWorker] Initial sync failed: %v", err)
	}

	// Setup periodic ticker
	ticker := time.NewTicker(w.syncInterval)
	defer ticker.Stop()

	log.Printf("[CitizenSyncWorker] Worker started successfully. Sync interval: %v", w.syncInterval)

	for {
		select {
		case <-ctx.Done():
			log.Println("[CitizenSyncWorker] Context cancelled, shutting down worker...")
			return ctx.Err()
		case <-ticker.C:
			if err := w.runSync(ctx); err != nil {
				log.Printf("[CitizenSyncWorker] Sync cycle failed: %v", err)
			}
		}
	}
}

// runSync executes a full sync cycle
func (w *CitizenSyncWorker) runSync(ctx context.Context) error {
	startTime := time.Now()
	log.Println("[CitizenSyncWorker] Starting sync cycle...")

	// Find CPFs that need syncing (have enrollments but missing/stale snapshot)
	cpfsToSync, err := w.citizenSnapshotRepo.GetCPFsWithEnrollments(ctx, w.staleThreshold, w.batchSize)
	if err != nil {
		log.Printf("[CitizenSyncWorker] Failed to get CPFs to sync: %v", err)
		return err
	}

	if len(cpfsToSync) == 0 {
		log.Println("[CitizenSyncWorker] No citizen snapshots need syncing")
		return nil
	}

	log.Printf("[CitizenSyncWorker] Found %d CPFs to sync", len(cpfsToSync))

	// Get service account token
	token, err := w.tokenManager.GetToken(ctx)
	if err != nil {
		log.Printf("[CitizenSyncWorker] Failed to get service token: %v", err)
		return err
	}

	// Sync citizens
	synced, failed := w.syncCitizens(ctx, cpfsToSync, token)

	duration := time.Since(startTime)
	log.Printf("[CitizenSyncWorker] Sync cycle completed in %v: %d synced, %d failed",
		duration, synced, failed)

	// Log stats
	w.logSyncStats(ctx)

	return nil
}

// syncCitizens syncs citizen data from RMI API
func (w *CitizenSyncWorker) syncCitizens(ctx context.Context, cpfs []string, token string) (synced int, failed int) {
	for _, cpf := range cpfs {
		if err := w.syncSingleCitizen(ctx, cpf, token); err != nil {
			log.Printf("[CitizenSyncWorker] Failed to sync citizen %s: %v", maskCPF(cpf), err)
			failed++
		} else {
			synced++
		}

		// Check context cancellation between iterations
		select {
		case <-ctx.Done():
			log.Println("[CitizenSyncWorker] Context cancelled during sync")
			return synced, failed
		default:
		}
	}

	return synced, failed
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

	log.Printf("[CitizenSyncWorker] Synced citizen %s (%s)", maskCPF(cpf), citizenInfo.Nome)
	return nil
}

// citizenInfoToSnapshot converts RMI CitizenContactInfo to CitizenSnapshot
func (w *CitizenSyncWorker) citizenInfoToSnapshot(cpf string, info *models.CitizenContactInfo) *models.CitizenSnapshot {
	snapshot := &models.CitizenSnapshot{
		CPF:           cpf,
		Nome:          info.Nome,
		Email:         info.GetEmail(),
		Celular:       info.GetCelular(),
		Raca:          info.GetRaca(),
		Genero:        info.GetGenero(),
		RendaFamiliar: info.GetRendaFamiliar(),
		Escolaridade:  info.GetEscolaridade(),
		Deficiencia:   info.GetDeficiencia(),
	}

	// Parse data_nascimento
	if dataNascimento := info.GetDataNascimento(); dataNascimento != "" {
		if t, err := parseDate(dataNascimento); err == nil {
			snapshot.DataNascimento = &t
		}
	}

	// Convert endereco
	if endereco := info.GetEndereco(); endereco != nil {
		snapshot.Endereco = &models.CitizenEndereco{
			Logradouro:  endereco.Principal.Logradouro,
			Numero:      endereco.Principal.Numero,
			Complemento: endereco.Principal.Complemento,
			Bairro:      endereco.Principal.Bairro,
			Cidade:      endereco.Principal.Cidade,
			UF:          endereco.Principal.UF,
			CEP:         endereco.Principal.CEP,
		}
	}

	return snapshot
}

// SyncCitizenOnDemand syncs a single citizen on-demand (called during enrollment creation)
func (w *CitizenSyncWorker) SyncCitizenOnDemand(ctx context.Context, cpf string) (*models.CitizenSnapshot, error) {
	// Check if we already have a fresh snapshot
	existing, err := w.citizenSnapshotRepo.GetByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}

	staleTime := time.Now().Add(-w.staleThreshold)
	if existing != nil && existing.LastSyncedAt.After(staleTime) {
		return existing, nil
	}

	// Fetch fresh data from RMI
	token, err := w.tokenManager.GetToken(ctx)
	if err != nil {
		// If token fails but we have existing data, return it
		if existing != nil {
			log.Printf("[CitizenSyncWorker] Token failed, returning stale snapshot for %s: %v", maskCPF(cpf), err)
			return existing, nil
		}
		return nil, err
	}

	citizenInfo, err := w.rmiClient.GetCitizenByCPF(ctx, token, cpf)
	if err != nil {
		// If RMI fails but we have existing data, return it
		if existing != nil {
			log.Printf("[CitizenSyncWorker] RMI failed, returning stale snapshot for %s: %v", maskCPF(cpf), err)
			return existing, nil
		}
		return nil, err
	}

	// Convert to snapshot
	snapshot := w.citizenInfoToSnapshot(cpf, citizenInfo)

	// Upsert snapshot
	if err := w.citizenSnapshotRepo.Upsert(ctx, snapshot); err != nil {
		return nil, err
	}

	log.Printf("[CitizenSyncWorker] On-demand sync completed for citizen %s", maskCPF(cpf))
	return snapshot, nil
}

// logSyncStats logs current sync statistics
func (w *CitizenSyncWorker) logSyncStats(ctx context.Context) {
	count, err := w.citizenSnapshotRepo.Count(ctx)
	if err != nil {
		log.Printf("[CitizenSyncWorker] Failed to get stats: %v", err)
		return
	}
	log.Printf("[CitizenSyncWorker] Total citizen snapshots: %d", count)
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
