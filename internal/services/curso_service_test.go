package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// MockCursoRepository implements CursoRepositoryInterface for testing
type MockCursoRepository struct {
	CreateFunc                        func(ctx context.Context, curso *models.Curso) (int, error)
	GetByIDFunc                       func(ctx context.Context, id int) (*models.Curso, error)
	UpdateFunc                        func(ctx context.Context, curso *models.Curso) error
	DeleteFunc                        func(ctx context.Context, id int) error
	ListFunc                          func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error)
	CreateCustomFieldsFunc            func(ctx context.Context, fields []models.CustomField) error
	CreateRemoteClassFunc             func(ctx context.Context, remoteClass *models.RemoteClass) error
	CreateLocationClassesFunc         func(ctx context.Context, locations []models.LocationClass) error
	ValidateForEnrollmentFunc         func(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error)
	CountEnrollmentsByScheduleIDFunc  func(ctx context.Context, scheduleID uuid.UUID) (int64, error)
	CountEnrollmentsByScheduleIDsFunc func(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	GetCourseScheduleByIDFunc         func(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error)
	GetRemoteScheduleByIDFunc         func(ctx context.Context, scheduleID uuid.UUID) (*models.RemoteSchedule, error)
}

func (m *MockCursoRepository) Create(ctx context.Context, curso *models.Curso) (int, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, curso)
	}
	return 1, nil
}

func (m *MockCursoRepository) GetByID(ctx context.Context, id int) (*models.Curso, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockCursoRepository) Update(ctx context.Context, curso *models.Curso) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, curso)
	}
	return nil
}

func (m *MockCursoRepository) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockCursoRepository) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filter, limit, offset)
	}
	return []*models.Curso{}, 0, nil
}

func (m *MockCursoRepository) CreateCustomFields(ctx context.Context, fields []models.CustomField) error {
	if m.CreateCustomFieldsFunc != nil {
		return m.CreateCustomFieldsFunc(ctx, fields)
	}
	return nil
}

func (m *MockCursoRepository) CreateRemoteClass(ctx context.Context, remoteClass *models.RemoteClass) error {
	if m.CreateRemoteClassFunc != nil {
		return m.CreateRemoteClassFunc(ctx, remoteClass)
	}
	return nil
}

func (m *MockCursoRepository) CreateLocationClasses(ctx context.Context, locations []models.LocationClass) error {
	if m.CreateLocationClassesFunc != nil {
		return m.CreateLocationClassesFunc(ctx, locations)
	}
	return nil
}

func (m *MockCursoRepository) ValidateForEnrollment(ctx context.Context, cursoID int) (status string, enrollmentStart *time.Time, enrollmentEnd *time.Time, autoApprove bool, err error) {
	if m.ValidateForEnrollmentFunc != nil {
		return m.ValidateForEnrollmentFunc(ctx, cursoID)
	}
	return string(models.StatusCursoOpened), nil, nil, false, nil
}

func (m *MockCursoRepository) CountEnrollmentsByScheduleID(ctx context.Context, scheduleID uuid.UUID) (int64, error) {
	if m.CountEnrollmentsByScheduleIDFunc != nil {
		return m.CountEnrollmentsByScheduleIDFunc(ctx, scheduleID)
	}
	return 0, nil
}

func (m *MockCursoRepository) CountEnrollmentsByScheduleIDs(ctx context.Context, scheduleIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	if m.CountEnrollmentsByScheduleIDsFunc != nil {
		return m.CountEnrollmentsByScheduleIDsFunc(ctx, scheduleIDs)
	}
	return make(map[uuid.UUID]int64), nil
}

func (m *MockCursoRepository) GetCourseScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.CourseSchedule, error) {
	if m.GetCourseScheduleByIDFunc != nil {
		return m.GetCourseScheduleByIDFunc(ctx, scheduleID)
	}
	return nil, nil
}

func (m *MockCursoRepository) GetRemoteScheduleByID(ctx context.Context, scheduleID uuid.UUID) (*models.RemoteSchedule, error) {
	if m.GetRemoteScheduleByIDFunc != nil {
		return m.GetRemoteScheduleByIDFunc(ctx, scheduleID)
	}
	return nil, nil
}

