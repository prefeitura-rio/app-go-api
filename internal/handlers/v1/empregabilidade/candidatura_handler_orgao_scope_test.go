package empregabilidade_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	empRepository "github.com/prefeitura-rio/app-go-api/internal/repository/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services"
	empServices "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// DEV-517: as rotas de gestão de candidaturas aplicam o escopo de órgão
// parceiro que VagaListFilter injeta no contexto, como a listagem de vagas.
// Admin total e go:empregabilidade:admin não recebem escopo (sem restrição).

// scopeCandidaturaRepo grava os filtros recebidos e conta as escritas.
type scopeCandidaturaRepo struct {
	mockCandidaturaRepoH
	listFilters  []empmodels.CandidaturaFilter
	countFilters []empmodels.CandidaturaFilter
	writes       int
}

func (m *scopeCandidaturaRepo) List(_ context.Context, f empmodels.CandidaturaFilter, _, _ int) ([]*empmodels.Candidatura, int, error) {
	m.listFilters = append(m.listFilters, f)
	return []*empmodels.Candidatura{}, 0, nil
}

func (m *scopeCandidaturaRepo) CountByStatus(_ context.Context, f empmodels.CandidaturaFilter) (map[empmodels.StatusCandidatura]int64, error) {
	m.countFilters = append(m.countFilters, f)
	return map[empmodels.StatusCandidatura]int64{}, nil
}

func (m *scopeCandidaturaRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ empmodels.StatusCandidatura) error {
	m.writes++
	return nil
}

func (m *scopeCandidaturaRepo) UpdateEtapa(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	m.writes++
	return nil
}

// bulkScopeRepo conta as escritas em lote.
type bulkScopeRepo struct {
	scopeCandidaturaRepo
	bulkWrites int
}

func (m *bulkScopeRepo) BulkUpdateStatus(_ context.Context, _ uuid.UUID, cpfs []string, _ empmodels.StatusCandidatura) (empRepository.BulkUpdateResult, error) {
	m.bulkWrites++
	return empRepository.BulkUpdateResult{Updated: len(cpfs)}, nil
}

func (m *bulkScopeRepo) BulkUpdateEtapa(_ context.Context, _ []uuid.UUID, _ uuid.UUID) error {
	m.bulkWrites++
	return nil
}

type scopeCaller struct {
	cpf           string
	role          string
	roles         []string
	secretariaIDs []string
	orgaoID       string
}

func (sc scopeCaller) inject() gin.HandlerFunc {
	return func(c *gin.Context) {
		if sc.cpf != "" {
			c.Set(middlewares.UserCPFKey, sc.cpf)
		}
		if sc.role != "" {
			c.Set(middlewares.UserRoleKey, sc.role)
		}
		if sc.roles != nil {
			c.Set(middlewares.UserRolesKey, sc.roles)
		}
		if sc.secretariaIDs != nil {
			c.Set(middlewares.UserSecretariaOrgaoIDsKey, sc.secretariaIDs)
		}
		if sc.orgaoID != "" {
			c.Set(middlewares.UserGroupsKey, []string{"go:orgao:" + sc.orgaoID})
		}
		c.Next()
	}
}

var (
	editorOrgao1    = scopeCaller{cpf: "11111111111", role: "USER", roles: []string{"go:empregabilidade:editor_com_curadoria"}, secretariaIDs: []string{"1"}}
	editorSemEscopo = scopeCaller{cpf: "11111111111", role: "USER", roles: []string{"go:empregabilidade:editor"}}
	adminEmp        = scopeCaller{cpf: "22222222222", role: "USER", roles: []string{"go:empregabilidade:admin"}}
	adminTotal      = scopeCaller{cpf: "33333333333", role: "ADMIN", roles: []string{"go:admin"}}
	cidadao         = scopeCaller{cpf: "44444444444", role: "USER"}
)

func newScopeRouter(t *testing.T, repo empServices.CandidaturaRepositoryInterface, vagaRepo empServices.VagaRepositoryInterface, caller scopeCaller) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	emailSvc := services.NewEmailNotificationService(nil, nil, nil, nil, false, "")
	svc := empServices.NewCandidaturaService(repo, vagaRepo, &mockCurriculoSvcH{}, nil, nil, emailSvc)
	h := handlers.NewCandidaturaHandler(svc)

	r := gin.New()
	r.Use(caller.inject())
	scope := middlewares.VagaListFilter()
	r.GET("/candidaturas", scope, h.List)
	r.GET("/candidaturas/usuario/:cpf", scope, h.ListByCPF)
	r.PUT("/candidaturas/bulk-status", scope, h.BulkUpdateStatus)
	r.PUT("/candidaturas/bulk-etapa", scope, h.BulkUpdateEtapa)
	r.PUT("/candidaturas/:id/status", scope, h.UpdateStatus)
	r.PUT("/candidaturas/:id/approve", scope, h.Approve)
	r.PUT("/candidaturas/:id/reject", scope, h.Reject)
	r.PUT("/candidaturas/:id/etapa", scope, h.UpdateEtapa)
	return r
}

