package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// EXPANDED TESTS FOR CURSO SERVICE
// This file adds 20+ edge cases and untested scenarios

// ==================== Complex Search/Filter Scenarios ====================

func TestCursoService_List_ComplexFilters(t *testing.T) {
	ctx := context.Background()

	t.Run("List with multiple filter combinations", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				// Verify filter contains expected keys
				if filter["status"] != string(models.StatusCursoOpened) {
					t.Error("Expected status filter")
				}
				if filter["modalidade"] != string(models.ModalidadePresencial) {
					t.Error("Expected modalidade filter")
				}
				return []*models.Curso{{ID: 1}}, 1, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		filter := map[string]interface{}{
			"status":     string(models.StatusCursoOpened),
			"modalidade": string(models.ModalidadePresencial),
		}

		_, total, err := svc.List(ctx, filter, 1, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
	})

	t.Run("List with empty filter returns all", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				return []*models.Curso{
					{ID: 1}, {ID: 2}, {ID: 3},
				}, 3, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}
	})

	t.Run("List with nil filter", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				return []*models.Curso{{ID: 1}}, 1, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, _, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List with nil filter should succeed: %v", err)
		}
	})
}

// ==================== Pagination Edge Cases ====================

func TestCursoService_List_PaginationEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Pagination with page 1", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				if offset != 0 {
					t.Errorf("Expected offset 0 for page 1, got %d", offset)
				}
				if limit != 10 {
					t.Errorf("Expected limit 10, got %d", limit)
				}
				return []*models.Curso{}, 0, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, _, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
	})

	t.Run("Pagination with large page number", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				if offset != 990 { // (100-1) * 10
					t.Errorf("Expected offset 990 for page 100, got %d", offset)
				}
				return []*models.Curso{}, 0, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, _, err := svc.List(ctx, nil, 100, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
	})

	t.Run("Pagination with very large page size", func(t *testing.T) {
		repo := &MockCursoRepository{
			ListFunc: func(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Curso, int, error) {
				if limit != 1000 {
					t.Errorf("Expected limit 1000, got %d", limit)
				}
				return []*models.Curso{}, 0, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		_, _, err := svc.List(ctx, nil, 1, 1000)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
	})
}

// ==================== Draft vs Published Workflow ====================

func TestCursoService_DraftToPublishedWorkflow(t *testing.T) {
	ctx := context.Background()

	t.Run("Create as draft then update to published", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 1, nil
			},
			UpdateFunc: func(ctx context.Context, curso *models.Curso) error {
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		// Create as draft with minimal data
		draft := &models.Curso{
			Titulo: "Draft Course",
			Status: models.StatusCursoDraft,
		}

		id, err := svc.Create(ctx, draft)
		if err != nil {
			t.Fatalf("Create draft failed: %v", err)
		}
		if id != 1 {
			t.Errorf("Expected ID 1, got %d", id)
		}

		// Update to published with full data
		draft.ID = 1
		draft.Status = models.StatusCursoOpened
		draft.Modalidade = models.ModalidadePresencial
		draft.NumeroVagas = 50
		draft.CargaHoraria = 40

		err = svc.Update(ctx, draft)
		if err != nil {
			t.Fatalf("Update to published failed: %v", err)
		}
	})

	t.Run("Draft accepts optional fields as empty", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 1, nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		draft := &models.Curso{
			Titulo:         "Draft Course",
			Status:         models.StatusCursoDraft,
			Modalidade:     "", // Optional for draft
			NumeroVagas:    0,  // Optional for draft
			CargaHoraria:   0,  // Optional for draft
			Organization:   "",
			Theme:          "",
			TargetAudience: "",
		}

		_, err := svc.Create(ctx, draft)
		if err != nil {
			t.Fatalf("Create draft with optional empty fields failed: %v", err)
		}
	})

	t.Run("Published course requires full validation", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		published := &models.Curso{
			Titulo:      "Published Course",
			Status:      models.StatusCursoOpened,
			Modalidade:  "", // Invalid: required for published
			NumeroVagas: 50,
		}

		_, err := svc.Create(ctx, published)
		if err == nil {
			t.Error("Expected validation error for published course without modalidade")
		}
	})
}

