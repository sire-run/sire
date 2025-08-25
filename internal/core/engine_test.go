package core

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

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

// MockStore is a mock implementation of the Store interface for testing.
type MockStore struct {
	Executions map[string]*Execution
}

func (m *MockStore) SaveExecution(execution *Execution) error {
	if m.Executions == nil {
		m.Executions = make(map[string]*Execution)
	}
	m.Executions[execution.ID] = execution
	return nil
}

func (m *MockStore) LoadExecution(id string) (*Execution, error) {
	if m.Executions == nil {
		return nil, fmt.Errorf("store is empty")
	}
	exec, ok := m.Executions[id]
	if !ok {
		return nil, fmt.Errorf("execution with ID %s not found", id)
	}
	return exec, nil
}

func (m *MockStore) ListPendingExecutions() ([]*Execution, error) {
	var pending []*Execution
	if m.Executions == nil {
		return pending, nil
	}
	for _, exec := range m.Executions {
		if exec.Status == ExecutionStatusRunning || exec.Status == ExecutionStatusRetrying {
			pending = append(pending, exec)
		}
	}
	return pending, nil
}

func TestEngine_Execute_LinearWorkflow(t *testing.T) {
	// 1. Setup
	dispatcher := &MockDispatcher{
		DispatchFunc: func(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error) {
			if tool == "sire:local/node1" {
				return map[string]interface{}{"node1_output": "hello"}, nil
			}
			if tool == "sire:local/node2" {
				return map[string]interface{}{"node2_output": params["node1_output"].(string) + " world"}, nil
			}
			return nil, fmt.Errorf("unknown tool: %s", tool)
		},
	}

	// Create a dummy store for the engine
	dummyStore := &MockStore{}

	engine := NewEngine(dispatcher, dummyStore)

	workflow := &Workflow{
		ID:   "wf-1",
		Name: "Test Workflow",
		Steps: []Step{
			{ID: "node-1", Tool: "sire:local/node1"},
			{ID: "node-2", Tool: "sire:local/node2"},
		},
		Edges: []Edge{
			{From: "node-1", To: "node-2"},
		},
	}

	execution := &Execution{
		ID:         "exec-1",
		WorkflowID: workflow.ID,
		Status:     ExecutionStatusRunning,
		StepStates: make(map[string]*StepState),
	}

	// 2. Execute
	execResult, err := engine.Execute(context.Background(), execution, workflow, nil)
	// 3. Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if execResult.Status != ExecutionStatusCompleted {
		t.Errorf("expected status %q, got %q", ExecutionStatusCompleted, execResult.Status)
	}
	if execResult.StepStates["node-1"].Status != StepStatusCompleted {
		t.Errorf("expected status %q, got %q", StepStatusCompleted, execResult.StepStates["node-1"].Status)
	}
	if execResult.StepStates["node-2"].Status != StepStatusCompleted {
		t.Errorf("expected status %q, got %q", StepStatusCompleted, execResult.StepStates["node-2"].Status)
	}
	if execResult.StepStates["node-2"].Output["node2_output"] != "hello world" {
		t.Errorf("expected output %q, got %q", "hello world", execResult.StepStates["node-2"].Output["node2_output"])
	}
}

func TestEngine_Execute_FailingWorkflow(t *testing.T) {
	// 1. Setup
	dispatcher := &MockDispatcher{
		DispatchFunc: func(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error) {
			return nil, fmt.Errorf("something went wrong")
		},
	}

	// Create a dummy store for the engine
	dummyStore := &MockStore{}

	engine := NewEngine(dispatcher, dummyStore)

	workflow := &Workflow{
		ID:   "wf-1",
		Name: "Test Workflow",
		Steps: []Step{
			{ID: "node-1", Tool: "sire:local/failing-node"},
		},
	}

	execution := &Execution{
		ID:         "exec-failing-1",
		WorkflowID: workflow.ID,
		Status:     ExecutionStatusRunning,
		StepStates: make(map[string]*StepState),
	}

	// 2. Execute
	execResult, err := engine.Execute(context.Background(), execution, workflow, nil)

	// 3. Assert
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if execResult.Status != ExecutionStatusFailed {
		t.Errorf("expected status %q, got %q", ExecutionStatusFailed, execResult.Status)
	}
	if execResult.StepStates["node-1"].Status != StepStatusFailed {
		t.Errorf("expected status %q, got %q", StepStatusFailed, execResult.StepStates["node-1"].Status)
	}
	if execResult.StepStates["node-1"].Error != "something went wrong" {
		t.Errorf("expected error %q, got %q", "something went wrong", execResult.StepStates["node-1"].Error)
	}
}

