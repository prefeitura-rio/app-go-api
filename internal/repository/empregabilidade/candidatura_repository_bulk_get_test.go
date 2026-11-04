package empregabilidade

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// DEV-238: a aprovação/reprovação em lote envia e-mail que lê candidatura.Vaga;
// BulkGetByCPFs precisa carregar a Vaga (e o Contratante) para o e-mail não chegar
// ao worker com vaga nula.
func TestCandidaturaRepository_BulkGetByCPFs_PreloadsVagaAndContratante(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCandidaturaRepository(db)
	vagaID := uuid.New()
	candidaturaID := uuid.New()
	cnpj := "11222333000181"

	mock.ExpectQuery(`SELECT \* FROM "emp_candidaturas" WHERE \(id_vaga = \$1 AND cpf IN \(\$2\)\)`).
		WithArgs(vagaID, "00000000191").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cpf", "id_vaga", "status"}).
			AddRow(candidaturaID, "00000000191", vagaID, "candidatura_enviada"))

	mock.ExpectQuery(`SELECT \* FROM "emp_vagas" WHERE "emp_vagas"."id" = \$1`).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "titulo", "id_contratante"}).
			AddRow(vagaID, "Vaga Sintética", cnpj))

	mock.ExpectQuery(`SELECT \* FROM "emp_empresas" WHERE "emp_empresas"."cnpj" = \$1`).
		WithArgs(cnpj).
		WillReturnRows(sqlmock.NewRows([]string{"cnpj", "razao_social", "nome_fantasia"}).
			AddRow(cnpj, "Empresa Sintética LTDA", "Empresa Sintética"))

	got, err := repo.BulkGetByCPFs(context.Background(), vagaID, []string{"00000000191"})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.NotNil(t, got[0].Vaga, "Vaga deve vir carregada")
	assert.Equal(t, "Vaga Sintética", got[0].Vaga.Titulo)
	require.NotNil(t, got[0].Vaga.Contratante, "Contratante deve vir carregado")
	assert.Equal(t, "Empresa Sintética", got[0].Vaga.Contratante.NomeFantasia)
	assert.NoError(t, mock.ExpectationsWereMet())
}
