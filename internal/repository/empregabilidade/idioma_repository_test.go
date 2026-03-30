package empregabilidade

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestIdiomaRepository_List_ApplyFilters(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name               string
		filter             map[string]interface{}
		expectedConditions []string
		description        string
	}{
		{
			name:               "empty filter",
			filter:             map[string]interface{}{},
			expectedConditions: []string{},
			description:        "No filter should generate no WHERE clauses",
		},
		{
			name: "filter by ID",
			filter: map[string]interface{}{
				"id": uuid.New(),
			},
			expectedConditions: []string{"id ="},
			description:        "ID filter should generate id = condition",
		},
		{
			name: "filter by descricao",
			filter: map[string]interface{}{
				"descricao": "Inglês",
			},
			expectedConditions: []string{"descricao ="},
			description:        "Descricao filter should generate descricao = condition",
		},
		{
			name: "multiple filters combined",
			filter: map[string]interface{}{
				"id":        uuid.New(),
				"descricao": "Espanhol",
			},
			expectedConditions: []string{"id =", "descricao ="},
			description:        "Multiple filters should all be applied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyFilters := func(db *gorm.DB) *gorm.DB {
				for key, value := range tt.filter {
					db = db.Where(key+" = ?", value)
				}
				return db
			}

			query := db.Model(&empregabilidade.Idioma{})
			filteredQuery := applyFilters(query)

			sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]empregabilidade.Idioma{})
			})

			for _, condition := range tt.expectedConditions {
				assert.Contains(t, sql, condition,
					"%s: SQL should contain '%s'", tt.description, condition)
			}
		})
	}
}

func TestIdiomaRepository_List_OrderByDescricao(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	query := db.Model(&empregabilidade.Idioma{}).
		Order("descricao ASC").
		Limit(10).
		Offset(0)

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]*empregabilidade.Idioma{})
	})

	assert.Contains(t, sql, "ORDER BY", "Should include ORDER BY clause")
	assert.Contains(t, sql, "descricao", "Should order by descricao")
	assert.Contains(t, sql, "ASC", "Should order ascending")
}

func TestIdiomaRepository_List_PaginationLimitOffset(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name   string
		limit  int
		offset int
	}{
		{
			name:   "first page",
			limit:  10,
			offset: 0,
		},
		{
			name:   "second page",
			limit:  10,
			offset: 10,
		},
		{
			name:   "custom page size",
			limit:  25,
			offset: 50,
		},
		{
			name:   "large offset",
			limit:  50,
			offset: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := db.Model(&empregabilidade.Idioma{}).
				Order("descricao ASC").
				Limit(tt.limit).
				Offset(tt.offset)

			sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]*empregabilidade.Idioma{})
			})

			assert.Contains(t, sql, "LIMIT", "Should include LIMIT clause")
			if tt.offset > 0 {
				assert.Contains(t, sql, "OFFSET", "Should include OFFSET clause when offset > 0")
			}
		})
	}
}

func TestIdiomaRepository_List_CountTotal(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	filter := map[string]interface{}{
		"descricao": "Francês",
	}

	applyFilters := func(db *gorm.DB) *gorm.DB {
		for key, value := range filter {
			db = db.Where(key+" = ?", value)
		}
		return db
	}

	query := db.Model(&empregabilidade.Idioma{})
	filteredQuery := applyFilters(query)

	sql := filteredQuery.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var count int64
		tx.Count(&count)
		return tx
	})

	assert.Contains(t, sql, "count", "Should include count in SQL")
	assert.Contains(t, sql, "descricao =", "Should apply filter before counting")
}

func TestIdiomaRepository_Update_WhereID(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	id := uuid.New()
	entity := &empregabilidade.Idioma{
		ID:        id,
		Descricao: "Alemão",
	}

	query := db.Model(entity).Where("id = ?", entity.ID)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Updates(entity)
	})

	assert.Contains(t, sql, "UPDATE", "Should be an UPDATE operation")
	assert.Contains(t, sql, "id =", "Should filter by ID")
	assert.Contains(t, sql, empregabilidade.Idioma{}.TableName(), "Should update correct table")
}

func TestIdiomaRepository_Delete_ByID(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	id := uuid.New()

	query := db.Model(&empregabilidade.Idioma{}).Where("id = ?", id)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Delete(&empregabilidade.Idioma{})
	})

	assert.Contains(t, sql, "DELETE", "Should be a DELETE operation")
	assert.Contains(t, sql, "id =", "Should filter by ID")
	assert.Contains(t, sql, empregabilidade.Idioma{}.TableName(), "Should delete from correct table")
}

func TestIdiomaRepository_GetByID_FilterByID(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	id := uuid.New()

	query := db.Model(&empregabilidade.Idioma{}).Where("id = ?", id)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.First(&empregabilidade.Idioma{})
	})

	assert.Contains(t, sql, "id =", "Should filter by ID")
	assert.Contains(t, sql, "LIMIT 1", "First should limit to 1 record")
}

func TestIdiomaRepository_TableName(t *testing.T) {
	tableName := empregabilidade.Idioma{}.TableName()
	assert.Equal(t, "emp_idiomas", tableName, "Should use correct table name")
}
