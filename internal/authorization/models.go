package authorization

import "time"

// CheckResourcesRequest represents a Cerbos check resources request
type CheckResourcesRequest struct {
	RequestID string     `json:"requestId"`
	Principal Principal  `json:"principal"`
	Resources []Resource `json:"resources"`
}

// Principal represents the user/entity making the request
type Principal struct {
	ID            string                 `json:"id"`
	Roles         []string               `json:"roles"`
	PolicyVersion string                 `json:"policyVersion"`
	Attr          map[string]interface{} `json:"attr"`
}

// Resource represents a resource being accessed
type Resource struct {
	Resource ResourceInfo `json:"resource"`
	Actions  []string     `json:"actions"`
}

// ResourceInfo contains resource metadata
type ResourceInfo struct {
	Kind string                 `json:"kind"`
	ID   string                 `json:"id"`
	Attr map[string]interface{} `json:"attr"`
}

// CheckResourcesResponse represents a Cerbos check resources response
type CheckResourcesResponse struct {
	RequestID string                   `json:"requestId"`
	Results   []ResourceActionResponse `json:"results"`
}

// ResourceActionResponse represents the response for a specific resource
type ResourceActionResponse struct {
	Resource ResourceInfo      `json:"resource"`
	Actions  map[string]string `json:"actions"`
}

// IsAllowed checks if an action is allowed based on the response
func (r *CheckResourcesResponse) IsAllowed(action string) bool {
	if len(r.Results) == 0 {
		return false
	}

	decision, exists := r.Results[0].Actions[action]
	if !exists {
		return false
	}

	return decision == "EFFECT_ALLOW"
}

// IsAnyAllowed checks if any of the given actions is allowed
func (r *CheckResourcesResponse) IsAnyAllowed(actions []string) bool {
	if len(r.Results) == 0 {
		return false
	}

	for _, action := range actions {
		decision, exists := r.Results[0].Actions[action]
		if exists && decision == "EFFECT_ALLOW" {
			return true
		}
	}

	return false
}

// GetDecision returns the decision for a specific action
func (r *CheckResourcesResponse) GetDecision(action string) string {
	if len(r.Results) == 0 {
		return ""
	}

	decision, exists := r.Results[0].Actions[action]
	if !exists {
		return ""
	}

	return decision
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405.000000")
}
