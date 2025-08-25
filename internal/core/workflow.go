package core

import "context"

// Node is the interface that all nodes must implement.
type Node interface {
	Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
}

// Workflow defines the structure of a workflow.
type Workflow struct {
	ID    string
	Name  string
	Nodes map[string]NodeDefinition
	Edges []Edge
}

// NodeDefinition represents a node in a workflow definition.
type NodeDefinition struct {
	Type   string
	Config map[string]interface{}
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
