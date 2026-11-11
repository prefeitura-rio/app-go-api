package empregabilidade_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// DEV-518: a listagem de bloqueios de candidatura (com CPF do cidadão) aplica o
// escopo de órgão parceiro que VagaListFilter injeta no contexto, como a
// listagem de vagas. Admin total e go:empregabilidade:admin não recebem escopo
// (sem restrição); editor sem escopo não vê nada.

type bloqueioScopeCaller struct {
	role          string
	roles         []string
	secretariaIDs []string
	orgaoID       string
}

func (bc bloqueioScopeCaller) inject() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middlewares.UserCPFKey, "11111111111")
		if bc.role != "" {
			c.Set(middlewares.UserRoleKey, bc.role)
		}
		if bc.roles != nil {
			c.Set(middlewares.UserRolesKey, bc.roles)
		}
		if bc.secretariaIDs != nil {
			c.Set(middlewares.UserSecretariaOrgaoIDsKey, bc.secretariaIDs)
		}
		if bc.orgaoID != "" {
			c.Set(middlewares.UserGroupsKey, []string{"go:orgao:" + bc.orgaoID})
		}
		c.Next()
	}
}

func newBloqueioScopeRouter(repo *mockCandidaturaBloqueioRepoH, caller bloqueioScopeCaller, comEscopo bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handlers.NewCandidaturaBloqueioHandler(
		services.NewCandidaturaBloqueioServiceWithInterface(repo, vagaExistenteBloqueio()),
	)
	r := gin.New()
	r.Use(caller.inject())
	if comEscopo {
		r.GET("/candidatura-bloqueios", middlewares.VagaListFilter(), h.List)
	} else {
		r.GET("/candidatura-bloqueios", h.List)
	}
	return r
}

func TestCandidaturaBloqueioHandler_List_AplicaEscopoDeOrgao(t *testing.T) {
	tests := []struct {
		name   string
		caller bloqueioScopeCaller
		query  string
		want   []string
	}{
		{"editor com secretaria fica restrito aos órgãos dela", bloqueioScopeCaller{role: "USER", roles: []string{"go:empregabilidade:editor_com_curadoria"}, secretariaIDs: []string{"1", "2600"}}, "", []string{"1", "2600"}},
		{"editor com órgão (go:orgao) fica restrito a ele", bloqueioScopeCaller{role: "USER", roles: []string{"go:empregabilidade:editor"}, orgaoID: "2600"}, "", []string{"2600"}},
		{"filtro de cpf e vaga não amplia o escopo do editor", bloqueioScopeCaller{role: "USER", roles: []string{"go:empregabilidade:editor"}, orgaoID: "1"}, "?cpf=99999999999&id_vaga=" + validUUID, []string{"1"}},
		{"editor sem escopo não vê nada", bloqueioScopeCaller{role: "USER", roles: []string{"go:empregabilidade:editor"}}, "", []string{}},
		{"go:empregabilidade:admin não é restrito", bloqueioScopeCaller{role: "USER", roles: []string{"go:empregabilidade:admin"}}, "", nil},
		{"admin total não é restrito", bloqueioScopeCaller{role: "ADMIN", roles: []string{"go:admin"}}, "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCandidaturaBloqueioRepoH{}
			r := newBloqueioScopeRouter(repo, tt.caller, true)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/candidatura-bloqueios"+tt.query, nil))

			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			require.Equal(t, 1, repo.listCalls)
			assert.Equal(t, tt.want, repo.listOrgaoIDs)
		})
	}
}

func TestCandidaturaBloqueioHandler_List_SemEscopoInjetado_NaoRestringe(t *testing.T) {
	// RBAC_ENABLED=false: VagaListFilter não roda e nada é injetado.
	repo := &mockCandidaturaBloqueioRepoH{}
	r := newBloqueioScopeRouter(repo, bloqueioScopeCaller{role: "USER", roles: []string{"go:empregabilidade:editor"}, orgaoID: "1"}, false)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/candidatura-bloqueios", nil))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Equal(t, 1, repo.listCalls)
	assert.Nil(t, repo.listOrgaoIDs)
}
