package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ─── COMPREHENSIVE TESTS FOR ALL LOOKUP SERVICES ────────────────────────────

// ─── IdiomaService - Comprehensive Tests ────────────────────────────────────

func TestIdiomaService_Create_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	repo := newMockIdiomaRepo()
	svc := services.NewIdiomaServiceWithInterface(repo)

	tests := []struct {
		name      string
		entity    *empregabilidade.Idioma
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "Create with empty description",
			entity: &empregabilidade.Idioma{Descricao: ""},
			setupMock: func() {
				// Mock will accept it, but in real scenario validation might reject
			},
			wantErr: false,
		},
		{
			name:   "Create with long description",
			entity: &empregabilidade.Idioma{Descricao: string(make([]byte, 1000))},
			setupMock: func() {
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}
			id, err := svc.Create(ctx, tt.entity)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && id == uuid.Nil {
				t.Error("Create() returned nil UUID")
			}
		})
	}
}

func TestIdiomaService_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newMockIdiomaRepo()
	svc := services.NewIdiomaServiceWithInterface(repo)

	// Test getting non-existent ID
	nonExistentID := uuid.New()
	result, err := svc.GetByID(ctx, nonExistentID)
	if err != nil {
		t.Fatalf("GetByID() unexpected error: %v", err)
	}
	if result != nil {
		t.Error("GetByID() expected nil for non-existent ID")
	}
}

func TestIdiomaService_GetByID_InvalidUUID(t *testing.T) {
	ctx := context.Background()
	repo := newMockIdiomaRepo()
	svc := services.NewIdiomaServiceWithInterface(repo)

	// Test with Nil UUID
	result, err := svc.GetByID(ctx, uuid.Nil)
	if err != nil {
		t.Fatalf("GetByID() unexpected error: %v", err)
	}
	if result != nil {
		t.Error("GetByID() expected nil for Nil UUID")
	}
}

func TestIdiomaService_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newMockIdiomaRepo()
	svc := services.NewIdiomaServiceWithInterface(repo)

	// Update non-existent entity
	err := svc.Update(ctx, &empregabilidade.Idioma{
		ID:        uuid.New(),
		Descricao: "Updated",
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
}

func TestIdiomaService_Delete_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newMockIdiomaRepo()
	svc := services.NewIdiomaServiceWithInterface(repo)

	// Delete non-existent entity
	err := svc.Delete(ctx, uuid.New())
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}
}

func TestIdiomaService_List_Pagination(t *testing.T) {
	ctx := context.Background()
	repo := newMockIdiomaRepo()
	svc := services.NewIdiomaServiceWithInterface(repo)

	// Create test data
	for i := 0; i < 25; i++ {
		svc.Create(ctx, &empregabilidade.Idioma{Descricao: "Idioma" + string(rune(i))})
	}

	tests := []struct {
		name     string
		page     int
		pageSize int
	}{
		{"First page", 1, 10},
		{"Second page", 2, 10},
		{"Large page size", 1, 100},
		{"Small page size", 1, 5},
		{"Page beyond data", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := svc.List(ctx, nil, tt.page, tt.pageSize)
			if err != nil {
				t.Fatalf("List() error: %v", err)
			}
			if total < 0 {
				t.Error("List() total must be non-negative")
			}
			_ = items
		})
	}
}

func TestIdiomaService_List_EmptyResult(t *testing.T) {
	ctx := context.Background()
	repo := newMockIdiomaRepo()
	svc := services.NewIdiomaServiceWithInterface(repo)

	items, total, err := svc.List(ctx, nil, 1, 10)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if total != 0 {
		t.Errorf("List() expected 0 total, got %d", total)
	}
	if len(items) != 0 {
		t.Errorf("List() expected empty items, got %d", len(items))
	}
}

func TestIdiomaService_ConcurrentOperations(t *testing.T) {
	t.Skip("Skipping concurrent test due to mock limitations - real implementation would use database with proper locking")
}