func TestTopologicalSort(t *testing.T) {
	t.Run("simple linear workflow", func(t *testing.T) {
		steps := map[string]Step{
			"node-1": {ID: "node-1"},
			"node-2": {ID: "node-2"},
			"node-3": {ID: "node-3"},
		}
		edges := []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-2", To: "node-3"},
		}

		sorted, err := topologicalSort(steps, edges)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sorted) != 3 || sorted[0] != "node-1" || sorted[1] != "node-2" || sorted[2] != "node-3" {
			t.Errorf("expected sorted %v, got %v", []string{"node-1", "node-2", "node-3"}, sorted)
		}
	})

	t.Run("workflow with a branch", func(t *testing.T) {
		steps := map[string]Step{
			"node-1": {ID: "node-1"},
			"node-2": {ID: "node-2"},
			"node-3": {ID: "node-3"},
			"node-4": {ID: "node-4"},
		}
		edges := []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-1", To: "node-3"},
			{From: "node-2", To: "node-4"},
			{From: "node-3", To: "node-4"},
		}

		sorted, err := topologicalSort(steps, edges)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// The exact order of node-2 and node-3 can vary, so we check for both possibilities.
		if !(sorted[1] == "node-2" && sorted[2] == "node-3") && !(sorted[1] == "node-3" && sorted[2] == "node-2") { //nolint:staticcheck // Ignoring De Morgan's law suggestion as current form is clear and linter is overly aggressive
			t.Errorf("expected sorted order of node-2 and node-3 to be flexible, got %v", sorted)
		}
		if sorted[0] != "node-1" {
			t.Errorf("expected first node to be %q, got %q", "node-1", sorted[0])
		}
		if sorted[3] != "node-4" {
			t.Errorf("expected last node to be %q, got %q", "node-4", sorted[3])
		}
	})

	t.Run("workflow with a cycle", func(t *testing.T) {
		steps := map[string]Step{
			"node-1": {ID: "node-1"},
			"node-2": {ID: "node-2"},
			"node-3": {ID: "node-3"},
		}
		edges := []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-2", To: "node-3"},
			{From: "node-3", To: "node-1"},
		}

		_, err := topologicalSort(steps, edges)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}
		if !strings.Contains(err.Error(), "workflow has a cycle") {
			t.Errorf("expected error to contain %q, got %q", "workflow has a cycle", err.Error())
		}
	})
}

func TestEngine_Execute_BranchingWorkflow(t *testing.T) {
	// 1. Setup
	dispatcher := &MockDispatcher{
		DispatchFunc: func(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error) {
			switch tool {
			case "sire:local/node1":
				return map[string]interface{}{"node1_output": "hello"}, nil
			case "sire:local/node2":
				return map[string]interface{}{"node2_output": params["node1_output"].(string) + " from node2"}, nil
			case "sire:local/node3":
				return map[string]interface{}{"node3_output": params["node1_output"].(string) + " from node3"}, nil
			case "sire:local/node4":
				return map[string]interface{}{
					"node4_output": params["node2_output"].(string) + " | " + params["node3_output"].(string),
				}, nil
			default:
				return nil, fmt.Errorf("unknown tool: %s", tool)
			}
		},
	}

	// Create a dummy store for the engine
	dummyStore := &MockStore{}

	engine := NewEngine(dispatcher, dummyStore)

	workflow := &Workflow{
		ID:   "wf-branching",
		Name: "Branching Workflow",
		Steps: []Step{
			{ID: "node-1", Tool: "sire:local/node1"},
			{ID: "node-2", Tool: "sire:local/node2"},
			{ID: "node-3", Tool: "sire:local/node3"},
			{ID: "node-4", Tool: "sire:local/node4"},
		},
		Edges: []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-1", To: "node-3"},
			{From: "node-2", To: "node-4"},
			{From: "node-3", To: "node-4"},
		},
	}

	execution := &Execution{
		ID:         "exec-branching-1",
		WorkflowID: workflow.ID,
		Status:     ExecutionStatusRunning,
		StepStates: make(map[string]*StepState),
	}

	// 2. Execute
	execResult, err := engine.Execute(context.Background(), execution, workflow, nil)
	// 3. Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if execResult.Status != ExecutionStatusCompleted {
		t.Errorf("expected status %q, got %q", ExecutionStatusCompleted, execResult.Status)
	}
	if execResult.StepStates["node-1"].Status != StepStatusCompleted {
		t.Errorf("expected status %q, got %q", StepStatusCompleted, execResult.StepStates["node-1"].Status)
	}
	if execResult.StepStates["node-2"].Status != StepStatusCompleted {
		t.Errorf("expected status %q, got %q", StepStatusCompleted, execResult.StepStates["node-2"].Status)
	}
	if execResult.StepStates["node-3"].Status != StepStatusCompleted {
		t.Errorf("expected status %q, got %q", StepStatusCompleted, execResult.StepStates["node-3"].Status)
	}
	if execResult.StepStates["node-4"].Status != StepStatusCompleted {
		t.Errorf("expected status %q, got %q", StepStatusCompleted, execResult.StepStates["node-4"].Status)
	}
	if execResult.StepStates["node-4"].Output["node4_output"] != "hello from node2 | hello from node3" {
		t.Errorf("expected output %q, got %q", "hello from node2 | hello from node3", execResult.StepStates["node-4"].Output["node4_output"])
	}
}

