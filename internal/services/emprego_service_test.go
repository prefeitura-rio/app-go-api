package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/prefeitura-rio/app-go-api/internal/services"
)

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
	if m.createID != 0 {
		return m.createID, nil
	}
	return 1, nil
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

func TestEmpregoService_Create(t *testing.T) {
	tests := []struct {
		name      string
		emprego   *models.Emprego
		repoID    int
		repoErr   error
		wantID    int
		wantErr   bool
		errMsg    string
	}{
		{
			name: "success - valid CLT job",
			emprego: &models.Emprego{
				Titulo:          "Desenvolvedor Backend",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
			},
			repoID: 10,
			wantID: 10,
		},
		{
			name: "success - valid ESTAGIO job",
			emprego: &models.Emprego{
				Titulo:          "Estagiário TI",
				TipoContratacao: models.TipoContratacaoEstagio,
				Status:          models.StatusEmpregoCriado,
			},
			repoID: 20,
			wantID: 20,
		},
		{
			name: "success - valid MEI job",
			emprego: &models.Emprego{
				Titulo:          "Prestador de Serviços",
				TipoContratacao: models.TipoContratacaoMEI,
				Status:          models.StatusEmpregoAberto,
				JornadaTrabalho: models.JornadaFlexivel,
			},
			repoID: 30,
			wantID: 30,
		},
		{
			name: "error - missing title",
			emprego: &models.Emprego{
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
			},
			wantErr: true,
			errMsg:  "título é obrigatório",
		},
		{
			name: "error - empty title",
			emprego: &models.Emprego{
				Titulo:          "   ",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
			},
			wantErr: true,
			errMsg:  "título é obrigatório",
		},
		{
			name: "error - missing tipo_contratacao",
			emprego: &models.Emprego{
				Titulo: "Desenvolvedor",
				Status: models.StatusEmpregoAberto,
			},
			wantErr: true,
			errMsg:  "tipo de contratação é obrigatório",
		},
		{
			name: "error - invalid tipo_contratacao",
			emprego: &models.Emprego{
				Titulo:          "Desenvolvedor",
				TipoContratacao: "INVALID",
				Status:          models.StatusEmpregoAberto,
			},
			wantErr: true,
			errMsg:  "tipo de contratação inválido",
		},
		{
			name: "error - missing status",
			emprego: &models.Emprego{
				Titulo:          "Desenvolvedor",
				TipoContratacao: models.TipoContratacaoCLT,
			},
			wantErr: true,
			errMsg:  "status é obrigatório",
		},
		{
			name: "error - invalid status",
			emprego: &models.Emprego{
				Titulo:          "Desenvolvedor",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          "INVALID",
			},
			wantErr: true,
			errMsg:  "status inválido",
		},
		{
			name: "error - invalid jornada_trabalho",
			emprego: &models.Emprego{
				Titulo:          "Desenvolvedor",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
				JornadaTrabalho: "INVALID",
			},
			wantErr: true,
			errMsg:  "jornada de trabalho inválida",
		},
		{
			name: "success - valid jornada_trabalho",
			emprego: &models.Emprego{
				Titulo:          "Desenvolvedor",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
				JornadaTrabalho: models.JornadaIntegral,
			},
			repoID: 40,
			wantID: 40,
		},
		{
			name: "repo error",
			emprego: &models.Emprego{
				Titulo:          "Desenvolvedor",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEmpregoRepo{createID: tt.repoID, createErr: tt.repoErr}
			svc := services.NewEmpregoServiceWithInterface(repo)
			ctx := context.Background()

			id, err := svc.Create(ctx, tt.emprego)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Create() error = %v, want contains %s", err, tt.errMsg)
				}
			}
			if !tt.wantErr && id != tt.wantID {
				t.Errorf("Create() id = %d, want %d", id, tt.wantID)
			}
		})
	}
}