// ==================== Location vs Remote Class Validation ====================

func TestCursoService_LocationClassConflicts(t *testing.T) {
	ctx := context.Background()

	t.Run("LIVRE_FORMACAO_ONLINE cannot have locations", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:       "Online Formation",
			Status:       models.StatusCursoOpened,
			Modalidade:   models.ModalidadeLivreFormacaoOnline,
			FormacaoLink: "https://example.com/course",
			LocationClasses: []models.LocationClass{
				{
					Address:      "Address 123",
					Neighborhood: "Neighborhood",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      30,
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "09:00",
							ClassDays:      "Mon-Fri",
						},
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected error: LIVRE_FORMACAO_ONLINE cannot have locations")
		}
		if !strings.Contains(err.Error(), "não deve ter locations") {
			t.Errorf("Expected 'não deve ter locations' error, got: %v", err)
		}
	})

	t.Run("LIVRE_FORMACAO_ONLINE cannot have remote class", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:       "Online Formation",
			Status:       models.StatusCursoOpened,
			Modalidade:   models.ModalidadeLivreFormacaoOnline,
			FormacaoLink: "https://example.com/course",
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{Vacancies: 100},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected error: LIVRE_FORMACAO_ONLINE cannot have remote_class")
		}
		if !strings.Contains(err.Error(), "não deve ter remote_class") {
			t.Errorf("Expected 'não deve ter remote_class' error, got: %v", err)
		}
	})

	t.Run("LIVRE_FORMACAO_ONLINE requires formacao_link", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:       "Online Formation",
			Status:       models.StatusCursoOpened,
			Modalidade:   models.ModalidadeLivreFormacaoOnline,
			FormacaoLink: "", // Missing required field
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected error: LIVRE_FORMACAO_ONLINE requires formacao_link")
		}
		if !strings.Contains(err.Error(), "formacao_link é obrigatório") {
			t.Errorf("Expected 'formacao_link é obrigatório' error, got: %v", err)
		}
	})

	t.Run("LIVRE_FORMACAO_ONLINE requires valid URL", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:       "Online Formation",
			Status:       models.StatusCursoOpened,
			Modalidade:   models.ModalidadeLivreFormacaoOnline,
			FormacaoLink: "invalid-url",
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected error: invalid URL")
		}
		if !strings.Contains(err.Error(), "URL válida") {
			t.Errorf("Expected 'URL válida' error, got: %v", err)
		}
	})
}

// ==================== Schedule Conflicts & Complex Cases ====================

func TestCursoService_ScheduleValidationEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("Location with zero vacancies in schedule", func(t *testing.T) {
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
							Vacancies:      0, // Invalid
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
			t.Error("Expected error for zero vacancies")
		}
		if !strings.Contains(err.Error(), "número de vagas") {
			t.Errorf("Expected 'número de vagas' error, got: %v", err)
		}
	})

	t.Run("Schedule with same start and end date is valid", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 1, nil
			},
			CreateLocationClassesFunc: func(ctx context.Context, locations []models.LocationClass) error {
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		sameDate := time.Now().Add(24 * time.Hour)

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
							ClassStartDate: sameDate,
							ClassEndDate:   sameDate, // Same as start
							ClassTime:      "09:00-12:00",
							ClassDays:      "Segunda",
						},
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Same date should be valid: %v", err)
		}
	})

	t.Run("Multiple locations with multiple schedules each", func(t *testing.T) {
		var capturedLocations []models.LocationClass
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 1, nil
			},
			CreateLocationClassesFunc: func(ctx context.Context, locations []models.LocationClass) error {
				capturedLocations = locations
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Multi-location Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Location 1 Address",
					Neighborhood: "Neighborhood 1",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      30,
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "09:00-12:00",
							ClassDays:      "Segunda a Sexta",
						},
						{
							Vacancies:      20,
							ClassStartDate: startDate.Add(168 * time.Hour),
							ClassEndDate:   endDate.Add(168 * time.Hour),
							ClassTime:      "14:00-17:00",
							ClassDays:      "Terça e Quinta",
						},
					},
				},
				{
					Address:      "Location 2 Address",
					Neighborhood: "Neighborhood 2",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      50,
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "08:00-11:00",
							ClassDays:      "Segunda a Quarta",
						},
					},
				},
			},
		}

		id, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id != 1 {
			t.Errorf("Expected ID 1, got %d", id)
		}
		if len(capturedLocations) != 2 {
			t.Errorf("Expected 2 locations, got %d", len(capturedLocations))
		}
		if len(capturedLocations[0].Schedules) != 2 {
			t.Errorf("Expected 2 schedules in location 1, got %d", len(capturedLocations[0].Schedules))
		}
	})
}

