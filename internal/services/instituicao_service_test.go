package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

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
	if m.createID != 0 {
		return m.createID, nil
	}
	return 1, nil
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

func TestInstituicaoService_Create(t *testing.T) {
	tests := []struct {
		name        string
		instituicao *models.InstituicaoEnsino
		repoID      int
		repoErr     error
		wantID      int
		wantErr     bool
	}{
		{
			name: "success - university",
			instituicao: &models.InstituicaoEnsino{
				Nome: "Universidade Federal do Rio de Janeiro",
			},
			repoID: 10,
			wantID: 10,
		},
		{
			name: "success - technical school",
			instituicao: &models.InstituicaoEnsino{
				Nome: "Escola Técnica Estadual",
			},
			repoID: 20,
			wantID: 20,
		},
		{
			name: "success - private college",
			instituicao: &models.InstituicaoEnsino{
				Nome: "Faculdade Privada XYZ",
			},
			repoID: 30,
			wantID: 30,
		},
		{
			name: "success - long name",
			instituicao: &models.InstituicaoEnsino{
				Nome: "Centro Universitário de Tecnologia e Desenvolvimento Profissional do Estado do Rio de Janeiro",
			},
			repoID: 40,
			wantID: 40,
		},
		{
			name: "success - name with special characters",
			instituicao: &models.InstituicaoEnsino{
				Nome: "Colégio & Instituto S/A",
			},
			repoID: 50,
			wantID: 50,
		},
		{
			name: "success - name with numbers",
			instituicao: &models.InstituicaoEnsino{
				Nome: "Instituto 123 de Ensino",
			},
			repoID: 60,
			wantID: 60,
		},
		{
			name: "repo error",
			instituicao: &models.InstituicaoEnsino{
				Nome: "UFRJ",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "success - empty name (no validation in service)",
			instituicao: &models.InstituicaoEnsino{
				Nome: "",
			},
			repoID: 70,
			wantID: 70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInstituicaoRepo{createID: tt.repoID, createErr: tt.repoErr}
			svc := services.NewInstituicaoServiceWithInterface(repo)
			ctx := context.Background()

			id, err := svc.Create(ctx, tt.instituicao)
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

func TestInstituicaoService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		entity  *models.InstituicaoEnsino
		getErr  error
		wantNil bool
		wantErr bool
	}{
		{
			name: "success",
			id:   1,
			entity: &models.InstituicaoEnsino{
				ID:   1,
				Nome: "UFRJ",
			},
		},
		{
			name: "success - with cursos",
			id:   2,
			entity: &models.InstituicaoEnsino{
				ID:   2,
				Nome: "UERJ",
				Cursos: []models.Curso{
					{ID: 10, Titulo: "Engenharia"},
					{ID: 11, Titulo: "Medicina"},
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
			repo := &mockInstituicaoRepo{entity: tt.entity, getErr: tt.getErr}
			svc := services.NewInstituicaoServiceWithInterface(repo)
			ctx := context.Background()

			instituicao, err := svc.GetByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantNil && instituicao != nil {
				t.Error("GetByID() expected nil instituicao")
			}
			if !tt.wantNil && !tt.wantErr && instituicao == nil {
				t.Error("GetByID() expected non-nil instituicao")
			}
			if instituicao != nil && instituicao.ID != tt.id {
				t.Errorf("GetByID() expected id %d, got %d", tt.id, instituicao.ID)
			}
			if instituicao != nil && tt.entity != nil && instituicao.Nome != tt.entity.Nome {
				t.Errorf("GetByID() expected nome %s, got %s", tt.entity.Nome, instituicao.Nome)
			}
		})
	}
}

func TestInstituicaoService_Update(t *testing.T) {
	tests := []struct {
		name        string
		instituicao *models.InstituicaoEnsino
		repoErr     error
		wantErr     bool
	}{
		{
			name: "success",
			instituicao: &models.InstituicaoEnsino{
				ID:   1,
				Nome: "UFRJ - Updated",
			},
		},
		{
			name: "success - update name",
			instituicao: &models.InstituicaoEnsino{
				ID:   2,
				Nome: "New Institution Name",
			},
		},
		{
			name: "success - empty name (no validation)",
			instituicao: &models.InstituicaoEnsino{
				ID:   3,
				Nome: "",
			},
		},
		{
			name: "repo error",
			instituicao: &models.InstituicaoEnsino{
				ID:   1,
				Nome: "UFRJ",
			},
			repoErr: errors.New("update failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInstituicaoRepo{updateErr: tt.repoErr}
			svc := services.NewInstituicaoServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.Update(ctx, tt.instituicao)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInstituicaoService_Delete(t *testing.T) {
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
			repoErr: errors.New("foreign key constraint: cursos reference this institution"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInstituicaoRepo{deleteErr: tt.repoErr}
			svc := services.NewInstituicaoServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.Delete(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInstituicaoService_List(t *testing.T) {
	inst1 := &models.InstituicaoEnsino{
		ID:   1,
		Nome: "UFRJ",
	}
	inst2 := &models.InstituicaoEnsino{
		ID:   2,
		Nome: "UERJ",
	}
	inst3 := &models.InstituicaoEnsino{
		ID:   3,
		Nome: "UFF",
	}
	inst4 := &models.InstituicaoEnsino{
		ID:   4,
		Nome: "PUC-Rio",
	}

	tests := []struct {
		name      string
		filter    map[string]interface{}
		page      int
		pageSize  int
		listItems []*models.InstituicaoEnsino
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
			listItems: []*models.InstituicaoEnsino{},
			listTotal: 0,
			wantLen:   0,
			wantTotal: 0,
		},
		{
			name:      "success - with items",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  10,
			listItems: []*models.InstituicaoEnsino{inst1, inst2, inst3, inst4},
			listTotal: 4,
			wantLen:   4,
			wantTotal: 4,
		},
		{
			name:      "success - filter by name",
			filter:    map[string]interface{}{"nome": "UFRJ"},
			page:      1,
			pageSize:  10,
			listItems: []*models.InstituicaoEnsino{inst1},
			listTotal: 1,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:      "success - pagination page 1",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  2,
			listItems: []*models.InstituicaoEnsino{inst1, inst2},
			listTotal: 4,
			wantLen:   2,
			wantTotal: 4,
		},
		{
			name:      "success - pagination page 2",
			filter:    map[string]interface{}{},
			page:      2,
			pageSize:  2,
			listItems: []*models.InstituicaoEnsino{inst3, inst4},
			listTotal: 4,
			wantLen:   2,
			wantTotal: 4,
		},
		{
			name:      "success - pagination page 3 empty",
			filter:    map[string]interface{}{},
			page:      3,
			pageSize:  2,
			listItems: []*models.InstituicaoEnsino{},
			listTotal: 4,
			wantLen:   0,
			wantTotal: 4,
		},
		{
			name:      "success - large page size",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  100,
			listItems: []*models.InstituicaoEnsino{inst1, inst2, inst3, inst4},
			listTotal: 4,
			wantLen:   4,
			wantTotal: 4,
		},
		{
			name:      "success - small page size",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  1,
			listItems: []*models.InstituicaoEnsino{inst1},
			listTotal: 4,
			wantLen:   1,
			wantTotal: 4,
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
			repo := &mockInstituicaoRepo{
				listItems: tt.listItems,
				listTotal: tt.listTotal,
				listErr:   tt.listErr,
			}
			svc := services.NewInstituicaoServiceWithInterface(repo)
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

func TestInstituicaoService_List_OffsetCalculation(t *testing.T) {
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
		{
			name:       "page 10, size 50",
			page:       10,
			pageSize:   50,
			wantOffset: 450,
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

func TestNewInstituicaoService(t *testing.T) {
	service := services.NewInstituicaoService(nil)

	if service == nil {
		t.Error("NewInstituicaoService() returned nil")
	}
}

func TestNewInstituicaoServiceWithInterface(t *testing.T) {
	repo := &mockInstituicaoRepo{}
	service := services.NewInstituicaoServiceWithInterface(repo)

	if service == nil {
		t.Error("NewInstituicaoServiceWithInterface() returned nil")
	}
}

func TestInstituicaoService_Create_MultipleInstitutions(t *testing.T) {
	// Test creating multiple institutions in sequence
	institutions := []string{
		"UFRJ",
		"UERJ",
		"UFF",
		"PUC-Rio",
		"UNIRIO",
		"CEFET",
		"IME",
		"ITA",
	}

	for i, nome := range institutions {
		t.Run(nome, func(t *testing.T) {
			repo := &mockInstituicaoRepo{createID: i + 1}
			svc := services.NewInstituicaoServiceWithInterface(repo)
			ctx := context.Background()

			instituicao := &models.InstituicaoEnsino{Nome: nome}
			id, err := svc.Create(ctx, instituicao)
			if err != nil {
				t.Errorf("Create() failed for %s: %v", nome, err)
			}
			if id != i+1 {
				t.Errorf("Create() expected id %d, got %d", i+1, id)
			}
		})
	}
}

func TestInstituicaoService_Update_SetID(t *testing.T) {
	// Test that SetID method works correctly
	instituicao := &models.InstituicaoEnsino{Nome: "Test"}
	instituicao.SetID(42)

	if instituicao.ID != 42 {
		t.Errorf("SetID() expected id 42, got %d", instituicao.ID)
	}
}

func TestInstituicaoService_CRUD_Flow(t *testing.T) {
	// Test a complete CRUD flow
	ctx := context.Background()

	// Create
	createRepo := &mockInstituicaoRepo{createID: 1}
	createSvc := services.NewInstituicaoServiceWithInterface(createRepo)
	instituicao := &models.InstituicaoEnsino{Nome: "Test University"}
	id, err := createSvc.Create(ctx, instituicao)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	if id != 1 {
		t.Errorf("Create() expected id 1, got %d", id)
	}

	// Read
	instituicao.ID = id
	getRepo := &mockInstituicaoRepo{entity: instituicao}
	getSvc := services.NewInstituicaoServiceWithInterface(getRepo)
	retrieved, err := getSvc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}
	if retrieved.Nome != "Test University" {
		t.Errorf("GetByID() expected nome 'Test University', got %s", retrieved.Nome)
	}

	// Update
	updateRepo := &mockInstituicaoRepo{}
	updateSvc := services.NewInstituicaoServiceWithInterface(updateRepo)
	instituicao.Nome = "Updated University"
	err = updateSvc.Update(ctx, instituicao)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// Delete
	deleteRepo := &mockInstituicaoRepo{}
	deleteSvc := services.NewInstituicaoServiceWithInterface(deleteRepo)
	err = deleteSvc.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}
}

func TestInstituicaoService_List_WithFilters(t *testing.T) {
	// Test various filter scenarios
	inst1 := &models.InstituicaoEnsino{ID: 1, Nome: "UFRJ"}

	filters := []map[string]interface{}{
		{"nome": "UFRJ"},
		{"id": 1},
		{"nome": "UFRJ", "id": 1},
		{},
	}

	for i, filter := range filters {
		t.Run("filter_"+string(rune(i+'0')), func(t *testing.T) {
			repo := &mockInstituicaoRepo{
				listItems: []*models.InstituicaoEnsino{inst1},
				listTotal: 1,
			}
			svc := services.NewInstituicaoServiceWithInterface(repo)
			ctx := context.Background()

			items, total, err := svc.List(ctx, filter, 1, 10)
			if err != nil {
				t.Errorf("List() failed: %v", err)
			}
			if len(items) != 1 {
				t.Errorf("List() expected 1 item, got %d", len(items))
			}
			if total != 1 {
				t.Errorf("List() expected total 1, got %d", total)
			}
		})
	}
}
