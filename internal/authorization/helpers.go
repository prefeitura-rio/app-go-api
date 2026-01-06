package authorization

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
)

// RequireOwnershipOrAnyPermission checks if the user is the owner OR has any of the specified permissions
// Returns nil if authorized, returns error if not authorized
func RequireOwnershipOrAnyPermission(c *gin.Context, checker *Checker, ownerCPF string, resourceKind string, actions []string) error {
	// Extract user CPF from context (set by ExtractUserContext middleware)
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		return fmt.Errorf("user CPF not found in context")
	}

	// Check if user is the owner
	if userCPF == ownerCPF {
		return nil // Owner is always authorized
	}

	// If checker is nil (Cerbos disabled), deny access for non-owners
	if checker == nil {
		return fmt.Errorf("not authorized: you can only access your own resources")
	}

	// User is not the owner, check if they have any of the required permissions
	userRole := middlewares.GetUserRole(c)
	roles := []string{}
	if userRole != "" {
		roles = append(roles, userRole)
	}

	// Check if user has any of the allowed actions
	hasPermission, err := checker.CheckAnyAction(c.Request.Context(), userCPF, roles, resourceKind, actions)
	if err != nil {
		return fmt.Errorf("authorization check failed: %w", err)
	}

	if !hasPermission {
		return fmt.Errorf("not authorized: you don't have permission to access this resource")
	}

	return nil
}

// RequireAnyPermission checks if the user has any of the specified permissions
// Does NOT check ownership, only permissions
func RequireAnyPermission(c *gin.Context, checker *Checker, resourceKind string, actions []string) error {
	// If checker is nil (Cerbos disabled), deny access
	if checker == nil {
		return fmt.Errorf("authorization is disabled")
	}

	// Extract user CPF from context
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		return fmt.Errorf("user CPF not found in context")
	}

	// Extract user role
	userRole := middlewares.GetUserRole(c)
	roles := []string{}
	if userRole != "" {
		roles = append(roles, userRole)
	}

	// Check if user has any of the allowed actions
	hasPermission, err := checker.CheckAnyAction(c.Request.Context(), userCPF, roles, resourceKind, actions)
	if err != nil {
		return fmt.Errorf("authorization check failed: %w", err)
	}

	if !hasPermission {
		return fmt.Errorf("not authorized: you don't have permission to perform this action")
	}

	return nil
}

// RequirePermission checks if the user has a specific permission
// Wrapper around RequireAnyPermission for single permission check
func RequirePermission(c *gin.Context, checker *Checker, resourceKind string, action string) error {
	return RequireAnyPermission(c, checker, resourceKind, []string{action})
}
