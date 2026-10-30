package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/prefeitura-rio/app-go-api/internal/apperrors"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/repository"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// Os testes deste arquivo exercitam o PUT do curso (serviço + repositório)
// contra um Postgres real: atualizar um curso nunca pode deixar inscrição com
// schedule_id de uma turma que deixou de existir (DEV-119). Gated by
// RUN_REPOSITORY_INTEGRATION.

func scheduleIntegrationTime(day int) time.Time {
	return time.Date(2026, time.March, day, 9, 0, 0, 0, time.UTC)
}

func newIntegrationRemoteSchedule(vacancies int) models.RemoteSchedule {
	return models.RemoteSchedule{Vacancies: vacancies}
}

func newIntegrationCourseSchedule(vacancies int) models.CourseSchedule {
	return models.CourseSchedule{
		Vacancies:      vacancies,
		ClassStartDate: scheduleIntegrationTime(10),
		ClassEndDate:   scheduleIntegrationTime(20),
		ClassTime:      "09h-12h",
		ClassDays:      "Seg,Qua",
	}
}

// createScheduleIntegrationCourse cria o curso pelo serviço, registra a
// limpeza e devolve o curso como o GET o devolve (com os ids das turmas).
func createScheduleIntegrationCourse(t *testing.T, db *gorm.DB, curso *models.Curso) (*services.CursoService, *repository.CursoRepository, *models.Curso) {
	t.Helper()
	ctx := context.Background()
	repo := repository.NewCursoRepository(db)
	svc := services.NewCursoService(repo)

	curso.Status = models.StatusCursoDraft
	id, err := svc.Create(ctx, curso)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		db.Where("curso_id = ?", id).Delete(&models.Inscricao{})
		db.Where("curso_id = ?", id).Delete(&models.LocationClass{})
		db.Where("curso_id = ?", id).Delete(&models.RemoteClass{})
		db.Delete(&models.Curso{}, id)
	})

	saved, err := repo.GetByID(ctx, id)
	if err != nil || saved == nil {
		t.Fatalf("GetByID: err=%v saved=%v", err, saved)
	}
	return svc, repo, saved
}

func enrollIntegration(t *testing.T, db *gorm.DB, cursoID int, scheduleID uuid.UUID, status models.StatusInscricao) uuid.UUID {
	t.Helper()
	sid := scheduleID
	inscricao := &models.Inscricao{
		CursoID:    cursoID,
		CPF:        "00000000000",
		Name:       "Pessoa Sintética de Teste",
		Status:     status,
		ScheduleID: &sid,
	}
	if err := db.Create(inscricao).Error; err != nil {
		t.Fatalf("criar inscrição: %v", err)
	}
	return inscricao.ID
}

// putCurso monta o payload como o Portal Interno o envia ao editar: o curso
// lido do GET, com as alterações aplicadas por mutate.
func putCurso(t *testing.T, svc *services.CursoService, saved *models.Curso, mutate func(c *models.Curso)) error {
	t.Helper()
	payload := *saved
	payload.Inscricoes = nil
	mutate(&payload)
	return svc.Update(context.Background(), &payload)
}

func countSchedules(t *testing.T, db *gorm.DB, model interface{}, id uuid.UUID) int64 {
	t.Helper()
	var n int64
	if err := db.Model(model).Where("id = ?", id).Count(&n).Error; err != nil {
		t.Fatalf("contar turma: %v", err)
	}
	return n
}

func assertConflict(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("esperava erro de conflito, PUT aceito")
	}
	if !apperrors.IsKind(err, apperrors.KindConflict) {
		t.Fatalf("esperava apperrors.KindConflict, veio: %v", err)
	}
}

