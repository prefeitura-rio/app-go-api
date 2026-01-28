package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type DisponibilidadeRepository struct {
	db *gorm.DB
}

func NewDisponibilidadeRepository(db *gorm.DB) *DisponibilidadeRepository {
	return &DisponibilidadeRepository{db: db}
}

func (r *DisponibilidadeRepository) Create(ctx context.Context, entity *empregabilidade.Disponibilidade) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar disponibilidade: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *DisponibilidadeRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Disponibilidade, error) {
	var entity empregabilidade.Disponibilidade
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar disponibilidade: %w", result.Error)
	}
	return &entity, nil
}

func (r *DisponibilidadeRepository) Update(ctx context.Context, entity *empregabilidade.Disponibilidade) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar disponibilidade: %w", result.Error)
	}
	return nil
}

func (r *DisponibilidadeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Disponibilidade{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir disponibilidade: %w", result.Error)
	}
	return nil
}

func (r *DisponibilidadeRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Disponibilidade, int, error) {
	var entities []*empregabilidade.Disponibilidade
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Disponibilidade{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.Order("descricao ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar disponibilidades: %w", result.Error)
	}

	return entities, int(total), nil
}
