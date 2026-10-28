package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newCorsRouter(allowed []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CorsMiddlewareWithOrigins(allowed))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Response")
	})
	return router
}

func TestCORS_Preflight_AllowedOrigin(t *testing.T) {
	router := newCorsRouter([]string{"https://example.com"})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORS_ActualRequest_AllowedOrigin(t *testing.T) {
	router := newCorsRouter([]string{"https://example.com"})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "Response", w.Body.String())
}

func TestCORS_DeniedOrigin(t *testing.T) {
	router := newCorsRouter([]string{"https://example.com"})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_Preflight_DeniedOrigin(t *testing.T) {
	router := newCorsRouter([]string{"https://example.com"})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_NoOriginHeader(t *testing.T) {
	router := newCorsRouter([]string{"https://example.com"})

	req := httptest.NewRequest("GET", "/test", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Response", w.Body.String())
}

func TestCORS_DefaultEmptyAllowlist(t *testing.T) {
	t.Setenv(CorsAllowedOriginsEnv, "")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CorsMiddleware())
	router.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_EnvAllowlist(t *testing.T) {
	t.Setenv(CorsAllowedOriginsEnv, " https://a.com/ , https://b.com,*")
	assert.Equal(t, []string{"https://a.com", "https://b.com"}, ParseAllowedOrigins("https://a.com/ , https://b.com,*"))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CorsMiddleware())
	router.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://b.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, "https://b.com", w.Header().Get("Access-Control-Allow-Origin"))
}
