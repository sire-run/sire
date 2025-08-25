package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/sire-run/sire/internal/mcp/inprocess" // Import the inprocess package
)

func TestTransform_Map(t *testing.T) {
	dispatcher := inprocess.NewInProcessDispatcher()

	params := map[string]interface{}{
		"operation":  "map",
		"expression": "item * 2",
		"data":       []interface{}{1, 2, 3, 4},
	}

	output, err := dispatcher.Dispatch(context.Background(), "sire:local/data.transform", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []interface{}{2, 4, 6, 8}
	actual := output["result"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestTransform_Filter(t *testing.T) {
	dispatcher := inprocess.NewInProcessDispatcher()

	params := map[string]interface{}{
		"operation":  "filter",
		"expression": "item > 2",
		"data":       []interface{}{1, 2, 3, 4},
	}

	output, err := dispatcher.Dispatch(context.Background(), "sire:local/data.transform", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []interface{}{3, 4}
	actual := output["result"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

	expected := float64(10)
	if actualResult != expected {
		t.Errorf("expected %v, got %v", expected, actualResult)
	}
}
