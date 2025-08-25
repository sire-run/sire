package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sire-run/sire/internal/core"
	"github.com/sire-run/sire/internal/mcp/inprocess" // Import inprocess dispatcher
	"github.com/sire-run/sire/internal/mcp/remote"    // Import remote dispatcher types
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMCPService represents a mock MCP service that can handle RPC calls.
type MockMCPService struct {
	methods map[string]func(params map[string]interface{}) (interface{}, error)
}

// NewMockMCPService creates a new MockMCPService.
func NewMockMCPService() *MockMCPService {
	return &MockMCPService{
		methods: make(map[string]func(params map[string]interface{}) (interface{}, error)),
	}
}

// RegisterMethod registers a method with the mock service.
func (s *MockMCPService) RegisterMethod(name string, handler func(params map[string]interface{}) (interface{}, error)) {
	s.methods[name] = handler
}

// ServeHTTP implements the http.Handler interface for the mock MCP server.
func (s *MockMCPService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	var req remote.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	var callParams map[string]interface{}
	if req.Params != nil {
		if p, ok := req.Params.(map[string]interface{}); ok {
			callParams = p
		} else {
			http.Error(w, "Bad Request: params must be an object", http.StatusBadRequest)
			return
		}
	} else {
		callParams = make(map[string]interface{}) // Empty map if params is nil
	}

	handler, ok := s.methods[req.Method]
	if !ok {
		resp := remote.JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &remote.JSONRPCError{Code: -32601, Message: "Method not found"},
			ID:      req.ID,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil { // Check error
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	result, err := handler(callParams)
	if err != nil {
		resp := remote.JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &remote.JSONRPCError{Code: -32000, Message: err.Error()},
			ID:      req.ID,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil { // Check error
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var resultBytes []byte
	resultBytes, err = json.Marshal(result)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := remote.JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  json.RawMessage(resultBytes),
		ID:      req.ID,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil { // Check error
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
	}
}

func TestEngine_RemoteToolExecution(t *testing.T) {
	// S7.2.1 Create a mock remote MCP server
	mockService := NewMockMCPService()
	mockService.RegisterMethod("math.add", func(params map[string]interface{}) (interface{}, error) {
		a, okA := params["a"].(float64)
		b, okB := params["b"].(float64)
		if !okA || !okB {
			return nil, fmt.Errorf("invalid parameters")
		}
		return map[string]interface{}{"sum": a + b}, nil
	})

	ts := httptest.NewServer(mockService)
	defer ts.Close()

	// S7.2.2 Write integration tests that use the core.Engine
	// to execute a workflow that calls the mock remote server.

	// Create a DispatcherMux and register both in-process and remote dispatchers
	dispatcherMux := core.NewDispatcherMux()
	dispatcherMux.Register("sire", inprocess.NewInProcessDispatcher())
	dispatcherMux.Register("mcp", remote.NewRemoteDispatcher())

	engine := core.NewEngine(dispatcherMux)

	// Define a workflow that calls the remote math.add tool
	workflow := &core.Workflow{
		ID:   "remote-math-workflow",
		Name: "Remote Math Workflow",
		Steps: []core.Step{
			{
				ID:   "add_numbers",
				Tool: fmt.Sprintf("mcp:%s#math.add", ts.URL), // Use the mock server's URL
				Params: map[string]interface{}{
					"a": 10,
					"b": 20,
				},
			},
		},
		Edges: []core.Edge{},
	}

	execution, err := engine.Execute(context.Background(), workflow, nil)
	require.NoError(t, err)
	assert.Equal(t, "success", execution.Status)
	assert.Equal(t, "success", execution.StepStates["add_numbers"].Status)
	assert.Equal(t, float64(30), execution.StepStates["add_numbers"].Output["sum"])

	// Test a workflow with a remote tool that returns an error
	mockService.RegisterMethod("math.divide", func(params map[string]interface{}) (interface{}, error) {
		numerator, okN := params["numerator"].(float64)
		denominator, okD := params["denominator"].(float64)
		if !okN || !okD {
			return nil, fmt.Errorf("invalid parameters")
		}
		if denominator == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return map[string]interface{}{"result": numerator / denominator}, nil
	})

	workflowWithError := &core.Workflow{
		ID:   "remote-error-workflow",
		Name: "Remote Error Workflow",
		Steps: []core.Step{
			{
				ID:   "divide_by_zero",
				Tool: fmt.Sprintf("mcp:%s#math.divide", ts.URL),
				Params: map[string]interface{}{
					"numerator":   10,
					"denominator": 0,
				},
			},
		},
		Edges: []core.Edge{},
	}

	executionWithError, err := engine.Execute(context.Background(), workflowWithError, nil)
	require.Error(t, err)
	assert.Equal(t, "failed", executionWithError.Status)
	assert.Equal(t, "failed", executionWithError.StepStates["divide_by_zero"].Status)
	assert.Contains(t, executionWithError.StepStates["divide_by_zero"].Error, "remote tool error (code -32000): division by zero")
}
