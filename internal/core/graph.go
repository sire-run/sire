package core

import "fmt"

// a simple implementation of Kahn's algorithm for topological sorting.
func topologicalSort(nodes map[string]Node, edges []Edge) ([]string, error) {
	// 1. Calculate in-degrees
	inDegree := make(map[string]int)
	for id := range nodes {
		inDegree[id] = 0
	}
	for _, edge := range edges {
		inDegree[edge.To]++
	}

	// 2. Initialize queue with nodes with in-degree 0
	queue := []string{}
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// 3. Process queue
	result := []string{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		result = append(result, id)

		// 4. Decrement in-degrees of neighbors
		for _, edge := range edges {
			if edge.From == id {
				inDegree[edge.To]--
				if inDegree[edge.To] == 0 {
					queue = append(queue, edge.To)
				}
			}
		}
	}

	// 5. Check for cycles
	if len(result) != len(nodes) {
		return nil, fmt.Errorf("workflow has a cycle")
	}

	return result, nil
}
