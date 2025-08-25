package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/sire-run/sire/internal/mcp/inprocess" // Import the inprocess package
)

// assertJSONEq compares two JSON strings for equality.
func assertJSONEq(t *testing.T, expected, actual string) bool {
	t.Helper()

	var expectedParsed, actualParsed interface{}

	err := json.Unmarshal([]byte(expected), &expectedParsed)
	if err != nil {
		t.Errorf("failed to unmarshal expected JSON: %v", err)
		return false
	}

	err = json.Unmarshal([]byte(actual), &actualParsed)
	if err != nil {
		t.Errorf("failed to unmarshal actual JSON: %v", err)
		return false
	}

	if !reflect.DeepEqual(expectedParsed, actualParsed) {
		t.Errorf("expected JSON %q, got %q", expected, actual)
		return false
	}
	return true
}

func TestHTTPRequest_Execute(t *testing.T) {
	t.Run("simple GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected method %q, got %q", "GET", r.Method)
			}
			if r.Header.Get("X-My-Header") != "my-value" {
				t.Errorf("expected header %q, got %q", "my-value", r.Header.Get("X-My-Header"))
			}
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"response": "ok"}`))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if output["statusCode"] != 200 {
			t.Errorf("expected status code %d, got %v", 200, output["statusCode"])
		}
		assertJSONEq(t, `{"response": "ok"}`, output["body"].(string))
	})

	t.Run("simple POST request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected method %q, got %q", "POST", r.Method)
			}
			body, _ := io.ReadAll(r.Body)
			assertJSONEq(t, `{"key": "value"}`, string(body))
			w.WriteHeader(http.StatusCreated)
			_, err := w.Write([]byte(`{"id": 123}`))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}))
		defer server.Close()

		dispatcher := inprocess.NewInProcessDispatcher()

		params := map[string]interface{}{
			"method": "POST",
			"url":    server.URL,
			"body":   `{"key": "value"}`,
		}

		output, err := dispatcher.Dispatch(context.Background(), "sire:local/http.request", params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if output["statusCode"] != 201 {
			t.Errorf("expected status code %d, got %v", 201, output["statusCode"])
		}
		assertJSONEq(t, `{"id": 123}`, output["body"].(string))
	})
}
