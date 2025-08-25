package core

import (
	"context"
	"fmt"
)

// Store defines the interface for storing and retrieving workflow executions.
// This is a copy from internal/storage/storage.go to avoid circular dependency.
// In a real project, this would be in a shared package or passed as an interface.
type Store interface {
	SaveExecution(execution *Execution) error
	LoadExecution(id string) (*Execution, error)
	ListPendingExecutions() ([]*Execution, error)
	// Add other necessary methods like DeleteExecution, etc.
}

// Engine is responsible for executing workflows.
type Engine struct {
	dispatcher Dispatcher
	store      Store // New field for storage
}

// NewEngine creates a new execution engine.
func NewEngine(dispatcher Dispatcher, store Store) *Engine {
	return &Engine{dispatcher: dispatcher, store: store}
}

// Execute executes a workflow.
// It now takes an existing execution object.
func (e *Engine) Execute(ctx context.Context, execution *Execution, workflow *Workflow, inputs map[string]interface{}) (*Execution, error) {
	// No longer creating a new execution here, it's passed in.
	// Ensure initial status is running if it's a new execution or resuming
	if execution.Status == "" { // Or some other initial state check
		execution.Status = ExecutionStatusRunning
	}

	steps := make(map[string]Step)
	for _, step := range workflow.Steps {
		steps[step.ID] = step
	}

	sortedSteps, err := topologicalSort(steps, workflow.Edges)
	if err != nil {
		execution.Status = ExecutionStatusFailed // Mark as failed if topological sort fails
		if e.store != nil {
			_ = e.store.SaveExecution(execution) // Attempt to save state
		}
		return execution, fmt.Errorf("workflow topological sort failed: %w", err)
	}

	// Load stepOutputs from execution.StepStates if resuming
	stepOutputs := make(map[string]map[string]interface{})
	for stepID, stepState := range execution.StepStates {
		if stepState.Status == StepStatusCompleted {
			stepOutputs[stepID] = stepState.Output
		}
	}

	for _, stepID := range sortedSteps {
		step := steps[stepID]

		// If step is already completed, skip it
		if ss, ok := execution.StepStates[stepID]; ok && ss.Status == StepStatusCompleted {
			continue
		}

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
			execution.Status = ExecutionStatusFailed // Use the new enum
			execution.StepStates[stepID] = &StepState{ // Assign pointer
				Status: StepStatusFailed, // Use the new enum
				Error:  err.Error(),
			}
			if e.store != nil {
				_ = e.store.SaveExecution(execution) // Attempt to save state
			}
			return execution, fmt.Errorf("error executing step %s: %w", stepID, err)
		}
		stepOutputs[stepID] = output

		execution.StepStates[stepID] = &StepState{ // Assign pointer
			Status: StepStatusCompleted, // Use the new enum
			Output: output,
		}
		// Save state after each step (S9.2.3)
		if e.store != nil {
			if err := e.store.SaveExecution(execution); err != nil {
				return execution, fmt.Errorf("failed to save execution state after step %s: %w", stepID, err)
			}
		}
	}

	execution.Status = ExecutionStatusCompleted // Use the new enum
	if e.store != nil {
		_ = e.store.SaveExecution(execution) // Final save
	}

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
