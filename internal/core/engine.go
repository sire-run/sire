package core

import (
	"context"
	"fmt"
)

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

// Engine is responsible for executing workflows.
type Engine struct{}

// NewEngine creates a new execution engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Execute executes a workflow.
func (e *Engine) Execute(ctx context.Context, workflow *Workflow, inputs map[string]interface{}) (*Execution, error) {
	execution := &Execution{
		WorkflowID: workflow.ID,
		Status:     "running",
		NodeStates: make(map[string]NodeState),
	}

	// For now, we assume a simple linear execution of nodes based on the order they are defined.
	// We will implement proper topological sorting and branching later.
	for nodeID, node := range workflow.Nodes {
		output, err := node.Execute(ctx, inputs)
		if err != nil {
			return nil, fmt.Errorf("error executing node %s: %w", nodeID, err)
		}
		inputs = output // Pass the output of the current node as input to the next.

		execution.NodeStates[nodeID] = NodeState{
			Status: "success",
			Output: output,
		}
	}

	execution.Status = "success"

	return execution, nil
}