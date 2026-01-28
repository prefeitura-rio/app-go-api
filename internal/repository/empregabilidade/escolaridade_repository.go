package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type EscolaridadeRepository struct {
	db *gorm.DB
}

func NewEscolaridadeRepository(db *gorm.DB) *EscolaridadeRepository {
	return &EscolaridadeRepository{db: db}
}

func (r *EscolaridadeRepository) Create(ctx context.Context, entity *empregabilidade.Escolaridade) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar escolaridade: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *EscolaridadeRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Escolaridade, error) {
	var entity empregabilidade.Escolaridade
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar escolaridade: %w", result.Error)
	}
	return &entity, nil
}

func (r *EscolaridadeRepository) Update(ctx context.Context, entity *empregabilidade.Escolaridade) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar escolaridade: %w", result.Error)
	}
	return nil
}

func (r *EscolaridadeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Escolaridade{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir escolaridade: %w", result.Error)
	}
	return nil
}

func (r *EscolaridadeRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Escolaridade, int, error) {
	var entities []*empregabilidade.Escolaridade
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Escolaridade{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.Order("descricao ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar escolaridades: %w", result.Error)
	}

	return entities, int(total), nil
}
