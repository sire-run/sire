package transform

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataTransformNode_Map(t *testing.T) {
	config := map[string]interface{}{
		"operation":  "map",
		"expression": "item * 2",
	}

	node, err := NewDataTransformNode(config)
	require.NoError(t, err)

	inputs := map[string]interface{}{
		"data": []interface{}{1, 2, 3, 4},
	}

	output, err := node.Execute(context.Background(), inputs)
	require.NoError(t, err)

	assert.Equal(t, []interface{}{2, 4, 6, 8}, output["result"])
}

func TestDataTransformNode_Filter(t *testing.T) {
	config := map[string]interface{}{
		"operation":  "filter",
		"expression": "item > 2",
	}

	node, err := NewDataTransformNode(config)
	require.NoError(t, err)

	inputs := map[string]interface{}{
		"data": []interface{}{1, 2, 3, 4},
	}

	output, err := node.Execute(context.Background(), inputs)
	require.NoError(t, err)

	assert.Equal(t, []interface{}{3, 4}, output["result"])
}

func TestDataTransformNode_Reduce(t *testing.T) {
	config := map[string]interface{}{
		"operation":  "reduce",
		"expression": "acc + item",
		"initial":    0,
	}

	node, err := NewDataTransformNode(config)
	require.NoError(t, err)

	inputs := map[string]interface{}{
		"data": []interface{}{1, 2, 3, 4},
	}

	output, err := node.Execute(context.Background(), inputs)
	require.NoError(t, err)

	assert.Equal(t, 10, output["result"])
}
