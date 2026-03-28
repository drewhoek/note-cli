package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "note",
	Short:         "Manage notes in an Obsidian vault",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command. Called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "note: %s\n", err)
		os.Exit(1)
	}
}
