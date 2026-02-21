package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type SituacaoAtualRepository struct {
	db *gorm.DB
}

func NewSituacaoAtualRepository(db *gorm.DB) *SituacaoAtualRepository {
	return &SituacaoAtualRepository{db: db}
}

func (r *SituacaoAtualRepository) Create(ctx context.Context, entity *empregabilidade.SituacaoAtual) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar situação atual: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *SituacaoAtualRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.SituacaoAtual, error) {
	var entity empregabilidade.SituacaoAtual
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar situação atual: %w", result.Error)
	}
	return &entity, nil
}

func (r *SituacaoAtualRepository) Update(ctx context.Context, entity *empregabilidade.SituacaoAtual) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar situação atual: %w", result.Error)
	}
	return nil
}

func (r *SituacaoAtualRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.SituacaoAtual{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir situação atual: %w", result.Error)
	}
	return nil
}

func (r *SituacaoAtualRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.SituacaoAtual, int, error) {
	var entities []*empregabilidade.SituacaoAtual
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.SituacaoAtual{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.Order("descricao ASC").Limit(limit).Offset(offset).Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar situações atuais: %w", result.Error)
	}

	return entities, int(total), nil
}
