package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type TipoPCDRepository struct {
	db *gorm.DB
}

func NewTipoPCDRepository(db *gorm.DB) *TipoPCDRepository {
	return &TipoPCDRepository{db: db}
}

func (r *TipoPCDRepository) Create(ctx context.Context, entity *empregabilidade.TipoPCD) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar tipo PCD: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *TipoPCDRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoPCD, error) {
	var entity empregabilidade.TipoPCD
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar tipo PCD: %w", result.Error)
	}
	return &entity, nil
}

func (r *TipoPCDRepository) Update(ctx context.Context, entity *empregabilidade.TipoPCD) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar tipo PCD: %w", result.Error)
	}
	return nil
}

func (r *TipoPCDRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.TipoPCD{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir tipo PCD: %w", result.Error)
	}
	return nil
}

func (r *TipoPCDRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.TipoPCD, int, error) {
	var entities []*empregabilidade.TipoPCD
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.TipoPCD{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.Order("descricao ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar tipos PCD: %w", result.Error)
	}

	return entities, int(total), nil
}
