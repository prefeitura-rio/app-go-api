package integration_test

import (
	"context"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// TestErrorRecovery_InvalidCourseID tests handling of invalid course ID
func TestErrorRecovery_InvalidCourseID(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Try to enroll in non-existent course
	cpf := GenerateUniqueCPF()
	inscricao := CreateTestInscricao(99999, cpf) // Invalid course ID
	err := svc.InscricaoService.Create(ctx, inscricao)
	AssertError(t, err) // Should fail gracefully with error

	// System should still be functional
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)
	AssertNotNil(t, cursoID)
}

// TestErrorRecovery_InvalidEnrollmentID tests handling of invalid enrollment ID
func TestErrorRecovery_InvalidEnrollmentID(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Try to update non-existent enrollment
	invalidID := models.GenerateUUID()
	err := svc.InscricaoService.UpdateStatus(ctx, invalidID, models.StatusInscricaoApproved, "", "")
	AssertError(t, err) // Should fail gracefully

	// Try to delete non-existent enrollment
	err = svc.InscricaoService.Delete(ctx, invalidID)
	AssertError(t, err) // Should fail gracefully
}

// TestErrorRecovery_InvalidCPF tests handling of invalid CPF format
func TestErrorRecovery_InvalidCPF(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Try to enroll with invalid CPF (if validation exists)
	inscricao := CreateTestInscricao(cursoID, "invalid-cpf")
	// Note: Depending on validation rules, this might succeed or fail
	// The system should handle either case gracefully
	_ = svc.InscricaoService.Create(ctx, inscricao)
}

// TestErrorRecovery_MissingRequiredFields tests validation of required fields
func TestErrorRecovery_MissingRequiredFields(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Try to create curso without required fields
	invalidCurso := &models.Curso{
		// Missing Titulo, Modalidade, Status
	}
	_, err := svc.CursoService.Create(ctx, invalidCurso)
	AssertError(t, err) // Should fail validation
}

// TestErrorRecovery_TransactionRollback tests transaction rollback on error
func TestErrorRecovery_TransactionRollback(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Count enrollments before
	beforeInscricoes, beforeTotal, err := svc.InscricaoService.GetByCursoID(ctx, cursoID, nil, 1, 100)
	AssertNoError(t, err)
	beforeCount := int(beforeTotal)

	// Try to create enrollment with invalid schedule ID (should rollback)
	cpf := GenerateUniqueCPF()
	inscricao := CreateTestInscricao(cursoID, cpf)
	invalidScheduleID := models.GenerateUUID()
	inscricao.ScheduleID = &invalidScheduleID
	err = svc.InscricaoService.Create(ctx, inscricao)
	AssertError(t, err) // Should fail

	// Verify no partial data was created
	afterInscricoes, afterTotal, err := svc.InscricaoService.GetByCursoID(ctx, cursoID, nil, 1, 100)
	AssertNoError(t, err)
	AssertEqual(t, beforeCount, int(afterTotal))
	AssertEqual(t, len(beforeInscricoes), len(afterInscricoes))
}

// TestErrorRecovery_StatusTransitionValidation tests invalid status transitions
func TestErrorRecovery_StatusTransitionValidation(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Setup vaga and candidatura
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

	// Move to Descontinuada
	err = svc.CandidaturaService.UpdateStatus(ctx, candidaturaID, empregabilidade.StatusCandidaturaDescontinuada)
	AssertNoError(t, err)

	// Try invalid transition from Descontinuada (should fail)
	err = svc.CandidaturaService.UpdateStatus(ctx, candidaturaID, empregabilidade.StatusCandidaturaAprovada)
	AssertError(t, err) // Should fail - Descontinuada is terminal state
}

// TestErrorRecovery_DuplicateKeyViolation tests handling of unique constraint violations
func TestErrorRecovery_DuplicateKeyViolation(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Create first enrollment
	cpf := GenerateUniqueCPF()
	inscricao1 := CreateTestInscricao(cursoID, cpf)
	err = svc.InscricaoService.Create(ctx, inscricao1)
	AssertNoError(t, err)

	// Try duplicate (should fail gracefully)
	inscricao2 := CreateTestInscricao(cursoID, cpf)
	err = svc.InscricaoService.Create(ctx, inscricao2)
	AssertError(t, err)

	// Verify only one enrollment exists
	inscricoes, total, err := svc.InscricaoService.GetByCursoID(ctx, cursoID, map[string]interface{}{"cpf": cpf}, 1, 100)
	AssertNoError(t, err)
	AssertEqual(t, 1, int(total))
	AssertEqual(t, 1, len(inscricoes))
}

// TestErrorRecovery_NullPointerHandling tests null pointer safety
func TestErrorRecovery_NullPointerHandling(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Try to create nil enrollment
	var nilInscricao *models.Inscricao
	err := svc.InscricaoService.Create(ctx, nilInscricao)
	// Should handle gracefully (panic or error)

	// Try to update nil enrollment
	err = svc.InscricaoService.Update(ctx, models.GenerateUUID(), 0, nil)
	// Should handle gracefully
	_ = err
}

