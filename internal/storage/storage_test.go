package storage

import (
	"os"
	"testing"
	"time"

	"github.com/sire-run/sire/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoltDBStore_SaveAndLoadExecution(t *testing.T) {
	// Create a temporary BoltDB file
	tmpfile, err := os.CreateTemp("", "test-boltdb-*.db")
	require.NoError(t, err)
	dbPath := tmpfile.Name()
	require.NoError(t, tmpfile.Close()) // Close the file so BoltDB can open it
	defer func() { assert.NoError(t, os.Remove(dbPath)) }()

	store, err := NewBoltDBStore(dbPath)
	require.NoError(t, err)
	defer func() { assert.NoError(t, store.Close()) }()

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
	require.NoError(t, err)
	assert.False(t, exec.CreatedAt.IsZero())
	assert.False(t, exec.UpdatedAt.IsZero())

	// Load the execution
	loadedExec, err := store.LoadExecution("exec-123")
	require.NoError(t, err)
	assert.Equal(t, exec.ID, loadedExec.ID)
	assert.Equal(t, exec.WorkflowID, loadedExec.WorkflowID)
	assert.Equal(t, exec.Status, loadedExec.Status)
	assert.Equal(t, exec.StepStates["step1"].Status, loadedExec.StepStates["step1"].Status)
	assert.Equal(t, exec.StepStates["step1"].Output["foo"], loadedExec.StepStates["step1"].Output["foo"])
	assert.WithinDuration(t, exec.CreatedAt, loadedExec.CreatedAt, time.Second)
	assert.WithinDuration(t, exec.UpdatedAt, loadedExec.UpdatedAt, time.Second)

	// Test loading non-existent execution
	_, err = store.LoadExecution("non-existent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBoltDBStore_ListPendingExecutions(t *testing.T) {
	// Create a temporary BoltDB file
	tmpfile, err := os.CreateTemp("", "test-boltdb-list-*.db")
	require.NoError(t, err)
	dbPath := tmpfile.Name()
	require.NoError(t, tmpfile.Close())
	defer func() { assert.NoError(t, os.Remove(dbPath)) }()

	store, err := NewBoltDBStore(dbPath)
	require.NoError(t, err)
	defer func() { assert.NoError(t, store.Close()) }()

	// Create sample executions
	exec1 := &core.Execution{ID: "exec-1", WorkflowID: "wf-a", Status: core.ExecutionStatusRunning}
	exec2 := &core.Execution{ID: "exec-2", WorkflowID: "wf-b", Status: core.ExecutionStatusCompleted}
	exec3 := &core.Execution{ID: "exec-3", WorkflowID: "wf-c", Status: core.ExecutionStatusRetrying}
	exec4 := &core.Execution{ID: "exec-4", WorkflowID: "wf-d", Status: core.ExecutionStatusFailed}

	require.NoError(t, store.SaveExecution(exec1))
	require.NoError(t, store.SaveExecution(exec2))
	require.NoError(t, store.SaveExecution(exec3))
	require.NoError(t, store.SaveExecution(exec4))

	// List pending executions
	pending, err := store.ListPendingExecutions()
	require.NoError(t, err)

	// Expect exec1 and exec3 to be pending
	assert.Len(t, pending, 2)
	var pendingIDs []string
	for _, e := range pending {
		pendingIDs = append(pendingIDs, e.ID)
	}
	assert.Contains(t, pendingIDs, "exec-1")
	assert.Contains(t, pendingIDs, "exec-3")
}

func TestBoltDBStore_OpenAndClose(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-boltdb-open-close-*.db")
	require.NoError(t, err)
	dbPath := tmpfile.Name()
	require.NoError(t, tmpfile.Close())
	defer func() { assert.NoError(t, os.Remove(dbPath)) }()

	store, err := NewBoltDBStore(dbPath)
	require.NoError(t, err)
	assert.NotNil(t, store.db)

	err = store.Close()
	require.NoError(t, err)

	// Test opening a non-existent path (should create it)
	nonExistentPath := dbPath + ".new"
	store2, err := NewBoltDBStore(nonExistentPath)
	require.NoError(t, err)
	assert.NotNil(t, store2.db)
	require.NoError(t, store2.Close())
	defer func() { assert.NoError(t, os.Remove(nonExistentPath)) }()
}