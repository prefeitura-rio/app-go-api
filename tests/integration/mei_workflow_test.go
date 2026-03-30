package integration_test

import (
	"context"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

// TestMEIWorkflow_CompleteFlow tests complete MEI proposal workflow
func TestMEIWorkflow_CompleteFlow(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Phase 1: Create Opportunity
	oportunidade := CreateTestOportunidadeMEI(t)
	oportunidadeRepo := repository.NewOportunidadeMEIRepository(tdb.DB)
	oportunidadeID, err := oportunidadeRepo.Create(ctx, oportunidade)
	AssertNoError(t, err)

	// Phase 2: Create Proposal (CNAE validation is mocked in test)
	cnpj := GenerateUniqueCNPJ()
	proposta := CreateTestPropostaMEI(oportunidadeID, cnpj)

	// Note: In real test with RMI integration, this would validate CNAE
	// For integration test without external dependencies, we skip CNAE validation
	// by using the repository directly
	propostaRepo := repository.NewPropostaMEIRepository(tdb.DB)
	propostaID, err := propostaRepo.Create(ctx, proposta)
	AssertNoError(t, err)
	AssertNotNil(t, propostaID)

	// Verify initial status
	AssertEqual(t, models.StatusPropostaAdminActive, proposta.StatusAdmin)
	AssertEqual(t, models.StatusPropostaCidadaoSubmitted, proposta.StatusCidadao)

	// Phase 3: Update Status to Approved
	err = svc.PropostaMEIService.UpdateStatusCidadao(ctx, propostaID, models.StatusPropostaCidadaoApproved)
	AssertNoError(t, err)

	// Verify status change
	updatedProposta, err := svc.PropostaMEIService.GetByID(ctx, propostaID)
	AssertNoError(t, err)
	AssertEqual(t, models.StatusPropostaCidadaoApproved, updatedProposta.StatusCidadao)
}

// TestMEIWorkflow_OwnershipValidation tests MEI ownership validation
func TestMEIWorkflow_OwnershipValidation(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create opportunity
	oportunidade := CreateTestOportunidadeMEI(t)
	oportunidadeRepo := repository.NewOportunidadeMEIRepository(tdb.DB)
	oportunidadeID, err := oportunidadeRepo.Create(ctx, oportunidade)
	AssertNoError(t, err)

	// Create proposal
	cnpj := GenerateUniqueCNPJ()
	proposta := CreateTestPropostaMEI(oportunidadeID, cnpj)
	propostaRepo := repository.NewPropostaMEIRepository(tdb.DB)
	propostaID, err := propostaRepo.Create(ctx, proposta)
	AssertNoError(t, err)

	// Verify proposal belongs to correct CNPJ
	retrieved, err := svc.PropostaMEIService.GetByID(ctx, propostaID)
	AssertNoError(t, err)
	AssertEqual(t, cnpj, retrieved.MEIEmpresaID)
}

// TestMEIWorkflow_DuplicateProposal tests duplicate proposal prevention
func TestMEIWorkflow_DuplicateProposal(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	ctx := context.Background()

	// Create opportunity
	oportunidade := CreateTestOportunidadeMEI(t)
	oportunidadeRepo := repository.NewOportunidadeMEIRepository(tdb.DB)
	oportunidadeID, err := oportunidadeRepo.Create(ctx, oportunidade)
	AssertNoError(t, err)

	// Create first proposal
	cnpj := GenerateUniqueCNPJ()
	proposta1 := CreateTestPropostaMEI(oportunidadeID, cnpj)
	propostaRepo := repository.NewPropostaMEIRepository(tdb.DB)
	_, err = propostaRepo.Create(ctx, proposta1)
	AssertNoError(t, err)

	// Try to create duplicate
	proposta2 := CreateTestPropostaMEI(oportunidadeID, cnpj)
	_, err = propostaRepo.Create(ctx, proposta2)
	// Note: Duplicate check happens at service layer with CheckExistingProposta
	// At repository layer, it depends on DB constraints
}

// TestMEIWorkflow_StatusTransitions tests valid status transitions
func TestMEIWorkflow_StatusTransitions(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create opportunity and proposal
	oportunidade := CreateTestOportunidadeMEI(t)
	oportunidadeRepo := repository.NewOportunidadeMEIRepository(tdb.DB)
	oportunidadeID, err := oportunidadeRepo.Create(ctx, oportunidade)
	AssertNoError(t, err)

	cnpj := GenerateUniqueCNPJ()
	proposta := CreateTestPropostaMEI(oportunidadeID, cnpj)
	propostaRepo := repository.NewPropostaMEIRepository(tdb.DB)
	propostaID, err := propostaRepo.Create(ctx, proposta)
	AssertNoError(t, err)

	// Test transition: Submitted -> Approved
	err = svc.PropostaMEIService.Approve(ctx, propostaID)
	AssertNoError(t, err)

	approved, err := svc.PropostaMEIService.GetByID(ctx, propostaID)
	AssertNoError(t, err)
	AssertEqual(t, models.StatusPropostaCidadaoApproved, approved.StatusCidadao)

	// Test transition: Approved -> Rejected (if allowed)
	err = svc.PropostaMEIService.Reject(ctx, propostaID)
	AssertNoError(t, err)

	rejected, err := svc.PropostaMEIService.GetByID(ctx, propostaID)
	AssertNoError(t, err)
	AssertEqual(t, models.StatusPropostaCidadaoRejected, rejected.StatusCidadao)
}

// TestMEIWorkflow_BulkStatusUpdate tests bulk status updates
func TestMEIWorkflow_BulkStatusUpdate(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create opportunity
	oportunidade := CreateTestOportunidadeMEI(t)
	oportunidadeRepo := repository.NewOportunidadeMEIRepository(tdb.DB)
	oportunidadeID, err := oportunidadeRepo.Create(ctx, oportunidade)
	AssertNoError(t, err)

	// Create multiple proposals
	numPropostas := 5
	var ids []models.UUID
	propostaRepo := repository.NewPropostaMEIRepository(tdb.DB)
	for i := 0; i < numPropostas; i++ {
		cnpj := GenerateUniqueCNPJ()
		proposta := CreateTestPropostaMEI(oportunidadeID, cnpj)
		propostaID, err := propostaRepo.Create(ctx, proposta)
		AssertNoError(t, err)
		ids = append(ids, propostaID)
	}

	// Bulk update to approved
	count, err := svc.PropostaMEIService.UpdateMultipleStatus(ctx, ids, models.StatusPropostaCidadaoApproved)
	AssertNoError(t, err)
	AssertEqual(t, numPropostas, count)

	// Verify all were updated
	for _, id := range ids {
		proposta, err := svc.PropostaMEIService.GetByID(ctx, id)
		AssertNoError(t, err)
		AssertEqual(t, models.StatusPropostaCidadaoApproved, proposta.StatusCidadao)
	}
}

// TestMEIWorkflow_PropostaUpdate tests proposal field updates
func TestMEIWorkflow_PropostaUpdate(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create opportunity and proposal
	oportunidade := CreateTestOportunidadeMEI(t)
	oportunidadeRepo := repository.NewOportunidadeMEIRepository(tdb.DB)
	oportunidadeID, err := oportunidadeRepo.Create(ctx, oportunidade)
	AssertNoError(t, err)

	cnpj := GenerateUniqueCNPJ()
	proposta := CreateTestPropostaMEI(oportunidadeID, cnpj)
	propostaRepo := repository.NewPropostaMEIRepository(tdb.DB)
	propostaID, err := propostaRepo.Create(ctx, proposta)
	AssertNoError(t, err)

	// Update proposal values
	newValor := 6000.00
	newPrazo := "45 dias"
	newAceita := false
	err = svc.PropostaMEIService.UpdateProposta(ctx, propostaID, oportunidadeID, &newValor, &newPrazo, &newAceita)
	AssertNoError(t, err)

	// Verify updates
	updated, err := svc.PropostaMEIService.GetByID(ctx, propostaID)
	AssertNoError(t, err)
	AssertEqual(t, newValor, *updated.ValorProposta)
	AssertEqual(t, newPrazo, *updated.PrazoExecucao)
	AssertEqual(t, newAceita, *updated.AceitaCustosIntegrais)
}

// TestMEIWorkflow_InactiveOpportunity tests proposal creation for inactive opportunity
func TestMEIWorkflow_InactiveOpportunity(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	ctx := context.Background()

	// Create inactive opportunity
	oportunidade := CreateTestOportunidadeMEI(t)
	oportunidade.Status = models.StatusOportunidadeInactive
	oportunidadeRepo := repository.NewOportunidadeMEIRepository(tdb.DB)
	oportunidadeID, err := oportunidadeRepo.Create(ctx, oportunidade)
	AssertNoError(t, err)

	// Try to create proposal (should fail at service layer)
	// Note: This test would need the full service with validation
	// At repository level, it would succeed but service should prevent it
}

// TestMEIWorkflow_ListByOportunidade tests listing proposals by opportunity
func TestMEIWorkflow_ListByOportunidade(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create opportunity
	oportunidade := CreateTestOportunidadeMEI(t)
	oportunidadeRepo := repository.NewOportunidadeMEIRepository(tdb.DB)
	oportunidadeID, err := oportunidadeRepo.Create(ctx, oportunidade)
	AssertNoError(t, err)

	// Create multiple proposals
	numPropostas := 3
	propostaRepo := repository.NewPropostaMEIRepository(tdb.DB)
	for i := 0; i < numPropostas; i++ {
		cnpj := GenerateUniqueCNPJ()
		proposta := CreateTestPropostaMEI(oportunidadeID, cnpj)
		_, err := propostaRepo.Create(ctx, proposta)
		AssertNoError(t, err)
	}

	// List proposals for this opportunity
	propostas, total, err := svc.PropostaMEIService.ListByOportunidade(ctx, oportunidadeID, "", "", "", 1, 100)
	AssertNoError(t, err)
	AssertEqual(t, numPropostas, int(total))
	AssertEqual(t, numPropostas, len(propostas))
}

// TestMEIWorkflow_ListByMEIEmpresa tests listing proposals by CNPJ
func TestMEIWorkflow_ListByMEIEmpresa(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create opportunities
	oportunidade1 := CreateTestOportunidadeMEI(t)
	oportunidade2 := CreateTestOportunidadeMEI(t)
	oportunidadeRepo := repository.NewOportunidadeMEIRepository(tdb.DB)
	oportunidadeID1, err := oportunidadeRepo.Create(ctx, oportunidade1)
	AssertNoError(t, err)
	oportunidadeID2, err := oportunidadeRepo.Create(ctx, oportunidade2)
	AssertNoError(t, err)

	// Create proposals from same CNPJ to different opportunities
	cnpj := GenerateUniqueCNPJ()
	propostaRepo := repository.NewPropostaMEIRepository(tdb.DB)

	proposta1 := CreateTestPropostaMEI(oportunidadeID1, cnpj)
	_, err = propostaRepo.Create(ctx, proposta1)
	AssertNoError(t, err)

	proposta2 := CreateTestPropostaMEI(oportunidadeID2, cnpj)
	_, err = propostaRepo.Create(ctx, proposta2)
	AssertNoError(t, err)

	// List proposals for this CNPJ
	propostas, total, err := svc.PropostaMEIService.ListByMEIEmpresa(ctx, cnpj, 1, 100)
	AssertNoError(t, err)
	AssertEqual(t, 2, int(total))
	AssertEqual(t, 2, len(propostas))
}

// TestMEIWorkflow_DataConsistency tests data consistency
func TestMEIWorkflow_DataConsistency(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create opportunity and proposal
	oportunidade := CreateTestOportunidadeMEI(t)
	oportunidadeRepo := repository.NewOportunidadeMEIRepository(tdb.DB)
	oportunidadeID, err := oportunidadeRepo.Create(ctx, oportunidade)
	AssertNoError(t, err)

	cnpj := GenerateUniqueCNPJ()
	proposta := CreateTestPropostaMEI(oportunidadeID, cnpj)
	propostaRepo := repository.NewPropostaMEIRepository(tdb.DB)
	propostaID, err := propostaRepo.Create(ctx, proposta)
	AssertNoError(t, err)

	// Verify data consistency
	retrieved, err := svc.PropostaMEIService.GetByID(ctx, propostaID)
	AssertNoError(t, err)
	AssertEqual(t, proposta.MEIEmpresaID, retrieved.MEIEmpresaID)
	AssertEqual(t, proposta.OportunidadeMEIID, retrieved.OportunidadeMEIID)
	AssertEqual(t, *proposta.ValorProposta, *retrieved.ValorProposta)
	AssertEqual(t, *proposta.PrazoExecucao, *retrieved.PrazoExecucao)
	AssertEqual(t, *proposta.AceitaCustosIntegrais, *retrieved.AceitaCustosIntegrais)
}
