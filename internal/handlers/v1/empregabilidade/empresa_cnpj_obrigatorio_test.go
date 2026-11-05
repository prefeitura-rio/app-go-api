package empregabilidade_test

import (
	"bytes"
	"context"
	"encoding/json"
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
// gravado e devolvido na forma canônica: 14 caracteres, maiúsculas, sem
// pontuação. Assim a chave cabe na coluna e é alcançável por /empresas/:cnpj.
func TestEmpresaHandler_Create_CNPJValido_GravaFormaCanonica(t *testing.T) {
	tests := []struct {
		name      string
		cnpj      string
		canonical string
	}{
		{"numérico", "11222333000181", "11222333000181"},
		{"numérico formatado", "11.222.333/0001-81", "11222333000181"},
		{"espaços em excesso", "      11 222 333 0001 81        ", "11222333000181"},
		{"alfanumérico", cnpjAlfanumericoSintetico, cnpjAlfanumericoSintetico},
		{"alfanumérico formatado em minúsculas", "12.abc.345/01de-35", cnpjAlfanumericoSintetico},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingEmpresaRepo{}
			w := doEmpresaRequest(newEmpresaRecordingRouter(repo), http.MethodPost, "/empresas",
				`{"cnpj":"`+tt.cnpj+`","razao_social":"Empresa Sintetica"}`)

			if w.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
			}
			if len(repo.created) != 1 || repo.created[0].CNPJ != tt.canonical {
				t.Fatalf("expected one Create with cnpj %q, got %+v", tt.canonical, repo.created)
			}
			var body empmodels.Empresa
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid JSON body: %v", err)
			}
			if body.CNPJ != tt.canonical {
				t.Errorf("expected cnpj %q in response, got %q", tt.canonical, body.CNPJ)
			}
		})
	}
}

// mapEmpresaRepo guarda as empresas por chave, como a tabela, para provar que o
// que o POST grava é alcançado por GET, PUT e DELETE /empresas/:cnpj.
type mapEmpresaRepo struct {
	rows map[string]*empmodels.Empresa
}

func (m *mapEmpresaRepo) Create(_ context.Context, e *empmodels.Empresa) (string, error) {
	copied := *e
	m.rows[e.CNPJ] = &copied
	return e.CNPJ, nil
}
func (m *mapEmpresaRepo) GetByID(_ context.Context, cnpj string) (*empmodels.Empresa, error) {
	return m.rows[cnpj], nil
}
func (m *mapEmpresaRepo) Update(_ context.Context, e *empmodels.Empresa) error {
	if row, ok := m.rows[e.CNPJ]; ok {
		row.RazaoSocial = e.RazaoSocial
	}
	return nil
}
func (m *mapEmpresaRepo) Delete(_ context.Context, cnpj string) error {
	delete(m.rows, cnpj)
	return nil
}
func (m *mapEmpresaRepo) List(_ context.Context, _ empmodels.EmpresaFilter, _, _ int) ([]*empmodels.Empresa, int, error) {
	return nil, 0, nil
}
func (m *mapEmpresaRepo) Upsert(_ context.Context, _ *empmodels.Empresa) error { return nil }

func TestEmpresaHandler_CNPJFormatado_AlcancavelPorGetPutDelete(t *testing.T) {
	repo := &mapEmpresaRepo{rows: map[string]*empmodels.Empresa{}}
	r := setupEmpresaHandlerRouter(repo)

	w := doEmpresaRequest(r, http.MethodPost, "/empresas", `{"cnpj":"11.222.333/0001-81","razao_social":"Empresa Sintetica"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created empmodels.Empresa
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("POST: invalid JSON body: %v", err)
	}
	path := "/empresas/" + created.CNPJ

	if w := doEmpresaRequest(r, http.MethodGet, path, ""); w.Code != http.StatusOK {
		t.Fatalf("GET %s: expected 200, got %d: %s", path, w.Code, w.Body.String())
	}
	if w := doEmpresaRequest(r, http.MethodPut, path, `{"razao_social":"Empresa Sintetica 2"}`); w.Code != http.StatusOK {
		t.Fatalf("PUT %s: expected 200, got %d: %s", path, w.Code, w.Body.String())
	}
	if got := repo.rows[created.CNPJ]; got == nil || got.RazaoSocial != "Empresa Sintetica 2" {
		t.Fatalf("PUT %s did not reach the stored row: %+v", path, got)
	}
	if w := doEmpresaRequest(r, http.MethodDelete, path, ""); w.Code != http.StatusOK {
		t.Fatalf("DELETE %s: expected 200, got %d: %s", path, w.Code, w.Body.String())
	}
	if len(repo.rows) != 0 {
		t.Fatalf("DELETE %s left rows behind: %v", path, repo.rows)
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

// DEV-261: o CNPJ do path manda, na forma canônica; um cnpj vazio no corpo não
// troca a chave.
func TestEmpresaHandler_Update_CNPJValido_UsaPathCanonico(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{"numérico", "/empresas/11222333000181", "11222333000181"},
		{"alfanumérico em minúsculas", "/empresas/12abc34501de35", cnpjAlfanumericoSintetico},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingEmpresaRepo{}
			w := doEmpresaRequest(newEmpresaRecordingRouter(repo), http.MethodPut, tt.path, `{"cnpj":"","razao_social":"Nova"}`)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
			}
			if len(repo.updated) != 1 || repo.updated[0].CNPJ != tt.want {
				t.Fatalf("expected one Update with cnpj %q, got %+v", tt.want, repo.updated)
			}
		})
	}
}
