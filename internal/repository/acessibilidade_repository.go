package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type AcessibilidadeRepository struct {
	db *gorm.DB
}

func NewAcessibilidadeRepository(db *gorm.DB) *AcessibilidadeRepository {
	return &AcessibilidadeRepository{
		db: db,
	}
}

func (r *AcessibilidadeRepository) Create(ctx context.Context, acessibilidade *models.Acessibilidade) (int, error) {
	result := r.db.WithContext(ctx).Create(acessibilidade)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar acessibilidade: %w", result.Error)
	}

	return acessibilidade.ID, nil
}

func (r *AcessibilidadeRepository) GetByID(ctx context.Context, id int) (*models.Acessibilidade, error) {
	var acessibilidade models.Acessibilidade

	result := r.db.WithContext(ctx).First(&acessibilidade, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar acessibilidade por ID: %w", result.Error)
	}

	return &acessibilidade, nil
}

func (r *AcessibilidadeRepository) Update(ctx context.Context, acessibilidade *models.Acessibilidade) error {
	result := r.db.WithContext(ctx).Model(acessibilidade).
		Where("id = ?", acessibilidade.ID).
		Updates(acessibilidade)

	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar acessibilidade: %w", result.Error)
	}

	return nil
}

func (r *AcessibilidadeRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.Acessibilidade{}, id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir acessibilidade: %w", result.Error)
	}
	return nil
}

func (r *AcessibilidadeRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Acessibilidade, int, error) {
	var acessibilidades []*models.Acessibilidade
	var total int64

	// Contar total de registros
	db := r.db.WithContext(ctx).Model(&models.Acessibilidade{})
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	db.Count(&total)

	// Buscar registros com paginação
	result := db.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&acessibilidades)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar acessibilidades: %w", result.Error)
	}

	return acessibilidades, int(total), nil
}
