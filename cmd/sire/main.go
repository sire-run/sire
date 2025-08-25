package main

import (
	"fmt"
	"os"
	"text/tabwriter" // New import for formatted output
	"time"           // New import for time.Format

	"github.com/sire-run/sire/internal/storage" // New import for storage
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sire",
	Short: "Sire is a Go-based workflow automation platform.",
	Long:  `A fast and flexible workflow automation platform built in Go.`,
}

// executionCmd represents the base command for execution-related operations
var executionCmd = &cobra.Command{
	Use:   "execution",
	Short: "Manage workflow executions",
	Long:  `Commands to list, view status, and manage workflow executions.`,
}

var dbPath // Global variable for DB path

// listCmd represents the list execution command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workflow executions",
	Run: func(cmd *cobra.Command, args []string) {
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

		executions, err := store.ListPendingExecutions() // Assuming ListPendingExecutions lists all for now
		if err != nil {
			fmt.Printf("Error listing executions: %v\n", err)
			os.Exit(1)
		}

		if len(executions) == 0 {
			fmt.Println("No executions found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tWORKFLOW ID\tSTATUS\tCREATED AT\tUPDATED AT")
		for _, exec := range executions {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				exec.ID,
				exec.WorkflowID,
				exec.Status,
				exec.CreatedAt.Format(time.RFC3339),
				exec.UpdatedAt.Format(time.RFC3339),
			)
		}
		w.Flush()
	},
}

// statusCmd represents the status execution command
var statusCmd = &cobra.Command{
	Use:   "status [execution-id]",
	Short: "View the status of a specific workflow execution",
	Args:  cobra.ExactArgs(1), // Requires exactly one argument (execution ID)
	Run: func(cmd *cobra.Command, args []string) {
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

		executionID := args[0]
		exec, err := store.LoadExecution(executionID)
		if err != nil {
			fmt.Printf("Error loading execution %s: %v\n", executionID, err)
			os.Exit(1)
		}

		fmt.Printf("Execution ID: %s\n", exec.ID)
		fmt.Printf("Workflow ID: %s\n", exec.WorkflowID)
		fmt.Printf("Status: %s\n", exec.Status)
		fmt.Printf("Created At: %s\n", exec.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated At: %s\n", exec.UpdatedAt.Format(time.RFC3339))
		fmt.Println("\nStep States:")

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "STEP ID\tSTATUS\tATTEMPTS\tERROR")
		for stepID, stepState := range exec.StepStates {
			errmsg := ""
			if stepState.Error != "" {
				errmsg = stepState.Error
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
				stepID,
				stepState.Status,
				stepState.Attempts,
				errmsg,
			)
		}
		w.Flush()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(workflowCmd) // workflowCmd is defined in workflow.go
	rootCmd.AddCommand(executionCmd)

	executionCmd.AddCommand(listCmd)
	executionCmd.AddCommand(statusCmd)

	// Add db-path flag to execution commands
	executionCmd.PersistentFlags().StringVarP(&dbPath, "db-path", "d", "sire.db", "Path to the BoltDB file for state persistence")
}
