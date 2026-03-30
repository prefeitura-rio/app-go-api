package integration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// TestCurriculoWorkflow_CompleteFlow tests building complete curriculum
func TestCurriculoWorkflow_CompleteFlow(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	cpf := GenerateUniqueCPF()

	// Create escolaridade for testing
	escolaridadeRepo := empRepository.NewEscolaridadeRepository(tdb.DB)
	escolaridade := &empregabilidade.Escolaridade{Nome: "Superior Completo"}
	escolaridadeID, err := escolaridadeRepo.Create(ctx, escolaridade)
	AssertNoError(t, err)

	// Phase 1: Create Formacao
	formacao := &empregabilidade.CurriculoFormacao{
		CPF:            cpf,
		Instituicao:    "Test University",
		Curso:          "Computer Science",
		IDEscolaridade: escolaridadeID,
	}
	formacaoID, err := svc.CurriculoService.CreateFormacao(ctx, formacao)
	AssertNoError(t, err)
	AssertNotNil(t, formacaoID)

	// Phase 2: Create Idioma
	idiomaRepo := empRepository.NewIdiomaRepository(tdb.DB)
	idioma := &empregabilidade.Idioma{Nome: "English"}
	idiomaID, err := idiomaRepo.Create(ctx, idioma)
	AssertNoError(t, err)

	nivelIdiomaRepo := empRepository.NewNivelIdiomaRepository(tdb.DB)
	nivelIdioma := &empregabilidade.NivelIdioma{Nome: "Fluente"}
	nivelIdiomaID, err := nivelIdiomaRepo.Create(ctx, nivelIdioma)
	AssertNoError(t, err)

	curriculoIdioma := &empregabilidade.CurriculoIdioma{
		CPF:           cpf,
		IDIdioma:      idiomaID,
		IDNivelIdioma: nivelIdiomaID,
	}
	idiomaIDCreated, err := svc.CurriculoService.CreateIdioma(ctx, curriculoIdioma)
	AssertNoError(t, err)
	AssertNotNil(t, idiomaIDCreated)

	// Phase 3: Create Curso Complementar
	cursoComplementar := &empregabilidade.CurriculoCursoComplementar{
		CPF:         cpf,
		Titulo:      "AWS Certification",
		Instituicao: "Amazon",
	}
	cursoID, err := svc.CurriculoService.CreateCursoComplementar(ctx, cursoComplementar)
	AssertNoError(t, err)
	AssertNotNil(t, cursoID)

	// Phase 4: Create Experiencia
	experiencia := &empregabilidade.CurriculoExperiencia{
		CPF:       cpf,
		Empresa:   "Test Company",
		Cargo:     "Software Engineer",
		Descricao: "Developed amazing software",
	}
	experienciaID, err := svc.CurriculoService.CreateExperiencia(ctx, experiencia)
	AssertNoError(t, err)
	AssertNotNil(t, experienciaID)

	// Phase 5: Create Conquista
	tipoConquistaRepo := empRepository.NewTipoConquistaRepository(tdb.DB)
	tipoConquista := &empregabilidade.TipoConquista{Nome: "Premio"}
	tipoConquistaID, err := tipoConquistaRepo.Create(ctx, tipoConquista)
	AssertNoError(t, err)

	conquista := &empregabilidade.CurriculoConquista{
		CPF:             cpf,
		Titulo:          "Best Developer Award",
		Descricao:       "Awarded for excellence",
		IDTipoConquista: tipoConquistaID,
	}
	conquistaID, err := svc.CurriculoService.CreateConquista(ctx, conquista)
	AssertNoError(t, err)
	AssertNotNil(t, conquistaID)

	// Verify complete curriculum
	curriculo, err := svc.CurriculoService.GetCurriculoCompleto(ctx, cpf)
	AssertNoError(t, err)
	AssertNotNil(t, curriculo)
	AssertEqual(t, 1, len(curriculo.Formacoes))
	AssertEqual(t, 1, len(curriculo.Idiomas))
	AssertEqual(t, 1, len(curriculo.CursosComplementares))
	AssertEqual(t, 1, len(curriculo.Experiencias))
	AssertEqual(t, 1, len(curriculo.Conquistas))
}

// TestCurriculoWorkflow_ReplaceAllFormacoes tests bulk replace formacoes
func TestCurriculoWorkflow_ReplaceAllFormacoes(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	cpf := GenerateUniqueCPF()

	// Create escolaridade
	escolaridadeRepo := empRepository.NewEscolaridadeRepository(tdb.DB)
	escolaridade := &empregabilidade.Escolaridade{Nome: "Superior Completo"}
	escolaridadeID, err := escolaridadeRepo.Create(ctx, escolaridade)
	AssertNoError(t, err)

	// Create initial formacao
	formacao1 := &empregabilidade.CurriculoFormacao{
		CPF:            cpf,
		Instituicao:    "University A",
		Curso:          "Course A",
		IDEscolaridade: escolaridadeID,
	}
	_, err = svc.CurriculoService.CreateFormacao(ctx, formacao1)
	AssertNoError(t, err)

	// ReplaceAll with new list
	formacao2 := &empregabilidade.CurriculoFormacao{
		CPF:            cpf,
		Instituicao:    "University B",
		Curso:          "Course B",
		IDEscolaridade: escolaridadeID,
	}
	formacao3 := &empregabilidade.CurriculoFormacao{
		CPF:            cpf,
		Instituicao:    "University C",
		Curso:          "Course C",
		IDEscolaridade: escolaridadeID,
	}
	newFormacoes := []*empregabilidade.CurriculoFormacao{formacao2, formacao3}
	err = svc.CurriculoService.ReplaceAllFormacoesByCPF(ctx, cpf, newFormacoes)
	AssertNoError(t, err)

	// Verify replacement
	formacoes, err := svc.CurriculoService.ListFormacoesByCPF(ctx, cpf)
	AssertNoError(t, err)
	AssertEqual(t, 2, len(formacoes))
}

