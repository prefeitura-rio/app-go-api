package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type TipoConquistaRepository struct {
	db *gorm.DB
}

func NewTipoConquistaRepository(db *gorm.DB) *TipoConquistaRepository {
	return &TipoConquistaRepository{db: db}
}

func (r *TipoConquistaRepository) Create(ctx context.Context, entity *empregabilidade.TipoConquista) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar tipo de conquista: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *TipoConquistaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoConquista, error) {
	var entity empregabilidade.TipoConquista
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar tipo de conquista: %w", result.Error)
	}
	return &entity, nil
}

func (r *TipoConquistaRepository) Update(ctx context.Context, entity *empregabilidade.TipoConquista) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar tipo de conquista: %w", result.Error)
	}
	return nil
}

func (r *TipoConquistaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.TipoConquista{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir tipo de conquista: %w", result.Error)
	}
	return nil
}

func (r *TipoConquistaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.TipoConquista, int, error) {
	var entities []*empregabilidade.TipoConquista
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.TipoConquista{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.Order("descricao ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar tipos de conquista: %w", result.Error)
	}

	return entities, int(total), nil
}
