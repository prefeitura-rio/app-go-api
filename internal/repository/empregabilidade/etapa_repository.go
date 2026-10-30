package empregabilidade

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// ErrEtapaNaoEncontrada indica que não há etapa com o id informado na vaga
// informada (inexistente ou pertencente a outra vaga).
var ErrEtapaNaoEncontrada = errors.New("etapa não encontrada")

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

// Update grava título, descrição e ordem da etapa entity.ID somente se ela
// pertence à vaga entity.IDVaga. Não cria etapa (sem upsert) nem muda a vaga
// dona; sem linha afetada devolve ErrEtapaNaoEncontrada.
func (r *EtapaRepository) Update(ctx context.Context, entity *empregabilidade.Etapa) error {
	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Etapa{}).
		Where("id = ? AND id_vaga = ?", entity.ID, entity.IDVaga).
		Select("titulo", "descricao", "ordem", "updated_at").
		Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar etapa: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrEtapaNaoEncontrada
	}
	return nil
}

// Delete exclui a etapa id somente se ela pertence à vaga vagaID; sem linha
// afetada devolve ErrEtapaNaoEncontrada.
func (r *EtapaRepository) Delete(ctx context.Context, vagaID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND id_vaga = ?", id, vagaID).
		Delete(&empregabilidade.Etapa{})
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir etapa: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrEtapaNaoEncontrada
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
