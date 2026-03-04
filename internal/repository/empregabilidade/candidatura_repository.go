package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type BulkUpdateResult struct {
	Updated    int
	FailedCPFs []string
}

type CandidaturaRepository struct {
	db *gorm.DB
}

func NewCandidaturaRepository(db *gorm.DB) *CandidaturaRepository {
	return &CandidaturaRepository{db: db}
}

func (r *CandidaturaRepository) Create(ctx context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar candidatura: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *CandidaturaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Candidatura, error) {
	var entity empregabilidade.Candidatura
	result := r.db.WithContext(ctx).
		Preload("Vaga").
		Preload("Vaga.Contratante").
		Preload("Vaga.Etapas").
		Preload("EtapaAtual").
		First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar candidatura: %w", result.Error)
	}
	return &entity, nil
}

func (r *CandidaturaRepository) Update(ctx context.Context, entity *empregabilidade.Candidatura) error {
	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar candidatura: %w", result.Error)
	}
	return nil
}

func (r *CandidaturaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Candidatura{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir candidatura: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("candidatura não encontrada")
	}
	return nil
}

func (r *CandidaturaRepository) List(ctx context.Context, filter empregabilidade.CandidaturaFilter, limit, offset int) ([]*empregabilidade.Candidatura, int, error) {
	var entities []*empregabilidade.Candidatura
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.CPF != "" {
			db = db.Where("cpf = ?", filter.CPF)
		}
		if filter.VagaID != nil {
			db = db.Where("id_vaga = ?", *filter.VagaID)
		}
		if filter.Status != "" {
			db = db.Where("status = ?", filter.Status)
		}
		if filter.EtapaID != nil {
			db = db.Where("id_etapa_atual = ?", *filter.EtapaID)
		}
		if filter.Search != "" {
			searchTerm := fmt.Sprintf("%%%s%%", filter.Search)
			db = db.Where("cpf ILIKE ? OR nome ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm)
		}
		return db
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Candidatura{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar candidaturas: %w", err)
	}

	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Candidatura{}))
	result := findDB.
		Preload("Vaga").
		Preload("Vaga.Contratante").
		Preload("Vaga.Etapas").
		Preload("EtapaAtual").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar candidaturas: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *CandidaturaRepository) BulkUpdateStatus(ctx context.Context, vagaID uuid.UUID, cpfs []string, status empregabilidade.StatusCandidatura) (BulkUpdateResult, error) {
	var result BulkUpdateResult

	// Fetch all candidaturas for this vaga and CPF list
	var candidaturas []*empregabilidade.Candidatura
	if err := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("id_vaga = ? AND cpf IN ?", vagaID, cpfs).
		Find(&candidaturas).Error; err != nil {
		return result, fmt.Errorf("erro ao buscar candidaturas: %w", err)
	}

	foundCPFs := make(map[string]*empregabilidade.Candidatura, len(candidaturas))
	for _, c := range candidaturas {
		foundCPFs[c.CPF] = c
	}

	// Identify CPFs not found in the database
	var updateIDs []uuid.UUID
	for _, cpf := range cpfs {
		c, found := foundCPFs[cpf]
		if !found {
			result.FailedCPFs = append(result.FailedCPFs, cpf)
			continue
		}
		updateIDs = append(updateIDs, c.ID)
	}

	if len(updateIDs) == 0 {
		return result, nil
	}

	res := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("id IN ?", updateIDs).
		Update("status", status)
	if res.Error != nil {
		return result, fmt.Errorf("erro ao atualizar status em lote: %w", res.Error)
	}

	result.Updated = int(res.RowsAffected)
	return result, nil
}

func (r *CandidaturaRepository) CheckExistingCandidatura(ctx context.Context, cpf string, vagaID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("cpf = ? AND id_vaga = ?", cpf, vagaID).
		Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("erro ao verificar candidatura existente: %w", result.Error)
	}
	return count > 0, nil
}

