package repository_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// TestInscricaoVacancyLock_Integration fires simultaneous auto-approved
// enrollments and turma changes against a real Postgres and checks that no turma
// ends with more approved enrollments than its vacancies (DEV-199, pentest
// vuln-0004). Gated by RUN_REPOSITORY_INTEGRATION.
func TestInscricaoVacancyLock_Integration(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}
	ctx := context.Background()

	const (
		vacanciesA = 3
		vacanciesB = 2
		requests   = 12
	)
	autoApprove := true
	classStart := time.Now().Add(30 * 24 * time.Hour).UTC().Truncate(time.Second)
	schedule := func(vacancies int) models.CourseSchedule {
		return models.CourseSchedule{
			Vacancies:      vacancies,
			ClassStartDate: classStart,
			ClassEndDate:   classStart.Add(30 * 24 * time.Hour),
			ClassTime:      "09h-12h",
			ClassDays:      "Seg,Qua",
		}
	}
	_, cursoRepo, curso := createScheduleIntegrationCourse(t, db, &models.Curso{
		Titulo:                 "[INTEGRATION] DEV-199 vagas concorrentes",
		Modalidade:             models.ModalidadePresencial,
		AutoApproveEnrollments: &autoApprove,
		LocationClasses: []models.LocationClass{{
			Address:      "Rua de Teste, 123 - Centro",
			Neighborhood: "Centro",
			Schedules:    []models.CourseSchedule{schedule(vacanciesA), schedule(vacanciesB)},
		}},
	})
	if len(curso.LocationClasses) != 1 || len(curso.LocationClasses[0].Schedules) != 2 {
		t.Fatalf("expected 1 location with 2 schedules, got %+v", curso.LocationClasses)
	}
	scheduleA, scheduleB := curso.LocationClasses[0].Schedules[0].ID, curso.LocationClasses[0].Schedules[1].ID
	if curso.LocationClasses[0].Schedules[0].Vacancies != vacanciesA {
		scheduleA, scheduleB = scheduleB, scheduleA
	}

	inscricaoRepo := repository.NewInscricaoRepository(db)
	svc := services.NewInscricaoServiceWithInterface(inscricaoRepo, cursoRepo, nil, nil, nil,
		&config.AppConfig{Enrollment: config.EnrollmentSettings{ScheduleChangeDeadlineHours: 48}})

	// 1. Simultaneous auto-approved enrollments into turma A.
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make(chan error, requests)
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sid := scheduleA
			inscricao := &models.Inscricao{
				CursoID:    curso.ID,
				CPF:        fmt.Sprintf("990000%05d", i),
				Name:       "Pessoa Sintética de Teste",
				ScheduleID: &sid,
			}
			<-start
			errs <- svc.CreateManual(ctx, inscricao)
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("CreateManual: %v", err)
		}
	}

	approvedA, err := cursoRepo.CountEnrollmentsByScheduleID(ctx, scheduleA)
	if err != nil {
		t.Fatalf("count A: %v", err)
	}
	if approvedA != vacanciesA {
		t.Fatalf("turma A: %d approved enrollments for %d vacancies", approvedA, vacanciesA)
	}
	var pendingA int64
	db.Model(&models.Inscricao{}).Where("schedule_id = ? AND status = ?", scheduleA, models.StatusInscricaoPending).Count(&pendingA)
	if pendingA != requests-vacanciesA {
		t.Fatalf("turma A: %d pending enrollments, want %d", pendingA, requests-vacanciesA)
	}

	// 2. The approved enrollments of turma A move to turma B at the same time.
	var approved []models.Inscricao
	if err := db.Where("schedule_id = ? AND status = ?", scheduleA, models.StatusInscricaoApproved).Find(&approved).Error; err != nil {
		t.Fatalf("list approved: %v", err)
	}
	start = make(chan struct{})
	changeErrs := make(chan error, len(approved))
	for _, in := range approved {
		wg.Add(1)
		go func(id uuid.UUID, cpf string) {
			defer wg.Done()
			target := scheduleB
			<-start
			_, err := svc.ChangeSchedule(ctx, id, cpf, &models.ScheduleChangeRequest{ScheduleID: &target})
			changeErrs <- err
		}(in.ID, in.CPF)
	}
	close(start)
	wg.Wait()
	close(changeErrs)

	moved, full := 0, 0
	for err := range changeErrs {
		switch {
		case err == nil:
			moved++
		case strings.Contains(err.Error(), "não há vagas"):
			full++
		default:
			t.Fatalf("ChangeSchedule: %v", err)
		}
	}
	approvedB, err := cursoRepo.CountEnrollmentsByScheduleID(ctx, scheduleB)
	if err != nil {
		t.Fatalf("count B: %v", err)
	}
	if moved != vacanciesB || full != vacanciesA-vacanciesB || approvedB != vacanciesB {
		t.Fatalf("turma B: moved=%d full=%d approved=%d, want moved=%d full=%d approved=%d",
			moved, full, approvedB, vacanciesB, vacanciesA-vacanciesB, vacanciesB)
	}
}
