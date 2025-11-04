package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	result := r.db.WithContext(ctx).Omit("CustomFields", "LocationClasses", "RemoteClass", "Inscricoes").Create(curso)
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
		Preload("CustomFields").
		Preload("LocationClasses.Schedules").
		Preload("RemoteClass").
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
	// Usar transação para garantir atomicidade
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Salvar as relações many-to-many
		if err := r.atualizarCategoriasWithTx(ctx, tx, curso); err != nil {
			return fmt.Errorf("erro ao atualizar categorias: %w", err)
		}

		if err := r.atualizarAcessibilidadesWithTx(ctx, tx, curso); err != nil {
			return fmt.Errorf("erro ao atualizar acessibilidades: %w", err)
		}
		
		// Atualizar custom fields
		if err := r.updateCustomFieldsWithTx(ctx, tx, curso); err != nil {
			return fmt.Errorf("erro ao atualizar custom fields: %w", err)
		}
		
		// Atualizar remote class
		if err := r.updateRemoteClassWithTx(ctx, tx, curso); err != nil {
			return fmt.Errorf("erro ao atualizar remote class: %w", err)
		}
		
		// Atualizar location classes
		if err := r.updateLocationClassesWithTx(ctx, tx, curso); err != nil {
			return fmt.Errorf("erro ao atualizar location classes: %w", err)
		}
		
		// Atualizar o curso - usar Select para incluir valores zero (como false para booleanos)
		// Omitir created_at para preservar a data original de criação
		result := tx.Model(curso).
			Where("id = ?", curso.ID).
			Omit("Categorias", "Acessibilidades", "Orgao", "Instituicao", "CustomFields", "RemoteClass", "LocationClasses", "created_at").
			Select("*").
			Updates(curso)
		
		if result.Error != nil {
			return fmt.Errorf("erro ao atualizar curso: %w", result.Error)
		}
		
		return nil
	})
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
		Preload("CustomFields").
		Preload("LocationClasses.Schedules").
		Preload("RemoteClass").
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
	return r.atualizarCategoriasWithTx(ctx, r.db, curso)
}

func (r *CursoRepository) atualizarCategoriasWithTx(ctx context.Context, tx *gorm.DB, curso *models.Curso) error {
	// Obter o curso atual com suas categorias
	var cursoAtual models.Curso
	if err := tx.WithContext(ctx).Preload("Categorias").First(&cursoAtual, curso.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // Curso não existe ainda, nada a fazer
		}
		return err
	}
	
	// Limpar associações existentes
	if err := tx.WithContext(ctx).Model(&cursoAtual).Association("Categorias").Clear(); err != nil {
		return err
	}
	
	// Adicionar novas associações
	if len(curso.Categorias) > 0 {
		if err := tx.WithContext(ctx).Model(curso).Association("Categorias").Replace(curso.Categorias); err != nil {
			return err
		}
	}
	
	return nil
}

func (r *CursoRepository) atualizarAcessibilidades(ctx context.Context, curso *models.Curso) error {
	return r.atualizarAcessibilidadesWithTx(ctx, r.db, curso)
}

func (r *CursoRepository) atualizarAcessibilidadesWithTx(ctx context.Context, tx *gorm.DB, curso *models.Curso) error {
	// Obter o curso atual com suas acessibilidades
	var cursoAtual models.Curso
	if err := tx.WithContext(ctx).Preload("Acessibilidades").First(&cursoAtual, curso.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // Curso não existe ainda, nada a fazer
		}
		return err
	}
	
	// Limpar associações existentes
	if err := tx.WithContext(ctx).Model(&cursoAtual).Association("Acessibilidades").Clear(); err != nil {
		return err
	}
	
	// Adicionar novas associações
	if len(curso.Acessibilidades) > 0 {
		if err := tx.WithContext(ctx).Model(curso).Association("Acessibilidades").Replace(curso.Acessibilidades); err != nil {
			return err
		}
	}
	
	return nil
}

// CreateCustomFields creates custom fields for a course
func (r *CursoRepository) CreateCustomFields(ctx context.Context, customFields []models.CustomField) error {
	if len(customFields) == 0 {
		return nil
	}
	
	result := r.db.WithContext(ctx).Create(&customFields)
	if result.Error != nil {
		return fmt.Errorf("erro ao criar custom fields: %w", result.Error)
	}
	
	return nil
}

// CreateRemoteClass creates a remote class for a course
func (r *CursoRepository) CreateRemoteClass(ctx context.Context, remoteClass *models.RemoteClass) error {
	if remoteClass == nil {
		return nil
	}
	
	result := r.db.WithContext(ctx).Create(remoteClass)
	if result.Error != nil {
		return fmt.Errorf("erro ao criar remote class: %w", result.Error)
	}
	
	return nil
}

// CreateLocationClasses creates location classes for a course
func (r *CursoRepository) CreateLocationClasses(ctx context.Context, locationClasses []models.LocationClass) error {
	if len(locationClasses) == 0 {
		return nil
	}
	
	result := r.db.WithContext(ctx).Create(&locationClasses)
	if result.Error != nil {
		return fmt.Errorf("erro ao criar location classes: %w", result.Error)
	}
	
	return nil
}

// updateCustomFields updates custom fields for a course
func (r *CursoRepository) updateCustomFields(ctx context.Context, curso *models.Curso) error {
	return r.updateCustomFieldsWithTx(ctx, r.db, curso)
}

