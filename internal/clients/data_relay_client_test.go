package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newDataRelayTestServer(t *testing.T, status int, contentType string, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/mailman" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func sendTestEmail(t *testing.T, baseURL string) error {
	t.Helper()
	client := NewDataRelayClient(baseURL, "test-api-key", 5*time.Second)
	return client.SendEmail(context.Background(), &EmailRequest{
		ToAddresses: []string{"cidadao.ficticio@example.com"},
		Subject:     "Assunto de teste",
		Body:        "Corpo de teste",
	})
}

func TestDataRelayClient_SendEmail_Success_ReturnsNil(t *testing.T) {
	server := newDataRelayTestServer(t, http.StatusOK, "application/json", `{"status":"sent"}`)

	if err := sendTestEmail(t, server.URL); err != nil {
		t.Fatalf("SendEmail() error = %v, want nil", err)
	}
}

func TestDataRelayClient_SendEmail_ErrorStatus_SanitizesResponseBody(t *testing.T) {
	const (
		syntheticEmail = "cidadao.ficticio@example.com"
		syntheticCPF   = "123.456.789-09"
	)

	tests := []struct {
		name        string
		status      int
		contentType string
		body        string
		want        string
	}{
		{
			name:        "JSON error echoing the recipient e-mail",
			status:      http.StatusBadRequest,
			contentType: "application/json",
			body:        `{"error":"invalid recipient ` + syntheticEmail + `","to_addresses":["` + syntheticEmail + `"]}`,
			want:        `data relay API returned status 400: {"error":"invalid recipient [omitido]"}`,
		},
		{
			name:        "JSON with only fields other than error and message",
			status:      http.StatusUnprocessableEntity,
			contentType: "application/json",
			body:        `{"detail":"` + syntheticEmail + `","body":"Olá, CPF ` + syntheticCPF + `"}`,
			want:        "data relay API returned status 422",
		},
		{
			name:        "JSON message with CPF and line breaks",
			status:      http.StatusInternalServerError,
			contentType: "application/json",
			body:        `{"message":"falha\nao enviar\r\npara CPF ` + syntheticCPF + `"}`,
			want:        `data relay API returned status 500: {"message":"falha ao enviar para CPF [omitido]"}`,
		},
		{
			name:        "plain text body echoing the e-mail content",
			status:      http.StatusBadGateway,
			contentType: "text/plain",
			body:        "upstream rejected message to " + syntheticEmail + "\nCorpo de teste",
			want:        "data relay API returned status 502",
		},
		{
			name:        "HTML error page",
			status:      http.StatusServiceUnavailable,
			contentType: "text/html",
			body:        "<html><body><h1>503</h1><p>" + syntheticEmail + "</p></body></html>",
			want:        "data relay API returned status 503",
		},
		{
			name:   "empty body",
			status: http.StatusTooManyRequests,
			body:   "",
			want:   "data relay API returned status 429",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newDataRelayTestServer(t, tt.status, tt.contentType, tt.body)

			err := sendTestEmail(t, server.URL)
			if err == nil {
				t.Fatal("SendEmail() error = nil, want error")
			}
			msg := err.Error()
			if msg != tt.want {
				t.Errorf("SendEmail() error = %q, want %q", msg, tt.want)
			}
			for _, leak := range []string{syntheticEmail, syntheticCPF, "Corpo de teste", "\n", "\r"} {
				if strings.Contains(msg, leak) {
					t.Errorf("SendEmail() error %q leaks %q", msg, leak)
				}
			}
		})
	}
}

func TestDataRelayClient_SendEmail_ErrorStatus_TruncatesLongMessage(t *testing.T) {
	server := newDataRelayTestServer(t, http.StatusBadRequest, "application/json",
		`{"message":"`+strings.Repeat("a", 500)+`"}`)

	err := sendTestEmail(t, server.URL)
	if err == nil {
		t.Fatal("SendEmail() error = nil, want error")
	}
	want := `data relay API returned status 400: {"message":"` + strings.Repeat("a", rmiErrorSnippetMaxRunes) + `…"}`
	if got := err.Error(); got != want {
		t.Errorf("SendEmail() error = %q, want %q", got, want)
	}
}

func TestDataRelayClient_SendEmail_ErrorStatus_LargeBodyIsBounded(t *testing.T) {
	// A JSON body larger than the read limit is cut, so it is no longer valid
	// JSON and nothing of it may reach the error message.
	large := `{"message":"` + strings.Repeat("x", 2*rmiErrorBodyDrainLimit) + `"}`
	server := newDataRelayTestServer(t, http.StatusInternalServerError, "application/json", large)

	err := sendTestEmail(t, server.URL)
	if err == nil {
		t.Fatal("SendEmail() error = nil, want error")
	}
	if got, want := err.Error(), "data relay API returned status 500"; got != want {
		t.Errorf("SendEmail() error = %q, want %q", got, want)
	}
}
