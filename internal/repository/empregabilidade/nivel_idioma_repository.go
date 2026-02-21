package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type NivelIdiomaRepository struct {
	db *gorm.DB
}

func NewNivelIdiomaRepository(db *gorm.DB) *NivelIdiomaRepository {
	return &NivelIdiomaRepository{db: db}
}

func (r *NivelIdiomaRepository) Create(ctx context.Context, entity *empregabilidade.NivelIdioma) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar nível de idioma: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *NivelIdiomaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.NivelIdioma, error) {
	var entity empregabilidade.NivelIdioma
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar nível de idioma: %w", result.Error)
	}
	return &entity, nil
}

func (r *NivelIdiomaRepository) Update(ctx context.Context, entity *empregabilidade.NivelIdioma) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar nível de idioma: %w", result.Error)
	}
	return nil
}

func (r *NivelIdiomaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.NivelIdioma{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir nível de idioma: %w", result.Error)
	}
	return nil
}

func (r *NivelIdiomaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.NivelIdioma, int, error) {
	var entities []*empregabilidade.NivelIdioma
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.NivelIdioma{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.Order("descricao ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar níveis de idioma: %w", result.Error)
	}

	return entities, int(total), nil
}
