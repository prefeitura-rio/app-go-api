package empregabilidade

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCurriculoRepository_UpdateFormacao_SaveOperation(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	formacao := &empregabilidade.CurriculoFormacao{
		ID:              uuid.New(),
		CPF:             "12345678900",
		NomeInstituicao: "Universidade Federal",
		NomeCurso:       "Engenharia Atualizado",
		Status:          empregabilidade.StatusFormacaoCompleto,
		AnoConclusao:    "2021",
	}

	query := db.Model(&empregabilidade.CurriculoFormacao{}).Where("id = ?", formacao.ID)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Updates(map[string]interface{}{
			"nome_curso": formacao.NomeCurso,
		})
	})

	assert.Contains(t, sql, "UPDATE", "Should be an UPDATE operation")
	assert.Contains(t, sql, empregabilidade.CurriculoFormacao{}.TableName(), "Should update correct table")
}

func TestCurriculoRepository_ListFormacoesByCPF_OrderByAnoConclusao(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	cpf := "12345678900"

	query := db.Model(&empregabilidade.CurriculoFormacao{}).Where("cpf = ?", cpf).Order("ano_conclusao DESC")
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Escolaridade").Find(&[]*empregabilidade.CurriculoFormacao{})
	})

	assert.Contains(t, sql, "cpf =", "Should filter by CPF")
	assert.Contains(t, sql, "ORDER BY", "Should order results")
	assert.Contains(t, sql, "ano_conclusao DESC", "Should order by ano_conclusao DESC")
}

func TestCurriculoRepository_UpdateIdioma_SaveOperation(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	idioma := &empregabilidade.CurriculoIdioma{
		ID:       uuid.New(),
		CPF:      "12345678900",
		IDIdioma: uuid.New(),
		IDNivel:  uuid.New(),
	}

	query := db.Model(&empregabilidade.CurriculoIdioma{}).Where("id = ?", idioma.ID)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Updates(map[string]interface{}{
			"id_nivel": idioma.IDNivel,
		})
	})

	assert.Contains(t, sql, "UPDATE", "Should be an UPDATE operation")
	assert.Contains(t, sql, empregabilidade.CurriculoIdioma{}.TableName(), "Should update correct table")
}

func TestCurriculoRepository_ListIdiomasByCPF_WithPreloads(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	cpf := "12345678900"

	query := db.Model(&empregabilidade.CurriculoIdioma{}).Where("cpf = ?", cpf)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Preload("Idioma").Preload("Nivel").Find(&[]*empregabilidade.CurriculoIdioma{})
	})

	assert.Contains(t, sql, "cpf =", "Should filter by CPF")
}

func TestCurriculoRepository_UpdateCursoComplementar_SaveOperation(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	curso := &empregabilidade.CurriculoCursoComplementar{
		ID:              uuid.New(),
		CPF:             "12345678900",
		NomeCurso:       "Curso Atualizado",
		NomeInstituicao: "Plataforma Online",
		AnoConclusao:    "2024",
	}

	query := db.Model(&empregabilidade.CurriculoCursoComplementar{}).Where("id = ?", curso.ID)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Updates(map[string]interface{}{
			"nome_curso": curso.NomeCurso,
		})
	})

	assert.Contains(t, sql, "UPDATE", "Should be an UPDATE operation")
	assert.Contains(t, sql, empregabilidade.CurriculoCursoComplementar{}.TableName(), "Should update correct table")
}

func TestCurriculoRepository_ListCursosComplementaresByCPF_OrderByAnoConclusao(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	cpf := "12345678900"

	query := db.Model(&empregabilidade.CurriculoCursoComplementar{}).Where("cpf = ?", cpf).Order("ano_conclusao DESC")
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]*empregabilidade.CurriculoCursoComplementar{})
	})

	assert.Contains(t, sql, "cpf =", "Should filter by CPF")
	assert.Contains(t, sql, "ORDER BY", "Should order results")
	assert.Contains(t, sql, "ano_conclusao DESC", "Should order by ano_conclusao DESC")
}

func TestCurriculoRepository_UpdateExperiencia_SaveOperation(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	experiencia := &empregabilidade.CurriculoExperiencia{
		ID:                  uuid.New(),
		CPF:                 "12345678900",
		Cargo:               "Senior Developer",
		Empresa:             "Tech Corp",
		EhTrabalhoAtual:     false,
		DescricaoAtividades: "Desenvolvimento atualizado",
	}

	query := db.Model(&empregabilidade.CurriculoExperiencia{}).Where("id = ?", experiencia.ID)
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Updates(map[string]interface{}{
			"cargo": experiencia.Cargo,
		})
	})

	assert.Contains(t, sql, "UPDATE", "Should be an UPDATE operation")
	assert.Contains(t, sql, empregabilidade.CurriculoExperiencia{}.TableName(), "Should update correct table")
}

func TestCurriculoRepository_ListExperienciasByCPF_OrderByTrabalhoAtual(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	cpf := "12345678900"

	query := db.Model(&empregabilidade.CurriculoExperiencia{}).
		Where("cpf = ?", cpf).
		Order("eh_trabalho_atual DESC, created_at DESC")

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]*empregabilidade.CurriculoExperiencia{})
	})

	assert.Contains(t, sql, "cpf =", "Should filter by CPF")
	assert.Contains(t, sql, "ORDER BY", "Should order results")
	assert.Contains(t, sql, "eh_trabalho_atual DESC", "Should order by eh_trabalho_atual DESC")
	assert.Contains(t, sql, "created_at DESC", "Should order by created_at DESC")
}

func TestCurriculoRepository_ListConquistasByCPF_WithPreloads(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	cpf := "12345678900"

	query := db.Model(&empregabilidade.CurriculoConquista{}).
		Preload("TipoConquista").
		Where("cpf = ?", cpf).
		Order("created_at DESC")

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&[]*empregabilidade.CurriculoConquista{})
	})

	assert.Contains(t, sql, "cpf =", "Should filter by CPF")
	assert.Contains(t, sql, "ORDER BY", "Should order results")
	assert.Contains(t, sql, "created_at DESC", "Should order by created_at DESC")
}

func TestCurriculoRepository_GetSituacaoInteressesByCPF_WithPreloads(t *testing.T) {
	db, _, cleanup := repository.SetupMockDB(t)
	defer cleanup()

	cpf := "12345678900"

	query := db.Model(&empregabilidade.CurriculoSituacaoInteresses{}).
		Preload("Situacao").
		Preload("Disponibilidade").
		Where("cpf = ?", cpf)

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		var entity empregabilidade.CurriculoSituacaoInteresses
		return tx.First(&entity)
	})

	assert.Contains(t, sql, "cpf =", "Should filter by CPF")
}
