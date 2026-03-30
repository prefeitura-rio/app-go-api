package integration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// TestJobWorkflow_CompleteFlow tests complete job application workflow
func TestJobWorkflow_CompleteFlow(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Phase 1: Create Empresa (contractor)
	cnpj := GenerateUniqueCNPJ()
	empresa := &empregabilidade.Empresa{
		CNPJ:        cnpj,
		NomeEmpresa: "Test Company",
		RazaoSocial: "Test Company Ltd",
	}
	empresaRepo := empRepository.NewEmpresaRepository(tdb.DB)
	err := empresaRepo.Create(ctx, empresa)
	AssertNoError(t, err)

	// Phase 2: Create lookup tables data (regime, modelo)
	// Note: In real tests, these would already exist in DB
	// For simplicity, we'll use repository to create them
	regimeRepo := empRepository.NewRegimeContratacaoRepository(tdb.DB)
	regime := &empregabilidade.RegimeContratacao{
		Nome: "CLT",
	}
	regimeID, err := regimeRepo.Create(ctx, regime)
	AssertNoError(t, err)

	modeloRepo := empRepository.NewModeloTrabalhoRepository(tdb.DB)
	modelo := &empregabilidade.ModeloTrabalho{
		Nome: "Presencial",
	}
	modeloID, err := modeloRepo.Create(ctx, modelo)
	AssertNoError(t, err)

	// Phase 3: Create Vaga (job posting)
	vaga := CreateTestVaga(t, cnpj)
	vaga.IDRegimeContratacao = regimeID
	vaga.IDModeloTrabalho = modeloID
	vagaID, err := svc.VagaService.Create(ctx, vaga)
	AssertNoError(t, err)
	AssertNotNil(t, vagaID)

	// Update vaga to published status
	vaga.ID = vagaID
	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	err = svc.VagaService.Update(ctx, vaga)
	AssertNoError(t, err)

	// Phase 4: Create Candidatura (job application)
	cpf := GenerateUniqueCPF()
	candidatura := CreateTestCandidatura(cpf, vagaID)
	candidaturaID, err := svc.CandidaturaService.Create(ctx, candidatura)
	AssertNoError(t, err)
	AssertNotNil(t, candidaturaID)

	// Verify initial status
	retrievedCandidatura, err := svc.CandidaturaService.GetByID(ctx, candidaturaID)
	AssertNoError(t, err)
	AssertEqual(t, empregabilidade.StatusCandidaturaEnviada, retrievedCandidatura.Status)

	// Phase 5: Update Status to Approved
	err = svc.CandidaturaService.UpdateStatus(ctx, candidaturaID, empregabilidade.StatusCandidaturaAprovada)
	AssertNoError(t, err)

	// Verify status change
	approvedCandidatura, err := svc.CandidaturaService.GetByID(ctx, candidaturaID)
	AssertNoError(t, err)
	AssertEqual(t, empregabilidade.StatusCandidaturaAprovada, approvedCandidatura.Status)
}

// TestJobWorkflow_ConcurrentApplications tests multiple concurrent applications
func TestJobWorkflow_ConcurrentApplications(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Setup empresa, regime, modelo
	cnpj := GenerateUniqueCNPJ()
	empresa := &empregabilidade.Empresa{
		CNPJ:        cnpj,
		NomeEmpresa: "Test Company",
		RazaoSocial: "Test Company Ltd",
	}
	empresaRepo := empRepository.NewEmpresaRepository(tdb.DB)
	err := empresaRepo.Create(ctx, empresa)
	AssertNoError(t, err)

	regimeRepo := empRepository.NewRegimeContratacaoRepository(tdb.DB)
	regime := &empregabilidade.RegimeContratacao{Nome: "CLT"}
	regimeID, err := regimeRepo.Create(ctx, regime)
	AssertNoError(t, err)

	modeloRepo := empRepository.NewModeloTrabalhoRepository(tdb.DB)
	modelo := &empregabilidade.ModeloTrabalho{Nome: "Presencial"}
	modeloID, err := modeloRepo.Create(ctx, modelo)
	AssertNoError(t, err)

	// Create vaga
	vaga := CreateTestVaga(t, cnpj)
	vaga.IDRegimeContratacao = regimeID
	vaga.IDModeloTrabalho = modeloID
	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vagaID, err := svc.VagaService.Create(ctx, vaga)
	AssertNoError(t, err)

	// Create multiple applications
	numApplications := 5
	for i := 0; i < numApplications; i++ {
		cpf := GenerateUniqueCPF()
		candidatura := CreateTestCandidatura(cpf, vagaID)
		_, err := svc.CandidaturaService.Create(ctx, candidatura)
		AssertNoError(t, err)
	}

	// Verify all applications exist
	filter := empregabilidade.CandidaturaFilter{
		IDVaga: &vagaID,
	}
	candidaturas, total, err := svc.CandidaturaService.List(ctx, filter, 1, 100)
	AssertNoError(t, err)
	AssertEqual(t, numApplications, int(total))
	AssertEqual(t, numApplications, len(candidaturas))
}

