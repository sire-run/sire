package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteDispatcher_Dispatch_Success(t *testing.T) {
	// Mock JSON-RPC server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req JSONRPCRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "2.0", req.JSONRPC)
		assert.Equal(t, "math.add", req.Method)
		assert.Equal(t, float64(1), req.Params.(map[string]interface{})["a"]) // JSON unmarshals numbers to float64
		assert.Equal(t, float64(2), req.Params.(map[string]interface{})["b"])

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
	require.NoError(t, err)
	assert.Equal(t, float64(3), output["sum"])
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
	require.Error(t, err)
	assert.Contains(t, err.Error(), "remote tool error (code -32000): Internal server error")
}

func TestRemoteDispatcher_Dispatch_InvalidURI(t *testing.T) {
	dispatcher := NewRemoteDispatcher()

	// Missing server URL (now missing scheme or host)
	_, err := dispatcher.Dispatch(context.Background(), "mcp:#math.add", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing scheme or host in RPC URL")

	// Missing tool name
	_, err = dispatcher.Dispatch(context.Background(), "mcp:http://localhost:8080#", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing tool name (service.method) in tool URI")

	// Unsupported scheme (now invalid mcp tool URI format)
	_, err = dispatcher.Dispatch(context.Background(), "http://localhost:8080#math.add", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mcp tool URI format")
}

func TestRemoteDispatcher_Dispatch_NetworkError(t *testing.T) {
	dispatcher := NewRemoteDispatcher()

	// Use a non-existent server to simulate network error
	toolURI := "mcp:http://localhost:9999#math.add"
	_, err := dispatcher.Dispatch(context.Background(), toolURI, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send HTTP request")
}

func TestRemoteDispatcher_Dispatch_Non200Status(t *testing.T) {
	// Mock server that returns non-200 status
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Server error"))
		assert.NoError(t, err)
	}))
	defer ts.Close()

	dispatcher := NewRemoteDispatcher()

	toolURI := fmt.Sprintf("mcp:%s#math.divide", ts.URL)
	_, err := dispatcher.Dispatch(context.Background(), toolURI, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "remote server returned non-OK status: 500, body: Server error")
}
