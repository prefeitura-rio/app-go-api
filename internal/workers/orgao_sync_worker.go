package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/prefeitura-rio/app-go-api/internal/clients"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"gorm.io/gorm"
)

const orgaoSyncLockKey = "orgao_sync:lock"

// OrgaoSyncWorker handles periodic synchronization of orgao data from RMI API
type OrgaoSyncWorker struct {
	db                   *gorm.DB
	rmiClient            *clients.RMIClient
	redisClient          *redis.Client
	orgaoSnapshotRepo    *repository.OrgaoSnapshotRepository
	cursoRepo            *repository.CursoRepository
	empregoRepo          *repository.EmpregoRepository
	oportunidadeMEIRepo  *repository.OportunidadeMEIRepository
	syncInterval         time.Duration
	staleThreshold       time.Duration
	failedRetryThreshold time.Duration
	batchSize            int
	maxRetries           int
}

// NewOrgaoSyncWorker creates a new orgao sync worker instance.
// redisClient is used for distributed locking so only one pod runs the sync at a time.
func NewOrgaoSyncWorker(
	db *gorm.DB,
	rmiClient *clients.RMIClient,
	redisClient *redis.Client,
	orgaoSnapshotRepo *repository.OrgaoSnapshotRepository,
	cursoRepo *repository.CursoRepository,
	empregoRepo *repository.EmpregoRepository,
	oportunidadeMEIRepo *repository.OportunidadeMEIRepository,
	cfg *config.OrgaoSyncSettings,
) *OrgaoSyncWorker {
	return &OrgaoSyncWorker{
		db:                   db,
		rmiClient:            rmiClient,
		redisClient:          redisClient,
		orgaoSnapshotRepo:    orgaoSnapshotRepo,
		cursoRepo:            cursoRepo,
		empregoRepo:          empregoRepo,
		oportunidadeMEIRepo:  oportunidadeMEIRepo,
		syncInterval:         cfg.SyncInterval,
		staleThreshold:       cfg.StaleThreshold,
		failedRetryThreshold: cfg.FailedRetryThreshold,
		batchSize:            cfg.BatchSize,
		maxRetries:           cfg.MaxRetries,
	}
}

// tryAcquireLock attempts to acquire a distributed Redis lock.
// Returns (token, true, nil) if acquired, ("", false, nil) if held by another instance,
// or ("", false, err) on Redis error.
// The lock TTL equals the sync interval so it expires automatically if the pod crashes.
func (w *OrgaoSyncWorker) tryAcquireLock(ctx context.Context) (string, bool, error) {
	if w.redisClient == nil {
		return "", true, nil // no Redis configured, run without lock
	}
	token := uuid.New().String()
	acquired, err := w.redisClient.SetNX(ctx, orgaoSyncLockKey, token, w.syncInterval).Result()
	if err != nil {
		return "", false, fmt.Errorf("redis lock error: %w", err)
	}
	if !acquired {
		return "", false, nil
	}
	return token, true, nil
}

// releaseLock releases the lock only if we still own it (atomic compare-and-delete via Lua).
// This prevents a pod from accidentally deleting another pod's lock after TTL expiry.
func (w *OrgaoSyncWorker) releaseLock(ctx context.Context, token string) {
	if w.redisClient == nil || token == "" {
		return
	}
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)
	if err := script.Run(ctx, w.redisClient, []string{orgaoSyncLockKey}, token).Err(); err != nil && err != redis.Nil {
		log.Printf("[OrgaoSyncWorker] Failed to release lock: %v", err)
	}
}

// Start begins the worker loop - runs until context is cancelled
func (w *OrgaoSyncWorker) Start(ctx context.Context) error {
	log.Println("[OrgaoSyncWorker] Starting orgao sync worker...")

	// Run initial sync immediately
	if err := w.runSync(ctx); err != nil {
		log.Printf("[OrgaoSyncWorker] Initial sync failed: %v", err)
		// Continue anyway - don't fail startup
	}

	// Setup periodic ticker
	ticker := time.NewTicker(w.syncInterval)
	defer ticker.Stop()

	log.Printf("[OrgaoSyncWorker] Worker started successfully. Sync interval: %v", w.syncInterval)

	for {
		select {
		case <-ctx.Done():
			log.Println("[OrgaoSyncWorker] Context cancelled, shutting down worker...")
			return ctx.Err()
		case <-ticker.C:
			if err := w.runSync(ctx); err != nil {
				log.Printf("[OrgaoSyncWorker] Sync cycle failed: %v", err)
				// Continue running despite errors
			}
		}
	}
}