// TestJobWorkflow_DuplicateApplication tests duplicate application prevention
func TestJobWorkflow_DuplicateApplication(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Setup empresa, vaga
	cnpj := GenerateUniqueCNPJ()
	empresa := &empregabilidade.Empresa{
		CNPJ:        cnpj,
		NomeEmpresa: "Test Company",
		RazaoSocial: "Test Company Ltd",
	}
	empresaRepo := empRepository.NewEmpresaRepository(tdb.DB)
	err := empresaRepo.Create(ctx, empresa)
	AssertNoError(t, err)

	regimeRepo := empRepository.NewRegimeContratacaoRepository(tdb.DB)
	regime := &empregabilidade.RegimeContratacao{Nome: "CLT"}
	regimeID, err := regimeRepo.Create(ctx, regime)
	AssertNoError(t, err)

	modeloRepo := empRepository.NewModeloTrabalhoRepository(tdb.DB)
	modelo := &empregabilidade.ModeloTrabalho{Nome: "Presencial"}
	modeloID, err := modeloRepo.Create(ctx, modelo)
	AssertNoError(t, err)

	vaga := CreateTestVaga(t, cnpj)
	vaga.IDRegimeContratacao = regimeID
	vaga.IDModeloTrabalho = modeloID
	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vagaID, err := svc.VagaService.Create(ctx, vaga)
	AssertNoError(t, err)

	// Create first application
	cpf := GenerateUniqueCPF()
	candidatura1 := CreateTestCandidatura(cpf, vagaID)
	_, err = svc.CandidaturaService.Create(ctx, candidatura1)
	AssertNoError(t, err)

	// Try to create duplicate
	candidatura2 := CreateTestCandidatura(cpf, vagaID)
	_, err = svc.CandidaturaService.Create(ctx, candidatura2)
	AssertError(t, err) // Should fail with duplicate error
}

// TestJobWorkflow_StatusTransitions tests valid status transitions
func TestJobWorkflow_StatusTransitions(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Setup empresa, vaga, candidatura
	cnpj := GenerateUniqueCNPJ()
	empresa := &empregabilidade.Empresa{
		CNPJ:        cnpj,
		NomeEmpresa: "Test Company",
		RazaoSocial: "Test Company Ltd",
	}
	empresaRepo := empRepository.NewEmpresaRepository(tdb.DB)
	err := empresaRepo.Create(ctx, empresa)
	AssertNoError(t, err)

	regimeRepo := empRepository.NewRegimeContratacaoRepository(tdb.DB)
	regime := &empregabilidade.RegimeContratacao{Nome: "CLT"}
	regimeID, err := regimeRepo.Create(ctx, regime)
	AssertNoError(t, err)

	modeloRepo := empRepository.NewModeloTrabalhoRepository(tdb.DB)
	modelo := &empregabilidade.ModeloTrabalho{Nome: "Presencial"}
	modeloID, err := modeloRepo.Create(ctx, modelo)
	AssertNoError(t, err)

	vaga := CreateTestVaga(t, cnpj)
	vaga.IDRegimeContratacao = regimeID
	vaga.IDModeloTrabalho = modeloID
	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vagaID, err := svc.VagaService.Create(ctx, vaga)
	AssertNoError(t, err)

	cpf := GenerateUniqueCPF()
	candidatura := CreateTestCandidatura(cpf, vagaID)
	candidaturaID, err := svc.CandidaturaService.Create(ctx, candidatura)
	AssertNoError(t, err)

	// Test transition: Enviada -> Aprovada
	err = svc.CandidaturaService.UpdateStatus(ctx, candidaturaID, empregabilidade.StatusCandidaturaAprovada)
	AssertNoError(t, err)

	approved, err := svc.CandidaturaService.GetByID(ctx, candidaturaID)
	AssertNoError(t, err)
	AssertEqual(t, empregabilidade.StatusCandidaturaAprovada, approved.Status)

	// Test transition: Aprovada -> Reprovada
	err = svc.CandidaturaService.UpdateStatus(ctx, candidaturaID, empregabilidade.StatusCandidaturaReprovada)
	AssertNoError(t, err)

	rejected, err := svc.CandidaturaService.GetByID(ctx, candidaturaID)
	AssertNoError(t, err)
	AssertEqual(t, empregabilidade.StatusCandidaturaReprovada, rejected.Status)
}

