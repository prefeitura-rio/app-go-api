package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestOportunidadeMEIRepository_Create(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOportunidadeMEIRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		oportunidade := &models.OportunidadeMEI{
			Titulo:           "Oportunidade Teste",
			DescricaoServico: "Descrição do serviço",
			Status:           models.StatusOportunidadeActive,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		id, err := repo.Create(ctx, oportunidade)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		oportunidade := &models.OportunidadeMEI{
			Titulo: "Oportunidade Teste",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "oportunidades_mei"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.Create(ctx, oportunidade)
		assert.Error(t, err)
		assert.Equal(t, 0, id)
		assert.Contains(t, err.Error(), "erro ao criar oportunidade MEI")
	})
}

func TestOportunidadeMEIRepository_GetByID(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOportunidadeMEIRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := 1
		rows := sqlmock.NewRows([]string{"id", "titulo", "descricao_servico", "status"}).
			AddRow(id, "Oportunidade 1", "Descrição", models.StatusOportunidadeActive)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(rows)

		oportunidade, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, oportunidade)
		assert.Equal(t, id, oportunidade.ID)
		assert.Equal(t, "Oportunidade 1", oportunidade.Titulo)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := 999
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		oportunidade, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, oportunidade)
	})

	t.Run("get by id error", func(t *testing.T) {
		id := 1
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnError(assert.AnError)

		oportunidade, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, oportunidade)
		assert.Contains(t, err.Error(), "erro ao buscar oportunidade MEI por ID")
	})
}

func TestOportunidadeMEIRepository_Update(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOportunidadeMEIRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		oportunidade := &models.OportunidadeMEI{
			ID:               1,
			Titulo:           "Oportunidade Atualizada",
			DescricaoServico: "Descrição atualizada",
			Status:           models.StatusOportunidadeActive,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "oportunidades_mei"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, oportunidade)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		oportunidade := &models.OportunidadeMEI{
			ID:     1,
			Titulo: "Atualizado",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "oportunidades_mei"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Update(ctx, oportunidade)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar oportunidade MEI")
	})
}

func TestOportunidadeMEIRepository_Delete(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOportunidadeMEIRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := 1

		// Soft delete uses UPDATE to set deleted_at
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "oportunidades_mei"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		id := 1

		// Soft delete uses UPDATE to set deleted_at
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "oportunidades_mei"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir oportunidade MEI")
	})
}

func TestOportunidadeMEIRepository_List(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOportunidadeMEIRepository(db)
	ctx := context.Background()

	t.Run("list with no filters", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		rows := sqlmock.NewRows([]string{"id", "titulo", "status"}).
			AddRow(1, "Oportunidade 1", models.StatusOportunidadeActive).
			AddRow(2, "Oportunidade 2", models.StatusOportunidadeDraft).
			AddRow(3, "Oportunidade 3", models.StatusOportunidadeActive)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(rows)

		oportunidades, total, err := repo.List(ctx, map[string]interface{}{}, "", 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, oportunidades, 3)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list with titulo filter", func(t *testing.T) {
		titulo := "Jardim"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "titulo"}).
			AddRow(1, "Manutenção de Jardins")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(rows)

		oportunidades, total, err := repo.List(ctx, map[string]interface{}{}, titulo, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, oportunidades, 1)
	})

	t.Run("list with exact filters", func(t *testing.T) {
		filters := map[string]interface{}{
			"status": models.StatusOportunidadeActive,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "status"}).
			AddRow(1, models.StatusOportunidadeActive).
			AddRow(2, models.StatusOportunidadeActive)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(rows)

		oportunidades, total, err := repo.List(ctx, filters, "", 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, oportunidades, 2)
	})
}

func TestOportunidadeMEIRepository_ListByStatus(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOportunidadeMEIRepository(db)
	ctx := context.Background()

	t.Run("list by status success", func(t *testing.T) {
		status := models.StatusOportunidadeActive

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "titulo", "status"}).
			AddRow(1, "Oportunidade 1", status).
			AddRow(2, "Oportunidade 2", status)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(rows)

		oportunidades, total, err := repo.ListByStatus(ctx, status, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, oportunidades, 2)
	})
}

func TestOportunidadeMEIRepository_ListByOrgao(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOportunidadeMEIRepository(db)
	ctx := context.Background()

	t.Run("list by orgao success", func(t *testing.T) {
		orgaoID := "ORG001"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "oportunidades_mei"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "titulo", "orgao_id"}).
			AddRow(1, "Oportunidade 1", orgaoID).
			AddRow(2, "Oportunidade 2", orgaoID)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "oportunidades_mei"`)).
			WillReturnRows(rows)

		oportunidades, total, err := repo.ListByOrgao(ctx, orgaoID, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, oportunidades, 2)
	})

	t.Run("list by orgao error", func(t *testing.T) {
		orgaoID := "ORG001"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "oportunidades_mei"`)).
			WillReturnError(assert.AnError)

		oportunidades, total, err := repo.ListByOrgao(ctx, orgaoID, 10, 0)
		assert.Error(t, err)
		assert.Equal(t, 0, total)
		assert.Nil(t, oportunidades)
		assert.Contains(t, err.Error(), "erro ao listar oportunidades MEI por órgão")
	})
}

func TestOportunidadeMEIRepository_UpdateExpiredOpportunities(t *testing.T) {
	db, mock, cleanup := SetupMockDB(t)
	defer cleanup()

	repo := NewOportunidadeMEIRepository(db)
	ctx := context.Background()

	t.Run("update expired opportunities success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "oportunidades_mei"`)).
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()

		err := repo.UpdateExpiredOpportunities(ctx)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update expired opportunities error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "oportunidades_mei"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateExpiredOpportunities(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar oportunidades expiradas")
	})

	t.Run("update expired opportunities no rows affected", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "oportunidades_mei"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := repo.UpdateExpiredOpportunities(ctx)
		assert.NoError(t, err)
	})
}
