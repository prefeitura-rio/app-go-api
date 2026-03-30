package empregabilidade

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestTermosUsoRepository_Upsert(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewTermosUsoRepository(db)
	ctx := context.Background()

	now := time.Now()
	entity := &empregabilidade.TermosUso{
		CPF:         "12345678900",
		UserConsent: true,
		AceitoEm:    &now,
	}

	mock.ExpectBegin()
	// GORM Save expects UPDATE (upsert behavior)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_termos_uso"`)).
		WithArgs(entity.UserConsent, entity.AceitoEm, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.CPF).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Upsert(ctx, entity)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTermosUsoRepository_GetByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewTermosUsoRepository(db)
	ctx := context.Background()

	cpf := "12345678900"
	now := time.Now()
	expectedEntity := &empregabilidade.TermosUso{
		CPF:         cpf,
		UserConsent: true,
		AceitoEm:    &now,
	}

	rows := sqlmock.NewRows([]string{"cpf", "user_consent", "aceito_em"}).
		AddRow(expectedEntity.CPF, expectedEntity.UserConsent, expectedEntity.AceitoEm)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_termos_uso"`)).
		WithArgs(cpf, 1).
		WillReturnRows(rows)

	result, err := repo.GetByCPF(ctx, cpf)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedEntity.CPF, result.CPF)
	assert.Equal(t, expectedEntity.UserConsent, result.UserConsent)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTermosUsoRepository_GetByCPF_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewTermosUsoRepository(db)
	ctx := context.Background()

	cpf := "12345678900"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_termos_uso"`)).
		WithArgs(cpf, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetByCPF(ctx, cpf)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTermosUsoRepository_AcceptTerms(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewTermosUsoRepository(db)
	ctx := context.Background()

	cpf := "12345678900"

	mock.ExpectBegin()
	// AcceptTerms uses INSERT with ON CONFLICT DO UPDATE
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "emp_termos_uso"`)).
		WithArgs(cpf, true, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.AcceptTerms(ctx, cpf)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
