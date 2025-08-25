package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sire-run/sire/internal/core"
	bolt "go.etcd.io/bbolt" // Using bbolt as the BoltDB implementation
)

// Bucket names for BoltDB
var (
	executionBucket = []byte("executions")
)

// Store defines the interface for storing and retrieving workflow executions.
type Store interface {
	SaveExecution(execution *core.Execution) error
	LoadExecution(id string) (*core.Execution, error)
	ListPendingExecutions() ([]*core.Execution, error)
	// Add other necessary methods like DeleteExecution, etc.
}

// BoltDBStore implements the Store interface using BoltDB.
type BoltDBStore struct {
	db *bolt.DB
}

// NewBoltDBStore creates a new BoltDBStore.
func NewBoltDBStore(dbPath string) (*BoltDBStore, error) {
	db, err := bolt.Open(dbPath, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open BoltDB: %w", err)
	}

	// Create buckets if they don't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(executionBucket)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create BoltDB buckets: %w", err)
	}

	return &BoltDBStore{db: db}, nil
}

// Close closes the BoltDB database.
func (s *BoltDBStore) Close() error {
	return s.db.Close()
}

// SaveExecution saves a workflow execution to BoltDB.
func (s *BoltDBStore) SaveExecution(execution *core.Execution) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(executionBucket)
		if b == nil {
			return fmt.Errorf("bucket %s not found", executionBucket)
		}

		// Update timestamps
		now := time.Now()
		if execution.CreatedAt.IsZero() {
			execution.CreatedAt = now
		}
		execution.UpdatedAt = now

		data, err := json.Marshal(execution)
		if err != nil {
			return fmt.Errorf("failed to marshal execution: %w", err)
		}
		return b.Put([]byte(execution.ID), data)
	})
}

// LoadExecution loads a workflow execution from BoltDB.
func (s *BoltDBStore) LoadExecution(id string) (*core.Execution, error) {
	var execution core.Execution
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(executionBucket)
		if b == nil {
			return fmt.Errorf("bucket %s not found", executionBucket)
		}
		data := b.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("execution with ID %s not found", id)
		}
		return json.Unmarshal(data, &execution)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load execution %s: %w", id, err)
	}
	return &execution, nil
}

// ListPendingExecutions lists all executions that are not yet completed or failed.
func (s *BoltDBStore) ListPendingExecutions() ([]*core.Execution, error) {
	var pendingExecutions []*core.Execution
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(executionBucket)
		if b == nil {
			return fmt.Errorf("bucket %s not found", executionBucket)
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var execution core.Execution
			if err := json.Unmarshal(v, &execution); err != nil {
				return fmt.Errorf("failed to unmarshal execution from DB: %w", err)
			}
			// Assuming "running" and "retrying" are pending states
			if execution.Status == "running" || execution.Status == "retrying" {
				pendingExecutions = append(pendingExecutions, &execution)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pending executions: %w", err)
	}
	return pendingExecutions, nil
}
