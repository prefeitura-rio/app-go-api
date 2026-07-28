package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRBACMiddleware_Enabled_UsesReal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	real := func(c *gin.Context) {
		c.JSON(http.StatusForbidden, gin.H{"error": "denied"})
		c.Abort()
	}

	r := gin.New()
	r.GET("/protected", rbacMiddleware(true, real), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRBACMiddleware_Disabled_UsesNoOp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	real := func(c *gin.Context) {
		c.JSON(http.StatusForbidden, gin.H{"error": "denied"})
		c.Abort()
	}

	r := gin.New()
	r.GET("/protected", rbacMiddleware(false, real), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRBACMiddleware_IndependentOfAppEnv(t *testing.T) {
	gin.SetMode(gin.TestMode)

	real := func(c *gin.Context) {
		c.JSON(http.StatusForbidden, gin.H{"error": "denied"})
		c.Abort()
	}

	// Even with APP_ENV=prod semantics historically disabling RBAC, the env var
	// alone controls whether the real middleware runs.
	r := gin.New()
	r.GET("/protected", rbacMiddleware(true, real), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusForbidden, w.Code)
}
