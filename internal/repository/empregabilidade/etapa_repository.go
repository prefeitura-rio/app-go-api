package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type EtapaRepository struct {
	db *gorm.DB
}

func NewEtapaRepository(db *gorm.DB) *EtapaRepository {
	return &EtapaRepository{db: db}
}

func (r *EtapaRepository) Create(ctx context.Context, entity *empregabilidade.Etapa) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar etapa: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *EtapaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Etapa, error) {
	var entity empregabilidade.Etapa
	result := r.db.WithContext(ctx).First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar etapa: %w", result.Error)
	}
	return &entity, nil
}

func (r *EtapaRepository) Update(ctx context.Context, entity *empregabilidade.Etapa) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar etapa: %w", result.Error)
	}
	return nil
}

func (r *EtapaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Etapa{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir etapa: %w", result.Error)
	}
	return nil
}

func (r *EtapaRepository) ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.Etapa, error) {
	var entities []*empregabilidade.Etapa

	result := r.db.WithContext(ctx).
		Where("id_vaga = ?", vagaID).
		Order("ordem ASC").
		Find(&entities)
	if result.Error != nil {
		return nil, fmt.Errorf("erro ao listar etapas: %w", result.Error)
	}

	return entities, nil
}

func (r *EtapaRepository) DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id_vaga = ?", vagaID).Delete(&empregabilidade.Etapa{})
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir etapas da vaga: %w", result.Error)
	}
	return nil
}
