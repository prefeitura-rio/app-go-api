package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type ModeloTrabalhoRepository struct {
	db *gorm.DB
}

func NewModeloTrabalhoRepository(db *gorm.DB) *ModeloTrabalhoRepository {
	return &ModeloTrabalhoRepository{db: db}
}

func (r *ModeloTrabalhoRepository) Create(ctx context.Context, entity *empregabilidade.ModeloTrabalho) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar modelo de trabalho: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *ModeloTrabalhoRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.ModeloTrabalho, error) {
	var entity empregabilidade.ModeloTrabalho
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar modelo de trabalho: %w", result.Error)
	}
	return &entity, nil
}

func (r *ModeloTrabalhoRepository) Update(ctx context.Context, entity *empregabilidade.ModeloTrabalho) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar modelo de trabalho: %w", result.Error)
	}
	return nil
}

func (r *ModeloTrabalhoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.ModeloTrabalho{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir modelo de trabalho: %w", result.Error)
	}
	return nil
}

func (r *ModeloTrabalhoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.ModeloTrabalho, int, error) {
	var entities []*empregabilidade.ModeloTrabalho
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.ModeloTrabalho{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.Order("descricao ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar modelos de trabalho: %w", result.Error)
	}

	return entities, int(total), nil
}
