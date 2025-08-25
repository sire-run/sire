package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time" // New import for time.Now()
	"github.com/google/uuid" // New import for generating UUIDs

	"github.com/sire-run/sire/internal/core"
	"github.com/sire-run/sire/internal/mcp/inprocess"
	"github.com/sire-run/sire/internal/storage" // New import for storage
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	runFile   string
	runInputs string
	dbPath    string // New global variable for DB path
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

		// 4. Initialize storage
		store, err := storage.NewBoltDBStore(dbPath)
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := store.Close(); err != nil {
				fmt.Printf("Error closing database: %v\n", err)
			}
		}()

		// 5. Create a new execution record
		executionID := uuid.New().String()
		execution := &core.Execution{
			ID:         executionID,
			WorkflowID: workflow.ID,
			Workflow:   &workflow, // Store the workflow definition
			Status:     core.ExecutionStatusRunning,
			StepStates: make(map[string]*core.StepState),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := store.SaveExecution(execution); err != nil {
			fmt.Printf("Error saving new execution: %v\n", err)
			os.Exit(1)
		}


		// 6. Execute workflow
		dispatcher := inprocess.NewInProcessDispatcher()
		// The engine will now take the store as well (part of S9.2.2)
		// For now, we'll just pass the dispatcher. The engine will be refactored later.
		engine := core.NewEngine(dispatcher, store) // Pass store to NewEngine

		// Pass the initial execution to the engine
		execution, err = engine.Execute(context.Background(), execution, &workflow, inputs) // Pass execution object
		if err != nil {
			fmt.Printf("Error executing workflow: %v\n", err)
			os.Exit(1)
		}

		// 7. Print output
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
	runCmd.Flags().StringVarP(&dbPath, "db-path", "d", "sire.db", "Path to the BoltDB file for state persistence") // New flag
}