// TestNewCursoService tests the NewCursoService constructor
func TestNewCursoService(t *testing.T) {
	// This test validates the constructor works with a mock repository
	// In production, NewCursoService is called with a concrete repository.NewCursoRepository
	// which requires a *gorm.DB. For unit testing, we verify the constructor pattern.
	t.Run("Constructor accepts repository", func(t *testing.T) {
		// We can't easily create a real gorm.DB in unit tests without a real database
		// So we test that the interface-based constructor works and the concrete one exists
		mockRepo := &MockCursoRepository{}

		// This uses the interface-based constructor (which the concrete one delegates to internally)
		svc := services.NewCursoServiceWithInterface(mockRepo)
		if svc == nil {
			t.Fatal("NewCursoServiceWithInterface returned nil")
		}
	})

	t.Run("Concrete constructor exists", func(t *testing.T) {
		// Verify the function signature exists at compile time
		// This ensures NewCursoService(repo *repository.CursoRepository) is available
		_ = services.NewCursoService
	})
}

// TestNewCursoServiceWithInterface tests the NewCursoServiceWithInterface constructor
func TestNewCursoServiceWithInterface(t *testing.T) {
	t.Run("With mock repository", func(t *testing.T) {
		mockRepo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(mockRepo)
		if svc == nil {
			t.Fatal("NewCursoServiceWithInterface returned nil")
		}
	})

	t.Run("Service is functional after creation", func(t *testing.T) {
		mockRepo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				return []*models.Curso{}, 0, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(mockRepo)
		ctx := context.Background()

		// Verify service can be used immediately
		_, _, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Errorf("Service should be functional after creation: %v", err)
		}
	})
}

// TestCursoService_List_WithFilters tests the List method with various filter scenarios
func TestCursoService_List_WithFilters(t *testing.T) {
	ctx := context.Background()

	t.Run("Filter by status", func(t *testing.T) {
		expectedFilter := map[string]interface{}{
			"status": "OPENED",
		}
		var capturedFilter map[string]interface{}

		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				capturedFilter = filter
				return []*models.Curso{
					{ID: 1, Titulo: "Opened Course", Status: models.StatusCursoOpened},
				}, 1, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		results, total, err := svc.List(ctx, expectedFilter, 1, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 curso, got %d", len(results))
		}
		if capturedFilter["status"] != "OPENED" {
			t.Errorf("Filter was not passed correctly: %v", capturedFilter)
		}
	})

	t.Run("Filter by modalidade", func(t *testing.T) {
		expectedFilter := map[string]interface{}{
			"modalidade": "PRESENCIAL",
		}
		var capturedFilter map[string]interface{}

		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				capturedFilter = filter
				return []*models.Curso{
					{ID: 1, Titulo: "Presential Course", Modalidade: models.ModalidadePresencial},
				}, 1, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, total, err := svc.List(ctx, expectedFilter, 1, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if capturedFilter["modalidade"] != "PRESENCIAL" {
			t.Errorf("Filter was not passed correctly: %v", capturedFilter)
		}
	})

	t.Run("Filter by organization", func(t *testing.T) {
		expectedFilter := map[string]interface{}{
			"organization": "Test Org",
		}
		var capturedFilter map[string]interface{}

		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				capturedFilter = filter
				return []*models.Curso{
					{ID: 1, Titulo: "Org Course", Organization: "Test Org"},
				}, 1, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, total, err := svc.List(ctx, expectedFilter, 1, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if capturedFilter["organization"] != "Test Org" {
			t.Errorf("Filter was not passed correctly: %v", capturedFilter)
		}
	})

	t.Run("Multiple filters combined", func(t *testing.T) {
		expectedFilter := map[string]interface{}{
			"status":       "OPENED",
			"modalidade":   "REMOTO",
			"organization": "Test Org",
		}
		var capturedFilter map[string]interface{}

		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				capturedFilter = filter
				return []*models.Curso{
					{
						ID:           1,
						Titulo:       "Filtered Course",
						Status:       models.StatusCursoOpened,
						Modalidade:   models.ModalidadeRemoto,
						Organization: "Test Org",
					},
				}, 1, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		cursos, total, err := svc.List(ctx, expectedFilter, 1, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if len(cursos) != 1 {
			t.Errorf("Expected 1 curso, got %d", len(cursos))
		}
		if capturedFilter["status"] != "OPENED" {
			t.Errorf("Status filter was not passed correctly")
		}
		if capturedFilter["modalidade"] != "REMOTO" {
			t.Errorf("Modalidade filter was not passed correctly")
		}
		if capturedFilter["organization"] != "Test Org" {
			t.Errorf("Organization filter was not passed correctly")
		}
	})

	t.Run("Empty filter returns all", func(t *testing.T) {
		var capturedFilter map[string]interface{}

		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				capturedFilter = filter
				return []*models.Curso{
					{ID: 1, Titulo: "Course 1"},
					{ID: 2, Titulo: "Course 2"},
					{ID: 3, Titulo: "Course 3"},
				}, 3, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		cursos, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}
		if len(cursos) != 3 {
			t.Errorf("Expected 3 cursos, got %d", len(cursos))
		}
		if capturedFilter != nil {
			t.Logf("Filter passed as nil is acceptable: %v", capturedFilter)
		}
	})

	t.Run("Nil filter is handled", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				// Nil filter should be acceptable
				return []*models.Curso{}, 0, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, _, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Errorf("List should handle nil filter: %v", err)
		}
	})
}