func TestCursoUpdate_RemoteScheduleWithEnrollments_Integration(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}

	newCourse := func(t *testing.T) (*services.CursoService, *models.Curso) {
		svc, _, saved := createScheduleIntegrationCourse(t, db, &models.Curso{
			Titulo:     "[INTEGRATION] DEV-119 remoto",
			Modalidade: models.ModalidadeRemoto,
			RemoteClass: &models.RemoteClass{Schedules: []models.RemoteSchedule{
				newIntegrationRemoteSchedule(10), newIntegrationRemoteSchedule(20),
			}},
		})
		if saved.RemoteClass == nil || len(saved.RemoteClass.Schedules) != 2 {
			t.Fatalf("esperava turma remota com 2 turmas, veio %+v", saved.RemoteClass)
		}
		return svc, saved
	}

	t.Run("PUT que reenvia as turmas com id preserva o schedule_id das inscrições", func(t *testing.T) {
		svc, saved := newCourse(t)
		enrolled := saved.RemoteClass.Schedules[0].ID
		enrollIntegration(t, db, saved.ID, enrolled, models.StatusInscricaoApproved)

		err := putCurso(t, svc, saved, func(c *models.Curso) {
			rc := *saved.RemoteClass
			rc.Schedules = append([]models.RemoteSchedule(nil), saved.RemoteClass.Schedules...)
			rc.Schedules[0].Vacancies = 15
			c.RemoteClass = &rc
		})
		if err != nil {
			t.Fatalf("PUT: %v", err)
		}
		if countSchedules(t, db, &models.RemoteSchedule{}, enrolled) != 1 {
			t.Fatal("turma com inscrição deixou de existir")
		}
	})

	t.Run("remover turma remota com inscrição é recusado e nada muda", func(t *testing.T) {
		svc, saved := newCourse(t)
		enrolled := saved.RemoteClass.Schedules[0].ID
		enrollIntegration(t, db, saved.ID, enrolled, models.StatusInscricaoCancelled)

		err := putCurso(t, svc, saved, func(c *models.Curso) {
			rc := *saved.RemoteClass
			rc.Schedules = []models.RemoteSchedule{saved.RemoteClass.Schedules[1]}
			c.RemoteClass = &rc
		})
		assertConflict(t, err)
		if countSchedules(t, db, &models.RemoteSchedule{}, enrolled) != 1 {
			t.Fatal("turma com inscrição foi apagada")
		}
	})

	t.Run("reenviar turma sem id no lugar da que tem inscrição é recusado", func(t *testing.T) {
		svc, saved := newCourse(t)
		enrolled := saved.RemoteClass.Schedules[0].ID
		enrollIntegration(t, db, saved.ID, enrolled, models.StatusInscricaoPending)

		err := putCurso(t, svc, saved, func(c *models.Curso) {
			rc := *saved.RemoteClass
			rc.Schedules = append([]models.RemoteSchedule(nil), saved.RemoteClass.Schedules...)
			for i := range rc.Schedules {
				rc.Schedules[i].ID = uuid.Nil
			}
			c.RemoteClass = &rc
		})
		assertConflict(t, err)
		if countSchedules(t, db, &models.RemoteSchedule{}, enrolled) != 1 {
			t.Fatal("turma com inscrição foi apagada")
		}
	})

	t.Run("tirar a turma remota inteira com inscrição é recusado", func(t *testing.T) {
		svc, saved := newCourse(t)
		enrolled := saved.RemoteClass.Schedules[1].ID
		enrollIntegration(t, db, saved.ID, enrolled, models.StatusInscricaoConcluded)

		err := putCurso(t, svc, saved, func(c *models.Curso) {
			c.RemoteClass = nil
		})
		assertConflict(t, err)
		if countSchedules(t, db, &models.RemoteSchedule{}, enrolled) != 1 {
			t.Fatal("turma com inscrição foi apagada")
		}
	})

	t.Run("remover turma remota sem inscrição continua permitido", func(t *testing.T) {
		svc, saved := newCourse(t)
		enrollIntegration(t, db, saved.ID, saved.RemoteClass.Schedules[1].ID, models.StatusInscricaoApproved)
		removed := saved.RemoteClass.Schedules[0].ID

		err := putCurso(t, svc, saved, func(c *models.Curso) {
			rc := *saved.RemoteClass
			rc.Schedules = []models.RemoteSchedule{saved.RemoteClass.Schedules[1]}
			c.RemoteClass = &rc
		})
		if err != nil {
			t.Fatalf("PUT: %v", err)
		}
		if countSchedules(t, db, &models.RemoteSchedule{}, removed) != 0 {
			t.Fatal("turma sem inscrição não foi removida")
		}
	})
}