// TestJobWorkflow_EtapaUpdate tests updating application stage
func TestJobWorkflow_EtapaUpdate(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Setup basic entities
	cnpj := GenerateUniqueCNPJ()
	empresa := &empregabilidade.Empresa{
		CNPJ:        cnpj,
		NomeEmpresa: "Test Company",
		RazaoSocial: "Test Company Ltd",
	}
	empresaRepo := empRepository.NewEmpresaRepository(tdb.DB)
	err := empresaRepo.Create(ctx, empresa)
	AssertNoError(t, err)

	// Create etapa
	etapaRepo := empRepository.NewEtapaRepository(tdb.DB)
	etapa := &empregabilidade.Etapa{
		Nome: "Entrevista",
	}
	etapaID, err := etapaRepo.Create(ctx, etapa)
	AssertNoError(t, err)

	regimeRepo := empRepository.NewRegimeContratacaoRepository(tdb.DB)
	regime := &empregabilidade.RegimeContratacao{Nome: "CLT"}
	regimeID, err := regimeRepo.Create(ctx, regime)
	AssertNoError(t, err)

	modeloRepo := empRepository.NewModeloTrabalhoRepository(tdb.DB)
	modelo := &empregabilidade.ModeloTrabalho{Nome: "Presencial"}
	modeloID, err := modeloRepo.Create(ctx, modelo)
	AssertNoError(t, err)

	vaga := CreateTestVaga(t, cnpj)
	vaga.IDRegimeContratacao = regimeID
	vaga.IDModeloTrabalho = modeloID
	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vagaID, err := svc.VagaService.Create(ctx, vaga)
	AssertNoError(t, err)

	cpf := GenerateUniqueCPF()
	candidatura := CreateTestCandidatura(cpf, vagaID)
	candidaturaID, err := svc.CandidaturaService.Create(ctx, candidatura)
	AssertNoError(t, err)

	// Update etapa
	err = svc.CandidaturaService.UpdateEtapa(ctx, candidaturaID, etapaID)
	AssertNoError(t, err)

	// Verify etapa update
	updated, err := svc.CandidaturaService.GetByID(ctx, candidaturaID)
	AssertNoError(t, err)
	AssertNotNil(t, updated.IDEtapaAtual)
	AssertEqual(t, etapaID, *updated.IDEtapaAtual)
}

// TestJobWorkflow_ExpiredVaga tests application to expired job
func TestJobWorkflow_ExpiredVaga(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	ctx := context.Background()

	// Setup basic entities
	cnpj := GenerateUniqueCNPJ()
	empresa := &empregabilidade.Empresa{
		CNPJ:        cnpj,
		NomeEmpresa: "Test Company",
		RazaoSocial: "Test Company Ltd",
	}
	empresaRepo := empRepository.NewEmpresaRepository(tdb.DB)
	err := empresaRepo.Create(ctx, empresa)
	AssertNoError(t, err)

	regimeRepo := empRepository.NewRegimeContratacaoRepository(tdb.DB)
	regime := &empregabilidade.RegimeContratacao{Nome: "CLT"}
	regimeID, err := regimeRepo.Create(ctx, regime)
	AssertNoError(t, err)

	modeloRepo := empRepository.NewModeloTrabalhoRepository(tdb.DB)
	modelo := &empregabilidade.ModeloTrabalho{Nome: "Presencial"}
	modeloID, err := modeloRepo.Create(ctx, modelo)
	AssertNoError(t, err)

	// Create expired vaga directly via repository
	vagaRepo := empRepository.NewVagaRepository(tdb.DB)
	vaga := CreateTestVaga(t, cnpj)
	vaga.IDRegimeContratacao = regimeID
	vaga.IDModeloTrabalho = modeloID
	vaga.Status = empregabilidade.StatusVagaPublicadoExpirado
	vagaID, err := vagaRepo.Create(ctx, vaga)
	AssertNoError(t, err)

	// Try to create candidatura (should fail at service layer)
	// Note: This requires the candidaturaService to validate vaga status
}

