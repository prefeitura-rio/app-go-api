package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	v1 "github.com/prefeitura-rio/app-go-api/internal/handlers/v1"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models"
)

const (
	meiDraftID     = 83
	meiActiveID    = 10
	meiPropostaCNP = "11222333000181"
)

// oportunidadeMEISvcStub atende as rotas de oportunidade MEI e conta as
// chamadas, para provar que um chamador barrado nunca chega ao serviço. O id
// meiDraftID é um rascunho e meiActiveID uma oportunidade publicada.
type oportunidadeMEISvcStub struct {
	calls int
}

func (s *oportunidadeMEISvcStub) Create(_ context.Context, _ *models.OportunidadeMEI, _ bool) (int, error) {
	s.calls++
	return 1, nil
}

func (s *oportunidadeMEISvcStub) GetByID(_ context.Context, id int) (*models.OportunidadeMEI, error) {
	s.calls++
	switch id {
	case meiDraftID:
		return &models.OportunidadeMEI{ID: id, Titulo: "Rascunho", Status: models.StatusOportunidadeDraft}, nil
	case meiActiveID:
		return &models.OportunidadeMEI{ID: id, Titulo: "Publicada", Status: models.StatusOportunidadeActive}, nil
	}
	return nil, nil
}

func (s *oportunidadeMEISvcStub) Update(_ context.Context, _ *models.OportunidadeMEI) error {
	s.calls++
	return nil
}

func (s *oportunidadeMEISvcStub) Publish(_ context.Context, _ int) error {
	s.calls++
	return nil
}

func (s *oportunidadeMEISvcStub) Delete(_ context.Context, _ int) error {
	s.calls++
	return nil
}

func (s *oportunidadeMEISvcStub) List(_ context.Context, _ map[string]interface{}, _ string, _, _ int) ([]*models.OportunidadeMEI, int, error) {
	s.calls++
	return []*models.OportunidadeMEI{}, 0, nil
}

func (s *oportunidadeMEISvcStub) ListByStatus(_ context.Context, _ models.StatusOportunidadeMEI, _, _ int) ([]*models.OportunidadeMEI, int, error) {
	s.calls++
	return []*models.OportunidadeMEI{}, 0, nil
}

func (s *oportunidadeMEISvcStub) ListDrafts(_ context.Context, _, _ string, _, _ int) ([]*models.OportunidadeMEI, int, error) {
	s.calls++
	return []*models.OportunidadeMEI{}, 0, nil
}

func (s *oportunidadeMEISvcStub) ListActive(_ context.Context, _, _ int) ([]*models.OportunidadeMEI, int, error) {
	s.calls++
	return []*models.OportunidadeMEI{}, 0, nil
}

func (s *oportunidadeMEISvcStub) ListByOrgao(_ context.Context, _ string, _, _ int) ([]*models.OportunidadeMEI, int, error) {
	s.calls++
	return []*models.OportunidadeMEI{}, 0, nil
}

func (s *oportunidadeMEISvcStub) UpdateExpiredOpportunities(_ context.Context) error {
	s.calls++
	return nil
}

// propostaMEISvcStub atende as rotas de proposta MEI e conta as chamadas.
type propostaMEISvcStub struct {
	calls int
}

func (s *propostaMEISvcStub) Create(_ context.Context, _ *models.PropostaMEI, _ string) (uuid.UUID, error) {
	s.calls++
	return uuid.New(), nil
}

func (s *propostaMEISvcStub) GetByID(_ context.Context, id uuid.UUID) (*models.PropostaMEI, error) {
	s.calls++
	return &models.PropostaMEI{ID: id, MEIEmpresaID: meiPropostaCNP}, nil
}

func (s *propostaMEISvcStub) Update(_ context.Context, _ *models.PropostaMEI) error {
	s.calls++
	return nil
}

func (s *propostaMEISvcStub) UpdateProposta(_ context.Context, _ uuid.UUID, _ int, _ *float64, _ *string, _ *bool) error {
	s.calls++
	return nil
}

func (s *propostaMEISvcStub) UpdateStatusCidadao(_ context.Context, _ uuid.UUID, _ models.StatusPropostaCidadao) error {
	s.calls++
	return nil
}

func (s *propostaMEISvcStub) Approve(_ context.Context, _ uuid.UUID) error {
	s.calls++
	return nil
}

func (s *propostaMEISvcStub) Reject(_ context.Context, _ uuid.UUID) error {
	s.calls++
	return nil
}