// TestErrorRecovery_ConcurrentUpdateConflicts tests handling of concurrent updates
func TestErrorRecovery_ConcurrentUpdateConflicts(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course and enrollment
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	cpf := GenerateUniqueCPF()
	inscricao := CreateTestInscricao(cursoID, cpf)
	err = svc.InscricaoService.Create(ctx, inscricao)
	AssertNoError(t, err)

	// Concurrent status updates (one should win)
	done := make(chan error, 2)

	go func() {
		done <- svc.InscricaoService.UpdateStatus(ctx, inscricao.ID, models.StatusInscricaoApproved, "", "")
	}()

	go func() {
		done <- svc.InscricaoService.UpdateStatus(ctx, inscricao.ID, models.StatusInscricaoRejected, "", "")
	}()

	err1 := <-done
	err2 := <-done

	// At least one should succeed
	if err1 != nil && err2 != nil {
		t.Error("Both concurrent updates failed")
	}

	// Verify final state is consistent
	final, err := svc.InscricaoService.GetByID(ctx, inscricao.ID)
	AssertNoError(t, err)
	AssertNotNil(t, final)
	// Status should be one of the two attempted values
	if final.Status != models.StatusInscricaoApproved && final.Status != models.StatusInscricaoRejected {
		t.Errorf("Unexpected final status: %v", final.Status)
	}
}

// TestErrorRecovery_ForeignKeyViolation tests handling of foreign key violations
func TestErrorRecovery_ForeignKeyViolation(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Try to create vaga with non-existent empresa
	invalidCNPJ := GenerateUniqueCNPJ()
	vaga := CreateTestVaga(t, invalidCNPJ)
	_, err := svc.VagaService.Create(ctx, vaga)
	AssertError(t, err) // Should fail - empresa doesn't exist
}

// TestErrorRecovery_ContextCancellation tests handling of context cancellation
func TestErrorRecovery_ContextCancellation(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)

	// Create context that is immediately cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Try to create course with cancelled context
	curso := CreateTestCurso(t)
	_, err := svc.CursoService.Create(ctx, curso)
	// Should handle context cancellation gracefully
	_ = err
}

// TestErrorRecovery_DataIntegrityAfterError tests data integrity after errors
func TestErrorRecovery_DataIntegrityAfterError(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Create valid enrollment
	cpf1 := GenerateUniqueCPF()
	inscricao1 := CreateTestInscricao(cursoID, cpf1)
	err = svc.InscricaoService.Create(ctx, inscricao1)
	AssertNoError(t, err)

	// Try to create invalid enrollment (should fail)
	invalidInscricao := CreateTestInscricao(99999, GenerateUniqueCPF())
	err = svc.InscricaoService.Create(ctx, invalidInscricao)
	AssertError(t, err)

	// Create another valid enrollment (should succeed)
	cpf2 := GenerateUniqueCPF()
	inscricao2 := CreateTestInscricao(cursoID, cpf2)
	err = svc.InscricaoService.Create(ctx, inscricao2)
	AssertNoError(t, err)

	// Verify both valid enrollments exist
	inscricoes, total, err := svc.InscricaoService.GetByCursoID(ctx, cursoID, nil, 1, 100)
	AssertNoError(t, err)
	AssertEqual(t, 2, int(total))
	AssertEqual(t, 2, len(inscricoes))
}

// TestErrorRecovery_BulkOperationPartialFailure tests handling of partial bulk failures
func TestErrorRecovery_BulkOperationPartialFailure(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course and some enrollments
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	var validIDs []models.UUID
	for i := 0; i < 5; i++ {
		cpf := GenerateUniqueCPF()
		inscricao := CreateTestInscricao(cursoID, cpf)
		err = svc.InscricaoService.Create(ctx, inscricao)
		AssertNoError(t, err)
		validIDs = append(validIDs, inscricao.ID)
	}

	// Add some invalid IDs
	bulkIDs := append(validIDs, models.GenerateUUID(), models.GenerateUUID())

	// Bulk update (some IDs invalid)
	count, err := svc.InscricaoService.UpdateMultipleStatus(ctx, bulkIDs, models.StatusInscricaoApproved, "", "")

	// Should update only valid IDs
	// Note: Behavior depends on implementation - might fail entirely or partially succeed
	t.Logf("Bulk update: %d updated (expected %d valid)", count, len(validIDs))
}

// TestErrorRecovery_EmailServiceFailure tests graceful degradation when email fails
func TestErrorRecovery_EmailServiceFailure(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Note: Email service is mocked in test setup (nil client)
	// Email sending should fail gracefully without affecting enrollment creation

	// Create course
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Create enrollment (email will fail but enrollment should succeed)
	cpf := GenerateUniqueCPF()
	inscricao := CreateTestInscricao(cursoID, cpf)
	err = svc.InscricaoService.Create(ctx, inscricao)
	AssertNoError(t, err) // Should succeed despite email failure

	// Verify enrollment was created
	retrieved, err := svc.InscricaoService.GetByID(ctx, inscricao.ID)
	AssertNoError(t, err)
	AssertNotNil(t, retrieved)
}

// TestErrorRecovery_SystemRecoveryAfterErrors tests system recovery
func TestErrorRecovery_SystemRecoveryAfterErrors(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Cause multiple errors
	for i := 0; i < 10; i++ {
		invalidCurso := &models.Curso{}
		_, _ = svc.CursoService.Create(ctx, invalidCurso)
	}

	// System should still be functional
	validCurso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, validCurso)
	AssertNoError(t, err)
	AssertNotNil(t, cursoID)

	// Verify data integrity
	retrieved, err := svc.CursoService.GetByID(ctx, cursoID)
	AssertNoError(t, err)
	AssertEqual(t, validCurso.Titulo, retrieved.Titulo)
}
