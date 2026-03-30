package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestPropostaMEIRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewPropostaMEIRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		proposta := &models.PropostaMEI{
			ID:                uuid.New(),
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
			StatusCidadao:     "pending",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(proposta.ID))
		mock.ExpectCommit()

		id, err := repo.Create(ctx, proposta)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		proposta := &models.PropostaMEI{
			ID:                uuid.New(),
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "propostas_mei"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.Create(ctx, proposta)
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
		assert.Contains(t, err.Error(), "erro ao criar proposta MEI")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPropostaMEIRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewPropostaMEIRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "oportunidade_mei_id", "mei_empresa_id", "status_cidadao"}).
			AddRow(id, 1, "12345678000190", "pending")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnRows(rows)

		// Expect oportunidade preload query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		proposta, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, proposta)
		assert.Equal(t, id, proposta.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		proposta, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, proposta)
	})

	t.Run("get by id error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnError(assert.AnError)

		proposta, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, proposta)
		assert.Contains(t, err.Error(), "erro ao buscar proposta MEI por ID")
	})
}

func TestPropostaMEIRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewPropostaMEIRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		proposta := &models.PropostaMEI{
			ID:                uuid.New(),
			OportunidadeMEIID: 1,
			MEIEmpresaID:      "12345678000190",
			StatusCidadao:     "approved",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "propostas_mei"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, proposta)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		proposta := &models.PropostaMEI{
			ID:           uuid.New(),
			StatusCidadao: "rejected",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "propostas_mei"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Update(ctx, proposta)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar proposta MEI")
	})
}

func TestPropostaMEIRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewPropostaMEIRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()

		// Soft delete uses UPDATE to set deleted_at
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "propostas_mei"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		id := uuid.New()

		// Soft delete uses UPDATE to set deleted_at
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "propostas_mei"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir proposta MEI")
	})
}

func TestPropostaMEIRepository_ListByOportunidade(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewPropostaMEIRepository(db)
	ctx := context.Background()

	t.Run("list with no filters", func(t *testing.T) {
		oportunidadeID := 1

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "oportunidade_mei_id", "mei_empresa_id"}).
			AddRow(uuid.New(), oportunidadeID, "12345678000190").
			AddRow(uuid.New(), oportunidadeID, "98765432000110")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnRows(rows)

		// Expect oportunidade preload queries
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		propostas, total, err := repo.ListByOportunidade(ctx, oportunidadeID, "", "", "", 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, propostas, 2)
	})

	t.Run("list with cnpj filter", func(t *testing.T) {
		oportunidadeID := 1
		cnpj := "12345678"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "oportunidade_mei_id", "mei_empresa_id"}).
			AddRow(uuid.New(), oportunidadeID, "12345678000190")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnRows(rows)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		propostas, total, err := repo.ListByOportunidade(ctx, oportunidadeID, "", cnpj, "", 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, propostas, 1)
	})

	t.Run("list with status filter", func(t *testing.T) {
		oportunidadeID := 1
		status := "approved"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "oportunidade_mei_id", "status_cidadao"}).
			AddRow(uuid.New(), oportunidadeID, status)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnRows(rows)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		propostas, total, err := repo.ListByOportunidade(ctx, oportunidadeID, "", "", status, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, propostas, 1)
	})

	t.Run("database error on find", func(t *testing.T) {
		oportunidadeID := 1

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnError(assert.AnError)

		propostas, total, err := repo.ListByOportunidade(ctx, oportunidadeID, "", "", "", 10, 0)
		assert.Error(t, err)
		assert.Nil(t, propostas)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "erro ao listar propostas MEI por oportunidade")
	})
}

func TestPropostaMEIRepository_ListByMEIEmpresa(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewPropostaMEIRepository(db)
	ctx := context.Background()

	t.Run("list by empresa success", func(t *testing.T) {
		meiEmpresaID := "12345678000190"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		rows := sqlmock.NewRows([]string{"id", "mei_empresa_id"}).
			AddRow(uuid.New(), meiEmpresaID).
			AddRow(uuid.New(), meiEmpresaID).
			AddRow(uuid.New(), meiEmpresaID)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnRows(rows)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		propostas, total, err := repo.ListByMEIEmpresa(ctx, meiEmpresaID, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, propostas, 3)
	})

	t.Run("database error on find", func(t *testing.T) {
		meiEmpresaID := "12345678000190"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnError(assert.AnError)

		propostas, total, err := repo.ListByMEIEmpresa(ctx, meiEmpresaID, 10, 0)
		assert.Error(t, err)
		assert.Nil(t, propostas)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "erro ao listar propostas MEI por empresa")
	})
}

func TestPropostaMEIRepository_ListByStatus(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewPropostaMEIRepository(db)
	ctx := context.Background()

	t.Run("list by status success", func(t *testing.T) {
		status := models.StatusPropostaCidadao("pending")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "status_cidadao"}).
			AddRow(uuid.New(), status).
			AddRow(uuid.New(), status)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnRows(rows)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		propostas, total, err := repo.ListByStatus(ctx, status, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, propostas, 2)
	})

	t.Run("database error on find", func(t *testing.T) {
		status := models.StatusPropostaCidadao("pending")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "propostas_mei"`)).
			WillReturnError(assert.AnError)

		propostas, total, err := repo.ListByStatus(ctx, status, 10, 0)
		assert.Error(t, err)
		assert.Nil(t, propostas)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "erro ao listar propostas MEI por status")
	})
}

func TestPropostaMEIRepository_CheckExistingProposta(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewPropostaMEIRepository(db)
	ctx := context.Background()

	t.Run("existing proposta found", func(t *testing.T) {
		oportunidadeID := 1
		meiEmpresaID := "12345678000190"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		exists, err := repo.CheckExistingProposta(ctx, oportunidadeID, meiEmpresaID)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("no existing proposta", func(t *testing.T) {
		oportunidadeID := 1
		meiEmpresaID := "12345678000190"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		exists, err := repo.CheckExistingProposta(ctx, oportunidadeID, meiEmpresaID)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("database error on count", func(t *testing.T) {
		oportunidadeID := 1
		meiEmpresaID := "12345678000190"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "propostas_mei"`)).
			WillReturnError(assert.AnError)

		exists, err := repo.CheckExistingProposta(ctx, oportunidadeID, meiEmpresaID)
		assert.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "erro ao verificar proposta existente")
	})
}

func TestPropostaMEIRepository_UpdateMultipleStatus(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewPropostaMEIRepository(db)
	ctx := context.Background()

	t.Run("update multiple status success", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
		status := models.StatusPropostaCidadao("approved")

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "propostas_mei"`)).
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()

		count, err := repo.UpdateMultipleStatus(ctx, ids, status)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("update multiple status error", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New()}
		status := models.StatusPropostaCidadao("rejected")

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "propostas_mei"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		count, err := repo.UpdateMultipleStatus(ctx, ids, status)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "erro ao atualizar status das propostas")
	})
}
