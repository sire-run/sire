package file

import (
	"context"
	"os"
	"testing"

	"github.com/sire-run/sire/internal/mcp/inprocess" // Import the inprocess package
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRead(t *testing.T) {
	// Create a temporary file with content
	tmpfile, err := os.CreateTemp("", "test-read-*.txt")
	require.NoError(t, err)
	defer func() { assert.NoError(t, os.Remove(tmpfile.Name())) }() // clean up

	content := "hello world"
	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	dispatcher := inprocess.NewInProcessDispatcher()

	params := map[string]interface{}{
		"path": tmpfile.Name(),
	}

	output, err := dispatcher.Dispatch(context.Background(), "sire:local/file.read", params)
	require.NoError(t, err)

	assert.Equal(t, content, output["content"])
}

func TestFileWrite(t *testing.T) {
	// Create a temporary file path
	tmpfile, err := os.CreateTemp("", "test-write-*.txt")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close()) // close it, we just need the name
	defer func() { assert.NoError(t, os.Remove(tmpfile.Name())) }()

	dispatcher := inprocess.NewInProcessDispatcher()

	params := map[string]interface{}{
		"path":    tmpfile.Name(),
		"content": "hello from test",
	}

	_, err = dispatcher.Dispatch(context.Background(), "sire:local/file.write", params)
	require.NoError(t, err)

	// Verify content
	data, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	assert.Equal(t, "hello from test", string(data))
}
