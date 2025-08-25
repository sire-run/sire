package inprocess

import (
	"context"
	"fmt"
	"strings"
	"testing"
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test duplicate registration
	err = server.RegisterServiceMethod("sire:local/test.hello", func(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	})
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "service method sire:local/test.hello already registered") {
		t.Errorf("expected error to contain %q, got %q", "service method sire:local/test.hello already registered", err.Error())
	}

	dispatcher := NewInProcessDispatcher()

	// Test successful dispatch
	output, err := dispatcher.Dispatch(context.Background(), "sire:local/test.hello", map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output["message"] != "Hello, World" {
		t.Errorf("expected message %q, got %q", "Hello, World", output["message"])
	}

	// Test dispatch to non-existent method
	_, err = dispatcher.Dispatch(context.Background(), "sire:local/test.nonexistent", nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "method \"nonexistent\" not found in service \"test\"") {
		t.Errorf("expected error to contain %q, got %q", "method \"nonexistent\" not found in service \"test\"", err.Error())
	}

	// Test dispatch to non-existent service
	_, err = dispatcher.Dispatch(context.Background(), "sire:local/nonexistent.method", nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "service \"nonexistent\" not found") {
		t.Errorf("expected error to contain %q, got %q", "service \"nonexistent\" not found", err.Error())
	}

	// Test invalid tool URI format
	_, err = dispatcher.Dispatch(context.Background(), "sire:local/invalid", nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "invalid sire:local tool URI format") {
		t.Errorf("expected error to contain %q, got %q", "invalid sire:local tool URI format", err.Error())
	}

	// Test unsupported scheme
	_, err = dispatcher.Dispatch(context.Background(), "http://example.com/test.hello", nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "unsupported scheme for in-process dispatcher: http") {
		t.Errorf("expected error to contain %q, got %q", "unsupported scheme for in-process dispatcher: http", err.Error())
	}
}

func TestInProcessServer_RegisterUnsupportedScheme(t *testing.T) {
	server := GetInProcessServer()
	err := server.RegisterServiceMethod("http://example.com/test.hello", func(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	})
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "unsupported scheme for in-process server: http") {
		t.Errorf("expected error to contain %q, got %q", "unsupported scheme for in-process server: http", err.Error())
	}
}

func TestInProcessServer_RegisterInvalidURI(t *testing.T) {
	server := GetInProcessServer()
	err := server.RegisterServiceMethod("invalid-uri", func(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
		return nil, nil
	})
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "unsupported scheme for in-process server: ") {
		t.Errorf("expected error to contain %q, got %q", "unsupported scheme for in-process server: ", err.Error())
	}
}