func TestEngine_Execute_WithCycle(t *testing.T) {
	// 1. Setup
	dispatcher := &MockDispatcher{}

	// Create a dummy store for the engine
	dummyStore := &MockStore{}

	engine := NewEngine(dispatcher, dummyStore)

	workflow := &Workflow{
		ID:   "wf-cycle",
		Name: "Cyclic Workflow",
		Steps: []Step{
			{ID: "node-1", Tool: "sire:local/node1"},
			{ID: "node-2", Tool: "sire:local/node2"},
		},
		Edges: []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-2", To: "node-1"},
		},
	}

	execution := &Execution{
		ID:         "exec-cycle-1",
		WorkflowID: workflow.ID,
		Status:     ExecutionStatusRunning,
		StepStates: make(map[string]*StepState),
	}

	// 2. Execute
	_, err := engine.Execute(context.Background(), execution, workflow, nil)

	// 3. Assert
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "workflow has a cycle") {
		t.Errorf("expected error to contain %q, got %q", "workflow has a cycle", err.Error())
	}
}

func TestEngine_RetryLogic(t *testing.T) {
	// Mock dispatcher that fails for a few attempts, then succeeds
	attemptCount := 0
	mockDispatcher := &MockDispatcher{
		DispatchFunc: func(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error) {
			attemptCount++
			if tool == "sire:local/flaky-tool" && attemptCount <= 2 { // Fail for first 2 attempts
				return nil, fmt.Errorf("simulated transient error on attempt %d", attemptCount)
			}
			return map[string]interface{}{"result": "success"}, nil
		},
	}

	// Create a dummy store for the engine
	dummyStore := &MockStore{}

	engine := NewEngine(mockDispatcher, dummyStore)

	workflow := &Workflow{
		ID:   "retry-workflow",
		Name: "Retry Test Workflow",
		Steps: []Step{
			{
				ID:   "flaky_step",
				Tool: "sire:local/flaky-tool",
				Retry: &RetryPolicy{
					MaxAttempts: 3,
					Backoff:     "exponential",
				},
			},
		},
		Edges: []Edge{},
	}

	execution := &Execution{
		ID:         "exec-retry-1",
		WorkflowID: workflow.ID,
		Status:     ExecutionStatusRunning,
		StepStates: make(map[string]*StepState),
	}

	// First execution attempt (should fail and retry)
	execResult, err := engine.Execute(context.Background(), execution, workflow, nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if execResult.Status != ExecutionStatusRunning { // Workflow still running as step is retrying
		t.Errorf("expected status %q, got %q", ExecutionStatusRunning, execResult.Status)
	}
	if execResult.StepStates["flaky_step"].Status != StepStatusRetrying {
		t.Errorf("expected status %q, got %q", StepStatusRetrying, execResult.StepStates["flaky_step"].Status)
	}
	if execResult.StepStates["flaky_step"].Attempts != 1 {
		t.Errorf("expected attempts %d, got %d", 1, execResult.StepStates["flaky_step"].Attempts)
	}
	if !strings.Contains(execResult.StepStates["flaky_step"].Error, "simulated transient error on attempt 1") {
		t.Errorf("expected error to contain %q, got %q", "simulated transient error on attempt 1", execResult.StepStates["flaky_step"].Error)
	}
	if execResult.StepStates["flaky_step"].NextAttempt.IsZero() {
		t.Errorf("expected NextAttempt to not be zero")
	}

	// Force NextAttempt to be in the past for the next run
	execResult.StepStates["flaky_step"].NextAttempt = time.Time{} // Set to zero value

	// Second execution attempt (should fail and retry again)
	execResult, err = engine.Execute(context.Background(), execResult, workflow, nil) // Pass previous state
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if execResult.Status != ExecutionStatusRunning {
		t.Errorf("expected status %q, got %q", ExecutionStatusRunning, execResult.Status)
	}
	if execResult.StepStates["flaky_step"].Status != StepStatusRetrying {
		t.Errorf("expected status %q, got %q", StepStatusRetrying, execResult.StepStates["flaky_step"].Status)
	}
	if execResult.StepStates["flaky_step"].Attempts != 2 {
		t.Errorf("expected attempts %d, got %d", 2, execResult.StepStates["flaky_step"].Attempts)
	}
	if !strings.Contains(execResult.StepStates["flaky_step"].Error, "simulated transient error on attempt 2") {
		t.Errorf("expected error to contain %q, got %q", "simulated transient error on attempt 2", execResult.StepStates["flaky_step"].Error)
	}
	if execResult.StepStates["flaky_step"].NextAttempt.IsZero() {
		t.Errorf("expected NextAttempt to not be zero")
	}

	// Force NextAttempt to be in the past for the next run
	execResult.StepStates["flaky_step"].NextAttempt = time.Time{} // Set to zero value

	// Third execution attempt (should succeed)
	execResult, err = engine.Execute(context.Background(), execResult, workflow, nil) // Pass previous state
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if execResult.Status != ExecutionStatusCompleted {
		t.Errorf("expected status %q, got %q", ExecutionStatusCompleted, execResult.Status)
	}
	if execResult.StepStates["flaky_step"].Status != StepStatusCompleted {
		t.Errorf("expected status %q, got %q", StepStatusCompleted, execResult.StepStates["flaky_step"].Status)
	}
	if execResult.StepStates["flaky_step"].Attempts != 3 { // Attempts count should still be 3
		t.Errorf("expected attempts %d, got %d", 3, execResult.StepStates["flaky_step"].Attempts)
	}
	if execResult.StepStates["flaky_step"].Output["result"] != "success" {
		t.Errorf("expected output %q, got %q", "success", execResult.StepStates["flaky_step"].Output["result"])
	}
	if execResult.StepStates["flaky_step"].Error != "" {
		t.Errorf("expected empty error, got %q", execResult.StepStates["flaky_step"].Error)
	}
	if !execResult.StepStates["flaky_step"].NextAttempt.IsZero() { // NextAttempt should be zeroed on success
		t.Errorf("expected NextAttempt to be zero")
	}

	// Test exceeding MaxAttempts
	attemptCount = 0 // Reset attempt count for new test
	workflowFailed := &Workflow{
		ID:   "retry-workflow-failed",
		Name: "Retry Test Workflow Failed",
		Steps: []Step{
			{
				ID:   "flaky_step_failed",
				Tool: "sire:local/flaky-tool",
				Retry: &RetryPolicy{
					MaxAttempts: 1, // Only 1 attempt allowed
					Backoff:     "exponential",
				},
			},
		},
		Edges: []Edge{},
	}

	executionFailed := &Execution{
		ID:         "exec-retry-failed-1",
		WorkflowID: workflowFailed.ID,
		Status:     ExecutionStatusRunning,
		StepStates: make(map[string]*StepState),
	}

	execResultFailed, err := engine.Execute(context.Background(), executionFailed, workflowFailed, nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if execResultFailed.Status != ExecutionStatusFailed { // Workflow should be failed
		t.Errorf("expected status %q, got %q", ExecutionStatusFailed, execResultFailed.Status)
	}
	if execResultFailed.StepStates["flaky_step_failed"].Status != StepStatusFailed {
		t.Errorf("expected status %q, got %q", StepStatusFailed, execResultFailed.StepStates["flaky_step_failed"].Status)
	}
	if execResultFailed.StepStates["flaky_step_failed"].Attempts != 1 {
		t.Errorf("expected attempts %d, got %d", 1, execResultFailed.StepStates["flaky_step_failed"].Attempts)
	}
	if !strings.Contains(execResultFailed.StepStates["flaky_step_failed"].Error, "simulated transient error on attempt 1") {
		t.Errorf("expected error to contain %q, got %q", "simulated transient error on attempt 1", execResultFailed.StepStates["flaky_step_failed"].Error)
	}
}
