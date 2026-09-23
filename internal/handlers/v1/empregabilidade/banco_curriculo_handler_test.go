package empregabilidade_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeBancoCurriculoService struct {
	items      []*empmodels.BancoCurriculoItem
	total      int64
	listErr    error
	gotFilter  empmodels.BancoCurriculoFilter
	gotPage    int
	gotSize    int
	detalhe    *empmodels.BancoCurriculoDetalhe
	detalheErr error
	gotCPF     string
	detalheHit bool
}

func (f *fakeBancoCurriculoService) List(_ context.Context, filter empmodels.BancoCurriculoFilter, page, pageSize int) ([]*empmodels.BancoCurriculoItem, int64, error) {
	f.gotFilter, f.gotPage, f.gotSize = filter, page, pageSize
	return f.items, f.total, f.listErr
}

func (f *fakeBancoCurriculoService) GetDetalhe(_ context.Context, cpf string) (*empmodels.BancoCurriculoDetalhe, error) {
	f.detalheHit = true
	f.gotCPF = cpf
	return f.detalhe, f.detalheErr
}

func newBancoCurriculoRouter(svc *fakeBancoCurriculoService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handlers.NewBancoCurriculoHandler(svc)
	r.GET("/banco-curriculos", h.List)
	r.GET("/banco-curriculos/:cpf", h.GetByCPF)
	return r
}

func doBancoGet(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

type bancoListResponse struct {
	Data []empmodels.BancoCurriculoItem `json:"data"`
	Meta struct {
		Page     int   `json:"page"`
		PageSize int   `json:"page_size"`
		Total    int64 `json:"total"`
	} `json:"meta"`
}

func TestBancoCurriculoHandler_List_PaginacaoPadrao(t *testing.T) {
	nome := "Ana"
	svc := &fakeBancoCurriculoService{
		items: []*empmodels.BancoCurriculoItem{{CPF: "11111111111", Nome: &nome, DataInclusao: time.Now()}},
		total: 1,
	}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, svc.gotPage)
	assert.Equal(t, 10, svc.gotSize)
	assert.Equal(t, "", svc.gotFilter.Search)

	var body bancoListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Data, 1)
	assert.Equal(t, "11111111111", body.Data[0].CPF)
	assert.Equal(t, 1, body.Meta.Page)
	assert.Equal(t, 10, body.Meta.PageSize)
	assert.Equal(t, int64(1), body.Meta.Total)
}

func TestBancoCurriculoHandler_List_RepassaBuscaEPagina(t *testing.T) {
	svc := &fakeBancoCurriculoService{items: []*empmodels.BancoCurriculoItem{}}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos?page=3&pageSize=25&search=ana%20claudia")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 3, svc.gotPage)
	assert.Equal(t, 25, svc.gotSize)
	assert.Equal(t, "ana claudia", svc.gotFilter.Search)
}

func TestBancoCurriculoHandler_List_PaginacaoInvalidaUsaPadrao(t *testing.T) {
	svc := &fakeBancoCurriculoService{items: []*empmodels.BancoCurriculoItem{}}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos?page=0&pageSize=500")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, svc.gotPage)
	assert.Equal(t, 10, svc.gotSize)
}

func TestBancoCurriculoHandler_List_Erro(t *testing.T) {
	svc := &fakeBancoCurriculoService{listErr: errors.New("pq: connection refused")}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), "pq:", "não expõe o erro do banco")
}

func TestBancoCurriculoHandler_GetByCPF_Encontrado(t *testing.T) {
	svc := &fakeBancoCurriculoService{detalhe: &empmodels.BancoCurriculoDetalhe{CPF: "11111111111"}}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos/11111111111")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "11111111111", svc.gotCPF)
	var body empmodels.BancoCurriculoDetalhe
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "11111111111", body.CPF)
}

func TestBancoCurriculoHandler_GetByCPF_AceitaCPFFormatado(t *testing.T) {
	svc := &fakeBancoCurriculoService{detalhe: &empmodels.BancoCurriculoDetalhe{CPF: "11111111111"}}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos/111.111.111-11")

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "11111111111", svc.gotCPF)
}

func TestBancoCurriculoHandler_GetByCPF_Invalido(t *testing.T) {
	svc := &fakeBancoCurriculoService{}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos/123")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, svc.detalheHit)
}

func TestBancoCurriculoHandler_GetByCPF_NaoEncontrado(t *testing.T) {
	svc := &fakeBancoCurriculoService{}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos/11111111111")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBancoCurriculoHandler_GetByCPF_Erro(t *testing.T) {
	svc := &fakeBancoCurriculoService{detalheErr: errors.New("pq: connection refused")}

	w := doBancoGet(newBancoCurriculoRouter(svc), "/banco-curriculos/11111111111")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NotContains(t, w.Body.String(), "pq:")
}
