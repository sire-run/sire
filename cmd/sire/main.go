package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sire",
	Short: "Sire is a Go-based workflow automation platform.",
	Long:  `A fast and flexible workflow automation platform built in Go.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
