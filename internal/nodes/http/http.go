package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sire-run/sire/internal/mcp/inprocess" // Import the inprocess package
)

// Request makes an HTTP request.
func Request(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	method, ok := params["method"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'method' is required and must be a string")
	}
	method = strings.ToUpper(method)

	urlStr, ok := params["url"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'url' is required and must be a string")
	}

	var bodyReader io.Reader
	if body, ok := params["body"].(string); ok {
		bodyReader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if headers, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if val, ok := v.(string); ok {
				req.Header.Set(k, val)
			}
		}
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

func init() {
	server := inprocess.GetInProcessServer()
	err := server.RegisterServiceMethod("sire:local/http.request", Request)
	if err != nil {
		panic(err) // Panics if registration fails
	}
}
