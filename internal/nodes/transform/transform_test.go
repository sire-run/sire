package transform

import (
	"context"
	"testing"

	"github.com/sire-run/sire/internal/mcp/inprocess" // Import the inprocess package
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransform_Map(t *testing.T) {
	dispatcher := inprocess.NewInProcessDispatcher()

	params := map[string]interface{}{
		"operation":  "map",
		"expression": "item * 2",
		"data":       []interface{}{1, 2, 3, 4},
	}

	output, err := dispatcher.Dispatch(context.Background(), "sire:local/data.transform", params)
	require.NoError(t, err)

	assert.Equal(t, []interface{}{2, 4, 6, 8}, output["result"])
}

func TestTransform_Filter(t *testing.T) {
	dispatcher := inprocess.NewInProcessDispatcher()

	params := map[string]interface{}{
		"operation":  "filter",
		"expression": "item > 2",
		"data":       []interface{}{1, 2, 3, 4},
	}

	output, err := dispatcher.Dispatch(context.Background(), "sire:local/data.transform", params)
	require.NoError(t, err)

	assert.Equal(t, []interface{}{3, 4}, output["result"])
}

func TestTransform_Reduce(t *testing.T) {
	dispatcher := inprocess.NewInProcessDispatcher()

	params := map[string]interface{}{
		"operation":  "reduce",
		"expression": "acc + item",
		"initial":    0,
		"data":       []interface{}{1, 2, 3, 4},
	}

	output, err := dispatcher.Dispatch(context.Background(), "sire:local/data.transform", params)
	require.NoError(t, err)

	// Convert the actual result to float64 for comparison, as expr can return float64 for integer results.
	actualResult, ok := output["result"].(float64)
	if !ok {
		// If it's an int, convert it to float64
		if intResult, isInt := output["result"].(int); isInt {
			actualResult = float64(intResult)
		} else {
			t.Fatalf("unexpected type for result: %T", output["result"])
		}
	}

	assert.Equal(t, float64(10), actualResult)
}
