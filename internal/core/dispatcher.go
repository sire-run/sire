package core

import (
	"context"
	"fmt"
	"net/url"
)

// Dispatcher is responsible for executing a tool.
type Dispatcher interface {
	Dispatch(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error)
}

// DispatcherMux is a multiplexer for dispatchers.
type DispatcherMux struct {
	dispatchers map[string]Dispatcher
}

// NewDispatcherMux creates a new DispatcherMux.
func NewDispatcherMux() *DispatcherMux {
	return &DispatcherMux{
		dispatchers: make(map[string]Dispatcher),
	}
}

// Register registers a dispatcher for a given scheme.
func (m *DispatcherMux) Register(scheme string, dispatcher Dispatcher) {
	m.dispatchers[scheme] = dispatcher
}

// Dispatch dispatches a tool execution to the appropriate dispatcher.
func (m *DispatcherMux) Dispatch(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error) {
	u, err := url.Parse(tool)
	if err != nil {
		return nil, fmt.Errorf("invalid tool URI: %w", err)
	}

	dispatcher, ok := m.dispatchers[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("no dispatcher for scheme %q", u.Scheme)
	}

	return dispatcher.Dispatch(ctx, tool, params)
}
