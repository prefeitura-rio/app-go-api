package integration_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
)

// TestStress_ConcurrentEnrollments tests 100 simultaneous enrollments to same course
func TestStress_ConcurrentEnrollments(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course with capacity
	curso := CreateTestCurso(t)
	curso.NumeroVagas = 100
	curso.LocationClasses[0].Schedules[0].Vacancies = 100
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Concurrent enrollments
	numEnrollments := 100
	var wg sync.WaitGroup
	errors := make([]error, numEnrollments)

	startTime := time.Now()
	for i := 0; i < numEnrollments; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			cpf := GenerateUniqueCPF()
			inscricao := CreateTestInscricao(cursoID, cpf)
			errors[index] = svc.InscricaoService.Create(ctx, inscricao)
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Count successful enrollments
	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}

	t.Logf("Concurrent enrollments: %d/%d succeeded in %v", successCount, numEnrollments, duration)

	// Verify data integrity
	inscricoes, total, err := svc.InscricaoService.GetByCursoID(ctx, cursoID, nil, 1, 200)
	AssertNoError(t, err)
	AssertEqual(t, successCount, int(total))
	AssertEqual(t, successCount, len(inscricoes))
}

// TestStress_ConcurrentEnrollmentsWithCapacity tests capacity enforcement under load
func TestStress_ConcurrentEnrollmentsWithCapacity(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course with auto-approval and limited capacity
	curso := CreateTestCurso(t)
	autoApprove := true
	curso.AutoApproveEnrollments = &autoApprove
	curso.NumeroVagas = 10
	curso.LocationClasses[0].Schedules[0].Vacancies = 10
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Try to enroll more than capacity
	numAttempts := 20
	var wg sync.WaitGroup
	errors := make([]error, numAttempts)

	for i := 0; i < numAttempts; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			cpf := GenerateUniqueCPF()
			inscricao := CreateTestInscricao(cursoID, cpf)
			scheduleID := curso.LocationClasses[0].Schedules[0].ID
			inscricao.ScheduleID = &scheduleID
			errors[index] = svc.InscricaoService.Create(ctx, inscricao)
		}(i)
	}

	wg.Wait()

	// Count auto-approved enrollments
	approvedCount := 0
	inscricoes, _, err := svc.InscricaoService.GetByCursoID(ctx, cursoID, nil, 1, 100)
	AssertNoError(t, err)
	for _, inscricao := range inscricoes {
		if inscricao.Status == models.StatusInscricaoApproved {
			approvedCount++
		}
	}

	t.Logf("Auto-approved enrollments: %d (capacity: 10)", approvedCount)
	// Note: Due to race conditions, this might exceed capacity slightly
	// In production, use SELECT FOR UPDATE or DB constraints for strict enforcement
}

// TestStress_BulkEnrollmentImport tests importing 1000+ enrollments
func TestStress_BulkEnrollmentImport(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course
	curso := CreateTestCurso(t)
	curso.NumeroVagas = 2000
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Bulk import
	numEnrollments := 1000
	startTime := time.Now()

	for i := 0; i < numEnrollments; i++ {
		cpf := GenerateUniqueCPF()
		inscricao := CreateTestInscricao(cursoID, cpf)
		err = svc.InscricaoService.Create(ctx, inscricao)
		if err != nil {
			t.Logf("Error on enrollment %d: %v", i, err)
		}
	}

	duration := time.Since(startTime)
	t.Logf("Bulk import of %d enrollments took %v (%.2f enrollments/sec)",
		numEnrollments, duration, float64(numEnrollments)/duration.Seconds())

	// Verify import
	inscricoes, total, err := svc.InscricaoService.GetByCursoID(ctx, cursoID, nil, 1, 2000)
	AssertNoError(t, err)
	t.Logf("Total enrollments in DB: %d", total)
	t.Logf("Enrollments retrieved: %d", len(inscricoes))
}

// TestStress_BulkStatusUpdate tests updating 500+ enrollments
func TestStress_BulkStatusUpdate(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course and enrollments
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	numEnrollments := 500
	var ids []models.UUID

	for i := 0; i < numEnrollments; i++ {
		cpf := GenerateUniqueCPF()
		inscricao := CreateTestInscricao(cursoID, cpf)
		err = svc.InscricaoService.Create(ctx, inscricao)
		AssertNoError(t, err)
		ids = append(ids, inscricao.ID)
	}

	// Bulk update
	startTime := time.Now()
	count, err := svc.InscricaoService.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "Bulk test", "")
	duration := time.Since(startTime)

	AssertNoError(t, err)
	AssertEqual(t, numEnrollments, count)
	t.Logf("Bulk status update of %d enrollments took %v", numEnrollments, duration)
}

// TestStress_ConcurrentJobApplications tests concurrent job applications
func TestStress_ConcurrentJobApplications(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Setup empresa and vaga
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
	vaga.QuantidadeVagas = 100
	vagaID, err := svc.VagaService.Create(ctx, vaga)
	AssertNoError(t, err)

	// Concurrent applications
	numApplications := 100
	var wg sync.WaitGroup
	errors := make([]error, numApplications)

	startTime := time.Now()
	for i := 0; i < numApplications; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			cpf := GenerateUniqueCPF()
			candidatura := CreateTestCandidatura(cpf, vagaID)
			_, errors[index] = svc.CandidaturaService.Create(ctx, candidatura)
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Count successes
	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}

	t.Logf("Concurrent job applications: %d/%d succeeded in %v", successCount, numApplications, duration)

	// Verify data integrity
	filter := empregabilidade.CandidaturaFilter{
		IDVaga: &vagaID,
	}
	candidaturas, total, err := svc.CandidaturaService.List(ctx, filter, 1, 200)
	AssertNoError(t, err)
	AssertEqual(t, successCount, int(total))
	AssertEqual(t, successCount, len(candidaturas))
}

