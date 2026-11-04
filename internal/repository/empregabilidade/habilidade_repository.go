package empregabilidade

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type HabilidadeRepository struct {
	db *gorm.DB
}

func NewHabilidadeRepository(db *gorm.DB) *HabilidadeRepository {
	return &HabilidadeRepository{db: db}
}

func (r *HabilidadeRepository) CreateHabilidade(ctx context.Context, entity *empregabilidade.Habilidade) (int64, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar habilidade: %w", result.Error)
	}
	return entity.ID, nil
}

// GetHabilidadeByID busca a habilidade pelo ID, sem Preload de áreas.
func (r *HabilidadeRepository) GetHabilidadeByID(ctx context.Context, id int64) (*empregabilidade.Habilidade, error) {
	var entity empregabilidade.Habilidade

	result := r.db.WithContext(ctx).
		First(&entity, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar habilidade por ID: %w", result.Error)
	}
	return &entity, nil
}

// UpdateHabilidade atualiza uma habilidade
func (r *HabilidadeRepository) UpdateHabilidade(ctx context.Context, entity *empregabilidade.Habilidade) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(map[string]interface{}{
		"nome":       entity.Nome,
		"updated_at": entity.UpdatedAt,
	})
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar habilidade: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteHabilidade exclui uma habilidade
func (r *HabilidadeRepository) DeleteHabilidade(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Habilidade{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir habilidade: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// AttachHabilidadeNaArea cria o par área × conhecimento.
func (r *HabilidadeRepository) AttachHabilidadeNaArea(ctx context.Context, areaID, habilidadeID int64) error {
	var areaCount int64
	if err := r.db.WithContext(ctx).Model(&empregabilidade.AreaAtuacao{}).Where("id = ?", areaID).Count(&areaCount).Error; err != nil {
		return fmt.Errorf("erro ao buscar área de atuação: %w", err)
	}
	if areaCount == 0 {
		return apperrors.NotFound("Área de atuação não encontrada")
	}

	var habilidadeCount int64
	if err := r.db.WithContext(ctx).Model(&empregabilidade.Habilidade{}).Where("id = ?", habilidadeID).Count(&habilidadeCount).Error; err != nil {
		return fmt.Errorf("erro ao buscar conhecimento: %w", err)
	}
	if habilidadeCount == 0 {
		return apperrors.NotFound("Conhecimento não encontrado")
	}

	var pairCount int64
	if err := r.db.WithContext(ctx).Model(&empregabilidade.AreaAtuacaoHabilidade{}).
		Where("id_area_atuacao = ? AND id_habilidade = ?", areaID, habilidadeID).
		Count(&pairCount).Error; err != nil {
		return fmt.Errorf("erro ao buscar vínculo: %w", err)
	}
	if pairCount > 0 {
		return apperrors.Conflict("Conhecimento já vinculado a esta área de atuação")
	}

	pair := &empregabilidade.AreaAtuacaoHabilidade{
		IDAreaAtuacao: areaID,
		IDHabilidade:  habilidadeID,
	}
	if err := r.db.WithContext(ctx).Create(pair).Error; err != nil {
		return fmt.Errorf("erro ao vincular conhecimento à área: %w", err)
	}
	return nil
}

// DetachHabilidadeDaArea remove só o par. A cascata tira o vínculo do currículo.
func (r *HabilidadeRepository) DetachHabilidadeDaArea(ctx context.Context, areaID, habilidadeID int64) error {
	result := r.db.WithContext(ctx).
		Where("id_area_atuacao = ? AND id_habilidade = ?", areaID, habilidadeID).
		Delete(&empregabilidade.AreaAtuacaoHabilidade{})
	if result.Error != nil {
		return fmt.Errorf("erro ao desvincular conhecimento da área: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFound("Conhecimento não vinculado a esta área de atuação")
	}
	return nil
}

// ListHabilidadesDaArea lista os conhecimentos de uma área, paginados.
func (r *HabilidadeRepository) ListHabilidadesDaArea(ctx context.Context, areaID int64, filter empregabilidade.HabilidadeFilter, limit, offset int) ([]*empregabilidade.CatalogoItem, int64, error) {
	var areaCount int64
	if err := r.db.WithContext(ctx).Model(&empregabilidade.AreaAtuacao{}).Where("id = ?", areaID).Count(&areaCount).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao buscar área de atuação: %w", err)
	}
	if areaCount == 0 {
		return nil, 0, apperrors.NotFound("Área de atuação não encontrada")
	}

	apply := func(db *gorm.DB) *gorm.DB {
		db = db.Table("emp_habilidades").
			Joins("JOIN area_atuacao_habilidade aah ON aah.id_habilidade = emp_habilidades.id").
			Where("aah.id_area_atuacao = ?", areaID)
		if filter.Search != "" {
			searchNome := fmt.Sprintf("%%%s%%", filter.Search)
			db = db.Where("lower(immutable_unaccent(emp_habilidades.nome)) LIKE lower(immutable_unaccent(?))", searchNome)
		}
		return db
	}

	var total int64
	if err := apply(r.db.WithContext(ctx)).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar conhecimentos da área: %w", err)
	}

	var items []*empregabilidade.CatalogoItem
	if err := apply(r.db.WithContext(ctx)).
		Select("emp_habilidades.id, emp_habilidades.nome").
		Order("emp_habilidades.nome ASC").
		Limit(limit).
		Offset(offset).
		Scan(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao listar conhecimentos da área: %w", err)
	}
	return items, total, nil
}

// ListHabilidades faz a paginação e a busca por nome, sem Preload de áreas.
func (r *HabilidadeRepository) ListHabilidades(ctx context.Context, filter empregabilidade.HabilidadeFilter, limit, offset int) ([]*empregabilidade.Habilidade, int64, error) {
	var entities []*empregabilidade.Habilidade
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.Search != "" {
			searchNome := fmt.Sprintf("%%%s%%", filter.Search)
			// Busca no nome da habilidade usando unaccent
			db = db.Where("lower(immutable_unaccent(emp_habilidades.nome)) LIKE lower(immutable_unaccent(?))", searchNome)
		}

		// Filtragem por ID da área de atuação usando o novo tipo int64 (> 0)
		if filter.AreaAtuacaoID > 0 {
			db = db.Joins("JOIN area_atuacao_habilidade aah ON aah.id_habilidade = emp_habilidades.id").
				Where("aah.id_area_atuacao = ?", filter.AreaAtuacaoID)
		}

		return db
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Habilidade{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar habilidades: %w", err)
	}

	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Habilidade{}))
	result := findDB.
		Order("emp_habilidades.nome ASC").
		Limit(limit).
		Offset(offset).
		Find(&entities)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar habilidades: %w", result.Error)
	}

	return entities, total, nil
}

