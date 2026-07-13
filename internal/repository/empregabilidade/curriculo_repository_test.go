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

// Formação Acadêmica Tests

func TestCurriculoRepository_CreateFormacao(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		formacao := &empregabilidade.CurriculoFormacao{
			ID:              uuid.New(),
			CPF:             "12345678900",
			NomeInstituicao: "Universidade Federal",
			NomeCurso:       "Engenharia de Software",
			Status:          empregabilidade.StatusFormacaoCompleto,
			AnoConclusao:    "2021",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_formacoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(formacao.ID))
		mock.ExpectCommit()

		id, err := repo.CreateFormacao(ctx, formacao)
		assert.NoError(t, err)
		assert.Equal(t, formacao.ID, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		formacao := &empregabilidade.CurriculoFormacao{
			CPF:       "12345678900",
			NomeCurso: "Engenharia",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_formacoes"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.CreateFormacao(ctx, formacao)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar formação")
		assert.Equal(t, uuid.Nil, id)
	})
}

func TestCurriculoRepository_GetFormacaoByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		escolaridadeID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "cpf", "nome_instituicao", "nome_curso", "id_escolaridade"}).
			AddRow(id, "12345678900", "Universidade Federal", "Engenharia", escolaridadeID)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_formacoes"`)).
			WillReturnRows(rows)

		// Expect Escolaridade preload query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		formacao, err := repo.GetFormacaoByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, formacao)
		assert.Equal(t, id, formacao.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_formacoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		formacao, err := repo.GetFormacaoByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, formacao)
	})

	t.Run("get by id error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_formacoes"`)).
			WillReturnError(assert.AnError)

		formacao, err := repo.GetFormacaoByID(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao buscar formação")
		assert.Nil(t, formacao)
	})
}

func TestCurriculoRepository_UpdateFormacao(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		formacao := &empregabilidade.CurriculoFormacao{
			ID:              uuid.New(),
			CPF:             "12345678900",
			NomeInstituicao: "Universidade Federal",
			NomeCurso:       "Engenharia Atualizado",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_formacoes"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateFormacao(ctx, formacao)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		formacao := &empregabilidade.CurriculoFormacao{
			ID:  uuid.New(),
			CPF: "12345678900",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_formacoes"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateFormacao(ctx, formacao)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar formação")
	})
}

func TestCurriculoRepository_DeleteFormacao(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteFormacao(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.DeleteFormacao(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir formação")
	})
}

func TestCurriculoRepository_ListFormacoesByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("list success", func(t *testing.T) {
		cpf := "12345678900"
		escID1 := uuid.New()
		escID2 := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "cpf", "nome_curso", "ano_conclusao", "id_escolaridade"}).
			AddRow(uuid.New(), cpf, "Engenharia", "2021", escID1).
			AddRow(uuid.New(), cpf, "Matemática", "2018", escID2)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_formacoes"`)).
			WillReturnRows(rows)

		// Expect Escolaridade preload
		escRows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(escID1, "Superior").
			AddRow(escID2, "Superior")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades"`)).
			WillReturnRows(escRows)

		formacoes, err := repo.ListFormacoesByCPF(ctx, cpf)
		assert.NoError(t, err)
		assert.Len(t, formacoes, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list empty", func(t *testing.T) {
		cpf := "12345678900"
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_formacoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "cpf"}))

		// Even with empty results, GORM may try to preload
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_escolaridades"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		formacoes, err := repo.ListFormacoesByCPF(ctx, cpf)
		assert.NoError(t, err)
		assert.Len(t, formacoes, 0)
	})

	t.Run("list error", func(t *testing.T) {
		cpf := "12345678900"
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_formacoes"`)).
			WillReturnError(assert.AnError)

		formacoes, err := repo.ListFormacoesByCPF(ctx, cpf)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao listar formações")
		assert.Nil(t, formacoes)
	})
}

// Idiomas Tests