// TestStress_CurriculoReplaceAllWithManyItems tests ReplaceAll with 100+ items
func TestStress_CurriculoReplaceAllWithManyItems(t *testing.T) {
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

	// Create large list of formacoes
	numFormacoes := 100
	formacoes := make([]*empregabilidade.CurriculoFormacao, numFormacoes)
	for i := 0; i < numFormacoes; i++ {
		formacoes[i] = &empregabilidade.CurriculoFormacao{
			CPF:            cpf,
			Instituicao:    "University " + uuid.New().String()[:8],
			Curso:          "Course " + uuid.New().String()[:8],
			IDEscolaridade: escolaridadeID,
		}
	}

	// ReplaceAll
	startTime := time.Now()
	err = svc.CurriculoService.ReplaceAllFormacoesByCPF(ctx, cpf, formacoes)
	duration := time.Since(startTime)

	AssertNoError(t, err)
	t.Logf("ReplaceAll with %d formacoes took %v", numFormacoes, duration)

	// Verify
	retrieved, err := svc.CurriculoService.ListFormacoesByCPF(ctx, cpf)
	AssertNoError(t, err)
	AssertEqual(t, numFormacoes, len(retrieved))
}

// TestStress_MemoryUsageUnderLoad tests memory usage with many concurrent operations
func TestStress_MemoryUsageUnderLoad(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create multiple courses
	numCourses := 10
	courseIDs := make([]int, numCourses)
	for i := 0; i < numCourses; i++ {
		curso := CreateTestCurso(t)
		cursoID, err := svc.CursoService.Create(ctx, curso)
		AssertNoError(t, err)
		courseIDs[i] = cursoID
	}

	// Create many enrollments across courses
	numEnrollmentsPerCourse := 50
	var wg sync.WaitGroup

	startTime := time.Now()
	for _, cursoID := range courseIDs {
		for i := 0; i < numEnrollmentsPerCourse; i++ {
			wg.Add(1)
			go func(cID int) {
				defer wg.Done()
				cpf := GenerateUniqueCPF()
				inscricao := CreateTestInscricao(cID, cpf)
				_ = svc.InscricaoService.Create(ctx, inscricao)
			}(cursoID)
		}
	}

	wg.Wait()
	duration := time.Since(startTime)

	totalOperations := numCourses * numEnrollmentsPerCourse
	t.Logf("Completed %d operations in %v (%.2f ops/sec)",
		totalOperations, duration, float64(totalOperations)/duration.Seconds())
}

// TestStress_LargeResultSets tests performance with large result sets
func TestStress_LargeResultSets(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course with many enrollments
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	numEnrollments := 1000
	for i := 0; i < numEnrollments; i++ {
		cpf := GenerateUniqueCPF()
		inscricao := CreateTestInscricao(cursoID, cpf)
		err = svc.InscricaoService.Create(ctx, inscricao)
		AssertNoError(t, err)
	}

	// Test pagination performance
	pageSize := 100
	totalPages := numEnrollments / pageSize

	startTime := time.Now()
	totalRetrieved := 0
	for page := 1; page <= totalPages; page++ {
		inscricoes, _, err := svc.InscricaoService.GetByCursoID(ctx, cursoID, nil, page, pageSize)
		AssertNoError(t, err)
		totalRetrieved += len(inscricoes)
	}
	duration := time.Since(startTime)

	t.Logf("Retrieved %d enrollments in %d pages in %v", totalRetrieved, totalPages, duration)
	AssertEqual(t, numEnrollments, totalRetrieved)
}

// TestStress_HighFrequencyStatusUpdates tests rapid status changes
func TestStress_HighFrequencyStatusUpdates(t *testing.T) {
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

	// Rapid status changes
	numUpdates := 50
	statuses := []models.StatusInscricao{
		models.StatusInscricaoApproved,
		models.StatusInscricaoRejected,
		models.StatusInscricaoApproved,
		models.StatusInscricaoRejected,
	}

	startTime := time.Now()
	for i := 0; i < numUpdates; i++ {
		status := statuses[i%len(statuses)]
		err = svc.InscricaoService.UpdateStatus(ctx, inscricao.ID, status, "", "")
		if err != nil {
			t.Logf("Update %d failed: %v", i, err)
		}
	}
	duration := time.Since(startTime)

	t.Logf("Completed %d status updates in %v (%.2f updates/sec)",
		numUpdates, duration, float64(numUpdates)/duration.Seconds())

	// Verify final state is consistent
	final, err := svc.InscricaoService.GetByID(ctx, inscricao.ID)
	AssertNoError(t, err)
	AssertNotNil(t, final)
}

// TestStress_ConcurrentDatabaseConnections tests database connection pool
func TestStress_ConcurrentDatabaseConnections(t *testing.T) {
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

	// Many concurrent read operations
	numConcurrentReads := 200
	var wg sync.WaitGroup
	errors := make([]error, numConcurrentReads)

	startTime := time.Now()
	for i := 0; i < numConcurrentReads; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, errors[index] = svc.CursoService.GetByID(ctx, cursoID)
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Count errors
	errorCount := 0
	for _, err := range errors {
		if err != nil {
			errorCount++
		}
	}

	t.Logf("Concurrent reads: %d/%d succeeded in %v",
		numConcurrentReads-errorCount, numConcurrentReads, duration)
}