func (r *CursoRepository) updateCustomFieldsWithTx(ctx context.Context, tx *gorm.DB, curso *models.Curso) error {
	// Delete all existing custom fields for this course
	if err := tx.WithContext(ctx).Where("curso_id = ?", curso.ID).Delete(&models.CustomField{}).Error; err != nil {
		return fmt.Errorf("erro ao deletar custom fields existentes: %w", err)
	}
	
	// Create new custom fields
	if len(curso.CustomFields) > 0 {
		for i := range curso.CustomFields {
			// Reset ID to ensure new ones are created
			curso.CustomFields[i].ID = uuid.UUID{}
			curso.CustomFields[i].CursoID = curso.ID
		}
		
		if err := tx.WithContext(ctx).Create(&curso.CustomFields).Error; err != nil {
			return fmt.Errorf("erro ao criar custom fields: %w", err)
		}
	}
	
	return nil
}

// updateRemoteClass updates remote class for a course
func (r *CursoRepository) updateRemoteClass(ctx context.Context, curso *models.Curso) error {
	return r.updateRemoteClassWithTx(ctx, r.db, curso)
}

func (r *CursoRepository) updateRemoteClassWithTx(ctx context.Context, tx *gorm.DB, curso *models.Curso) error {
	if curso.RemoteClass != nil {
		// Check if a remote class already exists
		var existingRemote models.RemoteClass
		err := tx.WithContext(ctx).Where("curso_id = ?", curso.ID).First(&existingRemote).Error
		
		curso.RemoteClass.CursoID = curso.ID
		
		if err == nil {
			// Update existing remote class
			curso.RemoteClass.ID = existingRemote.ID
			tx.WithContext(ctx).Model(&existingRemote).Updates(curso.RemoteClass)
		} else {
			// Create new remote class
			tx.WithContext(ctx).Create(curso.RemoteClass)
		}
	} else {
		// If no remote class provided, delete existing one if any
		tx.WithContext(ctx).Where("curso_id = ?", curso.ID).Delete(&models.RemoteClass{})
	}
	
	return nil
}

// updateLocationClasses updates location classes for a course
func (r *CursoRepository) updateLocationClasses(ctx context.Context, curso *models.Curso) error {
	return r.updateLocationClassesWithTx(ctx, r.db, curso)
}

func (r *CursoRepository) updateLocationClassesWithTx(ctx context.Context, tx *gorm.DB, curso *models.Curso) error {
	// Get existing location classes
	var existingLocations []models.LocationClass
	tx.WithContext(ctx).Where("curso_id = ?", curso.ID).Find(&existingLocations)

	// Build map of IDs to keep
	idsToKeep := make(map[string]bool)
	for _, location := range curso.LocationClasses {
		if location.ID.String() != "00000000-0000-0000-0000-000000000000" {
			idsToKeep[location.ID.String()] = true
		}
	}

	// Delete locations that are not in the update list (CASCADE will delete schedules)
	for _, existing := range existingLocations {
		if !idsToKeep[existing.ID.String()] {
			tx.WithContext(ctx).Delete(&existing)
		}
	}

	// Update or create locations and their schedules
	for i := range curso.LocationClasses {
		curso.LocationClasses[i].CursoID = curso.ID

		if curso.LocationClasses[i].ID.String() != "00000000-0000-0000-0000-000000000000" {
			// Update existing location (only address and neighborhood)
			tx.WithContext(ctx).Model(&curso.LocationClasses[i]).
				Select("address", "neighborhood", "updated_at").
				Updates(&curso.LocationClasses[i])

			// Delete all existing schedules for this location
			tx.WithContext(ctx).Where("location_id = ?", curso.LocationClasses[i].ID).Delete(&models.CourseSchedule{})
		} else {
			// Create new location
			tx.WithContext(ctx).Omit("Schedules").Create(&curso.LocationClasses[i])
		}

		// Create schedules for this location
		if len(curso.LocationClasses[i].Schedules) > 0 {
			for j := range curso.LocationClasses[i].Schedules {
				curso.LocationClasses[i].Schedules[j].LocationID = curso.LocationClasses[i].ID
				// Reset ID to ensure new schedule is created
				curso.LocationClasses[i].Schedules[j].ID = uuid.UUID{}
			}

			if err := tx.WithContext(ctx).Create(&curso.LocationClasses[i].Schedules).Error; err != nil {
				return fmt.Errorf("erro ao criar schedules: %w", err)
			}
		}
	}

	return nil
}

// CountEnrollmentsByScheduleID counts approved and concluded enrollments for a schedule
func (r *CursoRepository) CountEnrollmentsByScheduleID(ctx context.Context, scheduleID uuid.UUID) (int64, error) {
	var count int64

	result := r.db.WithContext(ctx).
		Model(&models.Inscricao{}).
		Where("schedule_id = ?", scheduleID).
		Where("status IN ?", []models.StatusInscricao{
			models.StatusInscricaoApproved,
			models.StatusInscricaoConcluded,
		}).
		Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("erro ao contar inscrições por schedule: %w", result.Error)
	}

	return count, nil
}

// CountEnrollmentsByScheduleIDs counts approved and concluded enrollments for multiple schedules
// Returns a map of schedule_id -> enrollment count
func (r *CursoRepository) CountEnrollmentsByScheduleIDs(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(scheduleIDs) == 0 {
		return make(map[uuid.UUID]int64), nil
	}

	type Result struct {
		ScheduleID uuid.UUID
		Count      int64
	}

	var results []Result

	err := r.db.WithContext(ctx).
		Model(&models.Inscricao{}).
		Select("schedule_id, COUNT(*) as count").
		Where("schedule_id IN ?", scheduleIDs).
		Where("status IN ?", []models.StatusInscricao{
			models.StatusInscricaoApproved,
			models.StatusInscricaoConcluded,
		}).
		Group("schedule_id").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("erro ao contar inscrições por schedules: %w", err)
	}

	// Build map
	countMap := make(map[uuid.UUID]int64)
	for _, result := range results {
		countMap[result.ScheduleID] = result.Count
	}

	return countMap, nil
}