func (r *CandidaturaRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status empregabilidade.StatusCandidatura) error {
	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar status da candidatura: %w", result.Error)
	}
	return nil
}

func (r *CandidaturaRepository) BulkGetByCPFs(ctx context.Context, vagaID uuid.UUID, cpfs []string) ([]*empregabilidade.Candidatura, error) {
	var candidaturas []*empregabilidade.Candidatura
	if err := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("id_vaga = ? AND cpf IN ?", vagaID, cpfs).
		Find(&candidaturas).Error; err != nil {
		return nil, fmt.Errorf("erro ao buscar candidaturas: %w", err)
	}
	return candidaturas, nil
}

func (r *CandidaturaRepository) BulkUpdateEtapa(ctx context.Context, ids []uuid.UUID, etapaID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("id IN ?", ids).
		Update("id_etapa_atual", etapaID)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar etapa em lote: %w", result.Error)
	}
	return nil
}

func (r *CandidaturaRepository) UpdateEtapa(ctx context.Context, id uuid.UUID, etapaID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("id = ?", id).
		Update("id_etapa_atual", etapaID)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar etapa da candidatura: %w", result.Error)
	}
	return nil
}

func (r *CandidaturaRepository) BulkSaveAndUpdateStatusByVagaID(ctx context.Context, vagaID uuid.UUID, newStatus empregabilidade.StatusCandidatura) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Save current status into status_anterior, then set new status
		result := tx.Exec(
			`UPDATE emp_candidaturas SET status_anterior = status, status = ?, updated_at = NOW() WHERE id_vaga = ? AND deleted_at IS NULL`,
			string(newStatus), vagaID,
		)
		if result.Error != nil {
			return fmt.Errorf("erro ao congelar/descontinuar candidaturas: %w", result.Error)
		}
		return nil
	})
}

func (r *CandidaturaRepository) CountByStatus(ctx context.Context, filter empregabilidade.CandidaturaFilter) (map[empregabilidade.StatusCandidatura]int64, error) {
	type statusCount struct {
		Status empregabilidade.StatusCandidatura
		Count  int64
	}
	var results []statusCount

	db := r.db.WithContext(ctx).Model(&empregabilidade.Candidatura{})

	if filter.CPF != "" {
		db = db.Where("cpf = ?", filter.CPF)
	}
	if filter.VagaID != nil {
		db = db.Where("id_vaga = ?", *filter.VagaID)
	}
	if filter.EtapaID != nil {
		db = db.Where("id_etapa_atual = ?", *filter.EtapaID)
	}
	if filter.Search != "" {
		searchTerm := fmt.Sprintf("%%%s%%", filter.Search)
		db = db.Where("cpf ILIKE ? OR nome ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm)
	}
	// Status filter is intentionally ignored to count all statuses

	if err := db.Select("status, COUNT(*) as count").Group("status").Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("erro ao contar candidaturas por status: %w", err)
	}

	counts := map[empregabilidade.StatusCandidatura]int64{
		empregabilidade.StatusCandidaturaEnviada:       0,
		empregabilidade.StatusCandidaturaAprovada:      0,
		empregabilidade.StatusCandidaturaReprovada:     0,
		empregabilidade.StatusCandidaturaVagaCongelada: 0,
		empregabilidade.StatusCandidaturaDescontinuada: 0,
	}
	for _, r := range results {
		counts[r.Status] = r.Count
	}

	return counts, nil
}

func (r *CandidaturaRepository) BulkRestoreStatusByVagaID(ctx context.Context, vagaID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(
			`UPDATE emp_candidaturas SET status = status_anterior, status_anterior = NULL, updated_at = NOW() WHERE id_vaga = ? AND deleted_at IS NULL AND status_anterior IS NOT NULL`,
			vagaID,
		)
		if result.Error != nil {
			return fmt.Errorf("erro ao restaurar status das candidaturas: %w", result.Error)
		}
		return nil
	})
}