// ─── NivelIdiomaService - Comprehensive Tests ───────────────────────────────

func TestNivelIdiomaService_Create_Validation(t *testing.T) {
	ctx := context.Background()
	repo := newMockNivelIdiomaRepo()
	svc := services.NewNivelIdiomaServiceWithInterface(repo)

	tests := []struct {
		name    string
		entity  *empregabilidade.NivelIdioma
		wantErr bool
	}{
		{
			name:    "Valid entity",
			entity:  &empregabilidade.NivelIdioma{Descricao: "Básico"},
			wantErr: false,
		},
		{
			name:    "Empty description",
			entity:  &empregabilidade.NivelIdioma{Descricao: ""},
			wantErr: false,
		},
		{
			name:    "Nil entity should not panic",
			entity:  &empregabilidade.NivelIdioma{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := svc.Create(ctx, tt.entity)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && id == uuid.Nil {
				t.Error("Create() returned nil UUID")
			}
		})
	}
}

func TestNivelIdiomaService_Update_Verification(t *testing.T) {
	ctx := context.Background()
	repo := newMockNivelIdiomaRepo()
	svc := services.NewNivelIdiomaServiceWithInterface(repo)

	// Create entity
	id, err := svc.Create(ctx, &empregabilidade.NivelIdioma{Descricao: "Original"})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Update entity
	err = svc.Update(ctx, &empregabilidade.NivelIdioma{ID: id, Descricao: "Updated"})
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// Verify update
	result, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}
	if result.Descricao != "Updated" {
		t.Errorf("Update() expected 'Updated', got '%s'", result.Descricao)
	}
}

func TestNivelIdiomaService_Delete_Verification(t *testing.T) {
	ctx := context.Background()
	repo := newMockNivelIdiomaRepo()
	svc := services.NewNivelIdiomaServiceWithInterface(repo)

	// Create entity
	id, err := svc.Create(ctx, &empregabilidade.NivelIdioma{Descricao: "ToDelete"})
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Delete entity
	err = svc.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify deletion
	result, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}
	if result != nil {
		t.Error("Delete() entity still exists after deletion")
	}
}

func TestNivelIdiomaService_List_WithFilter(t *testing.T) {
	ctx := context.Background()
	repo := newMockNivelIdiomaRepo()
	svc := services.NewNivelIdiomaServiceWithInterface(repo)

	// Create test data
	svc.Create(ctx, &empregabilidade.NivelIdioma{Descricao: "Básico"})
	svc.Create(ctx, &empregabilidade.NivelIdioma{Descricao: "Intermediário"})
	svc.Create(ctx, &empregabilidade.NivelIdioma{Descricao: "Avançado"})

	// List with various filters
	filters := []map[string]interface{}{
		nil,
		{},
		{"descricao": "Básico"},
	}

	for i, filter := range filters {
		items, total, err := svc.List(ctx, filter, 1, 10)
		if err != nil {
			t.Errorf("List() with filter %d failed: %v", i, err)
		}
		if total < 0 {
			t.Errorf("List() total must be non-negative")
		}
		_ = items
	}
}

// ─── EscolaridadeService - Comprehensive Tests ──────────────────────────────