// ==================== Capacity Management ====================

func TestCursoService_CapacityChecking(t *testing.T) {
	ctx := context.Background()

	t.Run("Maximum capacity 1000 is allowed", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 1, nil
			},
			CreateLocationClassesFunc: func(ctx context.Context, locations []models.LocationClass) error {
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Large Capacity Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Valid Address 123",
					Neighborhood: "Test Neighborhood",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      1000, // Maximum allowed
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
		if err != nil {
			t.Fatalf("Create with max capacity failed: %v", err)
		}
	})

	t.Run("Capacity 1001 exceeds limit", func(t *testing.T) {
		repo := &MockCursoRepository{}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Over Capacity Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Valid Address 123",
					Neighborhood: "Test Neighborhood",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      1001, // Over limit
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
			t.Error("Expected error for capacity over 1000")
		}
	})

	t.Run("Minimum capacity 1 is valid", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 1, nil
			},
			CreateLocationClassesFunc: func(ctx context.Context, locations []models.LocationClass) error {
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		startDate := time.Now().Add(24 * time.Hour)
		endDate := startDate.Add(72 * time.Hour)

		curso := &models.Curso{
			Titulo:     "Small Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{
					Address:      "Valid Address 123",
					Neighborhood: "Test Neighborhood",
					Schedules: []models.CourseSchedule{
						{
							Vacancies:      1, // Minimum
							ClassStartDate: startDate,
							ClassEndDate:   endDate,
							ClassTime:      "09:00-12:00",
							ClassDays:      "Segunda",
						},
					},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("Create with min capacity failed: %v", err)
		}
	})
}

// ==================== Error Propagation from Repository ====================

func TestCursoService_RepositoryErrorPropagation(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateCustomFields error is propagated", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 1, nil
			},
			CreateCustomFieldsFunc: func(ctx context.Context, fields []models.CustomField) error {
				return errors.New("custom fields database error")
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo: "Course with Custom Fields",
			Status: models.StatusCursoDraft,
			CustomFields: []models.CustomField{
				{Title: "Field1", Required: true},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected error from CreateCustomFields")
		}
		if !strings.Contains(err.Error(), "custom fields") {
			t.Errorf("Expected 'custom fields' error, got: %v", err)
		}
	})

	t.Run("CreateRemoteClass error is propagated", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 1, nil
			},
			CreateRemoteClassFunc: func(ctx context.Context, rc *models.RemoteClass) error {
				return errors.New("remote class database error")
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo:     "Remote Course",
			Status:     models.StatusCursoOpened,
			Modalidade: models.ModalidadeRemoto,
			RemoteClass: &models.RemoteClass{
				Schedules: []models.RemoteSchedule{
					{Vacancies: 30},
				},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err == nil {
			t.Error("Expected error from CreateRemoteClass")
		}
		if !strings.Contains(err.Error(), "remote class") {
			t.Errorf("Expected 'remote class' error, got: %v", err)
		}
	})

	t.Run("CreateLocationClasses error is propagated", func(t *testing.T) {
		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 1, nil
			},
			CreateLocationClassesFunc: func(ctx context.Context, locations []models.LocationClass) error {
				return errors.New("location classes database error")
			},
		}
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
			t.Error("Expected error from CreateLocationClasses")
		}
		if !strings.Contains(err.Error(), "location classes") {
			t.Errorf("Expected 'location classes' error, got: %v", err)
		}
	})
}

