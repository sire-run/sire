package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"

	// Blank imports to ensure init() functions of node packages are run
	_ "github.com/sire-run/sire/internal/nodes/file"
	_ "github.com/sire-run/sire/internal/nodes/http"
	_ "github.com/sire-run/sire/internal/nodes/transform"
)

func TestMCPService_ToolExecute_ListNodes(t *testing.T) {
	s := rpc.NewServer()
	s.RegisterCodec(json2.NewCodec(), "application/json")
	if err := s.RegisterService(NewService(), "mcp"); err != nil {
		t.Fatalf("failed to register service: %v", err)
	}
	server := httptest.NewServer(s)
	defer server.Close()

	// Construct a valid JSON-RPC 2.0 request
	reqBody := `{"jsonrpc":"2.0","method":"mcp.ToolExecute","params":[{"name":"sire/listTools"}],"id":1}`

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, bytes.NewBufferString(reqBody))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()

	var rpcResp struct {
		Jsonrpc string   `json:"jsonrpc"`
		Result  []string `json:"result"`
		ID      int      `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if rpcResp.Jsonrpc != "2.0" {
		t.Errorf("expected jsonrpc %q, got %q", "2.0", rpcResp.Jsonrpc)
	}
	if rpcResp.ID != 1 {
		t.Errorf("expected ID 1, got %d", rpcResp.ID)
	}

	expectedTools := []string{"sire:local/http.request", "sire:local/file.read", "sire:local/file.write", "sire:local/data.transform"}
	if len(rpcResp.Result) != len(expectedTools) {
		t.Errorf("expected %d tools, got %d", len(expectedTools), len(rpcResp.Result))
	} else {
		// Sort both slices to ensure consistent comparison
		sort.Strings(rpcResp.Result)
		sort.Strings(expectedTools)
		for i := range expectedTools {
			if rpcResp.Result[i] != expectedTools[i] {
				t.Errorf("expected tool %q, got %q at index %d", expectedTools[i], rpcResp.Result[i], i)
			}
		}
	}
}