func TestCurriculoRepository_CreateIdioma(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		idioma := &empregabilidade.CurriculoIdioma{
			ID:       uuid.New(),
			CPF:      "12345678900",
			IDIdioma: uuid.New(),
			IDNivel:  uuid.New(),
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_idiomas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(idioma.ID))
		mock.ExpectCommit()

		id, err := repo.CreateIdioma(ctx, idioma)
		assert.NoError(t, err)
		assert.Equal(t, idioma.ID, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		idioma := &empregabilidade.CurriculoIdioma{
			CPF: "12345678900",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_idiomas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.CreateIdioma(ctx, idioma)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar idioma")
		assert.Equal(t, uuid.Nil, id)
	})
}

func TestCurriculoRepository_GetIdiomaByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "cpf", "id_idioma", "id_nivel"}).
			AddRow(id, "12345678900", uuid.New(), uuid.New())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_idiomas"`)).
			WillReturnRows(rows)

		// Expect Idioma and Nivel preload queries
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_idiomas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_niveis_idioma"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		idioma, err := repo.GetIdiomaByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, idioma)
		assert.Equal(t, id, idioma.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_idiomas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		idioma, err := repo.GetIdiomaByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, idioma)
	})

	t.Run("get by id error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_idiomas"`)).
			WillReturnError(assert.AnError)

		idioma, err := repo.GetIdiomaByID(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao buscar idioma")
		assert.Nil(t, idioma)
	})
}

func TestCurriculoRepository_UpdateIdioma(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		idioma := &empregabilidade.CurriculoIdioma{
			ID:       uuid.New(),
			CPF:      "12345678900",
			IDIdioma: uuid.New(),
			IDNivel:  uuid.New(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_idiomas"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateIdioma(ctx, idioma)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		idioma := &empregabilidade.CurriculoIdioma{
			ID:  uuid.New(),
			CPF: "12345678900",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_idiomas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateIdioma(ctx, idioma)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar idioma")
	})
}

func TestCurriculoRepository_DeleteIdioma(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_idiomas"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteIdioma(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_idiomas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.DeleteIdioma(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir idioma")
	})
}

func TestCurriculoRepository_ListIdiomasByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("list success", func(t *testing.T) {
		cpf := "12345678900"
		rows := sqlmock.NewRows([]string{"id", "cpf", "id_idioma", "id_nivel"}).
			AddRow(uuid.New(), cpf, uuid.New(), uuid.New()).
			AddRow(uuid.New(), cpf, uuid.New(), uuid.New())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_idiomas"`)).
			WillReturnRows(rows)

		// Expect preloads
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_idiomas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_niveis_idioma"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		idiomas, err := repo.ListIdiomasByCPF(ctx, cpf)
		assert.NoError(t, err)
		assert.Len(t, idiomas, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list error", func(t *testing.T) {
		cpf := "12345678900"
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_idiomas"`)).
			WillReturnError(assert.AnError)

		idiomas, err := repo.ListIdiomasByCPF(ctx, cpf)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao listar idiomas")
		assert.Nil(t, idiomas)
	})
}

// Cursos Complementares Tests

func TestCurriculoRepository_CreateCursoComplementar(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		curso := &empregabilidade.CurriculoCursoComplementar{
			ID:              uuid.New(),
			CPF:             "12345678900",
			NomeCurso:       "Python Avançado",
			NomeInstituicao: "Plataforma Online",
			AnoConclusao:    "2024",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_cursos_complementares"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(curso.ID))
		mock.ExpectCommit()

		id, err := repo.CreateCursoComplementar(ctx, curso)
		assert.NoError(t, err)
		assert.Equal(t, curso.ID, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		curso := &empregabilidade.CurriculoCursoComplementar{
			CPF:       "12345678900",
			NomeCurso: "Curso Teste",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_cursos_complementares"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.CreateCursoComplementar(ctx, curso)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar curso complementar")
		assert.Equal(t, uuid.Nil, id)
	})
}

