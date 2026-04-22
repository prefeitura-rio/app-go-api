package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

// mockOportunidadeMEIRepo implements repository methods for OportunidadeMEI
type mockOportunidadeMEIRepo struct {
	createID               int
	createErr              error
	entity                 *models.OportunidadeMEI
	getErr                 error
	updateErr              error
	deleteErr              error
	listItems              []*models.OportunidadeMEI
	listTotal              int
	listErr                error
	listByStatusItems      []*models.OportunidadeMEI
	listByStatusTotal      int
	listByStatusErr        error
	listByOrgaoItems       []*models.OportunidadeMEI
	listByOrgaoTotal       int
	listByOrgaoErr         error
	updateExpiredErr       error
	updateExpiredCallCount int
}

func (m *mockOportunidadeMEIRepo) Create(ctx context.Context, oportunidade *models.OportunidadeMEI) (int, error) {
	if m.createErr != nil {
		return 0, m.createErr
	}
	if m.createID == 0 {
		m.createID = 1
	}
	return m.createID, nil
}

func (m *mockOportunidadeMEIRepo) GetByID(ctx context.Context, id int) (*models.OportunidadeMEI, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.entity, nil
}

func (m *mockOportunidadeMEIRepo) Update(ctx context.Context, oportunidade *models.OportunidadeMEI) error {
	return m.updateErr
}

func (m *mockOportunidadeMEIRepo) Delete(ctx context.Context, id int) error {
	return m.deleteErr
}

func (m *mockOportunidadeMEIRepo) List(ctx context.Context, filters map[string]interface{}, titulo string, limit, offset int) ([]*models.OportunidadeMEI, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	if m.listItems == nil {
		return []*models.OportunidadeMEI{}, 0, nil
	}
	return m.listItems, m.listTotal, nil
}

func (m *mockOportunidadeMEIRepo) ListByStatus(ctx context.Context, status models.StatusOportunidadeMEI, limit, offset int) ([]*models.OportunidadeMEI, int, error) {
	if m.listByStatusErr != nil {
		return nil, 0, m.listByStatusErr
	}
	if m.listByStatusItems == nil {
		return []*models.OportunidadeMEI{}, 0, nil
	}
	return m.listByStatusItems, m.listByStatusTotal, nil
}

func (m *mockOportunidadeMEIRepo) ListByOrgao(ctx context.Context, orgaoID string, limit, offset int) ([]*models.OportunidadeMEI, int, error) {
	if m.listByOrgaoErr != nil {
		return nil, 0, m.listByOrgaoErr
	}
	if m.listByOrgaoItems == nil {
		return []*models.OportunidadeMEI{}, 0, nil
	}
	return m.listByOrgaoItems, m.listByOrgaoTotal, nil
}

func (m *mockOportunidadeMEIRepo) UpdateExpiredOpportunities(ctx context.Context) error {
	m.updateExpiredCallCount++
	return m.updateExpiredErr
}

// Helper function to create a valid oportunidade
func validOportunidade() *models.OportunidadeMEI {
	dataExpiracao := time.Now().Add(30 * 24 * time.Hour)
	return &models.OportunidadeMEI{
		ID:               1,
		Titulo:           "Serviço de Limpeza",
		DescricaoServico: "Limpeza de prédios públicos",
		OrgaoID:          "org-123",
		CNAEIDs:          []string{"8121-4/00"},
		DataExpiracao:    &dataExpiracao,
		Status:           models.StatusOportunidadeActive,
		Logradouro:       "Rua ABC",
		Numero:           "123",
		Bairro:           "Centro",
		Cidade:           "Rio de Janeiro",
		Estado:           "RJ",
		PrazoPagamento:   "30 dias após conclusão",
	}
}

