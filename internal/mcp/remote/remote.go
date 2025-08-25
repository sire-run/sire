package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sire-run/sire/internal/core"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  interface{}   `json:"params"`
	ID      int           `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *JSONRPCError    `json:"error,omitempty"`
	ID      int              `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error object.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RemoteDispatcher dispatches tool executions to remote MCP servers.
type RemoteDispatcher struct {
	client *http.Client
}

// NewRemoteDispatcher creates a new RemoteDispatcher.
func NewRemoteDispatcher() *RemoteDispatcher {
	return &RemoteDispatcher{
		client: &http.Client{
			Timeout: 30 * time.Second, // Default timeout
		},
	}
}

// Dispatch dispatches a tool execution to a remote MCP server.
// The toolURI format is mcp:http://host/rpc#service.method
func (d *RemoteDispatcher) Dispatch(ctx context.Context, toolURI string, params map[string]interface{}) (map[string]interface{}, error) {
	// Split the toolURI into the MCP scheme part and the actual RPC URL + fragment
	parts := strings.SplitN(toolURI, ":", 2)
	if len(parts) != 2 || parts[0] != "mcp" {
		return nil, fmt.Errorf("invalid mcp tool URI format: %s. Expected mcp:http://host/rpc#service.method", toolURI)
	}
	rpcURLWithFragment := parts[1] // e.g., http://host/rpc#service.method

	// Parse the RPC URL and fragment
	u, err := url.Parse(rpcURLWithFragment)
	if err != nil {
		return nil, fmt.Errorf("invalid RPC URL in tool URI: %w", err)
	}

	// Check for missing scheme or host in the RPC URL
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("missing scheme or host in RPC URL: %s", rpcURLWithFragment)
	}

	// The actual HTTP URL for the RPC call
	httpURL := u.Scheme + "://" + u.Host + u.Path
	toolName := u.Fragment

	if httpURL == "" {
		return nil, fmt.Errorf("missing HTTP URL in tool URI: %s", toolURI)
	}
	if toolName == "" {
		return nil, fmt.Errorf("missing tool name (service.method) in tool URI: %s", toolURI)
	}


	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  toolName,
		Params:  params,
		ID:      1, // Request ID, can be dynamic
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", httpURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request to %s: %w", httpURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("remote server returned non-OK status: %d, body: %s", resp.StatusCode, string(respBodyBytes))
	}

	var rpcResp JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON-RPC response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("remote tool error (code %d): %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	var result map[string]interface{}
	if rpcResp.Result != nil {
		if err := json.Unmarshal(rpcResp.Result, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return result, nil
}

// Ensure RemoteDispatcher implements core.Dispatcher
var _ core.Dispatcher = (*RemoteDispatcher)(nil)
