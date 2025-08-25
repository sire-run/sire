package file

import (
	"context"
	"fmt"
	"os"

	"github.com/sire-run/sire/internal/mcp/inprocess" // Import the inprocess package
)

// ReadFile reads a file from the given path.
func ReadFile(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'path' is required and must be a string")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return map[string]interface{}{"content": string(data)}, nil
}

// WriteFile writes content to a file at the given path.
func WriteFile(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'path' is required and must be a string")
	}
	content, ok := params["content"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'content' is required and must be a string")
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write to file %s: %w", path, err)
	}

	return nil, nil // No output
}

func init() {
	server := inprocess.GetInProcessServer()
	err := server.RegisterServiceMethod("sire:local/file.read", ReadFile)
	if err != nil {
		panic(err) // Panics if registration fails, which should not happen in a well-formed app
	}
	err = server.RegisterServiceMethod("sire:local/file.write", WriteFile)
	if err != nil {
		panic(err) // Panics if registration fails
	}
}