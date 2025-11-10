package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type EscolaridadeRepository struct {
	db *gorm.DB
}

func NewEscolaridadeRepository(db *gorm.DB) *EscolaridadeRepository {
	return &EscolaridadeRepository{
		db: db,
	}
}

func (r *EscolaridadeRepository) Create(ctx context.Context, escolaridade *models.Escolaridade) (int, error) {
	result := r.db.WithContext(ctx).Create(escolaridade)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar escolaridade: %w", result.Error)
	}

	return escolaridade.ID, nil
}

func (r *EscolaridadeRepository) GetByID(ctx context.Context, id int) (*models.Escolaridade, error) {
	var escolaridade models.Escolaridade

	result := r.db.WithContext(ctx).First(&escolaridade, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar escolaridade por ID: %w", result.Error)
	}

	return &escolaridade, nil
}

func (r *EscolaridadeRepository) Update(ctx context.Context, escolaridade *models.Escolaridade) error {
	result := r.db.WithContext(ctx).Model(escolaridade).
		Where("id = ?", escolaridade.ID).
		Updates(escolaridade)

	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar escolaridade: %w", result.Error)
	}

	return nil
}

func (r *EscolaridadeRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.Escolaridade{}, id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir escolaridade: %w", result.Error)
	}
	return nil
}

func (r *EscolaridadeRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Escolaridade, int, error) {
	var escolaridades []*models.Escolaridade
	var total int64

	// Contar total de registros
	db := r.db.WithContext(ctx).Model(&models.Escolaridade{})
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	db.Count(&total)

	// Buscar registros com paginau00e7u00e3o
	result := db.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&escolaridades)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar escolaridades: %w", result.Error)
	}

	return escolaridades, int(total), nil
}
