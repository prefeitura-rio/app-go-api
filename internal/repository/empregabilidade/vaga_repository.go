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
		Preload("OrgaoParceiro").
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
	updates := map[string]interface{}{
		"titulo":                entity.Titulo,
		"descricao":             entity.Descricao,
		"id_contratante":        entity.IDContratante,
		"id_regime_contratacao": entity.IDRegimeContratacao,
		"id_modelo_trabalho":    entity.IDModeloTrabalho,
		"acessibilidade_pcd":    entity.AcessibilidadePCD,
		"valor_vaga":            entity.ValorVaga,
		"bairro":                entity.Bairro,
		"data_limite":           entity.DataLimite,
		"requisitos":            entity.Requisitos,
		"diferenciais":          entity.Diferenciais,
		"responsabilidades":     entity.Responsabilidades,
		"beneficios":            entity.Beneficios,
		"id_orgao_parceiro":     entity.IDOrgaoParceiro,
		"status":                entity.Status,
	}

	result := r.db.WithContext(ctx).
		Model(&empregabilidade.Vaga{}).
		Where("id = ?", entity.ID).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar vaga: %w", result.Error)
	}
	return nil
}

func (r *VagaRepository) UpdateWithAssociations(ctx context.Context, entity *empregabilidade.Vaga) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"titulo":                entity.Titulo,
			"descricao":             entity.Descricao,
			"id_contratante":        entity.IDContratante,
			"id_regime_contratacao": entity.IDRegimeContratacao,
			"id_modelo_trabalho":    entity.IDModeloTrabalho,
			"acessibilidade_pcd":    entity.AcessibilidadePCD,
			"valor_vaga":            entity.ValorVaga,
			"bairro":                entity.Bairro,
			"data_limite":           entity.DataLimite,
			"requisitos":            entity.Requisitos,
			"diferenciais":          entity.Diferenciais,
			"responsabilidades":     entity.Responsabilidades,
			"beneficios":            entity.Beneficios,
			"id_orgao_parceiro":     entity.IDOrgaoParceiro,
			"status":                entity.Status,
		}

		if err := tx.Model(&empregabilidade.Vaga{}).Where("id = ?", entity.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("erro ao atualizar vaga: %w", err)
		}

		if entity.Etapas != nil {
			// Clear id_etapa_atual from all candidaturas of this vaga before deleting the steps.
			// We must clear all statuses (not only active ones) because the FK has no ON DELETE SET NULL.
			if err := tx.Model(&empregabilidade.Candidatura{}).
				Where("id_etapa_atual IN (SELECT id FROM emp_etapas WHERE id_vaga = ?)", entity.ID).
				Update("id_etapa_atual", nil).Error; err != nil {
				return fmt.Errorf("erro ao desvincular etapas das candidaturas: %w", err)
			}

			if err := tx.Where("id_vaga = ?", entity.ID).Delete(&empregabilidade.Etapa{}).Error; err != nil {
				return fmt.Errorf("erro ao remover etapas existentes: %w", err)
			}

			for i := range entity.Etapas {
				entity.Etapas[i].ID = uuid.Nil
				entity.Etapas[i].IDVaga = entity.ID
				entity.Etapas[i].Vaga = nil
			}

			if len(entity.Etapas) > 0 {
				if err := tx.Create(&entity.Etapas).Error; err != nil {
					return fmt.Errorf("erro ao criar etapas: %w", err)
				}
			}
		}

		if entity.InformacoesComplementares != nil {
			if err := tx.Where("id_vaga = ?", entity.ID).Delete(&empregabilidade.InformacaoComplementar{}).Error; err != nil {
				return fmt.Errorf("erro ao remover informações complementares existentes: %w", err)
			}

			for i := range entity.InformacoesComplementares {
				entity.InformacoesComplementares[i].ID = uuid.Nil
				entity.InformacoesComplementares[i].IDVaga = entity.ID
				entity.InformacoesComplementares[i].Vaga = nil
			}

			if len(entity.InformacoesComplementares) > 0 {
				if err := tx.Create(&entity.InformacoesComplementares).Error; err != nil {
					return fmt.Errorf("erro ao criar informações complementares: %w", err)
				}
			}
		}

		if entity.TiposPCD != nil {
			if err := tx.Exec("DELETE FROM emp_vagas_tipos_pcd WHERE id_vaga = ?", entity.ID).Error; err != nil {
				return fmt.Errorf("erro ao remover tipos PCD: %w", err)
			}
			for _, tipo := range entity.TiposPCD {
				if err := tx.Exec("INSERT INTO emp_vagas_tipos_pcd (id_vaga, id_tipo_pcd) VALUES (?, ?)", entity.ID, tipo.ID).Error; err != nil {
					return fmt.Errorf("erro ao inserir tipo PCD: %w", err)
				}
			}
		}

		return nil
	})
}

