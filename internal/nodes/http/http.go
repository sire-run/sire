package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sire-run/sire/internal/core"
)

// HTTPRequestNode is a node that makes an HTTP request.
type HTTPRequestNode struct {
	method  string
	url     string
	headers map[string]string
	body    string
}

// NewHTTPRequestNode creates a new HTTPRequestNode.
func NewHTTPRequestNode(config map[string]interface{}) (core.Node, error) {
	node := &HTTPRequestNode{}

	if method, ok := config["method"].(string); ok {
		node.method = strings.ToUpper(method)
	} else {
		return nil, fmt.Errorf("config 'method' is required and must be a string")
	}

	if url, ok := config["url"].(string); ok {
		node.url = url
	} else {
		return nil, fmt.Errorf("config 'url' is required and must be a string")
	}

	if headers, ok := config["headers"].(map[string]interface{}); ok {
		node.headers = make(map[string]string)
		for k, v := range headers {
			if val, ok := v.(string); ok {
				node.headers[k] = val
			}
		}
	}

	if body, ok := config["body"].(string); ok {
		node.body = body
	}

	return node, nil
}

// Execute makes the HTTP request.
func (n *HTTPRequestNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Here we could override config with inputs, e.g. for dynamic URLs.
	// For now, we'll just use the config.

	req, err := http.NewRequestWithContext(ctx, n.method, n.url, bytes.NewBufferString(n.body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range n.headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	output := map[string]interface{}{
		"statusCode": resp.StatusCode,
		"body":       string(respBody),
		"headers":    resp.Header,
	}

	return output, nil
}

// init registers the node with the core registry.
func init() {
	core.RegisterNode("http.request", NewHTTPRequestNode)
}
