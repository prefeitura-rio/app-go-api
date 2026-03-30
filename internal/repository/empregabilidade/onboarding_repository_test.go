package empregabilidade

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestOnboardingRepository_Upsert(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewOnboardingRepository(db)
	ctx := context.Background()

	entity := &empregabilidade.Onboarding{
		CPF:                         "12345678900",
		IsEmpregabilidadeFirstLogin: true,
	}

	mock.ExpectBegin()
	// GORM Save expects UPDATE (upsert behavior)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_onboarding"`)).
		WithArgs(entity.IsEmpregabilidadeFirstLogin, sqlmock.AnyArg(), sqlmock.AnyArg(), entity.CPF).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Upsert(ctx, entity)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOnboardingRepository_GetByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewOnboardingRepository(db)
	ctx := context.Background()

	cpf := "12345678900"
	expectedEntity := &empregabilidade.Onboarding{
		CPF:                         cpf,
		IsEmpregabilidadeFirstLogin: false,
	}

	rows := sqlmock.NewRows([]string{"cpf", "is_empregabilidade_first_login"}).
		AddRow(expectedEntity.CPF, expectedEntity.IsEmpregabilidadeFirstLogin)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_onboarding"`)).
		WithArgs(cpf, 1).
		WillReturnRows(rows)

	result, err := repo.GetByCPF(ctx, cpf)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedEntity.CPF, result.CPF)
	assert.Equal(t, expectedEntity.IsEmpregabilidadeFirstLogin, result.IsEmpregabilidadeFirstLogin)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOnboardingRepository_GetByCPF_NotFound(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewOnboardingRepository(db)
	ctx := context.Background()

	cpf := "12345678900"

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_onboarding"`)).
		WithArgs(cpf, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetByCPF(ctx, cpf)

	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOnboardingRepository_MarkFirstLoginCompleted(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewOnboardingRepository(db)
	ctx := context.Background()

	cpf := "12345678900"

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO emp_onboarding`)).
		WithArgs(cpf).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.MarkFirstLoginCompleted(ctx, cpf)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