func TestCurriculoRepository_GetCursoComplementarByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "cpf", "nome_curso", "nome_instituicao"}).
			AddRow(id, "12345678900", "Python", "Plataforma X")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_cursos_complementares"`)).
			WillReturnRows(rows)

		curso, err := repo.GetCursoComplementarByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, curso)
		assert.Equal(t, id, curso.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_cursos_complementares"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		curso, err := repo.GetCursoComplementarByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, curso)
	})

	t.Run("get by id error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_cursos_complementares"`)).
			WillReturnError(assert.AnError)

		curso, err := repo.GetCursoComplementarByID(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao buscar curso complementar")
		assert.Nil(t, curso)
	})
}

func TestCurriculoRepository_UpdateCursoComplementar(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		curso := &empregabilidade.CurriculoCursoComplementar{
			ID:              uuid.New(),
			CPF:             "12345678900",
			NomeCurso:       "Curso Atualizado",
			NomeInstituicao: "Plataforma Online",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_cursos_complementares"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateCursoComplementar(ctx, curso)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		curso := &empregabilidade.CurriculoCursoComplementar{
			ID:  uuid.New(),
			CPF: "12345678900",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_cursos_complementares"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateCursoComplementar(ctx, curso)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar curso complementar")
	})
}

func TestCurriculoRepository_DeleteCursoComplementar(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_cursos_complementares"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteCursoComplementar(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_cursos_complementares"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.DeleteCursoComplementar(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir curso complementar")
	})
}

func TestCurriculoRepository_ListCursosComplementaresByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("list success", func(t *testing.T) {
		cpf := "12345678900"
		rows := sqlmock.NewRows([]string{"id", "cpf", "nome_curso", "ano_conclusao"}).
			AddRow(uuid.New(), cpf, "Python", "2024").
			AddRow(uuid.New(), cpf, "Java", "2023")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_cursos_complementares"`)).
			WillReturnRows(rows)

		cursos, err := repo.ListCursosComplementaresByCPF(ctx, cpf)
		assert.NoError(t, err)
		assert.Len(t, cursos, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list error", func(t *testing.T) {
		cpf := "12345678900"
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_cursos_complementares"`)).
			WillReturnError(assert.AnError)

		cursos, err := repo.ListCursosComplementaresByCPF(ctx, cpf)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao listar cursos complementares")
		assert.Nil(t, cursos)
	})
}

// Experiências Tests

func TestCurriculoRepository_CreateExperiencia(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		experiencia := &empregabilidade.CurriculoExperiencia{
			ID:                  uuid.New(),
			CPF:                 "12345678900",
			Cargo:               "Desenvolvedor Senior",
			Empresa:             "Tech Corp",
			EhTrabalhoAtual:     true,
			DescricaoAtividades: "Desenvolvimento de sistemas",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(experiencia.ID))
		mock.ExpectCommit()

		id, err := repo.CreateExperiencia(ctx, experiencia)
		assert.NoError(t, err)
		assert.Equal(t, experiencia.ID, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		experiencia := &empregabilidade.CurriculoExperiencia{
			CPF:   "12345678900",
			Cargo: "Desenvolvedor",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.CreateExperiencia(ctx, experiencia)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar experiência")
		assert.Equal(t, uuid.Nil, id)
	})
}

