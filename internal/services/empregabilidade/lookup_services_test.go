package empregabilidade_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// ─── RegimeContratacao ──────────────────────────────────────────────────────

type mockRegimeContratacaoRepo struct {
	items     map[uuid.UUID]*empregabilidade.RegimeContratacao
	createErr error
	getErr    error
	updateErr error
	deleteErr error
	listErr   error
}

func newMockRegimeContratacaoRepo() *mockRegimeContratacaoRepo {
	return &mockRegimeContratacaoRepo{items: make(map[uuid.UUID]*empregabilidade.RegimeContratacao)}
}

func (m *mockRegimeContratacaoRepo) Create(ctx context.Context, e *empregabilidade.RegimeContratacao) (uuid.UUID, error) {
	if m.createErr != nil {
		return uuid.Nil, m.createErr
	}
	e.ID = uuid.New()
	m.items[e.ID] = e
	return e.ID, nil
}

func (m *mockRegimeContratacaoRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.RegimeContratacao, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	e, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockRegimeContratacaoRepo) Update(ctx context.Context, e *empregabilidade.RegimeContratacao) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.items[e.ID] = e
	return nil
}

func (m *mockRegimeContratacaoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.items, id)
	return nil
}

func (m *mockRegimeContratacaoRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.RegimeContratacao, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var result []*empregabilidade.RegimeContratacao
	for _, v := range m.items {
		result = append(result, v)
	}
	return result, len(result), nil
}

func TestRegimeContratacaoService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockRegimeContratacaoRepo()
	svc := services.NewRegimeContratacaoServiceWithInterface(repo)

	t.Run("Create success", func(t *testing.T) {
		id, err := svc.Create(ctx, &empregabilidade.RegimeContratacao{Descricao: "CLT"})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if id == uuid.Nil {
			t.Error("Create: expected non-nil UUID")
		}
	})

	t.Run("Create error", func(t *testing.T) {
		repo.createErr = errors.New("db error")
		_, err := svc.Create(ctx, &empregabilidade.RegimeContratacao{Descricao: "PJ"})
		if err == nil {
			t.Error("Create: expected error")
		}
		repo.createErr = nil
	})

	t.Run("GetByID success", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.RegimeContratacao{Descricao: "Freelancer"})
		e, err := svc.GetByID(ctx, id)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v, entity=%v", err, e)
		}
		if e.Descricao != "Freelancer" {
			t.Errorf("GetByID: expected Freelancer, got %s", e.Descricao)
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		e, err := svc.GetByID(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if e != nil {
			t.Error("GetByID: expected nil")
		}
	})

	t.Run("Update success", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.RegimeContratacao{Descricao: "Estágio"})
		err := svc.Update(ctx, &empregabilidade.RegimeContratacao{ID: id, Descricao: "Estágio Updated"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete success", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.RegimeContratacao{Descricao: "Temporário"})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List success", func(t *testing.T) {
		items, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total < 0 {
			t.Error("List: total must be non-negative")
		}
		_ = items
	})

	t.Run("List error", func(t *testing.T) {
		repo.listErr = errors.New("db error")
		_, _, err := svc.List(ctx, nil, 1, 10)
		if err == nil {
			t.Error("List: expected error")
		}
		repo.listErr = nil
	})
}

// ─── ModeloTrabalho ──────────────────────────────────────────────────────────

type mockModeloTrabalhoRepo struct {
	items     map[uuid.UUID]*empregabilidade.ModeloTrabalho
	createErr error
	getErr    error
	updateErr error
	deleteErr error
	listErr   error
}

func newMockModeloTrabalhoRepo() *mockModeloTrabalhoRepo {
	return &mockModeloTrabalhoRepo{items: make(map[uuid.UUID]*empregabilidade.ModeloTrabalho)}
}

func (m *mockModeloTrabalhoRepo) Create(ctx context.Context, e *empregabilidade.ModeloTrabalho) (uuid.UUID, error) {
	if m.createErr != nil {
		return uuid.Nil, m.createErr
	}
	e.ID = uuid.New()
	m.items[e.ID] = e
	return e.ID, nil
}

