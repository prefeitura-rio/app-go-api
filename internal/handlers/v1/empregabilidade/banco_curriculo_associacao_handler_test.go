package empregabilidade_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAssociacaoService struct {
	resultados []*empmodels.AssociacaoVaga
	err        error
	chamado    bool
	params     services.AssociarParams
}

func (f *fakeAssociacaoService) Associar(_ context.Context, params services.AssociarParams) ([]*empmodels.AssociacaoVaga, error) {
	f.chamado = true
	f.params = params
	return f.resultados, f.err
}

// newAssociacaoRouter simula o operador logado e o escopo que a autorização
// injeta antes do handler.
func newAssociacaoRouter(svc *fakeAssociacaoService, orgaos []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middlewares.UserCPFKey, "22222222222")
		c.Set(middlewares.UserRolesKey, []string{"go:empregabilidade:editor_sem_curadoria"})
		c.Set(middlewares.UserSecretariaOrgaoIDsKey, orgaos)
		c.Next()
	})
	r.Use(middlewares.BancoCurriculosAssociacaoAuthorization())
	h := handlers.NewAssociacaoCandidatoHandler(svc)
	r.POST("/banco-curriculos/:cpf/associacoes", h.Associar)
	return r
}

func doAssociar(r *gin.Engine, cpf, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/banco-curriculos/"+cpf+"/associacoes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestAssociacaoCandidatoHandler_Associa(t *testing.T) {
	idVaga := uuid.New()
	idCandidatura := uuid.New()
	svc := &fakeAssociacaoService{resultados: []*empmodels.AssociacaoVaga{
		{IDVaga: idVaga, Titulo: "Auxiliar", Resultado: empmodels.AssociacaoRealizada, IDCandidatura: &idCandidatura},
	}}

	w := doAssociar(newAssociacaoRouter(svc, []string{"2600"}), "111.111.111-11", `{"ids_vagas":["`+idVaga.String()+`"]}`)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "11111111111", svc.params.CPF)
	assert.Equal(t, []uuid.UUID{idVaga}, svc.params.IDsVagas)
	assert.Equal(t, "22222222222", svc.params.OperadorCPF)
	assert.Equal(t, []string{"2600"}, svc.params.OrgaosPermitidos)

	var body empmodels.AssociarVagasResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Resultados, 1)
	assert.Equal(t, empmodels.AssociacaoRealizada, body.Resultados[0].Resultado)
}

func TestAssociacaoCandidatoHandler_PedidoInvalido(t *testing.T) {
	cases := map[string]struct{ cpf, body string }{
		"CPF inválido":      {"123", `{"ids_vagas":["` + uuid.NewString() + `"]}`},
		"sem vagas":         {"11111111111", `{"ids_vagas":[]}`},
		"id que não é uuid": {"11111111111", `{"ids_vagas":["abc"]}`},
		"corpo vazio":       {"11111111111", ``},
	}
	for nome, tc := range cases {
		t.Run(nome, func(t *testing.T) {
			svc := &fakeAssociacaoService{}
			w := doAssociar(newAssociacaoRouter(svc, []string{"2600"}), tc.cpf, tc.body)
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.False(t, svc.chamado)
		})
	}
}

func TestAssociacaoCandidatoHandler_CurriculoForaDoMinimo(t *testing.T) {
	svc := &fakeAssociacaoService{err: &services.CandidatoNaoAssociavelError{Pendencias: []string{"Termos de uso não aceitos"}}}

	w := doAssociar(newAssociacaoRouter(svc, []string{"2600"}), "11111111111", `{"ids_vagas":["`+uuid.NewString()+`"]}`)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var body empmodels.AssociacaoRecusadaResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, []string{"Termos de uso não aceitos"}, body.Pendencias)
}

func TestAssociacaoCandidatoHandler_SemCurriculo(t *testing.T) {
	svc := &fakeAssociacaoService{}

	w := doAssociar(newAssociacaoRouter(svc, []string{"2600"}), "11111111111", `{"ids_vagas":["`+uuid.NewString()+`"]}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAssociacaoCandidatoHandler_Erro(t *testing.T) {
	svc := &fakeAssociacaoService{err: errors.New("pq: connection refused")}

	w := doAssociar(newAssociacaoRouter(svc, []string{"2600"}), "11111111111", `{"ids_vagas":["`+uuid.NewString()+`"]}`)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), "pq:")
}

func TestAssociacaoCandidatoHandler_EditorSemSecretaria(t *testing.T) {
	svc := &fakeAssociacaoService{}

	w := doAssociar(newAssociacaoRouter(svc, nil), "11111111111", `{"ids_vagas":["`+uuid.NewString()+`"]}`)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, svc.chamado)
}
