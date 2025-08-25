package file

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileReadNode(t *testing.T) {
	// Create a temporary file with content
	tmpfile, err := os.CreateTemp("", "test-read-*.txt")
	require.NoError(t, err)
	defer func() { assert.NoError(t, os.Remove(tmpfile.Name())) }() // clean up

	content := "hello world"
	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	config := map[string]interface{}{
		"path": tmpfile.Name(),
	}

	node, err := NewFileReadNode(config)
	require.NoError(t, err)

	output, err := node.Execute(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, content, output["content"])
}

func TestFileWriteNode(t *testing.T) {
	// Create a temporary file path
	tmpfile, err := os.CreateTemp("", "test-write-*.txt")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close()) // close it, we just need the name
	defer func() { assert.NoError(t, os.Remove(tmpfile.Name())) }()

	config := map[string]interface{}{
		"path": tmpfile.Name(),
	}

	node, err := NewFileWriteNode(config)
	require.NoError(t, err)

	inputs := map[string]interface{}{
		"content": "hello from test",
	}

	_, err = node.Execute(context.Background(), inputs)
	require.NoError(t, err)

	// Verify content
	data, err := os.ReadFile(tmpfile.Name())
	require.NoError(t, err)
	assert.Equal(t, "hello from test", string(data))
}
