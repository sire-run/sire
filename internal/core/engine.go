package core

import (
	"context"
	"fmt"
)

// Engine is responsible for executing workflows.
type Engine struct {
	dispatcher Dispatcher
}

// NewEngine creates a new execution engine.
func NewEngine(dispatcher Dispatcher) *Engine {
	return &Engine{dispatcher: dispatcher}
}

// Execute executes a workflow.
func (e *Engine) Execute(ctx context.Context, workflow *Workflow, inputs map[string]interface{}) (*Execution, error) {
	execution := &Execution{
		WorkflowID: workflow.ID,
		Status:     "running",
		StepStates: make(map[string]StepState),
	}

	steps := make(map[string]Step)
	for _, step := range workflow.Steps {
		steps[step.ID] = step
	}

	sortedSteps, err := topologicalSort(steps, workflow.Edges)
	if err != nil {
		return nil, err
	}

	stepOutputs := make(map[string]map[string]interface{})

	for _, stepID := range sortedSteps {
		step := steps[stepID]

		stepInputs := make(map[string]interface{})
		// Start with the initial inputs to the workflow
		for k, v := range inputs {
			stepInputs[k] = v
		}
		// Add parameters defined in the step itself
		for k, v := range step.Params {
			stepInputs[k] = v
		}
		// Add outputs from parent steps
		for _, edge := range workflow.Edges {
			if edge.To == stepID {
				if parentOutput, ok := stepOutputs[edge.From]; ok {
					for k, v := range parentOutput {
						stepInputs[k] = v
					}
				}
			}
		}

		output, err := e.dispatcher.Dispatch(ctx, step.Tool, stepInputs)
		if err != nil {
			execution.Status = "failed"
			execution.StepStates[stepID] = StepState{
				Status: "failed",
				Error:  err.Error(),
			}
			return execution, fmt.Errorf("error executing step %s: %w", stepID, err)
		}
		stepOutputs[stepID] = output

		execution.StepStates[stepID] = StepState{
			Status: "success",
			Output: output,
		}
	}

	execution.Status = "success"

	return execution, nil
}

// a simple implementation of Kahn's algorithm for topological sorting.
func topologicalSort(steps map[string]Step, edges []Edge) ([]string, error) {
	// 1. Calculate in-degrees
	inDegree := make(map[string]int)
	for id := range steps {
		inDegree[id] = 0
	}
	for _, edge := range edges {
		inDegree[edge.To]++
	}

	// 2. Initialize queue with steps with in-degree 0
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
	if len(result) != len(steps) {
		return nil, fmt.Errorf("workflow has a cycle")
	}

	return result, nil
}