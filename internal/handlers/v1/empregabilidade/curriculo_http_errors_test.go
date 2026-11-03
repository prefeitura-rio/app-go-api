package empregabilidade_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"

	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
)

// DEV-49: erros dos handlers de currículo com o status certo e sem vazar o
// texto do banco no corpo da resposta.

const curriculoErrCPF = "11122233344"

// curriculoDBLeak imita uma falha de banco sem constraint conhecida: nada
// deste texto pode aparecer no corpo da resposta.
var curriculoDBLeak = fmt.Errorf("erro ao buscar formação: %w", &pgconn.PgError{
	Code:    "57014",
	Message: `canceling statement due to statement timeout on relation "curriculo_formacao"`,
	Detail:  "Key (cpf)=(00000000000)",
})

var curriculoHTTPLeakMarkers = []string{"SQLSTATE", "statement", "relation", "curriculo_formacao", "00000000000", "erro ao buscar", "DETAIL"}

func assertNoCurriculoLeak(t *testing.T, body string) {
	t.Helper()
	for _, marker := range curriculoHTTPLeakMarkers {
		assert.NotContains(t, body, marker, "o corpo vaza detalhe interno")
	}
}

// mutationErrRepo mantém os GETs por ID funcionando e falha só em
// Update*/Delete*, para exercitar o erro depois da checagem de existência.
type mutationErrRepo struct {
	*mockCurriculoRepoH
	mutationErr error
}

func (m *mutationErrRepo) UpdateFormacao(context.Context, *empmodels.CurriculoFormacao) error {
	return m.mutationErr
}
func (m *mutationErrRepo) DeleteFormacao(context.Context, uuid.UUID) error { return m.mutationErr }
func (m *mutationErrRepo) UpdateIdioma(context.Context, *empmodels.CurriculoIdioma) error {
	return m.mutationErr
}
func (m *mutationErrRepo) DeleteIdioma(context.Context, uuid.UUID) error { return m.mutationErr }
func (m *mutationErrRepo) UpdateCursoComplementar(context.Context, *empmodels.CurriculoCursoComplementar) error {
	return m.mutationErr
}
func (m *mutationErrRepo) DeleteCursoComplementar(context.Context, uuid.UUID) error {
	return m.mutationErr
}
func (m *mutationErrRepo) UpdateExperiencia(context.Context, *empmodels.CurriculoExperiencia) error {
	return m.mutationErr
}
func (m *mutationErrRepo) DeleteExperiencia(context.Context, uuid.UUID) error { return m.mutationErr }
func (m *mutationErrRepo) UpdateConquista(context.Context, *empmodels.CurriculoConquista) error {
	return m.mutationErr
}
func (m *mutationErrRepo) DeleteConquista(context.Context, uuid.UUID) error { return m.mutationErr }

func ownedItemsRepo() *mockCurriculoRepoH {
	id := uuid.MustParse(validUUID)
	return &mockCurriculoRepoH{
		formacao:    &empmodels.CurriculoFormacao{ID: id, CPF: curriculoErrCPF},
		idioma:      &empmodels.CurriculoIdioma{ID: id, CPF: curriculoErrCPF},
		curso:       &empmodels.CurriculoCursoComplementar{ID: id, CPF: curriculoErrCPF},
		experiencia: &empmodels.CurriculoExperiencia{ID: id, CPF: curriculoErrCPF},
		conquista:   &empmodels.CurriculoConquista{ID: id, CPF: curriculoErrCPF},
	}
}

type curriculoReq struct {
	method string
	path   string
	body   string
}

