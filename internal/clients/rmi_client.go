package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/sony/gobreaker"

	"github.com/prefeitura-rio/app-go-api/internal/models"
)

var (
	ErrCNPJNotFound     = errors.New("CNPJ não encontrado")
	ErrCNPJAccessDenied = errors.New("acesso negado ao CNPJ")
	ErrInvalidCNPJ      = errors.New("CNPJ inválido")

	// ErrCitizenNotFound is returned by GetCitizenByCPF when the RMI answers 404
	// for the CPF: an expected business result, not an infrastructure failure.
	ErrCitizenNotFound = errors.New("citizen not found")
)

// RMIStatusError is the error for a non-OK RMI response. Snippet is the
// sanitized excerpt of the response body (see sanitizeRMIErrorBody), empty when
// nothing of it is safe to keep. Callers use StatusCode to tell a rejected
// input (4xx) from an RMI failure (5xx) without parsing the message.
type RMIStatusError struct {
	StatusCode int
	Snippet    string
}

func (e *RMIStatusError) Error() string {
	if e.Snippet != "" {
		return fmt.Sprintf("RMI API returned status %d: %s", e.StatusCode, e.Snippet)
	}
	return fmt.Sprintf("RMI API returned status %d", e.StatusCode)
}

const (
	maxLegalEntities = 1000 // Safety limit to prevent infinite loops
)

// RMIClient handles communication with the RMI API.
// Requests go through a circuit breaker and are retried on transient failures;
// GetCPFSecretarias results are cached per CPF for a short TTL.
type RMIClient struct {
	baseURL         string
	httpClient      *http.Client
	name            string
	retryBaseDelay  time.Duration
	retryMaxDelay   time.Duration
	breaker         *gobreaker.CircuitBreaker
	secretariaCache *secretariaCache
}

// NewRMIClient creates a new RMI API client with the specified timeout per request
func NewRMIClient(baseURL string, timeout time.Duration, opts ...RMIClientOption) *RMIClient {
	o := defaultRMIClientOptions()
	for _, opt := range opts {
		opt(&o)
	}

	rmiCircuitState.WithLabelValues(o.name).Set(circuitStateValue(gobreaker.StateClosed))
	rmiCircuitRejectionsTotal.WithLabelValues(o.name)
	rmiSecretariaCacheTotal.WithLabelValues(o.name, "hit")
	rmiSecretariaCacheTotal.WithLabelValues(o.name, "miss")

	return &RMIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		name:            o.name,
		retryBaseDelay:  o.retryBaseDelay,
		retryMaxDelay:   o.retryMaxDelay,
		breaker:         newRMICircuitBreaker(o),
		secretariaCache: newSecretariaCache(o.secretariaCacheTTL, defaultSecretariaCacheEntries),
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

	resp, err := c.do(req)
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
		return nil, models.PaginationResponse{}, rmiStatusError(resp)
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

	resp, err := c.do(req)
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
		return nil, rmiStatusError(resp)
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

	resp, err := c.do(req)
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
		return nil, rmiStatusError(resp)
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

	resp, err = c.do(req)
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
		return nil, rmiStatusError(resp)
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

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute citizen request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing citizen response body: %v\n", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrCitizenNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, rmiStatusError(resp)
	}

	var citizenInfo models.CitizenContactInfo
	if err := json.NewDecoder(resp.Body).Decode(&citizenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode citizen response: %w", err)
	}

	return &citizenInfo, nil
}

// GetLegalEntityByCNPJ fetches complete legal entity data from RMI API by CNPJ
// Endpoint: GET /v1/legal-entity/{cnpj}
// Requires service account authentication token
func (c *RMIClient) GetLegalEntityByCNPJ(ctx context.Context, serviceToken string, cnpj string) (*models.LegalEntityFull, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("RMI base URL not configured")
	}

	normalizedCNPJ, err := NormalizeCNPJ(cnpj)
	if err != nil {
		return nil, err
	}

	legalEntityURL := fmt.Sprintf("%s/v1/legal-entity/%s", c.baseURL, normalizedCNPJ)

	req, err := http.NewRequestWithContext(ctx, "GET", legalEntityURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create legal entity request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", serviceToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute legal entity request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing legal entity response body: %v\n", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrCNPJNotFound
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrCNPJAccessDenied
	}

	if resp.StatusCode != http.StatusOK {
		return nil, rmiStatusError(resp)
	}

	var legalEntity models.LegalEntityFull
	if err := json.NewDecoder(resp.Body).Decode(&legalEntity); err != nil {
		return nil, fmt.Errorf("failed to decode legal entity response: %w", err)
	}

	return &legalEntity, nil
}