func TestCurriculoRepository_GetExperienciaByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "cpf", "cargo", "empresa"}).
			AddRow(id, "12345678900", "Desenvolvedor", "Tech Corp")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_experiencias"`)).
			WillReturnRows(rows)

		experiencia, err := repo.GetExperienciaByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, experiencia)
		assert.Equal(t, id, experiencia.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_experiencias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		experiencia, err := repo.GetExperienciaByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, experiencia)
	})

	t.Run("get by id error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_experiencias"`)).
			WillReturnError(assert.AnError)

		experiencia, err := repo.GetExperienciaByID(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao buscar experiência")
		assert.Nil(t, experiencia)
	})
}

func TestCurriculoRepository_UpdateExperiencia(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		experiencia := &empregabilidade.CurriculoExperiencia{
			ID:                  uuid.New(),
			CPF:                 "12345678900",
			Cargo:               "Senior Developer",
			Empresa:             "Tech Corp",
			EhTrabalhoAtual:     false,
			DescricaoAtividades: "Desenvolvimento atualizado",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_experiencias"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateExperiencia(ctx, experiencia)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		experiencia := &empregabilidade.CurriculoExperiencia{
			ID:  uuid.New(),
			CPF: "12345678900",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_experiencias"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateExperiencia(ctx, experiencia)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar experiência")
	})
}

func TestCurriculoRepository_DeleteExperiencia(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteExperiencia(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.DeleteExperiencia(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir experiência")
	})
}

func TestCurriculoRepository_ListExperienciasByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("list success", func(t *testing.T) {
		cpf := "12345678900"
		rows := sqlmock.NewRows([]string{"id", "cpf", "cargo", "eh_trabalho_atual"}).
			AddRow(uuid.New(), cpf, "Desenvolvedor", true).
			AddRow(uuid.New(), cpf, "Analista", false)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_experiencias"`)).
			WillReturnRows(rows)

		experiencias, err := repo.ListExperienciasByCPF(ctx, cpf)
		assert.NoError(t, err)
		assert.Len(t, experiencias, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list error", func(t *testing.T) {
		cpf := "12345678900"
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_experiencias"`)).
			WillReturnError(assert.AnError)

		experiencias, err := repo.ListExperienciasByCPF(ctx, cpf)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao listar experiências")
		assert.Nil(t, experiencias)
	})
}

// Conquistas Tests