func doRequest(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCandidaturaHandler_List_AplicaEscopoDeOrgao(t *testing.T) {
	tests := []struct {
		name   string
		caller scopeCaller
		want   []string
	}{
		{"editor com secretaria fica restrito ao órgão dela", editorOrgao1, []string{"1"}},
		{"editor com órgão (go:orgao) fica restrito a ele", scopeCaller{role: "USER", roles: []string{"go:empregabilidade:editor"}, orgaoID: "2600"}, []string{"2600"}},
		{"editor sem escopo não vê nada", editorSemEscopo, []string{}},
		{"go:empregabilidade:admin não é restrito", adminEmp, nil},
		{"admin total não é restrito", adminTotal, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &scopeCandidaturaRepo{}
			r := newScopeRouter(t, repo, &mockVagaForCandidaturaH{}, tt.caller)

			w := doRequest(r, http.MethodGet, "/candidaturas?vagaId=55555555-5555-4555-8555-555555555555", "")

			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			require.Len(t, repo.listFilters, 1)
			require.Len(t, repo.countFilters, 1)
			assert.Equal(t, tt.want, repo.listFilters[0].OrgaoParceiroIDs)
			assert.Equal(t, tt.want, repo.countFilters[0].OrgaoParceiroIDs, "o resumo por status segue o mesmo escopo")
		})
	}
}

func TestCandidaturaHandler_List_SemEscopoInjetado_NaoRestringe(t *testing.T) {
	// RBAC_ENABLED=false: VagaListFilter não roda e nada é injetado.
	repo := &scopeCandidaturaRepo{}
	r := setupCandidaturaRouterWithRoles(repo, &mockVagaForCandidaturaH{}, "11111111111", []string{"go:empregabilidade:editor"})

	w := doRequest(r, http.MethodGet, "/candidaturas", "")

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Len(t, repo.listFilters, 1)
	assert.Nil(t, repo.listFilters[0].OrgaoParceiroIDs)
}

func TestCandidaturaHandler_ListByCPF_AplicaEscopoDeOrgao(t *testing.T) {
	const outroCPF = "99999999999"
	tests := []struct {
		name   string
		caller scopeCaller
		cpf    string
		want   []string
	}{
		{"editor consultando outro CPF fica restrito ao seu órgão", editorOrgao1, outroCPF, []string{"1"}},
		{"editor sem escopo consultando outro CPF não vê nada", editorSemEscopo, outroCPF, []string{}},
		{"editor consultando o próprio CPF vê todas as suas candidaturas", editorOrgao1, editorOrgao1.cpf, nil},
		{"cidadão consultando o próprio CPF vê todas as suas candidaturas", cidadao, cidadao.cpf, nil},
		{"go:empregabilidade:admin não é restrito", adminEmp, outroCPF, nil},
		{"admin total não é restrito", adminTotal, outroCPF, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &scopeCandidaturaRepo{}
			r := newScopeRouter(t, repo, &mockVagaForCandidaturaH{}, tt.caller)

			w := doRequest(r, http.MethodGet, "/candidaturas/usuario/"+tt.cpf, "")

			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			require.Len(t, repo.listFilters, 1)
			require.Len(t, repo.countFilters, 1)
			assert.Equal(t, tt.cpf, repo.listFilters[0].CPF)
			assert.Equal(t, tt.want, repo.listFilters[0].OrgaoParceiroIDs)
			assert.Equal(t, tt.want, repo.countFilters[0].OrgaoParceiroIDs)
		})
	}
}

func TestCandidaturaHandler_ListByCPF_CidadaoOutroCPF_ContinuaForbidden(t *testing.T) {
	repo := &scopeCandidaturaRepo{}
	r := newScopeRouter(t, repo, &mockVagaForCandidaturaH{}, cidadao)

	w := doRequest(r, http.MethodGet, "/candidaturas/usuario/99999999999", "")

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, repo.listFilters)
}

func candidaturaDaVaga(orgaoID string) *empmodels.Candidatura {
	vagaID := uuid.New()
	etapaID := uuid.New()
	return &empmodels.Candidatura{
		ID:           uuid.New(),
		CPF:          "99999999999",
		IDVaga:       vagaID,
		Status:       empmodels.StatusCandidaturaEnviada,
		IDEtapaAtual: &etapaID,
		Vaga: &empmodels.Vaga{
			ID:              vagaID,
			IDOrgaoParceiro: orgaoID,
			Etapas:          []empmodels.Etapa{{ID: etapaID, IDVaga: vagaID, Ordem: 1}},
		},
	}
}

