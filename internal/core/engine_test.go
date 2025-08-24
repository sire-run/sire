package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_Execute_LinearWorkflow(t *testing.T) {
	// 1. Setup
	engine := NewEngine()

	node1 := &MockNode{
		ExecuteFunc: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"node1_output": "hello"}, nil
		},
	}

	node2 := &MockNode{
		ExecuteFunc: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"node2_output": inputs["node1_output"].(string) + " world"}, nil
		},
	}

	workflow := &Workflow{
		ID:   "wf-1",
		Name: "Test Workflow",
		Nodes: map[string]Node{
			"node-1": node1,
			"node-2": node2,
		},
		// Edges are not used yet, but we'll define them for future use.
		Edges: []Edge{
			{From: "node-1", To: "node-2"},
		},
	}

	// 2. Execute
	execution, err := engine.Execute(context.Background(), workflow, nil)

	// 3. Assert
	require.NoError(t, err)
	assert.Equal(t, "success", execution.Status)
	assert.Equal(t, "success", execution.NodeStates["node-1"].Status)
	assert.Equal(t, "success", execution.NodeStates["node-2"].Status)
	assert.Equal(t, "hello world", execution.NodeStates["node-2"].Output["node2_output"])
}
