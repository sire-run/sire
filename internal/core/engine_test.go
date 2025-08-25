package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockNode is a mock implementation of the Node interface for testing.
type MockNode struct {
	ExecuteFunc func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
}

// Execute calls the mock ExecuteFunc.
func (m *MockNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, inputs)
	}
	return nil, nil
}

func TestEngine_Execute_LinearWorkflow(t *testing.T) {
	// 1. Setup
	registry := NewNodeRegistry()
	registry.Register("node1", func(config map[string]interface{}) (Node, error) {
		return &MockNode{
			ExecuteFunc: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{"node1_output": "hello"}, nil
			},
		}, nil
	})
	registry.Register("node2", func(config map[string]interface{}) (Node, error) {
		return &MockNode{
			ExecuteFunc: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{"node2_output": inputs["node1_output"].(string) + " world"}, nil
			},
		}, nil
	})

	engine := NewEngine(registry)

	workflow := &Workflow{
		ID:   "wf-1",
		Name: "Test Workflow",
		Nodes: map[string]NodeDefinition{
			"node-1": {Type: "node1"},
			"node-2": {Type: "node2"},
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
	assert.Equal(t, "success", execution.NodeStates["node-1"].Status)
	assert.Equal(t, "success", execution.NodeStates["node-2"].Status)
	assert.Equal(t, "hello world", execution.NodeStates["node-2"].Output["node2_output"])
}

func TestEngine_Execute_FailingWorkflow(t *testing.T) {
	// 1. Setup
	registry := NewNodeRegistry()
	registry.Register("failing-node", func(config map[string]interface{}) (Node, error) {
		return &MockNode{
			ExecuteFunc: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
				return nil, fmt.Errorf("something went wrong")
			},
		}, nil
	})

	engine := NewEngine(registry)

	workflow := &Workflow{
		ID:   "wf-1",
		Name: "Test Workflow",
		Nodes: map[string]NodeDefinition{
			"node-1": {Type: "failing-node"},
		},
	}

	// 2. Execute
	execution, err := engine.Execute(context.Background(), workflow, nil)

	// 3. Assert
	require.Error(t, err)
	assert.Equal(t, "failed", execution.Status)
	assert.Equal(t, "failed", execution.NodeStates["node-1"].Status)
	assert.Equal(t, "something went wrong", execution.NodeStates["node-1"].Error)
}

func TestTopologicalSort(t *testing.T) {
	t.Run("simple linear workflow", func(t *testing.T) {
		nodes := map[string]Node{
			"node-1": &MockNode{},
			"node-2": &MockNode{},
			"node-3": &MockNode{},
		}
		edges := []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-2", To: "node-3"},
		}

		sorted, err := topologicalSort(nodes, edges)
		require.NoError(t, err)
		assert.Equal(t, []string{"node-1", "node-2", "node-3"}, sorted)
	})

	t.Run("workflow with a branch", func(t *testing.T) {
		nodes := map[string]Node{
			"node-1": &MockNode{},
			"node-2": &MockNode{},
			"node-3": &MockNode{},
			"node-4": &MockNode{},
		}
		edges := []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-1", To: "node-3"},
			{From: "node-2", To: "node-4"},
			{From: "node-3", To: "node-4"},
		}

		sorted, err := topologicalSort(nodes, edges)
		require.NoError(t, err)
		// The exact order of node-2 and node-3 can vary, so we check for both possibilities.
		assert.True(t, (sorted[1] == "node-2" && sorted[2] == "node-3") || (sorted[1] == "node-3" && sorted[2] == "node-2"))
		assert.Equal(t, "node-1", sorted[0])
		assert.Equal(t, "node-4", sorted[3])
	})

	t.Run("workflow with a cycle", func(t *testing.T) {
		nodes := map[string]Node{
			"node-1": &MockNode{},
			"node-2": &MockNode{},
			"node-3": &MockNode{},
		}
		edges := []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-2", To: "node-3"},
			{From: "node-3", To: "node-1"},
		}

		_, err := topologicalSort(nodes, edges)
		require.Error(t, err)
		assert.Equal(t, "workflow has a cycle", err.Error())
	})
}

func TestEngine_Execute_BranchingWorkflow(t *testing.T) {
	// 1. Setup
	registry := NewNodeRegistry()
	registry.Register("node1", func(config map[string]interface{}) (Node, error) {
		return &MockNode{
			ExecuteFunc: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{"node1_output": "hello"}, nil
			},
		}, nil
	})
	registry.Register("node2", func(config map[string]interface{}) (Node, error) {
		return &MockNode{
			ExecuteFunc: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{"node2_output": inputs["node1_output"].(string) + " from node2"}, nil
			},
		}, nil
	})
	registry.Register("node3", func(config map[string]interface{}) (Node, error) {
		return &MockNode{
			ExecuteFunc: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{"node3_output": inputs["node1_output"].(string) + " from node3"}, nil
			},
		}, nil
	})
	registry.Register("node4", func(config map[string]interface{}) (Node, error) {
		return &MockNode{
			ExecuteFunc: func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{
					"node4_output": inputs["node2_output"].(string) + " | " + inputs["node3_output"].(string),
				}, nil
			},
		}, nil
	})

	engine := NewEngine(registry)

	workflow := &Workflow{
		ID:   "wf-branching",
		Name: "Branching Workflow",
		Nodes: map[string]NodeDefinition{
			"node-1": {Type: "node1"},
			"node-2": {Type: "node2"},
			"node-3": {Type: "node3"},
			"node-4": {Type: "node4"},
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
	assert.Equal(t, "success", execution.NodeStates["node-1"].Status)
	assert.Equal(t, "success", execution.NodeStates["node-2"].Status)
	assert.Equal(t, "success", execution.NodeStates["node-3"].Status)
	assert.Equal(t, "success", execution.NodeStates["node-4"].Status)
	assert.Equal(t, "hello from node2 | hello from node3", execution.NodeStates["node-4"].Output["node4_output"])
}
