package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTimeout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TimeoutMiddleware(1 * time.Second))

	router.GET("/test", func(c *gin.Context) {
		// Fast handler
		time.Sleep(100 * time.Millisecond)
		c.String(http.StatusOK, "Success")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Success", w.Body.String())
}

func TestTimeout_Exceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TimeoutMiddleware(100 * time.Millisecond))

	router.GET("/test", func(c *gin.Context) {
		// Slow handler that exceeds timeout
		time.Sleep(500 * time.Millisecond)
		c.String(http.StatusOK, "This should not be returned")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Contains(t, w.Body.String(), "timeout")
}

func TestTimeout_ChecksContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TimeoutMiddleware(200 * time.Millisecond))

	router.GET("/test", func(c *gin.Context) {
		// Handler that checks context cancellation
		select {
		case <-time.After(1 * time.Second):
			c.String(http.StatusOK, "Completed")
		case <-c.Request.Context().Done():
			c.String(http.StatusRequestTimeout, "Cancelled")
			return
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// The timeout should trigger (either 408 or 504 depending on when context was checked)
	assert.True(t, w.Code == http.StatusRequestTimeout || w.Code == http.StatusGatewayTimeout,
		"Expected timeout status code (408 or 504), got %d", w.Code)
}
