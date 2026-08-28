package empregabilidade

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type HabilidadeRepository struct {
	db *gorm.DB
}

func NewHabilidadeRepository(db *gorm.DB) *HabilidadeRepository {
	return &HabilidadeRepository{db: db}
}

func (r *HabilidadeRepository) CreateHabilidade(ctx context.Context, entity *empregabilidade.Habilidade) (uuid.UUID, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("erro ao criar habilidade: %w", result.Error)
	}
	return entity.ID, nil
}

// GetByID traz a habilidade carregando suas áreas de atuação vinculadas
func (r *HabilidadeRepository) GetHabilidadeByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Habilidade, error) {
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

func (r *HabilidadeRepository) DeleteHabilidade(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.Habilidade{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir habilidade: %w", result.Error)
	}
	return nil
}

// List faz a paginação, busca por nome/área e carrega as Áreas de Atuação
func (r *HabilidadeRepository) ListHabilidades(ctx context.Context, filter empregabilidade.HabilidadeFilter, limit, offset int) ([]*empregabilidade.Habilidade, int, error) {
	var entities []*empregabilidade.Habilidade
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.Search != "" {
			searchNome := fmt.Sprintf("%%%s%%", filter.Search)
			// Busca no nome da habilidade usando unaccent + trigram
			db = db.Where("lower(immutable_unaccent(emp_habilidades.nome)) LIKE lower(immutable_unaccent(?))", searchNome)
		}

		// Caso queira filtrar especificamente por ID da área de atuação
		if filter.AreaAtuacaoID != uuid.Nil {
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

	return entities, int(total), nil
}

// ListAreas faz a paginação e busca insensível a acentos para Áreas de Atuação
func (r *HabilidadeRepository) ListAreas(ctx context.Context, filter empregabilidade.AreaAtuacaoFilter, limit, offset int) ([]*empregabilidade.AreaAtuacao, int, error) {
	var entities []*empregabilidade.AreaAtuacao
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.Search != "" {
			searchNome := fmt.Sprintf("%%%s%%", filter.Search)
			// Aproveita o índice trigram + unaccent criado na migration para area_atuacao
			db = db.Where("lower(immutable_unaccent(area_atuacao.nome)) LIKE lower(immutable_unaccent(?))", searchNome)
		}
		return db
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.AreaAtuacao{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar áreas de atuação: %w", err)
	}

	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.AreaAtuacao{}))
	result := findDB.
		Order("area_atuacao.nome ASC").
		Limit(limit).
		Offset(offset).
		Find(&entities)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar áreas de atuação: %w", result.Error)
	}

	return entities, int(total), nil
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
