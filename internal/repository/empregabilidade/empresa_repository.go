package empregabilidade

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type EmpresaRepository struct {
	db *gorm.DB
}

func NewEmpresaRepository(db *gorm.DB) *EmpresaRepository {
	return &EmpresaRepository{db: db}
}

func (r *EmpresaRepository) Create(ctx context.Context, entity *empregabilidade.Empresa) (string, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return "", fmt.Errorf("erro ao criar empresa: %w", result.Error)
	}
	return entity.CNPJ, nil
}

func (r *EmpresaRepository) GetByID(ctx context.Context, cnpj string) (*empregabilidade.Empresa, error) {
	var entity empregabilidade.Empresa
	result := r.db.WithContext(ctx).First(&entity, "cnpj = ?", cnpj)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar empresa: %w", result.Error)
	}
	return &entity, nil
}

func (r *EmpresaRepository) Update(ctx context.Context, entity *empregabilidade.Empresa) error {
	result := r.db.WithContext(ctx).Model(entity).Where("cnpj = ?", entity.CNPJ).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar empresa: %w", result.Error)
	}
	return nil
}

func (r *EmpresaRepository) Delete(ctx context.Context, cnpj string) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Empresa{}, "cnpj = ?", cnpj)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir empresa: %w", result.Error)
	}
	return nil
}

func (r *EmpresaRepository) List(ctx context.Context, filter empregabilidade.EmpresaFilter, limit, offset int) ([]*empregabilidade.Empresa, int, error) {
	var entities []*empregabilidade.Empresa
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.CNPJ != "" {
			db = db.Where("cnpj = ?", filter.CNPJ)
		}
		if filter.Search != "" {
			searchTerm := fmt.Sprintf("%%%s%%", filter.Search)
			db = db.Where("razao_social ILIKE ? OR nome_fantasia ILIKE ?", searchTerm, searchTerm)
		}
		return db
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Empresa{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar empresas: %w", err)
	}

	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Empresa{}))
	result := findDB.Order("razao_social ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar empresas: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *EmpresaRepository) Upsert(ctx context.Context, entity *empregabilidade.Empresa) error {
	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao upsert empresa: %w", result.Error)
	}
	return nil
}
