package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type OportunidadeMEIRepository struct {
	db *gorm.DB
}

func NewOportunidadeMEIRepository(db *gorm.DB) *OportunidadeMEIRepository {
	return &OportunidadeMEIRepository{
		db: db,
	}
}

func (r *OportunidadeMEIRepository) Create(ctx context.Context, oportunidade *models.OportunidadeMEI) (int, error) {
	result := r.db.WithContext(ctx).Create(oportunidade)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar oportunidade MEI: %w", result.Error)
	}

	return oportunidade.ID, nil
}

func (r *OportunidadeMEIRepository) GetByID(ctx context.Context, id int) (*models.OportunidadeMEI, error) {
	var oportunidade models.OportunidadeMEI

	result := r.db.WithContext(ctx).
		Preload("Orgao").
		Preload("CNAE").
		First(&oportunidade, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar oportunidade MEI por ID: %w", result.Error)
	}

	return &oportunidade, nil
}

func (r *OportunidadeMEIRepository) Update(ctx context.Context, oportunidade *models.OportunidadeMEI) error {
	result := r.db.WithContext(ctx).Model(oportunidade).
		Where("id = ?", oportunidade.ID).
		Updates(oportunidade)

	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar oportunidade MEI: %w", result.Error)
	}

	return nil
}

func (r *OportunidadeMEIRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.OportunidadeMEI{}, id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir oportunidade MEI: %w", result.Error)
	}
	return nil
}

func (r *OportunidadeMEIRepository) List(ctx context.Context, filters map[string]interface{}, titulo string, limit, offset int) ([]*models.OportunidadeMEI, int, error) {
	var oportunidades []*models.OportunidadeMEI
	var total int64

	db := r.db.WithContext(ctx).Model(&models.OportunidadeMEI{})

	// Aplicar filtros exatos
	for key, value := range filters {
		db = db.Where(key+" = ?", value)
	}

	// Filtro de busca por título (case-insensitive)
	if titulo != "" {
		db = db.Where("titulo ILIKE ?", "%"+titulo+"%")
	}

	db.Count(&total)

	result := db.
		Preload("Orgao").
		Preload("CNAE").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&oportunidades)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar oportunidades MEI: %w", result.Error)
	}

	return oportunidades, int(total), nil
}

func (r *OportunidadeMEIRepository) ListByStatus(ctx context.Context, status models.StatusOportunidadeMEI, limit, offset int) ([]*models.OportunidadeMEI, int, error) {
	var oportunidades []*models.OportunidadeMEI
	var total int64

	db := r.db.WithContext(ctx).Model(&models.OportunidadeMEI{}).
		Where("status = ?", status)

	db.Count(&total)

	result := db.
		Preload("Orgao").
		Preload("CNAE").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&oportunidades)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar oportunidades MEI por status: %w", result.Error)
	}

	return oportunidades, int(total), nil
}

func (r *OportunidadeMEIRepository) ListByOrgao(ctx context.Context, orgaoID int, limit, offset int) ([]*models.OportunidadeMEI, int, error) {
	var oportunidades []*models.OportunidadeMEI
	var total int64

	db := r.db.WithContext(ctx).Model(&models.OportunidadeMEI{}).
		Where("orgao_id = ?", orgaoID)

	db.Count(&total)

	result := db.
		Preload("Orgao").
		Preload("CNAE").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&oportunidades)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar oportunidades MEI por órgão: %w", result.Error)
	}

	return oportunidades, int(total), nil
}

func (r *OportunidadeMEIRepository) UpdateExpiredOpportunities(ctx context.Context) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.OportunidadeMEI{}).
		Where("status = ? AND data_expiracao < ?", models.StatusOportunidadeActive, now).
		Update("status", models.StatusOportunidadeExpired)

	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar oportunidades expiradas: %w", result.Error)
	}

	return nil
}
