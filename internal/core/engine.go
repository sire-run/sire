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

	sortedNodes, err := topologicalSort(workflow.Nodes, workflow.Edges)
	if err != nil {
		return nil, err
	}

	nodeOutputs := make(map[string]map[string]interface{})

	for _, nodeID := range sortedNodes {
		node := workflow.Nodes[nodeID]

		// For now, we'll just merge the outputs of all parent nodes.
		// A more sophisticated approach would be to allow the user to specify which outputs to use.
		nodeInputs := make(map[string]interface{})
		for k, v := range inputs { // start with the initial inputs
			nodeInputs[k] = v
		}
		for _, edge := range workflow.Edges {
			if edge.To == nodeID {
				if parentOutput, ok := nodeOutputs[edge.From]; ok {
					for k, v := range parentOutput {
						nodeInputs[k] = v
					}
				}
			}
		}

		output, err := node.Execute(ctx, nodeInputs)
		if err != nil {
			execution.Status = "failed"
			execution.NodeStates[nodeID] = NodeState{
				Status: "failed",
				Error:  err.Error(),
			}
			return execution, fmt.Errorf("error executing node %s: %w", nodeID, err)
		}
		nodeOutputs[nodeID] = output

		execution.NodeStates[nodeID] = NodeState{
			Status: "success",
			Output: output,
		}
	}

	execution.Status = "success"

	return execution, nil
}

// a simple implementation of Kahn's algorithm for topological sorting.
func topologicalSort(nodes map[string]Node, edges []Edge) ([]string, error) {
	// 1. Calculate in-degrees
	inDegree := make(map[string]int)
	for id := range nodes {
		inDegree[id] = 0
	}
	for _, edge := range edges {
		inDegree[edge.To]++
	}

	// 2. Initialize queue with nodes with in-degree 0
	queue := []string{}
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// 3. Process queue
	result := []string{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		result = append(result, id)

		// 4. Decrement in-degrees of neighbors
		for _, edge := range edges {
			if edge.From == id {
				inDegree[edge.To]--
				if inDegree[edge.To] == 0 {
					queue = append(queue, edge.To)
				}
			}
		}
	}

	// 5. Check for cycles
	if len(result) != len(nodes) {
		return nil, fmt.Errorf("workflow has a cycle")
	}

	return result, nil
}
