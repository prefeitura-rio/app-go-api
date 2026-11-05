package empregabilidade_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/clients"
	handlers "github.com/prefeitura-rio/app-go-api/internal/handlers/v1/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

func newConsultaCNPJRouterForbidden(t *testing.T, contentType, body string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	keycloak := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(auth.TokenResponse{AccessToken: "token-de-teste", ExpiresIn: 300})
	}))
	t.Cleanup(keycloak.Close)

	rmi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(rmi.Close)

	tokenManager := auth.NewServiceAccountTokenManager(keycloak.URL, "teste", "cliente", "segredo-de-teste")
	cnpjSvc := services.NewCNPJConsultaService(clients.NewRMIClient(rmi.URL, 5*time.Second), tokenManager)
	h := handlers.NewEmpresaHandler(services.NewEmpresaServiceWithInterface(&mockEmpresaRepoEmpH{})).WithCNPJConsulta(cnpjSvc)

	r := gin.New()
	r.GET("/empresas/consulta-cnpj/:cnpj", h.ConsultaCNPJ)
	return r
}

// DEV-364: a recusa do ext_authz da malha e a recusa de posse do RMI chegam
// com mensagens distintas na resposta e no log, sem o CNPJ em claro.
func TestEmpresaHandler_ConsultaCNPJ_Forbidden_DistinguishesOrigin(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		body         string
		wantError    string
		wantLogParts []string
	}{
		{
			name:         "ext_authz da malha",
			contentType:  "text/plain",
			body:         "denied by ext_authz: no action mapping found for endpoint",
			wantError:    "Consulta ao RMI recusada pelo controle de acesso do gateway (ext_authz), não pela regra de acesso ao CNPJ",
			wantLogParts: []string{"ConsultaCNPJ", "gateway da malha (ext_authz)", "denied by ext_authz: no action mapping found for endpoint"},
		},
		{
			name:         "regra de posse do RMI",
			contentType:  "application/json",
			body:         `{"error":"Access denied - you are not authorized to view this legal entity"}`,
			wantError:    "Acesso negado ao CNPJ",
			wantLogParts: []string{"ConsultaCNPJ", "acesso negado ao CNPJ", "Access denied - you are not authorized"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			previous := log.Writer()
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(previous) })

			r := newConsultaCNPJRouterForbidden(t, tt.contentType, tt.body)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/empresas/consulta-cnpj/"+cnpjAlfanumericoSintetico, nil))

			if w.Code != http.StatusForbidden {
				t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
			}
			var resp map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("invalid JSON body: %v", err)
			}
			if resp["error"] != tt.wantError {
				t.Errorf("expected error %q, got %v", tt.wantError, resp["error"])
			}

			logged := logs.String()
			for _, part := range tt.wantLogParts {
				if !strings.Contains(logged, part) {
					t.Errorf("expected log to contain %q, got %q", part, logged)
				}
			}
			if strings.Contains(strings.ToUpper(logged), cnpjAlfanumericoSintetico) {
				t.Errorf("log leaks the CNPJ: %q", logged)
			}
		})
	}
}