// ==================== FormatType Propagation ====================

func TestCursoService_Create_CustomField_FormatType(t *testing.T) {
	ctx := context.Background()

	t.Run("format_type is preserved when creating custom fields", func(t *testing.T) {
		var capturedFields []models.CustomField

		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 1, nil
			},
			CreateCustomFieldsFunc: func(ctx context.Context, fields []models.CustomField) error {
				capturedFields = fields
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		ft := "cpf"
		curso := &models.Curso{
			Titulo: "Curso com Máscara",
			Status: models.StatusCursoDraft,
			CustomFields: []models.CustomField{
				{Title: "CPF do participante", FieldType: "text", FormatType: &ft, Required: true},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(capturedFields) != 1 {
			t.Fatalf("expected 1 custom field, got %d", len(capturedFields))
		}
		if capturedFields[0].FormatType == nil || *capturedFields[0].FormatType != "cpf" {
			t.Errorf("expected FormatType 'cpf', got %v", capturedFields[0].FormatType)
		}
		if capturedFields[0].CursoID != 1 {
			t.Errorf("expected CursoID 1, got %d", capturedFields[0].CursoID)
		}
	})

	t.Run("format_type nil is preserved (free text field)", func(t *testing.T) {
		var capturedFields []models.CustomField

		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 2, nil
			},
			CreateCustomFieldsFunc: func(ctx context.Context, fields []models.CustomField) error {
				capturedFields = fields
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		curso := &models.Curso{
			Titulo: "Curso texto livre",
			Status: models.StatusCursoDraft,
			CustomFields: []models.CustomField{
				{Title: "Observações", FieldType: "text"},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(capturedFields) != 1 {
			t.Fatalf("expected 1 custom field, got %d", len(capturedFields))
		}
		if capturedFields[0].FormatType != nil {
			t.Errorf("expected FormatType nil, got %v", capturedFields[0].FormatType)
		}
	})

	t.Run("multiple fields with mixed format_type values", func(t *testing.T) {
		var capturedFields []models.CustomField

		repo := &MockCursoRepository{
			CreateFunc: func(ctx context.Context, curso *models.Curso) (int, error) {
				return 3, nil
			},
			CreateCustomFieldsFunc: func(ctx context.Context, fields []models.CustomField) error {
				capturedFields = fields
				return nil
			},
		}
		svc := services.NewCursoServiceWithInterface(repo)

		ftCPF := "cpf"
		ftPhone := "phone"
		curso := &models.Curso{
			Titulo: "Curso múltiplos campos",
			Status: models.StatusCursoDraft,
			CustomFields: []models.CustomField{
				{Title: "CPF", FieldType: "text", FormatType: &ftCPF},
				{Title: "Nome", FieldType: "text"},
				{Title: "Telefone", FieldType: "text", FormatType: &ftPhone},
			},
		}

		_, err := svc.Create(ctx, curso)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(capturedFields) != 3 {
			t.Fatalf("expected 3 custom fields, got %d", len(capturedFields))
		}
		if capturedFields[0].FormatType == nil || *capturedFields[0].FormatType != "cpf" {
			t.Errorf("field[0]: expected FormatType 'cpf', got %v", capturedFields[0].FormatType)
		}
		if capturedFields[1].FormatType != nil {
			t.Errorf("field[1]: expected FormatType nil, got %v", capturedFields[1].FormatType)
		}
		if capturedFields[2].FormatType == nil || *capturedFields[2].FormatType != "phone" {
			t.Errorf("field[2]: expected FormatType 'phone', got %v", capturedFields[2].FormatType)
		}
	})
}