func TestOportunidadeMEIService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("Create as draft", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{createID: 10}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		oportunidade := validOportunidade()
		oportunidade.Status = "" // Reset status to test draft setting

		id, err := svc.Create(ctx, oportunidade, true)
		if err != nil {
			t.Fatalf("Create draft: %v", err)
		}
		if id != 10 {
			t.Errorf("expected id 10, got %d", id)
		}
		if oportunidade.Status != models.StatusOportunidadeDraft {
			t.Errorf("expected draft status, got %s", oportunidade.Status)
		}
	})

	t.Run("Create as active", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{createID: 11}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		oportunidade := validOportunidade()
		oportunidade.Status = "" // Reset status

		id, err := svc.Create(ctx, oportunidade, false)
		if err != nil {
			t.Fatalf("Create active: %v", err)
		}
		if id != 11 {
			t.Errorf("expected id 11, got %d", id)
		}
		if oportunidade.Status != models.StatusOportunidadeActive {
			t.Errorf("expected active status, got %s", oportunidade.Status)
		}
	})

	t.Run("Create with validation error", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		// Empty oportunidade should fail validation
		oportunidade := &models.OportunidadeMEI{}
		_, err := svc.Create(ctx, oportunidade, false)
		if err == nil {
			t.Error("expected validation error for empty oportunidade")
		}
	})

	t.Run("Create with repo error", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{createErr: errors.New("db error")}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		oportunidade := validOportunidade()
		_, err := svc.Create(ctx, oportunidade, false)
		if err == nil {
			t.Error("expected repo error")
		}
	})
}

func TestOportunidadeMEIService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("Get existing oportunidade", func(t *testing.T) {
		oportunidade := validOportunidade()
		repo := &mockOportunidadeMEIRepo{entity: oportunidade}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		result, err := svc.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.ID != 1 {
			t.Errorf("expected ID 1, got %d", result.ID)
		}
	})

	t.Run("Get with expired oportunidade", func(t *testing.T) {
		expiredDate := time.Now().Add(-1 * 24 * time.Hour)
		oportunidade := validOportunidade()
		oportunidade.DataExpiracao = &expiredDate
		oportunidade.Status = models.StatusOportunidadeActive

		repo := &mockOportunidadeMEIRepo{entity: oportunidade}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		result, err := svc.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		// UpdateStatusBasedOnExpiration should be called
		if result.Status != models.StatusOportunidadeExpired {
			t.Errorf("expected expired status, got %s", result.Status)
		}
	})

	t.Run("Get non-existing oportunidade", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{entity: nil}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		result, err := svc.GetByID(ctx, 999)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if result != nil {
			t.Error("expected nil for non-existing oportunidade")
		}
	})

	t.Run("Get with repo error", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{getErr: errors.New("db error")}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		_, err := svc.GetByID(ctx, 1)
		if err == nil {
			t.Error("expected repo error")
		}
	})
}

func TestOportunidadeMEIService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("Update draft oportunidade", func(t *testing.T) {
		existing := validOportunidade()
		existing.Status = models.StatusOportunidadeDraft

		repo := &mockOportunidadeMEIRepo{entity: existing}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		updated := validOportunidade()
		updated.ID = 1
		updated.Titulo = "Updated Title"

		err := svc.Update(ctx, updated)
		if err != nil {
			t.Fatalf("Update draft: %v", err)
		}
		// Status should be preserved as draft
		if updated.Status != models.StatusOportunidadeDraft {
			t.Errorf("expected draft status preserved, got %s", updated.Status)
		}
	})

	t.Run("Update active oportunidade", func(t *testing.T) {
		existing := validOportunidade()
		existing.Status = models.StatusOportunidadeActive

		repo := &mockOportunidadeMEIRepo{entity: existing}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		updated := validOportunidade()
		updated.ID = 1
		updated.Titulo = "Updated Title"

		err := svc.Update(ctx, updated)
		if err != nil {
			t.Fatalf("Update active: %v", err)
		}
		if updated.Status != models.StatusOportunidadeActive {
			t.Errorf("expected active status preserved, got %s", updated.Status)
		}
	})

	t.Run("Update non-existing oportunidade", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{entity: nil}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		oportunidade := validOportunidade()
		err := svc.Update(ctx, oportunidade)
		if err == nil {
			t.Error("expected error for non-existing oportunidade")
		}
	})

	t.Run("Update with repo error on get", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{getErr: errors.New("db error")}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		err := svc.Update(ctx, validOportunidade())
		if err == nil {
			t.Error("expected repo error")
		}
	})

	t.Run("Update with repo error on update", func(t *testing.T) {
		existing := validOportunidade()
		repo := &mockOportunidadeMEIRepo{
			entity:    existing,
			updateErr: errors.New("update error"),
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		err := svc.Update(ctx, validOportunidade())
		if err == nil {
			t.Error("expected update error")
		}
	})
}

