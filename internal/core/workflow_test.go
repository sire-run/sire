package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWorkflow_YAMLUnmarshal(t *testing.T) {
	yamlData := `
id: my-workflow
name: My Test Workflow
steps:
  - id: step1
    tool: "sire:local/http.request"
    params:
      url: "https://example.com"
  - id: step2
    tool: "mcp:http://remote/rpc#file.write"
    params:
      path: "/tmp/test.txt"
      content: "hello"
    retry:
      max_attempts: 3
      backoff: "exponential"
edges:
  - from: step1
    to: step2
`
	var wf Workflow
	err := yaml.Unmarshal([]byte(yamlData), &wf)
	require.NoError(t, err)

	assert.Equal(t, "my-workflow", wf.ID)
	assert.Equal(t, "My Test Workflow", wf.Name)
	require.Len(t, wf.Steps, 2)
	assert.Equal(t, "step1", wf.Steps[0].ID)
	assert.Equal(t, "sire:local/http.request", wf.Steps[0].Tool)
	assert.Equal(t, "https://example.com", wf.Steps[0].Params["url"])
	assert.Nil(t, wf.Steps[0].Retry)

	assert.Equal(t, "step2", wf.Steps[1].ID)
	assert.Equal(t, "mcp:http://remote/rpc#file.write", wf.Steps[1].Tool)
	assert.NotNil(t, wf.Steps[1].Retry)
	assert.Equal(t, 3, wf.Steps[1].Retry.MaxAttempts)

	require.Len(t, wf.Edges, 1)
	assert.Equal(t, "step1", wf.Edges[0].From)
	assert.Equal(t, "step2", wf.Edges[0].To)
}
