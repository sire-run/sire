package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"github.com/sire-run/sire/internal/core"
)

func TestMCPService_ToolExecute_CreateWorkflow(t *testing.T) {
	s := rpc.NewServer()
	s.RegisterCodec(json2.NewCodec(), "application/json")
	if err := s.RegisterService(NewService(), "mcp"); err != nil {
		t.Fatalf("failed to register service: %v", err)
	}
	server := httptest.NewServer(s)
	defer server.Close()

	// Construct a valid JSON-RPC 2.0 request
	reqBody := `{"jsonrpc":"2.0","method":"mcp.ToolExecute","params":[{"name":"sire/createWorkflow","params":[{"id":"step1","tool":"file.write","params":{"path":"/tmp/foo.txt","content":"bar"}}]}],"id":1}`

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
		Jsonrpc string        `json:"jsonrpc"`
		Result  core.Workflow `json:"result"`
		ID      int           `json:"id"`
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
	if rpcResp.Result.ID != "generated-workflow" {
		t.Errorf("expected workflow ID %q, got %q", "generated-workflow", rpcResp.Result.ID)
	}
}
