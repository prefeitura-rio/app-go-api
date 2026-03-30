package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

// TestEnrollmentWorkflow_CompleteFlow tests the complete enrollment workflow
func TestEnrollmentWorkflow_CompleteFlow(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Phase 1: Create Course
	curso := CreateTestCurso(t)
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)
	AssertNotNil(t, cursoID)

	// Phase 2: Create Enrollment
	cpf := GenerateUniqueCPF()
	inscricao := CreateTestInscricao(cursoID, cpf)
	err = svc.InscricaoService.Create(ctx, inscricao)
	AssertNoError(t, err)
	AssertNotNil(t, inscricao.ID)

	// Verify initial status
	AssertEqual(t, models.StatusInscricaoPending, inscricao.Status)

	// Phase 3: Update Status to Approved
	err = svc.InscricaoService.UpdateStatus(ctx, inscricao.ID, models.StatusInscricaoApproved, "", "")
	AssertNoError(t, err)

	// Verify status change
	updatedInscricao, err := svc.InscricaoService.GetByID(ctx, inscricao.ID)
	AssertNoError(t, err)
	AssertEqual(t, models.StatusInscricaoApproved, updatedInscricao.Status)

	// Phase 4: Verify email would be sent (mocked in test)
	// In real implementation, email service should be called asynchronously
	// Here we just verify the workflow completed successfully
}

// TestEnrollmentWorkflow_MultipleEnrollments tests multiple enrollments to same course
func TestEnrollmentWorkflow_MultipleEnrollments(t *testing.T) {
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

	// Create multiple enrollments
	numEnrollments := 5
	for i := 0; i < numEnrollments; i++ {
		cpf := GenerateUniqueCPF()
		inscricao := CreateTestInscricao(cursoID, cpf)
		err = svc.InscricaoService.Create(ctx, inscricao)
		AssertNoError(t, err)
	}

	// Verify all enrollments exist
	inscricoes, total, err := svc.InscricaoService.GetByCursoID(ctx, cursoID, nil, 1, 100)
	AssertNoError(t, err)
	AssertEqual(t, numEnrollments, int(total))
	AssertEqual(t, numEnrollments, len(inscricoes))
}

// TestEnrollmentWorkflow_DuplicateEnrollment tests duplicate enrollment prevention
func TestEnrollmentWorkflow_DuplicateEnrollment(t *testing.T) {
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

	// Try to create duplicate enrollment
	inscricao2 := CreateTestInscricao(cursoID, cpf)
	err = svc.InscricaoService.Create(ctx, inscricao2)
	AssertError(t, err) // Should fail with duplicate error
}

// TestEnrollmentWorkflow_StatusTransitions tests valid status transitions
func TestEnrollmentWorkflow_StatusTransitions(t *testing.T) {
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

	// Test transition: Pending -> Approved
	err = svc.InscricaoService.UpdateStatus(ctx, inscricao.ID, models.StatusInscricaoApproved, "", "")
	AssertNoError(t, err)

	// Test transition: Approved -> Rejected
	err = svc.InscricaoService.UpdateStatus(ctx, inscricao.ID, models.StatusInscricaoRejected, "Test rejection", "")
	AssertNoError(t, err)

	// Verify final status
	final, err := svc.InscricaoService.GetByID(ctx, inscricao.ID)
	AssertNoError(t, err)
	AssertEqual(t, models.StatusInscricaoRejected, final.Status)
}

// TestEnrollmentWorkflow_BulkStatusUpdate tests bulk status updates
func TestEnrollmentWorkflow_BulkStatusUpdate(t *testing.T) {
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

	// Create multiple enrollments
	numEnrollments := 10
	var ids []models.UUID
	for i := 0; i < numEnrollments; i++ {
		cpf := GenerateUniqueCPF()
		inscricao := CreateTestInscricao(cursoID, cpf)
		err = svc.InscricaoService.Create(ctx, inscricao)
		AssertNoError(t, err)
		ids = append(ids, inscricao.ID)
	}

	// Bulk update to approved
	count, err := svc.InscricaoService.UpdateMultipleStatus(ctx, ids, models.StatusInscricaoApproved, "Bulk approval", "")
	AssertNoError(t, err)
	AssertEqual(t, numEnrollments, count)

	// Verify all were updated
	for _, id := range ids {
		inscricao, err := svc.InscricaoService.GetByID(ctx, id)
		AssertNoError(t, err)
		AssertEqual(t, models.StatusInscricaoApproved, inscricao.Status)
	}
}

