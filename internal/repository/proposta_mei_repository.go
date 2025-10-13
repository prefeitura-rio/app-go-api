package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type PropostaMEIRepository struct {
	db *gorm.DB
}

func NewPropostaMEIRepository(db *gorm.DB) *PropostaMEIRepository {
	return &PropostaMEIRepository{
		db: db,
	}
}

func (r *PropostaMEIRepository) Create(ctx context.Context, proposta *models.PropostaMEI) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(proposta)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar proposta MEI: %w", result.Error)
	}

	return proposta.ID, nil
}

func (r *PropostaMEIRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PropostaMEI, error) {
	var proposta models.PropostaMEI

	result := r.db.WithContext(ctx).
		Preload("Oportunidade").
		Preload("Oportunidade.Orgao").
		Preload("Oportunidade.CNAE").
		Preload("MEIEmpresa").
		Preload("MEIEmpresa.CNAEs").
		Where("id = ?", id).
		First(&proposta)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar proposta MEI por ID: %w", result.Error)
	}

	return &proposta, nil
}

func (r *PropostaMEIRepository) Update(ctx context.Context, proposta *models.PropostaMEI) error {
	result := r.db.WithContext(ctx).Model(proposta).
		Where("id = ?", proposta.ID).
		Updates(proposta)

	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar proposta MEI: %w", result.Error)
	}

	return nil
}

func (r *PropostaMEIRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.PropostaMEI{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir proposta MEI: %w", result.Error)
	}
	return nil
}

func (r *PropostaMEIRepository) ListByOportunidade(ctx context.Context, oportunidadeID int, nomeEmpresa, cnpj, status string, limit, offset int) ([]*models.PropostaMEI, int, error) {
	var propostas []*models.PropostaMEI
	var total int64

	db := r.db.WithContext(ctx).Model(&models.PropostaMEI{}).
		Joins("JOIN mei_empresas ON mei_empresas.id = propostas_mei.mei_empresa_id").
		Where("propostas_mei.oportunidade_mei_id = ?", oportunidadeID)

	// Filtro de busca por nome da empresa (case-insensitive)
	if nomeEmpresa != "" {
		db = db.Where("mei_empresas.razao_social ILIKE ?", "%"+nomeEmpresa+"%")
	}

	// Filtro de busca por CNPJ (permite busca parcial)
	if cnpj != "" {
		db = db.Where("mei_empresas.cnpj ILIKE ?", "%"+cnpj+"%")
	}

	// Filtro por status
	if status != "" {
		db = db.Where("propostas_mei.status_cidadao = ?", status)
	}

	db.Count(&total)

	result := db.
		Preload("MEIEmpresa").
		Preload("MEIEmpresa.CNAEs").
		Order("propostas_mei.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&propostas)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar propostas MEI por oportunidade: %w", result.Error)
	}

	return propostas, int(total), nil
}

func (r *PropostaMEIRepository) ListByMEIEmpresa(ctx context.Context, meiEmpresaID int, limit, offset int) ([]*models.PropostaMEI, int, error) {
	var propostas []*models.PropostaMEI
	var total int64

	db := r.db.WithContext(ctx).Model(&models.PropostaMEI{}).
		Where("mei_empresa_id = ?", meiEmpresaID)

	db.Count(&total)

	result := db.
		Preload("Oportunidade").
		Preload("Oportunidade.Orgao").
		Preload("Oportunidade.CNAE").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&propostas)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar propostas MEI por empresa: %w", result.Error)
	}

	return propostas, int(total), nil
}

func (r *PropostaMEIRepository) ListByStatus(ctx context.Context, statusCidadao models.StatusPropostaCidadao, limit, offset int) ([]*models.PropostaMEI, int, error) {
	var propostas []*models.PropostaMEI
	var total int64

	db := r.db.WithContext(ctx).Model(&models.PropostaMEI{}).
		Where("status_cidadao = ?", statusCidadao)

	db.Count(&total)

	result := db.
		Preload("Oportunidade").
		Preload("Oportunidade.Orgao").
		Preload("Oportunidade.CNAE").
		Preload("MEIEmpresa").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&propostas)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar propostas MEI por status: %w", result.Error)
	}

	return propostas, int(total), nil
}

func (r *PropostaMEIRepository) CheckExistingProposta(ctx context.Context, oportunidadeID int, meiEmpresaID int) (bool, error) {
	var count int64

	result := r.db.WithContext(ctx).
		Model(&models.PropostaMEI{}).
		Where("oportunidade_mei_id = ? AND mei_empresa_id = ?", oportunidadeID, meiEmpresaID).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("erro ao verificar proposta existente: %w", result.Error)
	}

	return count > 0, nil
}

func (r *PropostaMEIRepository) UpdateMultipleStatus(ctx context.Context, propostaIDs []uuid.UUID, status models.StatusPropostaCidadao) (int, error) {
	updates := map[string]interface{}{
		"status_cidadao": status,
	}

	result := r.db.WithContext(ctx).
		Model(&models.PropostaMEI{}).
		Where("id IN ?", propostaIDs).
		Updates(updates)

	if result.Error != nil {
		return 0, fmt.Errorf("erro ao atualizar status das propostas: %w", result.Error)
	}

	return int(result.RowsAffected), nil
}