// TestCursoService_List_PaginationBoundaries tests pagination edge cases
func TestCursoService_List_PaginationBoundaries(t *testing.T) {
	ctx := context.Background()

	t.Run("First page", func(t *testing.T) {
		var capturedOffset int
		var capturedLimit int

		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				capturedLimit = limit
				capturedOffset = offset
				return []*models.Curso{
					{ID: 1, Titulo: "Course 1"},
					{ID: 2, Titulo: "Course 2"},
				}, 10, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		cursos, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 10 {
			t.Errorf("Expected total 10, got %d", total)
		}
		if capturedOffset != 0 {
			t.Errorf("First page should have offset 0, got %d", capturedOffset)
		}
		if capturedLimit != 10 {
			t.Errorf("Expected limit 10, got %d", capturedLimit)
		}
		if len(cursos) != 2 {
			t.Errorf("Expected 2 cursos, got %d", len(cursos))
		}
	})

	t.Run("Second page", func(t *testing.T) {
		var capturedOffset int
		var capturedLimit int

		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				capturedLimit = limit
				capturedOffset = offset
				return []*models.Curso{
					{ID: 11, Titulo: "Course 11"},
					{ID: 12, Titulo: "Course 12"},
				}, 20, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		cursos, total, err := svc.List(ctx, nil, 2, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 20 {
			t.Errorf("Expected total 20, got %d", total)
		}
		if capturedOffset != 10 {
			t.Errorf("Second page with pageSize 10 should have offset 10, got %d", capturedOffset)
		}
		if capturedLimit != 10 {
			t.Errorf("Expected limit 10, got %d", capturedLimit)
		}
		if len(cursos) != 2 {
			t.Errorf("Expected 2 cursos, got %d", len(cursos))
		}
	})

	t.Run("Large page size", func(t *testing.T) {
		var capturedLimit int

		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				capturedLimit = limit
				items := make([]*models.Curso, limit)
				for i := 0; i < limit; i++ {
					items[i] = &models.Curso{ID: i + 1, Titulo: "Course"}
				}
				return items, 500, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		cursos, total, err := svc.List(ctx, nil, 1, 100)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 500 {
			t.Errorf("Expected total 500, got %d", total)
		}
		if capturedLimit != 100 {
			t.Errorf("Expected limit 100, got %d", capturedLimit)
		}
		if len(cursos) != 100 {
			t.Errorf("Expected 100 cursos, got %d", len(cursos))
		}
	})

	t.Run("Small page size", func(t *testing.T) {
		var capturedLimit int

		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				capturedLimit = limit
				return []*models.Curso{
					{ID: 1, Titulo: "Course 1"},
				}, 10, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		cursos, _, err := svc.List(ctx, nil, 1, 1)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if capturedLimit != 1 {
			t.Errorf("Expected limit 1, got %d", capturedLimit)
		}
		if len(cursos) != 1 {
			t.Errorf("Expected 1 curso, got %d", len(cursos))
		}
	})

	t.Run("Last page partial results", func(t *testing.T) {
		var capturedOffset int

		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				capturedOffset = offset
				// Total 25 items, page 3 with pageSize 10 should return 5 items
				return []*models.Curso{
					{ID: 21, Titulo: "Course 21"},
					{ID: 22, Titulo: "Course 22"},
					{ID: 23, Titulo: "Course 23"},
					{ID: 24, Titulo: "Course 24"},
					{ID: 25, Titulo: "Course 25"},
				}, 25, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		cursos, total, err := svc.List(ctx, nil, 3, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 25 {
			t.Errorf("Expected total 25, got %d", total)
		}
		if capturedOffset != 20 {
			t.Errorf("Third page should have offset 20, got %d", capturedOffset)
		}
		if len(cursos) != 5 {
			t.Errorf("Expected 5 cursos on last page, got %d", len(cursos))
		}
	})

	t.Run("Empty page beyond results", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				// Total 10 items, page 5 should return empty
				return []*models.Curso{}, 10, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		cursos, total, err := svc.List(ctx, nil, 5, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 10 {
			t.Errorf("Expected total 10, got %d", total)
		}
		if len(cursos) != 0 {
			t.Errorf("Expected 0 cursos beyond last page, got %d", len(cursos))
		}
	})
}

// TestCursoService_Create tests the Create method
func TestCursoService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Create valid draft curso", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 100, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Test Course",
			Status:     models.StatusCursoDraft,
			Modalidade: "",
		}

		id, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id != 100 {
			t.Errorf("Expected ID 100, got %d", id)
		}
	})

	t.Run("Create valid published curso", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 200, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:       "Published Course",
			Status:       models.StatusCursoOpened,
			Modalidade:   models.ModalidadePresencial,
			NumeroVagas:  50,
			CargaHoraria: 40,
		}

		id, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id != 200 {
			t.Errorf("Expected ID 200, got %d", id)
		}
	})

	t.Run("Create with empty titulo fails validation", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for empty titulo")
		}
	})

	t.Run("Create with titulo too long fails validation", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		longTitulo := make([]byte, 20001)
		for i := range longTitulo {
			longTitulo[i] = 'a'
		}

		curso := &models.Curso{
			Titulo:     string(longTitulo),
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for titulo too long")
		}
	})

	t.Run("Create with invalid modalidade fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Test Course",
			Status:     models.StatusCursoDraft,
			Modalidade: models.Modalidade("INVALID"),
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for invalid modalidade")
		}
	})

	t.Run("Create with negative numero_vagas fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:      "Test Course",
			Status:      models.StatusCursoOpened,
			Modalidade:  models.ModalidadePresencial,
			NumeroVagas: -10,
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for negative numero_vagas")
		}
	})

	t.Run("Create with repository error", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 0, errors.New("database error")
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Test Course",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected database error")
		}
	})

	t.Run("Create with custom fields", func(t *testing.T) {
		var createdFields []models.CustomField
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 300, nil
			},
			CreateCustomFieldsFunc: func(ctx context.Context, fields []models.CustomField) error {
				createdFields = fields
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Course with Custom Fields",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
			CustomFields: []models.CustomField{
				{Title: "Field1", Required: true},
			},
		}

		id, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id != 300 {
			t.Errorf("Expected ID 300, got %d", id)
		}
		if len(createdFields) != 1 {
			t.Errorf("Expected 1 custom field created, got %d", len(createdFields))
		}
	})

	t.Run("Create with location classes", func(t *testing.T) {
		var createdLocations []models.LocationClass
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 400, nil
			},
			CreateLocationClassesFunc: func(ctx context.Context, locations []models.LocationClass) error {
				createdLocations = locations
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Course with Locations",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Test Address 123",
					Neighborhood: "Test Neighborhood",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      30,
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "09:00-12:00",
							ClassDays:      "Segunda a Sexta",
						},
					},
				},
			},
		}

		id, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id != 400 {
			t.Errorf("Expected ID 400, got %d", id)
		}
		if len(createdLocations) != 1 {
			t.Errorf("Expected 1 location created, got %d", len(createdLocations))
		}
	})

	t.Run("Create with multiple locations", func(t *testing.T) {
		var createdLocations []models.LocationClass
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 500, nil
			},
			CreateLocationClassesFunc: func(ctx context.Context, locations []models.LocationClass) error {
				createdLocations = locations
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Course with Multiple Locations",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Test Address 123",
					Neighborhood: "Test Neighborhood 1",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      30,
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "09:00-12:00",
							ClassDays:      "Segunda a Sexta",
						},
					},
				},
				{
					Address:      "Another Address 456",
					Neighborhood: "Test Neighborhood 2",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      25,
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "14:00-17:00",
							ClassDays:      "Terça e Quinta",
						},
					},
				},
			},
		}

		id, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id != 500 {
			t.Errorf("Expected ID 500, got %d", id)
		}
		if len(createdLocations) != 2 {
			t.Errorf("Expected 2 locations created, got %d", len(createdLocations))
		}
	})

	t.Run("Create assigns curso_id to relationships", func(t *testing.T) {
		var createdFields []models.CustomField
		createdID := 999

		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return createdID, nil
			},
			CreateCustomFieldsFunc: func(ctx context.Context, fields []models.CustomField) error {
				createdFields = fields
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Course with Relations",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
			CustomFields: []models.CustomField{
				{Title: "Field1", Required: true},
				{Title: "Field2", Required: false},
			},
		}

		id, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id != createdID {
			t.Errorf("Expected ID %d, got %d", createdID, id)
		}

		// Verify curso_id was assigned to custom fields
		for i, field := range createdFields {
			if field.CursoID != createdID {
				t.Errorf("Custom field %d should have CursoID %d, got %d", i, createdID, field.CursoID)
			}
		}
	})

	t.Run("Create with custom fields error", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 600, nil
			},
			CreateCustomFieldsFunc: func(ctx context.Context, fields []models.CustomField) error {
				return errors.New("custom fields creation failed")
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Course with Custom Fields Error",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
			CustomFields: []models.CustomField{
				{Title: "Field1", Required: true},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected custom fields error")
		}
		if !strings.Contains(err.Error(), "custom fields") {
			t.Errorf("Expected custom fields error message, got: %v", err)
		}
	})

	t.Run("Create with location classes error", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 700, nil
			},
			CreateLocationClassesFunc: func(ctx context.Context, locations []models.LocationClass) error {
				return errors.New("location classes creation failed")
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Course with Location Error",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Test Address 123",
					Neighborhood: "Test Neighborhood",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      30,
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "09:00-12:00",
							ClassDays:      "Segunda a Sexta",
						},
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected location classes error")
		}
		if !strings.Contains(err.Error(), "location classes") {
			t.Errorf("Expected location classes error message, got: %v", err)
		}
	})

	t.Run("Create draft with minimal fields", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 800, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo: "Minimal Draft",
			Status: models.StatusCursoDraft,
			// Modalidade is optional for draft
		}

		id, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create minimal draft should succeed: %v", err)
		}
		if id != 800 {
			t.Errorf("Expected ID 800, got %d", id)
		}
	})

	t.Run("Create normalizes titulo", func(t *testing.T) {
		var createdCurso *models.Curso
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				createdCurso = curso
				return 900, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "  Curso com espaços  ",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
		}

		_, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if createdCurso.Titulo != "Curso com espaços" {
			t.Errorf("Expected trimmed titulo, got '%s'", createdCurso.Titulo)
		}
	})
}