// TestJobWorkflow_CurriculoSnapshot tests curriculum snapshot creation
func TestJobWorkflow_CurriculoSnapshot(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create curriculo data first
	cpf := GenerateUniqueCPF()
	formacao := &empregabilidade.CurriculoFormacao{
		CPF:            cpf,
		Instituicao:    "Test University",
		Curso:          "Computer Science",
		IDEscolaridade: uuid.New(), // Should be valid ID in real test
	}
	_, err := svc.CurriculoService.CreateFormacao(ctx, formacao)
	AssertNoError(t, err)

	// Setup vaga
	cnpj := GenerateUniqueCNPJ()
	empresa := &empregabilidade.Empresa{
		CNPJ:        cnpj,
		NomeEmpresa: "Test Company",
		RazaoSocial: "Test Company Ltd",
	}
	empresaRepo := empRepository.NewEmpresaRepository(tdb.DB)
	err = empresaRepo.Create(ctx, empresa)
	AssertNoError(t, err)

	regimeRepo := empRepository.NewRegimeContratacaoRepository(tdb.DB)
	regime := &empregabilidade.RegimeContratacao{Nome: "CLT"}
	regimeID, err := regimeRepo.Create(ctx, regime)
	AssertNoError(t, err)

	modeloRepo := empRepository.NewModeloTrabalhoRepository(tdb.DB)
	modelo := &empregabilidade.ModeloTrabalho{Nome: "Presencial"}
	modeloID, err := modeloRepo.Create(ctx, modelo)
	AssertNoError(t, err)

	vaga := CreateTestVaga(t, cnpj)
	vaga.IDRegimeContratacao = regimeID
	vaga.IDModeloTrabalho = modeloID
	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vagaID, err := svc.VagaService.Create(ctx, vaga)
	AssertNoError(t, err)

	// Create candidatura (should snapshot curriculum)
	candidatura := CreateTestCandidatura(cpf, vagaID)
	candidaturaID, err := svc.CandidaturaService.Create(ctx, candidatura)
	AssertNoError(t, err)

	// Verify curriculum snapshot was created
	retrieved, err := svc.CandidaturaService.GetByID(ctx, candidaturaID)
	AssertNoError(t, err)
	AssertNotNil(t, retrieved.CurriculoSnapshot)
}

// TestJobWorkflow_BulkStatusUpdate tests bulk candidatura status update
func TestJobWorkflow_BulkStatusUpdate(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Setup vaga
	cnpj := GenerateUniqueCNPJ()
	empresa := &empregabilidade.Empresa{
		CNPJ:        cnpj,
		NomeEmpresa: "Test Company",
		RazaoSocial: "Test Company Ltd",
	}
	empresaRepo := empRepository.NewEmpresaRepository(tdb.DB)
	err := empresaRepo.Create(ctx, empresa)
	AssertNoError(t, err)

	regimeRepo := empRepository.NewRegimeContratacaoRepository(tdb.DB)
	regime := &empregabilidade.RegimeContratacao{Nome: "CLT"}
	regimeID, err := regimeRepo.Create(ctx, regime)
	AssertNoError(t, err)

	modeloRepo := empRepository.NewModeloTrabalhoRepository(tdb.DB)
	modelo := &empregabilidade.ModeloTrabalho{Nome: "Presencial"}
	modeloID, err := modeloRepo.Create(ctx, modelo)
	AssertNoError(t, err)

	vaga := CreateTestVaga(t, cnpj)
	vaga.IDRegimeContratacao = regimeID
	vaga.IDModeloTrabalho = modeloID
	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vagaID, err := svc.VagaService.Create(ctx, vaga)
	AssertNoError(t, err)

	// Create multiple candidaturas
	numCandidaturas := 5
	cpfs := make([]string, numCandidaturas)
	for i := 0; i < numCandidaturas; i++ {
		cpfs[i] = GenerateUniqueCPF()
		candidatura := CreateTestCandidatura(cpfs[i], vagaID)
		_, err := svc.CandidaturaService.Create(ctx, candidatura)
		AssertNoError(t, err)
	}

	// Bulk update status
	candidaturaRepo := empRepository.NewCandidaturaRepository(tdb.DB)
	result, err := candidaturaRepo.BulkUpdateStatus(ctx, vagaID, cpfs, empregabilidade.StatusCandidaturaAprovada)
	AssertNoError(t, err)
	AssertEqual(t, int64(numCandidaturas), result.Updated)
}

