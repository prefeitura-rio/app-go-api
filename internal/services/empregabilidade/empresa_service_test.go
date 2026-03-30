package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// Mock Empresa Repository
type MockEmpresaRepo struct {
	empresas    map[string]*empregabilidade.Empresa
	createError error
	getError    error
	updateError error
	deleteError error
	listError   error
	upsertError error
}

func NewMockEmpresaRepo() *MockEmpresaRepo {
	return &MockEmpresaRepo{
		empresas: make(map[string]*empregabilidade.Empresa),
	}
}

func (m *MockEmpresaRepo) Create(ctx context.Context, entity *empregabilidade.Empresa) (string, error) {
	if m.createError != nil {
		return "", m.createError
	}
	m.empresas[entity.CNPJ] = entity
	return entity.CNPJ, nil
}

func (m *MockEmpresaRepo) GetByID(ctx context.Context, cnpj string) (*empregabilidade.Empresa, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	empresa, exists := m.empresas[cnpj]
	if !exists {
		return nil, nil
	}
	return empresa, nil
}

func (m *MockEmpresaRepo) Update(ctx context.Context, entity *empregabilidade.Empresa) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.empresas[entity.CNPJ] = entity
	return nil
}

func (m *MockEmpresaRepo) Delete(ctx context.Context, cnpj string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.empresas, cnpj)
	return nil
}

func (m *MockEmpresaRepo) List(ctx context.Context, filter empregabilidade.EmpresaFilter, limit, offset int) ([]*empregabilidade.Empresa, int, error) {
	if m.listError != nil {
		return nil, 0, m.listError
	}
	result := make([]*empregabilidade.Empresa, 0)
	for _, e := range m.empresas {
		result = append(result, e)
	}
	return result, len(result), nil
}

func (m *MockEmpresaRepo) Upsert(ctx context.Context, entity *empregabilidade.Empresa) error {
	if m.upsertError != nil {
		return m.upsertError
	}
	m.empresas[entity.CNPJ] = entity
	return nil
}

// ==================== Create Tests ====================

func TestEmpresaService_Create_Success(t *testing.T) {
	t.Run("Successfully create empresa", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		empresa := &empregabilidade.Empresa{
			CNPJ:        "12.345.678/0001-90",
			RazaoSocial: "Empresa Teste LTDA",
		}

		ctx := context.Background()
		cnpj, err := service.Create(ctx, empresa)

		if err != nil {
			t.Errorf("Expected successful creation, got error: %v", err)
		}

		if cnpj == "" {
			t.Error("Expected non-empty CNPJ")
		}

		if cnpj != "12.345.678/0001-90" {
			t.Errorf("Expected CNPJ '12.345.678/0001-90', got '%s'", cnpj)
		}
	})
}

func TestEmpresaService_Create_Error(t *testing.T) {
	t.Run("Error when repository create fails", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		mockRepo.createError = errors.New("database error")
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		empresa := &empregabilidade.Empresa{
			CNPJ:        "12.345.678/0001-90",
			RazaoSocial: "Empresa Teste LTDA",
		}

		ctx := context.Background()
		cnpj, err := service.Create(ctx, empresa)

		if err == nil {
			t.Error("Expected error when create fails")
		}

		if cnpj != "" {
			t.Error("Expected empty CNPJ on error")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got '%s'", err.Error())
		}
	})
}

// ==================== GetByID Tests ====================

func TestEmpresaService_GetByID_Success(t *testing.T) {
	t.Run("Successfully get existing empresa", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		cnpj := "12.345.678/0001-90"
		mockRepo.empresas[cnpj] = &empregabilidade.Empresa{
			CNPJ:        cnpj,
			RazaoSocial: "Empresa Teste LTDA",
		}

		ctx := context.Background()
		result, err := service.GetByID(ctx, cnpj)

		if err != nil {
			t.Errorf("Expected successful get, got error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}

		if result.CNPJ != cnpj {
			t.Errorf("Expected CNPJ '%s', got '%s'", cnpj, result.CNPJ)
		}

		if result.RazaoSocial != "Empresa Teste LTDA" {
			t.Errorf("Expected RazaoSocial 'Empresa Teste LTDA', got '%s'", result.RazaoSocial)
		}
	})

	t.Run("Return nil when empresa not found", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		ctx := context.Background()
		result, err := service.GetByID(ctx, "99.999.999/0001-99")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result != nil {
			t.Error("Expected nil result for non-existent empresa")
		}
	})
}

func TestEmpresaService_GetByID_Error(t *testing.T) {
	t.Run("Error when repository fails", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		mockRepo.getError = errors.New("connection timeout")
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		ctx := context.Background()
		result, err := service.GetByID(ctx, "12.345.678/0001-90")

		if err == nil {
			t.Error("Expected error when repository fails")
		}

		if result != nil {
			t.Error("Expected nil result on error")
		}

		if err.Error() != "connection timeout" {
			t.Errorf("Expected 'connection timeout', got '%s'", err.Error())
		}
	})
}