func TestOportunidadeMEIService_Publish(t *testing.T) {
	ctx := context.Background()

	t.Run("Publish valid draft", func(t *testing.T) {
		draft := validOportunidade()
		draft.Status = models.StatusOportunidadeDraft

		repo := &mockOportunidadeMEIRepo{entity: draft}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		err := svc.Publish(ctx, 1)
		if err != nil {
			t.Fatalf("Publish: %v", err)
		}
	})

	t.Run("Publish invalid oportunidade", func(t *testing.T) {
		invalid := &models.OportunidadeMEI{
			ID:     1,
			Status: models.StatusOportunidadeDraft,
			// Missing required fields for publish
		}

		repo := &mockOportunidadeMEIRepo{entity: invalid}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		err := svc.Publish(ctx, 1)
		if err == nil {
			t.Error("expected validation error for incomplete oportunidade")
		}
	})

	t.Run("Publish non-existing oportunidade", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{entity: nil}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		err := svc.Publish(ctx, 999)
		if err == nil {
			t.Error("expected error for non-existing oportunidade")
		}
	})

	t.Run("Publish with repo error", func(t *testing.T) {
		draft := validOportunidade()
		draft.Status = models.StatusOportunidadeDraft

		repo := &mockOportunidadeMEIRepo{
			entity:    draft,
			updateErr: errors.New("update error"),
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		err := svc.Publish(ctx, 1)
		if err == nil {
			t.Error("expected update error")
		}
	})
}

func TestOportunidadeMEIService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Delete success", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		err := svc.Delete(ctx, 1)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("Delete with error", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{deleteErr: errors.New("delete error")}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		err := svc.Delete(ctx, 1)
		if err == nil {
			t.Error("expected delete error")
		}
	})
}