func TestEmpregoService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		entity  *models.Emprego
		getErr  error
		wantNil bool
		wantErr bool
	}{
		{
			name: "success",
			id:   1,
			entity: &models.Emprego{
				ID:              1,
				Titulo:          "Desenvolvedor",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
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
			repo := &mockEmpregoRepo{entity: tt.entity, getErr: tt.getErr}
			svc := services.NewEmpregoServiceWithInterface(repo)
			ctx := context.Background()

			emprego, err := svc.GetByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantNil && emprego != nil {
				t.Error("GetByID() expected nil emprego")
			}
			if !tt.wantNil && !tt.wantErr && emprego == nil {
				t.Error("GetByID() expected non-nil emprego")
			}
			if emprego != nil && emprego.ID != tt.id {
				t.Errorf("GetByID() expected id %d, got %d", tt.id, emprego.ID)
			}
		})
	}
}

func TestEmpregoService_Update(t *testing.T) {
	tests := []struct {
		name      string
		emprego   *models.Emprego
		repoErr   error
		wantErr   bool
		errMsg    string
	}{
		{
			name: "success",
			emprego: &models.Emprego{
				ID:              1,
				Titulo:          "Desenvolvedor Sênior",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
			},
		},
		{
			name: "error - validation fails (missing title)",
			emprego: &models.Emprego{
				ID:              1,
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
			},
			wantErr: true,
			errMsg:  "título é obrigatório",
		},
		{
			name: "error - validation fails (invalid tipo_contratacao)",
			emprego: &models.Emprego{
				ID:              1,
				Titulo:          "Desenvolvedor",
				TipoContratacao: "INVALID",
				Status:          models.StatusEmpregoAberto,
			},
			wantErr: true,
			errMsg:  "tipo de contratação inválido",
		},
		{
			name: "error - validation fails (invalid status)",
			emprego: &models.Emprego{
				ID:              1,
				Titulo:          "Desenvolvedor",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          "INVALID",
			},
			wantErr: true,
			errMsg:  "status inválido",
		},
		{
			name: "success - update with jornada_trabalho",
			emprego: &models.Emprego{
				ID:              1,
				Titulo:          "Desenvolvedor",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
				JornadaTrabalho: models.JornadaHomeOffice,
			},
		},
		{
			name: "repo error",
			emprego: &models.Emprego{
				ID:              1,
				Titulo:          "Desenvolvedor",
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
			},
			repoErr: errors.New("update failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEmpregoRepo{updateErr: tt.repoErr}
			svc := services.NewEmpregoServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.Update(ctx, tt.emprego)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Update() error = %v, want contains %s", err, tt.errMsg)
				}
			}
		})
	}
}

