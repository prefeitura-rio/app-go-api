package empregabilidade_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	empmodels "github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

// DEV-183: CNPJ alfanumérico sintético no formato novo da Receita Federal.
const cnpjAlfanumericoSintetico = "12ABC34501DE35"

// newConsultaCNPJRouter monta o handler de consulta com um Keycloak e um RMI
// falsos (httptest). rmiPaths recebe o caminho de cada chamada ao RMI.
func newConsultaCNPJRouter(t *testing.T, rmiPaths *[]string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	keycloak := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(auth.TokenResponse{AccessToken: "token-de-teste", ExpiresIn: 300})
	}))
	t.Cleanup(keycloak.Close)

	rmi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*rmiPaths = append(*rmiPaths, r.URL.Path)
		_ = json.NewEncoder(w).Encode(models.LegalEntityFull{CNPJ: cnpjAlfanumericoSintetico, RazaoSocial: "Empresa Sintetica"})
	}))
	t.Cleanup(rmi.Close)

	tokenManager := auth.NewServiceAccountTokenManager(keycloak.URL, "teste", "cliente", "segredo-de-teste")
	cnpjSvc := services.NewCNPJConsultaService(clients.NewRMIClient(rmi.URL, 5*time.Second), tokenManager)
	h := handlers.NewEmpresaHandler(services.NewEmpresaServiceWithInterface(&mockEmpresaRepoEmpH{})).WithCNPJConsulta(cnpjSvc)

	r := gin.New()
	r.GET("/empresas/consulta-cnpj/:cnpj", h.ConsultaCNPJ)
	return r
}

func TestEmpresaHandler_ConsultaCNPJ_Alfanumerico_Success(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"sem formatação", "/empresas/consulta-cnpj/" + cnpjAlfanumericoSintetico},
		{"minúsculas", "/empresas/consulta-cnpj/12abc34501de35"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rmiPaths []string
			r := newConsultaCNPJRouter(t, &rmiPaths)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
			}
			if len(rmiPaths) != 1 || rmiPaths[0] != "/v1/legal-entity/"+cnpjAlfanumericoSintetico {
				t.Errorf("expected one RMI call to /v1/legal-entity/%s, got %v", cnpjAlfanumericoSintetico, rmiPaths)
			}
			var body models.LegalEntityConsultaResponse
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid JSON body: %v", err)
			}
			if body.CNPJ != cnpjAlfanumericoSintetico {
				t.Errorf("expected cnpj %s, got %s", cnpjAlfanumericoSintetico, body.CNPJ)
			}
		})
	}
}

func TestEmpresaHandler_ConsultaCNPJ_FormatoInvalido_BadRequest(t *testing.T) {
	tests := []struct {
		name string
		cnpj string
	}{
		{"curto demais", "12ABC34501DE3"},
		{"longo demais", "12ABC34501DE355"},
		{"letra no dígito verificador", "12ABC34501DEA5"},
		{"caractere fora do formato", "12ABC34501D%2A35"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rmiPaths []string
			r := newConsultaCNPJRouter(t, &rmiPaths)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/empresas/consulta-cnpj/"+tt.cnpj, nil))

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
			want := `{"error":"CNPJ inválido: deve ter 14 caracteres, 12 letras ou dígitos seguidos de 2 dígitos verificadores"}`
			if got := w.Body.String(); got != want {
				t.Errorf("expected body %s, got %s", want, got)
			}
			if len(rmiPaths) != 0 {
				t.Errorf("RMI must not be called for an invalid CNPJ, got %v", rmiPaths)
			}
		})
	}
}

func TestEmpresaHandler_CRUD_AceitaChaveAlfanumerica(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &mockEmpresaRepoEmpH{
		cnpj:   cnpjAlfanumericoSintetico,
		entity: &empmodels.Empresa{CNPJ: cnpjAlfanumericoSintetico, RazaoSocial: "Empresa Sintetica"},
	}
	r := setupEmpresaHandlerRouter(repo)

	requests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{http.MethodPost, "/empresas", `{"cnpj":"` + cnpjAlfanumericoSintetico + `","razao_social":"Empresa Sintetica"}`, http.StatusCreated},
		{http.MethodGet, "/empresas/" + cnpjAlfanumericoSintetico, "", http.StatusOK},
		{http.MethodPut, "/empresas/" + cnpjAlfanumericoSintetico, `{"razao_social":"Empresa Sintetica 2"}`, http.StatusOK},
		{http.MethodDelete, "/empresas/" + cnpjAlfanumericoSintetico, "", http.StatusOK},
	}
	for _, rq := range requests {
		t.Run(rq.method, func(t *testing.T) {
			req := httptest.NewRequest(rq.method, rq.path, bytes.NewBufferString(rq.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != rq.want {
				t.Fatalf("%s %s: expected %d, got %d: %s", rq.method, rq.path, rq.want, w.Code, w.Body.String())
			}
			if rq.method != http.MethodDelete && !bytes.Contains(w.Body.Bytes(), []byte(`"cnpj":"`+cnpjAlfanumericoSintetico+`"`)) {
				t.Errorf("%s %s: expected cnpj %s in body, got %s", rq.method, rq.path, cnpjAlfanumericoSintetico, w.Body.String())
			}
		})
	}
}
