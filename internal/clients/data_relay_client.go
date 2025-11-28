package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DataRelayClient handles communication with the Data Relay API for email sending
type DataRelayClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewDataRelayClient creates a new Data Relay API client
func NewDataRelayClient(baseURL string, apiKey string, timeout time.Duration) *DataRelayClient {
	return &DataRelayClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// EmailRequest represents the request body for sending an email via Data Relay
type EmailRequest struct {
	ToAddresses  []string `json:"to_addresses"`
	CCAddresses  []string `json:"cc_addresses,omitempty"`
	BCCAddresses []string `json:"bcc_addresses,omitempty"`
	ReplyTo      []string `json:"reply_to,omitempty"`
	Subject      string   `json:"subject"`
	Body         string   `json:"body"`
	IsHTMLBody   bool     `json:"is_html_body"`
}

// SendEmail sends an email via the Data Relay API
func (c *DataRelayClient) SendEmail(ctx context.Context, req *EmailRequest) error {
	if c.baseURL == "" {
		return fmt.Errorf("data relay base URL not configured")
	}

	if c.apiKey == "" {
		return fmt.Errorf("data relay API key not configured")
	}

	// Marshal request body
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/data/mailman", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", c.apiKey)

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing Data Relay response body: %v\n", err)
		}
	}()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("data relay API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
