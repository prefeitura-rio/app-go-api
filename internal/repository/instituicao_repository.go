package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type InstituicaoRepository struct {
	db *gorm.DB
}

func NewInstituicaoRepository(db *gorm.DB) *InstituicaoRepository {
	return &InstituicaoRepository{
		db: db,
	}
}

func (r *InstituicaoRepository) Create(ctx context.Context, instituicao *models.InstituicaoEnsino) (int, error) {
	result := r.db.WithContext(ctx).Create(instituicao)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar instituição de ensino: %w", result.Error)
	}
	
	return instituicao.ID, nil
}

func (r *InstituicaoRepository) GetByID(ctx context.Context, id int) (*models.InstituicaoEnsino, error) {
	var instituicao models.InstituicaoEnsino
	
	result := r.db.WithContext(ctx).First(&instituicao, id)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar instituição de ensino por ID: %w", result.Error)
	}
	
	return &instituicao, nil
}

func (r *InstituicaoRepository) Update(ctx context.Context, instituicao *models.InstituicaoEnsino) error {
	result := r.db.WithContext(ctx).Model(instituicao).
		Where("id = ?", instituicao.ID).
		Updates(instituicao)
	
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar instituição de ensino: %w", result.Error)
	}
	
	return nil
}

func (r *InstituicaoRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.InstituicaoEnsino{}, id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir instituição de ensino: %w", result.Error)
	}
	return nil
}

func (r *InstituicaoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.InstituicaoEnsino, int, error) {
	var instituicoes []*models.InstituicaoEnsino
	var total int64
	
	// Contar total de registros
	db := r.db.WithContext(ctx).Model(&models.InstituicaoEnsino{})
	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}
	db.Count(&total)
	
	// Buscar registros com paginação
	result := db.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&instituicoes)
		
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar instituições de ensino: %w", result.Error)
	}
	
	return instituicoes, int(total), nil
}
