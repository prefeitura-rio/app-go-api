package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type ZonaRepository struct {
	db *gorm.DB
}

func NewZonaRepository(db *gorm.DB) *ZonaRepository {
	return &ZonaRepository{db: db}
}

func (r *ZonaRepository) Create(ctx context.Context, entity *empregabilidade.Zona) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar zona: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *ZonaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Zona, error) {
	var entity empregabilidade.Zona
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar zona: %w", result.Error)
	}
	return &entity, nil
}

func (r *ZonaRepository) Update(ctx context.Context, entity *empregabilidade.Zona) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar zona: %w", result.Error)
	}
	return nil
}

func (r *ZonaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Zona{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir zona: %w", result.Error)
	}
	return nil
}

func (r *ZonaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Zona, int, error) {
	var entities []*empregabilidade.Zona
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Zona{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.Order("descricao ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar zonas: %w", result.Error)
	}

	return entities, int(total), nil
}