func (s *propostaMEISvcStub) Delete(_ context.Context, _ uuid.UUID) error {
	s.calls++
	return nil
}

func (s *propostaMEISvcStub) ListByOportunidade(_ context.Context, _ int, _, _, _ string, _, _ int) ([]*models.PropostaMEI, int, error) {
	s.calls++
	return []*models.PropostaMEI{}, 0, nil
}

func (s *propostaMEISvcStub) ListByMEIEmpresa(_ context.Context, _ string, _, _ int) ([]*models.PropostaMEI, int, error) {
	s.calls++
	return []*models.PropostaMEI{}, 0, nil
}

func (s *propostaMEISvcStub) ListByStatus(_ context.Context, _ models.StatusPropostaCidadao, _, _ int) ([]*models.PropostaMEI, int, error) {
	s.calls++
	return []*models.PropostaMEI{}, 0, nil
}

func (s *propostaMEISvcStub) UpdateMultipleStatus(_ context.Context, ids []uuid.UUID, _ models.StatusPropostaCidadao) (int, error) {
	s.calls++
	return len(ids), nil
}

// cnpjOwnerStub diz que o chamador é dono só do CNPJ owned.
type cnpjOwnerStub struct {
	owned string
}

func (s *cnpjOwnerStub) ValidatePropostaForCNAE(_ context.Context, _, _ string, _ []string) error {
	return nil
}

func (s *cnpjOwnerStub) CheckCNPJOwnership(_ context.Context, _, _, cnpj string) (bool, error) {
	return cnpj == s.owned, nil
}

type meiTestUser struct {
	cpf   string
	role  string
	roles []string
}

var (
	meiAnon         = meiTestUser{}
	meiCidadao      = meiTestUser{cpf: "00000000191", role: "USER"}
	meiCursosEditor = meiTestUser{cpf: "00000000272", role: "USER", roles: []string{"go:cursos:editor"}}
	meiEmpEditor    = meiTestUser{cpf: "00000000353", role: "USER", roles: []string{"go:empregabilidade:editor"}}
	meiEmpAdmin     = meiTestUser{cpf: "00000000434", role: "USER", roles: []string{"go:empregabilidade:admin"}}
	meiAdmin        = meiTestUser{cpf: "00000000515", role: "ADMIN", roles: []string{"go:admin"}}
)

// newMEIRouter monta as rotas core reais com os stubs MEI e um contexto de
// usuário fixo, com RBAC desligado (como em produção hoje): a proteção das
// rotas MEI não pode depender de RBAC_ENABLED.
func newMEIRouter(user meiTestUser, ownedCNPJ string) (*gin.Engine, *oportunidadeMEISvcStub, *propostaMEISvcStub) {
	router := gin.New()
	apiV1 := router.Group("/api/v1")
	apiV1.Use(func(c *gin.Context) {
		if user.cpf != "" {
			c.Set(middlewares.UserCPFKey, user.cpf)
		}
		if user.role != "" {
			c.Set(middlewares.UserRoleKey, user.role)
		}
		if user.roles != nil {
			c.Set(middlewares.UserRolesKey, user.roles)
		}
		c.Next()
	})
	apiPublic := router.Group("/api/public")

	opSvc := &oportunidadeMEISvcStub{}
	propSvc := &propostaMEISvcStub{}
	cfg := &config.AppConfig{
		App:         config.AppSettings{RBACEnabled: false},
		PropostaMEI: config.PropostaMEIPermissions{ReadPermissions: []string{"read"}},
	}

	app := createMockApplicationContainer()
	app.OportunidadeMEIHandler = v1.NewOportunidadeMEIHandler(opSvc)
	app.PropostaMEIHandler = v1.NewPropostaMEIHandler(propSvc, &cnpjOwnerStub{owned: ownedCNPJ}, nil, cfg)
	registerCoreRoutes(apiV1, apiPublic, app, cfg)

	return router, opSvc, propSvc
}

