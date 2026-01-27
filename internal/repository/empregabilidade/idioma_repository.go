package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type IdiomaRepository struct {
	db *gorm.DB
}

func NewIdiomaRepository(db *gorm.DB) *IdiomaRepository {
	return &IdiomaRepository{db: db}
}

func (r *IdiomaRepository) Create(ctx context.Context, entity *empregabilidade.Idioma) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar idioma: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *IdiomaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Idioma, error) {
	var entity empregabilidade.Idioma
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar idioma: %w", result.Error)
	}
	return &entity, nil
}

func (r *IdiomaRepository) Update(ctx context.Context, entity *empregabilidade.Idioma) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar idioma: %w", result.Error)
	}
	return nil
}

func (r *IdiomaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Idioma{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir idioma: %w", result.Error)
	}
	return nil
}

func (r *IdiomaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Idioma, int, error) {
	var entities []*empregabilidade.Idioma
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Idioma{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.Order("descricao ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar idiomas: %w", result.Error)
	}

	return entities, int(total), nil
}
