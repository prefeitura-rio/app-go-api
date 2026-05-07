package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type CategoriaRepository struct {
	db *gorm.DB
}

func NewCategoriaRepository(db *gorm.DB) *CategoriaRepository {
	return &CategoriaRepository{
		db: db,
	}
}

func (r *CategoriaRepository) Create(ctx context.Context, categoria *models.Categoria) (int, error) {
	result := r.db.WithContext(ctx).Create(categoria)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar categoria: %w", result.Error)
	}

	return categoria.ID, nil
}

func (r *CategoriaRepository) GetByID(ctx context.Context, id int) (*models.Categoria, error) {
	var categoria models.Categoria

	result := r.db.WithContext(ctx).First(&categoria, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar categoria por ID: %w", result.Error)
	}

	return &categoria, nil
}

func (r *CategoriaRepository) Update(ctx context.Context, categoria *models.Categoria) error {
	result := r.db.WithContext(ctx).Model(categoria).
		Where("id = ?", categoria.ID).
		Updates(categoria)

	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar categoria: %w", result.Error)
	}

	return nil
}

func (r *CategoriaRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.Categoria{}, id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir categoria: %w", result.Error)
	}
	return nil
}

func (r *CategoriaRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Categoria, int, error) {
	var categorias []*models.Categoria
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Categoria{})

	if _, ok := filter["days_tolerance"]; ok {
		delete(filter, "days_tolerance")

		isVisible := true
		if iv, ok := filter["is_visible"].(bool); ok {
			isVisible = iv
			delete(filter, "is_visible")
		}

		if isVisible {
			db = db.Where(`id IN (
				SELECT DISTINCT cc.categoria_id
				FROM cursos_categorias cc
				INNER JOIN cursos c ON cc.curso_id = c.id
				WHERE c.status IN ('published', 'opened', 'ABERTO')
				AND (c.is_visible = true OR c.is_visible IS NULL)
			)`)
		} else {
			db = db.Where(`id IN (
				SELECT DISTINCT cc.categoria_id
				FROM cursos_categorias cc
				INNER JOIN cursos c ON cc.curso_id = c.id
				WHERE c.status IN ('published', 'opened', 'ABERTO')
			)`)
		}
	}

	for key, value := range filter {
		db = db.Where(key+" = ?", value)
	}

	db.Count(&total)

	result := db.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&categorias)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar categorias: %w", result.Error)
	}

	return categorias, int(total), nil
}