func doMEIRequest(router *gin.Engine, method, path, body string, withToken bool) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if withToken {
		req.Header.Set("Authorization", "Bearer token-sintetico")
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

type meiRoute struct {
	name, method, path, body string
	okStatus                 int
}

func meiGestaoRoutes() []meiRoute {
	propostaID := uuid.New().String()
	bulkBody := `{"proposta_ids":["` + uuid.New().String() + `"],"status":"approved"}`
	return []meiRoute{
		{"criar oportunidade", http.MethodPost, "/api/v1/oportunidades-mei", `{"titulo":"X"}`, http.StatusCreated},
		{"criar rascunho", http.MethodPost, "/api/v1/oportunidades-mei/draft", `{"titulo":"X"}`, http.StatusCreated},
		{"listar rascunhos", http.MethodGet, "/api/v1/oportunidades-mei/drafts", "", http.StatusOK},
		{"editar oportunidade", http.MethodPut, "/api/v1/oportunidades-mei/10", `{"titulo":"X"}`, http.StatusOK},
		{"publicar oportunidade", http.MethodPut, "/api/v1/oportunidades-mei/83/publish", "", http.StatusOK},
		{"excluir oportunidade", http.MethodDelete, "/api/v1/oportunidades-mei/10", "", http.StatusOK},
		{"status de propostas em lote", http.MethodPut, "/api/v1/oportunidades-mei/10/propostas/status", bulkBody, http.StatusOK},
		{"status de uma proposta", http.MethodPut, "/api/v1/oportunidades-mei/10/propostas/" + propostaID + "/status", `{"status":"approved"}`, http.StatusOK},
	}
}

// TestMEIRoutes_Gestao_DeniesWithoutMEIProfile prova que as mutações de
// oportunidade, os rascunhos e a mudança de status das propostas recusam
// anônimo, cidadão e perfil de outro módulo antes de chegar ao serviço.
func TestMEIRoutes_Gestao_DeniesWithoutMEIProfile(t *testing.T) {
	for _, route := range meiGestaoRoutes() {
		for _, user := range []struct {
			name      string
			u         meiTestUser
			withToken bool
		}{
			{"anônimo sem token", meiAnon, false},
			{"cidadão", meiCidadao, true},
			{"editor de cursos", meiCursosEditor, true},
		} {
			t.Run(route.name+"/"+user.name, func(t *testing.T) {
				router, opSvc, propSvc := newMEIRouter(user.u, "")
				w := doMEIRequest(router, route.method, route.path, route.body, user.withToken)

				assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
				assert.Zero(t, opSvc.calls, "serviço de oportunidade não pode ser chamado")
				assert.Zero(t, propSvc.calls, "serviço de proposta não pode ser chamado")
			})
		}
	}
}

// TestMEIRoutes_Gestao_AllowsMEIProfiles prova que os perfis de gestão MEI
// seguem operando as mesmas rotas.
func TestMEIRoutes_Gestao_AllowsMEIProfiles(t *testing.T) {
	for _, route := range meiGestaoRoutes() {
		for _, user := range []struct {
			name string
			u    meiTestUser
		}{
			{"editor de empregabilidade", meiEmpEditor},
			{"admin de empregabilidade", meiEmpAdmin},
			{"ADMIN", meiAdmin},
		} {
			t.Run(route.name+"/"+user.name, func(t *testing.T) {
				router, _, _ := newMEIRouter(user.u, "")
				w := doMEIRequest(router, route.method, route.path, route.body, true)

				assert.Equal(t, route.okStatus, w.Code, w.Body.String())
			})
		}
	}
}

// TestMEIRoutes_GetByID_DraftHiddenWithoutMEIProfile prova que um rascunho não
// é exposto pelo GET por id (v1 e público) a quem não gere MEI, enquanto a
// oportunidade publicada segue pública.
func TestMEIRoutes_GetByID_DraftHiddenWithoutMEIProfile(t *testing.T) {
	for _, path := range []string{"/api/v1/oportunidades-mei/83", "/api/public/oportunidades-mei/83"} {
		for _, user := range []struct {
			name string
			u    meiTestUser
		}{
			{"anônimo", meiAnon},
			{"cidadão", meiCidadao},
			{"editor de cursos", meiCursosEditor},
		} {
			t.Run(path+"/"+user.name, func(t *testing.T) {
				router, _, _ := newMEIRouter(user.u, "")
				w := doMEIRequest(router, http.MethodGet, path, "", user.u.cpf != "")

				assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
				assert.NotContains(t, w.Body.String(), "Rascunho")
			})
		}
	}

	t.Run("editor de empregabilidade vê o rascunho no v1", func(t *testing.T) {
		router, _, _ := newMEIRouter(meiEmpEditor, "")
		w := doMEIRequest(router, http.MethodGet, "/api/v1/oportunidades-mei/83", "", true)

		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, w.Body.String(), "Rascunho")
	})

	for _, path := range []string{"/api/v1/oportunidades-mei/10", "/api/public/oportunidades-mei/10"} {
		t.Run("publicada segue aberta "+path, func(t *testing.T) {
			router, _, _ := newMEIRouter(meiAnon, "")
			w := doMEIRequest(router, http.MethodGet, path, "", false)

			assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), "Publicada")
		})
	}
}

