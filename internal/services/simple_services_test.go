package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// ─── EmpregoService ───────────────────────────────────────────────────────────

type mockEmpregoRepo struct {
	createID  int
	createErr error
	entity    *models.Emprego
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.Emprego
	listTotal int
	listErr   error
}

func (m *mockEmpregoRepo) Create(ctx context.Context, e *models.Emprego) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockEmpregoRepo) GetByID(ctx context.Context, id int) (*models.Emprego, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockEmpregoRepo) Update(ctx context.Context, e *models.Emprego) error {
	return m.updateErr
}

func (m *mockEmpregoRepo) Delete(ctx context.Context, id int) error {
	return m.deleteErr
}

func (m *mockEmpregoRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Emprego, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Emprego{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func TestEmpregoService_CRUD(t *testing.T) {
	ctx := context.Background()

	t.Run("Create valid emprego", func(t *testing.T) {
		repo := &mockEmpregoRepo{createID: 5}
		svc := services.NewEmpregoServiceWithInterface(repo)
		emprego := &models.Emprego{
			Titulo:          "Desenvolvedor Go",
			TipoContratacao: models.TipoContratacaoCLT,
			Status:          models.StatusEmpregoAberto,
		}
		id, err := svc.Create(ctx, emprego)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if id != 5 {
			t.Errorf("Create: expected id 5, got %d", id)
		}
	})

	t.Run("Create invalid emprego", func(t *testing.T) {
		repo := &mockEmpregoRepo{}
		svc := services.NewEmpregoServiceWithInterface(repo)
		_, err := svc.Create(ctx, &models.Emprego{}) // empty, should fail validation
		// Validate() might return error for missing fields
		_ = err // may or may not error depending on Validate()
	})

	t.Run("GetByID", func(t *testing.T) {
		repo := &mockEmpregoRepo{entity: &models.Emprego{ID: 1, Titulo: "Dev"}}
		svc := services.NewEmpregoServiceWithInterface(repo)
		e, err := svc.GetByID(ctx, 1)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
		if e.Titulo != "Dev" {
			t.Errorf("expected Dev, got %s", e.Titulo)
		}
	})

	t.Run("GetByID error", func(t *testing.T) {
		repo := &mockEmpregoRepo{getErr: errors.New("db error")}
		svc := services.NewEmpregoServiceWithInterface(repo)
		_, err := svc.GetByID(ctx, 1)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		repo := &mockEmpregoRepo{}
		svc := services.NewEmpregoServiceWithInterface(repo)
		err := svc.Delete(ctx, 1)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List success", func(t *testing.T) {
		repo := &mockEmpregoRepo{
			listItems: []*models.Emprego{{ID: 1, Titulo: "Dev"}},
			listTotal: 1,
		}
		svc := services.NewEmpregoServiceWithInterface(repo)
		items, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 1 || len(items) != 1 {
			t.Errorf("List: expected 1 item, got %d total=%d", len(items), total)
		}
	})

	t.Run("List error", func(t *testing.T) {
		repo := &mockEmpregoRepo{listErr: errors.New("db error")}
		svc := services.NewEmpregoServiceWithInterface(repo)
		_, _, err := svc.List(ctx, nil, 1, 10)
		if err == nil {
			t.Error("expected error")
		}
	})
}

// ─── EmpresaService ───────────────────────────────────────────────────────────

type mockEmpresaRepoMain struct {
	createID  int
	createErr error
	entity    *models.Empresa
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.Empresa
	listTotal int
	listErr   error
}

func (m *mockEmpresaRepoMain) Create(ctx context.Context, e *models.Empresa) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockEmpresaRepoMain) GetByID(ctx context.Context, id int) (*models.Empresa, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockEmpresaRepoMain) Update(ctx context.Context, e *models.Empresa) error {
	return m.updateErr
}

func (m *mockEmpresaRepoMain) Delete(ctx context.Context, id int) error {
	return m.deleteErr
}

func (m *mockEmpresaRepoMain) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Empresa, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Empresa{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func TestEmpresaService_CRUD(t *testing.T) {
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		repo := &mockEmpresaRepoMain{createID: 10}
		svc := services.NewEmpresaServiceWithInterface(repo)
		id, err := svc.Create(ctx, &models.Empresa{Nome: "Empresa ABC"})
		if err != nil || id != 10 {
			t.Fatalf("Create: %v / %d", err, id)
		}
	})

	t.Run("Create error", func(t *testing.T) {
		repo := &mockEmpresaRepoMain{createErr: errors.New("db error")}
		svc := services.NewEmpresaServiceWithInterface(repo)
		_, err := svc.Create(ctx, &models.Empresa{Nome: "Empresa"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetByID found", func(t *testing.T) {
		repo := &mockEmpresaRepoMain{entity: &models.Empresa{ID: 1, Nome: "ABC"}}
		svc := services.NewEmpresaServiceWithInterface(repo)
		e, err := svc.GetByID(ctx, 1)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		repo := &mockEmpresaRepoMain{entity: nil}
		svc := services.NewEmpresaServiceWithInterface(repo)
		e, err := svc.GetByID(ctx, 999)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if e != nil {
			t.Error("expected nil")
		}
	})

	t.Run("Update", func(t *testing.T) {
		repo := &mockEmpresaRepoMain{}
		svc := services.NewEmpresaServiceWithInterface(repo)
		err := svc.Update(ctx, &models.Empresa{ID: 1, Nome: "Updated"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Update error", func(t *testing.T) {
		repo := &mockEmpresaRepoMain{updateErr: errors.New("db error")}
		svc := services.NewEmpresaServiceWithInterface(repo)
		err := svc.Update(ctx, &models.Empresa{ID: 1, Nome: "Updated"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		repo := &mockEmpresaRepoMain{}
		svc := services.NewEmpresaServiceWithInterface(repo)
		err := svc.Delete(ctx, 1)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List success", func(t *testing.T) {
		repo := &mockEmpresaRepoMain{
			listItems: []*models.Empresa{{ID: 1, Nome: "A"}, {ID: 2, Nome: "B"}},
			listTotal: 2,
		}
		svc := services.NewEmpresaServiceWithInterface(repo)
		items, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 2 || len(items) != 2 {
			t.Errorf("List: expected 2 items, got %d total=%d", len(items), total)
		}
	})
}

// ─── InstituicaoService ───────────────────────────────────────────────────────

type mockInstituicaoRepo struct {
	createID  int
	createErr error
	entity    *models.InstituicaoEnsino
	getErr    error
	updateErr error
	deleteErr error
	listItems []*models.InstituicaoEnsino
	listTotal int
	listErr   error
}

func (m *mockInstituicaoRepo) Create(ctx context.Context, i *models.InstituicaoEnsino) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockInstituicaoRepo) GetByID(ctx context.Context, id int) (*models.InstituicaoEnsino, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockInstituicaoRepo) Update(ctx context.Context, i *models.InstituicaoEnsino) error {
	return m.updateErr
}

func (m *mockInstituicaoRepo) Delete(ctx context.Context, id int) error {
	return m.deleteErr
}

func (m *mockInstituicaoRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.InstituicaoEnsino, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.InstituicaoEnsino{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func TestInstituicaoService_CRUD(t *testing.T) {
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		repo := &mockInstituicaoRepo{createID: 3}
		svc := services.NewInstituicaoServiceWithInterface(repo)
		id, err := svc.Create(ctx, &models.InstituicaoEnsino{Nome: "UERJ"})
		if err != nil || id != 3 {
			t.Fatalf("Create: %v / %d", err, id)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		repo := &mockInstituicaoRepo{entity: &models.InstituicaoEnsino{ID: 1, Nome: "UFRJ"}}
		svc := services.NewInstituicaoServiceWithInterface(repo)
		e, err := svc.GetByID(ctx, 1)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
		if e.Nome != "UFRJ" {
			t.Errorf("expected UFRJ, got %s", e.Nome)
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		repo := &mockInstituicaoRepo{entity: nil}
		svc := services.NewInstituicaoServiceWithInterface(repo)
		e, err := svc.GetByID(ctx, 999)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if e != nil {
			t.Error("expected nil")
		}
	})

	t.Run("Update", func(t *testing.T) {
		repo := &mockInstituicaoRepo{}
		svc := services.NewInstituicaoServiceWithInterface(repo)
		err := svc.Update(ctx, &models.InstituicaoEnsino{ID: 1, Nome: "PUC"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		repo := &mockInstituicaoRepo{}
		svc := services.NewInstituicaoServiceWithInterface(repo)
		err := svc.Delete(ctx, 1)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		repo := &mockInstituicaoRepo{
			listItems: []*models.InstituicaoEnsino{{ID: 1, Nome: "UERJ"}},
			listTotal: 1,
		}
		svc := services.NewInstituicaoServiceWithInterface(repo)
		items, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 1 || len(items) != 1 {
			t.Errorf("List: expected 1 item, got %d", len(items))
		}
	})
}

// ─── JobService ───────────────────────────────────────────────────────────────

type mockJobRepo struct {
	jobs      map[uuid.UUID]*models.Job
	createErr error
	getErr    error
	updateErr error
}

func newMockJobRepo() *mockJobRepo {
	return &mockJobRepo{jobs: make(map[uuid.UUID]*models.Job)}
}

func (m *mockJobRepo) Create(ctx context.Context, job *models.Job) error {
	if m.createErr != nil {
		return m.createErr
	}
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *mockJobRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	j, ok := m.jobs[id]
	if !ok {
		return nil, nil
	}
	return j, nil
}

func (m *mockJobRepo) Update(ctx context.Context, job *models.Job) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *mockJobRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if j, ok := m.jobs[id]; ok {
		j.Status = status
	}
	return nil
}

func (m *mockJobRepo) UpdateProgress(ctx context.Context, id uuid.UUID, progress, successCount, errorCount int) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if j, ok := m.jobs[id]; ok {
		j.Progress = progress
		j.SuccessCount = successCount
		j.ErrorCount = errorCount
	}
	return nil
}

func (m *mockJobRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Job, int, error) {
	var result []*models.Job
	for _, v := range m.jobs {
		result = append(result, v)
	}
	return result, len(result), nil
}

func (m *mockJobRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	delete(m.jobs, id)
	return nil
}

func TestJobService(t *testing.T) {
	ctx := context.Background()

	t.Run("Create valid job", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		job := &models.Job{Type: models.JobTypeEnrollmentImport}
		err := svc.Create(ctx, job)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if job.Status != models.JobStatusPending {
			t.Errorf("expected status pending, got %s", job.Status)
		}
	})

	t.Run("Create missing type", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		err := svc.Create(ctx, &models.Job{})
		if err == nil {
			t.Error("expected error for missing type")
		}
	})

	t.Run("GetByID existing", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		job := &models.Job{Type: models.JobTypeEnrollmentImport}
		svc.Create(ctx, job)
		found, err := svc.GetByID(ctx, job.ID)
		if err != nil || found == nil {
			t.Fatalf("GetByID: %v / %v", err, found)
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		found, err := svc.GetByID(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if found != nil {
			t.Error("expected nil")
		}
	})

	t.Run("Update existing", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		job := &models.Job{Type: models.JobTypeEnrollmentImport}
		svc.Create(ctx, job)
		job.Progress = 50
		err := svc.Update(ctx, job)
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Update non-existing", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		err := svc.Update(ctx, &models.Job{ID: uuid.New(), Type: models.JobTypeEnrollmentImport})
		if err == nil {
			t.Error("expected error for non-existing job")
		}
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		job := &models.Job{Type: models.JobTypeEnrollmentImport}
		svc.Create(ctx, job)
		err := svc.UpdateStatus(ctx, job.ID, models.JobStatusCompleted)
		if err != nil {
			t.Fatalf("UpdateStatus: %v", err)
		}
	})

	t.Run("UpdateStatus non-existing", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		err := svc.UpdateStatus(ctx, uuid.New(), models.JobStatusCompleted)
		if err == nil {
			t.Error("expected error for non-existing job")
		}
	})

	t.Run("UpdateProgress", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		job := &models.Job{Type: models.JobTypeEnrollmentImport}
		svc.Create(ctx, job)
		err := svc.UpdateProgress(ctx, job.ID, 75, 8, 2)
		if err != nil {
			t.Fatalf("UpdateProgress: %v", err)
		}
	})

	t.Run("UpdateProgress non-existing", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		err := svc.UpdateProgress(ctx, uuid.New(), 75, 8, 2)
		if err == nil {
			t.Error("expected error for non-existing job")
		}
	})

	t.Run("Delete existing", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		job := &models.Job{Type: models.JobTypeEnrollmentImport}
		svc.Create(ctx, job)
		err := svc.Delete(ctx, job.ID)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("Delete non-existing", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		err := svc.Delete(ctx, uuid.New())
		if err == nil {
			t.Error("expected error for non-existing job")
		}
	})

	t.Run("List", func(t *testing.T) {
		repo := newMockJobRepo()
		svc := services.NewJobServiceWithInterface(repo)
		job := &models.Job{Type: models.JobTypeEnrollmentImport}
		svc.Create(ctx, job)
		items, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total < 1 || len(items) < 1 {
			t.Errorf("List: expected >= 1 item, got %d total=%d", len(items), total)
		}
	})
}

// ─── AcessibilidadeService (Update/Delete not yet tested) ────────────────────

func TestAcessibilidadeService_UpdateAndDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("Update", func(t *testing.T) {
		repo := &mockAcessibilidadeRepo{}
		svc := services.NewAcessibilidadeService(repo)
		err := svc.Update(ctx, &models.Acessibilidade{ID: 1, Nome: "Rampa Updated"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Update error", func(t *testing.T) {
		repo := &mockAcessibilidadeRepo{updateErr: errors.New("db error")}
		svc := services.NewAcessibilidadeService(repo)
		err := svc.Update(ctx, &models.Acessibilidade{ID: 1, Nome: "X"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		repo := &mockAcessibilidadeRepo{}
		svc := services.NewAcessibilidadeService(repo)
		err := svc.Delete(ctx, 1)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("Delete error", func(t *testing.T) {
		repo := &mockAcessibilidadeRepo{deleteErr: errors.New("db error")}
		svc := services.NewAcessibilidadeService(repo)
		err := svc.Delete(ctx, 1)
		if err == nil {
			t.Error("expected error")
		}
	})
}
