package empregabilidade

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestInformacaoComplementarRepository_Create(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	entity := &empregabilidade.InformacaoComplementar{
		ID:          uuid.New(),
		IDVaga:      vagaID,
		Titulo:      "Possui veículo próprio?",
		Obrigatorio: true,
		TipoCampo:   empregabilidade.TipoCampoRespostaCurta,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_informacoes_complementares"`)).
		WithArgs(
			entity.IDVaga,
			entity.Titulo,
			entity.Obrigatorio,
			entity.TipoCampo,
			sqlmock.AnyArg(), // ValorMinimo
			sqlmock.AnyArg(), // ValorMaximo
			sqlmock.AnyArg(), // Opcoes
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			entity.ID,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(entity.ID))
	mock.ExpectCommit()

	id, err := repo.Create(ctx, entity)

	assert.NoError(t, err)
	assert.Equal(t, entity.ID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_Create_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.InformacaoComplementar{
		ID:          uuid.New(),
		IDVaga:      uuid.New(),
		Titulo:      "Experiência na área",
		Obrigatorio: false,
		TipoCampo:   empregabilidade.TipoCampoSelecaoUnica,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_informacoes_complementares"`)).
		WithArgs(
			entity.IDVaga,
			entity.Titulo,
			entity.Obrigatorio,
			entity.TipoCampo,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			entity.ID,
		).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	id, err := repo.Create(ctx, entity)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "erro ao criar informação complementar")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_GetByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	id := uuid.New()
	vagaID := uuid.New()
	expectedEntity := &empregabilidade.InformacaoComplementar{
		ID:          id,
		IDVaga:      vagaID,
		Titulo:      "CNH categoria B?",
		Obrigatorio: true,
		TipoCampo:   empregabilidade.TipoCampoRespostaCurta,
	}

	rows := sqlmock.NewRows([]string{"id", "id_vaga", "titulo", "obrigatorio", "tipo_campo"}).
		AddRow(expectedEntity.ID, expectedEntity.IDVaga, expectedEntity.Titulo, expectedEntity.Obrigatorio, expectedEntity.TipoCampo)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_informacoes_complementares"`)).
		WithArgs(id, 1).
		WillReturnRows(rows)

	result, err := repo.GetByID(ctx, id)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedEntity.ID, result.ID)
	assert.Equal(t, expectedEntity.IDVaga, result.IDVaga)
	assert.Equal(t, expectedEntity.Titulo, result.Titulo)
	assert.Equal(t, expectedEntity.Obrigatorio, result.Obrigatorio)
	assert.Equal(t, expectedEntity.TipoCampo, result.TipoCampo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_GetByID_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_informacoes_complementares"`)).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetByID(ctx, id)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_GetByID_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_informacoes_complementares"`)).
		WithArgs(id, 1).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetByID(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "erro ao buscar informação complementar")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_Update(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.InformacaoComplementar{
		ID:          uuid.New(),
		IDVaga:      uuid.New(),
		Titulo:      "Disponibilidade de viagem?",
		Obrigatorio: false,
		TipoCampo:   empregabilidade.TipoCampoRespostaCurta,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_informacoes_complementares"`)).
		WithArgs(
			entity.IDVaga,
			entity.Titulo,
			entity.Obrigatorio,
			entity.TipoCampo,
			sqlmock.AnyArg(), // ValorMinimo
			sqlmock.AnyArg(), // ValorMaximo
			sqlmock.AnyArg(), // Opcoes
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			entity.ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, entity)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_Update_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.InformacaoComplementar{
		ID:          uuid.New(),
		IDVaga:      uuid.New(),
		Titulo:      "Idiomas",
		Obrigatorio: true,
		TipoCampo:   empregabilidade.TipoCampoSelecaoUnica,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_informacoes_complementares"`)).
		WithArgs(
			entity.IDVaga,
			entity.Titulo,
			entity.Obrigatorio,
			entity.TipoCampo,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			entity.ID,
		).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.Update(ctx, entity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao atualizar informação complementar")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_Delete(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_informacoes_complementares"`)).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, id)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_Delete_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_informacoes_complementares"`)).
		WithArgs(id).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.Delete(ctx, id)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir informação complementar")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_ListByVaga(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	expectedEntities := []*empregabilidade.InformacaoComplementar{
		{
			ID:          uuid.New(),
			IDVaga:      vagaID,
			Titulo:      "CNH",
			Obrigatorio: true,
			TipoCampo:   empregabilidade.TipoCampoRespostaCurta,
		},
		{
			ID:          uuid.New(),
			IDVaga:      vagaID,
			Titulo:      "Disponibilidade",
			Obrigatorio: false,
			TipoCampo:   empregabilidade.TipoCampoSelecaoUnica,
		},
	}

	rows := sqlmock.NewRows([]string{"id", "id_vaga", "titulo", "obrigatorio", "tipo_campo"})
	for _, e := range expectedEntities {
		rows.AddRow(e.ID, e.IDVaga, e.Titulo, e.Obrigatorio, e.TipoCampo)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_informacoes_complementares" WHERE id_vaga = $1 ORDER BY created_at ASC`)).
		WithArgs(vagaID).
		WillReturnRows(rows)

	results, err := repo.ListByVaga(ctx, vagaID)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, expectedEntities[0].Titulo, results[0].Titulo)
	assert.Equal(t, expectedEntities[1].Titulo, results[1].Titulo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_ListByVaga_Empty(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_informacoes_complementares" WHERE id_vaga = $1 ORDER BY created_at ASC`)).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "id_vaga", "titulo", "obrigatorio", "tipo_campo"}))

	results, err := repo.ListByVaga(ctx, vagaID)

	assert.NoError(t, err)
	assert.Len(t, results, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_ListByVaga_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_informacoes_complementares" WHERE id_vaga = $1 ORDER BY created_at ASC`)).
		WithArgs(vagaID).
		WillReturnError(errors.New("database error"))

	results, err := repo.ListByVaga(ctx, vagaID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao listar informações complementares")
	assert.Nil(t, results)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_DeleteByVaga(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_informacoes_complementares" WHERE id_vaga = $1`)).
		WithArgs(vagaID).
		WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	err := repo.DeleteByVaga(ctx, vagaID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_DeleteByVaga_Error(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_informacoes_complementares" WHERE id_vaga = $1`)).
		WithArgs(vagaID).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.DeleteByVaga(ctx, vagaID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao excluir informações complementares da vaga")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_CreateWithLimitCheck(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	entity := &empregabilidade.InformacaoComplementar{
		ID:          uuid.New(),
		IDVaga:      vagaID,
		Titulo:      "Teste",
		Obrigatorio: true,
		TipoCampo:   empregabilidade.TipoCampoRespostaCurta,
	}
	maxLimit := 5

	mock.ExpectBegin()

	// Lock query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM emp_vagas WHERE id = $1 FOR UPDATE`)).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(vagaID))

	// Count query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_informacoes_complementares" WHERE id_vaga = $1`)).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Create query
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_informacoes_complementares"`)).
		WithArgs(
			entity.IDVaga,
			entity.Titulo,
			entity.Obrigatorio,
			entity.TipoCampo,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			entity.ID,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(entity.ID))

	mock.ExpectCommit()

	id, err := repo.CreateWithLimitCheck(ctx, entity, maxLimit)

	assert.NoError(t, err)
	assert.Equal(t, entity.ID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_CreateWithLimitCheck_LimitReached(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	entity := &empregabilidade.InformacaoComplementar{
		ID:          uuid.New(),
		IDVaga:      vagaID,
		Titulo:      "Teste",
		Obrigatorio: true,
		TipoCampo:   empregabilidade.TipoCampoRespostaCurta,
	}
	maxLimit := 5

	mock.ExpectBegin()

	// Lock query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM emp_vagas WHERE id = $1 FOR UPDATE`)).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(vagaID))

	// Count query - returns 5 (limit reached)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_informacoes_complementares" WHERE id_vaga = $1`)).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectRollback()

	id, err := repo.CreateWithLimitCheck(ctx, entity, maxLimit)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "limite máximo de 5 informações complementares por vaga atingido")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_CreateWithLimitCheck_LockError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	entity := &empregabilidade.InformacaoComplementar{
		ID:          uuid.New(),
		IDVaga:      vagaID,
		Titulo:      "Teste",
		Obrigatorio: true,
		TipoCampo:   empregabilidade.TipoCampoRespostaCurta,
	}
	maxLimit := 5

	mock.ExpectBegin()

	// Lock query fails
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM emp_vagas WHERE id = $1 FOR UPDATE`)).
		WithArgs(vagaID).
		WillReturnError(errors.New("lock error"))

	mock.ExpectRollback()

	id, err := repo.CreateWithLimitCheck(ctx, entity, maxLimit)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "erro ao obter lock na vaga")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_CreateWithLimitCheck_CountError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	entity := &empregabilidade.InformacaoComplementar{
		ID:          uuid.New(),
		IDVaga:      vagaID,
		Titulo:      "Teste",
		Obrigatorio: true,
		TipoCampo:   empregabilidade.TipoCampoRespostaCurta,
	}
	maxLimit := 5

	mock.ExpectBegin()

	// Lock query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM emp_vagas WHERE id = $1 FOR UPDATE`)).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(vagaID))

	// Count query fails
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_informacoes_complementares" WHERE id_vaga = $1`)).
		WithArgs(vagaID).
		WillReturnError(errors.New("count error"))

	mock.ExpectRollback()

	id, err := repo.CreateWithLimitCheck(ctx, entity, maxLimit)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "erro ao contar informações complementares")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInformacaoComplementarRepository_CreateWithLimitCheck_CreateError(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewInformacaoComplementarRepository(db)
	ctx := context.Background()

	vagaID := uuid.New()
	entity := &empregabilidade.InformacaoComplementar{
		ID:          uuid.New(),
		IDVaga:      vagaID,
		Titulo:      "Teste",
		Obrigatorio: true,
		TipoCampo:   empregabilidade.TipoCampoRespostaCurta,
	}
	maxLimit := 5

	mock.ExpectBegin()

	// Lock query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM emp_vagas WHERE id = $1 FOR UPDATE`)).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(vagaID))

	// Count query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "emp_informacoes_complementares" WHERE id_vaga = $1`)).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Create query fails
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_informacoes_complementares"`)).
		WithArgs(
			entity.IDVaga,
			entity.Titulo,
			entity.Obrigatorio,
			entity.TipoCampo,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			entity.ID,
		).
		WillReturnError(errors.New("insert error"))

	mock.ExpectRollback()

	id, err := repo.CreateWithLimitCheck(ctx, entity, maxLimit)

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, id)
	assert.Contains(t, err.Error(), "erro ao criar informação complementar")
	assert.NoError(t, mock.ExpectationsWereMet())
}