// TestEnrollmentWorkflow_AutoApproval tests auto-approval workflow
func TestEnrollmentWorkflow_AutoApproval(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course with auto-approval enabled
	curso := CreateTestCurso(t)
	autoApprove := true
	curso.AutoApproveEnrollments = &autoApprove
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Create enrollment
	cpf := GenerateUniqueCPF()
	inscricao := CreateTestInscricao(cursoID, cpf)
	err = svc.InscricaoService.Create(ctx, inscricao)
	AssertNoError(t, err)

	// Verify auto-approved status
	AssertEqual(t, models.StatusInscricaoApproved, inscricao.Status)
}

// TestEnrollmentWorkflow_ScheduleChange tests schedule change workflow
func TestEnrollmentWorkflow_ScheduleChange(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course with multiple schedules
	curso := CreateTestCurso(t)
	secondSchedule := models.CourseSchedule{
		ID:             models.GenerateUUID(),
		Vacancies:      30,
		ClassTime:      "14:00-18:00",
		ClassDays:      "Segunda a Sexta",
		ClassStartDate: time.Now().Add(72 * time.Hour), // 3 days from now
		ClassEndDate:   time.Now().Add(40 * 24 * time.Hour),
	}
	acceptingEnrollments := true
	secondSchedule.AcceptingEnrollments = &acceptingEnrollments
	curso.LocationClasses[0].Schedules = append(curso.LocationClasses[0].Schedules, secondSchedule)

	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Create enrollment with first schedule
	cpf := GenerateUniqueCPF()
	inscricao := CreateTestInscricao(cursoID, cpf)
	firstScheduleID := curso.LocationClasses[0].Schedules[0].ID
	inscricao.ScheduleID = &firstScheduleID
	err = svc.InscricaoService.Create(ctx, inscricao)
	AssertNoError(t, err)

	// Change to second schedule
	changeRequest := &models.ScheduleChangeRequest{
		ScheduleID: &secondSchedule.ID,
		EnrolledUnit: &models.EnrolledUnit{
			Schedules: []models.EnrolledSchedule{
				{
					ClassStartDate: secondSchedule.ClassStartDate.Format(time.RFC3339),
				},
			},
		},
	}
	updatedInscricao, err := svc.InscricaoService.ChangeSchedule(ctx, inscricao.ID, cpf, changeRequest)
	AssertNoError(t, err)
	AssertEqual(t, secondSchedule.ID, *updatedInscricao.ScheduleID)
}

// TestEnrollmentWorkflow_ClosedCourse tests enrollment to closed course
func TestEnrollmentWorkflow_ClosedCourse(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create closed course
	curso := CreateTestCurso(t)
	curso.Status = models.StatusCursoClosed
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Try to enroll
	cpf := GenerateUniqueCPF()
	inscricao := CreateTestInscricao(cursoID, cpf)
	err = svc.InscricaoService.Create(ctx, inscricao)
	AssertError(t, err) // Should fail - course is closed
}

// TestEnrollmentWorkflow_EnrollmentPeriodValidation tests enrollment period checks
func TestEnrollmentWorkflow_EnrollmentPeriodValidation(t *testing.T) {
	tdb := GetTestDB(t)
	if tdb == nil {
		return
	}
	defer tdb.Cleanup(context.Background())

	svc := SetupTestServices(t, tdb.DB)
	ctx := context.Background()

	// Create course with future enrollment period
	curso := CreateTestCurso(t)
	futureStart := time.Now().Add(48 * time.Hour)
	curso.EnrollmentStart = &futureStart
	cursoID, err := svc.CursoService.Create(ctx, curso)
	AssertNoError(t, err)

	// Try to enroll before period starts
	cpf := GenerateUniqueCPF()
	inscricao := CreateTestInscricao(cursoID, cpf)
	err = svc.InscricaoService.Create(ctx, inscricao)
	AssertError(t, err) // Should fail - enrollment period hasn't started
}

// TestEnrollmentWorkflow_DataConsistency tests data consistency across operations
func TestEnrollmentWorkflow_DataConsistency(t *testing.T) {
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

	// Create enrollment
	cpf := GenerateUniqueCPF()
	inscricao := CreateTestInscricao(cursoID, cpf)
	err = svc.InscricaoService.Create(ctx, inscricao)
	AssertNoError(t, err)

	// Verify data consistency
	retrieved, err := svc.InscricaoService.GetByID(ctx, inscricao.ID)
	AssertNoError(t, err)
	AssertEqual(t, inscricao.CPF, retrieved.CPF)
	AssertEqual(t, inscricao.CursoID, retrieved.CursoID)
	AssertEqual(t, inscricao.Name, retrieved.Name)
	AssertEqual(t, inscricao.Email, retrieved.Email)
	AssertEqual(t, inscricao.Phone, retrieved.Phone)

	// Verify timestamps are set
	AssertNotNil(t, retrieved.EnrolledAt)
	AssertNotNil(t, retrieved.UpdatedAt)
}
