package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RegistryTestNode is a mock node for testing the registry
type RegistryTestNode struct {
	configValue string
}

func (n *RegistryTestNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"config": n.configValue}, nil
}

// NewRegistryTestNode is a constructor for RegistryTestNode
func NewRegistryTestNode(config map[string]interface{}) (Node, error) {
	val, ok := config["key"].(string)
	if !ok {
		return nil, fmt.Errorf("config key 'key' must be a string")
	}
	return &RegistryTestNode{configValue: val}, nil
}

func TestNodeRegistry(t *testing.T) {
	registry := NewNodeRegistry()

	t.Run("register and get a node constructor", func(t *testing.T) {
		registry.Register("test-node", NewRegistryTestNode)

		constructor, err := registry.GetNodeConstructor("test-node")
		require.NoError(t, err)

		node, err := constructor(map[string]interface{}{"key": "value"})
		require.NoError(t, err)
		assert.IsType(t, &RegistryTestNode{}, node)
		assert.Equal(t, "value", node.(*RegistryTestNode).configValue)
	})

	t.Run("get a constructor for a non-existent node", func(t *testing.T) {
		_, err := registry.GetNodeConstructor("non-existent-node")
		require.Error(t, err)
		assert.Equal(t, "node type 'non-existent-node' not found", err.Error())
	})

	t.Run("register a node twice", func(t *testing.T) {
		// This subtest expects a panic
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			} else {
				assert.Equal(t, "node type 'test-node' already registered", r)
			}
		}()
		registry.Register("test-node", NewRegistryTestNode) // Already registered in a previous subtest
	})

	t.Run("create a node with bad config", func(t *testing.T) {
		// Re-register in a new registry to avoid panic
		freshRegistry := NewNodeRegistry()
		freshRegistry.Register("test-node-bad-config", NewRegistryTestNode)
		constructor, err := freshRegistry.GetNodeConstructor("test-node-bad-config")
		require.NoError(t, err)

		_, err = constructor(map[string]interface{}{"key": 123})
		require.Error(t, err)
		assert.Equal(t, "config key 'key' must be a string", err.Error())
	})
}