// TestMEIRoutes_List_DraftStatusRequiresMEIProfile prova que a listagem não
// devolve rascunhos via ?status=draft a quem não gere MEI, e que a listagem
// das publicadas segue aberta.
func TestMEIRoutes_List_DraftStatusRequiresMEIProfile(t *testing.T) {
	for _, path := range []string{"/api/v1/oportunidades-mei?status=draft", "/api/public/oportunidades-mei?status=draft"} {
		t.Run("anônimo "+path, func(t *testing.T) {
			router, opSvc, _ := newMEIRouter(meiAnon, "")
			w := doMEIRequest(router, http.MethodGet, path, "", false)

			assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
			assert.Zero(t, opSvc.calls)
		})
	}

	t.Run("editor de cursos no v1", func(t *testing.T) {
		router, opSvc, _ := newMEIRouter(meiCursosEditor, "")
		w := doMEIRequest(router, http.MethodGet, "/api/v1/oportunidades-mei?status=draft", "", true)

		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		assert.Zero(t, opSvc.calls)
	})

	t.Run("editor de empregabilidade no v1", func(t *testing.T) {
		router, _, _ := newMEIRouter(meiEmpEditor, "")
		w := doMEIRequest(router, http.MethodGet, "/api/v1/oportunidades-mei?status=draft", "", true)

		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	for _, path := range []string{"/api/v1/oportunidades-mei", "/api/public/oportunidades-mei", "/api/public/oportunidades-mei?status=active"} {
		t.Run("publicadas seguem abertas "+path, func(t *testing.T) {
			router, _, _ := newMEIRouter(meiAnon, "")
			w := doMEIRequest(router, http.MethodGet, path, "", false)

			assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
		})
	}
}

// TestMEIRoutes_ListPropostas_RequiresMEIProfileOrReadPermission prova que a
// listagem das propostas de uma oportunidade (CNPJ e valores das empresas)
// exige perfil de gestão MEI; sem Cerbos, ninguém mais lista.
func TestMEIRoutes_ListPropostas_RequiresMEIProfileOrReadPermission(t *testing.T) {
	for _, user := range []struct {
		name      string
		u         meiTestUser
		withToken bool
		want      int
	}{
		{"anônimo", meiAnon, false, http.StatusForbidden},
		{"cidadão", meiCidadao, true, http.StatusForbidden},
		{"editor de cursos", meiCursosEditor, true, http.StatusForbidden},
		{"editor de empregabilidade", meiEmpEditor, true, http.StatusOK},
		{"ADMIN", meiAdmin, true, http.StatusOK},
	} {
		t.Run(user.name, func(t *testing.T) {
			router, _, propSvc := newMEIRouter(user.u, meiPropostaCNP)
			w := doMEIRequest(router, http.MethodGet, "/api/v1/oportunidades-mei/10/propostas", "", user.withToken)

			assert.Equal(t, user.want, w.Code, w.Body.String())
			if user.want != http.StatusOK {
				assert.Zero(t, propSvc.calls)
			}
		})
	}
}

// TestMEIRoutes_ListByMEIEmpresa_RequiresOwnershipOrMEIProfile prova que as
// propostas de uma empresa só são listadas pelo dono do CNPJ ou por um perfil
// de gestão MEI.
func TestMEIRoutes_ListByMEIEmpresa_RequiresOwnershipOrMEIProfile(t *testing.T) {
	path := "/api/v1/propostas-mei/por-empresa?meiEmpresaId=" + meiPropostaCNP
	for _, tc := range []struct {
		name      string
		u         meiTestUser
		owned     string
		withToken bool
		want      int
	}{
		{"anônimo", meiAnon, "", false, http.StatusForbidden},
		{"cidadão que não é dono", meiCidadao, "99888777000166", true, http.StatusForbidden},
		{"editor de cursos que não é dono", meiCursosEditor, "", true, http.StatusForbidden},
		{"cidadão dono do CNPJ", meiCidadao, meiPropostaCNP, true, http.StatusOK},
		{"editor de empregabilidade", meiEmpEditor, "", true, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router, _, propSvc := newMEIRouter(tc.u, tc.owned)
			w := doMEIRequest(router, http.MethodGet, path, "", tc.withToken)

			assert.Equal(t, tc.want, w.Code, w.Body.String())
			if tc.want != http.StatusOK {
				assert.Zero(t, propSvc.calls)
			}
		})
	}
}
