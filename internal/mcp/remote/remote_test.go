package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRemoteDispatcher_Dispatch_Success(t *testing.T) {
	// Mock JSON-RPC server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected method %q, got %q", "POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type %q, got %q", "application/json", r.Header.Get("Content-Type"))
		}

		var req JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if req.JSONRPC != "2.0" {
			t.Errorf("expected JSONRPC %q, got %q", "2.0", req.JSONRPC)
		}
		if req.Method != "math.add" {
			t.Errorf("expected method %q, got %q", "math.add", req.Method)
		}
		// JSON unmarshals numbers to float64
		if req.Params.(map[string]interface{})["a"] != float64(1) {
			t.Errorf("expected param a %v, got %v", float64(1), req.Params.(map[string]interface{})["a"])
		}
		if req.Params.(map[string]interface{})["b"] != float64(2) {
			t.Errorf("expected param b %v, got %v", float64(2), req.Params.(map[string]interface{})["b"])
		}

		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  json.RawMessage(`{"sum": 3}`),
			ID:      req.ID,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil { // Check error
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	dispatcher := NewRemoteDispatcher()

	toolURI := fmt.Sprintf("mcp:%s#math.add", ts.URL)
	params := map[string]interface{}{
		"a": 1,
		"b": 2,
	}

	output, err := dispatcher.Dispatch(context.Background(), toolURI, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output["sum"] != float64(3) {
		t.Errorf("expected sum %v, got %v", float64(3), output["sum"])
	}
}

func TestRemoteDispatcher_Dispatch_RemoteError(t *testing.T) {
	// Mock JSON-RPC server that returns an error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32000,
				Message: "Internal server error",
			},
			ID: 1,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil { // Check error
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	dispatcher := NewRemoteDispatcher()

	toolURI := fmt.Sprintf("mcp:%s#math.subtract", ts.URL)
	params := map[string]interface{}{
		"a": 5,
		"b": 2,
	}

	_, err := dispatcher.Dispatch(context.Background(), toolURI, params)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "remote tool error (code -32000): Internal server error") {
		t.Errorf("expected error to contain %q, got %q", "remote tool error (code -32000): Internal server error", err.Error())
	}
}

func TestRemoteDispatcher_Dispatch_InvalidURI(t *testing.T) {
	dispatcher := NewRemoteDispatcher()

	// Missing server URL (now missing scheme or host)
	_, err := dispatcher.Dispatch(context.Background(), "mcp:#math.add", nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "missing scheme or host in RPC URL") {
		t.Errorf("expected error to contain %q, got %q", "missing scheme or host in RPC URL", err.Error())
	}

	// Missing tool name
	_, err = dispatcher.Dispatch(context.Background(), "mcp:http://localhost:8080#", nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "missing tool name (service.method) in tool URI") {
		t.Errorf("expected error to contain %q, got %q", "missing tool name (service.method) in tool URI", err.Error())
	}

	// Unsupported scheme (now invalid mcp tool URI format)
	_, err = dispatcher.Dispatch(context.Background(), "http://localhost:8080#math.add", nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "invalid mcp tool URI format") {
		t.Errorf("expected error to contain %q, got %q", "invalid mcp tool URI format", err.Error())
	}
}

func TestRemoteDispatcher_Dispatch_NetworkError(t *testing.T) {
	dispatcher := NewRemoteDispatcher()

	// Use a non-existent server to simulate network error
	toolURI := "mcp:http://localhost:9999#math.add"
	_, err := dispatcher.Dispatch(context.Background(), toolURI, nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "failed to send HTTP request") {
		t.Errorf("expected error to contain %q, got %q", "failed to send HTTP request", err.Error())
	}
}

func TestRemoteDispatcher_Dispatch_Non200Status(t *testing.T) {
	// Mock server that returns non-200 status
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Server error"))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}))
	defer ts.Close()

	dispatcher := NewRemoteDispatcher()

	toolURI := fmt.Sprintf("mcp:%s#math.divide", ts.URL)
	_, err := dispatcher.Dispatch(context.Background(), toolURI, nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "remote server returned non-OK status: 500, body: Server error") {
		t.Errorf("expected error to contain %q, got %q", "remote server returned non-OK status: 500, body: Server error", err.Error())
	}
}
