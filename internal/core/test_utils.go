package core

import "context"

// MockNode is a mock implementation of the Node interface for testing.
type MockNode struct {
	ExecuteFunc func(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error)
}

// Execute calls the mock ExecuteFunc.
func (m *MockNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, inputs)
	}
	return nil, nil
}
