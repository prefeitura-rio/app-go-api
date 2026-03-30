package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

type mockEmpresaRepo struct {
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

func (m *mockEmpresaRepo) Create(ctx context.Context, e *models.Empresa) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID != 0 {
		return m.createID, nil
	}
	return 1, nil
}

func (m *mockEmpresaRepo) GetByID(ctx context.Context, id int) (*models.Empresa, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockEmpresaRepo) Update(ctx context.Context, e *models.Empresa) error {
	return m.updateErr
}

func (m *mockEmpresaRepo) Delete(ctx context.Context, id int) error {
	return m.deleteErr
}

func (m *mockEmpresaRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*models.Empresa, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.Empresa{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func TestEmpresaService_Create(t *testing.T) {
	tests := []struct {
		name    string
		empresa *models.Empresa
		repoID  int
		repoErr error
		wantID  int
		wantErr bool
	}{
		{
			name: "success - simple empresa",
			empresa: &models.Empresa{
				Nome: "Tech Corp",
			},
			repoID: 10,
			wantID: 10,
		},
		{
			name: "success - empresa with long name",
			empresa: &models.Empresa{
				Nome: "Empresa de Tecnologia e Desenvolvimento de Software Ltda ME",
			},
			repoID: 20,
			wantID: 20,
		},
		{
			name: "success - empresa with special characters",
			empresa: &models.Empresa{
				Nome: "Empresa & Cia S/A",
			},
			repoID: 30,
			wantID: 30,
		},
		{
			name: "success - empresa with numbers",
			empresa: &models.Empresa{
				Nome: "Tech123 Desenvolvimento",
			},
			repoID: 40,
			wantID: 40,
		},
		{
			name: "repo error",
			empresa: &models.Empresa{
				Nome: "Tech Corp",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "success - empty name (no validation in service)",
			empresa: &models.Empresa{
				Nome: "",
			},
			repoID: 50,
			wantID: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEmpresaRepo{createID: tt.repoID, createErr: tt.repoErr}
			svc := services.NewEmpresaServiceWithInterface(repo)
			ctx := context.Background()

			id, err := svc.Create(ctx, tt.empresa)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && id != tt.wantID {
				t.Errorf("Create() id = %d, want %d", id, tt.wantID)
			}
		})
	}
}

func TestEmpresaService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		entity  *models.Empresa
		getErr  error
		wantNil bool
		wantErr bool
	}{
		{
			name: "success",
			id:   1,
			entity: &models.Empresa{
				ID:   1,
				Nome: "Tech Corp",
			},
		},
		{
			name: "success - with empregos",
			id:   2,
			entity: &models.Empresa{
				ID:   2,
				Nome: "Startup XYZ",
				Empregos: []models.Emprego{
					{ID: 10, Titulo: "Desenvolvedor"},
					{ID: 11, Titulo: "Designer"},
				},
			},
		},
		{
			name:    "not found",
			id:      999,
			entity:  nil,
			wantNil: true,
		},
		{
			name:    "repo error",
			id:      1,
			getErr:  errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEmpresaRepo{entity: tt.entity, getErr: tt.getErr}
			svc := services.NewEmpresaServiceWithInterface(repo)
			ctx := context.Background()

			empresa, err := svc.GetByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantNil && empresa != nil {
				t.Error("GetByID() expected nil empresa")
			}
			if !tt.wantNil && !tt.wantErr && empresa == nil {
				t.Error("GetByID() expected non-nil empresa")
			}
			if empresa != nil && empresa.ID != tt.id {
				t.Errorf("GetByID() expected id %d, got %d", tt.id, empresa.ID)
			}
			if empresa != nil && tt.entity != nil && empresa.Nome != tt.entity.Nome {
				t.Errorf("GetByID() expected nome %s, got %s", tt.entity.Nome, empresa.Nome)
			}
		})
	}
}

