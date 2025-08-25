package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os" // New import for os.CreateTemp
	"testing"
	"time" // New import for time.Now()

	"github.com/sire-run/sire/internal/core"
	"github.com/sire-run/sire/internal/mcp/inprocess" // Import inprocess dispatcher
	"github.com/sire-run/sire/internal/mcp/remote"   // Import remote dispatcher types
	"github.com/sire-run/sire/internal/storage"      // New import for storage
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMCPService represents a mock MCP service that can handle RPC calls.
type MockMCPService struct {
	methods map[string]func(params map[string]interface{}) (interface{}, error)
}

// MockDispatcher is a mock implementation of the Dispatcher interface for testing.
type MockDispatcher struct {
	DispatchFunc func(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error)
}

// Dispatch calls the mock DispatchFunc.
func (m *MockDispatcher) Dispatch(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error) {
	if m.DispatchFunc != nil {
		return m.DispatchFunc(ctx, tool, params)
	}
	return nil, nil
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

	engine := core.NewEngine(dispatcherMux, nil) // Pass nil for store for now

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

	execution := &core.Execution{
		ID:         "exec-remote-1",
		WorkflowID: workflow.ID,
		Status:     core.ExecutionStatusRunning,
		StepStates: make(map[string]*core.StepState),
	}

	execResult, err := engine.Execute(context.Background(), execution, workflow, nil)
	require.NoError(t, err)
	assert.Equal(t, core.ExecutionStatusCompleted, execResult.Status)
	assert.Equal(t, core.StepStatusCompleted, execResult.StepStates["add_numbers"].Status)
	assert.Equal(t, float64(30), execResult.StepStates["add_numbers"].Output["sum"])

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

	executionWithError := &core.Execution{
		ID:         "exec-remote-error-1",
		WorkflowID: workflowWithError.ID,
		Status:     core.ExecutionStatusRunning,
		StepStates: make(map[string]*core.StepState),
	}

	execResultWithError, err := engine.Execute(context.Background(), executionWithError, workflowWithError, nil)
	require.Error(t, err)
	assert.Equal(t, core.ExecutionStatusFailed, execResultWithError.Status)
	assert.Equal(t, core.StepStatusFailed, execResultWithError.StepStates["divide_by_zero"].Status)
	assert.Contains(t, execResultWithError.StepStates["divide_by_zero"].Error, "remote tool error (code -32000): division by zero")
}

func TestEngine_ResumeWorkflow(t *testing.T) {
	// Create a temporary BoltDB file
	tmpfile, err := os.CreateTemp("", "test-boltdb-resume-*.db")
	require.NoError(t, err)
	dbPath := tmpfile.Name()
	require.NoError(t, tmpfile.Close()) // Close the file so BoltDB can open it
	defer func() { assert.NoError(t, os.Remove(dbPath)) }()

	store, err := storage.NewBoltDBStore(dbPath)
	require.NoError(t, err)
	defer func() { assert.NoError(t, store.Close()) }()

	// Mock dispatcher that fails on the second step initially
	mockDispatcher := &MockDispatcher{
		DispatchFunc: func(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error) {
			if tool == "sire:local/step2" && !params["allow_step2_success"].(bool) {
				return nil, fmt.Errorf("simulated failure for step2")
			}
			switch tool {
			case "sire:local/step1":
				return map[string]interface{}{"output1": "hello"}, nil
			case "sire:local/step2":
				return map[string]interface{}{"output2": params["output1"].(string) + " world"}, nil
			case "sire:local/step3":
				return map[string]interface{}{"output3": params["output2"].(string) + "!"}, nil
			default:
				return nil, fmt.Errorf("unknown tool: %s", tool)
			}
		},
	}

	// Create a DispatcherMux and register the mock dispatcher
	dispatcherMux := core.NewDispatcherMux()
	dispatcherMux.Register("sire", mockDispatcher)

	// First run: Workflow should fail at step2
	workflow := &core.Workflow{
		ID:   "resume-workflow",
		Name: "Resume Test Workflow",
		Steps: []core.Step{
			{ID: "step1", Tool: "sire:local/step1"},
			{ID: "step2", Tool: "sire:local/step2", Params: map[string]interface{}{"allow_step2_success": false}}, // Fails initially
			{ID: "step3", Tool: "sire:local/step3"},
		},
		Edges: []core.Edge{
			{From: "step1", To: "step2"},
			{From: "step2", To: "step3"},
		},
	}

	engine := core.NewEngine(dispatcherMux, store) // Pass the store

	execution := &core.Execution{
		ID:         "exec-resume-1",
		WorkflowID: workflow.ID,
		Status:     core.ExecutionStatusRunning,
		StepStates: make(map[string]*core.StepState),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Execute the workflow for the first time
	execResult, err := engine.Execute(context.Background(), execution, workflow, nil)
	require.Error(t, err)
	assert.Equal(t, core.ExecutionStatusFailed, execResult.Status)
	assert.Equal(t, core.StepStatusCompleted, execResult.StepStates["step1"].Status)
	assert.Equal(t, core.StepStatusFailed, execResult.StepStates["step2"].Status)
	assert.Contains(t, execResult.StepStates["step2"].Error, "simulated failure for step2")
	assert.Nil(t, execResult.StepStates["step3"]) // step3 should not have run

	// Simulate orchestrator restart: Load execution from DB
	loadedExec, err := store.LoadExecution("exec-resume-1")
	require.NoError(t, err)
	assert.Equal(t, core.ExecutionStatusFailed, loadedExec.Status) // Still failed from previous run
	assert.Equal(t, core.StepStatusCompleted, loadedExec.StepStates["step1"].Status)
	assert.Equal(t, core.StepStatusFailed, loadedExec.StepStates["step2"].Status)

	// Modify the workflow to allow step2 to succeed on resume
	workflow.Steps[1].Params["allow_step2_success"] = true

	// Re-execute the workflow with the loaded execution state
	// The engine should resume from step2
	resumedExecResult, err := engine.Execute(context.Background(), loadedExec, workflow, nil)
	require.NoError(t, err)
	assert.Equal(t, core.ExecutionStatusCompleted, resumedExecResult.Status)
	assert.Equal(t, core.StepStatusCompleted, resumedExecResult.StepStates["step1"].Status)
	assert.Equal(t, core.StepStatusCompleted, resumedExecResult.StepStates["step2"].Status)
	assert.Equal(t, core.StepStatusCompleted, resumedExecResult.StepStates["step3"].Status)
	assert.Equal(t, "hello world!", resumedExecResult.StepStates["step3"].Output["output3"])
}
