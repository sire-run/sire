package core

import "fmt"

// NodeConstructor creates a new instance of a Node from a config.
type NodeConstructor func(config map[string]interface{}) (Node, error)

// NodeRegistry holds the registered node types.
type NodeRegistry struct {
	constructors map[string]NodeConstructor
}

// NewNodeRegistry creates a new node registry.
func NewNodeRegistry() *NodeRegistry {
	return &NodeRegistry{
		constructors: make(map[string]NodeConstructor),
	}
}

// Register registers a new node type.
func (r *NodeRegistry) Register(typeName string, constructor NodeConstructor) {
	if _, exists := r.constructors[typeName]; exists {
		panic(fmt.Sprintf("node type '%s' already registered", typeName))
	}
	r.constructors[typeName] = constructor
}

// GetNodeConstructor retrieves a node constructor from the registry.
func (r *NodeRegistry) GetNodeConstructor(typeName string) (NodeConstructor, error) {
	constructor, ok := r.constructors[typeName]
	if !ok {
		return nil, fmt.Errorf("node type '%s' not found", typeName)
	}
	return constructor, nil
}
