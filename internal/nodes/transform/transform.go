package transform

import (
	"context"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/sire-run/sire/internal/mcp/inprocess" // Import the inprocess package
)

// Transform performs data transformation operations.
func Transform(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'operation' is required and must be a string")
	}
	expression, ok := params["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("parameter 'expression' is required and must be a string")
	}
	data, ok := params["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("parameter 'data' is required and must be an array")
	}

	var initial interface{}
	if initVal, ok := params["initial"]; ok {
		initial = initVal
	}

	var result interface{}
	var err error

	switch operation {
	case "map":
		result, err = mapData(data, expression)
	case "filter":
		result, err = filterData(data, expression)
	case "reduce":
		result, err = reduceData(data, expression, initial)
	default:
		err = fmt.Errorf("unsupported operation: %s", operation)
	}

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"result": result}, nil
}

func mapData(data []interface{}, expression string) ([]interface{}, error) {
	result := make([]interface{}, len(data))
	for i, item := range data {
		env := map[string]interface{}{"item": item}
		output, err := expr.Eval(expression, env)
		if err != nil {
			return nil, fmt.Errorf("map expression error: %w", err)
		}
		result[i] = output
	}
	return result, nil
}

func filterData(data []interface{}, expression string) ([]interface{}, error) {
	result := []interface{}{}
	for _, item := range data {
		env := map[string]interface{}{"item": item}
		output, err := expr.Eval(expression, env)
		if err != nil {
			return nil, fmt.Errorf("filter expression error: %w", err)
		}
		if val, ok := output.(bool); ok && val {
			result = append(result, item)
		}
	}
	return result, nil
}

func reduceData(data []interface{}, expression string, initial interface{}) (interface{}, error) {
	acc := initial
	for _, item := range data {
		env := map[string]interface{}{"acc": acc, "item": item}
		output, err := expr.Eval(expression, env)
		if err != nil {
			return nil, fmt.Errorf("reduce expression error: %w", err)
		}
		acc = output
	}
	return acc, nil
}

func init() {
	server := inprocess.GetInProcessServer()
	err := server.RegisterServiceMethod("sire:local/data.transform", Transform)
	if err != nil {
		panic(err) // Panics if registration fails
	}
}