func TestCandidaturaHandler_EscritaPorID_AplicaEscopoDeOrgao(t *testing.T) {
	etapa := uuid.New().String()
	routes := []struct {
		name string
		path string
		body string
	}{
		{"status", "/status", `{"status":"aprovada"}`},
		{"approve", "/approve", ""},
		{"reject", "/reject", ""},
		{"etapa", "/etapa", `{"id_etapa_atual":"` + etapa + `"}`},
	}
	callers := []struct {
		name      string
		caller    scopeCaller
		orgaoVaga string
		forbidden bool
	}{
		{"editor de outro órgão", editorOrgao1, "1100", true},
		{"editor sem escopo", editorSemEscopo, "1", true},
		{"editor do órgão da vaga", editorOrgao1, "1", false},
		{"go:empregabilidade:admin", adminEmp, "1100", false},
		{"admin total", adminTotal, "1100", false},
	}

	for _, rt := range routes {
		for _, tt := range callers {
			t.Run(rt.name+" "+tt.name, func(t *testing.T) {
				cand := candidaturaDaVaga(tt.orgaoVaga)
				repo := &scopeCandidaturaRepo{mockCandidaturaRepoH: mockCandidaturaRepoH{entity: cand}}
				r := newScopeRouter(t, repo, &mockVagaForCandidaturaH{entity: cand.Vaga}, tt.caller)

				w := doRequest(r, http.MethodPut, "/candidaturas/"+cand.ID.String()+rt.path, rt.body)

				if tt.forbidden {
					assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
					assert.Equal(t, 0, repo.writes, "candidatura fora do escopo não pode ser alterada")
				} else {
					assert.NotEqual(t, http.StatusForbidden, w.Code, w.Body.String())
				}
			})
		}
	}
}

func TestCandidaturaHandler_EscritaPorID_CandidaturaInexistente_404(t *testing.T) {
	repo := &scopeCandidaturaRepo{}
	r := newScopeRouter(t, repo, &mockVagaForCandidaturaH{}, editorOrgao1)

	w := doRequest(r, http.MethodPut, "/candidaturas/"+uuid.New().String()+"/approve", "")

	assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	assert.Equal(t, 0, repo.writes)
}

func TestCandidaturaHandler_EscritaEmLote_AplicaEscopoDeOrgao(t *testing.T) {
	routes := []struct {
		name string
		path string
		body func(vagaID, etapaID string) string
	}{
		{"bulk-status", "/candidaturas/bulk-status", func(v, _ string) string {
			return `{"cpfs":["99999999999"],"vaga_id":"` + v + `","status":"aprovada"}`
		}},
		{"bulk-etapa", "/candidaturas/bulk-etapa", func(v, e string) string {
			return `{"cpfs":["99999999999"],"vaga_id":"` + v + `","id_etapa":"` + e + `"}`
		}},
	}
	callers := []struct {
		name      string
		caller    scopeCaller
		orgaoVaga string
		forbidden bool
	}{
		{"editor de outro órgão", editorOrgao1, "1100", true},
		{"editor sem escopo", editorSemEscopo, "1", true},
		{"editor do órgão da vaga", editorOrgao1, "1", false},
		{"go:empregabilidade:admin", adminEmp, "1100", false},
		{"admin total", adminTotal, "1100", false},
	}

	for _, rt := range routes {
		for _, tt := range callers {
			t.Run(rt.name+" "+tt.name, func(t *testing.T) {
				cand := candidaturaDaVaga(tt.orgaoVaga)
				repo := &bulkScopeRepo{scopeCandidaturaRepo: scopeCandidaturaRepo{mockCandidaturaRepoH: mockCandidaturaRepoH{
					entity:    cand,
					listItems: []*empmodels.Candidatura{cand},
				}}}
				r := newScopeRouter(t, repo, &mockVagaForCandidaturaH{entity: cand.Vaga}, tt.caller)

				w := doRequest(r, http.MethodPut, rt.path, rt.body(cand.IDVaga.String(), cand.Vaga.Etapas[0].ID.String()))

				if tt.forbidden {
					assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
					assert.Equal(t, 0, repo.bulkWrites, "vaga fora do escopo não pode ter candidaturas alteradas")
				} else {
					assert.NotEqual(t, http.StatusForbidden, w.Code, w.Body.String())
				}
			})
		}
	}
}

func TestCandidaturaHandler_EscritaEmLote_VagaInexistente_404(t *testing.T) {
	repo := &bulkScopeRepo{}
	r := newScopeRouter(t, repo, &mockVagaForCandidaturaH{}, editorOrgao1)

	w := doRequest(r, http.MethodPut, "/candidaturas/bulk-status",
		`{"cpfs":["99999999999"],"vaga_id":"`+uuid.New().String()+`","status":"aprovada"}`)

	assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	assert.Equal(t, 0, repo.bulkWrites)
}