func TestCursoUpdate_LocationScheduleWithEnrollments_Integration(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}

	newCourse := func(t *testing.T) (*services.CursoService, *models.Curso) {
		svc, _, saved := createScheduleIntegrationCourse(t, db, &models.Curso{
			Titulo:     "[INTEGRATION] DEV-119 presencial",
			Modalidade: models.ModalidadePresencial,
			LocationClasses: []models.LocationClass{
				{Address: "Rua Sintética, 1", Neighborhood: "Centro", Schedules: []models.CourseSchedule{
					newIntegrationCourseSchedule(10), newIntegrationCourseSchedule(20),
				}},
				{Address: "Rua Sintética, 2", Neighborhood: "Tijuca", Schedules: []models.CourseSchedule{
					newIntegrationCourseSchedule(30),
				}},
			},
		})
		if len(saved.LocationClasses) != 2 {
			t.Fatalf("esperava 2 locais, veio %d", len(saved.LocationClasses))
		}
		return svc, saved
	}
	locationByAddress := func(t *testing.T, c *models.Curso, address string) models.LocationClass {
		t.Helper()
		for _, loc := range c.LocationClasses {
			if loc.Address == address {
				return loc
			}
		}
		t.Fatalf("local %q não encontrado", address)
		return models.LocationClass{}
	}

	t.Run("PUT que reenvia locais e turmas com id preserva o schedule_id", func(t *testing.T) {
		svc, saved := newCourse(t)
		loc1 := locationByAddress(t, saved, "Rua Sintética, 1")
		enrolled := loc1.Schedules[0].ID
		enrollIntegration(t, db, saved.ID, enrolled, models.StatusInscricaoApproved)

		err := putCurso(t, svc, saved, func(c *models.Curso) {
			c.LocationClasses[0].Address = c.LocationClasses[0].Address + " "
		})
		if err != nil {
			t.Fatalf("PUT: %v", err)
		}
		if countSchedules(t, db, &models.CourseSchedule{}, enrolled) != 1 {
			t.Fatal("turma com inscrição deixou de existir")
		}
	})

	t.Run("remover turma presencial com inscrição é recusado", func(t *testing.T) {
		svc, saved := newCourse(t)
		loc1 := locationByAddress(t, saved, "Rua Sintética, 1")
		enrolled := loc1.Schedules[0].ID
		enrollIntegration(t, db, saved.ID, enrolled, models.StatusInscricaoRejected)

		err := putCurso(t, svc, saved, func(c *models.Curso) {
			for i := range c.LocationClasses {
				if c.LocationClasses[i].ID == loc1.ID {
					c.LocationClasses[i].Schedules = []models.CourseSchedule{loc1.Schedules[1]}
				}
			}
		})
		assertConflict(t, err)
		if countSchedules(t, db, &models.CourseSchedule{}, enrolled) != 1 {
			t.Fatal("turma com inscrição foi apagada")
		}
	})

	t.Run("remover local cuja turma tem inscrição é recusado", func(t *testing.T) {
		svc, saved := newCourse(t)
		loc2 := locationByAddress(t, saved, "Rua Sintética, 2")
		enrolled := loc2.Schedules[0].ID
		enrollIntegration(t, db, saved.ID, enrolled, models.StatusInscricaoApproved)

		err := putCurso(t, svc, saved, func(c *models.Curso) {
			c.LocationClasses = []models.LocationClass{locationByAddress(t, saved, "Rua Sintética, 1")}
		})
		assertConflict(t, err)
		if countSchedules(t, db, &models.CourseSchedule{}, enrolled) != 1 {
			t.Fatal("turma com inscrição foi apagada")
		}
	})

	t.Run("turma com inscrição que muda de local mantém o id", func(t *testing.T) {
		svc, saved := newCourse(t)
		loc1 := locationByAddress(t, saved, "Rua Sintética, 1")
		loc2 := locationByAddress(t, saved, "Rua Sintética, 2")
		moved := loc2.Schedules[0]
		enrollIntegration(t, db, saved.ID, moved.ID, models.StatusInscricaoApproved)

		// O local 2 sai do curso e a turma dele passa para o local 1.
		err := putCurso(t, svc, saved, func(c *models.Curso) {
			keep := loc1
			keep.Schedules = append(append([]models.CourseSchedule(nil), loc1.Schedules...), moved)
			c.LocationClasses = []models.LocationClass{keep}
		})
		if err != nil {
			t.Fatalf("PUT: %v", err)
		}
		var sched models.CourseSchedule
		if err := db.First(&sched, "id = ?", moved.ID).Error; err != nil {
			t.Fatalf("turma movida deixou de existir: %v", err)
		}
		if sched.LocationID != loc1.ID {
			t.Fatalf("turma movida ficou no local %s, esperava %s", sched.LocationID, loc1.ID)
		}
	})

	t.Run("remover local sem inscrições continua permitido", func(t *testing.T) {
		svc, saved := newCourse(t)
		loc1 := locationByAddress(t, saved, "Rua Sintética, 1")
		loc2 := locationByAddress(t, saved, "Rua Sintética, 2")
		enrollIntegration(t, db, saved.ID, loc1.Schedules[0].ID, models.StatusInscricaoApproved)

		err := putCurso(t, svc, saved, func(c *models.Curso) {
			c.LocationClasses = []models.LocationClass{loc1}
		})
		if err != nil {
			t.Fatalf("PUT: %v", err)
		}
		if countSchedules(t, db, &models.CourseSchedule{}, loc2.Schedules[0].ID) != 0 {
			t.Fatal("turma do local removido continua existindo")
		}
	})
}