// runSync executes a full sync cycle, guarded by a distributed Redis lock.
// If another pod already holds the lock, the cycle is skipped.
func (w *OrgaoSyncWorker) runSync(ctx context.Context) error {
	token, acquired, err := w.tryAcquireLock(ctx)
	if err != nil {
		log.Printf("[OrgaoSyncWorker] Could not acquire distributed lock: %v", err)
		return fmt.Errorf("could not acquire distributed lock: %w", err)
	}
	if !acquired {
		log.Println("[OrgaoSyncWorker] Another instance is running sync, skipping cycle")
		return nil
	}
	defer w.releaseLock(ctx, token)

	startTime := time.Now()
	log.Println("[OrgaoSyncWorker] Starting sync cycle (lock acquired)...")

	// Discover all unique orgao IDs from database
	orgaoIDs, err := w.discoverOrgaoIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover orgao IDs: %w", err)
	}

	if len(orgaoIDs) == 0 {
		log.Println("[OrgaoSyncWorker] No orgao IDs found to sync")
		return nil
	}

	log.Printf("[OrgaoSyncWorker] Discovered %d unique orgao IDs", len(orgaoIDs))

	// Filter IDs that need syncing (missing or stale)
	idsToSync, err := w.filterIDsToSync(ctx, orgaoIDs)
	if err != nil {
		return fmt.Errorf("failed to filter IDs to sync: %w", err)
	}

	if len(idsToSync) == 0 {
		log.Println("[OrgaoSyncWorker] All orgao snapshots are up to date")
		return nil
	}

	log.Printf("[OrgaoSyncWorker] Syncing %d orgao snapshots...", len(idsToSync))

	// Sync in batches
	synced, failed := w.syncOrgaos(ctx, idsToSync)

	duration := time.Since(startTime)
	log.Printf("[OrgaoSyncWorker] Sync cycle completed in %v: %d synced, %d failed",
		duration, synced, failed)

	// Log stats
	if err := w.logSyncStats(ctx); err != nil {
		log.Printf("[OrgaoSyncWorker] Failed to log sync stats: %v", err)
	}

	return nil
}

// discoverOrgaoIDs queries all tables to find unique orgao_id values
func (w *OrgaoSyncWorker) discoverOrgaoIDs(ctx context.Context) ([]string, error) {
	orgaoIDSet := make(map[string]struct{})

	// Query cursos
	var cursoOrgaoIDs []string
	if err := w.db.WithContext(ctx).
		Model(&models.Curso{}).
		Where("orgao_id IS NOT NULL AND orgao_id != ''").
		Distinct("orgao_id").
		Pluck("orgao_id", &cursoOrgaoIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to query curso orgao_ids: %w", err)
	}
	for _, id := range cursoOrgaoIDs {
		orgaoIDSet[id] = struct{}{}
	}

	// Query empregos
	var empregoOrgaoIDs []string
	if err := w.db.WithContext(ctx).
		Model(&models.Emprego{}).
		Where("orgao_id IS NOT NULL AND orgao_id != ''").
		Distinct("orgao_id").
		Pluck("orgao_id", &empregoOrgaoIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to query emprego orgao_ids: %w", err)
	}
	for _, id := range empregoOrgaoIDs {
		orgaoIDSet[id] = struct{}{}
	}

	// Query oportunidades_mei
	var meiOrgaoIDs []string
	if err := w.db.WithContext(ctx).
		Model(&models.OportunidadeMEI{}).
		Where("orgao_id IS NOT NULL AND orgao_id != ''").
		Distinct("orgao_id").
		Pluck("orgao_id", &meiOrgaoIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to query oportunidade_mei orgao_ids: %w", err)
	}
	for _, id := range meiOrgaoIDs {
		orgaoIDSet[id] = struct{}{}
	}

	// Convert set to slice
	orgaoIDs := make([]string, 0, len(orgaoIDSet))
	for id := range orgaoIDSet {
		orgaoIDs = append(orgaoIDs, id)
	}

	return orgaoIDs, nil
}

