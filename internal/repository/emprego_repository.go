package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type EmpregoRepository struct {
	db *gorm.DB
}

func NewEmpregoRepository(db *gorm.DB) *EmpregoRepository {
	return &EmpregoRepository{
		db: db,
	}
}

func (r *EmpregoRepository) Create(ctx context.Context, emprego *models.Emprego) (int, error) {
	result := r.db.WithContext(ctx).Create(emprego)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar emprego: %w", result.Error)
	}
	
	return emprego.ID, nil
}

func (r *EmpregoRepository) GetByID(ctx context.Context, id int) (*models.Emprego, error) {
	var emprego models.Emprego
	
	result := r.db.WithContext(ctx).
		Preload("Orgao").
		Preload("Empresa").
		Preload("Escolaridade").
		First(&emprego, id)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar emprego por ID: %w", result.Error)
	}
	
	return &emprego, nil
}

func (r *EmpregoRepository) Update(ctx context.Context, emprego *models.Emprego) error {
	result := r.db.WithContext(ctx).Model(emprego).
		Where("id = ?", emprego.ID).
		Omit("Orgao", "Empresa", "Escolaridade"). // Ignorar relações
		Updates(emprego)
	
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar emprego: %w", result.Error)
	}
	
	return nil
}

func (r *EmpregoRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.Emprego{}, id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir emprego: %w", result.Error)
	}
	return nil
}

func (r *EmpregoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Emprego, int, error) {
	var empregos []*models.Emprego
	var total int64
	
	// Contar total de registros
	db := r.db.WithContext(ctx).Model(&models.Emprego{})
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	db.Count(&total)
	
	// Buscar registros com paginação
	result := db.
		Preload("Orgao").
		Preload("Empresa").
		Preload("Escolaridade").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&empregos)
		
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar empregos: %w", result.Error)
	}
	
	return empregos, int(total), nil
} 