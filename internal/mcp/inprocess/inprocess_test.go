package inprocess

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInProcessServer_RegisterAndDispatch(t *testing.T) {
	server := GetInProcessServer()

	// Test successful registration
	err := server.RegisterServiceMethod("sire:local/test.hello", func(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
		name, ok := params["name"].(string)
		if !ok {
			return nil, fmt.Errorf("name parameter missing or invalid")
		}
		return map[string]interface{}{"message": "Hello, " + name}, nil
	})
	require.NoError(t, err)

	// Test duplicate registration
	err = server.RegisterServiceMethod("sire:local/test.hello", func(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service method sire:local/test.hello already registered")

	dispatcher := NewInProcessDispatcher()

	// Test successful dispatch
	output, err := dispatcher.Dispatch(context.Background(), "sire:local/test.hello", map[string]interface{}{"name": "World"})
	require.NoError(t, err)
	assert.Equal(t, "Hello, World", output["message"])

	// Test dispatch to non-existent method
	_, err = dispatcher.Dispatch(context.Background(), "sire:local/test.nonexistent", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "method \"nonexistent\" not found in service \"test\"")

	// Test dispatch to non-existent service
	_, err = dispatcher.Dispatch(context.Background(), "sire:local/nonexistent.method", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service \"nonexistent\" not found")

	// Test invalid tool URI format
	_, err = dispatcher.Dispatch(context.Background(), "sire:local/invalid", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sire:local tool URI format")

	// Test unsupported scheme
	_, err = dispatcher.Dispatch(context.Background(), "http://example.com/test.hello", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported scheme for in-process dispatcher: http")
}

func TestInProcessServer_RegisterUnsupportedScheme(t *testing.T) {
	server := GetInProcessServer()
	err := server.RegisterServiceMethod("http://example.com/test.hello", func(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported scheme for in-process server: http")
}

func TestInProcessServer_RegisterInvalidURI(t *testing.T) {
	server := GetInProcessServer()
	err := server.RegisterServiceMethod("invalid-uri", func(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported scheme for in-process server: ")
}
