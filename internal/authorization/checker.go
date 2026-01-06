package authorization

import (
	"context"
	"fmt"
	"time"
)

// Checker provides high-level authorization checking functions
type Checker struct {
	client *Client
}

// NewChecker creates a new authorization checker
func NewChecker(endpoint string, timeout time.Duration) *Checker {
	return &Checker{
		client: NewClient(endpoint, timeout),
	}
}

// CheckAction checks if a principal has permission to perform a specific action on a resource
func (ch *Checker) CheckAction(ctx context.Context, principalCPF string, roles []string, resourceKind string, action string) (bool, error) {
	// Build Cerbos request
	request := &CheckResourcesRequest{
		RequestID: generateRequestID(),
		Principal: Principal{
			ID:            principalCPF,
			Roles:         roles,
			PolicyVersion: "default",
			Attr: map[string]interface{}{
				"cpf": principalCPF,
			},
		},
		Resources: []Resource{
			{
				Resource: ResourceInfo{
					Kind: resourceKind,
					ID:   "resource",
					Attr: map[string]interface{}{},
				},
				Actions: []string{action},
			},
		},
	}

	// Call Cerbos
	response, err := ch.client.CheckResources(ctx, request)
	if err != nil {
		return false, fmt.Errorf("cerbos check failed: %w", err)
	}

	return response.IsAllowed(action), nil
}

// CheckAnyAction checks if a principal has permission to perform ANY of the given actions on a resource
// Returns true if at least one action is allowed
func (ch *Checker) CheckAnyAction(ctx context.Context, principalCPF string, roles []string, resourceKind string, actions []string) (bool, error) {
	if len(actions) == 0 {
		return false, nil
	}

	// Build Cerbos request with all actions
	request := &CheckResourcesRequest{
		RequestID: generateRequestID(),
		Principal: Principal{
			ID:            principalCPF,
			Roles:         roles,
			PolicyVersion: "default",
			Attr: map[string]interface{}{
				"cpf": principalCPF,
			},
		},
		Resources: []Resource{
			{
				Resource: ResourceInfo{
					Kind: resourceKind,
					ID:   "resource",
					Attr: map[string]interface{}{},
				},
				Actions: actions,
			},
		},
	}

	// Call Cerbos
	response, err := ch.client.CheckResources(ctx, request)
	if err != nil {
		return false, fmt.Errorf("cerbos check failed: %w", err)
	}

	// Check if any action is allowed
	return response.IsAnyAllowed(actions), nil
}

// CanAccessResource checks if a principal can perform an action on a specific resource instance
func (ch *Checker) CanAccessResource(ctx context.Context, principalCPF string, roles []string, resourceKind string, resourceID string, action string, resourceAttrs map[string]interface{}) (bool, error) {
	// Build Cerbos request
	request := &CheckResourcesRequest{
		RequestID: generateRequestID(),
		Principal: Principal{
			ID:            principalCPF,
			Roles:         roles,
			PolicyVersion: "default",
			Attr: map[string]interface{}{
				"cpf": principalCPF,
			},
		},
		Resources: []Resource{
			{
				Resource: ResourceInfo{
					Kind: resourceKind,
					ID:   resourceID,
					Attr: resourceAttrs,
				},
				Actions: []string{action},
			},
		},
	}

	// Call Cerbos
	response, err := ch.client.CheckResources(ctx, request)
	if err != nil {
		return false, fmt.Errorf("cerbos check failed: %w", err)
	}

	return response.IsAllowed(action), nil
}