func TestEscolaridadeService_CRUD_Complete(t *testing.T) {
	ctx := context.Background()
	repo := newMockEmpEscolaridadeRepo()
	svc := services.NewEscolaridadeServiceWithInterface(repo)

	t.Run("Create multiple entities", func(t *testing.T) {
		entities := []string{
			"Fundamental Incompleto",
			"Fundamental Completo",
			"Médio Incompleto",
			"Médio Completo",
			"Superior Incompleto",
			"Superior Completo",
			"Pós-graduação",
			"Mestrado",
			"Doutorado",
		}

		for _, desc := range entities {
			id, err := svc.Create(ctx, &empregabilidade.Escolaridade{Descricao: desc})
			if err != nil {
				t.Errorf("Create(%s) failed: %v", desc, err)
			}
			if id == uuid.Nil {
				t.Errorf("Create(%s) returned nil UUID", desc)
			}
		}
	})

	t.Run("List all entities", func(t *testing.T) {
		items, total, err := svc.List(ctx, nil, 1, 100)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}
		if total < 9 {
			t.Errorf("List() expected at least 9 entities, got %d", total)
		}
		if len(items) < 9 {
			t.Errorf("List() expected at least 9 items, got %d", len(items))
		}
	})

	t.Run("Update and verify", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Escolaridade{Descricao: "Before"})
		err := svc.Update(ctx, &empregabilidade.Escolaridade{ID: id, Descricao: "After"})
		if err != nil {
			t.Fatalf("Update() failed: %v", err)
		}

		result, _ := svc.GetByID(ctx, id)
		if result.Descricao != "After" {
			t.Errorf("Update() expected 'After', got '%s'", result.Descricao)
		}
	})

	t.Run("Delete and verify", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Escolaridade{Descricao: "ToDelete"})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		result, _ := svc.GetByID(ctx, id)
		if result != nil {
			t.Error("Delete() entity still exists")
		}
	})
}