func TestCurriculoRepository_CreateConquista(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("create success", func(t *testing.T) {
		tipoConquistaID := uuid.New()
		conquista := &empregabilidade.CurriculoConquista{
			ID:              uuid.New(),
			CPF:             "12345678900",
			IDTipoConquista: &tipoConquistaID,
			Titulo:          "Certificação AWS",
			Descricao:       "Certificação AWS Solutions Architect",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_conquistas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(conquista.ID))
		mock.ExpectCommit()

		id, err := repo.CreateConquista(ctx, conquista)
		assert.NoError(t, err)
		assert.Equal(t, conquista.ID, id)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("create error", func(t *testing.T) {
		conquista := &empregabilidade.CurriculoConquista{
			CPF:       "12345678900",
			Descricao: "Teste",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_conquistas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		id, err := repo.CreateConquista(ctx, conquista)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao criar conquista")
		assert.Equal(t, uuid.Nil, id)
	})
}

func TestCurriculoRepository_GetConquistaByID(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("get by id found", func(t *testing.T) {
		id := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "cpf", "descricao", "id_tipo_conquista"}).
			AddRow(id, "12345678900", "Certificação", uuid.New())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_conquistas"`)).
			WillReturnRows(rows)

		// Expect TipoConquista preload query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_tipos_conquista"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		conquista, err := repo.GetConquistaByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, conquista)
		assert.Equal(t, id, conquista.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by id not found", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_conquistas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		conquista, err := repo.GetConquistaByID(ctx, id)
		assert.NoError(t, err)
		assert.Nil(t, conquista)
	})

	t.Run("get by id error", func(t *testing.T) {
		id := uuid.New()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_conquistas"`)).
			WillReturnError(assert.AnError)

		conquista, err := repo.GetConquistaByID(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao buscar conquista")
		assert.Nil(t, conquista)
	})
}

func TestCurriculoRepository_UpdateConquista(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("update success", func(t *testing.T) {
		conquista := &empregabilidade.CurriculoConquista{
			ID:        uuid.New(),
			CPF:       "12345678900",
			Descricao: "Certificação Atualizada",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_conquistas"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateConquista(ctx, conquista)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update error", func(t *testing.T) {
		conquista := &empregabilidade.CurriculoConquista{
			ID:  uuid.New(),
			CPF: "12345678900",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_conquistas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpdateConquista(ctx, conquista)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao atualizar conquista")
	})
}

func TestCurriculoRepository_DeleteConquista(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("delete success", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_conquistas"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteConquista(ctx, id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		id := uuid.New()

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_conquistas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.DeleteConquista(ctx, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao excluir conquista")
	})
}

func TestCurriculoRepository_ListConquistasByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("list success", func(t *testing.T) {
		cpf := "12345678900"
		tipoID1 := uuid.New()
		tipoID2 := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "cpf", "titulo", "descricao", "id_tipo_conquista"}).
			AddRow(uuid.New(), cpf, "Certificação AWS", "AWS Solutions Architect", tipoID1).
			AddRow(uuid.New(), cpf, "Hackathon Winner", "1st Place", tipoID2)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_conquistas"`)).
			WillReturnRows(rows)

		// Expect TipoConquista preload
		tipoRows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(tipoID1, "Certificação").
			AddRow(tipoID2, "Premiação")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_tipos_conquista"`)).
			WillReturnRows(tipoRows)

		conquistas, err := repo.ListConquistasByCPF(ctx, cpf)
		assert.NoError(t, err)
		assert.Len(t, conquistas, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list error", func(t *testing.T) {
		cpf := "12345678900"
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_conquistas"`)).
			WillReturnError(assert.AnError)

		conquistas, err := repo.ListConquistasByCPF(ctx, cpf)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao listar conquistas")
		assert.Nil(t, conquistas)
	})
}

// ReplaceAll Tests

func TestCurriculoRepository_ReplaceAllFormacoesByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("replace with items", func(t *testing.T) {
		cpf := "12345678900"
		formacoes := []*empregabilidade.CurriculoFormacao{
			{NomeCurso: "Engenharia", AnoConclusao: "2021"},
			{NomeCurso: "Matemática", AnoConclusao: "2018"},
		}

		mock.ExpectBegin()
		// Delete existing
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnResult(sqlmock.NewResult(0, 2))
		// Insert new
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_formacoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()).AddRow(uuid.New()))
		mock.ExpectCommit()

		err := repo.ReplaceAllFormacoesByCPF(ctx, cpf, formacoes)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("replace with empty list", func(t *testing.T) {
		cpf := "12345678900"

		mock.ExpectBegin()
		// Delete existing
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		err := repo.ReplaceAllFormacoesByCPF(ctx, cpf, []*empregabilidade.CurriculoFormacao{})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		cpf := "12345678900"
		formacoes := []*empregabilidade.CurriculoFormacao{
			{NomeCurso: "Engenharia"},
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllFormacoesByCPF(ctx, cpf, formacoes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao remover formações")
	})

	t.Run("insert error", func(t *testing.T) {
		cpf := "12345678900"
		formacoes := []*empregabilidade.CurriculoFormacao{
			{NomeCurso: "Engenharia"},
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_formacoes"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllFormacoesByCPF(ctx, cpf, formacoes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao inserir formações")
	})
}

func TestCurriculoRepository_ReplaceAllFormacaoAccordionByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("replace formacoes and idiomas", func(t *testing.T) {
		cpf := "12345678900"
		formacoes := []*empregabilidade.CurriculoFormacao{
			{NomeCurso: "Engenharia"},
		}
		idiomas := []*empregabilidade.CurriculoIdioma{
			{IDIdioma: uuid.New()},
		}

		mock.ExpectBegin()
		// Delete existing formacoes
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		// Insert formacoes
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_formacoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		// Delete existing idiomas
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_idiomas"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		// Insert idiomas
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_idiomas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		err := repo.ReplaceAllFormacaoAccordionByCPF(ctx, cpf, formacoes, idiomas)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete formacoes error", func(t *testing.T) {
		cpf := "12345678900"
		formacoes := []*empregabilidade.CurriculoFormacao{{NomeCurso: "Engenharia"}}
		idiomas := []*empregabilidade.CurriculoIdioma{{IDIdioma: uuid.New()}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllFormacaoAccordionByCPF(ctx, cpf, formacoes, idiomas)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao remover formações")
	})

	t.Run("insert idiomas error", func(t *testing.T) {
		cpf := "12345678900"
		formacoes := []*empregabilidade.CurriculoFormacao{{NomeCurso: "Engenharia"}}
		idiomas := []*empregabilidade.CurriculoIdioma{{IDIdioma: uuid.New()}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_formacoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_idiomas"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_idiomas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllFormacaoAccordionByCPF(ctx, cpf, formacoes, idiomas)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao inserir idiomas")
	})

	t.Run("delete idiomas error", func(t *testing.T) {
		cpf := "12345678900"
		formacoes := []*empregabilidade.CurriculoFormacao{{NomeCurso: "Engenharia"}}
		idiomas := []*empregabilidade.CurriculoIdioma{{IDIdioma: uuid.New()}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_formacoes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_idiomas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllFormacaoAccordionByCPF(ctx, cpf, formacoes, idiomas)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao remover idiomas")
	})

	t.Run("insert formacoes error", func(t *testing.T) {
		cpf := "12345678900"
		formacoes := []*empregabilidade.CurriculoFormacao{{NomeCurso: "Engenharia"}}
		idiomas := []*empregabilidade.CurriculoIdioma{{IDIdioma: uuid.New()}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_formacoes"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_formacoes"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllFormacaoAccordionByCPF(ctx, cpf, formacoes, idiomas)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao inserir formações")
	})
}

func TestCurriculoRepository_ReplaceAllExperienciasByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("replace with items", func(t *testing.T) {
		cpf := "12345678900"
		experiencias := []*empregabilidade.CurriculoExperiencia{
			{Cargo: "Desenvolvedor", Empresa: "Tech Corp"},
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		err := repo.ReplaceAllExperienciasByCPF(ctx, cpf, experiencias)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		cpf := "12345678900"
		experiencias := []*empregabilidade.CurriculoExperiencia{{Cargo: "Dev"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllExperienciasByCPF(ctx, cpf, experiencias)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao remover experiências")
	})

	t.Run("insert error", func(t *testing.T) {
		cpf := "12345678900"
		experiencias := []*empregabilidade.CurriculoExperiencia{{Cargo: "Dev"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllExperienciasByCPF(ctx, cpf, experiencias)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao inserir experiências")
	})
}

func TestCurriculoRepository_ReplaceAllExperienciaProfissionalAccordionByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("replace experiencias and conquistas", func(t *testing.T) {
		cpf := "12345678900"
		experiencias := []*empregabilidade.CurriculoExperiencia{
			{Cargo: "Desenvolvedor"},
		}
		conquistas := []*empregabilidade.CurriculoConquista{
			{Descricao: "Certificação"},
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_conquistas"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_conquistas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_perfil" ("cpf","resumo_profissional","created_at","updated_at") VALUES ($1,$2,$3,$4) ON CONFLICT ("cpf") DO UPDATE SET "resumo_profissional"="excluded"."resumo_profissional","updated_at"="excluded"."updated_at"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.ReplaceAllExperienciaProfissionalAccordionByCPF(ctx, cpf, experiencias, conquistas, "texto resumo")
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete experiencias error", func(t *testing.T) {
		cpf := "12345678900"
		experiencias := []*empregabilidade.CurriculoExperiencia{{Cargo: "Dev"}}
		conquistas := []*empregabilidade.CurriculoConquista{{Descricao: "Award"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllExperienciaProfissionalAccordionByCPF(ctx, cpf, experiencias, conquistas, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao remover experiências")
	})

	t.Run("insert conquistas error", func(t *testing.T) {
		cpf := "12345678900"
		experiencias := []*empregabilidade.CurriculoExperiencia{{Cargo: "Dev"}}
		conquistas := []*empregabilidade.CurriculoConquista{{Descricao: "Award"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_conquistas"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_conquistas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllExperienciaProfissionalAccordionByCPF(ctx, cpf, experiencias, conquistas, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao inserir conquistas")
	})

	t.Run("delete conquistas error", func(t *testing.T) {
		cpf := "12345678900"
		experiencias := []*empregabilidade.CurriculoExperiencia{{Cargo: "Dev"}}
		conquistas := []*empregabilidade.CurriculoConquista{{Descricao: "Award"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_conquistas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllExperienciaProfissionalAccordionByCPF(ctx, cpf, experiencias, conquistas, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao remover conquistas")
	})

	t.Run("insert experiencias error", func(t *testing.T) {
		cpf := "12345678900"
		experiencias := []*empregabilidade.CurriculoExperiencia{{Cargo: "Dev"}}
		conquistas := []*empregabilidade.CurriculoConquista{{Descricao: "Award"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllExperienciaProfissionalAccordionByCPF(ctx, cpf, experiencias, conquistas, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao inserir experiências")
	})

	t.Run("upsert perfil error", func(t *testing.T) {
		cpf := "12345678900"
		experiencias := []*empregabilidade.CurriculoExperiencia{{Cargo: "Dev"}}
		conquistas := []*empregabilidade.CurriculoConquista{{Descricao: "Award"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_experiencias"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_experiencias"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_conquistas"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_conquistas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_perfil" ("cpf","resumo_profissional","created_at","updated_at") VALUES ($1,$2,$3,$4) ON CONFLICT ("cpf") DO UPDATE SET "resumo_profissional"="excluded"."resumo_profissional","updated_at"="excluded"."updated_at"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllExperienciaProfissionalAccordionByCPF(ctx, cpf, experiencias, conquistas, "texto")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao salvar resumo profissional")
	})
}

func TestCurriculoRepository_ReplaceAllConquistasByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("replace with items", func(t *testing.T) {
		cpf := "12345678900"
		conquistas := []*empregabilidade.CurriculoConquista{
			{Descricao: "Certificação AWS"},
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_conquistas"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_conquistas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		err := repo.ReplaceAllConquistasByCPF(ctx, cpf, conquistas)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		cpf := "12345678900"
		conquistas := []*empregabilidade.CurriculoConquista{{Descricao: "Award"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_conquistas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllConquistasByCPF(ctx, cpf, conquistas)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao remover conquistas")
	})

	t.Run("insert error", func(t *testing.T) {
		cpf := "12345678900"
		conquistas := []*empregabilidade.CurriculoConquista{{Descricao: "Award"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_conquistas"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_conquistas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllConquistasByCPF(ctx, cpf, conquistas)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao inserir conquistas")
	})
}

func TestCurriculoRepository_ReplaceAllIdiomasByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("replace with items", func(t *testing.T) {
		cpf := "12345678900"
		idiomas := []*empregabilidade.CurriculoIdioma{
			{IDIdioma: uuid.New(), IDNivel: uuid.New()},
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_idiomas"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_idiomas"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		err := repo.ReplaceAllIdiomasByCPF(ctx, cpf, idiomas)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		cpf := "12345678900"
		idiomas := []*empregabilidade.CurriculoIdioma{{IDIdioma: uuid.New()}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_idiomas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllIdiomasByCPF(ctx, cpf, idiomas)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao remover idiomas")
	})

	t.Run("insert error", func(t *testing.T) {
		cpf := "12345678900"
		idiomas := []*empregabilidade.CurriculoIdioma{{IDIdioma: uuid.New()}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_idiomas"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_idiomas"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllIdiomasByCPF(ctx, cpf, idiomas)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao inserir idiomas")
	})
}

func TestCurriculoRepository_ReplaceAllCursosComplementaresByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("replace with items", func(t *testing.T) {
		cpf := "12345678900"
		cursos := []*empregabilidade.CurriculoCursoComplementar{
			{NomeCurso: "Python Avançado"},
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_cursos_complementares"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_cursos_complementares"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
		mock.ExpectCommit()

		err := repo.ReplaceAllCursosComplementaresByCPF(ctx, cpf, cursos)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete error", func(t *testing.T) {
		cpf := "12345678900"
		cursos := []*empregabilidade.CurriculoCursoComplementar{{NomeCurso: "Python"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_cursos_complementares"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllCursosComplementaresByCPF(ctx, cpf, cursos)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao remover cursos complementares")
	})

	t.Run("insert error", func(t *testing.T) {
		cpf := "12345678900"
		cursos := []*empregabilidade.CurriculoCursoComplementar{{NomeCurso: "Python"}}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "emp_curriculo_cursos_complementares"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_cursos_complementares"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.ReplaceAllCursosComplementaresByCPF(ctx, cpf, cursos)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao inserir cursos complementares")
	})
}

// Situação e Interesses Tests

func TestCurriculoRepository_UpsertSituacaoInteresses(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("upsert success", func(t *testing.T) {
		situacaoID := uuid.New()
		disponibilidadeID := uuid.New()
		situacao := &empregabilidade.CurriculoSituacaoInteresses{
			CPF:                    "12345678900",
			IDSituacao:             &situacaoID,
			IDDisponibilidade:      &disponibilidadeID,
			TempoProcurandoEmprego: "1-3 meses",
		}

		mock.ExpectBegin()
		// GORM.Save() with a primary key value performs UPDATE with ON CONFLICT INSERT
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "emp_curriculo_situacao_interesses"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpsertSituacaoInteresses(ctx, situacao)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("upsert error", func(t *testing.T) {
		situacao := &empregabilidade.CurriculoSituacaoInteresses{
			CPF: "12345678900",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "emp_curriculo_situacao_interesses"`)).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err := repo.UpsertSituacaoInteresses(ctx, situacao)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao salvar situação e interesses")
	})
}

func TestCurriculoRepository_GetSituacaoInteressesByCPF(t *testing.T) {
	db, mock, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	repo := NewCurriculoRepository(db)
	ctx := context.Background()

	t.Run("get by cpf found", func(t *testing.T) {
		cpf := "12345678900"
		situacaoID := uuid.New()
		disponibilidadeID := uuid.New()
		rows := sqlmock.NewRows([]string{"cpf", "id_situacao", "id_disponibilidade", "tempo_procurando_emprego"}).
			AddRow(cpf, situacaoID, disponibilidadeID, "1-3 meses")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_situacao_interesses"`)).
			WillReturnRows(rows)

		// Preloads happen in alphabetical order: Disponibilidade, then Situacao
		dispRows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(disponibilidadeID, "Imediato")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_disponibilidades"`)).
			WillReturnRows(dispRows)

		sitRows := sqlmock.NewRows([]string{"id", "nome"}).
			AddRow(situacaoID, "Empregado")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_situacoes_atual"`)).
			WillReturnRows(sitRows)

		situacao, err := repo.GetSituacaoInteressesByCPF(ctx, cpf)
		assert.NoError(t, err)
		assert.NotNil(t, situacao)
		assert.Equal(t, cpf, situacao.CPF)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by cpf not found", func(t *testing.T) {
		cpf := "12345678900"
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_situacao_interesses"`)).
			WillReturnRows(sqlmock.NewRows([]string{"cpf"}))

		situacao, err := repo.GetSituacaoInteressesByCPF(ctx, cpf)
		assert.NoError(t, err)
		assert.Nil(t, situacao)
	})

	t.Run("get by cpf error", func(t *testing.T) {
		cpf := "12345678900"
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "emp_curriculo_situacao_interesses"`)).
			WillReturnError(assert.AnError)

		situacao, err := repo.GetSituacaoInteressesByCPF(ctx, cpf)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "erro ao buscar situação e interesses")
		assert.Nil(t, situacao)
	})
}
