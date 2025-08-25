package core

import (
	"testing"

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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wf.ID != "my-workflow" {
		t.Errorf("expected ID %q, got %q", "my-workflow", wf.ID)
	}
	if wf.Name != "My Test Workflow" {
		t.Errorf("expected Name %q, got %q", "My Test Workflow", wf.Name)
	}
	if len(wf.Steps) != 2 {
		t.Fatalf("expected %d steps, got %d", 2, len(wf.Steps))
	}
	if wf.Steps[0].ID != "step1" {
		t.Errorf("expected step1 ID %q, got %q", "step1", wf.Steps[0].ID)
	}
	if wf.Steps[0].Tool != "sire:local/http.request" {
		t.Errorf("expected step1 Tool %q, got %q", "sire:local/http.request", wf.Steps[0].Tool)
	}
	if wf.Steps[0].Params["url"] != "https://example.com" {
		t.Errorf("expected step1 Params url %q, got %q", "https://example.com", wf.Steps[0].Params["url"])
	}
	if wf.Steps[0].Retry != nil {
		t.Errorf("expected step1 Retry to be nil, got %v", wf.Steps[0].Retry)
	}

	if wf.Steps[1].ID != "step2" {
		t.Errorf("expected step2 ID %q, got %q", "step2", wf.Steps[1].ID)
	}
	if wf.Steps[1].Tool != "mcp:http://remote/rpc#file.write" {
		t.Errorf("expected step2 Tool %q, got %q", "mcp:http://remote/rpc#file.write", wf.Steps[1].Tool)
	}
	if wf.Steps[1].Retry == nil {
		t.Errorf("expected step2 Retry to not be nil")
	}
	if wf.Steps[1].Retry.MaxAttempts != 3 {
		t.Errorf("expected step2 MaxAttempts %d, got %d", 3, wf.Steps[1].Retry.MaxAttempts)
	}

	if len(wf.Edges) != 1 {
		t.Fatalf("expected %d edges, got %d", 1, len(wf.Edges))
	}
	if wf.Edges[0].From != "step1" {
		t.Errorf("expected edge From %q, got %q", "step1", wf.Edges[0].From)
	}
	if wf.Edges[0].To != "step2" {
		t.Errorf("expected edge To %q, got %q", "step2", wf.Edges[0].To)
	}
}