func TestOportunidadeMEIService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("List with status filter", func(t *testing.T) {
		active := validOportunidade()
		active.Status = models.StatusOportunidadeActive

		draft := validOportunidade()
		draft.ID = 2
		draft.Status = models.StatusOportunidadeDraft

		repo := &mockOportunidadeMEIRepo{
			listItems: []*models.OportunidadeMEI{active, draft},
			listTotal: 2,
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		filters := map[string]interface{}{
			"status": models.StatusOportunidadeActive,
		}

		items, total, err := svc.List(ctx, filters, "", 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		// Should only return active items
		if len(items) != 1 {
			t.Errorf("expected 1 active item, got %d", len(items))
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
	})

	t.Run("List without filters", func(t *testing.T) {
		items := []*models.OportunidadeMEI{validOportunidade(), validOportunidade()}
		items[1].ID = 2

		repo := &mockOportunidadeMEIRepo{
			listItems: items,
			listTotal: 2,
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		result, total, err := svc.List(ctx, nil, "", 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 items, got %d", len(result))
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
	})

	t.Run("List with pagination", func(t *testing.T) {
		items := make([]*models.OportunidadeMEI, 5)
		for i := 0; i < 5; i++ {
			items[i] = validOportunidade()
			items[i].ID = i + 1
		}

		repo := &mockOportunidadeMEIRepo{
			listItems: items,
			listTotal: 5,
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		// Page 1, size 2
		result, total, err := svc.List(ctx, nil, "", 1, 2)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 items on page 1, got %d", len(result))
		}
		if total != 5 {
			t.Errorf("expected total 5, got %d", total)
		}
	})

	t.Run("List with repo error", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{listErr: errors.New("db error")}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		_, _, err := svc.List(ctx, nil, "", 1, 10)
		if err == nil {
			t.Error("expected repo error")
		}
	})

	t.Run("List updates expired statuses", func(t *testing.T) {
		expiredDate := time.Now().Add(-1 * 24 * time.Hour)
		expired := validOportunidade()
		expired.DataExpiracao = &expiredDate
		expired.Status = models.StatusOportunidadeActive

		repo := &mockOportunidadeMEIRepo{
			listItems: []*models.OportunidadeMEI{expired},
			listTotal: 1,
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		filters := map[string]interface{}{
			"status": models.StatusOportunidadeExpired,
		}

		items, total, err := svc.List(ctx, filters, "", 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		// Should return 1 expired item after status update
		if len(items) != 1 {
			t.Errorf("expected 1 expired item, got %d", len(items))
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
	})
}

func TestOportunidadeMEIService_ListByStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("ListByStatus success", func(t *testing.T) {
		items := []*models.OportunidadeMEI{validOportunidade()}
		repo := &mockOportunidadeMEIRepo{
			listByStatusItems: items,
			listByStatusTotal: 1,
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		result, total, err := svc.ListByStatus(ctx, models.StatusOportunidadeActive, 1, 10)
		if err != nil {
			t.Fatalf("ListByStatus: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 item, got %d", len(result))
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		// Should call UpdateExpiredOpportunities
		if repo.updateExpiredCallCount != 1 {
			t.Errorf("expected UpdateExpiredOpportunities to be called once, got %d", repo.updateExpiredCallCount)
		}
	})

	t.Run("ListByStatus with update error (non-fatal)", func(t *testing.T) {
		items := []*models.OportunidadeMEI{validOportunidade()}
		repo := &mockOportunidadeMEIRepo{
			listByStatusItems: items,
			listByStatusTotal: 1,
			updateExpiredErr:  errors.New("update error"),
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		// Should not fail even if UpdateExpiredOpportunities fails
		result, total, err := svc.ListByStatus(ctx, models.StatusOportunidadeActive, 1, 10)
		if err != nil {
			t.Fatalf("ListByStatus: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 item, got %d", len(result))
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
	})
}

func TestOportunidadeMEIService_ListDrafts(t *testing.T) {
	ctx := context.Background()

	t.Run("ListDrafts without orgao filter", func(t *testing.T) {
		draft := validOportunidade()
		draft.Status = models.StatusOportunidadeDraft

		repo := &mockOportunidadeMEIRepo{
			listItems: []*models.OportunidadeMEI{draft},
			listTotal: 1,
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		result, total, err := svc.ListDrafts(ctx, "", "", 1, 10)
		if err != nil {
			t.Fatalf("ListDrafts: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 draft, got %d", len(result))
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
	})

	t.Run("ListDrafts with orgao filter", func(t *testing.T) {
		draft := validOportunidade()
		draft.Status = models.StatusOportunidadeDraft
		draft.OrgaoID = "org-123"

		repo := &mockOportunidadeMEIRepo{
			listItems: []*models.OportunidadeMEI{draft},
			listTotal: 1,
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		result, total, err := svc.ListDrafts(ctx, "org-123", "", 1, 10)
		if err != nil {
			t.Fatalf("ListDrafts: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 draft, got %d", len(result))
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
	})
}

func TestOportunidadeMEIService_ListActive(t *testing.T) {
	ctx := context.Background()

	t.Run("ListActive success", func(t *testing.T) {
		active := validOportunidade()
		repo := &mockOportunidadeMEIRepo{
			listByStatusItems: []*models.OportunidadeMEI{active},
			listByStatusTotal: 1,
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		result, total, err := svc.ListActive(ctx, 1, 10)
		if err != nil {
			t.Fatalf("ListActive: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 active item, got %d", len(result))
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
	})
}

func TestOportunidadeMEIService_ListByOrgao(t *testing.T) {
	ctx := context.Background()

	t.Run("ListByOrgao success", func(t *testing.T) {
		items := []*models.OportunidadeMEI{validOportunidade()}
		repo := &mockOportunidadeMEIRepo{
			listByOrgaoItems: items,
			listByOrgaoTotal: 1,
		}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		result, total, err := svc.ListByOrgao(ctx, "org-123", 1, 10)
		if err != nil {
			t.Fatalf("ListByOrgao: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 item, got %d", len(result))
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
	})

	t.Run("ListByOrgao with error", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{listByOrgaoErr: errors.New("db error")}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		_, _, err := svc.ListByOrgao(ctx, "org-123", 1, 10)
		if err == nil {
			t.Error("expected repo error")
		}
	})
}

func TestOportunidadeMEIService_UpdateExpiredOpportunities(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdateExpiredOpportunities success", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		err := svc.UpdateExpiredOpportunities(ctx)
		if err != nil {
			t.Fatalf("UpdateExpiredOpportunities: %v", err)
		}
	})

	t.Run("UpdateExpiredOpportunities with error", func(t *testing.T) {
		repo := &mockOportunidadeMEIRepo{updateExpiredErr: errors.New("update error")}
		svc := services.NewOportunidadeMEIServiceWithInterface(repo)

		err := svc.UpdateExpiredOpportunities(ctx)
		if err == nil {
			t.Error("expected update error")
		}
	})
}

func TestNewOportunidadeMEIService(t *testing.T) {
	service := services.NewOportunidadeMEIService(nil)

	if service == nil {
		t.Error("NewOportunidadeMEIService() returned nil")
	}
}

func TestNewOportunidadeMEIServiceWithInterface(t *testing.T) {
	repo := &mockOportunidadeMEIRepo{}
	service := services.NewOportunidadeMEIServiceWithInterface(repo)

	if service == nil {
		t.Error("NewOportunidadeMEIServiceWithInterface() returned nil")
	}
}