func (m *mockModeloTrabalhoRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.ModeloTrabalho, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	e, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockModeloTrabalhoRepo) Update(ctx context.Context, e *empregabilidade.ModeloTrabalho) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.items[e.ID] = e
	return nil
}

func (m *mockModeloTrabalhoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.items, id)
	return nil
}

func (m *mockModeloTrabalhoRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.ModeloTrabalho, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var result []*empregabilidade.ModeloTrabalho
	for _, v := range m.items {
		result = append(result, v)
	}
	return result, len(result), nil
}

func TestModeloTrabalhoService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockModeloTrabalhoRepo()
	svc := services.NewModeloTrabalhoServiceWithInterface(repo)

	t.Run("Create", func(t *testing.T) {
		id, err := svc.Create(ctx, &empregabilidade.ModeloTrabalho{Descricao: "Presencial"})
		if err != nil || id == uuid.Nil {
			t.Fatalf("Create: err=%v, id=%v", err, id)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.ModeloTrabalho{Descricao: "Remoto"})
		e, err := svc.GetByID(ctx, id)
		if err != nil || e == nil {
			t.Fatalf("GetByID: err=%v, e=%v", err, e)
		}
	})

	t.Run("Update", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.ModeloTrabalho{Descricao: "Híbrido"})
		err := svc.Update(ctx, &empregabilidade.ModeloTrabalho{ID: id, Descricao: "Híbrido Updated"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.ModeloTrabalho{Descricao: "Temporário"})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		items, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		_ = items
		_ = total
	})
}

// ─── TipoPCD ─────────────────────────────────────────────────────────────────

type mockTipoPCDRepo struct {
	items     map[uuid.UUID]*empregabilidade.TipoPCD
	createErr error
	listErr   error
}

func newMockTipoPCDRepo() *mockTipoPCDRepo {
	return &mockTipoPCDRepo{items: make(map[uuid.UUID]*empregabilidade.TipoPCD)}
}

func (m *mockTipoPCDRepo) Create(ctx context.Context, e *empregabilidade.TipoPCD) (uuid.UUID, error) {
	if m.createErr != nil {
		return uuid.Nil, m.createErr
	}
	e.ID = uuid.New()
	m.items[e.ID] = e
	return e.ID, nil
}

func (m *mockTipoPCDRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoPCD, error) {
	e, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockTipoPCDRepo) Update(ctx context.Context, e *empregabilidade.TipoPCD) error {
	m.items[e.ID] = e
	return nil
}

func (m *mockTipoPCDRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockTipoPCDRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.TipoPCD, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var result []*empregabilidade.TipoPCD
	for _, v := range m.items {
		result = append(result, v)
	}
	return result, len(result), nil
}

func TestTipoPCDService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockTipoPCDRepo()
	svc := services.NewTipoPCDServiceWithInterface(repo)

	t.Run("Create", func(t *testing.T) {
		id, err := svc.Create(ctx, &empregabilidade.TipoPCD{Descricao: "Visual"})
		if err != nil || id == uuid.Nil {
			t.Fatalf("Create: %v / %v", err, id)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.TipoPCD{Descricao: "Auditiva"})
		e, err := svc.GetByID(ctx, id)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
	})

	t.Run("Update", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.TipoPCD{Descricao: "Física"})
		err := svc.Update(ctx, &empregabilidade.TipoPCD{ID: id, Descricao: "Física Motora"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.TipoPCD{Descricao: "Intelectual"})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		_, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total < 0 {
			t.Error("List: expected non-negative total")
		}
	})
}

// ─── Idioma ───────────────────────────────────────────────────────────────────

type mockIdiomaRepo struct {
	items map[uuid.UUID]*empregabilidade.Idioma
}

func newMockIdiomaRepo() *mockIdiomaRepo {
	return &mockIdiomaRepo{items: make(map[uuid.UUID]*empregabilidade.Idioma)}
}

func (m *mockIdiomaRepo) Create(ctx context.Context, e *empregabilidade.Idioma) (uuid.UUID, error) {
	e.ID = uuid.New()
	m.items[e.ID] = e
	return e.ID, nil
}

