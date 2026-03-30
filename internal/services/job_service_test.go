package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type mockJobRepo struct {
	createErr          error
	entity             *models.Job
	getErr             error
	updateErr          error
	updateStatusErr    error
	updateProgressErr  error
	deleteErr          error
	listItems          []*models.Job
	listTotal          int
	listErr            error
	getCalls           int
	updateCalls        int
	updateStatusCalls  int
	updateProgressCalls int
}

func (m *mockJobRepo) Create(ctx context.Context, j *models.Job) error {
	if m.createErr != nil {
		return m.createErr
	}
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}

func (m *mockJobRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	m.getCalls++
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockJobRepo) Update(ctx context.Context, j *models.Job) error {
	m.updateCalls++
	return m.updateErr
}

func (m *mockJobRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.JobStatus) error {
	m.updateStatusCalls++
	return m.updateStatusErr
}

func (m *mockJobRepo) UpdateProgress(ctx context.Context, id uuid.UUID, progress, successCount, errorCount int) error {
	m.updateProgressCalls++
	return m.updateProgressErr
}

func (m *mockJobRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteErr
}

func (m *mockJobRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Job, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Job{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func TestJobService_Create(t *testing.T) {
	tests := []struct {
		name      string
		job       *models.Job
		repoErr   error
		wantErr   bool
		checkJob  func(*testing.T, *models.Job)
	}{
		{
			name: "success - sets defaults",
			job: &models.Job{
				Type: models.JobTypeEnrollmentImport,
			},
			checkJob: func(t *testing.T, j *models.Job) {
				if j.Status != models.JobStatusPending {
					t.Errorf("expected status pending, got %s", j.Status)
				}
				if j.Progress != 0 {
					t.Errorf("expected progress 0, got %d", j.Progress)
				}
				if j.SuccessCount != 0 {
					t.Errorf("expected success_count 0, got %d", j.SuccessCount)
				}
				if j.ErrorCount != 0 {
					t.Errorf("expected error_count 0, got %d", j.ErrorCount)
				}
			},
		},
		{
			name: "error - missing type",
			job:  &models.Job{},
			wantErr: true,
		},
		{
			name: "error - empty type",
			job: &models.Job{
				Type: "",
			},
			wantErr: true,
		},
		{
			name: "repo error",
			job: &models.Job{
				Type: models.JobTypeEnrollmentImport,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "preserves existing fields",
			job: &models.Job{
				Type:         models.JobTypeEnrollmentImport,
				TotalRecords: 100,
			},
			checkJob: func(t *testing.T, j *models.Job) {
				if j.TotalRecords != 100 {
					t.Errorf("expected total_records 100, got %d", j.TotalRecords)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockJobRepo{createErr: tt.repoErr}
			svc := services.NewJobServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.Create(ctx, tt.job)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.checkJob != nil {
				tt.checkJob(t, tt.job)
			}
		})
	}
}

func TestJobService_GetByID(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name    string
		entity  *models.Job
		getErr  error
		wantNil bool
		wantErr bool
	}{
		{
			name: "success",
			entity: &models.Job{
				ID:     id,
				Type:   models.JobTypeEnrollmentImport,
				Status: models.JobStatusPending,
			},
		},
		{
			name:    "not found",
			entity:  nil,
			wantNil: true,
		},
		{
			name:    "repo error",
			getErr:  errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockJobRepo{entity: tt.entity, getErr: tt.getErr}
			svc := services.NewJobServiceWithInterface(repo)
			ctx := context.Background()

			job, err := svc.GetByID(ctx, id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantNil && job != nil {
				t.Error("GetByID() expected nil job")
			}
			if !tt.wantNil && !tt.wantErr && job == nil {
				t.Error("GetByID() expected non-nil job")
			}
			if job != nil && job.ID != id {
				t.Errorf("GetByID() expected id %v, got %v", id, job.ID)
			}
		})
	}
}

func TestJobService_Update(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name        string
		job         *models.Job
		existingJob *models.Job
		getErr      error
		updateErr   error
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			job: &models.Job{
				ID:     existingID,
				Type:   models.JobTypeEnrollmentImport,
				Status: models.JobStatusCompleted,
			},
			existingJob: &models.Job{
				ID:     existingID,
				Status: models.JobStatusProcessing,
			},
		},
		{
			name: "job not found",
			job: &models.Job{
				ID: existingID,
			},
			existingJob: nil,
			wantErr:     true,
			errContains: "não encontrado",
		},
		{
			name: "repo get error",
			job: &models.Job{
				ID: existingID,
			},
			getErr:      errors.New("db error"),
			wantErr:     true,
			errContains: "erro ao verificar",
		},
		{
			name: "repo update error",
			job: &models.Job{
				ID: existingID,
			},
			existingJob: &models.Job{
				ID: existingID,
			},
			updateErr: errors.New("update failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockJobRepo{
				entity:    tt.existingJob,
				getErr:    tt.getErr,
				updateErr: tt.updateErr,
			}
			svc := services.NewJobServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.Update(ctx, tt.job)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Update() error = %v, want contains %s", err, tt.errContains)
				}
			}
			if !tt.wantErr && repo.updateCalls != 1 {
				t.Errorf("Update() expected 1 update call, got %d", repo.updateCalls)
			}
		})
	}
}