// CreateAreaAtuacao cria uma Área de Atuação
func (r *HabilidadeRepository) CreateAreaAtuacao(ctx context.Context, entity *empregabilidade.AreaAtuacao) (int64, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar Área de Atuação: %w", result.Error)
	}
	return entity.ID, nil
}

// GetAreaAtuacaoByID busca a área de atuação pelo ID, sem Preload de conhecimentos.
func (r *HabilidadeRepository) GetAreaAtuacaoByID(ctx context.Context, id int64) (*empregabilidade.AreaAtuacao, error) {
	var entity empregabilidade.AreaAtuacao

	result := r.db.WithContext(ctx).
		First(&entity, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar área de atuação por ID: %w", result.Error)
	}
	return &entity, nil
}

// UpdateAreaAtuacao atualiza a área de atuação
func (r *HabilidadeRepository) UpdateAreaAtuacao(ctx context.Context, entity *empregabilidade.AreaAtuacao) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(map[string]interface{}{
		"nome":       entity.Nome,
		"updated_at": entity.UpdatedAt,
	})
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar área de atuação: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteAreaAtuacao exclui uma área de atuação
func (r *HabilidadeRepository) DeleteAreaAtuacao(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.AreaAtuacao{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir uma área de atuação: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListAreas faz a paginação e busca insensível a acentos para Áreas de Atuação
func (r *HabilidadeRepository) ListAreasAtuacao(ctx context.Context, filter empregabilidade.AreaAtuacaoFilter, limit, offset int) ([]*empregabilidade.AreaAtuacao, int64, error) {
	var entities []*empregabilidade.AreaAtuacao
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.Search != "" {
			searchNome := fmt.Sprintf("%%%s%%", filter.Search)
			// Ajuste o nome da tabela base se no seu banco não for emp_areas_atuacao
			db = db.Where("lower(immutable_unaccent(emp_areas_atuacao.nome)) LIKE lower(immutable_unaccent(?))", searchNome)
		}

		// Filtragem por ID da Habilidade
		if filter.HabilidadeID > 0 {
			db = db.Joins("JOIN area_atuacao_habilidade aah ON aah.id_area_atuacao = emp_areas_atuacao.id").
				Where("aah.id_habilidade = ?", filter.HabilidadeID)
		}

		return db
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.AreaAtuacao{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar áreas de atuação: %w", err)
	}

	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.AreaAtuacao{}))
	result := findDB.
		Order("emp_areas_atuacao.nome ASC").
		Limit(limit).
		Offset(offset).
		Find(&entities)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar áreas de atuação: %w", result.Error)
	}

	return entities, total, nil
}