// filterIDsToSync returns IDs that are missing or stale.
// Uses a single bulk query instead of N individual queries to avoid hammering the DB.
// Failed syncs are only retried after failedRetryThreshold to prevent thundering herd against the RMI API.
func (w *OrgaoSyncWorker) filterIDsToSync(ctx context.Context, orgaoIDs []string) ([]string, error) {
	staleTime := time.Now().Add(-w.staleThreshold)
	failedRetryTime := time.Now().Add(-w.failedRetryThreshold)

	snapshotMap, err := w.orgaoSnapshotRepo.GetByOrgaoIDs(ctx, orgaoIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	idsToSync := []string{}
	for _, orgaoID := range orgaoIDs {
		snapshot, exists := snapshotMap[orgaoID]
		if !exists {
			idsToSync = append(idsToSync, orgaoID)
			continue
		}

		switch snapshot.SyncStatus {
		case models.SyncStatusFailed:
			// Back off: only retry failed syncs after failedRetryThreshold
			if snapshot.LastSyncedAt.Before(failedRetryTime) {
				idsToSync = append(idsToSync, orgaoID)
			}
		case models.SyncStatusPending:
			idsToSync = append(idsToSync, orgaoID)
		default: // SyncStatusSynced
			if snapshot.LastSyncedAt.Before(staleTime) {
				idsToSync = append(idsToSync, orgaoID)
			}
		}
	}

	return idsToSync, nil
}

// syncOrgaos syncs orgao data in batches
func (w *OrgaoSyncWorker) syncOrgaos(ctx context.Context, orgaoIDs []string) (synced int, failed int) {
	for i := 0; i < len(orgaoIDs); i += w.batchSize {
		end := i + w.batchSize
		if end > len(orgaoIDs) {
			end = len(orgaoIDs)
		}

		batch := orgaoIDs[i:end]
		log.Printf("[OrgaoSyncWorker] Processing batch %d-%d of %d", i+1, end, len(orgaoIDs))

		for _, orgaoID := range batch {
			if err := w.syncSingleOrgao(ctx, orgaoID); err != nil {
				log.Printf("[OrgaoSyncWorker] Failed to sync orgao %s: %v", orgaoID, err)
				failed++
			} else {
				synced++
			}

			// Check context cancellation between iterations
			select {
			case <-ctx.Done():
				log.Println("[OrgaoSyncWorker] Context cancelled during sync")
				return synced, failed
			default:
				// Continue
			}
		}
	}

	return synced, failed
}

// syncSingleOrgao fetches and upserts a single orgao snapshot
func (w *OrgaoSyncWorker) syncSingleOrgao(ctx context.Context, orgaoID string) error {
	// Fetch from RMI API
	orgao, err := w.rmiClient.GetOrgao(ctx, orgaoID)
	if err != nil {
		// Create/update snapshot with error status
		snapshot := &models.OrgaoSnapshot{
			OrgaoID:      orgaoID,
			Name:         "",
			Sigla:        nil,
			Metadata:     models.OrgaoMetadata{},
			LastSyncedAt: time.Now(),
			SyncStatus:   models.SyncStatusFailed,
			SyncError:    stringPtr(err.Error()),
		}
		_ = w.orgaoSnapshotRepo.Upsert(ctx, snapshot)
		return err
	}

	// Prepare metadata (store all additional fields)
	metadata := models.OrgaoMetadata{
		"id":              orgao.ID,
		"cd_ua":           orgao.CdUA,
		"cd_ua_pai":       orgao.CdUAPai,
		"nivel":           orgao.Nivel,
		"ordem_ua_basica": orgao.OrdemUABasica,
		"ordem_absoluta":  orgao.OrdemAbsoluta,
		"ordem_relativa":  orgao.OrdemRelativa,
	}

	// Prepare sigla pointer
	var sigla *string
	if orgao.SiglaUA != "" {
		sigla = &orgao.SiglaUA
	}

	// Upsert snapshot
	snapshot := &models.OrgaoSnapshot{
		OrgaoID:      orgaoID,
		CdUA:         orgao.CdUA,
		Name:         orgao.NomeUA,
		Sigla:        sigla,
		Metadata:     metadata,
		LastSyncedAt: time.Now(),
		SyncStatus:   models.SyncStatusSynced,
		SyncError:    nil,
	}

	if err := w.orgaoSnapshotRepo.Upsert(ctx, snapshot); err != nil {
		return fmt.Errorf("failed to upsert snapshot: %w", err)
	}

	displayName := orgao.NomeUA
	if sigla != nil {
		displayName = fmt.Sprintf("%s (%s)", orgao.NomeUA, *sigla)
	}
	log.Printf("[OrgaoSyncWorker] Synced orgao %s: %s", orgaoID, displayName)
	return nil
}

// logSyncStats logs current sync statistics
func (w *OrgaoSyncWorker) logSyncStats(ctx context.Context) error {
	counts, err := w.orgaoSnapshotRepo.CountByStatus(ctx)
	if err != nil {
		return err
	}

	statsJSON, _ := json.MarshalIndent(counts, "", "  ")
	log.Printf("[OrgaoSyncWorker] Current snapshot statistics:\n%s", string(statsJSON))

	return nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
