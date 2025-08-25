package core

import (
	"sort"
	"testing"
)

// assertElementsMatch checks if two string slices contain the same elements, regardless of order.
func assertElementsMatch(t *testing.T, expected, actual []string, msg string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf("%s: lengths differ, expected %d, got %d", msg, len(expected), len(actual))
		return
	}

	sort.Strings(expected)
	sort.Strings(actual)

	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("%s: elements mismatch at index %d, expected %s, got %s", msg, i, expected[i], actual[i])
			return
		}
	}
}

// assertEmpty checks if a string slice is empty.
func assertEmpty(t *testing.T, actual []string, msg string) {
	t.Helper()
	if len(actual) != 0 {
		t.Errorf("%s: expected empty slice, got %v", msg, actual)
	}
}

func TestGetExecutableSteps(t *testing.T) {
	// Workflow:
	// A -> B
	// A -> C
	// D
	workflow := Workflow{
		ID:   "test-workflow",
		Name: "Test Workflow",
		Steps: []Step{
			{ID: "A", Tool: "toolA"},
			{ID: "B", Tool: "toolB"},
			{ID: "C", Tool: "toolC"},
			{ID: "D", Tool: "toolD"},
		},
		Edges: []Edge{
			{From: "A", To: "B"},
			{From: "A", To: "C"},
		},
	}

	// Initial state: all pending
	stepStates := map[string]*StepState{
		"A": {Status: StepStatusPending},
		"B": {Status: StepStatusPending},
		"C": {Status: StepStatusPending},
		"D": {Status: StepStatusPending},
	}

	// Expected: A and D are executable
	executable := GetExecutableSteps(&workflow, stepStates)
	assertElementsMatch(t, []string{"A", "D"}, executable, "Expected A and D to be executable initially")

	// Mark A as completed
	stepStates["A"].Status = StepStatusCompleted
	executable = GetExecutableSteps(&workflow, stepStates)
	assertElementsMatch(t, []string{"B", "C", "D"}, executable, "Expected B, C, and D to be executable after A completes")

	// Mark B as completed
	stepStates["B"].Status = StepStatusCompleted
	executable = GetExecutableSteps(&workflow, stepStates)
	assertElementsMatch(t, []string{"C", "D"}, executable, "Expected C and D to be executable after B completes")

	// Mark C as completed
	stepStates["C"].Status = StepStatusCompleted
	executable = GetExecutableSteps(&workflow, stepStates)
	assertElementsMatch(t, []string{"D"}, executable, "Expected D to be executable after C completes")

	// Mark D as completed
	stepStates["D"].Status = StepStatusCompleted
	executable = GetExecutableSteps(&workflow, stepStates)
	assertEmpty(t, executable, "Expected no executable steps when all are completed")

	// Test with a workflow where all steps are dependent on each other
	// A -> B -> C
	workflow2 := Workflow{
		ID:   "test-workflow-linear",
		Name: "Test Workflow Linear",
		Steps: []Step{
			{ID: "A", Tool: "toolA"},
			{ID: "B", Tool: "toolB"},
			{ID: "C", Tool: "toolC"},
		},
		Edges: []Edge{
			{From: "A", To: "B"},
			{From: "B", To: "C"},
		},
	}
	stepStates2 := map[string]*StepState{
		"A": {Status: StepStatusPending},
		"B": {Status: StepStatusPending},
		"C": {Status: StepStatusPending},
	}

	executable = GetExecutableSteps(&workflow2, stepStates2)
	assertElementsMatch(t, []string{"A"}, executable, "Expected A to be executable initially in linear workflow")

	stepStates2["A"].Status = StepStatusCompleted
	executable = GetExecutableSteps(&workflow2, stepStates2)
	assertElementsMatch(t, []string{"B"}, executable, "Expected B to be executable after A completes in linear workflow")

	stepStates2["B"].Status = StepStatusCompleted
	executable = GetExecutableSteps(&workflow2, stepStates2)
	assertElementsMatch(t, []string{"C"}, executable, "Expected C to be executable after B completes in linear workflow")

	stepStates2["C"].Status = StepStatusCompleted
	executable = GetExecutableSteps(&workflow2, stepStates2)
	assertEmpty(t, executable, "Expected no executable steps when all are completed in linear workflow")
}