// TestCursoService_GetByID tests the GetByID method
func TestCursoService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("GetByID existing curso", func(t *testing.T) {
		expectedCurso := &models.Curso{
			ID:         1,
			Titulo:     "Test Course",
			Modalidade: models.ModalidadePresencial,
		}

		repo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				if id == 1 {
					return expectedCurso, nil
				}
				return nil, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso, err := svc.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if curso == nil {
			t.Fatal("Expected curso, got nil")
		}
		if curso.ID != 1 || curso.Titulo != "Test Course" {
			t.Errorf("Got unexpected curso: %+v", curso)
		}
	})

	t.Run("GetByID non-existing curso", func(t *testing.T) {
		repo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return nil, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso, err := svc.GetByID(ctx, 999)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if curso != nil {
			t.Error("Expected nil for non-existing curso")
		}
	})

	t.Run("GetByID with repository error", func(t *testing.T) {
		repo := &MockCursoRepository{
			GetByIDFunc: func(ctx context.Context, id int) (*models.Curso, error) {
				return nil, errors.New("database error")
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, err := svc.GetByID(ctx, 1)
		if err == nil {
			t.Error("Expected database error")
		}
	})
}

// TestCursoService_Update tests the Update method
func TestCursoService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("Update valid curso", func(t *testing.T) {
		repo := &MockCursoRepository{
			UpdateFunc: func(ctx context.Context, curso *models.Curso) error {
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			ID:         1,
			Titulo:     "Updated Course",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
		}

		err := svc.Update(ctx, curso)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	})

	t.Run("Update with empty titulo fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			ID:         1,
			Titulo:     "",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
		}

		err := svc.Update(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for empty titulo")
		}
	})

	t.Run("Update with repository error", func(t *testing.T) {
		repo := &MockCursoRepository{
			UpdateFunc: func(ctx context.Context, curso *models.Curso) error {
				return errors.New("database error")
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			ID:         1,
			Titulo:     "Test Course",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
		}

		err := svc.Update(ctx, curso)
		if err == nil {
			t.Error("Expected database error")
		}
	})

	t.Run("Update draft to published requires full validation", func(t *testing.T) {
		var updatedCurso *models.Curso
		repo := &MockCursoRepository{
			UpdateFunc: func(ctx context.Context, curso *models.Curso) error {
				updatedCurso = curso
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			ID:           1,
			Titulo:       "Course Becoming Published",
			Status:       models.StatusCursoOpened,
			Modalidade:   models.ModalidadePresencial,
			NumeroVagas:  30,
			CargaHoraria: 40,
		}

		err := svc.Update(ctx, curso)
		if err != nil {
			t.Fatalf("Update should succeed: %v", err)
		}
		if updatedCurso == nil {
			t.Fatal("Update was not called")
		}
		if updatedCurso.Status != models.StatusCursoOpened {
			t.Errorf("Status should be OPENED, got %s", updatedCurso.Status)
		}
	})

	t.Run("Update with invalid status transition", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			ID:         1,
			Titulo:     "Test Course",
			Status:     models.StatusCurso("INVALID_STATUS"),
			Modalidade: models.ModalidadePresencial,
		}

		err := svc.Update(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for invalid status")
		}
		if !strings.Contains(err.Error(), "status inválido") {
			t.Errorf("Expected 'status inválido' error, got: %v", err)
		}
	})

	t.Run("Update normalizes data", func(t *testing.T) {
		var updatedCurso *models.Curso
		repo := &MockCursoRepository{
			UpdateFunc: func(ctx context.Context, curso *models.Curso) error {
				updatedCurso = curso
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			ID:         1,
			Titulo:     "  Course With Spaces  ",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
		}

		err := svc.Update(ctx, curso)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		if updatedCurso.Titulo != "Course With Spaces" {
			t.Errorf("Titulo should be trimmed, got: '%s'", updatedCurso.Titulo)
		}
	})

	t.Run("Update with numero_vagas boundary", func(t *testing.T) {
		repo := &MockCursoRepository{
			UpdateFunc: func(ctx context.Context, curso *models.Curso) error {
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			ID:          1,
			Titulo:      "Boundary Test",
			Status:      models.StatusCursoOpened,
			Modalidade:  models.ModalidadePresencial,
			NumeroVagas: 0, // Edge case: zero vagas is valid
		}

		err := svc.Update(ctx, curso)
		if err != nil {
			t.Fatalf("Update with zero vagas should be valid: %v", err)
		}
	})
}

// TestCursoService_Delete tests the Delete method
func TestCursoService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Delete existing curso", func(t *testing.T) {
		repo := &MockCursoRepository{
			DeleteFunc: func(ctx context.Context, id int) error {
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		err := svc.Delete(ctx, 1)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})

	t.Run("Delete with repository error", func(t *testing.T) {
		repo := &MockCursoRepository{
			DeleteFunc: func(ctx context.Context, id int) error {
				return errors.New("database error")
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		err := svc.Delete(ctx, 1)
		if err == nil {
			t.Error("Expected database error")
		}
	})
}

// TestCursoService_List tests the List method
func TestCursoService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("List with results", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				return []*models.Curso{
					{ID: 1, Titulo: "Course 1"},
					{ID: 2, Titulo: "Course 2"},
				}, 2, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		cursos, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 2 {
			t.Errorf("Expected total 2, got %d", total)
		}
		if len(cursos) != 2 {
			t.Errorf("Expected 2 cursos, got %d", len(cursos))
		}
	})

	t.Run("List with pagination", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				if limit != 5 || offset != 10 {
					t.Errorf("Expected limit=5 offset=10, got limit=%d offset=%d", limit, offset)
				}
				return []*models.Curso{}, 0, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, _, err := svc.List(ctx, nil, 3, 5) // page 3, pageSize 5 = offset 10
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
	})

	t.Run("List with repository error", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				return nil, 0, errors.New("database error")
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, _, err := svc.List(ctx, nil, 1, 10)
		if err == nil {
			t.Error("Expected database error")
		}
	})
}

