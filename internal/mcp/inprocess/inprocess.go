package inprocess

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/sire-run/sire/internal/core"
)

// ServiceMethod represents a callable service method.
type ServiceMethod func(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)

// InProcessServer hosts MCP services in-process.
type InProcessServer struct {
	mu       sync.RWMutex
	services map[string]map[string]ServiceMethod // scheme -> serviceName -> method
}

var (
	inProcessServerInstance *InProcessServer
	once                    sync.Once
)

// GetInProcessServer returns the singleton InProcessServer instance.
func GetInProcessServer() *InProcessServer {
	once.Do(func() {
		inProcessServerInstance = &InProcessServer{
			services: make(map[string]map[string]ServiceMethod),
		}
	})
	return inProcessServerInstance
}

// RegisterServiceMethod registers a service method with the in-process server.
// The serviceName and methodName are extracted from the toolURI.
// Example toolURI: sire:local/file.write
func (s *InProcessServer) RegisterServiceMethod(toolURI string, method ServiceMethod) error {
	u, err := url.Parse(toolURI)
	if err != nil {
		return fmt.Errorf("invalid tool URI: %w", err)
	}

	if u.Scheme != "sire" {
		return fmt.Errorf("unsupported scheme for in-process server: %s", u.Scheme)
	}

	// Expecting opaque to be in format "local/service.method"
	opaqueParts := strings.SplitN(u.Opaque, "/", 2)
	if len(opaqueParts) != 2 || opaqueParts[0] != "local" {
		return fmt.Errorf("invalid sire:local tool URI format: %s. Expected sire:local/service.method", toolURI)
	}
	serviceMethod := opaqueParts[1] // This should be "service.method"

	smParts := strings.SplitN(serviceMethod, ".", 2)
	if len(smParts) != 2 {
		return fmt.Errorf("invalid sire:local tool URI format: %s. Expected sire:local/service.method", toolURI)
	}
	serviceName := smParts[0]
	methodName := smParts[1]

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.services[serviceName]; !ok {
		s.services[serviceName] = make(map[string]ServiceMethod)
	}
	if _, ok := s.services[serviceName][methodName]; ok {
		return fmt.Errorf("service method %s already registered", toolURI)
	}
	s.services[serviceName][methodName] = method
	return nil
}

// InProcessDispatcher implements the core.Dispatcher interface for sire:local tools.
type InProcessDispatcher struct {
	server *InProcessServer
}

// NewInProcessDispatcher creates a new InProcessDispatcher.
func NewInProcessDispatcher() *InProcessDispatcher {
	return &InProcessDispatcher{
		server: GetInProcessServer(),
	}
}

// Dispatch dispatches a tool execution to the in-process server.
func (d *InProcessDispatcher) Dispatch(ctx context.Context, toolURI string, params map[string]interface{}) (map[string]interface{}, error) {
	u, err := url.Parse(toolURI)
	if err != nil {
		return nil, fmt.Errorf("invalid tool URI: %w", err)
	}

	if u.Scheme != "sire" {
		return nil, fmt.Errorf("unsupported scheme for in-process dispatcher: %s", u.Scheme)
	}

	// Expecting opaque to be in format "local/service.method"
	opaqueParts := strings.SplitN(u.Opaque, "/", 2)
	if len(opaqueParts) != 2 || opaqueParts[0] != "local" {
		return nil, fmt.Errorf("invalid sire:local tool URI format: %s. Expected sire:local/service.method", toolURI)
	}
	serviceMethod := opaqueParts[1] // This should be "service.method"

	smParts := strings.SplitN(serviceMethod, ".", 2)
	if len(smParts) != 2 {
		return nil, fmt.Errorf("invalid sire:local tool URI format: %s. Expected sire:local/service.method", toolURI)
	}
	serviceName := smParts[0]
	methodName := smParts[1]

	d.server.mu.RLock()
	defer d.server.mu.RUnlock()

	service, ok := d.server.services[serviceName]
	if !ok {
		return nil, fmt.Errorf("service %q not found", serviceName)
	}
	method, ok := service[methodName]
	if !ok {
		return nil, fmt.Errorf("method %q not found in service %q", methodName, serviceName)
	}

	return method(ctx, params)
}

// Ensure InProcessDispatcher implements core.Dispatcher
var _ core.Dispatcher = (*InProcessDispatcher)(nil)