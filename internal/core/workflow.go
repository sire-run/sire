package core

import "time" // Import time package for time.Time

// Step represents a single unit of work in a workflow.
type Step struct {
	ID     string                 `yaml:"id"`
	Tool   string                 `yaml:"tool"`
	Params map[string]interface{} `yaml:"params,omitempty"`
	Retry  *RetryPolicy           `yaml:"retry,omitempty"`
}

// RetryPolicy defines the retry behavior for a step.
type RetryPolicy struct {
	MaxAttempts int    `yaml:"max_attempts"` //nolint:tagliatelle
	Backoff     string `yaml:"backoff"` // e.g., "exponential"
}

// Workflow defines the structure of a workflow.
type Workflow struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
	Edges []Edge `yaml:"edges"`
}

// Edge represents a connection between two steps in a workflow.
type Edge struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// ExecutionStatus defines the status of a workflow execution.
type ExecutionStatus string

const (
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusRetrying  ExecutionStatus = "retrying"
)

// StepStatus defines the status of a single step in an execution.
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusRetrying  StepStatus = "retrying"
)

// Execution represents a single, durable run of a workflow.
type Execution struct {
	ID         string                `json:"id"`
	WorkflowID string                `json:"workflowId"`
	Workflow   *Workflow             `json:"workflow"` // New field to store the workflow definition
	Status     ExecutionStatus       `json:"status"`   // e.g., running, completed, failed, retrying
	StepStates map[string]*StepState `json:"stepStates"`
	CreatedAt  time.Time             `json:"createdAt"`
	UpdatedAt  time.Time             `json:"updatedAt"`
}

// StepState represents the state of a single step in an execution.
type StepState struct {
	Status      StepStatus             `json:"status"` // e.g., pending, running, completed, failed
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Attempts    int                    `json:"attempts"`
	NextAttempt time.Time              `json:"nextAttempt,omitempty"` // For exponential backoff
}
