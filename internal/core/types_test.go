package core

import (
	"context"
	"testing"
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

func TestWorkflow(t *testing.T) {
	_ = Workflow{
		ID:   "wf-1",
		Name: "Test Workflow",
		Nodes: map[string]Node{
			"node-1": &MockNode{},
		},
		Edges: []Edge{
			{From: "node-1", To: "node-2"},
		},
	}
}

func TestExecution(t *testing.T) {
	_ = Execution{
		ID:         "exec-1",
		WorkflowID: "wf-1",
		Status:     "running",
		NodeStates: map[string]NodeState{
			"node-1": {
				Status: "success",
				Output: map[string]interface{}{"foo": "bar"},
				Error:  "",
			},
		},
	}
}
