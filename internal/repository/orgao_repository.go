package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type OrgaoRepository struct {
	db *gorm.DB
}

func NewOrgaoRepository(db *gorm.DB) *OrgaoRepository {
	return &OrgaoRepository{
		db: db,
	}
}

func (r *OrgaoRepository) Create(ctx context.Context, orgao *models.Orgao) (int, error) {
	result := r.db.WithContext(ctx).Create(orgao)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar u00f3rgu00e3o: %w", result.Error)
	}
	
	return orgao.ID, nil
}

func (r *OrgaoRepository) GetByID(ctx context.Context, id int) (*models.Orgao, error) {
	var orgao models.Orgao
	
	result := r.db.WithContext(ctx).First(&orgao, id)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar u00f3rgu00e3o por ID: %w", result.Error)
	}
	
	return &orgao, nil
}

func (r *OrgaoRepository) Update(ctx context.Context, orgao *models.Orgao) error {
	result := r.db.WithContext(ctx).Model(orgao).
		Where("id = ?", orgao.ID).
		Updates(orgao)
	
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar u00f3rgu00e3o: %w", result.Error)
	}
	
	return nil
}

func (r *OrgaoRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.Orgao{}, id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir u00f3rgu00e3o: %w", result.Error)
	}
	return nil
}

func (r *OrgaoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Orgao, int, error) {
	var orgaos []*models.Orgao
	var total int64
	
	// Contar total de registros
	db := r.db.WithContext(ctx).Model(&models.Orgao{})
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	db.Count(&total)
	
	// Buscar registros com paginau00e7u00e3o
	result := db.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&orgaos)
		
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar u00f3rgu00e3os: %w", result.Error)
	}
	
	return orgaos, int(total), nil
}
