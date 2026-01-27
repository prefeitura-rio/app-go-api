package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type InformacaoComplementarRepository struct {
	db *gorm.DB
}

func NewInformacaoComplementarRepository(db *gorm.DB) *InformacaoComplementarRepository {
	return &InformacaoComplementarRepository{db: db}
}

func (r *InformacaoComplementarRepository) Create(ctx context.Context, entity *empregabilidade.InformacaoComplementar) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar informação complementar: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *InformacaoComplementarRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.InformacaoComplementar, error) {
	var entity empregabilidade.InformacaoComplementar
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar informação complementar: %w", result.Error)
	}
	return &entity, nil
}

func (r *InformacaoComplementarRepository) Update(ctx context.Context, entity *empregabilidade.InformacaoComplementar) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar informação complementar: %w", result.Error)
	}
	return nil
}

func (r *InformacaoComplementarRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.InformacaoComplementar{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir informação complementar: %w", result.Error)
	}
	return nil
}

func (r *InformacaoComplementarRepository) ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.InformacaoComplementar, error) {
	var entities []*empregabilidade.InformacaoComplementar

	result := r.db.WithContext(ctx).
		Where("id_vaga = ?", vagaID).
		Order("created_at ASC").
		Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("erro ao listar informações complementares: %w", result.Error)
	}

	return entities, nil
}

func (r *InformacaoComplementarRepository) DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id_vaga = ?", vagaID).Delete(&empregabilidade.InformacaoComplementar{})
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir informações complementares da vaga: %w", result.Error)
	}
	return nil
}
