package storage

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sire-run/sire/internal/core"
)

// contains checks if a slice contains an element.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestBoltDBStore_SaveAndLoadExecution(t *testing.T) {
	// Create a temporary BoltDB file
	tmpfile, err := os.CreateTemp("", "test-boltdb-*.db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dbPath := tmpfile.Name()
	if err := tmpfile.Close(); err != nil { // Close the file so BoltDB can open it
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			t.Errorf("failed to remove temporary DB file: %v", err)
		}
	}()

	store, err := NewBoltDBStore(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close store: %v", err)
		}
	}()

	// Create a sample execution
	exec := &core.Execution{
		ID:         "exec-123",
		WorkflowID: "wf-abc",
		Status:     core.ExecutionStatusRunning, // Use enum
		StepStates: map[string]*core.StepState{ // Use pointers
			"step1": {Status: core.StepStatusCompleted, Output: map[string]interface{}{"foo": "bar"}}, // Use enum and pointer
		},
	}

	// Save the execution
	err = store.SaveExecution(exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to not be zero")
	}
	if exec.UpdatedAt.IsZero() {
		t.Errorf("expected UpdatedAt to not be zero")
	}

	// Load the execution
	loadedExec, err := store.LoadExecution("exec-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loadedExec.ID != exec.ID {
		t.Errorf("expected ID %q, got %q", exec.ID, loadedExec.ID)
	}
	if loadedExec.WorkflowID != exec.WorkflowID {
		t.Errorf("expected WorkflowID %q, got %q", exec.WorkflowID, loadedExec.WorkflowID)
	}
	if loadedExec.Status != exec.Status {
		t.Errorf("expected Status %q, got %q", exec.Status, loadedExec.Status)
	}
	if loadedExec.StepStates["step1"].Status != exec.StepStates["step1"].Status {
		t.Errorf("expected Step1 Status %q, got %q", exec.StepStates["step1"].Status, loadedExec.StepStates["step1"].Status)
	}
	if loadedExec.StepStates["step1"].Output["foo"] != exec.StepStates["step1"].Output["foo"] {
		t.Errorf("expected Step1 Output foo %q, got %q", exec.StepStates["step1"].Output["foo"], loadedExec.StepStates["step1"].Output["foo"])
	}

	delta := time.Second
	if loadedExec.CreatedAt.Before(exec.CreatedAt.Add(-delta)) || loadedExec.CreatedAt.After(exec.CreatedAt.Add(delta)) {
		t.Errorf("expected CreatedAt %v to be within %v of %v", loadedExec.CreatedAt, delta, exec.CreatedAt)
	}
	if loadedExec.UpdatedAt.Before(exec.UpdatedAt.Add(-delta)) || loadedExec.UpdatedAt.After(exec.UpdatedAt.Add(delta)) {
		t.Errorf("expected UpdatedAt %v to be within %v of %v", loadedExec.UpdatedAt, delta, exec.UpdatedAt)
	}

	// Test loading non-existent execution
	_, err = store.LoadExecution("non-existent")
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to contain %q, got %q", "not found", err.Error())
	}
}

func TestBoltDBStore_ListPendingExecutions(t *testing.T) {
	// Create a temporary BoltDB file
	tmpfile, err := os.CreateTemp("", "test-boltdb-list-*.db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dbPath := tmpfile.Name()
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			t.Errorf("failed to remove temporary DB file: %v", err)
		}
	}()

	store, err := NewBoltDBStore(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("failed to close store: %v", err)
		}
	}()

	// Create sample executions
	exec1 := &core.Execution{ID: "exec-1", WorkflowID: "wf-a", Status: core.ExecutionStatusRunning}
	exec2 := &core.Execution{ID: "exec-2", WorkflowID: "wf-b", Status: core.ExecutionStatusCompleted}
	exec3 := &core.Execution{ID: "exec-3", WorkflowID: "wf-c", Status: core.ExecutionStatusRetrying}
	exec4 := &core.Execution{ID: "exec-4", WorkflowID: "wf-d", Status: core.ExecutionStatusFailed}

	if err := store.SaveExecution(exec1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := store.SaveExecution(exec2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := store.SaveExecution(exec3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := store.SaveExecution(exec4); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// List pending executions
	pending, err := store.ListPendingExecutions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect exec1 and exec3 to be pending
	if len(pending) != 2 {
		t.Errorf("expected %d pending executions, got %d", 2, len(pending))
	}
	var pendingIDs []string
	for _, e := range pending {
		pendingIDs = append(pendingIDs, e.ID)
	}
	if !contains(pendingIDs, "exec-1") {
		t.Errorf("expected pending IDs to contain %q", "exec-1")
	}
	if !contains(pendingIDs, "exec-3") {
		t.Errorf("expected pending IDs to contain %q", "exec-3")
	}
}

func TestBoltDBStore_OpenAndClose(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-boltdb-open-close-*.db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dbPath := tmpfile.Name()
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			t.Errorf("failed to remove temporary DB file: %v", err)
		}
	}()

	store, err := NewBoltDBStore(dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.db == nil {
		t.Errorf("expected store.db to not be nil")
	}

	err = store.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test opening a non-existent path (should create it)
	nonExistentPath := dbPath + ".new"
	store2, err := NewBoltDBStore(nonExistentPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store2.db == nil {
		t.Errorf("expected store2.db to not be nil")
	}
	if err := store2.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if err := os.Remove(nonExistentPath); err != nil {
			t.Errorf("failed to remove temporary DB file: %v", err)
		}
	}()
}
