package service

import (
	"fmt"

	"github.com/sire-run/sire/internal/core"
	"github.com/sire-run/sire/internal/mcp/inprocess"
)

// ToolProvider provides the implementation for the MCP tools.
type ToolProvider struct{}

// NewToolProvider creates a new ToolProvider.
func NewToolProvider() *ToolProvider {
	return &ToolProvider{}
}

// ListTools lists the available Sire tools.
func (p *ToolProvider) ListTools() []string {
	server := inprocess.GetInProcessServer()
	return server.ListRegisteredTools()
}

// CreateWorkflow creates a new Sire workflow from a list of steps.
// Each step in the input map should contain "id", "tool", and optionally "params".
func (p *ToolProvider) CreateWorkflow(inputSteps []map[string]interface{}) (*core.Workflow, error) {
	steps := make([]core.Step, len(inputSteps))
	edges := []core.Edge{}

	for i, inputStep := range inputSteps {
		stepID, ok := inputStep["id"].(string)
		if !ok {
			return nil, fmt.Errorf("step %d: 'id' is required and must be a string", i)
		}
		toolURI, ok := inputStep["tool"].(string)
		if !ok {
			return nil, fmt.Errorf("step %d: 'tool' is required and must be a string", i)
		}

		params, _ := inputStep["params"].(map[string]interface{}) // params is optional

		steps[i] = core.Step{
			ID:     stepID,
			Tool:   toolURI,
			Params: params,
		}

		// For simplicity, create a linear workflow for now
		if i > 0 {
			edges = append(edges, core.Edge{
				From: inputSteps[i-1]["id"].(string),
				To:   stepID,
			})
		}
	}

	return &core.Workflow{
		ID:    "generated-workflow",
		Name:  "Generated Workflow",
		Steps: steps,
		Edges: edges,
	}, nil
}
