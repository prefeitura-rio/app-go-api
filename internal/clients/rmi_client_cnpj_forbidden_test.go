package clients

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// DEV-364: um 403 na consulta de CNPJ pode vir do ext_authz da malha
// (cerbos-ext-authz/Heimdall, corpo em texto puro "denied by ext_authz: ...")
// ou da regra de posse do RMI (corpo JSON {"error": "..."}). As duas causas
// precisam chegar distintas ao handler e ao log, com o corpo saneado.
func TestGetLegalEntityByCNPJ_Forbidden_DistinguishesOrigin(t *testing.T) {
	tests := []struct {
		name        string
		cnpj        string
		contentType string
		body        string
		wantIs      error
		wantNotIs   error
		wantMessage string
		forbidden   []string
	}{
		{
			name:        "recusa do ext_authz da malha",
			cnpj:        "12ABC34501DE35",
			contentType: "text/plain",
			body:        "denied by ext_authz: no action mapping found for endpoint",
			wantIs:      ErrRMIGatewayDenied,
			wantNotIs:   ErrCNPJAccessDenied,
			wantMessage: "acesso ao RMI negado pelo gateway da malha (ext_authz): denied by ext_authz: no action mapping found for endpoint",
		},
		{
			name:        "RBAC do Istio",
			cnpj:        "12345678000195",
			contentType: "text/plain",
			body:        "RBAC: access denied",
			wantIs:      ErrRMIGatewayDenied,
			wantNotIs:   ErrCNPJAccessDenied,
			wantMessage: "acesso ao RMI negado pelo gateway da malha (ext_authz): RBAC: access denied",
		},
		{
			name:        "ext_authz que ecoa o CNPJ alfanumérico e dados pessoais",
			cnpj:        "12.abc.345/01de-35",
			contentType: "text/plain",
			body:        "denied by ext_authz: authorization error: GET /v1/legal-entity/12ABC34501DE35 user 11122233344 x@y.com",
			wantIs:      ErrRMIGatewayDenied,
			wantNotIs:   ErrCNPJAccessDenied,
			forbidden:   []string{"12ABC34501DE35", "12abc34501de35", "11122233344", "x@y.com"},
		},
		{
			name:        "ext_authz que ecoa o CNPJ alfanumérico em outra formatação",
			cnpj:        "12ABC34501DE35",
			contentType: "text/plain",
			body:        "denied by ext_authz: authorization error for 12.abc.345/01DE-35",
			wantIs:      ErrRMIGatewayDenied,
			wantNotIs:   ErrCNPJAccessDenied,
			wantMessage: "acesso ao RMI negado pelo gateway da malha (ext_authz): denied by ext_authz: authorization error for [omitido]",
		},
		{
			name:        "RMI que ecoa o CNPJ numérico formatado",
			cnpj:        "12345678000195",
			contentType: "application/json",
			body:        `{"error":"Access denied to 12.345.678/0001-95"}`,
			wantIs:      ErrCNPJAccessDenied,
			wantNotIs:   ErrRMIGatewayDenied,
			wantMessage: `acesso negado ao CNPJ: {"error":"Access denied to [omitido]"}`,
		},
		{
			name:        "regra de posse do RMI",
			cnpj:        "12345678000195",
			contentType: "application/json",
			body:        `{"error":"Access denied - you are not authorized to view this legal entity"}`,
			wantIs:      ErrCNPJAccessDenied,
			wantNotIs:   ErrRMIGatewayDenied,
			wantMessage: `acesso negado ao CNPJ: {"error":"Access denied - you are not authorized to view this legal entity"}`,
		},
		{
			name:        "regra de posse do RMI que ecoa o CNPJ",
			cnpj:        "12ABC34501DE35",
			contentType: "application/json",
			body:        `{"error":"Access denied to 12ABC34501DE35"}`,
			wantIs:      ErrCNPJAccessDenied,
			wantNotIs:   ErrRMIGatewayDenied,
			wantMessage: `acesso negado ao CNPJ: {"error":"Access denied to [omitido]"}`,
		},
		{
			name:        "corpo vazio mantém a recusa do RMI",
			cnpj:        "12345678000195",
			body:        "",
			wantIs:      ErrCNPJAccessDenied,
			wantNotIs:   ErrRMIGatewayDenied,
			wantMessage: "acesso negado ao CNPJ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tt.contentType != "" {
					w.Header().Set("Content-Type", tt.contentType)
				}
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			client := NewRMIClient(srv.URL, 5*time.Second, WithName(uniqueClientName(t, "cnpj-403-")))
			_, err := client.GetLegalEntityByCNPJ(context.Background(), "test-token", tt.cnpj)
			if err == nil {
				t.Fatal("expected error for 403")
			}
			if !errors.Is(err, tt.wantIs) {
				t.Errorf("expected errors.Is(err, %v), got %v", tt.wantIs, err)
			}
			if errors.Is(err, tt.wantNotIs) {
				t.Errorf("did not expect errors.Is(err, %v), got %v", tt.wantNotIs, err)
			}
			var forbidden *CNPJForbiddenError
			if !errors.As(err, &forbidden) {
				t.Errorf("expected *CNPJForbiddenError, got %T", err)
			}
			if tt.wantMessage != "" && err.Error() != tt.wantMessage {
				t.Errorf("expected message %q, got %q", tt.wantMessage, err.Error())
			}
			for _, leak := range tt.forbidden {
				if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(leak)) {
					t.Errorf("error message leaks %q: %q", leak, err.Error())
				}
			}
		})
	}
}
