package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type InscricaoRepository struct {
	db *gorm.DB
}

func NewInscricaoRepository(db *gorm.DB) *InscricaoRepository {
	return &InscricaoRepository{
		db: db,
	}
}

func (r *InscricaoRepository) Create(ctx context.Context, inscricao *models.Inscricao) error {
	result := r.db.WithContext(ctx).Create(inscricao)
	if result.Error != nil {
		return fmt.Errorf("erro ao criar inscrição: %w", result.Error)
	}
	return nil
}

func (r *InscricaoRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Inscricao, error) {
	var inscricao models.Inscricao
	
	result := r.db.WithContext(ctx).
		Preload("Curso").
		First(&inscricao, "id = ?", id)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar inscrição por ID: %w", result.Error)
	}
	
	return &inscricao, nil
}

func (r *InscricaoRepository) GetByCursoID(ctx context.Context, cursoID int, filter map[string]interface{}, limit, offset int) ([]*models.Inscricao, int, error) {
	var inscricoes []*models.Inscricao
	var total int64
	
	// Base query for the course
	baseQuery := r.db.WithContext(ctx).Model(&models.Inscricao{}).Where("curso_id = ?", cursoID)
	
	// Apply filters
	for key, value := range filter {
		switch key {
		case "status":
			baseQuery = baseQuery.Where("status = ?", value)
		case "search":
			// Search in CPF, name, or email
			searchTerm := fmt.Sprintf("%%%s%%", value)
			baseQuery = baseQuery.Where("cpf ILIKE ? OR name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm)
		}
	}
	
	// Count total records
	baseQuery.Count(&total)
	
	// Get paginated results
	result := baseQuery.
		Order("enrolled_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&inscricoes)
		
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar inscrições: %w", result.Error)
	}
	
	return inscricoes, int(total), nil
}

func (r *InscricaoRepository) UpdateStatus(ctx context.Context, inscricaoID uuid.UUID, status models.StatusInscricao, reason, adminNotes string) error {
	updates := map[string]interface{}{
		"status":      status,
		"updated_at":  time.Now(),
	}
	
	if reason != "" {
		updates["reason"] = reason
	}
	
	if adminNotes != "" {
		updates["admin_notes"] = adminNotes
	}
	
	if status == models.StatusInscricaoConcluded {
		now := time.Now()
		updates["concluded_at"] = now
	}
	
	result := r.db.WithContext(ctx).
		Model(&models.Inscricao{}).
		Where("id = ?", inscricaoID).
		Updates(updates)
	
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar status da inscrição: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("inscrição não encontrada")
	}
	
	return nil
}

func (r *InscricaoRepository) UpdateMultipleStatus(ctx context.Context, inscricaoIDs []uuid.UUID, status models.StatusInscricao, reason, adminNotes string) (int, error) {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	
	if reason != "" {
		updates["reason"] = reason
	}
	
	if adminNotes != "" {
		updates["admin_notes"] = adminNotes
	}
	
	if status == models.StatusInscricaoConcluded {
		now := time.Now()
		updates["concluded_at"] = now
	}
	
	result := r.db.WithContext(ctx).
		Model(&models.Inscricao{}).
		Where("id IN ?", inscricaoIDs).
		Updates(updates)
	
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao atualizar status das inscrições: %w", result.Error)
	}
	
	return int(result.RowsAffected), nil
}

func (r *InscricaoRepository) GetSummaryByCursoID(ctx context.Context, cursoID int) (*models.EnrollmentSummary, error) {
	var summary models.EnrollmentSummary
	var total int64
	
	// Count total enrollments
	r.db.WithContext(ctx).
		Model(&models.Inscricao{}).
		Where("curso_id = ?", cursoID).
		Count(&total)
	
	summary.Total = int(total)
	
	// Count by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	
	r.db.WithContext(ctx).
		Model(&models.Inscricao{}).
		Select("status, count(*) as count").
		Where("curso_id = ?", cursoID).
		Group("status").
		Scan(&statusCounts)
	
	// Map counts to summary
	for _, sc := range statusCounts {
		switch models.StatusInscricao(sc.Status) {
		case models.StatusInscricaoPending:
			summary.Pending = int(sc.Count)
		case models.StatusInscricaoApproved:
			summary.Approved = int(sc.Count)
		case models.StatusInscricaoRejected:
			summary.Rejected = int(sc.Count)
		case models.StatusInscricaoCancelled:
			summary.Cancelled = int(sc.Count)
		case models.StatusInscricaoConcluded:
			summary.Concluded = int(sc.Count)
		}
	}
	
	return &summary, nil
}

func (r *InscricaoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Inscricao{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir inscrição: %w", result.Error)
	}
	return nil
}

func (r *InscricaoRepository) ExistsByCPFAndCurso(ctx context.Context, cpf string, cursoID int) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&models.Inscricao{}).
		Where("cpf = ? AND curso_id = ?", cpf, cursoID).
		Count(&count)
	
	if result.Error != nil {
		return false, fmt.Errorf("erro ao verificar inscrição existente: %w", result.Error)
	}
	
	return count > 0, nil
}

func (r *InscricaoRepository) ListByCPF(ctx context.Context, cpf string, filter map[string]interface{}, offset, limit int) ([]*models.Inscricao, int, error) {
	var inscricoes []*models.Inscricao
	var total int64

	// Base query
	query := r.db.WithContext(ctx).Model(&models.Inscricao{}).Where("cpf = ?", cpf)

	// Apply additional filters
	for key, value := range filter {
		if key != "cpf" { // cpf already applied above
			query = query.Where(key+" = ?", value)
		}
	}

	// Count total
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar inscrições: %w", err)
	}

	// Get records with pagination and preload course data
	result := query.
		Preload("Curso").
		Preload("Curso.Orgao").
		Preload("Curso.Instituicao").
		Offset(offset).
		Limit(limit).
		Order("enrolled_at DESC").
		Find(&inscricoes)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao buscar inscrições por CPF: %w", result.Error)
	}

	return inscricoes, int(total), nil
}

func (r *InscricaoRepository) UpdateCertificate(ctx context.Context, inscricaoID uuid.UUID, certificateURL string) error {
	updates := map[string]interface{}{
		"certificate_url": certificateURL,
		"updated_at":      time.Now(),
	}
	
	result := r.db.WithContext(ctx).
		Model(&models.Inscricao{}).
		Where("id = ?", inscricaoID).
		Updates(updates)
	
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar certificado: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("inscrição não encontrada")
	}
	
	return nil
}