// rmiErrorBodyDrainLimit caps how much of an error response body is read and
// discarded before the connection is released.
const rmiErrorBodyDrainLimit = 64 << 10

const rmiErrorSnippetMaxRunes = 120

// rmiErrorMessageFields are the only fields of a JSON error body kept in the
// error message; every other field may carry citizen data and is dropped.
var rmiErrorMessageFields = []string{"error", "message"}

var (
	// Runs of 6+ digits with optional single separators: CPF, CNPJ, phone
	// numbers, formatted or not.
	rmiDigitRunPattern = regexp.MustCompile(`\d(?:[.\-/ ]?\d){5,}`)
	rmiEmailPattern    = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
)

// rmiStatusError builds the error for a non-OK RMI response. Callers log these
// errors and the RMI may echo citizen data in error bodies, so only a short,
// sanitized excerpt of the body goes into the message (see sanitizeRMIErrorBody).
func rmiStatusError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, rmiErrorBodyDrainLimit))
	return &RMIStatusError{StatusCode: resp.StatusCode, Snippet: sanitizeRMIErrorBody(body)}
}

// sanitizeRMIErrorBody returns a single-line excerpt of an RMI error body that
// is safe to log, or "" when nothing of it should be kept. Only the string
// fields "error" and "message" of a complete JSON object are kept, re-encoded
// as a JSON object, with digit runs that look like CPF, CNPJ or phone numbers
// and e-mail addresses replaced by "[omitido]" and each value truncated. Any
// other body (other JSON fields, arrays, scalars, JSON cut at the read limit,
// HTML, plain text) is dropped, since it can carry citizen data verbatim.
func sanitizeRMIErrorBody(body []byte) string {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return ""
	}

	if !json.Valid(trimmed) {
		return ""
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return ""
	}
	kept := map[string]string{}
	for _, name := range rmiErrorMessageFields {
		var value string
		if err := json.Unmarshal(fields[name], &value); err != nil {
			continue
		}
		if value = sanitizeRMIErrorText(value); value != "" {
			kept[name] = value
		}
	}
	if len(kept) == 0 {
		return ""
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(kept); err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}

func sanitizeRMIErrorText(text string) string {
	text = rmiEmailPattern.ReplaceAllString(text, "[omitido]")
	text = rmiDigitRunPattern.ReplaceAllString(text, "[omitido]")
	text = strings.Join(strings.Fields(text), " ")
	if runes := []rune(text); len(runes) > rmiErrorSnippetMaxRunes {
		text = string(runes[:rmiErrorSnippetMaxRunes]) + "…"
	}
	return text
}

// GetCPFSecretarias fetches the list of cd_uas linked to a CPF from the app-rmi CPF-Secretaria registry.
// Endpoint: GET /cpf-secretaria/{cpf}
// Requires service account authentication token.
// Returns an empty slice when the CPF has no secretaria mappings.
// Successful results (including "no mapping") are cached per CPF; errors are not.
func (c *RMIClient) GetCPFSecretarias(ctx context.Context, serviceToken string, cpf string) ([]string, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("RMI base URL not configured")
	}

	if cdUAs, ok := c.secretariaCache.get(cpf); ok {
		rmiSecretariaCacheTotal.WithLabelValues(c.name, "hit").Inc()
		return cdUAs, nil
	}
	rmiSecretariaCacheTotal.WithLabelValues(c.name, "miss").Inc()

	cdUAs, err := c.fetchCPFSecretarias(ctx, serviceToken, cpf)
	if err != nil {
		return nil, err
	}
	c.secretariaCache.set(cpf, cdUAs)
	return cdUAs, nil
}

func (c *RMIClient) fetchCPFSecretarias(ctx context.Context, serviceToken string, cpf string) ([]string, error) {
	url := fmt.Sprintf("%s/v1/cpf-secretaria/%s", c.baseURL, cpf)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", serviceToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing GetCPFSecretarias response body: %v\n", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return []string{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		// The body is drained (bounded) for connection reuse but never put in the
		// error: callers log it and the RMI may echo citizen data in error bodies.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, rmiErrorBodyDrainLimit))
		return nil, fmt.Errorf("RMI API returned status %d", resp.StatusCode)
	}

	var response struct {
		CPF   string   `json:"cpf"`
		CdUAs []string `json:"cd_uas"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.CdUAs == nil {
		return []string{}, nil
	}
	return response.CdUAs, nil
}

// NormalizeCNPJ removes formatting and validates that the CNPJ has 14 digits
func NormalizeCNPJ(cnpj string) (string, error) {
	var digits strings.Builder
	for _, r := range cnpj {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}

	normalized := digits.String()
	if len(normalized) != 14 {
		return "", ErrInvalidCNPJ
	}

	return normalized, nil
}