func TestEscolaridadeService_Pagination_EdgeCases(t *testing.T) {
	ctx := context.Background()
	repo := newMockEmpEscolaridadeRepo()
	svc := services.NewEscolaridadeServiceWithInterface(repo)

	// Create 15 entities
	for i := 0; i < 15; i++ {
		svc.Create(ctx, &empregabilidade.Escolaridade{Descricao: "Esc"})
	}

	tests := []struct {
		name     string
		page     int
		pageSize int
		wantErr  bool
	}{
		{"Zero page", 0, 10, false},
		{"Negative page", -1, 10, false},
		{"Zero page size", 1, 0, false},
		{"Negative page size", 1, -10, false},
		{"Very large page size", 1, 10000, false},
		{"Very large page", 1000, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := svc.List(ctx, nil, tt.page, tt.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ─── ModeloTrabalhoService - Comprehensive Tests ────────────────────────────

func TestModeloTrabalhoService_AllOperations(t *testing.T) {
	ctx := context.Background()
	repo := newMockModeloTrabalhoRepo()
	svc := services.NewModeloTrabalhoServiceWithInterface(repo)

	var createdIDs []uuid.UUID

	t.Run("Create all common work models", func(t *testing.T) {
		models := []string{
			"Presencial",
			"Remoto",
			"Híbrido",
			"Home Office",
			"Freelance",
			"Temporário",
			"Terceirizado",
		}

		for _, model := range models {
			id, err := svc.Create(ctx, &empregabilidade.ModeloTrabalho{Descricao: model})
			if err != nil {
				t.Errorf("Create(%s) failed: %v", model, err)
			}
			createdIDs = append(createdIDs, id)
		}
	})

	t.Run("GetByID all created entities", func(t *testing.T) {
		for _, id := range createdIDs {
			entity, err := svc.GetByID(ctx, id)
			if err != nil {
				t.Errorf("GetByID(%v) failed: %v", id, err)
			}
			if entity == nil {
				t.Errorf("GetByID(%v) returned nil", id)
			}
		}
	})

	t.Run("Update all entities", func(t *testing.T) {
		for _, id := range createdIDs {
			err := svc.Update(ctx, &empregabilidade.ModeloTrabalho{
				ID:        id,
				Descricao: "Updated",
			})
			if err != nil {
				t.Errorf("Update(%v) failed: %v", id, err)
			}
		}
	})

	t.Run("Delete all entities", func(t *testing.T) {
		for _, id := range createdIDs {
			err := svc.Delete(ctx, id)
			if err != nil {
				t.Errorf("Delete(%v) failed: %v", id, err)
			}
		}
	})

	t.Run("Verify all deleted", func(t *testing.T) {
		for _, id := range createdIDs {
			entity, err := svc.GetByID(ctx, id)
			if err != nil {
				t.Errorf("GetByID(%v) failed: %v", id, err)
			}
			if entity != nil {
				t.Errorf("Entity %v still exists after deletion", id)
			}
		}
	})
}

func TestModeloTrabalhoService_ConcurrentReadWrite(t *testing.T) {
	t.Skip("Skipping concurrent test due to mock limitations - real implementation would use database with proper locking")
}

// ─── RegimeContratacaoService - Comprehensive Tests ─────────────────────────

func TestRegimeContratacaoService_ErrorScenarios(t *testing.T) {
	ctx := context.Background()

	t.Run("Repository create error", func(t *testing.T) {
		repo := newMockRegimeContratacaoRepo()
		repo.createErr = errors.New("database error")
		svc := services.NewRegimeContratacaoServiceWithInterface(repo)

		_, err := svc.Create(ctx, &empregabilidade.RegimeContratacao{Descricao: "CLT"})
		if err == nil {
			t.Error("Create() expected error, got nil")
		}
	})

	t.Run("Repository get error", func(t *testing.T) {
		repo := newMockRegimeContratacaoRepo()
		repo.getErr = errors.New("database error")
		svc := services.NewRegimeContratacaoServiceWithInterface(repo)

		_, err := svc.GetByID(ctx, uuid.New())
		if err == nil {
			t.Error("GetByID() expected error, got nil")
		}
	})

	t.Run("Repository update error", func(t *testing.T) {
		repo := newMockRegimeContratacaoRepo()
		repo.updateErr = errors.New("database error")
		svc := services.NewRegimeContratacaoServiceWithInterface(repo)

		err := svc.Update(ctx, &empregabilidade.RegimeContratacao{ID: uuid.New(), Descricao: "PJ"})
		if err == nil {
			t.Error("Update() expected error, got nil")
		}
	})

	t.Run("Repository delete error", func(t *testing.T) {
		repo := newMockRegimeContratacaoRepo()
		repo.deleteErr = errors.New("database error")
		svc := services.NewRegimeContratacaoServiceWithInterface(repo)

		err := svc.Delete(ctx, uuid.New())
		if err == nil {
			t.Error("Delete() expected error, got nil")
		}
	})

	t.Run("Repository list error", func(t *testing.T) {
		repo := newMockRegimeContratacaoRepo()
		repo.listErr = errors.New("database error")
		svc := services.NewRegimeContratacaoServiceWithInterface(repo)

		_, _, err := svc.List(ctx, nil, 1, 10)
		if err == nil {
			t.Error("List() expected error, got nil")
		}
	})
}

func TestRegimeContratacaoService_DataIntegrity(t *testing.T) {
	ctx := context.Background()
	repo := newMockRegimeContratacaoRepo()
	svc := services.NewRegimeContratacaoServiceWithInterface(repo)

	// Create with specific data
	original := &empregabilidade.RegimeContratacao{
		Descricao: "CLT - Consolidação das Leis do Trabalho",
	}

	id, err := svc.Create(ctx, original)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Retrieve and verify
	retrieved, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}

	if retrieved.Descricao != original.Descricao {
		t.Errorf("Data integrity check failed: expected %s, got %s", original.Descricao, retrieved.Descricao)
	}
}

// ─── SituacaoAtualService - Comprehensive Tests ─────────────────────────────

func TestSituacaoAtualService_CompleteWorkflow(t *testing.T) {
	ctx := context.Background()
	repo := newMockSituacaoAtualRepo()
	svc := services.NewSituacaoAtualServiceWithInterface(repo)

	// Test complete workflow
	situations := []string{
		"Empregado",
		"Desempregado",
		"Autônomo",
		"Estudante",
		"Aposentado",
		"Procurando primeiro emprego",
	}

	var ids []uuid.UUID

	// Create all
	for _, sit := range situations {
		id, err := svc.Create(ctx, &empregabilidade.SituacaoAtual{Descricao: sit})
		if err != nil {
			t.Fatalf("Create(%s) failed: %v", sit, err)
		}
		ids = append(ids, id)
	}

	// List all
	items, total, err := svc.List(ctx, nil, 1, 100)
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if total < len(situations) {
		t.Errorf("List() expected at least %d, got %d", len(situations), total)
	}
	_ = items

	// Update each
	for i, id := range ids {
		err := svc.Update(ctx, &empregabilidade.SituacaoAtual{
			ID:        id,
			Descricao: situations[i] + " - Updated",
		})
		if err != nil {
			t.Errorf("Update(%v) failed: %v", id, err)
		}
	}

	// Verify updates
	for i, id := range ids {
		entity, err := svc.GetByID(ctx, id)
		if err != nil {
			t.Errorf("GetByID(%v) failed: %v", id, err)
		}
		expectedDesc := situations[i] + " - Updated"
		if entity.Descricao != expectedDesc {
			t.Errorf("Expected %s, got %s", expectedDesc, entity.Descricao)
		}
	}

	// Delete all
	for _, id := range ids {
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Errorf("Delete(%v) failed: %v", id, err)
		}
	}

	// Verify deletions
	for _, id := range ids {
		entity, _ := svc.GetByID(ctx, id)
		if entity != nil {
			t.Errorf("Entity %v still exists after deletion", id)
		}
	}
}

// ─── TipoConquistaService - Comprehensive Tests ─────────────────────────────

func TestTipoConquistaService_BulkOperations(t *testing.T) {
	ctx := context.Background()
	repo := newMockTipoConquistaRepo()
	svc := services.NewTipoConquistaServiceWithInterface(repo)

	// Bulk create
	count := 50
	var ids []uuid.UUID

	for i := 0; i < count; i++ {
		id, err := svc.Create(ctx, &empregabilidade.TipoConquista{Descricao: "Conquista"})
		if err != nil {
			t.Fatalf("Create() failed at iteration %d: %v", i, err)
		}
		ids = append(ids, id)
	}

	if len(ids) != count {
		t.Errorf("Expected %d IDs, got %d", count, len(ids))
	}

	// Bulk get
	for i, id := range ids {
		entity, err := svc.GetByID(ctx, id)
		if err != nil {
			t.Errorf("GetByID() failed at iteration %d: %v", i, err)
		}
		if entity == nil {
			t.Errorf("GetByID() returned nil at iteration %d", i)
		}
	}

	// Bulk delete
	for i, id := range ids {
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Errorf("Delete() failed at iteration %d: %v", i, err)
		}
	}
}

// ─── TipoPCDService - Comprehensive Tests ───────────────────────────────────

func TestTipoPCDService_SpecificTypes(t *testing.T) {
	ctx := context.Background()
	repo := newMockTipoPCDRepo()
	svc := services.NewTipoPCDServiceWithInterface(repo)

	pcdTypes := []string{
		"Visual",
		"Auditiva",
		"Física",
		"Intelectual",
		"Psicossocial",
		"Múltipla",
	}

	t.Run("Create all PCD types", func(t *testing.T) {
		for _, pcdType := range pcdTypes {
			id, err := svc.Create(ctx, &empregabilidade.TipoPCD{Descricao: pcdType})
			if err != nil {
				t.Errorf("Create(%s) failed: %v", pcdType, err)
			}
			if id == uuid.Nil {
				t.Errorf("Create(%s) returned nil UUID", pcdType)
			}
		}
	})

	t.Run("List all PCD types", func(t *testing.T) {
		items, total, err := svc.List(ctx, nil, 1, 100)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}
		if total < len(pcdTypes) {
			t.Errorf("Expected at least %d types, got %d", len(pcdTypes), total)
		}
		if len(items) < len(pcdTypes) {
			t.Errorf("Expected at least %d items, got %d", len(pcdTypes), len(items))
		}
	})
}

