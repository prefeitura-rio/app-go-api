package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type RegimeContratacaoRepository struct {
	db *gorm.DB
}

func NewRegimeContratacaoRepository(db *gorm.DB) *RegimeContratacaoRepository {
	return &RegimeContratacaoRepository{db: db}
}

func (r *RegimeContratacaoRepository) Create(ctx context.Context, entity *empregabilidade.RegimeContratacao) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar regime de contratação: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *RegimeContratacaoRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.RegimeContratacao, error) {
	var entity empregabilidade.RegimeContratacao
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar regime de contratação: %w", result.Error)
	}
	return &entity, nil
}

func (r *RegimeContratacaoRepository) Update(ctx context.Context, entity *empregabilidade.RegimeContratacao) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar regime de contratação: %w", result.Error)
	}
	return nil
}

func (r *RegimeContratacaoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.RegimeContratacao{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir regime de contratação: %w", result.Error)
	}
	return nil
}

func (r *RegimeContratacaoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.RegimeContratacao, int, error) {
	var entities []*empregabilidade.RegimeContratacao
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.RegimeContratacao{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.Order("descricao ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar regimes de contratação: %w", result.Error)
	}

	return entities, int(total), nil
}
