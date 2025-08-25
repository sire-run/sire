package core

// Step represents a single unit of work in a workflow.
type Step struct {
	ID     string                 `yaml:"id"`
	Tool   string                 `yaml:"tool"`
	Params map[string]interface{} `yaml:"params,omitempty"`
	Retry  *RetryPolicy           `yaml:"retry,omitempty"`
}

// RetryPolicy defines the retry behavior for a step.
type RetryPolicy struct {
	MaxAttempts int    `yaml:"max_attempts"`
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

// Execution represents a single run of a workflow.
type Execution struct {
	ID         string
	WorkflowID string
	Status     string
	StepStates map[string]StepState
}

// StepState represents the state of a single step in an execution.
type StepState struct {
	Status string
	Output map[string]interface{}
	Error  string
}