func (m *mockIdiomaRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Idioma, error) {
	e, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockIdiomaRepo) Update(ctx context.Context, e *empregabilidade.Idioma) error {
	m.items[e.ID] = e
	return nil
}

func (m *mockIdiomaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockIdiomaRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Idioma, int, error) {
	var result []*empregabilidade.Idioma
	for _, v := range m.items {
		result = append(result, v)
	}
	return result, len(result), nil
}

func TestIdiomaService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockIdiomaRepo()
	svc := services.NewIdiomaServiceWithInterface(repo)

	t.Run("Create", func(t *testing.T) {
		id, err := svc.Create(ctx, &empregabilidade.Idioma{Descricao: "Inglês"})
		if err != nil || id == uuid.Nil {
			t.Fatalf("Create: %v / %v", err, id)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Idioma{Descricao: "Português"})
		e, err := svc.GetByID(ctx, id)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
		if e.Descricao != "Português" {
			t.Errorf("GetByID: expected Português, got %s", e.Descricao)
		}
	})

	t.Run("Update", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Idioma{Descricao: "Espanhol"})
		err := svc.Update(ctx, &empregabilidade.Idioma{ID: id, Descricao: "Espanhol Updated"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Idioma{Descricao: "Francês"})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		items, total, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		_ = items
		_ = total
	})
}

// ─── NivelIdioma ─────────────────────────────────────────────────────────────

type mockNivelIdiomaRepo struct {
	items map[uuid.UUID]*empregabilidade.NivelIdioma
}

func newMockNivelIdiomaRepo() *mockNivelIdiomaRepo {
	return &mockNivelIdiomaRepo{items: make(map[uuid.UUID]*empregabilidade.NivelIdioma)}
}

func (m *mockNivelIdiomaRepo) Create(ctx context.Context, e *empregabilidade.NivelIdioma) (uuid.UUID, error) {
	e.ID = uuid.New()
	m.items[e.ID] = e
	return e.ID, nil
}

func (m *mockNivelIdiomaRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.NivelIdioma, error) {
	e, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockNivelIdiomaRepo) Update(ctx context.Context, e *empregabilidade.NivelIdioma) error {
	m.items[e.ID] = e
	return nil
}

func (m *mockNivelIdiomaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockNivelIdiomaRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.NivelIdioma, int, error) {
	var result []*empregabilidade.NivelIdioma
	for _, v := range m.items {
		result = append(result, v)
	}
	return result, len(result), nil
}

func TestNivelIdiomaService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockNivelIdiomaRepo()
	svc := services.NewNivelIdiomaServiceWithInterface(repo)

	t.Run("Create", func(t *testing.T) {
		id, err := svc.Create(ctx, &empregabilidade.NivelIdioma{Descricao: "Básico"})
		if err != nil || id == uuid.Nil {
			t.Fatalf("Create: %v / %v", err, id)
		}
	})

	t.Run("GetByID success", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.NivelIdioma{Descricao: "Intermediário"})
		e, err := svc.GetByID(ctx, id)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		e, err := svc.GetByID(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if e != nil {
			t.Error("expected nil")
		}
	})

	t.Run("Update", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.NivelIdioma{Descricao: "Avançado"})
		err := svc.Update(ctx, &empregabilidade.NivelIdioma{ID: id, Descricao: "Fluente"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.NivelIdioma{Descricao: "Nativo"})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		_, _, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
	})
}

// ─── EmpEscolaridade ─────────────────────────────────────────────────────────

type mockEmpEscolaridadeRepo struct {
	items map[uuid.UUID]*empregabilidade.Escolaridade
}

func newMockEmpEscolaridadeRepo() *mockEmpEscolaridadeRepo {
	return &mockEmpEscolaridadeRepo{items: make(map[uuid.UUID]*empregabilidade.Escolaridade)}
}

func (m *mockEmpEscolaridadeRepo) Create(ctx context.Context, e *empregabilidade.Escolaridade) (uuid.UUID, error) {
	e.ID = uuid.New()
	m.items[e.ID] = e
	return e.ID, nil
}

func (m *mockEmpEscolaridadeRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Escolaridade, error) {
	e, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockEmpEscolaridadeRepo) Update(ctx context.Context, e *empregabilidade.Escolaridade) error {
	m.items[e.ID] = e
	return nil
}

func (m *mockEmpEscolaridadeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockEmpEscolaridadeRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Escolaridade, int, error) {
	var result []*empregabilidade.Escolaridade
	for _, v := range m.items {
		result = append(result, v)
	}
	return result, len(result), nil
}

func TestEmpEscolaridadeService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockEmpEscolaridadeRepo()
	svc := services.NewEscolaridadeServiceWithInterface(repo)

	t.Run("Create", func(t *testing.T) {
		id, err := svc.Create(ctx, &empregabilidade.Escolaridade{Descricao: "Superior"})
		if err != nil || id == uuid.Nil {
			t.Fatalf("Create: %v / %v", err, id)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Escolaridade{Descricao: "Médio"})
		e, err := svc.GetByID(ctx, id)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
		if e.Descricao != "Médio" {
			t.Errorf("expected Médio, got %s", e.Descricao)
		}
	})

	t.Run("Update", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Escolaridade{Descricao: "Fundamental"})
		err := svc.Update(ctx, &empregabilidade.Escolaridade{ID: id, Descricao: "Fundamental Completo"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Escolaridade{Descricao: "Pós-graduação"})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		_, _, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
	})
}

// ─── TipoConquista ───────────────────────────────────────────────────────────

type mockTipoConquistaRepo struct {
	items map[uuid.UUID]*empregabilidade.TipoConquista
}

func newMockTipoConquistaRepo() *mockTipoConquistaRepo {
	return &mockTipoConquistaRepo{items: make(map[uuid.UUID]*empregabilidade.TipoConquista)}
}

func (m *mockTipoConquistaRepo) Create(ctx context.Context, e *empregabilidade.TipoConquista) (uuid.UUID, error) {
	e.ID = uuid.New()
	m.items[e.ID] = e
	return e.ID, nil
}

func (m *mockTipoConquistaRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.TipoConquista, error) {
	e, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockTipoConquistaRepo) Update(ctx context.Context, e *empregabilidade.TipoConquista) error {
	m.items[e.ID] = e
	return nil
}

func (m *mockTipoConquistaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockTipoConquistaRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.TipoConquista, int, error) {
	var result []*empregabilidade.TipoConquista
	for _, v := range m.items {
		result = append(result, v)
	}
	return result, len(result), nil
}

func TestTipoConquistaService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockTipoConquistaRepo()
	svc := services.NewTipoConquistaServiceWithInterface(repo)

	t.Run("Create", func(t *testing.T) {
		id, err := svc.Create(ctx, &empregabilidade.TipoConquista{Descricao: "Premiação"})
		if err != nil || id == uuid.Nil {
			t.Fatalf("Create: %v / %v", err, id)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.TipoConquista{Descricao: "Certificação"})
		e, err := svc.GetByID(ctx, id)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
	})

	t.Run("Update", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.TipoConquista{Descricao: "Publicação"})
		err := svc.Update(ctx, &empregabilidade.TipoConquista{ID: id, Descricao: "Publicação Científica"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.TipoConquista{Descricao: "Voluntariado"})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		_, _, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
	})
}

// ─── SituacaoAtual ───────────────────────────────────────────────────────────

type mockSituacaoAtualRepo struct {
	items map[uuid.UUID]*empregabilidade.SituacaoAtual
}

func newMockSituacaoAtualRepo() *mockSituacaoAtualRepo {
	return &mockSituacaoAtualRepo{items: make(map[uuid.UUID]*empregabilidade.SituacaoAtual)}
}

func (m *mockSituacaoAtualRepo) Create(ctx context.Context, e *empregabilidade.SituacaoAtual) (uuid.UUID, error) {
	e.ID = uuid.New()
	m.items[e.ID] = e
	return e.ID, nil
}

