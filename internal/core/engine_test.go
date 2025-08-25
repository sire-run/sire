package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	engine := NewEngine(dispatcher)

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

	// 2. Execute
	execution, err := engine.Execute(context.Background(), workflow, nil)

	// 3. Assert
	require.NoError(t, err)
	assert.Equal(t, "success", execution.Status)
	assert.Equal(t, "success", execution.StepStates["node-1"].Status)
	assert.Equal(t, "success", execution.StepStates["node-2"].Status)
	assert.Equal(t, "hello world", execution.StepStates["node-2"].Output["node2_output"])
}

func TestEngine_Execute_FailingWorkflow(t *testing.T) {
	// 1. Setup
	dispatcher := &MockDispatcher{
		DispatchFunc: func(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error) {
			return nil, fmt.Errorf("something went wrong")
		},
	}

	engine := NewEngine(dispatcher)

	workflow := &Workflow{
		ID:   "wf-1",
		Name: "Test Workflow",
		Steps: []Step{
			{ID: "node-1", Tool: "sire:local/failing-node"},
		},
	}

	// 2. Execute
	execution, err := engine.Execute(context.Background(), workflow, nil)

	// 3. Assert
	require.Error(t, err)
	assert.Equal(t, "failed", execution.Status)
	assert.Equal(t, "failed", execution.StepStates["node-1"].Status)
	assert.Equal(t, "something went wrong", execution.StepStates["node-1"].Error)
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
		require.NoError(t, err)
		assert.Equal(t, []string{"node-1", "node-2", "node-3"}, sorted)
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
		require.NoError(t, err)
		// The exact order of node-2 and node-3 can vary, so we check for both possibilities.
		assert.True(t, (sorted[1] == "node-2" && sorted[2] == "node-3") || (sorted[1] == "node-3" && sorted[2] == "node-2"))
		assert.Equal(t, "node-1", sorted[0])
		assert.Equal(t, "node-4", sorted[3])
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
		require.Error(t, err)
		assert.Equal(t, "workflow has a cycle", err.Error())
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

	engine := NewEngine(dispatcher)

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

	// 2. Execute
	execution, err := engine.Execute(context.Background(), workflow, nil)

	// 3. Assert
	require.NoError(t, err)
	assert.Equal(t, "success", execution.Status)
	assert.Equal(t, "success", execution.StepStates["node-1"].Status)
	assert.Equal(t, "success", execution.StepStates["node-2"].Status)
	assert.Equal(t, "success", execution.StepStates["node-3"].Status)
	assert.Equal(t, "success", execution.StepStates["node-4"].Status)
	assert.Equal(t, "hello from node2 | hello from node3", execution.StepStates["node-4"].Output["node4_output"])
}

func TestEngine_Execute_WithCycle(t *testing.T) {
	// 1. Setup
	dispatcher := &MockDispatcher{}

	engine := NewEngine(dispatcher)

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

	// 2. Execute
	_, err := engine.Execute(context.Background(), workflow, nil)

	// 3. Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow has a cycle")
}