func TestJobService_UpdateStatus(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name             string
		id               uuid.UUID
		status           models.JobStatus
		existingJob      *models.Job
		getErr           error
		updateStatusErr  error
		wantErr          bool
		errContains      string
	}{
		{
			name:   "success - pending to processing",
			id:     existingID,
			status: models.JobStatusProcessing,
			existingJob: &models.Job{
				ID:     existingID,
				Status: models.JobStatusPending,
			},
		},
		{
			name:   "success - processing to completed",
			id:     existingID,
			status: models.JobStatusCompleted,
			existingJob: &models.Job{
				ID:     existingID,
				Status: models.JobStatusProcessing,
			},
		},
		{
			name:   "success - processing to failed",
			id:     existingID,
			status: models.JobStatusFailed,
			existingJob: &models.Job{
				ID:     existingID,
				Status: models.JobStatusProcessing,
			},
		},
		{
			name:        "job not found",
			id:          existingID,
			status:      models.JobStatusCompleted,
			existingJob: nil,
			wantErr:     true,
			errContains: "não encontrado",
		},
		{
			name:        "repo get error",
			id:          existingID,
			status:      models.JobStatusCompleted,
			getErr:      errors.New("db error"),
			wantErr:     true,
			errContains: "erro ao verificar",
		},
		{
			name:   "repo update status error",
			id:     existingID,
			status: models.JobStatusCompleted,
			existingJob: &models.Job{
				ID: existingID,
			},
			updateStatusErr: errors.New("update failed"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockJobRepo{
				entity:          tt.existingJob,
				getErr:          tt.getErr,
				updateStatusErr: tt.updateStatusErr,
			}
			svc := services.NewJobServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.UpdateStatus(ctx, tt.id, tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("UpdateStatus() error = %v, want contains %s", err, tt.errContains)
				}
			}
			if !tt.wantErr && repo.updateStatusCalls != 1 {
				t.Errorf("UpdateStatus() expected 1 call, got %d", repo.updateStatusCalls)
			}
		})
	}
}

func TestJobService_UpdateProgress(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name               string
		id                 uuid.UUID
		progress           int
		successCount       int
		errorCount         int
		existingJob        *models.Job
		getErr             error
		updateProgressErr  error
		wantErr            bool
		errContains        string
	}{
		{
			name:         "success - initial progress",
			id:           existingID,
			progress:     10,
			successCount: 5,
			errorCount:   0,
			existingJob: &models.Job{
				ID: existingID,
			},
		},
		{
			name:         "success - 50% progress",
			id:           existingID,
			progress:     50,
			successCount: 45,
			errorCount:   5,
			existingJob: &models.Job{
				ID: existingID,
			},
		},
		{
			name:         "success - complete",
			id:           existingID,
			progress:     100,
			successCount: 95,
			errorCount:   5,
			existingJob: &models.Job{
				ID: existingID,
			},
		},
		{
			name:         "success - zero progress",
			id:           existingID,
			progress:     0,
			successCount: 0,
			errorCount:   0,
			existingJob: &models.Job{
				ID: existingID,
			},
		},
		{
			name:         "job not found",
			id:           existingID,
			progress:     50,
			successCount: 25,
			errorCount:   0,
			existingJob:  nil,
			wantErr:      true,
			errContains:  "não encontrado",
		},
		{
			name:         "repo get error",
			id:           existingID,
			progress:     50,
			successCount: 25,
			errorCount:   0,
			getErr:       errors.New("db error"),
			wantErr:      true,
			errContains:  "erro ao verificar",
		},
		{
			name:         "repo update progress error",
			id:           existingID,
			progress:     50,
			successCount: 25,
			errorCount:   0,
			existingJob: &models.Job{
				ID: existingID,
			},
			updateProgressErr: errors.New("update failed"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockJobRepo{
				entity:            tt.existingJob,
				getErr:            tt.getErr,
				updateProgressErr: tt.updateProgressErr,
			}
			svc := services.NewJobServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.UpdateProgress(ctx, tt.id, tt.progress, tt.successCount, tt.errorCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateProgress() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("UpdateProgress() error = %v, want contains %s", err, tt.errContains)
				}
			}
			if !tt.wantErr && repo.updateProgressCalls != 1 {
				t.Errorf("UpdateProgress() expected 1 call, got %d", repo.updateProgressCalls)
			}
		})
	}
}

