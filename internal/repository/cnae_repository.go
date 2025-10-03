package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type CNAERepository struct {
	db *gorm.DB
}

func NewCNAERepository(db *gorm.DB) *CNAERepository {
	return &CNAERepository{
		db: db,
	}
}

func (r *CNAERepository) GetByID(ctx context.Context, id int) (*models.CNAE, error) {
	var cnae models.CNAE

	result := r.db.WithContext(ctx).First(&cnae, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar CNAE por ID: %w", result.Error)
	}

	return &cnae, nil
}

func (r *CNAERepository) GetByCodigo(ctx context.Context, codigo string) (*models.CNAE, error) {
	var cnae models.CNAE

	result := r.db.WithContext(ctx).Where("codigo = ?", codigo).First(&cnae)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar CNAE por código: %w", result.Error)
	}

	return &cnae, nil
}

func (r *CNAERepository) List(ctx context.Context, ocupacao string, limit, offset int) ([]*models.CNAE, int, error) {
	var cnaes []*models.CNAE
	var total int64

	db := r.db.WithContext(ctx).Model(&models.CNAE{})

	if ocupacao != "" {
		db = db.Where("ocupacao ILIKE ?", "%"+ocupacao+"%")
	}

	db.Count(&total)

	result := db.
		Order("codigo ASC").
		Limit(limit).
		Offset(offset).
		Find(&cnaes)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar CNAEs: %w", result.Error)
	}

	return cnaes, int(total), nil
}

func (r *CNAERepository) ListByOcupacao(ctx context.Context, ocupacao string) ([]*models.CNAE, error) {
	var cnaes []*models.CNAE

	result := r.db.WithContext(ctx).
		Where("ocupacao = ?", ocupacao).
		Order("servico ASC").
		Find(&cnaes)

	if result.Error != nil {
		return nil, fmt.Errorf("erro ao buscar CNAEs por ocupação: %w", result.Error)
	}

	return cnaes, nil
}

func (r *CNAERepository) Create(ctx context.Context, cnae *models.CNAE) error {
	result := r.db.WithContext(ctx).Create(cnae)
	if result.Error != nil {
		return fmt.Errorf("erro ao criar CNAE: %w", result.Error)
	}
	return nil
}

func (r *CNAERepository) Update(ctx context.Context, cnae *models.CNAE) error {
	result := r.db.WithContext(ctx).
		Model(&models.CNAE{}).
		Where("codigo = ?", cnae.Codigo).
		Updates(map[string]interface{}{
			"ocupacao": cnae.Ocupacao,
			"servico":  cnae.Servico,
		})

	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar CNAE: %w", result.Error)
	}

	return nil
}

func (r *CNAERepository) Delete(ctx context.Context, codigo string) error {
	result := r.db.WithContext(ctx).Delete(&models.CNAE{}, "codigo = ?", codigo)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir CNAE: %w", result.Error)
	}
	return nil
}