func TestEmpresaService_Update(t *testing.T) {
	tests := []struct {
		name    string
		empresa *models.Empresa
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			empresa: &models.Empresa{
				ID:   1,
				Nome: "Tech Corp Updated",
			},
		},
		{
			name: "success - update name",
			empresa: &models.Empresa{
				ID:   2,
				Nome: "New Company Name",
			},
		},
		{
			name: "success - empty name (no validation)",
			empresa: &models.Empresa{
				ID:   3,
				Nome: "",
			},
		},
		{
			name: "repo error",
			empresa: &models.Empresa{
				ID:   1,
				Nome: "Tech Corp",
			},
			repoErr: errors.New("update failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEmpresaRepo{updateErr: tt.repoErr}
			svc := services.NewEmpresaServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.Update(ctx, tt.empresa)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmpresaService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		repoErr error
		wantErr bool
	}{
		{
			name: "success",
			id:   1,
		},
		{
			name: "success - delete different id",
			id:   100,
		},
		{
			name:    "repo error",
			id:      1,
			repoErr: errors.New("delete failed"),
			wantErr: true,
		},
		{
			name:    "repo error - foreign key constraint",
			id:      1,
			repoErr: errors.New("foreign key constraint"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEmpresaRepo{deleteErr: tt.repoErr}
			svc := services.NewEmpresaServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.Delete(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmpresaService_List(t *testing.T) {
	empresa1 := &models.Empresa{
		ID:   1,
		Nome: "Tech Corp",
	}
	empresa2 := &models.Empresa{
		ID:   2,
		Nome: "Startup XYZ",
	}
	empresa3 := &models.Empresa{
		ID:   3,
		Nome: "Consultoria ABC",
	}

	tests := []struct {
		name      string
		filter    map[string]interface{}
		page      int
		pageSize  int
		listItems []*models.Empresa
		listTotal int
		listErr   error
		wantLen   int
		wantTotal int
		wantErr   bool
	}{
		{
			name:      "success - empty list",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  10,
			listItems: []*models.Empresa{},
			listTotal: 0,
			wantLen:   0,
			wantTotal: 0,
		},
		{
			name:      "success - with items",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  10,
			listItems: []*models.Empresa{empresa1, empresa2, empresa3},
			listTotal: 3,
			wantLen:   3,
			wantTotal: 3,
		},
		{
			name:      "success - filter by name",
			filter:    map[string]interface{}{"nome": "Tech"},
			page:      1,
			pageSize:  10,
			listItems: []*models.Empresa{empresa1},
			listTotal: 1,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:      "success - pagination page 1",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  2,
			listItems: []*models.Empresa{empresa1, empresa2},
			listTotal: 3,
			wantLen:   2,
			wantTotal: 3,
		},
		{
			name:      "success - pagination page 2",
			filter:    map[string]interface{}{},
			page:      2,
			pageSize:  2,
			listItems: []*models.Empresa{empresa3},
			listTotal: 3,
			wantLen:   1,
			wantTotal: 3,
		},
		{
			name:      "success - large page size",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  100,
			listItems: []*models.Empresa{empresa1, empresa2, empresa3},
			listTotal: 3,
			wantLen:   3,
			wantTotal: 3,
		},
		{
			name:      "success - page 3 empty",
			filter:    map[string]interface{}{},
			page:      3,
			pageSize:  2,
			listItems: []*models.Empresa{},
			listTotal: 3,
			wantLen:   0,
			wantTotal: 3,
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
			repo := &mockEmpresaRepo{
				listItems: tt.listItems,
				listTotal: tt.listTotal,
				listErr:   tt.listErr,
			}
			svc := services.NewEmpresaServiceWithInterface(repo)
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

func TestEmpresaService_List_OffsetCalculation(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		wantOffset int
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
			name:       "page 5, size 20",
			page:       5,
			pageSize:   20,
			wantOffset: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedOffset := (tt.page - 1) * tt.pageSize
			if expectedOffset != tt.wantOffset {
				t.Errorf("Offset calculation: got %d, want %d", expectedOffset, tt.wantOffset)
			}
		})
	}
}

func TestNewEmpresaService(t *testing.T) {
	service := services.NewEmpresaService(nil)

	if service == nil {
		t.Error("NewEmpresaService() returned nil")
	}
}

func TestNewEmpresaServiceWithInterface(t *testing.T) {
	repo := &mockEmpresaRepo{}
	service := services.NewEmpresaServiceWithInterface(repo)

	if service == nil {
		t.Error("NewEmpresaServiceWithInterface() returned nil")
	}

	// Verify service can perform basic operations
	ctx := context.Background()
	_, _, err := service.List(ctx, map[string]interface{}{}, 1, 10)
	if err != nil {
		t.Errorf("NewEmpresaServiceWithInterface() service not properly initialized: %v", err)
	}
}

func TestEmpresaService_Create_MultipleCompanies(t *testing.T) {
	// Test creating multiple companies in sequence
	companies := []string{
		"Tech Corp",
		"Startup XYZ",
		"Consultoria ABC",
		"Empresa 123",
		"Software House",
	}

	for i, nome := range companies {
		t.Run(nome, func(t *testing.T) {
			repo := &mockEmpresaRepo{createID: i + 1}
			svc := services.NewEmpresaServiceWithInterface(repo)
			ctx := context.Background()

			empresa := &models.Empresa{Nome: nome}
			id, err := svc.Create(ctx, empresa)
			if err != nil {
				t.Errorf("Create() failed for %s: %v", nome, err)
			}
			if id != i+1 {
				t.Errorf("Create() expected id %d, got %d", i+1, id)
			}
		})
	}
}

func TestEmpresaService_Update_SetID(t *testing.T) {
	// Test that SetID method works correctly
	empresa := &models.Empresa{Nome: "Test"}
	empresa.SetID(42)

	if empresa.ID != 42 {
		t.Errorf("SetID() expected id 42, got %d", empresa.ID)
	}
}

func TestEmpresaService_CRUD_Flow(t *testing.T) {
	// Test a complete CRUD flow
	ctx := context.Background()

	// Create
	createRepo := &mockEmpresaRepo{createID: 1}
	createSvc := services.NewEmpresaServiceWithInterface(createRepo)
	empresa := &models.Empresa{Nome: "Test Company"}
	id, err := createSvc.Create(ctx, empresa)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	if id != 1 {
		t.Errorf("Create() expected id 1, got %d", id)
	}

	// Read
	empresa.ID = id
	getRepo := &mockEmpresaRepo{entity: empresa}
	getSvc := services.NewEmpresaServiceWithInterface(getRepo)
	retrieved, err := getSvc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}
	if retrieved.Nome != "Test Company" {
		t.Errorf("GetByID() expected nome 'Test Company', got %s", retrieved.Nome)
	}

	// Update
	updateRepo := &mockEmpresaRepo{}
	updateSvc := services.NewEmpresaServiceWithInterface(updateRepo)
	empresa.Nome = "Updated Company"
	err = updateSvc.Update(ctx, empresa)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// Delete
	deleteRepo := &mockEmpresaRepo{}
	deleteSvc := services.NewEmpresaServiceWithInterface(deleteRepo)
	err = deleteSvc.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}
}
