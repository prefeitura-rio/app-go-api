package empregabilidade

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

func novaVagaParaHabilidades() *empregabilidade.Vaga {
	return &empregabilidade.Vaga{
		ID:                  uuid.New(),
		Titulo:              "Desenvolvedor Go",
		Descricao:           "Vaga para desenvolvedor Go",
		IDContratante:       "12345678000190",
		IDRegimeContratacao: uuid.New(),
		IDModeloTrabalho:    uuid.New(),
		Status:              empregabilidade.StatusVagaEmEdicao,
	}
}

func areasDaVaga(areas ...empregabilidade.VagaAreaAtuacao) *[]empregabilidade.VagaAreaAtuacao {
	return &areas
}

func idsDaVaga(ids ...int64) *[]int64 {
	return &ids
}

// Criar a vaga com áreas, conhecimentos e comportamentos grava tudo na mesma
// transação, ligado ao id da vaga, com as áreas antes dos conhecimentos (FK
// composta).
func TestVagaRepository_Create_GravaHabilidadesCompetenciasNaMesmaTransacao(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	vaga := novaVagaParaHabilidades()
	vaga.ID = uuid.Nil
	vaga.AreasAtuacao = areasDaVaga(
		empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 3, ConhecimentoIDs: []int64{10, 11}},
		empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 4},
	)
	vaga.ComportamentoAtitudesIDs = idsDaVaga(7)
	novoID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(novoID))
	mock.ExpectExec(`INSERT INTO emp_vagas_areas_atuacao \(id_vaga, id_area_atuacao\) VALUES`).
		WithArgs(novoID, int64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO emp_vagas_areas_atuacao \(id_vaga, id_area_atuacao\) VALUES`).
		WithArgs(novoID, int64(4)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO emp_vagas_conhecimentos \(id_vaga, id_area_atuacao, id_habilidade\) VALUES`).
		WithArgs(novoID, int64(3), int64(10)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO emp_vagas_conhecimentos \(id_vaga, id_area_atuacao, id_habilidade\) VALUES`).
		WithArgs(novoID, int64(3), int64(11)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO emp_vagas_comportamento_atitudes \(id_vaga, id_comportamento_atitudes\) VALUES`).
		WithArgs(novoID, int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	id, err := repo.Create(context.Background(), vaga)
	require.NoError(t, err)
	assert.Equal(t, novoID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Falha ao gravar um conhecimento desfaz a vaga inteira.
func TestVagaRepository_Create_ErroNosConhecimentosDesfazTudo(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	vaga := novaVagaParaHabilidades()
	vaga.ID = uuid.Nil
	vaga.AreasAtuacao = areasDaVaga(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 3, ConhecimentoIDs: []int64{10}})
	novoID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "emp_vagas"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(novoID))
	mock.ExpectExec(`INSERT INTO emp_vagas_areas_atuacao`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO emp_vagas_conhecimentos`).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	_, err := repo.Create(context.Background(), vaga)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao gravar conhecimento da vaga")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Editar com as listas enviadas substitui o conjunto anterior: apaga o que
// estava gravado e insere o que veio.
func TestVagaRepository_UpdateWithAssociations_SubstituiHabilidadesCompetencias(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	vaga := novaVagaParaHabilidades()
	vaga.AreasAtuacao = areasDaVaga(empregabilidade.VagaAreaAtuacao{IDAreaAtuacao: 5, ConhecimentoIDs: []int64{20}})
	vaga.ComportamentoAtitudesIDs = idsDaVaga(8, 9)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM emp_vagas_conhecimentos WHERE id_vaga = \$1`).
		WithArgs(vaga.ID).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`DELETE FROM emp_vagas_areas_atuacao WHERE id_vaga = \$1`).
		WithArgs(vaga.ID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO emp_vagas_areas_atuacao`).
		WithArgs(vaga.ID, int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO emp_vagas_conhecimentos`).
		WithArgs(vaga.ID, int64(5), int64(20)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM emp_vagas_comportamento_atitudes WHERE id_vaga = \$1`).
		WithArgs(vaga.ID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO emp_vagas_comportamento_atitudes`).
		WithArgs(vaga.ID, int64(8)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO emp_vagas_comportamento_atitudes`).
		WithArgs(vaga.ID, int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.UpdateWithAssociations(context.Background(), vaga))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Listas vazias apagam tudo e não inserem nada.
func TestVagaRepository_UpdateWithAssociations_ListasVaziasApagamTudo(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	vaga := novaVagaParaHabilidades()
	vaga.AreasAtuacao = areasDaVaga()
	vaga.ComportamentoAtitudesIDs = idsDaVaga()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM emp_vagas_conhecimentos`).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(`DELETE FROM emp_vagas_areas_atuacao`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM emp_vagas_comportamento_atitudes`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.UpdateWithAssociations(context.Background(), vaga))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Sem as listas no corpo (nil), nada é apagado nem inserido: o gravado fica.
func TestVagaRepository_UpdateWithAssociations_SemListasNaoTocaHabilidades(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	vaga := novaVagaParaHabilidades()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.UpdateWithAssociations(context.Background(), vaga))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_UpdateWithAssociations_ErroAoApagarAreasDesfazTudo(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	vaga := novaVagaParaHabilidades()
	vaga.AreasAtuacao = areasDaVaga()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "emp_vagas"`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM emp_vagas_conhecimentos`).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.UpdateWithAssociations(context.Background(), vaga)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao remover conhecimentos da vaga")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// A leitura monta as áreas com os conhecimentos de cada uma e os
// comportamentos, só de vaga não excluída.
func TestVagaRepository_GetHabilidadesCompetencias_MontaAreasComConhecimentos(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	vagaID := uuid.New()

	mock.ExpectQuery(`(?s)FROM emp_vagas_areas_atuacao va.*JOIN emp_areas_atuacao a ON a.id = va.id_area_atuacao.*JOIN emp_vagas v ON v.id = va.id_vaga AND v.deleted_at IS NULL.*WHERE va.id_vaga = \$1`).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).AddRow(4, "Saúde").AddRow(3, "Tecnologia"))
	mock.ExpectQuery(`(?s)FROM emp_vagas_conhecimentos vc.*JOIN emp_habilidades h ON h.id = vc.id_habilidade.*JOIN emp_vagas v ON v.id = vc.id_vaga AND v.deleted_at IS NULL.*WHERE vc.id_vaga = \$1`).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"id_area_atuacao", "id", "nome"}).
			AddRow(3, 10, "Go").AddRow(3, 11, "SQL"))
	mock.ExpectQuery(`(?s)FROM emp_vagas_comportamento_atitudes vca.*JOIN emp_comportamento_atitudes c ON c.id = vca.id_comportamento_atitudes.*JOIN emp_vagas v ON v.id = vca.id_vaga AND v.deleted_at IS NULL.*WHERE vca.id_vaga = \$1`).
		WithArgs(vagaID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}).AddRow(7, "Proatividade"))

	got, err := repo.GetHabilidadesCompetencias(context.Background(), vagaID)
	require.NoError(t, err)
	assert.Equal(t, []empregabilidade.VagaAreaAtuacao{
		{IDAreaAtuacao: 4, Nome: "Saúde", ConhecimentoIDs: []int64{}, Conhecimentos: []empregabilidade.CatalogoItem{}},
		{IDAreaAtuacao: 3, Nome: "Tecnologia", ConhecimentoIDs: []int64{10, 11}, Conhecimentos: []empregabilidade.CatalogoItem{{ID: 10, Nome: "Go"}, {ID: 11, Nome: "SQL"}}},
	}, got.AreasAtuacao)
	assert.Equal(t, []empregabilidade.CatalogoItem{{ID: 7, Nome: "Proatividade"}}, got.ComportamentosAtitudes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Vaga antiga, sem registro: listas vazias, sem erro.
func TestVagaRepository_GetHabilidadesCompetencias_VagaSemRegistrosLeVazio(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	vagaID := uuid.New()

	mock.ExpectQuery(`FROM emp_vagas_areas_atuacao`).WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}))
	mock.ExpectQuery(`FROM emp_vagas_conhecimentos`).WillReturnRows(sqlmock.NewRows([]string{"id_area_atuacao", "id", "nome"}))
	mock.ExpectQuery(`FROM emp_vagas_comportamento_atitudes`).WillReturnRows(sqlmock.NewRows([]string{"id", "nome"}))

	got, err := repo.GetHabilidadesCompetencias(context.Background(), vagaID)
	require.NoError(t, err)
	assert.Equal(t, []empregabilidade.VagaAreaAtuacao{}, got.AreasAtuacao)
	assert.Equal(t, []empregabilidade.CatalogoItem{}, got.ComportamentosAtitudes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVagaRepository_GetHabilidadesCompetencias_ErroDoBanco(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewVagaRepository(db)
	mock.ExpectQuery(`FROM emp_vagas_areas_atuacao`).WillReturnError(assert.AnError)

	_, err := repo.GetHabilidadesCompetencias(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "erro ao buscar áreas de atuação da vaga")
}

func TestVagaRepository_CatalogoDeHabilidades(t *testing.T) {
	t.Run("áreas existentes", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectQuery(`SELECT id FROM emp_areas_atuacao WHERE id IN \(\$1,\$2\)`).
			WithArgs(int64(3), int64(99)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

		got, err := NewVagaRepository(db).AreasAtuacaoExistentes(context.Background(), []int64{3, 99})
		require.NoError(t, err)
		assert.Equal(t, []int64{3}, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("pares área x conhecimento do catálogo", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectQuery(`SELECT id_area_atuacao, id_habilidade FROM area_atuacao_habilidade WHERE id_area_atuacao IN \(\$1\) AND id_habilidade IN \(\$2,\$3\)`).
			WithArgs(int64(3), int64(10), int64(50)).
			WillReturnRows(sqlmock.NewRows([]string{"id_area_atuacao", "id_habilidade"}).AddRow(3, 10))

		got, err := NewVagaRepository(db).ConhecimentosDasAreas(context.Background(), []int64{3}, []int64{10, 50})
		require.NoError(t, err)
		assert.True(t, got[3][10])
		assert.False(t, got[3][50])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("comportamentos existentes", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		mock.ExpectQuery(`SELECT id FROM emp_comportamento_atitudes WHERE id IN \(\$1\)`).
			WithArgs(int64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

		got, err := NewVagaRepository(db).ComportamentosExistentes(context.Background(), []int64{7})
		require.NoError(t, err)
		assert.Equal(t, []int64{7}, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("listas vazias não consultam o banco", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewVagaRepository(db)

		areas, err := repo.AreasAtuacaoExistentes(context.Background(), nil)
		require.NoError(t, err)
		assert.Empty(t, areas)
		pares, err := repo.ConhecimentosDasAreas(context.Background(), []int64{3}, nil)
		require.NoError(t, err)
		assert.Empty(t, pares)
		comportamentos, err := repo.ComportamentosExistentes(context.Background(), nil)
		require.NoError(t, err)
		assert.Empty(t, comportamentos)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("erro do banco", func(t *testing.T) {
		db, mock, cleanup := repository.SetupMockDB(t)
		defer cleanup()
		repo := NewVagaRepository(db)
		mock.ExpectQuery(`FROM emp_areas_atuacao`).WillReturnError(assert.AnError)
		mock.ExpectQuery(`FROM area_atuacao_habilidade`).WillReturnError(assert.AnError)
		mock.ExpectQuery(`FROM emp_comportamento_atitudes`).WillReturnError(assert.AnError)

		_, err := repo.AreasAtuacaoExistentes(context.Background(), []int64{3})
		assert.Error(t, err)
		_, err = repo.ConhecimentosDasAreas(context.Background(), []int64{3}, []int64{10})
		assert.Error(t, err)
		_, err = repo.ComportamentosExistentes(context.Background(), []int64{7})
		assert.Error(t, err)
	})
}
