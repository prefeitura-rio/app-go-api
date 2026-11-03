package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CPF, nome e e-mail sintéticos.
const (
	accessLogCPF   = "00011122233"
	accessLogName  = "Fulano Sintetico"
	accessLogEmail = "fulano.sintetico@example.invalid"
)

func newAccessLogRouter(out *bytes.Buffer) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AccessLogger(out))
	r.GET("/api/v1/courses/:courseId/enrollments", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/api/v1/enrollments/user/:cpf", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func serveAccessLog(t *testing.T, target string) string {
	t.Helper()
	var out bytes.Buffer
	r := newAccessLogRouter(&out)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
	return out.String()
}

func TestAccessLogger_SearchQuery_RedactsValues(t *testing.T) {
	tests := []struct {
		name   string
		target string
		secret string
	}{
		{"cpf", "/api/v1/courses/2882/enrollments?search=" + accessLogCPF + "&limit=1", accessLogCPF},
		{"nome", "/api/v1/courses/2882/enrollments?search=Fulano+Sintetico&page=2", accessLogName},
		{"e-mail", "/api/v1/courses/2882/enrollments?search=fulano.sintetico%40example.invalid", accessLogEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := serveAccessLog(t, tt.target)

			require.NotEmpty(t, out)
			assert.Contains(t, out, "[GIN]")
			assert.Contains(t, out, "GET")
			assert.Contains(t, out, "200")
			assert.Contains(t, out, "/api/v1/courses/:courseId/enrollments?")
			assert.Contains(t, out, "search=[omitido]")
			assert.NotContains(t, out, tt.secret)
			assert.NotContains(t, out, "fulano")
			assert.NotContains(t, out, "2882")
		})
	}
}

func TestAccessLogger_PaginationParams_AreKept(t *testing.T) {
	out := serveAccessLog(t, "/api/v1/courses/7/enrollments?search="+accessLogCPF+"&limit=1&page=3")

	assert.Contains(t, out, "limit=1")
	assert.Contains(t, out, "page=3")
	assert.Contains(t, out, "search=[omitido]")
	assert.NotContains(t, out, accessLogCPF)
}

func TestAccessLogger_PathParam_LogsRouteTemplate(t *testing.T) {
	out := serveAccessLog(t, "/api/v1/enrollments/user/"+accessLogCPF)

	assert.Contains(t, out, "/api/v1/enrollments/user/:cpf")
	assert.NotContains(t, out, accessLogCPF)
}

func TestAccessLogger_UnmatchedRoute_DoesNotLogRawPath(t *testing.T) {
	out := serveAccessLog(t, "/api/v1/inexistente/"+accessLogCPF+"?q="+accessLogCPF)

	assert.Contains(t, out, "404")
	assert.Contains(t, out, unmatchedRoute)
	assert.NotContains(t, out, accessLogCPF)
}

func TestAccessLogger_NoQuery_LogsPlainRoute(t *testing.T) {
	out := serveAccessLog(t, "/health")

	assert.Contains(t, out, `"/health"`)
	assert.NotContains(t, out, "?")
}
