package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sire-run/sire/internal/mcp/inprocess" // Import the inprocess package
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPRequest_Execute(t *testing.T) {
	t.Run("simple GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "my-value", r.Header.Get("X-My-Header"))
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"response": "ok"}`))
			assert.NoError(t, err)
		}))
		defer server.Close()

		dispatcher := inprocess.NewInProcessDispatcher()

		params := map[string]interface{}{
			"method": "GET",
			"url":    server.URL,
			"headers": map[string]interface{}{
				"X-My-Header": "my-value",
			},
		}

		output, err := dispatcher.Dispatch(context.Background(), "sire:local/http.request", params)
		require.NoError(t, err)

		assert.Equal(t, 200, output["statusCode"])
		assert.JSONEq(t, `{"response": "ok"}`, output["body"].(string))
	})

	t.Run("simple POST request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			body, _ := io.ReadAll(r.Body)
			assert.JSONEq(t, `{"key": "value"}`, string(body))
			w.WriteHeader(http.StatusCreated)
			_, err := w.Write([]byte(`{"id": 123}`))
			assert.NoError(t, err)
		}))
		defer server.Close()

		dispatcher := inprocess.NewInProcessDispatcher()

		params := map[string]interface{}{
			"method": "POST",
			"url":    server.URL,
			"body":   `{"key": "value"}`,
		}

		output, err := dispatcher.Dispatch(context.Background(), "sire:local/http.request", params)
		require.NoError(t, err)

		assert.Equal(t, 201, output["statusCode"])
		assert.JSONEq(t, `{"id": 123}`, output["body"].(string))
	})
}