// TestCurriculoWorkflow_ReplaceAllExperiencias tests bulk replace experiencias
func TestCurriculoWorkflow_ReplaceAllExperiencias(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	cpf := GenerateUniqueCPF()

	// Create initial experiencias
	exp1 := &empregabilidade.CurriculoExperiencia{
		CPF:       cpf,
		Empresa:   "Company A",
		Cargo:     "Position A",
		Descricao: "Description A",
	}
	_, err := svc.CurriculoService.CreateExperiencia(ctx, exp1)
	AssertNoError(t, err)

	// ReplaceAll with new list
	exp2 := &empregabilidade.CurriculoExperiencia{
		CPF:       cpf,
		Empresa:   "Company B",
		Cargo:     "Position B",
		Descricao: "Description B",
	}
	exp3 := &empregabilidade.CurriculoExperiencia{
		CPF:       cpf,
		Empresa:   "Company C",
		Cargo:     "Position C",
		Descricao: "Description C",
	}
	newExperiencias := []*empregabilidade.CurriculoExperiencia{exp2, exp3}
	err = svc.CurriculoService.ReplaceAllExperienciasByCPF(ctx, cpf, newExperiencias)
	AssertNoError(t, err)

	// Verify replacement
	experiencias, err := svc.CurriculoService.ListExperienciasByCPF(ctx, cpf)
	AssertNoError(t, err)
	AssertEqual(t, 2, len(experiencias))
}

// TestCurriculoWorkflow_ValidationOnCreate tests validation during creation
func TestCurriculoWorkflow_ValidationOnCreate(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	cpf := GenerateUniqueCPF()

	// Try to create invalid formacao (missing required fields)
	invalidFormacao := &empregabilidade.CurriculoFormacao{
		CPF: cpf,
		// Missing Instituicao, Curso, IDEscolaridade
	}
	_, err := svc.CurriculoService.CreateFormacao(ctx, invalidFormacao)
	AssertError(t, err) // Should fail validation
}

// TestCurriculoWorkflow_UpdateAndDelete tests update and delete operations
func TestCurriculoWorkflow_UpdateAndDelete(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	cpf := GenerateUniqueCPF()

	// Create escolaridade
	escolaridadeRepo := empRepository.NewEscolaridadeRepository(tdb.DB)
	escolaridade := &empregabilidade.Escolaridade{Nome: "Superior Completo"}
	escolaridadeID, err := escolaridadeRepo.Create(ctx, escolaridade)
	AssertNoError(t, err)

	// Create formacao
	formacao := &empregabilidade.CurriculoFormacao{
		CPF:            cpf,
		Instituicao:    "Original University",
		Curso:          "Original Course",
		IDEscolaridade: escolaridadeID,
	}
	formacaoID, err := svc.CurriculoService.CreateFormacao(ctx, formacao)
	AssertNoError(t, err)

	// Update formacao
	formacao.ID = formacaoID
	formacao.Instituicao = "Updated University"
	formacao.Curso = "Updated Course"
	err = svc.CurriculoService.UpdateFormacao(ctx, formacao)
	AssertNoError(t, err)

	// Verify update
	updated, err := svc.CurriculoService.GetFormacaoByID(ctx, formacaoID)
	AssertNoError(t, err)
	AssertEqual(t, "Updated University", updated.Instituicao)
	AssertEqual(t, "Updated Course", updated.Curso)

	// Delete formacao
	err = svc.CurriculoService.DeleteFormacao(ctx, formacaoID)
	AssertNoError(t, err)

	// Verify deletion
	formacoes, err := svc.CurriculoService.ListFormacoesByCPF(ctx, cpf)
	AssertNoError(t, err)
	AssertEqual(t, 0, len(formacoes))
}

// TestCurriculoWorkflow_DataConsistency tests data consistency
func TestCurriculoWorkflow_DataConsistency(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	cpf := GenerateUniqueCPF()

	// Create experiencia
	experiencia := &empregabilidade.CurriculoExperiencia{
		CPF:       cpf,
		Empresa:   "Test Company",
		Cargo:     "Test Position",
		Descricao: "Test Description",
	}
	experienciaID, err := svc.CurriculoService.CreateExperiencia(ctx, experiencia)
	AssertNoError(t, err)

	// Verify data consistency
	retrieved, err := svc.CurriculoService.GetExperienciaByID(ctx, experienciaID)
	AssertNoError(t, err)
	AssertEqual(t, experiencia.CPF, retrieved.CPF)
	AssertEqual(t, experiencia.Empresa, retrieved.Empresa)
	AssertEqual(t, experiencia.Cargo, retrieved.Cargo)
	AssertEqual(t, experiencia.Descricao, retrieved.Descricao)
}