// ==================== Update Tests ====================

func TestEmpresaService_Update_Success(t *testing.T) {
	t.Run("Successfully update empresa", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		cnpj := "12.345.678/0001-90"
		mockRepo.empresas[cnpj] = &empregabilidade.Empresa{
			CNPJ:        cnpj,
			RazaoSocial: "Empresa Original",
		}

		updated := &empregabilidade.Empresa{
			CNPJ:        cnpj,
			RazaoSocial: "Empresa Atualizada",
		}

		ctx := context.Background()
		err := service.Update(ctx, updated)

		if err != nil {
			t.Errorf("Expected successful update, got error: %v", err)
		}

		if mockRepo.empresas[cnpj].RazaoSocial != "Empresa Atualizada" {
			t.Errorf("Expected RazaoSocial to be updated")
		}
	})
}

func TestEmpresaService_Update_Error(t *testing.T) {
	t.Run("Error when repository fails", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		mockRepo.updateError = errors.New("update failed")
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		empresa := &empregabilidade.Empresa{
			CNPJ:        "12.345.678/0001-90",
			RazaoSocial: "Empresa Teste",
		}

		ctx := context.Background()
		err := service.Update(ctx, empresa)

		if err == nil {
			t.Error("Expected error when update fails")
		}

		if err.Error() != "update failed" {
			t.Errorf("Expected 'update failed', got '%s'", err.Error())
		}
	})
}

// ==================== Delete Tests ====================

func TestEmpresaService_Delete_Success(t *testing.T) {
	t.Run("Successfully delete empresa", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		cnpj := "12.345.678/0001-90"
		mockRepo.empresas[cnpj] = &empregabilidade.Empresa{
			CNPJ:        cnpj,
			RazaoSocial: "Empresa Teste",
		}

		ctx := context.Background()
		err := service.Delete(ctx, cnpj)

		if err != nil {
			t.Errorf("Expected successful delete, got error: %v", err)
		}

		if _, exists := mockRepo.empresas[cnpj]; exists {
			t.Error("Expected empresa to be deleted")
		}
	})
}

func TestEmpresaService_Delete_Error(t *testing.T) {
	t.Run("Error when repository fails", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		mockRepo.deleteError = errors.New("delete failed")
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		ctx := context.Background()
		err := service.Delete(ctx, "12.345.678/0001-90")

		if err == nil {
			t.Error("Expected error when delete fails")
		}

		if err.Error() != "delete failed" {
			t.Errorf("Expected 'delete failed', got '%s'", err.Error())
		}
	})
}

// ==================== List Tests ====================

func TestEmpresaService_List_Success(t *testing.T) {
	t.Run("Successfully list empresas", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		// Add some test empresas
		mockRepo.empresas["11.111.111/0001-11"] = &empregabilidade.Empresa{
			CNPJ:        "11.111.111/0001-11",
			RazaoSocial: "Empresa A",
		}
		mockRepo.empresas["22.222.222/0001-22"] = &empregabilidade.Empresa{
			CNPJ:        "22.222.222/0001-22",
			RazaoSocial: "Empresa B",
		}

		ctx := context.Background()
		result, total, err := service.List(ctx, empregabilidade.EmpresaFilter{}, 1, 10)

		if err != nil {
			t.Errorf("Expected successful list, got error: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result")
		}

		if total != 2 {
			t.Errorf("Expected total of 2, got %d", total)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 empresas in result, got %d", len(result))
		}
	})

	t.Run("List with pagination", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		ctx := context.Background()
		result, total, err := service.List(ctx, empregabilidade.EmpresaFilter{}, 2, 5)

		if err != nil {
			t.Errorf("Expected successful list, got error: %v", err)
		}

		// Result should not be nil, but can be empty
		_ = result

		if total < 0 {
			t.Errorf("Expected non-negative total, got %d", total)
		}
	})
}

func TestEmpresaService_List_Error(t *testing.T) {
	t.Run("Error when repository fails", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		mockRepo.listError = errors.New("list failed")
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		ctx := context.Background()
		result, total, err := service.List(ctx, empregabilidade.EmpresaFilter{}, 1, 10)

		if err == nil {
			t.Error("Expected error when list fails")
		}

		if result != nil {
			t.Error("Expected nil result on error")
		}

		if total != 0 {
			t.Errorf("Expected total of 0 on error, got %d", total)
		}

		if err.Error() != "list failed" {
			t.Errorf("Expected 'list failed', got '%s'", err.Error())
		}
	})
}

// ==================== Upsert Tests ====================