func doCurriculoReq(r *gin.Engine, req curriculoReq) *httptest.ResponseRecorder {
	var httpReq *http.Request
	if req.body != "" {
		httpReq = httptest.NewRequest(req.method, req.path, strings.NewReader(req.body))
		httpReq.Header.Set("Content-Type", "application/json")
	} else {
		httpReq = httptest.NewRequest(req.method, req.path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)
	return w
}

// Toda rota de currículo que falha no banco responde 500 com mensagem genérica.
func TestCurriculoHandler_DBError_Returns500WithoutLeak(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cpf := curriculoErrCPF
	reqs := map[string]curriculoReq{
		"GetCurriculoCompleto":            {http.MethodGet, "/curriculo/" + cpf, ""},
		"CreateFormacao":                  {http.MethodPost, "/curriculo/formacoes", `{"nome_instituicao":"X"}`},
		"GetFormacaoByID":                 {http.MethodGet, "/curriculo/formacoes/" + validUUID, ""},
		"UpdateFormacao":                  {http.MethodPut, "/curriculo/formacoes/" + validUUID, `{}`},
		"DeleteFormacao":                  {http.MethodDelete, "/curriculo/formacoes/" + validUUID, ""},
		"ListFormacoesByCPF":              {http.MethodGet, "/curriculo/" + cpf + "/formacoes", ""},
		"CreateIdioma":                    {http.MethodPost, "/curriculo/idiomas", `{}`},
		"GetIdiomaByID":                   {http.MethodGet, "/curriculo/idiomas/" + validUUID, ""},
		"UpdateIdioma":                    {http.MethodPut, "/curriculo/idiomas/" + validUUID, `{}`},
		"DeleteIdioma":                    {http.MethodDelete, "/curriculo/idiomas/" + validUUID, ""},
		"ListIdiomasByCPF":                {http.MethodGet, "/curriculo/" + cpf + "/idiomas", ""},
		"CreateCursoComplementar":         {http.MethodPost, "/curriculo/cursos", `{}`},
		"GetCursoComplementarByID":        {http.MethodGet, "/curriculo/cursos/" + validUUID, ""},
		"UpdateCursoComplementar":         {http.MethodPut, "/curriculo/cursos/" + validUUID, `{}`},
		"DeleteCursoComplementar":         {http.MethodDelete, "/curriculo/cursos/" + validUUID, ""},
		"ListCursosComplementaresByCPF":   {http.MethodGet, "/curriculo/" + cpf + "/cursos", ""},
		"CreateExperiencia":               {http.MethodPost, "/curriculo/experiencias", `{}`},
		"GetExperienciaByID":              {http.MethodGet, "/curriculo/experiencias/" + validUUID, ""},
		"UpdateExperiencia":               {http.MethodPut, "/curriculo/experiencias/" + validUUID, `{}`},
		"DeleteExperiencia":               {http.MethodDelete, "/curriculo/experiencias/" + validUUID, ""},
		"ListExperienciasByCPF":           {http.MethodGet, "/curriculo/" + cpf + "/experiencias", ""},
		"CreateConquista":                 {http.MethodPost, "/curriculo/conquistas", `{}`},
		"GetConquistaByID":                {http.MethodGet, "/curriculo/conquistas/" + validUUID, ""},
		"UpdateConquista":                 {http.MethodPut, "/curriculo/conquistas/" + validUUID, `{}`},
		"DeleteConquista":                 {http.MethodDelete, "/curriculo/conquistas/" + validUUID, ""},
		"ListConquistasByCPF":             {http.MethodGet, "/curriculo/" + cpf + "/conquistas", ""},
		"ReplaceAllFormacoesAccordion":    {http.MethodPut, "/curriculo/" + cpf + "/formacoes", `{"formacoes":[],"idiomas":[]}`},
		"ReplaceAllExperienciasAccordion": {http.MethodPut, "/curriculo/" + cpf + "/experiencias", `{"experiencias":[],"conquistas":[]}`},
		"ReplaceAllConquistas":            {http.MethodPut, "/curriculo/" + cpf + "/conquistas", `[]`},
		"ReplaceAllIdiomas":               {http.MethodPut, "/curriculo/" + cpf + "/idiomas", `[]`},
		"ReplaceAllCursosComplementares":  {http.MethodPut, "/curriculo/" + cpf + "/cursos-complementares", `[]`},
		"UpsertSituacaoInteresses":        {http.MethodPut, "/curriculo/" + cpf + "/situacao-interesses", `{}`},
		"GetSituacaoInteresses":           {http.MethodGet, "/curriculo/" + cpf + "/situacao-interesses", ""},
	}

	for name, req := range reqs {
		t.Run(name, func(t *testing.T) {
			r := setupCurriculoRouter(&mockCurriculoRepoH{err: curriculoDBLeak}, cpf, false)
			w := doCurriculoReq(r, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
			assert.Contains(t, w.Body.String(), `"error":"Erro ao `)
			assertNoCurriculoLeak(t, w.Body.String())
		})
	}
}

// Falha no Update/Delete depois de achar o item também não vaza o erro.
func TestCurriculoHandler_MutationDBError_Returns500WithoutLeak(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reqs := map[string]curriculoReq{}
	for _, item := range []string{"formacoes", "idiomas", "cursos", "experiencias", "conquistas"} {
		reqs["PUT "+item] = curriculoReq{http.MethodPut, "/curriculo/" + item + "/" + validUUID, `{}`}
		reqs["DELETE "+item] = curriculoReq{http.MethodDelete, "/curriculo/" + item + "/" + validUUID, ""}
	}

	for name, req := range reqs {
		t.Run(name, func(t *testing.T) {
			repo := &mutationErrRepo{mockCurriculoRepoH: ownedItemsRepo(), mutationErr: curriculoDBLeak}
			r := setupCurriculoRouter(repo, curriculoErrCPF, false)
			w := doCurriculoReq(r, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
			assertNoCurriculoLeak(t, w.Body.String())
		})
	}
}

func TestCurriculoHandler_ItemNotFound_Returns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, item := range []string{"formacoes", "idiomas", "cursos", "experiencias", "conquistas"} {
		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			t.Run(method+" "+item, func(t *testing.T) {
				r := setupCurriculoRouter(&mockCurriculoRepoH{}, curriculoErrCPF, false)
				body := ""
				if method == http.MethodPut {
					body = `{}`
				}
				w := doCurriculoReq(r, curriculoReq{method, "/curriculo/" + item + "/" + validUUID, body})

				assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
			})
		}
	}
}

func TestCurriculoHandler_Create_ConstraintViolation_MapsTo409And400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	unique := fmt.Errorf("erro ao criar formação: %w", &pgconn.PgError{
		Code: "23505", ConstraintName: "uni_curriculo_formacao_cpf_escolaridade",
		Message: `duplicate key value violates unique constraint "uni_curriculo_formacao_cpf_escolaridade"`,
		Detail:  "Key (cpf)=(00000000000) already exists.",
	})
	fk := fmt.Errorf("erro ao criar formação: %w", &pgconn.PgError{
		Code: "23503", ConstraintName: "fk_curriculo_formacao_escolaridade_id",
		Message: `insert or update on table "curriculo_formacao" violates foreign key constraint`,
	})

	tests := []struct {
		name       string
		err        error
		path       string
		wantStatus int
	}{
		{"formacao duplicada", unique, "/curriculo/formacoes", http.StatusConflict},
		{"idioma duplicado", unique, "/curriculo/idiomas", http.StatusConflict},
		{"formacao com escolaridade inexistente", fk, "/curriculo/formacoes", http.StatusBadRequest},
		{"experiencia com referencia inexistente", fk, "/curriculo/experiencias", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupCurriculoRouter(&mockCurriculoRepoH{err: tt.err}, curriculoErrCPF, false)
			w := doCurriculoReq(r, curriculoReq{http.MethodPost, tt.path, `{}`})

			assert.Equal(t, tt.wantStatus, w.Code, w.Body.String())
			assertNoCurriculoLeak(t, w.Body.String())
			assert.NotContains(t, w.Body.String(), "uni_curriculo")
			assert.NotContains(t, w.Body.String(), "fk_curriculo")
		})
	}
}

func TestCurriculoHandler_Create_InvalidPayload_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, path := range []string{"/curriculo/formacoes", "/curriculo/idiomas", "/curriculo/cursos", "/curriculo/experiencias", "/curriculo/conquistas"} {
		t.Run(path, func(t *testing.T) {
			r := setupCurriculoRouter(&mockCurriculoRepoH{}, curriculoErrCPF, false)
			w := doCurriculoReq(r, curriculoReq{http.MethodPost, path, `{"cpf": 123`})

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
	}
}
