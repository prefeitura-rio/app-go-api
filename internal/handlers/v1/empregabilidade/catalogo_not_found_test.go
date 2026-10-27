package empregabilidade_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// DEV-11: PUT/DELETE de habilidade, área de atuação e comportamento/atitude
// com id inexistente devem responder 404 (como o GET por ID), não 200.

var catalogoNotFoundErrs = map[string]error{
	"sentinel": gorm.ErrRecordNotFound,
	"wrapped":  fmt.Errorf("erro ao atualizar: %w", gorm.ErrRecordNotFound),
}

func TestHabilidadeHandler_UpdateDelete_NotFound_Returns404(t *testing.T) {
	cases := map[string]struct {
		req     catalogoRequest
		wantMsg string
	}{
		"UpdateHabilidade":  {catalogoRequest{http.MethodPut, "/habilidades/999999", `{"nome":"X"}`}, "Habilidade não encontrada"},
		"DeleteHabilidade":  {catalogoRequest{http.MethodDelete, "/habilidades/999999", ""}, "Habilidade não encontrada"},
		"UpdateAreaAtuacao": {catalogoRequest{http.MethodPut, "/habilidades/areas-atuacao/999999", `{"nome":"X"}`}, "Área de atuação não encontrada"},
		"DeleteAreaAtuacao": {catalogoRequest{http.MethodDelete, "/habilidades/areas-atuacao/999999", ""}, "Área de atuação não encontrada"},
	}
	for op, tc := range cases {
		for errName, err := range catalogoNotFoundErrs {
			t.Run(op+"/"+errName, func(t *testing.T) {
				r := setupHabilidadeRouter(&mockHabilidadeRepo{err: err})
				w := httptest.NewRecorder()
				r.ServeHTTP(w, tc.req.newRequest())

				assert.Equal(t, http.StatusNotFound, w.Code, "body: %s", w.Body.String())
				assert.Equal(t, tc.wantMsg, decodeErrorMessage(t, w.Body.Bytes()))
			})
		}
	}
}

func TestComportamentoAtitudesHandler_UpdateDelete_NotFound_Returns404(t *testing.T) {
	cases := map[string]catalogoRequest{
		"UpdateComportamentoAtitudes": {http.MethodPut, "/comportamentos-atitudes/999999", `{"nome":"X"}`},
		"DeleteComportamentoAtitudes": {http.MethodDelete, "/comportamentos-atitudes/999999", ""},
	}
	for op, req := range cases {
		for errName, err := range catalogoNotFoundErrs {
			t.Run(op+"/"+errName, func(t *testing.T) {
				r := setupComportametoAtitudesHandlerRouter(&mockComportamentoAtitudesRepo{err: err})
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req.newRequest())

				assert.Equal(t, http.StatusNotFound, w.Code, "body: %s", w.Body.String())
				assert.Equal(t, "Comportamento/atitude não encontrado", decodeErrorMessage(t, w.Body.Bytes()))
			})
		}
	}
}