// TestCursoService_ValidationScenarios tests comprehensive validation scenarios
func TestCursoService_ValidationScenarios(t *testing.T) {
	ctx := context.Background()

	t.Run("Validate LIVRE_FORMACAO_ONLINE requires formacao_link", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Online Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadeLivreFormacaoOnline,
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for missing formacao_link")
		}
	})

	t.Run("Validate LIVRE_FORMACAO_ONLINE with valid URL", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 100, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:       "Online Course",
			Status:       models.StatusCursoOpened,
			Modalidade:   models.ModalidadeLivreFormacaoOnline,
			FormacaoLink: "https://example.com/course",
		}

		id, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id != 100 {
			t.Errorf("Expected ID 100, got %d", id)
		}
	})

	t.Run("Validate LIVRE_FORMACAO_ONLINE cannot have locations", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:       "Online Course",
			Status:       models.StatusCursoOpened,
			Modalidade:   models.ModalidadeLivreFormacaoOnline,
			FormacaoLink: "https://example.com/course",
			LocationClasses: []models.LocationClass{
				{Address: "Test Address"},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for locations in LIVRE_FORMACAO_ONLINE")
		}
	})

	t.Run("Validate LIVRE_FORMACAO_ONLINE cannot have remote class", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:       "Online Course",
			Status:       models.StatusCursoOpened,
			Modalidade:   models.ModalidadeLivreFormacaoOnline,
			FormacaoLink: "https://example.com/course",
			RemoteClass:  &models.RemoteClass{},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for remote_class in LIVRE_FORMACAO_ONLINE")
		}
	})

	t.Run("Validate location with short address fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Presential Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Short",
					Neighborhood: "Test Neighborhood",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      30,
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "09:00-12:00",
							ClassDays:      "Segunda a Sexta",
						},
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for short address")
		}
	})

	t.Run("Validate location with short neighborhood fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Presential Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Valid Address 123",
					Neighborhood: "AB",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      30,
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "09:00-12:00",
							ClassDays:      "Segunda a Sexta",
						},
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for short neighborhood")
		}
	})

	t.Run("Validate location without schedules fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Presential Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Valid Address 123",
					Neighborhood: "Test Neighborhood",
					Schedules:    []models.CourseSchedule{},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for location without schedules")
		}
	})

	t.Run("Validate schedule with invalid vacancies fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Presential Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Valid Address 123",
					Neighborhood: "Test Neighborhood",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      1001, // Too many
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "09:00-12:00",
							ClassDays:      "Segunda a Sexta",
						},
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for too many vacancies")
		}
	})

	t.Run("Validate schedule with end date before start date fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(72 * time.Hour)
		endDate := time.Now().Add(24 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Presential Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Valid Address 123",
					Neighborhood: "Test Neighborhood",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      30,
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "09:00-12:00",
							ClassDays:      "Segunda a Sexta",
						},
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for end date before start date")
		}
	})

	t.Run("Validate CourseManagementType OWN_ORG", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 100, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:               "Course Own Org",
			Status:               models.StatusCursoOpened,
			Modalidade:           models.ModalidadePresencial,
			CourseManagementType: models.CourseManagementOwnOrg,
			ExternalPartnerName:  "", // Must be empty
			ExternalPartnerURL:   "",
		}

		id, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id != 100 {
			t.Errorf("Expected ID 100, got %d", id)
		}
	})

	t.Run("Validate CourseManagementType OWN_ORG with partner fields fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:               "Course Own Org",
			Status:               models.StatusCursoOpened,
			Modalidade:           models.ModalidadePresencial,
			CourseManagementType: models.CourseManagementOwnOrg,
			ExternalPartnerName:  "Partner Name", // Should be empty
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for partner fields in OWN_ORG")
		}
	})

	t.Run("Validate CourseManagementType EXTERNAL_MANAGED_BY_ORG requires partner name", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:               "External Course",
			Status:               models.StatusCursoOpened,
			Modalidade:           models.ModalidadePresencial,
			CourseManagementType: models.CourseManagementExternalManagedByOrg,
			ExternalPartnerName:  "", // Required
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for missing partner name")
		}
	})

	t.Run("Validate CourseManagementType EXTERNAL_MANAGED_BY_PARTNER requires URL", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:               "External Course",
			Status:               models.StatusCursoOpened,
			Modalidade:           models.ModalidadePresencial,
			CourseManagementType: models.CourseManagementExternalManagedByPartner,
			ExternalPartnerName:  "Partner Name",
			ExternalPartnerURL:   "", // Required
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for missing partner URL")
		}
	})

	t.Run("Validate invalid FormatoAula fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:      "Test Course",
			Status:      models.StatusCursoDraft,
			Modalidade:  models.ModalidadePresencial,
			FormatoAula: models.FormatoAula("INVALID"),
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for invalid FormatoAula")
		}
	})

	t.Run("Validate invalid Turno fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Test Course",
			Status:     models.StatusCursoDraft,
			Modalidade: models.ModalidadePresencial,
			Turno:      models.Turno("INVALID"),
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for invalid Turno")
		}
	})

	// Remote class validation tests
	t.Run("RemoteClass with no schedules fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadeRemoto,
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for remote class without schedules")
		}
		if !strings.Contains(err.Error(), "pelo menos 1 turma") {
			t.Errorf("Expected 'pelo menos 1 turma' error, got: %v", err)
		}
	})

	t.Run("RemoteClass nil is valid", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 200, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:      "Remote Course",
			Status:      models.StatusCursoOpened,
			Modalidade:  models.ModalidadeRemoto,
			RemoteClass: nil, // nil is valid
		}

		_, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create should succeed with nil remote class: %v", err)
		}
	})

	t.Run("RemoteSchedule with zero vacancies fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial, // Non-online mode
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies: 0, // Invalid
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for zero vacancies")
		}
		if !strings.Contains(err.Error(), "número de vagas") {
			t.Errorf("Expected 'número de vagas' error, got: %v", err)
		}
	})

	t.Run("RemoteSchedule with excessive vacancies fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies: 1001, // Too many
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for excessive vacancies")
		}
		if !strings.Contains(err.Error(), "número de vagas") {
			t.Errorf("Expected 'número de vagas' error, got: %v", err)
		}
	})

	t.Run("RemoteSchedule non-online missing start date fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial, // Non-online requires dates
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies:      30,
						ClassStartDate: nil, // Required for non-online
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for missing start date")
		}
		if !strings.Contains(err.Error(), "data de início") {
			t.Errorf("Expected 'data de início' error, got: %v", err)
		}
	})

	t.Run("RemoteSchedule non-online missing end date fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now()
		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies:      30,
						ClassStartDate: &startDate,
						ClassEndDate:   nil, // Required for non-online
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for missing end date")
		}
		if !strings.Contains(err.Error(), "data de término") {
			t.Errorf("Expected 'data de término' error, got: %v", err)
		}
	})

	t.Run("RemoteSchedule non-online end before start fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now()
		endDate := startDate.Add(-24 * time.Hour) // Before start
		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies:      30,
						ClassStartDate: &startDate,
						ClassEndDate:   &endDate,
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for end date before start date")
		}
		if !strings.Contains(err.Error(), "data de término deve ser maior") {
			t.Errorf("Expected date comparison error, got: %v", err)
		}
	})

	t.Run("RemoteSchedule non-online missing class time fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now()
		endDate := startDate.Add(24 * time.Hour)
		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies:      30,
						ClassStartDate: &startDate,
						ClassEndDate:   &endDate,
						ClassTime:      nil, // Required for non-online
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for missing class time")
		}
		if !strings.Contains(err.Error(), "horário da aula") {
			t.Errorf("Expected 'horário da aula' error, got: %v", err)
		}
	})

	t.Run("RemoteSchedule non-online missing class days fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now()
		endDate := startDate.Add(24 * time.Hour)
		classTime := "09:00-12:00"
		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies:      30,
						ClassStartDate: &startDate,
						ClassEndDate:   &endDate,
						ClassTime:      &classTime,
						ClassDays:      nil, // Required for non-online
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for missing class days")
		}
		if !strings.Contains(err.Error(), "dias da semana") {
			t.Errorf("Expected 'dias da semana' error, got: %v", err)
		}
	})

	t.Run("RemoteSchedule online mode allows optional fields", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 300, nil
			},
			CreateRemoteClassFunc: func(ctx context.Context, rc *models.RemoteClass) error {
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Online Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadeRemoto, // Online mode
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies: 30,
						// All date/time/days fields are nil - should be valid for online
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Online course should succeed with optional fields: %v", err)
		}
	})

	t.Run("RemoteSchedule online mode with end before start fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now()
		endDate := startDate.Add(-24 * time.Hour) // Before start
		curso := &models.Curso{
			Titulo:     "Online Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadeRemoto, // Online mode
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies:      30,
						ClassStartDate: &startDate,
						ClassEndDate:   &endDate, // Even in online mode, if provided must be valid
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for end date before start date even in online mode")
		}
	})

	t.Run("RemoteSchedule class time too long fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now()
		endDate := startDate.Add(24 * time.Hour)
		longString := strings.Repeat("A", 20001)
		classTime := &longString
		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies:      30,
						ClassStartDate: &startDate,
						ClassEndDate:   &endDate,
						ClassTime:      classTime, // Too long
						ClassDays:      new(string),
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for class time too long")
		}
		if !strings.Contains(err.Error(), "20000 caracteres") {
			t.Errorf("Expected character limit error, got: %v", err)
		}
	})

	t.Run("RemoteSchedule class days too long fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now()
		endDate := startDate.Add(24 * time.Hour)
		longString := strings.Repeat("A", 20001)
		classTime := "09:00"
		classDays := &longString
		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies:      30,
						ClassStartDate: &startDate,
						ClassEndDate:   &endDate,
						ClassTime:      &classTime,
						ClassDays:      classDays, // Too long
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for class days too long")
		}
		if !strings.Contains(err.Error(), "20000 caracteres") {
			t.Errorf("Expected character limit error, got: %v", err)
		}
	})

	t.Run("RemoteSchedule online mode class time too long fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		longString := strings.Repeat("A", 20001)
		classTime := &longString
		curso := &models.Curso{
			Titulo:     "Online Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadeRemoto, // Online mode
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies: 30,
						ClassTime: classTime, // Even in online mode, if provided must be valid
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for class time too long even in online mode")
		}
	})

	t.Run("RemoteSchedule online mode class days too long fails", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		longString := strings.Repeat("A", 20001)
		classDays := &longString
		curso := &models.Curso{
			Titulo:     "Online Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadeRemoto, // Online mode
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{
						Vacancies: 30,
						ClassDays: classDays, // Even in online mode, if provided must be valid
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected validation error for class days too long even in online mode")
		}
	})
}