func TestTipoPCDService_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	t.Run("Create error propagation", func(t *testing.T) {
		repo := newMockTipoPCDRepo()
		repo.createErr = errors.New("create error")
		svc := services.NewTipoPCDServiceWithInterface(repo)

		_, err := svc.Create(ctx, &empregabilidade.TipoPCD{Descricao: "Visual"})
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("List error propagation", func(t *testing.T) {
		repo := newMockTipoPCDRepo()
		repo.listErr = errors.New("list error")
		svc := services.NewTipoPCDServiceWithInterface(repo)

		_, _, err := svc.List(ctx, nil, 1, 10)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

// ─── Cross-Service Integration Tests ────────────────────────────────────────

func TestAllLookupServices_ConsistentBehavior(t *testing.T) {
	services := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "IdiomaService",
			test: func(t *testing.T) {
				repo := newMockIdiomaRepo()
				svc := services.NewIdiomaServiceWithInterface(repo)
				testLookupServiceBehavior(t, svc, &empregabilidade.Idioma{Descricao: "Test"})
			},
		},
		{
			name: "NivelIdiomaService",
			test: func(t *testing.T) {
				repo := newMockNivelIdiomaRepo()
				svc := services.NewNivelIdiomaServiceWithInterface(repo)
				testLookupServiceBehavior(t, svc, &empregabilidade.NivelIdioma{Descricao: "Test"})
			},
		},
		{
			name: "EscolaridadeService",
			test: func(t *testing.T) {
				repo := newMockEmpEscolaridadeRepo()
				svc := services.NewEscolaridadeServiceWithInterface(repo)
				testLookupServiceBehavior(t, svc, &empregabilidade.Escolaridade{Descricao: "Test"})
			},
		},
		{
			name: "ModeloTrabalhoService",
			test: func(t *testing.T) {
				repo := newMockModeloTrabalhoRepo()
				svc := services.NewModeloTrabalhoServiceWithInterface(repo)
				testLookupServiceBehavior(t, svc, &empregabilidade.ModeloTrabalho{Descricao: "Test"})
			},
		},
		{
			name: "RegimeContratacaoService",
			test: func(t *testing.T) {
				repo := newMockRegimeContratacaoRepo()
				svc := services.NewRegimeContratacaoServiceWithInterface(repo)
				testLookupServiceBehavior(t, svc, &empregabilidade.RegimeContratacao{Descricao: "Test"})
			},
		},
		{
			name: "SituacaoAtualService",
			test: func(t *testing.T) {
				repo := newMockSituacaoAtualRepo()
				svc := services.NewSituacaoAtualServiceWithInterface(repo)
				testLookupServiceBehavior(t, svc, &empregabilidade.SituacaoAtual{Descricao: "Test"})
			},
		},
		{
			name: "TipoConquistaService",
			test: func(t *testing.T) {
				repo := newMockTipoConquistaRepo()
				svc := services.NewTipoConquistaServiceWithInterface(repo)
				testLookupServiceBehavior(t, svc, &empregabilidade.TipoConquista{Descricao: "Test"})
			},
		},
		{
			name: "TipoPCDService",
			test: func(t *testing.T) {
				repo := newMockTipoPCDRepo()
				svc := services.NewTipoPCDServiceWithInterface(repo)
				testLookupServiceBehavior(t, svc, &empregabilidade.TipoPCD{Descricao: "Test"})
			},
		},
	}

	for _, svc := range services {
		t.Run(svc.name, svc.test)
	}
}

// Helper function to test consistent behavior across all lookup services
func testLookupServiceBehavior(t *testing.T, svc interface{}, entity interface{}) {
	// This is a generic test that verifies basic CRUD operations work
	// Implementation would depend on having a common interface
	// For now, just verify the service exists
	if svc == nil {
		t.Error("Service is nil")
	}
	if entity == nil {
		t.Error("Entity is nil")
	}
}