func TestEmpregoService_Delete(t *testing.T) {
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
			name:    "repo error",
			id:      1,
			repoErr: errors.New("delete failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEmpregoRepo{deleteErr: tt.repoErr}
			svc := services.NewEmpregoServiceWithInterface(repo)
			ctx := context.Background()

			err := svc.Delete(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmpregoService_List(t *testing.T) {
	emprego1 := &models.Emprego{
		ID:              1,
		Titulo:          "Desenvolvedor Backend",
		TipoContratacao: models.TipoContratacaoCLT,
		Status:          models.StatusEmpregoAberto,
	}
	emprego2 := &models.Emprego{
		ID:              2,
		Titulo:          "Estagiário TI",
		TipoContratacao: models.TipoContratacaoEstagio,
		Status:          models.StatusEmpregoCriado,
	}
	emprego3 := &models.Emprego{
		ID:              3,
		Titulo:          "Desenvolvedor Frontend",
		TipoContratacao: models.TipoContratacaoCLT,
		Status:          models.StatusEmpregoEncerrado,
	}

	tests := []struct {
		name      string
		filter    map[string]interface{}
		page      int
		pageSize  int
		listItems []*models.Emprego
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
			listItems: []*models.Emprego{},
			listTotal: 0,
			wantLen:   0,
			wantTotal: 0,
		},
		{
			name:      "success - with items",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  10,
			listItems: []*models.Emprego{emprego1, emprego2, emprego3},
			listTotal: 3,
			wantLen:   3,
			wantTotal: 3,
		},
		{
			name:      "success - filter by status",
			filter:    map[string]interface{}{"status": "ABERTO"},
			page:      1,
			pageSize:  10,
			listItems: []*models.Emprego{emprego1},
			listTotal: 1,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:      "success - filter by tipo_contratacao",
			filter:    map[string]interface{}{"tipo_contratacao": "CLT"},
			page:      1,
			pageSize:  10,
			listItems: []*models.Emprego{emprego1, emprego3},
			listTotal: 2,
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name:      "success - pagination page 1",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  2,
			listItems: []*models.Emprego{emprego1, emprego2},
			listTotal: 3,
			wantLen:   2,
			wantTotal: 3,
		},
		{
			name:      "success - pagination page 2",
			filter:    map[string]interface{}{},
			page:      2,
			pageSize:  2,
			listItems: []*models.Emprego{emprego3},
			listTotal: 3,
			wantLen:   1,
			wantTotal: 3,
		},
		{
			name:      "success - large page size",
			filter:    map[string]interface{}{},
			page:      1,
			pageSize:  100,
			listItems: []*models.Emprego{emprego1, emprego2, emprego3},
			listTotal: 3,
			wantLen:   3,
			wantTotal: 3,
		},
		{
			name:      "success - filter by empresa_id",
			filter:    map[string]interface{}{"empresa_id": 5},
			page:      1,
			pageSize:  10,
			listItems: []*models.Emprego{emprego1},
			listTotal: 1,
			wantLen:   1,
			wantTotal: 1,
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
			repo := &mockEmpregoRepo{
				listItems: tt.listItems,
				listTotal: tt.listTotal,
				listErr:   tt.listErr,
			}
			svc := services.NewEmpregoServiceWithInterface(repo)
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

func TestEmpregoService_Create_AllTipoContratacao(t *testing.T) {
	types := []models.TipoContratacao{
		models.TipoContratacaoCLT,
		models.TipoContratacaoEstagio,
		models.TipoContratacaoJovemAprendiz,
		models.TipoContratacaoMEI,
		models.TipoContratacaoPJ,
	}

	for _, tc := range types {
		t.Run(string(tc), func(t *testing.T) {
			repo := &mockEmpregoRepo{createID: 1}
			svc := services.NewEmpregoServiceWithInterface(repo)
			ctx := context.Background()

			emprego := &models.Emprego{
				Titulo:          "Teste " + string(tc),
				TipoContratacao: tc,
				Status:          models.StatusEmpregoAberto,
			}

			id, err := svc.Create(ctx, emprego)
			if err != nil {
				t.Errorf("Create() with %s failed: %v", tc, err)
			}
			if id != 1 {
				t.Errorf("Create() with %s: expected id 1, got %d", tc, id)
			}
		})
	}
}

func TestEmpregoService_Create_AllStatus(t *testing.T) {
	statuses := []models.StatusEmprego{
		models.StatusEmpregoCriado,
		models.StatusEmpregoAberto,
		models.StatusEmpregoEmAnalise,
		models.StatusEmpregoEncerrado,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			repo := &mockEmpregoRepo{createID: 1}
			svc := services.NewEmpregoServiceWithInterface(repo)
			ctx := context.Background()

			emprego := &models.Emprego{
				Titulo:          "Teste " + string(status),
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          status,
			}

			id, err := svc.Create(ctx, emprego)
			if err != nil {
				t.Errorf("Create() with %s failed: %v", status, err)
			}
			if id != 1 {
				t.Errorf("Create() with %s: expected id 1, got %d", status, id)
			}
		})
	}
}

func TestEmpregoService_Create_AllJornadaTrabalho(t *testing.T) {
	jornadas := []models.JornadaTrabalho{
		models.JornadaIntegral,
		models.JornadaMeioPeriodo,
		models.JornadaFlexivel,
		models.JornadaEstagio,
		models.JornadaHomeOffice,
	}

	for _, jornada := range jornadas {
		t.Run(string(jornada), func(t *testing.T) {
			repo := &mockEmpregoRepo{createID: 1}
			svc := services.NewEmpregoServiceWithInterface(repo)
			ctx := context.Background()

			emprego := &models.Emprego{
				Titulo:          "Teste " + string(jornada),
				TipoContratacao: models.TipoContratacaoCLT,
				Status:          models.StatusEmpregoAberto,
				JornadaTrabalho: jornada,
			}

			id, err := svc.Create(ctx, emprego)
			if err != nil {
				t.Errorf("Create() with %s failed: %v", jornada, err)
			}
			if id != 1 {
				t.Errorf("Create() with %s: expected id 1, got %d", jornada, id)
			}
		})
	}
}
