package empregabilidade_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DEV-10: os handlers do catálogo de Habilidades/Áreas de Atuação e
// Comportamentos-Atitudes devem mapear erros de banco como os demais
// handlers de empregabilidade (DEV-3): unicidade → 409; FK/not-null/check → 400.

func foreignKeyViolationErr(constraint string) error {
	return &pq.Error{Code: pq.ErrorCode("23503"), Constraint: constraint}
}

func checkViolationErr(constraint string) error {
	return &pq.Error{Code: pq.ErrorCode("23514"), Constraint: constraint}
}

type catalogoRequest struct {
	method string
	path   string
	body   string
}

func (cr catalogoRequest) newRequest() *http.Request {
	var body io.Reader
	if cr.body != "" {
		body = bodyOf(cr.body)
	}
	req := httptest.NewRequest(cr.method, cr.path, body)
	if cr.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

var habilidadeMutations = map[string]catalogoRequest{
	"CreateHabilidade":    {http.MethodPost, "/habilidades", `{"nome":"Go"}`},
	"UpdateHabilidade":    {http.MethodPut, "/habilidades/10", `{"nome":"Go"}`},
	"DeleteHabilidade":    {http.MethodDelete, "/habilidades/10", ""},
	"CreateAreaAtuacao":   {http.MethodPost, "/habilidades/areas-atuacao", `{"nome":"TI"}`},
	"UpdateAreaAtuacao":   {http.MethodPut, "/habilidades/areas-atuacao/2", `{"nome":"TI"}`},
	"DeleteAreaAtuacao":   {http.MethodDelete, "/habilidades/areas-atuacao/2", ""},
	"AttachAreaAtuacao":   {http.MethodPost, "/habilidades/10/areas-atuacao/2", ""},
	"DetachAreaAtuacao":   {http.MethodDelete, "/habilidades/10/areas-atuacao/2", ""},
	"ReplaceAreasAtuacao": {http.MethodPut, "/habilidades/10/areas-atuacao", `{"area_ids":[1,2]}`},
}

var comportamentoMutations = map[string]catalogoRequest{
	"CreateComportamentoAtitudes": {http.MethodPost, "/comportamentos-atitudes", `{"nome":"Proatividade"}`},
	"UpdateComportamentoAtitudes": {http.MethodPut, "/comportamentos-atitudes/3", `{"nome":"Proatividade"}`},
	"DeleteComportamentoAtitudes": {http.MethodDelete, "/comportamentos-atitudes/3", ""},
}

var dbErrorCases = []struct {
	name       string
	err        error
	wantStatus int
	wantMsg    string
}{
	{"unique violation", uniqueViolationErr("emp_habilidades_nome_key"), http.StatusConflict, "Este nome já está cadastrado no sistema."},
	{"unique violation pivot", uniqueViolationErr("uk_habilidade_area"), http.StatusConflict, "Já existe um registro com essas informações no sistema."},
	{"foreign key violation", foreignKeyViolationErr("fk_emp_area_atuacao_habilidade_habilidade"), http.StatusBadRequest, "Referência inválida. Verifique se todos os IDs fornecidos existem no sistema."},
	{"not null violation", notNullViolationErr("nome"), http.StatusBadRequest, "O campo 'nome' é obrigatório e não pode estar vazio."},
	{"check violation", checkViolationErr("chk_x"), http.StatusBadRequest, "Valor inválido fornecido para um dos campos."},
}

func decodeErrorMessage(t *testing.T, body []byte) string {
	t.Helper()
	var resp map[string]any
	require.NoError(t, json.Unmarshal(body, &resp), "body: %s", string(body))
	msg, _ := resp["error"].(string)
	return msg
}

func TestHabilidadeHandler_DatabaseErrors_MappedToClientStatus(t *testing.T) {
	for op, cr := range habilidadeMutations {
		for _, tc := range dbErrorCases {
			t.Run(op+"/"+tc.name, func(t *testing.T) {
				r := setupHabilidadeRouter(&mockHabilidadeRepo{err: tc.err})
				w := httptest.NewRecorder()
				r.ServeHTTP(w, cr.newRequest())

				assert.Equal(t, tc.wantStatus, w.Code, "body: %s", w.Body.String())
				assert.Equal(t, tc.wantMsg, decodeErrorMessage(t, w.Body.Bytes()))
			})
		}
	}
}

func TestComportamentoAtitudesHandler_DatabaseErrors_MappedToClientStatus(t *testing.T) {
	for op, cr := range comportamentoMutations {
		for _, tc := range dbErrorCases {
			t.Run(op+"/"+tc.name, func(t *testing.T) {
				r := setupComportametoAtitudesHandlerRouter(&mockComportamentoAtitudesRepo{err: tc.err})
				w := httptest.NewRecorder()
				r.ServeHTTP(w, cr.newRequest())

				assert.Equal(t, tc.wantStatus, w.Code, "body: %s", w.Body.String())
				assert.Equal(t, tc.wantMsg, decodeErrorMessage(t, w.Body.Bytes()))
			})
		}
	}
}

// Erro que não é de constraint do banco continua 500, sem vazar o detalhe interno.
func TestCatalogoHandlers_UnknownError_StillInternalServerErrorWithoutLeak(t *testing.T) {
	internal := errors.New("pq: connection refused segredo-interno")
	for op, cr := range habilidadeMutations {
		t.Run(op, func(t *testing.T) {
			r := setupHabilidadeRouter(&mockHabilidadeRepo{err: internal})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, cr.newRequest())

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.NotContains(t, w.Body.String(), "segredo-interno")
		})
	}
	for op, cr := range comportamentoMutations {
		t.Run(op, func(t *testing.T) {
			r := setupComportametoAtitudesHandlerRouter(&mockComportamentoAtitudesRepo{err: internal})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, cr.newRequest())

			assert.Equal(t, http.StatusInternalServerError, w.Code)
			assert.NotContains(t, w.Body.String(), "segredo-interno")
		})
	}
}

// IDs de rota não positivos são entrada inválida (400), não erro interno.
func TestHabilidadeHandler_NonPositiveRouteIDs_ReturnBadRequest(t *testing.T) {
	cases := map[string]catalogoRequest{
		"Attach habilidade 0":  {http.MethodPost, "/habilidades/0/areas-atuacao/2", ""},
		"Attach area -1":       {http.MethodPost, "/habilidades/10/areas-atuacao/-1", ""},
		"Detach habilidade -5": {http.MethodDelete, "/habilidades/-5/areas-atuacao/2", ""},
		"Detach area 0":        {http.MethodDelete, "/habilidades/10/areas-atuacao/0", ""},
		"Replace habilidade 0": {http.MethodPut, "/habilidades/0/areas-atuacao", `{"area_ids":[1]}`},
	}
	for name, cr := range cases {
		t.Run(name, func(t *testing.T) {
			r := setupHabilidadeRouter(&mockHabilidadeRepo{})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, cr.newRequest())

			assert.Equal(t, http.StatusBadRequest, w.Code, "body: %s", w.Body.String())
		})
	}
}