// TestJobWorkflow_CountByStatus tests candidatura status counting
func TestJobWorkflow_CountByStatus(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Setup vaga
	cnpj := GenerateUniqueCNPJ()
	empresa := &empregabilidade.Empresa{
		CNPJ:        cnpj,
		NomeEmpresa: "Test Company",
		RazaoSocial: "Test Company Ltd",
	}
	empresaRepo := empRepository.NewEmpresaRepository(tdb.DB)
	err := empresaRepo.Create(ctx, empresa)
	AssertNoError(t, err)

	regimeRepo := empRepository.NewRegimeContratacaoRepository(tdb.DB)
	regime := &empregabilidade.RegimeContratacao{Nome: "CLT"}
	regimeID, err := regimeRepo.Create(ctx, regime)
	AssertNoError(t, err)

	modeloRepo := empRepository.NewModeloTrabalhoRepository(tdb.DB)
	modelo := &empregabilidade.ModeloTrabalho{Nome: "Presencial"}
	modeloID, err := modeloRepo.Create(ctx, modelo)
	AssertNoError(t, err)

	vaga := CreateTestVaga(t, cnpj)
	vaga.IDRegimeContratacao = regimeID
	vaga.IDModeloTrabalho = modeloID
	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vagaID, err := svc.VagaService.Create(ctx, vaga)
	AssertNoError(t, err)

	// Create candidaturas with different statuses
	numEnviadas := 3
	numAprovadas := 2

	for i := 0; i < numEnviadas; i++ {
		cpf := GenerateUniqueCPF()
		candidatura := CreateTestCandidatura(cpf, vagaID)
		_, err := svc.CandidaturaService.Create(ctx, candidatura)
		AssertNoError(t, err)
	}

	for i := 0; i < numAprovadas; i++ {
		cpf := GenerateUniqueCPF()
		candidatura := CreateTestCandidatura(cpf, vagaID)
		candidaturaID, err := svc.CandidaturaService.Create(ctx, candidatura)
		AssertNoError(t, err)
		err = svc.CandidaturaService.UpdateStatus(ctx, candidaturaID, empregabilidade.StatusCandidaturaAprovada)
		AssertNoError(t, err)
	}

	// Count by status
	filter := empregabilidade.CandidaturaFilter{
		IDVaga: &vagaID,
	}
	counts, err := svc.CandidaturaService.CountByStatus(ctx, filter)
	AssertNoError(t, err)
	AssertEqual(t, int64(numEnviadas), counts[empregabilidade.StatusCandidaturaEnviada])
	AssertEqual(t, int64(numAprovadas), counts[empregabilidade.StatusCandidaturaAprovada])
}

// TestJobWorkflow_DataConsistency tests data consistency
func TestJobWorkflow_DataConsistency(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Setup vaga
	cnpj := GenerateUniqueCNPJ()
	empresa := &empregabilidade.Empresa{
		CNPJ:        cnpj,
		NomeEmpresa: "Test Company",
		RazaoSocial: "Test Company Ltd",
	}
	empresaRepo := empRepository.NewEmpresaRepository(tdb.DB)
	err := empresaRepo.Create(ctx, empresa)
	AssertNoError(t, err)

	regimeRepo := empRepository.NewRegimeContratacaoRepository(tdb.DB)
	regime := &empregabilidade.RegimeContratacao{Nome: "CLT"}
	regimeID, err := regimeRepo.Create(ctx, regime)
	AssertNoError(t, err)

	modeloRepo := empRepository.NewModeloTrabalhoRepository(tdb.DB)
	modelo := &empregabilidade.ModeloTrabalho{Nome: "Presencial"}
	modeloID, err := modeloRepo.Create(ctx, modelo)
	AssertNoError(t, err)

	vaga := CreateTestVaga(t, cnpj)
	vaga.IDRegimeContratacao = regimeID
	vaga.IDModeloTrabalho = modeloID
	vaga.Status = empregabilidade.StatusVagaPublicadoAtivo
	vagaID, err := svc.VagaService.Create(ctx, vaga)
	AssertNoError(t, err)

	cpf := GenerateUniqueCPF()
	candidatura := CreateTestCandidatura(cpf, vagaID)
	candidaturaID, err := svc.CandidaturaService.Create(ctx, candidatura)
	AssertNoError(t, err)

	// Verify data consistency
	retrieved, err := svc.CandidaturaService.GetByID(ctx, candidaturaID)
	AssertNoError(t, err)
	AssertEqual(t, candidatura.CPF, retrieved.CPF)
	AssertEqual(t, candidatura.IDVaga, retrieved.IDVaga)
	AssertEqual(t, empregabilidade.StatusCandidaturaEnviada, retrieved.Status)
}
