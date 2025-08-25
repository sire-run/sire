package file

import (
	"context"
	"os"
	"testing"

	"github.com/sire-run/sire/internal/mcp/inprocess" // Import the inprocess package
)

func TestFileRead(t *testing.T) {
	// Create a temporary file with content
	tmpfile, err := os.CreateTemp("", "test-read-*.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}() // clean up

	content := "hello world"
	_, err = tmpfile.Write([]byte(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dispatcher := inprocess.NewInProcessDispatcher()

	params := map[string]interface{}{
		"path": tmpfile.Name(),
	}

	output, err := dispatcher.Dispatch(context.Background(), "sire:local/file.read", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output["content"] != content {
		t.Errorf("expected content %q, got %q", content, output["content"])
	}
}

func TestFileWrite(t *testing.T) {
	// Create a temporary file path
	tmpfile, err := os.CreateTemp("", "test-write-*.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tmpfile.Close(); err != nil { // close it, we just need the name
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	dispatcher := inprocess.NewInProcessDispatcher()

	params := map[string]interface{}{
		"path":    tmpfile.Name(),
		"content": "hello from test",
	}

	_, err = dispatcher.Dispatch(context.Background(), "sire:local/file.write", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify content
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "hello from test" {
		t.Errorf("expected content %q, got %q", "hello from test", string(data))
	}
}