func (m *mockSituacaoAtualRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.SituacaoAtual, error) {
	e, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockSituacaoAtualRepo) Update(ctx context.Context, e *empregabilidade.SituacaoAtual) error {
	m.items[e.ID] = e
	return nil
}

func (m *mockSituacaoAtualRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockSituacaoAtualRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.SituacaoAtual, int, error) {
	var result []*empregabilidade.SituacaoAtual
	for _, v := range m.items {
		result = append(result, v)
	}
	return result, len(result), nil
}

func TestSituacaoAtualService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockSituacaoAtualRepo()
	svc := services.NewSituacaoAtualServiceWithInterface(repo)

	t.Run("Create", func(t *testing.T) {
		id, err := svc.Create(ctx, &empregabilidade.SituacaoAtual{Descricao: "Empregado"})
		if err != nil || id == uuid.Nil {
			t.Fatalf("Create: %v / %v", err, id)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.SituacaoAtual{Descricao: "Desempregado"})
		e, err := svc.GetByID(ctx, id)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
	})

	t.Run("Update", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.SituacaoAtual{Descricao: "Estudante"})
		err := svc.Update(ctx, &empregabilidade.SituacaoAtual{ID: id, Descricao: "Estudante Universitário"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.SituacaoAtual{Descricao: "Aposentado"})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		_, _, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
	})
}

// ─── Disponibilidade ─────────────────────────────────────────────────────────

type mockDisponibilidadeRepo struct {
	items map[uuid.UUID]*empregabilidade.Disponibilidade
}

func newMockDisponibilidadeRepo() *mockDisponibilidadeRepo {
	return &mockDisponibilidadeRepo{items: make(map[uuid.UUID]*empregabilidade.Disponibilidade)}
}

func (m *mockDisponibilidadeRepo) Create(ctx context.Context, e *empregabilidade.Disponibilidade) (uuid.UUID, error) {
	e.ID = uuid.New()
	m.items[e.ID] = e
	return e.ID, nil
}

func (m *mockDisponibilidadeRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Disponibilidade, error) {
	e, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockDisponibilidadeRepo) Update(ctx context.Context, e *empregabilidade.Disponibilidade) error {
	m.items[e.ID] = e
	return nil
}

func (m *mockDisponibilidadeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockDisponibilidadeRepo) List(ctx context.Context, filter map[string]interface{}, limit, offset int) ([]*empregabilidade.Disponibilidade, int, error) {
	var result []*empregabilidade.Disponibilidade
	for _, v := range m.items {
		result = append(result, v)
	}
	return result, len(result), nil
}

func TestDisponibilidadeService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockDisponibilidadeRepo()
	svc := services.NewDisponibilidadeServiceWithInterface(repo)

	t.Run("Create", func(t *testing.T) {
		id, err := svc.Create(ctx, &empregabilidade.Disponibilidade{Descricao: "Imediata"})
		if err != nil || id == uuid.Nil {
			t.Fatalf("Create: %v / %v", err, id)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Disponibilidade{Descricao: "Em 30 dias"})
		e, err := svc.GetByID(ctx, id)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
	})

	t.Run("Update", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Disponibilidade{Descricao: "A combinar"})
		err := svc.Update(ctx, &empregabilidade.Disponibilidade{ID: id, Descricao: "Negociável"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Disponibilidade{Descricao: "Em 60 dias"})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		_, _, err := svc.List(ctx, nil, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
	})
}

// ─── Empresa ─────────────────────────────────────────────────────────────────

type mockEmpresaRepo struct {
	items    map[string]*empregabilidade.Empresa
	createErr error
}

func newMockEmpresaRepo() *mockEmpresaRepo {
	return &mockEmpresaRepo{items: make(map[string]*empregabilidade.Empresa)}
}

func (m *mockEmpresaRepo) Create(ctx context.Context, e *empregabilidade.Empresa) (string, error) {
	if m.createErr != nil {
		return "", m.createErr
	}
	m.items[e.CNPJ] = e
	return e.CNPJ, nil
}

