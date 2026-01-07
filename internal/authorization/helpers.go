package authorization

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prefeitura-rio/app-go-api/internal/middlewares"
)

// CNPJOwnershipChecker defines the interface for checking CNPJ ownership
type CNPJOwnershipChecker interface {
	CheckCNPJOwnership(ctx context.Context, authToken string, cpf string, cnpj string) (bool, error)
}

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

// RequireCNPJOwnershipOrAnyPermission checks if the user owns the CNPJ OR has any of the specified permissions
// This is for resources identified by CNPJ instead of CPF
// Returns nil if authorized, returns error if not authorized
func RequireCNPJOwnershipOrAnyPermission(
	c *gin.Context,
	checker *Checker,
	ownershipChecker CNPJOwnershipChecker,
	cnpj string,
	resourceKind string,
	actions []string,
) error {
	// Extract user CPF from context (set by ExtractUserContext middleware)
	userCPF := middlewares.GetUserCPF(c)
	if userCPF == "" {
		return fmt.Errorf("user CPF not found in context")
	}

	// Extract auth token from headers
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		return fmt.Errorf("authorization token not found")
	}

	// Check if user owns the CNPJ
	isOwner, err := ownershipChecker.CheckCNPJOwnership(c.Request.Context(), authToken, userCPF, cnpj)
	if err != nil {
		return fmt.Errorf("failed to check CNPJ ownership: %w", err)
	}

	if isOwner {
		return nil // Owner is always authorized
	}

	// If checker is nil (Cerbos disabled), deny access for non-owners
	if checker == nil {
		return fmt.Errorf("not authorized: you can only access resources for CNPJs you own")
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