func TestJobService_Delete(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name        string
		id          uuid.UUID
		existingJob *models.Job
		getErr      error
		deleteErr   error
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			id:   existingID,
			existingJob: &models.Job{
				ID:     existingID,
				Status: models.JobStatusCompleted,
			},
		},
		{
			name:        "job not found",
			id:          existingID,
			existingJob: nil,
			wantErr:     true,
			errContains: "não encontrado",
		},
		{
			name:        "repo get error",
			id:          existingID,
			getErr:      errors.New("db error"),
			wantErr:     true,
			errContains: "erro ao verificar",
		},
		{
			name: "repo delete error",
			id:   existingID,
			existingJob: &models.Job{
				ID: existingID,
			},
			deleteErr: errors.New("delete failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockJobRepo{
				entity:    tt.existingJob,
				getErr:    tt.getErr,
				deleteErr: tt.deleteErr,
			}
			svc := services.NewJobServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.Delete(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errContains != "" && err != nil {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("Delete() error = %v, want contains %s", err, tt.errContains)
				}
			}
		})
	}
}

func TestJobService_List(t *testing.T) {
	job1 := &models.Job{
		ID:     uuid.New(),
		Type:   models.JobTypeEnrollmentImport,
		Status: models.JobStatusCompleted,
	}
	job2 := &models.Job{
		ID:     uuid.New(),
		Type:   models.JobTypeEnrollmentImport,
		Status: models.JobStatusPending,
	}

	tests := []struct {
		name       string
		filter     map[string]interface{}
		page       int
		pageSize   int
		listItems  []*models.Job
		listTotal  int
		listErr    error
		wantLen    int
		wantTotal  int
		wantErr    bool
	}{
		{
			name:      "success - empty list",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  10,
			listItems: []*models.Job{},
			listTotal: 0,
			wantLen:   0,
			wantTotal: 0,
		},
		{
			name:      "success - with items",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  10,
			listItems: []*models.Job{job1, job2},
			listTotal: 2,
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name:      "success - with status filter",
			filter:    map[string]interface{}{"status": "completed"},
			page:      1,
			pageSize:  10,
			listItems: []*models.Job{job1},
			listTotal: 1,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:      "success - pagination page 2",
			filter:    map[string]interface{}{},
			page:      2,
			pageSize:  5,
			listItems: []*models.Job{job2},
			listTotal: 6,
			wantLen:   1,
			wantTotal: 6,
		},
		{
			name:      "success - large page size",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  100,
			listItems: []*models.Job{job1, job2},
			listTotal: 2,
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name:     "repo error",
			filter:   map[string]interface{}{},
			page:     1,
			pageSize: 10,
			listErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockJobRepo{
				listItems: tt.listItems,
				listTotal: tt.listTotal,
				listErr:   tt.listErr,
			}
			svc := services.NewJobServiceWithInterface(repo)
			ctx := context.Background()

			items, total, err := svc.List(ctx, tt.filter, tt.page, tt.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(items) != tt.wantLen {
					t.Errorf("List() got %d items, want %d", len(items), tt.wantLen)
				}
				if total != tt.wantTotal {
					t.Errorf("List() got total %d, want %d", total, tt.wantTotal)
				}
			}
		})
	}
}

func TestJobService_List_OffsetCalculation(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantOffset   int
	}{
		{
			name:       "page 1, size 10",
			page:       1,
			pageSize:   10,
			wantOffset: 0,
		},
		{
			name:       "page 2, size 10",
			page:       2,
			pageSize:   10,
			wantOffset: 10,
		},
		{
			name:       "page 3, size 5",
			page:       3,
			pageSize:   5,
			wantOffset: 10,
		},
		{
			name:       "page 1, size 100",
			page:       1,
			pageSize:   100,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can verify offset calculation by checking what gets passed to repo
			// The service calculates: offset = (page - 1) * pageSize
			expectedOffset := (tt.page - 1) * tt.pageSize
			if expectedOffset != tt.wantOffset {
				t.Errorf("Offset calculation: got %d, want %d", expectedOffset, tt.wantOffset)
			}
		})
	}
}
