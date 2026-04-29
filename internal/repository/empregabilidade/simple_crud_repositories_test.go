package empregabilidade

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
)

// TestDisponibilidadeRepository tests all CRUD operations
func TestDisponibilidadeRepository_Create(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewDisponibilidadeRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		entity := &empregabilidade.Disponibilidade{
			Descricao: "Integral",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_disponibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		id, err := repo.Create(ctx, entity)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		entity := &empregabilidade.Disponibilidade{
			Descricao: "Integral",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_disponibilidades"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.Create(ctx, entity)
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
		assert.Contains(t, err.Error(), "erro ao criar disponibilidade")
	})
}

func TestDisponibilidadeRepository_GetByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewDisponibilidadeRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "descricao"}).
			AddRow(id, "Integral")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_disponibilidades"`)).
			WillReturnRows(rows)

		entity, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, entity)
		assert.Equal(t, id, entity.ID)
		assert.Equal(t, "Integral", entity.Descricao)
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_disponibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		entity, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, entity)
	})

	t.Run("get by id database error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_disponibilidades"`)).
			WillReturnError(assert.AnError)

		entity, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, entity)
		assert.Contains(t, err.Error(), "erro ao buscar disponibilidade")
	})
}

func TestDisponibilidadeRepository_Update(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewDisponibilidadeRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		entity := &empregabilidade.Disponibilidade{
			ID:        uuid.New(),
			Descricao: "Parcial",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_disponibilidades"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, entity)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		entity := &empregabilidade.Disponibilidade{
			ID:        uuid.New(),
			Descricao: "Parcial",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_disponibilidades"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Update(ctx, entity)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar disponibilidade")
	})
}

func TestDisponibilidadeRepository_Delete(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewDisponibilidadeRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_disponibilidades"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_disponibilidades"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir disponibilidade")
	})
}

func TestDisponibilidadeRepository_List(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewDisponibilidadeRepository(db)
	ctx := context.Background()

	t.Run("list success with filters", func(t *testing.T) {
		filter := map[string]interface{}{
			"descricao": "Integral",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_disponibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "descricao"}).
			AddRow(uuid.New(), "Integral").
			AddRow(uuid.New(), "Integral")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_disponibilidades"`)).
			WillReturnRows(rows)

		entities, total, err := repo.List(ctx, filter, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, entities, 2)
	})

	t.Run("list error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_disponibilidades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_disponibilidades"`)).
			WillReturnError(assert.AnError)

		entities, total, err := repo.List(ctx, nil, 10, 0)
		assert.Error(t, err)
		assert.Nil(t, entities)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "erro ao listar disponibilidades")
	})
}

// TestIdiomaRepository tests all CRUD operations
func TestIdiomaRepository_Create(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewIdiomaRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		entity := &empregabilidade.Idioma{
			Descricao: "Inglês",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_idiomas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		id, err := repo.Create(ctx, entity)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
	})

	t.Run("create error", func(t *testing.T) {
		entity := &empregabilidade.Idioma{
			Descricao: "Espanhol",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_idiomas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.Create(ctx, entity)
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
		assert.Contains(t, err.Error(), "erro ao criar idioma")
	})
}

func TestIdiomaRepository_GetByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewIdiomaRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "descricao"}).
			AddRow(id, "Inglês")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_idiomas"`)).
			WillReturnRows(rows)

		entity, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, entity)
		assert.Equal(t, id, entity.ID)
		assert.Equal(t, "Inglês", entity.Descricao)
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_idiomas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		entity, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, entity)
	})

	t.Run("get by id database error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_idiomas"`)).
			WillReturnError(assert.AnError)

		entity, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, entity)
		assert.Contains(t, err.Error(), "erro ao buscar idioma")
	})
}

func TestIdiomaRepository_Update(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewIdiomaRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		entity := &empregabilidade.Idioma{
			ID:        uuid.New(),
			Descricao: "Francês",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_idiomas"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, entity)
		assert.NoError(t, err)
	})

	t.Run("update error", func(t *testing.T) {
		entity := &empregabilidade.Idioma{
			ID:        uuid.New(),
			Descricao: "Alemão",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_idiomas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Update(ctx, entity)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar idioma")
	})
}

func TestIdiomaRepository_Delete(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewIdiomaRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_idiomas"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_idiomas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir idioma")
	})
}

func TestIdiomaRepository_List(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewIdiomaRepository(db)
	ctx := context.Background()

	t.Run("list success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_idiomas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "descricao"}).
			AddRow(uuid.New(), "Inglês").
			AddRow(uuid.New(), "Espanhol")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_idiomas"`)).
			WillReturnRows(rows)

		entities, total, err := repo.List(ctx, nil, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, entities, 2)
	})

	t.Run("list error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_idiomas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_idiomas"`)).
			WillReturnError(assert.AnError)

		entities, total, err := repo.List(ctx, nil, 10, 0)
		assert.Error(t, err)
		assert.Nil(t, entities)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "erro ao listar idiomas")
	})
}
