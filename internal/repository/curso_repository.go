package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

type CursoRepository struct {
	db *gorm.DB
}

func NewCursoRepository(db *gorm.DB) *CursoRepository {
	return &CursoRepository{
		db: db,
	}
}

func (r *CursoRepository) Create(ctx context.Context, curso *models.Curso) (int, error) {
	result := r.db.WithContext(ctx).Create(curso)
	if result.Error != nil {
		return 0, fmt.Errorf("erro ao criar curso: %w", result.Error)
	}
	
	return curso.ID, nil
}

func (r *CursoRepository) GetByID(ctx context.Context, id int) (*models.Curso, error) {
	var curso models.Curso
	
	result := r.db.WithContext(ctx).
		Preload("Categorias").
		Preload("Acessibilidades").
		Preload("Orgao").
		Preload("Instituicao").
		First(&curso, id)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar curso por ID: %w", result.Error)
	}
	
	return &curso, nil
}

func (r *CursoRepository) Update(ctx context.Context, curso *models.Curso) error {
	// Salvar as relações many-to-many
	if err := r.atualizarCategorias(ctx, curso); err != nil {
		return fmt.Errorf("erro ao atualizar categorias: %w", err)
	}

	if err := r.atualizarAcessibilidades(ctx, curso); err != nil {
		return fmt.Errorf("erro ao atualizar acessibilidades: %w", err)
	}
	
	// Atualizar o curso
	result := r.db.WithContext(ctx).Model(curso).
		Where("id = ?", curso.ID).
		Omit("Categorias", "Acessibilidades", "Orgao", "Instituicao"). // Ignorar relações
		Updates(curso)
	
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar curso: %w", result.Error)
	}
	
	return nil
}

func (r *CursoRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&models.Curso{}, id)
	if result.Error != nil {
		return fmt.Errorf("erro ao excluir curso: %w", result.Error)
	}
	return nil
}

func (r *CursoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
	var cursos []*models.Curso
	var total int64
	
	// Contar total de registros
	db := r.db.WithContext(ctx).Model(&models.Curso{})
	db = r.applyFilters(db, filter)
	db.Count(&total)
	
	// Buscar registros com paginação
	db = r.db.WithContext(ctx).Model(&models.Curso{})
	db = r.applyFilters(db, filter)
	result := db.
		Preload("Categorias").
		Preload("Acessibilidades").
		Preload("Orgao").
		Preload("Instituicao").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&cursos)
		
	if result.Error != nil {
		return nil, 0, fmt.Errorf("erro ao listar cursos: %w", result.Error)
	}
	
	return cursos, int(total), nil
}

func (r *CursoRepository) applyFilters(db *gorm.DB, filter map[string]interface{}) *gorm.DB {
	for key, value := range filter {
		switch key {
		case "status NOT":
			db = db.Where("status != ?", value)
		case "title ILIKE":
			db = db.Where("titulo ILIKE ?", value)
		default:
			db = db.Where(key+" = ?", value)
		}
	}
	return db
}

// Métodos auxiliares para manipulação dos relacionamentos

func (r *CursoRepository) atualizarCategorias(ctx context.Context, curso *models.Curso) error {
	// Obter o curso atual com suas categorias
	var cursoAtual models.Curso
	if err := r.db.WithContext(ctx).Preload("Categorias").First(&cursoAtual, curso.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // Curso não existe ainda, nada a fazer
		}
		return err
	}
	
	// Limpar associações existentes
	if err := r.db.WithContext(ctx).Model(&cursoAtual).Association("Categorias").Clear(); err != nil {
		return err
	}
	
	// Adicionar novas associações
	if len(curso.Categorias) > 0 {
		if err := r.db.WithContext(ctx).Model(curso).Association("Categorias").Replace(curso.Categorias); err != nil {
			return err
		}
	}
	
	return nil
}

func (r *CursoRepository) atualizarAcessibilidades(ctx context.Context, curso *models.Curso) error {
	// Obter o curso atual com suas acessibilidades
	var cursoAtual models.Curso
	if err := r.db.WithContext(ctx).Preload("Acessibilidades").First(&cursoAtual, curso.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // Curso não existe ainda, nada a fazer
		}
		return err
	}
	
	// Limpar associações existentes
	if err := r.db.WithContext(ctx).Model(&cursoAtual).Association("Acessibilidades").Clear(); err != nil {
		return err
	}
	
	// Adicionar novas associações
	if len(curso.Acessibilidades) > 0 {
		if err := r.db.WithContext(ctx).Model(curso).Association("Acessibilidades").Replace(curso.Acessibilidades); err != nil {
			return err
		}
	}
	
	return nil
} 