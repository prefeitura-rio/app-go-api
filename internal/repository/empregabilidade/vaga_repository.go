package empregabilidade

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type VagaRepository struct {
	db *gorm.DB
}

func NewVagaRepository(db *gorm.DB) *VagaRepository {
	return &VagaRepository{db: db}
}

func (r *VagaRepository) Create(ctx context.Context, entity *empregabilidade.Vaga) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar vaga: %w", result.Error)
	}
	return entity.ID, nil
}

func (r *VagaRepository) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Vaga, error) {
	var entity empregabilidade.Vaga
	result := r.db.WithContext(ctx).
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("TiposPCD").
		Preload("Etapas", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC")
		}).
		Preload("InformacoesComplementares").
		First(&entity, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar vaga: %w", result.Error)
	}
	return &entity, nil
}

func (r *VagaRepository) Update(ctx context.Context, entity *empregabilidade.Vaga) error {
	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar vaga: %w", result.Error)
	}
	return nil
}

func (r *VagaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Vaga{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir vaga: %w", result.Error)
	}
	return nil
}

func (r *VagaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Vaga{})

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *VagaRepository) UpdateTiposPCD(ctx context.Context, vagaID uuid.UUID, tiposPCDIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM emp_vagas_tipos_pcd WHERE id_vaga = ?", vagaID).Error; err != nil {
			return fmt.Errorf("erro ao remover tipos PCD: %w", err)
		}

		for _, tipoPCDID := range tiposPCDIDs {
			if err := tx.Exec("INSERT INTO emp_vagas_tipos_pcd (id_vaga, id_tipo_pcd) VALUES (?, ?)", vagaID, tipoPCDID).Error; err != nil {
				return fmt.Errorf("erro ao inserir tipo PCD: %w", err)
			}
		}

		return nil
	})
}

func (r *VagaRepository) ListByContratante(ctx context.Context, cnpj string, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}).Where("id_contratante = ?", cnpj)

	db.Count(&total)

	result := db.
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas por contratante: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *VagaRepository) ListByOrgaoParceiro(ctx context.Context, orgaoID string, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}).Where("id_orgao_parceiro = ?", orgaoID)

	db.Count(&total)

	result := db.
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas por órgão parceiro: %w", result.Error)
	}

	return entities, int(total), nil
}
