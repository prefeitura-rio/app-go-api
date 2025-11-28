package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

const (
	maxLegalEntities = 1000 // Safety limit to prevent infinite loops
)

// RMIClient handles communication with the RMI API
type RMIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRMIClient creates a new RMI API client with the specified timeout per request
func NewRMIClient(baseURL string, timeout time.Duration) *RMIClient {
	return &RMIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetUserLegalEntities fetches ALL legal entities for a CPF (handles pagination)
// Returns all CNPJs with their CNAEs
// Implements safety limit of maxLegalEntities to prevent infinite loops
func (c *RMIClient) GetUserLegalEntities(ctx context.Context, authToken string, cpf string) ([]models.LegalEntity, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("RMI base URL not configured")
	}

	allEntities := []models.LegalEntity{}
	page := 1

	for {
		// Check if we've hit the safety limit
		if len(allEntities) >= maxLegalEntities {
			return nil, fmt.Errorf("exceeded maximum legal entities limit (%d)", maxLegalEntities)
		}

		// Fetch page
		entities, pagination, err := c.fetchPage(ctx, authToken, cpf, page)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}

		allEntities = append(allEntities, entities...)

		// Check if we have more pages
		if page >= pagination.TotalPages {
			break
		}

		page++
	}

	return allEntities, nil
}

// fetchPage fetches a single page of legal entities
func (c *RMIClient) fetchPage(ctx context.Context, authToken string, cpf string, page int) ([]models.LegalEntity, models.PaginationResponse, error) {
	url := fmt.Sprintf("%s/v1/citizen/%s/legal-entities?page=%d", c.baseURL, cpf, page)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, models.PaginationResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Forward the Authorization header
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, models.PaginationResponse{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing RMI response body: %v\n", err)
		}
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, models.PaginationResponse{}, fmt.Errorf("RMI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response models.LegalEntitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, models.PaginationResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, response.Pagination, nil
}

// GetOrgao fetches orgao (department) details from RMI API
// Endpoint: GET /v1/departments/{orgao_id}
// No authentication required for this endpoint
func (c *RMIClient) GetOrgao(ctx context.Context, orgaoID string) (*models.Orgao, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("RMI base URL not configured")
	}

	url := fmt.Sprintf("%s/v1/departments/%s", c.baseURL, orgaoID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing RMI response body: %v\n", err)
		}
	}()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RMI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response - note: API returns the object directly, not wrapped in "data"
	var orgao models.Orgao
	if err := json.NewDecoder(resp.Body).Decode(&orgao); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &orgao, nil
}
