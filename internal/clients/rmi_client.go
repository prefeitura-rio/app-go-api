package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// GetCNPJOwnerInfo fetches contact information for the owner of a CNPJ
// This is a two-step process:
// 1. GET /v1/legal-entity/{cnpj} to get the list of socios CPFs
// 2. GET /v1/citizen/{cpf} for the first socio to get contact info
// Requires service account authentication token
func (c *RMIClient) GetCNPJOwnerInfo(ctx context.Context, serviceToken string, cnpj string) (*models.CNPJOwnerInfo, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("RMI base URL not configured")
	}

	// Normalize CNPJ (remove formatting)
	normalizedCNPJ := cnpj
	for _, char := range []string{".", "/", "-"} {
		normalizedCNPJ = strings.ReplaceAll(normalizedCNPJ, char, "")
	}

	// Step 1: Get legal entity details to find socios
	legalEntityURL := fmt.Sprintf("%s/v1/legal-entity/%s", c.baseURL, normalizedCNPJ)

	req, err := http.NewRequestWithContext(ctx, "GET", legalEntityURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create legal entity request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", serviceToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute legal entity request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing legal entity response body: %v\n", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("legal entity not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RMI API returned status %d: %s", resp.StatusCode, string(body))
	}

	var legalEntity models.LegalEntityDetails
	if err := json.NewDecoder(resp.Body).Decode(&legalEntity); err != nil {
		return nil, fmt.Errorf("failed to decode legal entity response: %w", err)
	}

	// Get all socio CPFs
	socioCPFs := legalEntity.GetSocioCPFs()

	if len(socioCPFs) == 0 {
		return nil, fmt.Errorf("no socios found for CNPJ")
	}

	// Step 2: Get contact info for the first socio
	// TODO: In the future, we might want to match the socio CPF with the requesting user's CPF
	firstCPF := socioCPFs[0]

	// Normalize CPF (remove formatting)
	normalizedCPF := firstCPF
	for _, char := range []string{".", "-"} {
		normalizedCPF = strings.ReplaceAll(normalizedCPF, char, "")
	}

	citizenURL := fmt.Sprintf("%s/v1/citizen/%s", c.baseURL, normalizedCPF)

	req, err = http.NewRequestWithContext(ctx, "GET", citizenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create citizen request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", serviceToken))
	req.Header.Set("Accept", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute citizen request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing citizen response body: %v\n", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("citizen contact info not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RMI API returned status %d: %s", resp.StatusCode, string(body))
	}

	var citizenInfo models.CitizenContactInfo
	if err := json.NewDecoder(resp.Body).Decode(&citizenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode citizen response: %w", err)
	}

	// Build the final CNPJOwnerInfo
	ownerInfo := &models.CNPJOwnerInfo{
		CNPJ:                normalizedCNPJ,
		EmailPessoaFisica:   citizenInfo.GetEmail(),
		CelularPessoaFisica: citizenInfo.GetCelular(),
	}

	return ownerInfo, nil
}

// GetCitizenByCPF fetches citizen data from RMI API by CPF
// Endpoint: GET /v1/citizen/{cpf}
// Requires service account authentication token
func (c *RMIClient) GetCitizenByCPF(ctx context.Context, serviceToken string, cpf string) (*models.CitizenContactInfo, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("RMI base URL not configured")
	}

	// Normalize CPF (remove formatting)
	normalizedCPF := cpf
	for _, char := range []string{".", "-"} {
		normalizedCPF = strings.ReplaceAll(normalizedCPF, char, "")
	}

	citizenURL := fmt.Sprintf("%s/v1/citizen/%s", c.baseURL, normalizedCPF)

	req, err := http.NewRequestWithContext(ctx, "GET", citizenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create citizen request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", serviceToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute citizen request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing citizen response body: %v\n", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("citizen not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RMI API returned status %d: %s", resp.StatusCode, string(body))
	}

	var citizenInfo models.CitizenContactInfo
	if err := json.NewDecoder(resp.Body).Decode(&citizenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode citizen response: %w", err)
	}

	return &citizenInfo, nil
}
