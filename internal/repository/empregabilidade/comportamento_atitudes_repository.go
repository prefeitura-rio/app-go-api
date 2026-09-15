package empregabilidade

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

type ComportamentoAtitudesRepository struct {
	db *gorm.DB
}

func NewComportamentoAtitudesRepository(db *gorm.DB) *ComportamentoAtitudesRepository {
	return &ComportamentoAtitudesRepository{db: db}
}

// CreateComportamentoAtitudes insere um novo comportamento no catálogo mestre
func (r *ComportamentoAtitudesRepository) CreateComportamentoAtitudes(ctx context.Context, entity *empregabilidade.ComportamentoAtitudes) (int64, error) {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar Comportamento e Atitudes: %w", result.Error)
	}
	return entity.ID, nil
}

// GetComportamentoAtitudesByID busca um comportamento mestre por ID
func (r *ComportamentoAtitudesRepository) GetComportamentoAtitudesByID(ctx context.Context, id int64) (*empregabilidade.ComportamentoAtitudes, error) {
	var entity empregabilidade.ComportamentoAtitudes

	result := r.db.WithContext(ctx).
		First(&entity, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar comportamento e atitude por ID: %w", result.Error)
	}
	return &entity, nil
}

// UpdateComportamentoAtitudes atualiza um comportamento no catálogo mestre
func (r *ComportamentoAtitudesRepository) UpdateComportamentoAtitudes(ctx context.Context, entity *empregabilidade.ComportamentoAtitudes) error {
	result := r.db.WithContext(ctx).Model(entity).Where("id = ?", entity.ID).Updates(map[string]interface{}{
		"nome":       entity.Nome,
		"updated_at": entity.UpdatedAt,
	})
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar comportamento e atitude: %w", result.Error)
	}
	return nil
}

// DeleteComportamentoAtitudes exclui um comportamento do catálogo mestre
func (r *ComportamentoAtitudesRepository) DeleteComportamentoAtitudes(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&empregabilidade.ComportamentoAtitudes{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir comportamento e atitude: %w", result.Error)
	}
	return nil
}

// ListComportamentoAtitudes faz a paginação e busca por nome com unaccent
func (r *ComportamentoAtitudesRepository) ListComportamentoAtitudes(ctx context.Context, filter empregabilidade.ComportamentoAtitudesFilter, limit, offset int) ([]*empregabilidade.ComportamentoAtitudes, int64, error) {
	var entities []*empregabilidade.ComportamentoAtitudes
	var total int64

	applyFilters := func(db *gorm.DB) *gorm.DB {
		if filter.Search != "" {
			searchNome := fmt.Sprintf("%%%s%%", filter.Search)
			db = db.Where("lower(immutable_unaccent(emp_comportamento_atitudes.nome)) LIKE lower(immutable_unaccent(?))", searchNome)
		}
		return db
	}

	countDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.ComportamentoAtitudes{}))
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar comportamentos e atitudes: %w", err)
	}

	findDB := applyFilters(r.db.WithContext(ctx).Model(&empregabilidade.ComportamentoAtitudes{}))
	result := findDB.
		Order("emp_comportamento_atitudes.nome ASC").
		Limit(limit).
		Offset(offset).
		Find(&entities)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar comportamentos e atitudes: %w", result.Error)
	}

	return entities, total, nil
}

// --- MÉTODOS DE VÍNCULO COM O CURRÍCULO (emp_curriculo_comportamento_atitudes) ---

// AddComportamentoAtitudesAoCurriculo vincula um comportamento ao currículo
func (r *ComportamentoAtitudesRepository) AddComportamentoAtitudesAoCurriculo(ctx context.Context, vinculo *empregabilidade.CurriculoComportamentoAtitudes) error {
	result := r.db.WithContext(ctx).Create(vinculo)
	if result.Error != nil {
		return fmt.Errorf("erro ao vincular comportamento ao currículo: %w", result.Error)
	}
	return nil
}

// DetachComportamentoAtitudesDoCurriculo remove o vínculo garantindo ownership pelo CPF
func (r *ComportamentoAtitudesRepository) DetachComportamentoAtitudesDoCurriculo(ctx context.Context, vinculo *empregabilidade.CurriculoComportamentoAtitudes) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND cpf = ?", vinculo.ID, vinculo.CPF).
		Delete(&empregabilidade.CurriculoComportamentoAtitudes{})

	if result.Error != nil {
		return fmt.Errorf("erro ao desvincular comportamento do currículo: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// ListComportamentoAtitudesPorCPF lista todos os comportamentos vinculados ao currículo do usuário carregando os dados mestre
func (r *ComportamentoAtitudesRepository) ListComportamentoAtitudesPorCPF(ctx context.Context, cpf string) ([]*empregabilidade.CurriculoComportamentoAtitudes, error) {
	var vinculos []*empregabilidade.CurriculoComportamentoAtitudes
	result := r.db.WithContext(ctx).
		Preload("ComportamentoAtitudes").
		Where("cpf = ?", cpf).
		Find(&vinculos)

	if result.Error != nil {
		return nil, fmt.Errorf("erro ao buscar comportamentos do currículo por CPF: %w", result.Error)
	}
	return vinculos, nil
}