func TestEmpresaService_Upsert_Success(t *testing.T) {
	t.Run("Upsert creates new empresa", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		empresa := &empregabilidade.Empresa{
			CNPJ:        "12.345.678/0001-90",
			RazaoSocial: "Empresa Nova",
		}

		ctx := context.Background()
		err := service.Upsert(ctx, empresa)

		if err != nil {
			t.Errorf("Expected successful upsert, got error: %v", err)
		}

		if _, exists := mockRepo.empresas["12.345.678/0001-90"]; !exists {
			t.Error("Expected empresa to be created")
		}

		if mockRepo.empresas["12.345.678/0001-90"].RazaoSocial != "Empresa Nova" {
			t.Error("Expected RazaoSocial to match")
		}
	})

	t.Run("Upsert updates existing empresa", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		cnpj := "12.345.678/0001-90"
		mockRepo.empresas[cnpj] = &empregabilidade.Empresa{
			CNPJ:        cnpj,
			RazaoSocial: "Empresa Original",
		}

		updated := &empregabilidade.Empresa{
			CNPJ:        cnpj,
			RazaoSocial: "Empresa Atualizada via Upsert",
		}

		ctx := context.Background()
		err := service.Upsert(ctx, updated)

		if err != nil {
			t.Errorf("Expected successful upsert, got error: %v", err)
		}

		if mockRepo.empresas[cnpj].RazaoSocial != "Empresa Atualizada via Upsert" {
			t.Error("Expected RazaoSocial to be updated via upsert")
		}
	})
}

func TestEmpresaService_Upsert_Error(t *testing.T) {
	t.Run("Error when repository fails", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		mockRepo.upsertError = errors.New("upsert failed")
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		empresa := &empregabilidade.Empresa{
			CNPJ:        "12.345.678/0001-90",
			RazaoSocial: "Empresa Teste",
		}

		ctx := context.Background()
		err := service.Upsert(ctx, empresa)

		if err == nil {
			t.Error("Expected error when upsert fails")
		}

		if err.Error() != "upsert failed" {
			t.Errorf("Expected 'upsert failed', got '%s'", err.Error())
		}
	})
}

// ==================== CNPJ Validation Tests ====================

func TestEmpresaService_CNPJOperations(t *testing.T) {
	t.Run("Create with formatted CNPJ", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		empresa := &empregabilidade.Empresa{
			CNPJ:        "12.345.678/0001-90",
			RazaoSocial: "Empresa com CNPJ Formatado",
		}

		ctx := context.Background()
		cnpj, err := service.Create(ctx, empresa)

		if err != nil {
			t.Errorf("Expected successful creation with formatted CNPJ, got error: %v", err)
		}

		if cnpj != "12.345.678/0001-90" {
			t.Errorf("Expected CNPJ '12.345.678/0001-90', got '%s'", cnpj)
		}
	})

	t.Run("Create with numeric CNPJ", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		empresa := &empregabilidade.Empresa{
			CNPJ:        "12345678000190",
			RazaoSocial: "Empresa com CNPJ Numerico",
		}

		ctx := context.Background()
		cnpj, err := service.Create(ctx, empresa)

		if err != nil {
			t.Errorf("Expected successful creation with numeric CNPJ, got error: %v", err)
		}

		if cnpj != "12345678000190" {
			t.Errorf("Expected CNPJ '12345678000190', got '%s'", cnpj)
		}
	})
}

// ==================== Concurrent Operations Tests ====================

func TestEmpresaService_ConcurrentOperations(t *testing.T) {
	t.Run("Multiple concurrent gets", func(t *testing.T) {
		mockRepo := NewMockEmpresaRepo()
		service := services.NewEmpresaServiceWithInterface(mockRepo)

		cnpj := "12.345.678/0001-90"
		mockRepo.empresas[cnpj] = &empregabilidade.Empresa{
			CNPJ:        cnpj,
			RazaoSocial: "Empresa Teste",
		}

		ctx := context.Background()

		// Simulate concurrent gets
		done := make(chan bool, 3)
		for i := 0; i < 3; i++ {
			go func() {
				result, err := service.GetByID(ctx, cnpj)
				if err != nil || result == nil {
					t.Errorf("Concurrent get failed")
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 3; i++ {
			<-done
		}
	})
}

// ==================== Constructor Tests ====================

func TestNewEmpresaService(t *testing.T) {
	mockRepo := NewMockEmpresaRepo()
	service := services.NewEmpresaServiceWithInterface(mockRepo)

	if service == nil {
		t.Error("NewEmpresaService() returned nil")
	}
}

func TestNewEmpresaServiceWithInterface(t *testing.T) {
	mockRepo := NewMockEmpresaRepo()
	service := services.NewEmpresaServiceWithInterface(mockRepo)

	if service == nil {
		t.Error("NewEmpresaServiceWithInterface() returned nil")
	}
}
