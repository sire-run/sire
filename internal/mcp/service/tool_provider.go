package service

import (
	"fmt"
	"github.com/sire-run/sire/internal/core"
	"github.com/sire-run/sire/internal/mcp/integration"
)

// ToolProvider provides the implementation for the MCP tools.
type ToolProvider struct {
	sireAdapter *integration.SireAdapter
}

// NewToolProvider creates a new ToolProvider.
func NewToolProvider() *ToolProvider {
	return &ToolProvider{
		sireAdapter: integration.NewSireAdapter(),
	}
}

// ListNodes lists the available Sire nodes.
func (p *ToolProvider) ListNodes() []string {
	return p.sireAdapter.GetNodeTypes()
}

// CreateWorkflow creates a new Sire workflow.
func (p *ToolProvider) CreateWorkflow(steps []map[string]interface{}) (*core.Workflow, error) {
	// This is a placeholder implementation.
	// A real implementation would be more sophisticated.
	nodes := make(map[string]core.NodeDefinition)
	edges := []core.Edge{}
	for i, step := range steps {
		nodeID := fmt.Sprintf("node-%d", i)
		nodes[nodeID] = core.NodeDefinition{
			Type:   step["type"].(string),
			Config: step["config"].(map[string]interface{}),
		}
		if i > 0 {
			edges = append(edges, core.Edge{
				From: fmt.Sprintf("node-%d", i-1),
				To:   nodeID,
			})
		}
	}

	return &core.Workflow{
		ID:    "new-workflow",
		Name:  "New Workflow",
		Nodes: nodes,
		Edges: edges,
	}, nil
}
