package repository

import (
	"context"
	"fmt"
	"time"

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
		Preload("Instituicao").
		Preload("CustomFields").
		Preload("LocationClasses.Schedules", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("RemoteClass.Schedules", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
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
			Omit("Categorias", "Acessibilidades", "Instituicao", "CustomFields", "RemoteClass", "LocationClasses", "created_at").
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

func (r *CursoRepository) UpdateStatus(ctx context.Context, id int, status models.StatusCurso) error {
	result := r.db.WithContext(ctx).Model(&models.Curso{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("erro ao atualizar status do curso: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("curso não encontrado")
	}
	return nil
}

func (r *CursoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
	var cursos []*models.Curso
	var total int64

	// Build base query
	baseQuery := r.db.WithContext(ctx).Model(&models.Curso{})
	baseQuery = r.applyFilters(baseQuery, filter)

	// Count total records (reuse filtered query)
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("erro ao contar cursos: %w", err)
	}

	// Fetch records with selective preloading
	// List view only needs basic info + schedules for vacancy calculation
	// Skip CustomFields, Instituicao details to reduce data transfer
	result := baseQuery.
		Preload("Categorias").
		Preload("Acessibilidades").
		Preload("LocationClasses.Schedules", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("RemoteClass.Schedules", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
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
		case "categoria_id":
			db = db.Where("cursos.id IN (SELECT curso_id FROM cursos_categorias WHERE categoria_id = ?)", value)
		case "acessibilidade_id":
			db = db.Where("cursos.id IN (SELECT curso_id FROM cursos_acessibilidades WHERE acessibilidade_id = ?)", value)
		case "neighborhood_zone":
			db = db.Where("cursos.id IN (SELECT curso_id FROM location_classes WHERE neighborhood_zone = ?)", value)
		default:
			db = db.Where(key+" = ?", value)
		}
	}
	return db
}

// Métodos auxiliares para manipulação dos relacionamentos

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

	for j := range remoteClass.Schedules {
		remoteClass.Schedules[j].DisplayOrder = j + 1
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

	for i := range locationClasses {
		for j := range locationClasses[i].Schedules {
			locationClasses[i].Schedules[j].DisplayOrder = j + 1
		}
	}

	result := r.db.WithContext(ctx).Create(&locationClasses)
	if result.Error != nil {
		return fmt.Errorf("erro ao criar location classes: %w", result.Error)
	}

	return nil
}

// updateCustomFieldsWithTx updates custom fields for a course
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

// updateRemoteClassWithTx updates remote class for a course
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

			// Build map of schedule IDs to keep
			scheduleIDsToKeep := make(map[string]bool)
			for _, schedule := range curso.RemoteClass.Schedules {
				if schedule.ID.String() != "00000000-0000-0000-0000-000000000000" {
					scheduleIDsToKeep[schedule.ID.String()] = true
				}
			}

			// Delete only schedules that are not in the update list
			var existingSchedules []models.RemoteSchedule
			tx.WithContext(ctx).Where("remote_class_id = ?", curso.RemoteClass.ID).Find(&existingSchedules)
			for _, existing := range existingSchedules {
				if !scheduleIDsToKeep[existing.ID.String()] {
					tx.WithContext(ctx).Delete(&existing)
				}
			}
		} else {
			// Create new remote class
			tx.WithContext(ctx).Omit("Schedules").Create(curso.RemoteClass)
		}

		// Update or create schedules for this remote class
		if len(curso.RemoteClass.Schedules) > 0 {
			var schedulesToCreate []models.RemoteSchedule

			for j := range curso.RemoteClass.Schedules {
				curso.RemoteClass.Schedules[j].RemoteClassID = curso.RemoteClass.ID
				curso.RemoteClass.Schedules[j].DisplayOrder = j + 1

				if curso.RemoteClass.Schedules[j].ID.String() != "00000000-0000-0000-0000-000000000000" {
					// Update existing schedule
					if err := tx.WithContext(ctx).Model(&curso.RemoteClass.Schedules[j]).
						Select("vacancies", "class_start_date", "class_end_date", "class_time", "class_days", "display_order", "accepting_enrollments", "updated_at").
						Updates(&curso.RemoteClass.Schedules[j]).Error; err != nil {
						return fmt.Errorf("erro ao atualizar remote schedule: %w", err)
					}
				} else {
					// Queue for batch creation
					schedulesToCreate = append(schedulesToCreate, curso.RemoteClass.Schedules[j])
				}
			}

			// Batch create new schedules
			if len(schedulesToCreate) > 0 {
				if err := tx.WithContext(ctx).Create(&schedulesToCreate).Error; err != nil {
					return fmt.Errorf("erro ao criar remote schedules: %w", err)
				}
			}
		}
	} else {
		// If no remote class provided, delete existing one if any (CASCADE will delete schedules)
		tx.WithContext(ctx).Where("curso_id = ?", curso.ID).Delete(&models.RemoteClass{})
	}

	return nil
}

// updateLocationClassesWithTx updates location classes for a course
func (r *CursoRepository) updateLocationClassesWithTx(ctx context.Context, tx *gorm.DB, curso *models.Curso) error {
	// Get existing location classes
	var existingLocations []models.LocationClass
	tx.WithContext(ctx).Where("curso_id = ?", curso.ID).Find(&existingLocations)

	// Build map of location IDs to keep
	locationIDsToKeep := make(map[string]bool)
	for _, location := range curso.LocationClasses {
		if location.ID.String() != "00000000-0000-0000-0000-000000000000" {
			locationIDsToKeep[location.ID.String()] = true
		}
	}

	// Delete locations that are not in the update list (CASCADE will delete schedules)
	for _, existing := range existingLocations {
		if !locationIDsToKeep[existing.ID.String()] {
			tx.WithContext(ctx).Delete(&existing)
		}
	}

	// Update or create locations and their schedules
	for i := range curso.LocationClasses {
		curso.LocationClasses[i].CursoID = curso.ID
		locationID := curso.LocationClasses[i].ID

		if locationID.String() != "00000000-0000-0000-0000-000000000000" {
			// Update existing location
			tx.WithContext(ctx).Model(&curso.LocationClasses[i]).
				Select("address", "neighborhood", "neighborhood_zone", "updated_at").
				Updates(&curso.LocationClasses[i])

			// Build map of schedule IDs to keep for this location
			scheduleIDsToKeep := make(map[string]bool)
			for _, schedule := range curso.LocationClasses[i].Schedules {
				if schedule.ID.String() != "00000000-0000-0000-0000-000000000000" {
					scheduleIDsToKeep[schedule.ID.String()] = true
				}
			}

			// Delete only schedules that are not in the update list
			var existingSchedules []models.CourseSchedule
			tx.WithContext(ctx).Where("location_id = ?", locationID).Find(&existingSchedules)
			for _, existing := range existingSchedules {
				if !scheduleIDsToKeep[existing.ID.String()] {
					tx.WithContext(ctx).Delete(&existing)
				}
			}
		} else {
			// Create new location
			tx.WithContext(ctx).Omit("Schedules").Create(&curso.LocationClasses[i])
		}

		// Update or create schedules for this location
		if len(curso.LocationClasses[i].Schedules) > 0 {
			var schedulesToCreate []models.CourseSchedule

			for j := range curso.LocationClasses[i].Schedules {
				curso.LocationClasses[i].Schedules[j].LocationID = curso.LocationClasses[i].ID
				curso.LocationClasses[i].Schedules[j].DisplayOrder = j + 1

				if curso.LocationClasses[i].Schedules[j].ID.String() != "00000000-0000-0000-0000-000000000000" {
					// Update existing schedule
					if err := tx.WithContext(ctx).Model(&curso.LocationClasses[i].Schedules[j]).
						Select("vacancies", "class_start_date", "class_end_date", "class_time", "class_days", "display_order", "accepting_enrollments", "updated_at").
						Updates(&curso.LocationClasses[i].Schedules[j]).Error; err != nil {
						return fmt.Errorf("erro ao atualizar course schedule: %w", err)
					}
				} else {
					// Queue for batch creation
					schedulesToCreate = append(schedulesToCreate, curso.LocationClasses[i].Schedules[j])
				}
			}

			// Batch create new schedules
			if len(schedulesToCreate) > 0 {
				if err := tx.WithContext(ctx).Create(&schedulesToCreate).Error; err != nil {
					return fmt.Errorf("erro ao criar schedules: %w", err)
				}
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

// ValidateForEnrollment checks if a course exists and can accept enrollments
// This is a lightweight query that only fetches the fields needed for validation,
// avoiding the overhead of loading all relationships
func (r *CursoRepository) ValidateForEnrollment(ctx context.Context, cursoID int) (status string, enrollmentStart, enrollmentEnd *time.Time, autoApprove bool, err error) {
	var result struct {
		Status                 string     `gorm:"column:status"`
		EnrollmentStartDate    *time.Time `gorm:"column:enrollment_start_date"`
		EnrollmentEndDate      *time.Time `gorm:"column:enrollment_end_date"`
		AutoApproveEnrollments *bool      `gorm:"column:auto_approve_enrollments"`
	}

	err = r.db.WithContext(ctx).
		Model(&models.Curso{}).
		Select("status, enrollment_start_date, enrollment_end_date, auto_approve_enrollments").
		Where("id = ?", cursoID).
		First(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil, nil, false, fmt.Errorf("curso não encontrado")
		}
		return "", nil, nil, false, fmt.Errorf("erro ao validar curso: %w", err)
	}

	autoApproveValue := result.AutoApproveEnrollments != nil && *result.AutoApproveEnrollments

	return result.Status, result.EnrollmentStartDate, result.EnrollmentEndDate, autoApproveValue, nil
}

// GetCourseScheduleByID fetches a course schedule by its ID
func (r *CursoRepository) GetCourseScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error) {
	var schedule models.CourseSchedule

	result := r.db.WithContext(ctx).
		Preload("Location").
		First(&schedule, "id = ?", scheduleID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar course schedule: %w", result.Error)
	}

	return &schedule, nil
}

// GetRemoteScheduleByID fetches a remote schedule by its ID
func (r *CursoRepository) GetRemoteScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.RemoteSchedule, error) {
	var schedule models.RemoteSchedule

	result := r.db.WithContext(ctx).First(&schedule, "id = ?", scheduleID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar remote schedule: %w", result.Error)
	}

	return &schedule, nil
}
