package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPService_ToolExecute_ListNodes(t *testing.T) {
	s := rpc.NewServer()
	s.RegisterCodec(json2.NewCodec(), "application/json")
	require.NoError(t, s.RegisterService(NewService(), "mcp"))
	server := httptest.NewServer(s)
	defer server.Close()

	// Construct a valid JSON-RPC 2.0 request
	reqBody := `{"jsonrpc":"2.0","method":"mcp.ToolExecute","params":[{"name":"sire/listNodes"}],"id":1}`

	resp, err := http.Post(server.URL, "application/json", bytes.NewBufferString(reqBody))
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	var rpcResp struct {
		Jsonrpc string   `json:"jsonrpc"`
		Result  []string `json:"result"`
		ID      int      `json:"id"`
	}

	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	require.NoError(t, err)

	assert.Equal(t, "2.0", rpcResp.Jsonrpc)
	assert.Equal(t, 1, rpcResp.ID)
	assert.Equal(t, []string{"http.request", "file.read", "file.write", "data.transform"}, rpcResp.Result)
}
