package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type EmpresaRepository struct {
	db *gorm.DB
}

func NewEmpresaRepository(db *gorm.DB) *EmpresaRepository {
	return &EmpresaRepository{
		db: db,
	}
}

func (r *EmpresaRepository) Create(ctx context.Context, empresa *models.Empresa) (int, error) {
	result := r.db.WithContext(ctx).Create(empresa)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar empresa: %w", result.Error)
	}
	
	return empresa.ID, nil
}

func (r *EmpresaRepository) GetByID(ctx context.Context, id int) (*models.Empresa, error) {
	var empresa models.Empresa
	
	result := r.db.WithContext(ctx).First(&empresa, id)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar empresa por ID: %w", result.Error)
	}
	
	return &empresa, nil
}

func (r *EmpresaRepository) Update(ctx context.Context, empresa *models.Empresa) error {
	result := r.db.WithContext(ctx).Model(empresa).
		Where("id = ?", empresa.ID).
		Updates(empresa)
	
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar empresa: %w", result.Error)
	}
	
	return nil
}

func (r *EmpresaRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.Empresa{}, id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir empresa: %w", result.Error)
	}
	return nil
}

func (r *EmpresaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Empresa, int, error) {
	var empresas []*models.Empresa
	var total int64
	
	// Contar total de registros
	db := r.db.WithContext(ctx).Model(&models.Empresa{})
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	db.Count(&total)
	
	// Buscar registros com paginação
	result := db.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&empresas)
		
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar empresas: %w", result.Error)
	}
	
	return empresas, int(total), nil
}
