package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type CandidaturaRepository struct {
	db *gorm.DB
}

func NewCandidaturaRepository(db *gorm.DB) *CandidaturaRepository {
	return &CandidaturaRepository{db: db}
}

func (r *CandidaturaRepository) Create(ctx context.Context, entity *empregabilidade.Candidatura) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar candidatura: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *CandidaturaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Candidatura, error) {
	var entity empregabilidade.Candidatura
	result := r.db.WithContext(ctx).
		Preload("Vaga").
		Preload("Vaga.Contratante").
		Preload("EtapaAtual").
		First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar candidatura: %w", result.Error)
	}
	return &entity, nil
}

func (r *CandidaturaRepository) Update(ctx context.Context, entity *empregabilidade.Candidatura) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar candidatura: %w", result.Error)
	}
	return nil
}

func (r *CandidaturaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Candidatura{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir candidatura: %w", result.Error)
	}
	return nil
}

func (r *CandidaturaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Candidatura, int, error) {
	var entities []*empregabilidade.Candidatura
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Candidatura{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.
		Preload("Vaga").
		Preload("Vaga.Contratante").
		Preload("EtapaAtual").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar candidaturas: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *CandidaturaRepository) ListByCPF(ctx context.Context, cpf string, limit, offset int) ([]*empregabilidade.Candidatura, int, error) {
	var entities []*empregabilidade.Candidatura
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Candidatura{}).Where("cpf = ?", cpf)

	db.Count(&total)

	result := db.
		Preload("Vaga").
		Preload("Vaga.Contratante").
		Preload("EtapaAtual").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar candidaturas por CPF: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *CandidaturaRepository) ListByVaga(ctx context.Context, vagaID uuid.UUID, status string, limit, offset int) ([]*empregabilidade.Candidatura, int, error) {
	var entities []*empregabilidade.Candidatura
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Candidatura{}).Where("id_vaga = ?", vagaID)

	if status != "" {
		db = db.Where("status = ?", status)
	}

	db.Count(&total)

	result := db.
		Preload("EtapaAtual").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar candidaturas por vaga: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *CandidaturaRepository) CheckExistingCandidatura(ctx context.Context, cpf string, vagaID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("cpf = ? AND id_vaga = ?", cpf, vagaID).
		Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("erro ao verificar candidatura existente: %w", result.Error)
	}
	return count > 0, nil
}

func (r *CandidaturaRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status empregabilidade.StatusCandidatura) error {
	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar status da candidatura: %w", result.Error)
	}
	return nil
}

func (r *CandidaturaRepository) UpdateEtapa(ctx context.Context, id uuid.UUID, etapaID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Candidatura{}).
		Where("id = ?", id).
		Update("id_etapa_atual", etapaID)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar etapa da candidatura: %w", result.Error)
	}
	return nil
}
