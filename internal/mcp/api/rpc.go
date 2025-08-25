package api

import (
	"fmt"
	"net/http"

	"github.com/sire-run/sire/internal/mcp/service"
)

// MCPService provides the implementation of the MCP methods.
type MCPService struct {
	toolProvider *service.ToolProvider
}

// NewService creates a new MCPService.
func NewService() *MCPService {
	return &MCPService{
		toolProvider: service.NewToolProvider(),
	}
}

// Initialize implements the mcp/initialize method.
func (s *MCPService) Initialize(r *http.Request, args *struct{}, reply *map[string]interface{}) error {
	*reply = map[string]interface{}{
		"capabilities": map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "sire/listTools", // Changed from listNodes
					"description": "List available Sire tools.",
				},
				{
					"name":        "sire/createWorkflow",
					"description": "Create a new Sire workflow.",
				},
			},
		},
	}
	return nil
}

// Shutdown implements the mcp/shutdown method.
func (s *MCPService) Shutdown(r *http.Request, args *struct{}, reply *struct{}) error {
	return nil
}

// ToolExecute implements the mcp/tool/execute method.
func (s *MCPService) ToolExecute(r *http.Request, args *struct {
	Name   string
	Params []map[string]interface{}
}, reply *interface{}) error {
	switch args.Name {
	case "sire/listTools": // Changed from listNodes
		*reply = s.toolProvider.ListTools()
	case "sire/createWorkflow":
		wf, err := s.toolProvider.CreateWorkflow(args.Params)
		if err != nil {
			return err
		}
		*reply = wf
	default:
		return fmt.Errorf("method not found")
	}
	return nil
}
