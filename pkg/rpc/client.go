package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"golang.org/x/crypto/sha3"
)

// RPCClient represents a JSON-RPC client for Ethereum
type RPCClient struct {
	URL        string
	HTTPClient *http.Client
	Debug      bool
}

// NewRPCClient creates a new Ethereum RPC client
func NewRPCClient(url string, debug bool) *RPCClient {
	return &RPCClient{
		URL:        url,
		HTTPClient: &http.Client{},
		Debug:      debug,
	}
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Transaction represents an Ethereum transaction
type Transaction struct {
	Hash        string `json:"hash"`
	BlockNumber string `json:"blockNumber"`
	BlockHash   string `json:"blockHash"`
	From        string `json:"from"`
	To          string `json:"to"`
	ChainID     string `json:"chainId"`
}

// TransactionReceipt represents an Ethereum transaction receipt
type TransactionReceipt struct {
	TransactionHash  string `json:"transactionHash"`
	TransactionIndex string `json:"transactionIndex"`
	BlockNumber      string `json:"blockNumber"`
	BlockHash        string `json:"blockHash"`
	Status           string `json:"status"`
	Logs             []Log  `json:"logs"`
}

// Log represents a log entry in a transaction receipt
type Log struct {
	LogIndex         string   `json:"logIndex"`
	TransactionIndex string   `json:"transactionIndex"`
	TransactionHash  string   `json:"transactionHash"`
	BlockHash        string   `json:"blockHash"`
	BlockNumber      string   `json:"blockNumber"`
	Address          string   `json:"address"`
	Data             string   `json:"data"`
	Topics           []string `json:"topics"`
}

// GetTransaction fetches transaction information by hash
func (c *RPCClient) GetTransaction(txHash string) (*Transaction, error) {
	// Ensure the hash is prefixed with 0x
	if !strings.HasPrefix(txHash, "0x") {
		txHash = "0x" + txHash
	}

	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_getTransactionByHash",
		Params:  []interface{}{txHash},
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if c.Debug {
		fmt.Printf("DEBUG: Sending RPC request to %s\n", c.URL)
		fmt.Printf("DEBUG: Request body: %s\n", string(reqBody))
	}

	resp, err := c.HTTPClient.Post(c.URL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.Debug {
		fmt.Printf("DEBUG: Response status: %s\n", resp.Status)
		fmt.Printf("DEBUG: Response body: %s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RPC request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response JSONRPCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC returned error: %s", response.Error.Message)
	}

	var tx Transaction
	if err := json.Unmarshal(response.Result, &tx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &tx, nil
}

// GetTransactionReceipt fetches the transaction receipt
func (c *RPCClient) GetTransactionReceipt(txHash string) (*TransactionReceipt, error) {
	// Ensure the hash is prefixed with 0x
	if !strings.HasPrefix(txHash, "0x") {
		txHash = "0x" + txHash
	}

	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_getTransactionReceipt",
		Params:  []interface{}{txHash},
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if c.Debug {
		fmt.Printf("DEBUG: Sending RPC request to %s\n", c.URL)
		fmt.Printf("DEBUG: Request body: %s\n", string(reqBody))
	}

	resp, err := c.HTTPClient.Post(c.URL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if c.Debug {
		fmt.Printf("DEBUG: Response status: %s\n", resp.Status)
		fmt.Printf("DEBUG: Response body: %s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RPC request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response JSONRPCResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC returned error: %s", response.Error.Message)
	}

	var receipt TransactionReceipt
	if err := json.Unmarshal(response.Result, &receipt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal receipt: %w", err)
	}

	return &receipt, nil
}

// GetEventSignatureHash calculates the Keccak256 hash of an event signature
func (c *RPCClient) GetEventSignatureHash(eventSignature string) (string, error) {
	// Ethereum uses Keccak-256 for event signatures
	hasher := sha3.NewLegacyKeccak256()

	// Write the event signature to the hasher
	_, err := hasher.Write([]byte(eventSignature))
	if err != nil {
		return "", fmt.Errorf("failed to hash event signature: %w", err)
	}

	// Get the hash result
	hash := hasher.Sum(nil)

	// Format as 0x-prefixed hex string
	hexHash := "0x" + hex.EncodeToString(hash)

	return hexHash, nil
}

// HexToUint64 converts a hexadecimal string to uint64
func HexToUint64(hex string) (uint64, error) {
	// If the hex is "0x0", just return 0
	if hex == "0x0" {
		return 0, nil
	}

	// Remove "0x" prefix if present
	if len(hex) >= 2 && hex[0:2] == "0x" {
		hex = hex[2:]
	}

	// Parse hex as a big integer
	value := new(big.Int)
	value, ok := value.SetString(hex, 16)
	if !ok {
		return 0, fmt.Errorf("invalid hex value: %s", hex)
	}

	// Check if value fits in uint64
	if !value.IsUint64() {
		return 0, fmt.Errorf("hex value too large for uint64: %s", hex)
	}

	return value.Uint64(), nil
}