func (m *mockEmpresaRepo) GetByID(ctx context.Context, cnpj string) (*empregabilidade.Empresa, error) {
	e, ok := m.items[cnpj]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockEmpresaRepo) Update(ctx context.Context, e *empregabilidade.Empresa) error {
	m.items[e.CNPJ] = e
	return nil
}

func (m *mockEmpresaRepo) Delete(ctx context.Context, cnpj string) error {
	delete(m.items, cnpj)
	return nil
}

func (m *mockEmpresaRepo) List(ctx context.Context, filter empregabilidade.EmpresaFilter, limit, offset int) ([]*empregabilidade.Empresa, int, error) {
	var result []*empregabilidade.Empresa
	for _, v := range m.items {
		result = append(result, v)
	}
	return result, len(result), nil
}

func (m *mockEmpresaRepo) Upsert(ctx context.Context, e *empregabilidade.Empresa) error {
	m.items[e.CNPJ] = e
	return nil
}

func TestEmpresaService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockEmpresaRepo()
	svc := services.NewEmpresaServiceWithInterface(repo)

	t.Run("Create", func(t *testing.T) {
		cnpj, err := svc.Create(ctx, &empregabilidade.Empresa{CNPJ: "12345678000190", RazaoSocial: "Empresa Teste"})
		if err != nil || cnpj == "" {
			t.Fatalf("Create: %v / %v", err, cnpj)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		svc.Create(ctx, &empregabilidade.Empresa{CNPJ: "98765432000199", RazaoSocial: "Outra Empresa"})
		e, err := svc.GetByID(ctx, "98765432000199")
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
		if e.RazaoSocial != "Outra Empresa" {
			t.Errorf("expected Outra Empresa, got %s", e.RazaoSocial)
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		e, err := svc.GetByID(ctx, "00000000000000")
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if e != nil {
			t.Error("expected nil")
		}
	})

	t.Run("Update", func(t *testing.T) {
		svc.Create(ctx, &empregabilidade.Empresa{CNPJ: "11111111000111", RazaoSocial: "Original"})
		err := svc.Update(ctx, &empregabilidade.Empresa{CNPJ: "11111111000111", RazaoSocial: "Updated"})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		svc.Create(ctx, &empregabilidade.Empresa{CNPJ: "22222222000122", RazaoSocial: "Para Deletar"})
		err := svc.Delete(ctx, "22222222000122")
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("List", func(t *testing.T) {
		_, _, err := svc.List(ctx, empregabilidade.EmpresaFilter{}, 1, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
	})

	t.Run("Upsert", func(t *testing.T) {
		err := svc.Upsert(ctx, &empregabilidade.Empresa{CNPJ: "33333333000133", RazaoSocial: "Upserted"})
		if err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	})
}

// ─── Onboarding ───────────────────────────────────────────────────────────────

type mockOnboardingRepo struct {
	items map[string]*empregabilidade.Onboarding
	getErr error
}

func newMockOnboardingRepo() *mockOnboardingRepo {
	return &mockOnboardingRepo{items: make(map[string]*empregabilidade.Onboarding)}
}

func (m *mockOnboardingRepo) GetByCPF(ctx context.Context, cpf string) (*empregabilidade.Onboarding, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	e, ok := m.items[cpf]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockOnboardingRepo) Upsert(ctx context.Context, e *empregabilidade.Onboarding) error {
	m.items[e.CPF] = e
	return nil
}

func (m *mockOnboardingRepo) MarkFirstLoginCompleted(ctx context.Context, cpf string) error {
	if e, ok := m.items[cpf]; ok {
		e.IsEmpregabilidadeFirstLogin = false
	}
	return nil
}

func TestOnboardingService(t *testing.T) {
	ctx := context.Background()
	repo := newMockOnboardingRepo()
	svc := services.NewOnboardingServiceWithInterface(repo)

	t.Run("IsFirstLogin true when no record", func(t *testing.T) {
		isFirst, err := svc.IsFirstLogin(ctx, "12345678900")
		if err != nil {
			t.Fatalf("IsFirstLogin: %v", err)
		}
		if !isFirst {
			t.Error("expected true for new user")
		}
	})

	t.Run("IsFirstLogin false after mark completed", func(t *testing.T) {
		repo.Upsert(ctx, &empregabilidade.Onboarding{CPF: "98765432100", IsEmpregabilidadeFirstLogin: true})
		svc.MarkFirstLoginCompleted(ctx, "98765432100")
		isFirst, err := svc.IsFirstLogin(ctx, "98765432100")
		if err != nil {
			t.Fatalf("IsFirstLogin: %v", err)
		}
		if isFirst {
			t.Error("expected false after completing")
		}
	})

	t.Run("IsFirstLogin error propagates", func(t *testing.T) {
		repo.getErr = errors.New("db error")
		_, err := svc.IsFirstLogin(ctx, "anyuser")
		if err == nil {
			t.Error("expected error")
		}
		repo.getErr = nil
	})

	t.Run("GetByCPF", func(t *testing.T) {
		repo.Upsert(ctx, &empregabilidade.Onboarding{CPF: "11111111111"})
		e, err := svc.GetByCPF(ctx, "11111111111")
		if err != nil || e == nil {
			t.Fatalf("GetByCPF: %v / %v", err, e)
		}
	})

	t.Run("Upsert", func(t *testing.T) {
		err := svc.Upsert(ctx, &empregabilidade.Onboarding{CPF: "22222222222"})
		if err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	})

	t.Run("MarkFirstLoginCompleted", func(t *testing.T) {
		err := svc.MarkFirstLoginCompleted(ctx, "33333333333")
		if err != nil {
			t.Fatalf("MarkFirstLoginCompleted: %v", err)
		}
	})
}

// ─── TermosUso ────────────────────────────────────────────────────────────────

type mockTermosUsoRepo struct {
	items  map[string]*empregabilidade.TermosUso
	getErr error
}

func newMockTermosUsoRepo() *mockTermosUsoRepo {
	return &mockTermosUsoRepo{items: make(map[string]*empregabilidade.TermosUso)}
}

func (m *mockTermosUsoRepo) GetByCPF(ctx context.Context, cpf string) (*empregabilidade.TermosUso, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	e, ok := m.items[cpf]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockTermosUsoRepo) Upsert(ctx context.Context, e *empregabilidade.TermosUso) error {
	m.items[e.CPF] = e
	return nil
}

func (m *mockTermosUsoRepo) AcceptTerms(ctx context.Context, cpf string) error {
	if e, ok := m.items[cpf]; ok {
		e.UserConsent = true
	}
	return nil
}

func TestTermosUsoService(t *testing.T) {
	ctx := context.Background()
	repo := newMockTermosUsoRepo()
	svc := services.NewTermosUsoServiceWithInterface(repo)

	t.Run("HasAcceptedTerms false when no record", func(t *testing.T) {
		accepted, err := svc.HasAcceptedTerms(ctx, "12345678900")
		if err != nil {
			t.Fatalf("HasAcceptedTerms: %v", err)
		}
		if accepted {
			t.Error("expected false for new user")
		}
	})

	t.Run("HasAcceptedTerms true after accept", func(t *testing.T) {
		repo.Upsert(ctx, &empregabilidade.TermosUso{CPF: "98765432100", UserConsent: false})
		svc.AcceptTerms(ctx, "98765432100")
		accepted, err := svc.HasAcceptedTerms(ctx, "98765432100")
		if err != nil {
			t.Fatalf("HasAcceptedTerms: %v", err)
		}
		if !accepted {
			t.Error("expected true after accepting")
		}
	})

	t.Run("HasAcceptedTerms error propagates", func(t *testing.T) {
		repo.getErr = errors.New("db error")
		_, err := svc.HasAcceptedTerms(ctx, "any")
		if err == nil {
			t.Error("expected error")
		}
		repo.getErr = nil
	})

	t.Run("GetByCPF", func(t *testing.T) {
		repo.Upsert(ctx, &empregabilidade.TermosUso{CPF: "11111111111"})
		e, err := svc.GetByCPF(ctx, "11111111111")
		if err != nil || e == nil {
			t.Fatalf("GetByCPF: %v / %v", err, e)
		}
	})

	t.Run("Upsert", func(t *testing.T) {
		err := svc.Upsert(ctx, &empregabilidade.TermosUso{CPF: "22222222222"})
		if err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	})

	t.Run("AcceptTerms", func(t *testing.T) {
		err := svc.AcceptTerms(ctx, "33333333333")
		if err != nil {
			t.Fatalf("AcceptTerms: %v", err)
		}
	})
}

// ─── Etapa ────────────────────────────────────────────────────────────────────

type mockEtapaRepo struct {
	items map[uuid.UUID]*empregabilidade.Etapa
}

func newMockEtapaRepo() *mockEtapaRepo {
	return &mockEtapaRepo{items: make(map[uuid.UUID]*empregabilidade.Etapa)}
}

func (m *mockEtapaRepo) Create(ctx context.Context, e *empregabilidade.Etapa) (uuid.UUID, error) {
	e.ID = uuid.New()
	m.items[e.ID] = e
	return e.ID, nil
}

func (m *mockEtapaRepo) GetByID(ctx context.Context, id uuid.UUID) (*empregabilidade.Etapa, error) {
	e, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockEtapaRepo) Update(ctx context.Context, e *empregabilidade.Etapa) error {
	m.items[e.ID] = e
	return nil
}

func (m *mockEtapaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *mockEtapaRepo) ListByVaga(ctx context.Context, vagaID uuid.UUID) ([]*empregabilidade.Etapa, error) {
	var result []*empregabilidade.Etapa
	for _, v := range m.items {
		if v.IDVaga == vagaID {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *mockEtapaRepo) DeleteByVaga(ctx context.Context, vagaID uuid.UUID) error {
	for id, v := range m.items {
		if v.IDVaga == vagaID {
			delete(m.items, id)
		}
	}
	return nil
}

func TestEtapaService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockEtapaRepo()
	svc := services.NewEtapaServiceWithInterface(repo)
	vagaID := uuid.New()

	t.Run("Create", func(t *testing.T) {
		id, err := svc.Create(ctx, &empregabilidade.Etapa{IDVaga: vagaID, Titulo: "Triagem", Ordem: 1})
		if err != nil || id == uuid.Nil {
			t.Fatalf("Create: %v / %v", err, id)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Etapa{IDVaga: vagaID, Titulo: "Entrevista", Ordem: 2})
		e, err := svc.GetByID(ctx, id)
		if err != nil || e == nil {
			t.Fatalf("GetByID: %v / %v", err, e)
		}
		if e.Titulo != "Entrevista" {
			t.Errorf("expected Entrevista, got %s", e.Titulo)
		}
	})

	t.Run("Update", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Etapa{IDVaga: vagaID, Titulo: "Teste", Ordem: 3})
		err := svc.Update(ctx, &empregabilidade.Etapa{ID: id, IDVaga: vagaID, Titulo: "Teste Técnico", Ordem: 3})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		id, _ := svc.Create(ctx, &empregabilidade.Etapa{IDVaga: vagaID, Titulo: "Dinâmica", Ordem: 4})
		err := svc.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}
	})

	t.Run("ListByVaga", func(t *testing.T) {
		svc.Create(ctx, &empregabilidade.Etapa{IDVaga: vagaID, Titulo: "Etapa 1", Ordem: 5})
		svc.Create(ctx, &empregabilidade.Etapa{IDVaga: vagaID, Titulo: "Etapa 2", Ordem: 6})
		etapas, err := svc.ListByVaga(ctx, vagaID)
		if err != nil {
			t.Fatalf("ListByVaga: %v", err)
		}
		if len(etapas) < 2 {
			t.Errorf("ListByVaga: expected >= 2, got %d", len(etapas))
		}
	})

	t.Run("DeleteByVaga", func(t *testing.T) {
		err := svc.DeleteByVaga(ctx, vagaID)
		if err != nil {
			t.Fatalf("DeleteByVaga: %v", err)
		}
	})
}
