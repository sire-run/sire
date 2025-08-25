package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sire-run/sire/internal/core"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var validateFile string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a workflow file",
	Run: func(cmd *cobra.Command, args []string) {
		absValidateFile, err := filepath.Abs(validateFile)
		if err != nil {
			fmt.Printf("Error resolving workflow file path: %v\n", err)
			os.Exit(1)
		}
		// Clean the path to remove any ../ or ./ components.
		// Note: This does not prevent directory traversal if the initial path is outside
		// the intended working directory. A more robust solution would involve
		// checking if the cleaned path is within a trusted base directory.
		cleanedValidateFile := filepath.Clean(absValidateFile)
		data, err := os.ReadFile(cleanedValidateFile)
		if err != nil {
			fmt.Printf("Error reading workflow file: %v\n", err)
			os.Exit(1)
		}

		var workflow core.Workflow
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			fmt.Printf("Error parsing workflow file: %v\n", err)
			os.Exit(1)
		}

		// Basic validation is done by unmarshalling. We can add more complex validation here later,
		// like checking for circular dependencies, which is already handled by the engine's topological sort.

		fmt.Println("Workflow file is valid.")
	},
}

func init() {
	workflowCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVarP(&validateFile, "file", "f", "", "Path to the workflow file (YAML or JSON)")
	if err := validateCmd.MarkFlagRequired("file"); err != nil {
		fmt.Printf("Error marking flag as required: %v\n", err)
		os.Exit(1)
	}
}
