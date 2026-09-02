package empregabilidade

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

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

// GetHabilidadeByID traz a habilidade carregando suas áreas de atuação vinculadas
func (r *HabilidadeRepository) GetHabilidadeByID(ctx context.Context, id int64) (*empregabilidade.Habilidade, error) {
	var entity empregabilidade.Habilidade

	// Preload("Areas") faz o JOIN automático com a tabela area_atuacao via tabela de junção
	result := r.db.WithContext(ctx).
		Preload("Areas").
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
	return nil
}

// DeleteHabilidade exclui uma habilidade
func (r *HabilidadeRepository) DeleteHabilidade(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Habilidade{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir habilidade: %w", result.Error)
	}
	return nil
}

// AttachAreaAtuacao vincula uma Área de Atuação a uma Habilidade
func (r *HabilidadeRepository) AttachAreaAtuacao(ctx context.Context, habilidadeID, areaID int64) error {
	habilidade := &empregabilidade.Habilidade{ID: habilidadeID}
	area := empregabilidade.AreaAtuacao{ID: areaID}

	// db.Model.Association gerencia a inclusão na tabela pivô (area_atuacao_habilidade)
	err := r.db.WithContext(ctx).
		Omit("Areas.*").
		Model(habilidade).
		Association("Areas").
		Append(&area)

	if err != nil {
		return fmt.Errorf("erro ao vincular área de atuação %d à habilidade %d: %w", areaID, habilidadeID, err)
	}

	return nil
}

// DetachAreaAtuacao remove o vínculo entre uma Área de Atuação e uma Habilidade
func (r *HabilidadeRepository) DetachAreaAtuacao(ctx context.Context, habilidadeID, areaID int64) error {
	habilidade := &empregabilidade.Habilidade{ID: habilidadeID}
	area := empregabilidade.AreaAtuacao{ID: areaID}

	// Remove o registro específico da tabela pivô sem deletar as entidades principais
	err := r.db.WithContext(ctx).
		Model(habilidade).
		Association("Areas").
		Delete(&area)

	if err != nil {
		return fmt.Errorf("erro ao desvincular área de atuação %d da habilidade %d: %w", areaID, habilidadeID, err)
	}

	return nil
}

// ReplaceAreasAtuacao substitui todas as áreas de atuação de uma habilidade por uma nova lista de IDs (útil para atualizações em lote)
func (r *HabilidadeRepository) ReplaceAreasAtuacao(ctx context.Context, habilidadeID int64, areaIDs []int64) error {
	habilidade := &empregabilidade.Habilidade{ID: habilidadeID}

	var areas []empregabilidade.AreaAtuacao
	if len(areaIDs) > 0 {
		if err := r.db.WithContext(ctx).Where("id IN ?", areaIDs).Find(&areas).Error; err != nil {
			return fmt.Errorf("erro ao buscar áreas de atuação para substituição: %w", err)
		}
	}

	// Omit("Areas.*") instrui o GORM a NÃO tentar inserir/atualizar a tabela area_atuacao,
	// manipulando estritamente os registros da tabela pivô
	err := r.db.WithContext(ctx).
		Omit("Areas.*").
		Model(habilidade).
		Association("Areas").
		Replace(&areas)

	if err != nil {
		return fmt.Errorf("erro ao substituir áreas de atuação da habilidade %d: %w", habilidadeID, err)
	}

	return nil
}

// ListHabilidades faz a paginação, busca por nome/área e carrega as Áreas de Atuação
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

	// Carrega as habilidades com Preload na relação Areas
	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.Habilidade{}))
	result := findDB.
		Preload("Areas").
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

// GetAreaAtuacaoByID traz a área de atuação carregando suas habilidades vinculadas
func (r *HabilidadeRepository) GetAreaAtuacaoByID(ctx context.Context, id int64) (*empregabilidade.AreaAtuacao, error) {
	var entity empregabilidade.AreaAtuacao

	// Preload("Habilidades") faz o JOIN automático com a tabela emp_habilidades via tabela de junção
	result := r.db.WithContext(ctx).
		Preload("Habilidades").
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
	return nil
}

// DeleteAreaAtuacao exclui uma área de atuação
func (r *HabilidadeRepository) DeleteAreaAtuacao(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.AreaAtuacao{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir uma área de atuação: %w", result.Error)
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
			db = db.Where("lower(immutable_unaccent(area_atuacao.nome)) LIKE lower(immutable_unaccent(?))", searchNome)
		}

		// Filtragem por ID da Habilidade
		if filter.HabilidadeID > 0 {
			db = db.Joins("JOIN area_atuacao_habilidade aah ON aah.id_area_atuacao = area_atuacao.id").
				Where("aah.id_habilidade = ?", filter.HabilidadeID)
		}

		return db
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.AreaAtuacao{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar áreas de atuação: %w", err)
	}

	// Carrega as áreas de atuação com Preload na relação Habilidades
	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.AreaAtuacao{}))
	result := findDB.
		Preload("Habilidades").
		Order("area_atuacao.nome ASC").
		Limit(limit).
		Offset(offset).
		Find(&entities)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar áreas de atuação: %w", result.Error)
	}

	return entities, total, nil
}

// Métodos para Vínculo com Currículo e Áreas

func (r *HabilidadeRepository) AddHabilidadeAoCurriculo(ctx context.Context, vinculo *empregabilidade.CurriculoHabilidade) error {
	result := r.db.WithContext(ctx).Create(vinculo)
	if result.Error != nil {
		return fmt.Errorf("erro ao vincular habilidade ao currículo: %w", result.Error)
	}
	return nil
}

func (r *HabilidadeRepository) ListHabilidadesPorCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoHabilidade, error) {
	var vinculos []*empregabilidade.CurriculoHabilidade
	result := r.db.WithContext(ctx).
		Preload("Habilidade.Areas"). // Traz a Habilidade E as Áreas de Atuação dela
		Where("cpf = ?", cpf).
		Find(&vinculos)

	if result.Error != nil {
		return nil, fmt.Errorf("erro ao buscar habilidades por CPF: %w", result.Error)
	}
	return vinculos, nil
}