// Uma turma de outro curso enviada no payload não pode ser alterada nem
// levada para este curso.
func TestCursoUpdate_ForeignScheduleID_Integration(t *testing.T) {
	db := getDBForIntegration(t)
	if db == nil {
		return
	}

	_, _, other := createScheduleIntegrationCourse(t, db, &models.Curso{
		Titulo:     "[INTEGRATION] DEV-119 outro curso",
		Modalidade: models.ModalidadeRemoto,
		RemoteClass: &models.RemoteClass{Schedules: []models.RemoteSchedule{
			newIntegrationRemoteSchedule(10),
		}},
	})
	foreign := other.RemoteClass.Schedules[0]

	svc, _, saved := createScheduleIntegrationCourse(t, db, &models.Curso{
		Titulo:     "[INTEGRATION] DEV-119 curso editado",
		Modalidade: models.ModalidadeRemoto,
		RemoteClass: &models.RemoteClass{Schedules: []models.RemoteSchedule{
			newIntegrationRemoteSchedule(10),
		}},
	})

	err := putCurso(t, svc, saved, func(c *models.Curso) {
		rc := *saved.RemoteClass
		intruder := foreign
		intruder.Vacancies = 999
		rc.Schedules = append(append([]models.RemoteSchedule(nil), saved.RemoteClass.Schedules...), intruder)
		c.RemoteClass = &rc
	})
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}

	var got models.RemoteSchedule
	if err := db.First(&got, "id = ?", foreign.ID).Error; err != nil {
		t.Fatalf("turma do outro curso sumiu: %v", err)
	}
	if got.Vacancies != foreign.Vacancies || got.RemoteClassID != foreign.RemoteClassID {
		t.Fatalf("turma do outro curso foi alterada: vagas %d (antes %d), turma remota %s (antes %s)",
			got.Vacancies, foreign.Vacancies, got.RemoteClassID, foreign.RemoteClassID)
	}

	var created []models.RemoteSchedule
	if err := db.Where("remote_class_id = ?", saved.RemoteClass.ID).Find(&created).Error; err != nil {
		t.Fatalf("listar turmas do curso editado: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("esperava a turma enviada criada como nova no curso editado (2 turmas), veio %d", len(created))
	}
	for _, sched := range created {
		if sched.ID == foreign.ID {
			t.Fatal("a turma do outro curso passou para o curso editado")
		}
	}
}
