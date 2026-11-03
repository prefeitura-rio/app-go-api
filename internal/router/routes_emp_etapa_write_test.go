package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prefeitura-rio/app-go-api/internal/config"
	empHandlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empModels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// etapaWriteRepoStub guarda as etapas em memória e reproduz o comportamento
// antigo do repositório (Update grava pelo id, inclusive movendo a etapa de
// vaga; Delete apaga pelo id ignorando a vaga). Assim o teste prova que a
// conferência da vaga dona da etapa acontece antes de chegar ao repositório.
type etapaWriteRepoStub struct {
	etapas map[uuid.UUID]*empModels.Etapa
}

func (s *etapaWriteRepoStub) Create(_ context.Context, e *empModels.Etapa) (uuid.UUID, error) {
	e.ID = uuid.New()
	copied := *e
	s.etapas[e.ID] = &copied
	return e.ID, nil
}

func (s *etapaWriteRepoStub) GetByID(_ context.Context, id uuid.UUID) (*empModels.Etapa, error) {
	e, ok := s.etapas[id]
	if !ok {
		return nil, nil
	}
	copied := *e
	return &copied, nil
}

func (s *etapaWriteRepoStub) Update(_ context.Context, e *empModels.Etapa) error {
	copied := *e
	s.etapas[e.ID] = &copied
	return nil
}

func (s *etapaWriteRepoStub) Delete(_ context.Context, _, id uuid.UUID) error {
	delete(s.etapas, id)
	return nil
}

func (s *etapaWriteRepoStub) ListByVaga(_ context.Context, vagaID uuid.UUID) ([]*empModels.Etapa, error) {
	var out []*empModels.Etapa
	for _, e := range s.etapas {
		if e.IDVaga == vagaID {
			copied := *e
			out = append(out, &copied)
		}
	}
	return out, nil
}

func (s *etapaWriteRepoStub) DeleteByVaga(_ context.Context, vagaID uuid.UUID) error {
	for id, e := range s.etapas {
		if e.IDVaga == vagaID {
			delete(s.etapas, id)
		}
	}
	return nil
}

