// Harness temporário: curso presencial publicado com 2 turmas com janela de
// inscrição, pra conferir o card da turma no superapp (layout do "Inscrições até").
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=app_db port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("open:", err)
		os.Exit(1)
	}
	repo := repository.NewCursoRepository(db)
	ctx := context.Background()
	p := func(v time.Time) *time.Time { return &v }
	pb := func(b bool) *bool { return &b }

	curso := &models.Curso{
		Titulo: "TESTE Card Turma (layout)", Descricao: "teste layout inscricoes ate",
		Modalidade: models.ModalidadePresencial, Status: models.StatusCursoPublished,
		Workload: "20 horas", TargetAudience: "Todos", OrgaoID: "1",
		EnrollmentStartDate: p(time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)),
		EnrollmentEndDate:   p(time.Date(2027, 3, 20, 12, 0, 0, 0, time.UTC)),
		IsVisible:           pb(true),
	}
	id, err := repo.Create(ctx, curso)
	if err != nil {
		fmt.Println("create curso:", err)
		os.Exit(1)
	}
	// Turma 1: janela aberta; Turma 2: janela já encerrada (end no passado).
	loc := []models.LocationClass{{
		CursoID: id, Address: "Rua Felipe Cardoso, 100 - Santa Cruz", Neighborhood: "Santa Cruz",
		Schedules: []models.CourseSchedule{
			{
				Vacancies: 30, DisplayOrder: 1,
				EnrollmentStartDate: p(time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)),
				EnrollmentEndDate:   p(time.Date(2027, 3, 20, 12, 0, 0, 0, time.UTC)),
				ClassStartDate:      time.Date(2027, 4, 1, 9, 0, 0, 0, time.UTC),
				ClassEndDate:        time.Date(2027, 5, 1, 12, 0, 0, 0, time.UTC),
				ClassTime:           "09:00-12:00", ClassDays: "Seg,Ter",
			},
			{
				Vacancies: 2, DisplayOrder: 2,
				EnrollmentStartDate: p(time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)),
				EnrollmentEndDate:   p(time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)),
				ClassStartDate:      time.Date(2026, 7, 28, 15, 0, 0, 0, time.UTC),
				ClassEndDate:        time.Date(2026, 7, 29, 16, 0, 0, 0, time.UTC),
				ClassTime:           "15:00-16:00", ClassDays: "Qua",
			},
		},
	}}
	if err := repo.CreateLocationClasses(ctx, loc); err != nil {
		fmt.Println("create loc:", err)
		os.Exit(1)
	}
	fmt.Printf("CURSO id=%d publicado, presencial, 2 turmas com enrollment_end_date\n", id)
}
