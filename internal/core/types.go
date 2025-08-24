package core

import "context"

// Workflow is the top-level structure for a workflow definition.
type Workflow struct {
	ID    string
	Name  string
	Nodes map[string]Node
	Edges []Edge
}

// Node is the interface that all nodes must implement.
type Node interface {
	Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
}

// Edge represents a connection between two nodes in a workflow.
type Edge struct {
	From string
	To   string
}

// Execution represents a single run of a workflow.
type Execution struct {
	ID         string
	WorkflowID string
	Status     string
	NodeStates map[string]NodeState
}

// NodeState represents the state of a single node in an execution.
type NodeState struct {
	Status string
	Output map[string]interface{}
	Error  string
}
