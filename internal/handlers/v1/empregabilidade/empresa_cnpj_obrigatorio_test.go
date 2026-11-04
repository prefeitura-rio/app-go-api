package empregabilidade_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

const cnpjInvalidoMsg = `{"error":"CNPJ inválido: deve ter 14 caracteres, 12 letras ou dígitos seguidos de 2 dígitos verificadores válidos"}`

// recordingEmpresaRepo registra as escritas para provar que nada é gravado
// quando o CNPJ é recusado.
type recordingEmpresaRepo struct {
	created []*empmodels.Empresa
	updated []*empmodels.Empresa
}

func (m *recordingEmpresaRepo) Create(_ context.Context, e *empmodels.Empresa) (string, error) {
	m.created = append(m.created, e)
	return e.CNPJ, nil
}
func (m *recordingEmpresaRepo) GetByID(_ context.Context, _ string) (*empmodels.Empresa, error) {
	return nil, nil
}
func (m *recordingEmpresaRepo) Update(_ context.Context, e *empmodels.Empresa) error {
	m.updated = append(m.updated, e)
	return nil
}
func (m *recordingEmpresaRepo) Delete(_ context.Context, _ string) error { return nil }
func (m *recordingEmpresaRepo) List(_ context.Context, _ empmodels.EmpresaFilter, _, _ int) ([]*empmodels.Empresa, int, error) {
	return nil, 0, nil
}
func (m *recordingEmpresaRepo) Upsert(_ context.Context, _ *empmodels.Empresa) error { return nil }

func newEmpresaRecordingRouter(repo *recordingEmpresaRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handlers.NewEmpresaHandler(services.NewEmpresaServiceWithInterface(repo))
	r := gin.New()
	r.POST("/empresas", h.Create)
	r.PUT("/empresas/:cnpj", h.Update)
	return r
}

func doEmpresaRequest(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// DEV-261: POST sem CNPJ válido responde 400 e não grava chave primária.
func TestEmpresaHandler_Create_CNPJInvalido_BadRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"corpo vazio", `{}`},
		{"cnpj ausente", `{"razao_social":"Empresa Sintetica"}`},
		{"cnpj vazio", `{"cnpj":"","razao_social":"Empresa Sintetica"}`},
		{"cnpj só espaços", `{"cnpj":"   ","razao_social":"Empresa Sintetica"}`},
		{"cnpj curto", `{"cnpj":"1122233300018"}`},
		{"dígito verificador errado", `{"cnpj":"11222333000182"}`},
		{"alfanumérico com DV errado", `{"cnpj":"12ABC34501DE36"}`},
		{"caracteres repetidos", `{"cnpj":"00000000000000"}`},
		{"caractere fora do formato", `{"cnpj":"11222333000#81"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingEmpresaRepo{}
			w := doEmpresaRequest(newEmpresaRecordingRouter(repo), http.MethodPost, "/empresas", tt.body)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
			if got := w.Body.String(); got != cnpjInvalidoMsg {
				t.Errorf("expected body %s, got %s", cnpjInvalidoMsg, got)
			}
			if len(repo.created) != 0 {
				t.Errorf("repository Create must not be called, got %d calls", len(repo.created))
			}
		})
	}
}

// DEV-261: CNPJ válido, numérico ou alfanumérico, com ou sem formatação, é
// gravado como o cliente enviou.
func TestEmpresaHandler_Create_CNPJValido_Created(t *testing.T) {
	tests := []struct {
		name string
		cnpj string
	}{
		{"numérico", "11222333000181"},
		{"numérico formatado", "11.222.333/0001-81"},
		{"alfanumérico", cnpjAlfanumericoSintetico},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingEmpresaRepo{}
			w := doEmpresaRequest(newEmpresaRecordingRouter(repo), http.MethodPost, "/empresas",
				`{"cnpj":"`+tt.cnpj+`","razao_social":"Empresa Sintetica"}`)

			if w.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
			}
			if len(repo.created) != 1 || repo.created[0].CNPJ != tt.cnpj {
				t.Fatalf("expected one Create with cnpj %q, got %+v", tt.cnpj, repo.created)
			}
		})
	}
}

// DEV-261: PUT com CNPJ inválido no path responde 400 e não chega ao banco.
func TestEmpresaHandler_Update_CNPJInvalido_BadRequest(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"espaço", "/empresas/%20"},
		{"curto", "/empresas/1122233300018"},
		{"dígito verificador errado", "/empresas/11222333000182"},
		{"não numérico", "/empresas/abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingEmpresaRepo{}
			w := doEmpresaRequest(newEmpresaRecordingRouter(repo), http.MethodPut, tt.path, `{"razao_social":"Nova"}`)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
			if got := w.Body.String(); got != cnpjInvalidoMsg {
				t.Errorf("expected body %s, got %s", cnpjInvalidoMsg, got)
			}
			if len(repo.updated) != 0 {
				t.Errorf("repository Update must not be called, got %d calls", len(repo.updated))
			}
		})
	}
}

// DEV-261: o CNPJ do path manda; um cnpj vazio no corpo não troca a chave.
func TestEmpresaHandler_Update_CNPJValido_UsaPath(t *testing.T) {
	repo := &recordingEmpresaRepo{}
	w := doEmpresaRequest(newEmpresaRecordingRouter(repo), http.MethodPut, "/empresas/11222333000181",
		`{"cnpj":"","razao_social":"Nova"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if len(repo.updated) != 1 || repo.updated[0].CNPJ != "11222333000181" {
		t.Fatalf("expected one Update with cnpj from path, got %+v", repo.updated)
	}
}
