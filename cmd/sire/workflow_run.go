package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sire-run/sire/internal/core"
	"github.com/sire-run/sire/internal/mcp/inprocess" // New import
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	runFile   string
	runInputs string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a workflow",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Read workflow file
		data, err := os.ReadFile(runFile)
		if err != nil {
			fmt.Printf("Error reading workflow file: %v\n", err)
			os.Exit(1)
		}

		// 2. Parse workflow
		var workflow core.Workflow
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			fmt.Printf("Error parsing workflow file: %v\n", err)
			os.Exit(1)
		}

		// 3. Parse inputs
		var inputs map[string]interface{}
		if runInputs != "" {
			if err := json.Unmarshal([]byte(runInputs), &inputs); err != nil {
				fmt.Printf("Error parsing inputs: %v\n", err)
				os.Exit(1)
			}
		}

		// 4. Execute workflow
		// Instantiate the in-process dispatcher
		dispatcher := inprocess.NewInProcessDispatcher()
		engine := core.NewEngine(dispatcher)

		execution, err := engine.Execute(context.Background(), &workflow, inputs)
		if err != nil {
			fmt.Printf("Error executing workflow: %v\n", err)
			os.Exit(1)
		}

		// 5. Print output
		outputJSON, err := json.MarshalIndent(execution, "", "  ")
		if err != nil {
			fmt.Printf("Error marshalling execution output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(outputJSON))
	},
}

func init() {
	workflowCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&runFile, "file", "f", "", "Path to the workflow file (YAML or JSON)")
	if err := runCmd.MarkFlagRequired("file"); err != nil {
		fmt.Printf("Error marking flag as required: %v\n", err)
		os.Exit(1)
	}
	runCmd.Flags().StringVarP(&runInputs, "inputs", "i", "", "JSON string of inputs to the workflow")
}
