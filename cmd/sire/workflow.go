package main

import (
	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage workflows",
}

func init() {
	rootCmd.AddCommand(workflowCmd)
}
