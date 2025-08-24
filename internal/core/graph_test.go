package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopologicalSort(t *testing.T) {
	t.Run("simple linear workflow", func(t *testing.T) {
		nodes := map[string]Node{
			"node-1": &MockNode{},
			"node-2": &MockNode{},
			"node-3": &MockNode{},
		}
		edges := []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-2", To: "node-3"},
		}

		sorted, err := topologicalSort(nodes, edges)
		require.NoError(t, err)
		assert.Equal(t, []string{"node-1", "node-2", "node-3"}, sorted)
	})

	t.Run("workflow with a branch", func(t *testing.T) {
		nodes := map[string]Node{
			"node-1": &MockNode{},
			"node-2": &MockNode{},
			"node-3": &MockNode{},
			"node-4": &MockNode{},
		}
		edges := []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-1", To: "node-3"},
			{From: "node-2", To: "node-4"},
			{From: "node-3", To: "node-4"},
		}

		sorted, err := topologicalSort(nodes, edges)
		require.NoError(t, err)
		// The exact order of node-2 and node-3 can vary, so we check for both possibilities.
		assert.True(t, (sorted[1] == "node-2" && sorted[2] == "node-3") || (sorted[1] == "node-3" && sorted[2] == "node-2"))
		assert.Equal(t, "node-1", sorted[0])
		assert.Equal(t, "node-4", sorted[3])
	})

	t.Run("workflow with a cycle", func(t *testing.T) {
		nodes := map[string]Node{
			"node-1": &MockNode{},
			"node-2": &MockNode{},
			"node-3": &MockNode{},
		}
		edges := []Edge{
			{From: "node-1", To: "node-2"},
			{From: "node-2", To: "node-3"},
			{From: "node-3", To: "node-1"},
		}

		_, err := topologicalSort(nodes, edges)
		require.Error(t, err)
		assert.Equal(t, "workflow has a cycle", err.Error())
	})
}
