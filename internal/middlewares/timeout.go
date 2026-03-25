package middlewares

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware adds a timeout to all requests to prevent long-running
// queries from accumulating under high load.
//
// This middleware sets a timeout context for the request. Handlers should
// check c.Request.Context().Done() to respect the timeout.
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context with timeout context
		c.Request = c.Request.WithContext(ctx)

		// Continue with the request
		// Handlers are responsible for checking ctx.Done() to respect timeout
		c.Next()
	}
}
