package file

import (
	"context"
	"fmt"
	"os"

	"github.com/sire-run/sire/internal/core"
)

// FileReadNode reads a file.
type FileReadNode struct {
	path string
}

// NewFileReadNode creates a new FileReadNode.
func NewFileReadNode(config map[string]interface{}) (core.Node, error) {
	path, ok := config["path"].(string)
	if !ok {
		return nil, fmt.Errorf("config 'path' is required and must be a string")
	}
	return &FileReadNode{path: path}, nil
}

// Execute reads the file content.
func (n *FileReadNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Here we could allow overriding path from inputs.
	data, err := os.ReadFile(n.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", n.path, err)
	}
	return map[string]interface{}{"content": string(data)}, nil
}

// FileWriteNode writes to a file.
type FileWriteNode struct {
	path string
}

// NewFileWriteNode creates a new FileWriteNode.
func NewFileWriteNode(config map[string]interface{}) (core.Node, error) {
	path, ok := config["path"].(string)
	if !ok {
		return nil, fmt.Errorf("config 'path' is required and must be a string")
	}
	return &FileWriteNode{path: path}, nil
}

// Execute writes the content to a file.
func (n *FileWriteNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	content, ok := inputs["content"].(string)
	if !ok {
		return nil, fmt.Errorf("input 'content' is required and must be a string")
	}

	err := os.WriteFile(n.path, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write to file %s: %w", n.path, err)
	}

	return nil, nil // No output
}

func init() {
	core.RegisterNode("file.read", NewFileReadNode)
	core.RegisterNode("file.write", NewFileWriteNode)
}
