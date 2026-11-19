package empregabilidade

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// tipo_etapa (DEV-532): gravado quando informado; ausente mantém o valor do banco.

func TestEtapaRepository_Update_GravaTipoEtapaInformado(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEtapaRepository(db)
	tipo := "entrevista"
	entity := &empregabilidade.Etapa{ID: uuid.New(), IDVaga: uuid.New(), Titulo: "Entrevista", Descricao: "RH", Ordem: 2, TipoEtapa: &tipo}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_etapas" SET "titulo"=\$1,"descricao"=\$2,"ordem"=\$3,"tipo_etapa"=\$4,"updated_at"=\$5 WHERE id = \$6 AND id_vaga = \$7$`).
		WithArgs(entity.Titulo, entity.Descricao, entity.Ordem, tipo, sqlmock.AnyArg(), entity.ID, entity.IDVaga).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.Update(context.Background(), entity))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestEtapaRepository_Create_GravaTipoEtapa(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewEtapaRepository(db)
	tipo := "triagem"
	entity := &empregabilidade.Etapa{IDVaga: uuid.New(), Titulo: "Triagem", Ordem: 1, TipoEtapa: &tipo}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_etapas" ("id_vaga","titulo","descricao","ordem","tipo_etapa","created_at","updated_at")`)).
		WithArgs(entity.IDVaga, entity.Titulo, "", 1, tipo, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	_, err := repo.Create(context.Background(), entity)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCamposAtualizaveisEtapa(t *testing.T) {
	tipo := "entrevista"
	assert.Equal(t, []string{"titulo", "descricao", "ordem"}, camposAtualizaveisEtapa(&empregabilidade.Etapa{}))
	assert.Equal(t, []string{"titulo", "descricao", "ordem", "tipo_etapa", "updated_at"}, camposAtualizaveisEtapa(&empregabilidade.Etapa{TipoEtapa: &tipo}, "updated_at"))
}
