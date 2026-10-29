package empregabilidade_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

type fakeParticipacaoService struct {
	status    *empregabilidade.ParticipacaoBancoStatus
	err       error
	gotCPF    string
	gotLigado *bool
}

func (f *fakeParticipacaoService) Get(_ context.Context, cpf string) (*empregabilidade.ParticipacaoBancoStatus, error) {
	f.gotCPF = cpf
	return f.status, f.err
}

func (f *fakeParticipacaoService) Atualizar(_ context.Context, cpf string, disponivel bool) (*empregabilidade.ParticipacaoBancoStatus, error) {
	f.gotCPF, f.gotLigado = cpf, &disponivel
	return f.status, f.err
}

const cpfDono = "11144477735"

func participacaoRouter(svc *fakeParticipacaoService, cpfLogado, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if cpfLogado != "" {
			c.Set(middlewares.UserCPFKey, cpfLogado)
		}
		if role != "" {
			c.Set(middlewares.UserRoleKey, role)
		}
	})
	h := handlers.NewParticipacaoBancoHandler(svc)
	r.GET("/curriculo/:cpf/banco-curriculos", h.Get)
	r.PUT("/curriculo/:cpf/banco-curriculos", h.Atualizar)
	return r
}

func TestParticipacaoBancoHandler_Get_DonoVeStatus(t *testing.T) {
	svc := &fakeParticipacaoService{status: &empregabilidade.ParticipacaoBancoStatus{TemCurriculo: true, Disponivel: true, Pendencias: []string{}}}
	w := httptest.NewRecorder()

	participacaoRouter(svc, cpfDono, "").ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/curriculo/"+cpfDono+"/banco-curriculos", nil))

	require.Equal(t, http.StatusOK, w.Code)
	var corpo map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &corpo))
	assert.Equal(t, true, corpo["disponivel"])
	assert.Equal(t, cpfDono, svc.gotCPF)
}

func TestParticipacaoBancoHandler_Get_OutroCPFNegado(t *testing.T) {
	svc := &fakeParticipacaoService{}
	w := httptest.NewRecorder()

	participacaoRouter(svc, "22222222222", "").ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/curriculo/"+cpfDono+"/banco-curriculos", nil))

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, svc.gotCPF)
}

func TestParticipacaoBancoHandler_Get_SemLogin(t *testing.T) {
	w := httptest.NewRecorder()

	participacaoRouter(&fakeParticipacaoService{}, "", "").ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/curriculo/"+cpfDono+"/banco-curriculos", nil))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestParticipacaoBancoHandler_Atualizar_Liga(t *testing.T) {
	svc := &fakeParticipacaoService{status: &empregabilidade.ParticipacaoBancoStatus{Disponivel: true, Pendencias: []string{}}}
	w := httptest.NewRecorder()

	participacaoRouter(svc, cpfDono, "").ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/curriculo/"+cpfDono+"/banco-curriculos", bodyOf(`{"disponivel": true}`)))

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, svc.gotLigado)
	assert.True(t, *svc.gotLigado)
}

func TestParticipacaoBancoHandler_Atualizar_SemCampoDisponivel(t *testing.T) {
	svc := &fakeParticipacaoService{}
	w := httptest.NewRecorder()

	participacaoRouter(svc, cpfDono, "").ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/curriculo/"+cpfDono+"/banco-curriculos", bodyOf(`{}`)))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Nil(t, svc.gotLigado)
}

func TestParticipacaoBancoHandler_Atualizar_SemCurriculo(t *testing.T) {
	svc := &fakeParticipacaoService{err: services.ErrSemCurriculoParaBanco}
	w := httptest.NewRecorder()

	participacaoRouter(svc, cpfDono, "").ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/curriculo/"+cpfDono+"/banco-curriculos", bodyOf(`{"disponivel": true}`)))

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var corpo map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &corpo))
	assert.Equal(t, "Crie o currículo antes de escolher se ele aparece no banco de currículos", corpo["error"])
}

// A escolha é do cidadão: nem o admin liga ou desliga o currículo de outra pessoa.
func TestParticipacaoBancoHandler_Atualizar_AdminNaoAlteraOutroCPF(t *testing.T) {
	svc := &fakeParticipacaoService{}
	w := httptest.NewRecorder()

	participacaoRouter(svc, "22222222222", "ADMIN").ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/curriculo/"+cpfDono+"/banco-curriculos", bodyOf(`{"disponivel": false}`)))

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Nil(t, svc.gotLigado)
}

func TestParticipacaoBancoHandler_Get_ErroInterno(t *testing.T) {
	w := httptest.NewRecorder()

	participacaoRouter(&fakeParticipacaoService{err: errTest}, cpfDono, "").ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/curriculo/"+cpfDono+"/banco-curriculos", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
