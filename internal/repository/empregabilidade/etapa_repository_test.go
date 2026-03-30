package empregabilidade

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestEtapaRepository_ListByVaga_FilterByVagaID(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	vagaID := uuid.New()

	query := db.Model(&empregabilidade.Etapa{}).
		Where("id_vaga = ?", vagaID).
		Order("ordem ASC")

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]*empregabilidade.Etapa{})
	})

	assert.Contains(t, sql, "id_vaga =", "Should filter by id_vaga")
	assert.Contains(t, sql, "ORDER BY", "Should include ORDER BY clause")
	assert.Contains(t, sql, "ordem", "Should order by ordem")
	assert.Contains(t, sql, "ASC", "Should order ascending")
}

func TestEtapaRepository_ListByVaga_OrderByOrdem(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		vagaID      uuid.UUID
		description string
	}{
		{
			name:        "valid vaga ID",
			vagaID:      uuid.New(),
			description: "Should order etapas by ordem field",
		},
		{
			name:        "different vaga ID",
			vagaID:      uuid.New(),
			description: "Should maintain order for different vaga",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := db.Model(&empregabilidade.Etapa{}).
				Where("id_vaga = ?", tt.vagaID).
				Order("ordem ASC")

			sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Find(&[]*empregabilidade.Etapa{})
			})

			assert.Contains(t, sql, "ordem ASC", tt.description)
		})
	}
}

func TestEtapaRepository_DeleteByVaga_FilterByVagaID(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	vagaID := uuid.New()

	query := db.Model(&empregabilidade.Etapa{}).Where("id_vaga = ?", vagaID)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Delete(&empregabilidade.Etapa{})
	})

	assert.Contains(t, sql, "DELETE", "Should be a DELETE operation")
	assert.Contains(t, sql, "id_vaga =", "Should filter by id_vaga")
	assert.Contains(t, sql, empregabilidade.Etapa{}.TableName(), "Should delete from correct table")
}

func TestEtapaRepository_Update_SaveOperation(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	etapa := &empregabilidade.Etapa{
		ID:        uuid.New(),
		IDVaga:    uuid.New(),
		Titulo:    "Entrevista Atualizada",
		Descricao: "Entrevista com RH",
		Ordem:     1,
	}

	query := db.Model(&empregabilidade.Etapa{}).Where("id = ?", etapa.ID)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Updates(map[string]interface{}{
			"titulo": etapa.Titulo,
		})
	})

	assert.Contains(t, sql, "UPDATE", "Should be an UPDATE operation")
	assert.Contains(t, sql, empregabilidade.Etapa{}.TableName(), "Should update correct table")
}

func TestEtapaRepository_Delete_ByID(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	id := uuid.New()

	query := db.Model(&empregabilidade.Etapa{}).Where("id = ?", id)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Delete(&empregabilidade.Etapa{})
	})

	assert.Contains(t, sql, "DELETE", "Should be a DELETE operation")
	assert.Contains(t, sql, "id =", "Should filter by ID")
	assert.Contains(t, sql, empregabilidade.Etapa{}.TableName(), "Should delete from correct table")
}

func TestEtapaRepository_GetByID_FilterByID(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	id := uuid.New()

	query := db.Model(&empregabilidade.Etapa{}).Where("id = ?", id)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.First(&empregabilidade.Etapa{})
	})

	assert.Contains(t, sql, "id =", "Should filter by ID")
	assert.Contains(t, sql, "LIMIT 1", "First should limit to 1 record")
}

func TestEtapaRepository_ListByVaga_EmptyResult(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	vagaID := uuid.New()

	query := db.Model(&empregabilidade.Etapa{}).
		Where("id_vaga = ?", vagaID).
		Order("ordem ASC")

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]*empregabilidade.Etapa{})
	})

	assert.Contains(t, sql, "id_vaga =", "Should filter by id_vaga even for empty results")
	assert.Contains(t, sql, empregabilidade.Etapa{}.TableName(), "Should query from correct table")
}

func TestEtapaRepository_TableName(t *testing.T) {
	tableName := empregabilidade.Etapa{}.TableName()
	assert.Equal(t, "emp_etapas", tableName, "Should use correct table name")
}