func (r *VagaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Vaga{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir vaga: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("vaga não encontrada")
	}
	return nil
}

func (r *VagaRepository) List(ctx context.Context, filter empregabilidade.VagaFilter, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.Contratante != "" {
			db = db.Where("id_contratante = ?", filter.Contratante)
		}
		if filter.OrgaoParceiroID != "" {
			db = db.Where("id_orgao_parceiro = ?", filter.OrgaoParceiroID)
		}
		if filter.Search != "" {
			searchTerm := fmt.Sprintf("%%%s%%", filter.Search)
			db = db.Where("titulo ILIKE ?", searchTerm)
		}
		if filter.Status != "" {
			switch filter.Status {
			case string(empregabilidade.StatusVagaPublicadoAtivo):
				db = db.Where("status = ? AND (data_limite IS NULL OR data_limite > NOW())", empregabilidade.StatusVagaPublicadoAtivo)
			case string(empregabilidade.StatusVagaPublicadoExpirado):
				db = db.Where("status = ? AND data_limite IS NOT NULL AND data_limite <= NOW()", empregabilidade.StatusVagaPublicadoAtivo)
			default:
				db = db.Where("status = ?", filter.Status)
			}
		}
		return db
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar vagas: %w", err)
	}

	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}))
	result := findDB.
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *VagaRepository) ListPublicActive(ctx context.Context, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}).
		Where("status = ? AND (data_limite IS NULL OR data_limite > NOW())", empregabilidade.StatusVagaPublicadoAtivo)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar vagas: %w", err)
	}

	result := db.
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas públicas ativas: %w", result.Error)
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

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar vagas por contratante: %w", err)
	}

	result := db.
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas por contratante: %w", result.Error)
	}

	return entities, int(total), nil
}

func (r *VagaRepository) GetByIDPrefix(ctx context.Context, idPrefix string) (*empregabilidade.Vaga, error) {
	var entity empregabilidade.Vaga
	result := r.db.WithContext(ctx).
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Preload("TiposPCD").
		Preload("Etapas", func(db *gorm.DB) *gorm.DB {
			return db.Order("ordem ASC")
		}).
		Preload("InformacoesComplementares").
		First(&entity, "id::text LIKE ?", idPrefix+"%")
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar vaga por slug: %w", result.Error)
	}
	return &entity, nil
}

func (r *VagaRepository) ListByOrgaoParceiro(ctx context.Context, orgaoID string, limit, offset int) ([]*empregabilidade.Vaga, int, error) {
	var entities []*empregabilidade.Vaga
	var total int64

	db := r.db.WithContext(ctx).Model(&empregabilidade.Vaga{}).Where("id_orgao_parceiro = ?", orgaoID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar vagas por órgão parceiro: %w", err)
	}

	result := db.
		Preload("Contratante").
		Preload("RegimeContratacao").
		Preload("ModeloTrabalho").
		Preload("OrgaoParceiro").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar vagas por órgão parceiro: %w", result.Error)
	}

	return entities, int(total), nil
}