// PUT e DELETE /api/v1/empregabilidade/vagas/:id/etapas/:etapaId só agem sobre
// etapa cuja vaga é a do path (DEV-139). vagaOwnership confere o órgão da vaga
// do path; uma etapa de vaga de outro órgão responde 404 e fica intacta.
func TestRegisterEmpregabilidadeRoutes_EtapaWrite_EscopoDaVaga(t *testing.T) {
	gin.SetMode(gin.TestMode)

	vagaA := uuid.MustParse("11111111-1111-4111-8111-111111111111") // órgão A
	vagaB := uuid.MustParse("22222222-2222-4222-8222-222222222222") // órgão B
	etapaA := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	etapaB := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	inexistente := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")

	vagaRepo := &vagaGetRepoStub{vagas: map[uuid.UUID]*empModels.Vaga{
		vagaA: {ID: vagaA, Titulo: "Vaga A", IDOrgaoParceiro: "orgao-a", Status: empModels.StatusVagaEmEdicao},
		vagaB: {ID: vagaB, Titulo: "Vaga B", IDOrgaoParceiro: "orgao-b", Status: empModels.StatusVagaEmEdicao},
	}}
	vagaSvc := empServices.NewVagaServiceWithInterfaces(vagaRepo, nil, nil)

	editorA := injectCaller("", []string{"go:empregabilidade:editor"}, "orgao-a")
	empAdmin := injectCaller("", []string{"go:empregabilidade:admin"}, "")

	tests := []struct {
		name   string
		caller gin.HandlerFunc
		method string
		vaga   uuid.UUID
		etapa  uuid.UUID
		want   int
		// estado esperado da etapa A depois da chamada: "intacta", "alterada" ou "excluida"
		wantEtapaA string
	}{
		{"editor A, DELETE etapa da vaga B pela vaga A", editorA, http.MethodDelete, vagaA, etapaB, http.StatusNotFound, "intacta"},
		{"editor A, PUT etapa da vaga B pela vaga A", editorA, http.MethodPut, vagaA, etapaB, http.StatusNotFound, "intacta"},
		{"emp admin, PUT etapa da vaga B pela vaga A", empAdmin, http.MethodPut, vagaA, etapaB, http.StatusNotFound, "intacta"},
		{"emp admin, DELETE etapa da vaga B pela vaga A", empAdmin, http.MethodDelete, vagaA, etapaB, http.StatusNotFound, "intacta"},
		{"editor A, PUT etapa inexistente pela vaga A", editorA, http.MethodPut, vagaA, inexistente, http.StatusNotFound, "intacta"},
		{"editor A, DELETE etapa inexistente pela vaga A", editorA, http.MethodDelete, vagaA, inexistente, http.StatusNotFound, "intacta"},
		{"editor A, PUT etapa da vaga B pela vaga B", editorA, http.MethodPut, vagaB, etapaB, http.StatusForbidden, "intacta"},
		{"editor A, DELETE etapa da vaga B pela vaga B", editorA, http.MethodDelete, vagaB, etapaB, http.StatusForbidden, "intacta"},
		{"editor A, PUT etapa da vaga A pela vaga A", editorA, http.MethodPut, vagaA, etapaA, http.StatusOK, "alterada"},
		{"editor A, DELETE etapa da vaga A pela vaga A", editorA, http.MethodDelete, vagaA, etapaA, http.StatusOK, "excluida"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etapaRepo := &etapaWriteRepoStub{etapas: map[uuid.UUID]*empModels.Etapa{
				etapaA: {ID: etapaA, IDVaga: vagaA, Titulo: "Triagem A", Ordem: 1},
				etapaB: {ID: etapaB, IDVaga: vagaB, Titulo: "Triagem B", Ordem: 1},
			}}
			etapaSvc := empServices.NewEtapaServiceWithInterface(etapaRepo)

			router := gin.New()
			router.Use(tt.caller)
			app := createMockAppContainer()
			app.EmpVagaService = vagaSvc
			app.EmpVagaHandler = empHandlers.NewVagaHandler(vagaSvc)
			app.EmpEtapaService = etapaSvc
			app.EmpEtapaHandler = empHandlers.NewEtapaHandler(etapaSvc)
			cfg := &config.AppConfig{App: config.AppSettings{RBACEnabled: true}}
			registerEmpregabilidadeRoutes(router.Group("/api/v1"), router.Group("/api/public"), app, cfg)

			url := "/api/v1/empregabilidade/vagas/" + tt.vaga.String() + "/etapas/" + tt.etapa.String()
			var req *http.Request
			if tt.method == http.MethodPut {
				req = httptest.NewRequest(http.MethodPut, url, strings.NewReader(`{"titulo":"Alterada","ordem":9}`))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(http.MethodDelete, url, nil)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.want, w.Code, w.Body.String())

			_, criada := etapaRepo.etapas[inexistente]
			assert.False(t, criada, "o PUT não pode criar etapa inexistente")

			b, ok := etapaRepo.etapas[etapaB]
			require.True(t, ok, "etapa da vaga B não pode ser excluída")
			assert.Equal(t, vagaB, b.IDVaga, "etapa da vaga B não pode ser movida")
			assert.Equal(t, "Triagem B", b.Titulo, "etapa da vaga B não pode ser alterada")
			assert.Equal(t, 1, b.Ordem, "etapa da vaga B não pode ser alterada")

			a, ok := etapaRepo.etapas[etapaA]
			switch tt.wantEtapaA {
			case "excluida":
				assert.False(t, ok, "etapa da vaga A deveria ter sido excluída")
			case "alterada":
				require.True(t, ok)
				assert.Equal(t, vagaA, a.IDVaga)
				assert.Equal(t, "Alterada", a.Titulo)
			default:
				require.True(t, ok)
				assert.Equal(t, "Triagem A", a.Titulo)
			}
		})
	}
}
