package agent

import (
	"context"
	"log"
	"time"

	"github.com/sire-run/sire/internal/core"
	"github.com/sire-run/sire/internal/storage"
)

// Agent is a background worker that scans for and resumes pending/retrying executions.
type Agent struct {
	store    storage.Store
	engine   *core.Engine
	interval time.Duration
}

// NewAgent creates a new Agent.
func NewAgent(store storage.Store, engine *core.Engine, interval time.Duration) *Agent {
	return &Agent{
		store:    store,
		engine:   engine,
		interval: interval,
	}
}

// Run starts the agent's periodic scanning and resumption process.
func (a *Agent) Run(ctx context.Context) {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	log.Println("Agent started, scanning for pending executions...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Agent shutting down.")
			return
		case <-ticker.C:
			a.scanAndResume(ctx)
		}
	}
}

func (a *Agent) scanAndResume(ctx context.Context) {
	executions, err := a.store.ListPendingExecutions()
	if err != nil {
		log.Printf("Agent: failed to list pending executions: %v", err)
		return
	}

	if len(executions) > 0 {
		log.Printf("Agent: found %d pending/retrying executions.", len(executions))
	}

	for _, exec := range executions {
		// Check if the execution is actually ready for retry (NextAttempt time has passed)
		// This check is also in the engine, but good to have here to avoid unnecessary processing
		readyForRetry := true
		for _, stepState := range exec.StepStates {
			if stepState.Status == core.StepStatusRetrying && time.Now().Before(stepState.NextAttempt) {
				readyForRetry = false
				break
			}
		}

		if !readyForRetry {
			continue
		}

		log.Printf("Agent: Resuming execution %s (Workflow: %s)", exec.ID, exec.WorkflowID)

		// Use the workflow definition stored in the execution object
		wf := exec.Workflow

		// Use a background context for the execution, so the agent can continue scanning
		go func(e *core.Execution, wf *core.Workflow) {
			_, err := a.engine.Execute(context.Background(), e, wf, nil) // Pass nil for inputs for now
			if err != nil {
				log.Printf("Agent: Error resuming execution %s: %v", e.ID, err)
			} else {
				log.Printf("Agent: Execution %s completed successfully.", e.ID)
			}
		}(exec, wf)
	}
}
