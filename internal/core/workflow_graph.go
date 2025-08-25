package core

// GetExecutableSteps identifies steps that are ready to be executed.
// A step is executable if:
// 1. It is in a Pending state.
// 2. All its 'From' dependencies (predecessors) are in a Completed state.
// 3. It has no 'From' dependencies (it's a root step).
func GetExecutableSteps(workflow *Workflow, stepStates map[string]*StepState) []string {
	var executable []string

	// Build a map of step ID to its incoming dependencies
	dependencies := make(map[string]map[string]bool)
	for _, edge := range workflow.Edges {
		if _, ok := dependencies[edge.To]; !ok {
			dependencies[edge.To] = make(map[string]bool)
		}
		dependencies[edge.To][edge.From] = true
	}

	for _, step := range workflow.Steps {
		state, ok := stepStates[step.ID]
		if !ok || state.Status != StepStatusPending {
			// Skip if step state is not found or not pending
			continue
		}

		// Check if all dependencies are met
		allDependenciesMet := true
		if preds, hasDeps := dependencies[step.ID]; hasDeps {
			for predID := range preds {
				predState, predOk := stepStates[predID]
				if !predOk || predState.Status != StepStatusCompleted {
					allDependenciesMet = false
					break
				}
			}
		}

		if allDependenciesMet {
			executable = append(executable, step.ID)
		}
	}

	return executable
}
