package transform

import (
	"context"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/sire-run/sire/internal/core"
)

// DataTransformNode performs data transformation operations.
type DataTransformNode struct {
	operation  string
	expression string
	initial    interface{}
}

// NewDataTransformNode creates a new DataTransformNode.
func NewDataTransformNode(config map[string]interface{}) (core.Node, error) {
	node := &DataTransformNode{}

	op, ok := config["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("config 'operation' is required and must be a string")
	}
	node.operation = op

	exp, ok := config["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("config 'expression' is required and must be a string")
	}
	node.expression = exp

	if initial, ok := config["initial"]; ok {
		node.initial = initial
	}

	return node, nil
}

// Execute performs the data transformation.
func (n *DataTransformNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	data, ok := inputs["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("input 'data' is required and must be an array")
	}

	var result interface{}
	var err error

	switch n.operation {
	case "map":
		result, err = n.mapData(data)
	case "filter":
		result, err = n.filterData(data)
	case "reduce":
		result, err = n.reduceData(data)
	default:
		err = fmt.Errorf("unsupported operation: %s", n.operation)
	}

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"result": result}, nil
}

func (n *DataTransformNode) mapData(data []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(data))
	for i, item := range data {
		env := map[string]interface{}{"item": item}
		output, err := expr.Eval(n.expression, env)
		if err != nil {
			return nil, fmt.Errorf("map expression error: %w", err)
		}
		result[i] = output
	}
	return result, nil
}

func (n *DataTransformNode) filterData(data []interface{}) ([]interface{}, error) {
	result := []interface{}{}
	for _, item := range data {
		env := map[string]interface{}{"item": item}
		output, err := expr.Eval(n.expression, env)
		if err != nil {
			return nil, fmt.Errorf("filter expression error: %w", err)
		}
		if val, ok := output.(bool); ok && val {
			result = append(result, item)
		}
	}
	return result, nil
}

func (n *DataTransformNode) reduceData(data []interface{}) (interface{}, error) {
	acc := n.initial
	for _, item := range data {
		env := map[string]interface{}{"acc": acc, "item": item}
		output, err := expr.Eval(n.expression, env)
		if err != nil {
			return nil, fmt.Errorf("reduce expression error: %w", err)
		}
		acc = output
	}
	return acc, nil
}

func init() {
	core.RegisterNode("data.transform", NewDataTransformNode)
}
