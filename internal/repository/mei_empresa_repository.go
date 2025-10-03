package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type MEIEmpresaRepository struct {
	db *gorm.DB
}

func NewMEIEmpresaRepository(db *gorm.DB) *MEIEmpresaRepository {
	return &MEIEmpresaRepository{
		db: db,
	}
}

func (r *MEIEmpresaRepository) Create(ctx context.Context, meiEmpresa *models.MEIEmpresa) (int, error) {
	result := r.db.WithContext(ctx).Create(meiEmpresa)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar MEI empresa: %w", result.Error)
	}

	return meiEmpresa.ID, nil
}

func (r *MEIEmpresaRepository) GetByID(ctx context.Context, id int) (*models.MEIEmpresa, error) {
	var meiEmpresa models.MEIEmpresa

	result := r.db.WithContext(ctx).
		Preload("CNAEs").
		First(&meiEmpresa, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar MEI empresa por ID: %w", result.Error)
	}

	return &meiEmpresa, nil
}

func (r *MEIEmpresaRepository) GetByCNPJ(ctx context.Context, cnpj string) (*models.MEIEmpresa, error) {
	var meiEmpresa models.MEIEmpresa

	result := r.db.WithContext(ctx).
		Preload("CNAEs").
		Where("cnpj = ?", cnpj).
		First(&meiEmpresa)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar MEI empresa por CNPJ: %w", result.Error)
	}

	return &meiEmpresa, nil
}

func (r *MEIEmpresaRepository) Update(ctx context.Context, meiEmpresa *models.MEIEmpresa) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Atualizar dados básicos
		if err := tx.Model(meiEmpresa).
			Where("id = ?", meiEmpresa.ID).
			Updates(meiEmpresa).Error; err != nil {
			return fmt.Errorf("erro ao atualizar MEI empresa: %w", err)
		}

		// Atualizar relacionamento com CNAEs
		if err := tx.Model(meiEmpresa).Association("CNAEs").Replace(meiEmpresa.CNAEs); err != nil {
			return fmt.Errorf("erro ao atualizar CNAEs da MEI empresa: %w", err)
		}

		return nil
	})
}

func (r *MEIEmpresaRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.MEIEmpresa{}, id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir MEI empresa: %w", result.Error)
	}
	return nil
}

func (r *MEIEmpresaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.MEIEmpresa, int, error) {
	var meiEmpresas []*models.MEIEmpresa
	var total int64

	db := r.db.WithContext(ctx).Model(&models.MEIEmpresa{})
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	db.Count(&total)

	result := db.
		Preload("CNAEs").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&meiEmpresas)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar MEI empresas: %w", result.Error)
	}

	return meiEmpresas, int(total), nil
}
