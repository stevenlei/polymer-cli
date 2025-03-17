package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Client represents a Polymer API client
type Client struct {
	APIKey     string
	APIBaseURL string
	HTTPClient *http.Client
	Debug      bool
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ProofStatusResponse represents the status of a proof job
type ProofStatusResponse struct {
	Status string          `json:"status"`
	Proof  json.RawMessage `json:"proof,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// NewClient creates a new Polymer API client
func NewClient(apiKey, apiBaseURL string, debug bool) *Client {
	return &Client{
		APIKey:     apiKey,
		APIBaseURL: apiBaseURL,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		Debug: debug,
	}
}

// RequestProof sends a request to generate a proof for a transaction
func (c *Client) RequestProof(srcChainID uint64, srcBlockNumber uint64, txIndex uint, logIndex uint) (string, error) {
	// Create JSON-RPC request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "log_requestProof",
		Params:  []interface{}{srcChainID, srcBlockNumber, txIndex, logIndex},
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	if c.Debug {
		fmt.Printf("DEBUG: Sending request to %s\n", c.APIBaseURL)
		fmt.Printf("DEBUG: Request body: %s\n", string(reqBody))
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.APIBaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// Send request
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if c.Debug {
		fmt.Printf("DEBUG: Response status: %s\n", resp.Status)
		fmt.Printf("DEBUG: Response body: %s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON-RPC response
	var response JSONRPCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return "", fmt.Errorf("API returned error: %s", response.Error.Message)
	}

	// Get job ID from result
	var jobID string
	switch v := response.Result.(type) {
	case string:
		jobID = v
	case float64:
		jobID = fmt.Sprintf("%.0f", v)
	default:
		return "", fmt.Errorf("unexpected result type: %T", response.Result)
	}

	return jobID, nil
}

// GetProofStatus checks the status of a proof generation job
func (c *Client) GetProofStatus(jobID string) (*ProofStatusResponse, error) {
	// Convert job ID to numeric format
	jobIDNum, err := strconv.ParseFloat(jobID, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	// Create JSON-RPC request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "log_queryProof",
		Params:  []interface{}{jobIDNum},
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if c.Debug {
		fmt.Printf("DEBUG: Sending request to %s\n", c.APIBaseURL)
		fmt.Printf("DEBUG: Request body: %s\n", string(reqBody))
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.APIBaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// Send request
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.Debug {
		fmt.Printf("DEBUG: Response status: %s\n", resp.Status)
		fmt.Printf("DEBUG: Response body: %s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON-RPC response
	var response JSONRPCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("API returned error: %s", response.Error.Message)
	}

	// Parse status response from result
	resultJSON, err := json.Marshal(response.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var statusResponse ProofStatusResponse
	if err := json.Unmarshal(resultJSON, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status response: %w", err)
	}

	return &statusResponse, nil
}

// WaitForProof polls for a proof until it's generated or max attempts is reached
func (c *Client) WaitForProof(jobID string, maxAttempts int, interval time.Duration) (*ProofStatusResponse, error) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if c.Debug {
			fmt.Printf("DEBUG: Polling attempt %d/%d for job %s\n", attempt+1, maxAttempts, jobID)
		}

		status, err := c.GetProofStatus(jobID)
		if err != nil {
			return nil, err
		}

		switch status.Status {
		case "complete", "completed":
			return status, nil
		case "failed":
			return nil, fmt.Errorf("proof generation failed: %s", status.Error)
		case "pending", "processing":
			// Continue polling
			if c.Debug {
				fmt.Printf("DEBUG: Job status: %s, waiting...\n", status.Status)
			}
			time.Sleep(interval)
		default:
			return nil, fmt.Errorf("unknown job status: %s", status.Status)
		}
	}

	return nil, fmt.Errorf("max polling attempts (%d) reached without completion", maxAttempts)
}
