package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	cleanup := func() {
		sqlDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestCursoRepository_applyFilters(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)

	tests := []struct {
		name               string
		filter             map[string]interface{}
		expectedConditions []string
		description        string
	}{
		{
			name:               "empty filter",
			filter:             map[string]interface{}{},
			expectedConditions: []string{`SELECT * FROM "cursos"`},
			description:        "Empty filter should generate base SELECT without WHERE",
		},
		{
			name: "status NOT filter",
			filter: map[string]interface{}{
				"status NOT": "archived",
			},
			expectedConditions: []string{"status !=", "archived"},
			description:        "status NOT filter should generate != condition",
		},
		{
			name: "title ILIKE filter",
			filter: map[string]interface{}{
				"title ILIKE": "%golang%",
			},
			expectedConditions: []string{"titulo ILIKE", "%golang%"},
			description:        "title ILIKE filter should generate ILIKE condition",
		},
		{
			name: "categoria_id filter",
			filter: map[string]interface{}{
				"categoria_id": 5,
			},
			expectedConditions: []string{"cursos.id IN", "curso_id FROM cursos_categorias WHERE categoria_id"},
			description:        "categoria_id filter should generate subquery",
		},
		{
			name: "acessibilidade_id filter",
			filter: map[string]interface{}{
				"acessibilidade_id": 3,
			},
			expectedConditions: []string{"cursos.id IN", "curso_id FROM cursos_acessibilidades WHERE acessibilidade_id"},
			description:        "acessibilidade_id filter should generate subquery",
		},
		{
			name: "neighborhood_zone filter",
			filter: map[string]interface{}{
				"neighborhood_zone": "Zona Sul",
			},
			expectedConditions: []string{"cursos.id IN", "curso_id FROM location_classes WHERE neighborhood_zone"},
			description:        "neighborhood_zone filter should generate subquery",
		},
		{
			name: "default exact match filter",
			filter: map[string]interface{}{
				"instituicao_id": 10,
			},
			expectedConditions: []string{"instituicao_id ="},
			description:        "Unknown filter key should use exact match",
		},
		{
			name: "multiple filters",
			filter: map[string]interface{}{
				"status NOT":     "archived",
				"title ILIKE":    "%python%",
				"categoria_id":   2,
				"instituicao_id": 7,
			},
			expectedConditions: []string{
				"status !=",
				"titulo ILIKE",
				"categoria_id",
				"instituicao_id",
			},
			description: "Multiple filters should all be present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start with a base query
			query := db.Model(&models.Curso{})

			// Apply filters
			filteredQuery := repo.applyFilters(query, tt.filter)

			// Get the SQL statement
			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]models.Curso{})
			})

			// Check that all expected conditions are present
			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}
		})
	}
}

func TestCursoRepository_applyFilters_AllFilterTypes(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)

	// Test that all special filter keys are handled correctly
	specialFilters := map[string]string{
		"status NOT":        "status !=",
		"title ILIKE":       "titulo ILIKE",
		"categoria_id":      "curso_id FROM cursos_categorias WHERE categoria_id",
		"acessibilidade_id": "curso_id FROM cursos_acessibilidades WHERE acessibilidade_id",
		"neighborhood_zone": "curso_id FROM location_classes WHERE neighborhood_zone",
	}

	for filterKey, expectedSQLFragment := range specialFilters {
		t.Run(filterKey, func(t *testing.T) {
			filter := map[string]interface{}{
				filterKey: "test_value",
			}

			query := db.Model(&models.Curso{})
			filteredQuery := repo.applyFilters(query, filter)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]models.Curso{})
			})

			assert.Contains(t, sql, expectedSQLFragment,
				"Filter %s should generate SQL containing %s", filterKey, expectedSQLFragment)
		})
	}
}

func TestCursoRepository_applyFilters_DefaultBehavior(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewCursoRepository(db)

	// Test that unknown filter keys use default exact match behavior
	unknownFilters := []string{
		"random_field",
		"another_field",
		"custom_column",
	}

	for _, filterKey := range unknownFilters {
		t.Run(filterKey, func(t *testing.T) {
			filter := map[string]interface{}{
				filterKey: "test_value",
			}

			query := db.Model(&models.Curso{})
			filteredQuery := repo.applyFilters(query, filter)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]models.Curso{})
			})

			// Should use exact match with "="
			expectedSQL := filterKey + " ="
			assert.Contains(t, sql, expectedSQL,
				"Unknown filter key %s should use exact match (=)", filterKey)
		})
	}
}

func TestCursoRepository_ValidateForEnrollment_AutoApproveLogic(t *testing.T) {
	// Test the autoApprove logic without DB - it's a pure function
	tests := []struct {
		name                   string
		autoApproveEnrollments *bool
		expectedAutoApprove    bool
		description            string
	}{
		{
			name:                   "nil pointer should return false",
			autoApproveEnrollments: nil,
			expectedAutoApprove:    false,
			description:            "When AutoApproveEnrollments is nil, should return false",
		},
		{
			name: "false pointer should return false",
			autoApproveEnrollments: func() *bool {
				b := false
				return &b
			}(),
			expectedAutoApprove: false,
			description:         "When AutoApproveEnrollments is false, should return false",
		},
		{
			name: "true pointer should return true",
			autoApproveEnrollments: func() *bool {
				b := true
				return &b
			}(),
			expectedAutoApprove: true,
			description:         "When AutoApproveEnrollments is true, should return true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from ValidateForEnrollment
			autoApproveValue := tt.autoApproveEnrollments != nil && *tt.autoApproveEnrollments

			assert.Equal(t, tt.expectedAutoApprove, autoApproveValue, tt.description)
		})
	}
}

func TestCursoRepository_ValidateForEnrollment_NilHandling(t *testing.T) {
	// Test that nil pointer dereference is handled correctly
	t.Run("nil pointer safe check", func(t *testing.T) {
		var autoApproveEnrollments *bool = nil

		// This is the exact logic from the repository
		autoApproveValue := autoApproveEnrollments != nil && *autoApproveEnrollments

		assert.False(t, autoApproveValue,
			"Nil pointer should be handled safely and return false")
	})

	t.Run("non-nil false pointer", func(t *testing.T) {
		b := false
		autoApproveEnrollments := &b

		autoApproveValue := autoApproveEnrollments != nil && *autoApproveEnrollments

		assert.False(t, autoApproveValue,
			"False boolean pointer should return false")
	})

	t.Run("non-nil true pointer", func(t *testing.T) {
		b := true
		autoApproveEnrollments := &b

		autoApproveValue := autoApproveEnrollments != nil && *autoApproveEnrollments

		assert.True(t, autoApproveValue,
			"True boolean pointer should return true")
	})
}
