package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestEmpregoRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	empresaID := 1
	emprego := &models.Emprego{
		Titulo:                "Desenvolvedor Go",
		Descricao:             "Desenvolvimento de APIs",
		EmpresaID:             &empresaID,
		TipoContratacao:       models.TipoContratacaoCLT,
		NumeroVagas:           5,
		Latitude:              -22.9068,
		Longitude:             -43.1729,
		SalarioMin:            5000,
		SalarioMax:            10000,
		Beneficios:            "VT, VR",
		PreRequisitos:         "Experiência com Go",
		DataInicioPrevista:    time.Now(),
		DataLimiteCandidatura: time.Now().AddDate(0, 1, 0),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "empregos"`).
		WithArgs(
			sqlmock.AnyArg(), // titulo
			sqlmock.AnyArg(), // descricao
			sqlmock.AnyArg(), // orgao_id
			sqlmock.AnyArg(), // empresa_id
			sqlmock.AnyArg(), // tipo_contratacao
			sqlmock.AnyArg(), // numero_vagas
			sqlmock.AnyArg(), // latitude
			sqlmock.AnyArg(), // longitude
			sqlmock.AnyArg(), // salario_min
			sqlmock.AnyArg(), // salario_max
			sqlmock.AnyArg(), // beneficios
			sqlmock.AnyArg(), // pre_requisitos
			sqlmock.AnyArg(), // data_inicio_prevista
			sqlmock.AnyArg(), // data_limite_candidatura
			sqlmock.AnyArg(), // status
			sqlmock.AnyArg(), // escolaridade_id
			sqlmock.AnyArg(), // jornada_trabalho
			sqlmock.AnyArg(), // turno
			sqlmock.AnyArg(), // contato_duvidas
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	id, err := repo.Create(ctx, emprego)

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpregoRepository_Create_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	emprego := &models.Emprego{
		Titulo: "Desenvolvedor Go",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "empregos"`).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	id, err := repo.Create(ctx, emprego)

	assert.Error(t, err)
	assert.Equal(t, 0, id)
	assert.Contains(t, err.Error(), "erro ao criar emprego")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpregoRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	empresaID := 1
	escolaridadeID := 2
	expectedEmprego := &models.Emprego{
		ID:             1,
		Titulo:         "Desenvolvedor Go",
		Descricao:      "Desenvolvimento de APIs",
		EmpresaID:      &empresaID,
		EscolaridadeID: &escolaridadeID,
	}

	// Mock for main query
	mock.ExpectQuery(`SELECT \* FROM "empregos" WHERE "empregos"\."id" = \$1 ORDER BY "empregos"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "titulo", "descricao", "empresa_id", "escolaridade_id"}).
			AddRow(1, "Desenvolvedor Go", "Desenvolvimento de APIs", empresaID, escolaridadeID))

	// Mock for Empresa preload
	mock.ExpectQuery(`SELECT \* FROM "empresas" WHERE "empresas"\."id" = \$1`).
		WithArgs(empresaID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Empresa Teste"))

	// Mock for Escolaridade preload
	mock.ExpectQuery(`SELECT \* FROM "escolaridades" WHERE "escolaridades"\."id" = \$1`).
		WithArgs(escolaridadeID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nivel"}).
			AddRow(2, "Superior Completo"))

	emprego, err := repo.GetByID(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, emprego)
	assert.Equal(t, expectedEmprego.ID, emprego.ID)
	assert.Equal(t, expectedEmprego.Titulo, emprego.Titulo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpregoRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "empregos" WHERE "empregos"\."id" = \$1 ORDER BY "empregos"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "titulo"}))

	emprego, err := repo.GetByID(ctx, 999)

	assert.NoError(t, err)
	assert.Nil(t, emprego)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpregoRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	emprego := &models.Emprego{
		ID:          1,
		Titulo:      "Desenvolvedor Go Senior",
		Descricao:   "Desenvolvimento de APIs escaláveis",
		NumeroVagas: 3,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "empregos" SET`).
		WithArgs(
			sqlmock.AnyArg(), // titulo
			sqlmock.AnyArg(), // descricao
			sqlmock.AnyArg(), // numero_vagas
			sqlmock.AnyArg(), // updated_at
			1,                // id in WHERE clause (from .Where("id = ?"))
			1,                // id in WHERE clause (from GORM's model ID)
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, emprego)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpregoRepository_Update_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	emprego := &models.Emprego{
		ID:     1,
		Titulo: "Desenvolvedor Go",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "empregos" SET`).
		WillReturnError(errors.New("update failed"))
	mock.ExpectRollback()

	err := repo.Update(ctx, emprego)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar emprego")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpregoRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "empregos" WHERE "empregos"\."id" = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpregoRepository_Delete_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "empregos"`).
		WillReturnError(errors.New("delete failed"))
	mock.ExpectRollback()

	err := repo.Delete(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir emprego")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpregoRepository_List(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{}
	limit := 10
	offset := 0

	// Mock count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "empregos"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock list query
	mock.ExpectQuery(`SELECT \* FROM "empregos" ORDER BY id DESC LIMIT \$1`).
		WithArgs(limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "titulo", "empresa_id", "escolaridade_id"}).
			AddRow(1, "Emprego 1", 1, 1).
			AddRow(2, "Emprego 2", 2, 2))

	// Mock for Empresa preload
	mock.ExpectQuery(`SELECT \* FROM "empresas" WHERE "empresas"\."id" IN \(\$1,\$2\)`).
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Empresa A").
			AddRow(2, "Empresa B"))

	// Mock for Escolaridade preload
	mock.ExpectQuery(`SELECT \* FROM "escolaridades" WHERE "escolaridades"\."id" IN \(\$1,\$2\)`).
		WithArgs(1, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nivel"}).
			AddRow(1, "Ensino Médio").
			AddRow(2, "Superior"))

	empregos, total, err := repo.List(ctx, filter, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, empregos, 2)
	assert.Equal(t, "Emprego 1", empregos[0].Titulo)
	assert.Equal(t, "Emprego 2", empregos[1].Titulo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpregoRepository_List_WithFilter(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	empresaID := 1
	filter := map[string]interface{}{
		"empresa_id": empresaID,
	}
	limit := 10
	offset := 0

	// Mock count query with filter
	mock.ExpectQuery(`SELECT count\(\*\) FROM "empregos" WHERE empresa_id = \$1`).
		WithArgs(empresaID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Mock list query with filter
	mock.ExpectQuery(`SELECT \* FROM "empregos" WHERE empresa_id = \$1 ORDER BY id DESC LIMIT \$2`).
		WithArgs(empresaID, limit).
		WillReturnRows(sqlmock.NewRows([]string{"id", "titulo", "empresa_id", "escolaridade_id"}).
			AddRow(1, "Emprego 1", empresaID, 1))

	// Mock for Empresa preload
	mock.ExpectQuery(`SELECT \* FROM "empresas" WHERE "empresas"\."id" = \$1`).
		WithArgs(empresaID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(1, "Empresa A"))

	// Mock for Escolaridade preload
	mock.ExpectQuery(`SELECT \* FROM "escolaridades" WHERE "escolaridades"\."id" = \$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nivel"}).
			AddRow(1, "Ensino Médio"))

	empregos, total, err := repo.List(ctx, filter, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, empregos, 1)
	assert.Equal(t, "Emprego 1", empregos[0].Titulo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEmpregoRepository_List_Error(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewEmpregoRepository(db)
	ctx := context.Background()

	filter := map[string]interface{}{}
	limit := 10
	offset := 0

	// Mock count query
	mock.ExpectQuery(`SELECT count\(\*\) FROM "empregos"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock list query error
	mock.ExpectQuery(`SELECT \* FROM "empregos"`).
		WillReturnError(errors.New("query failed"))

	empregos, total, err := repo.List(ctx, filter, limit, offset)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao listar empregos")
	assert.Nil(t, empregos)
	assert.Equal(t, 0, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}
