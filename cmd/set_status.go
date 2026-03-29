package cmd

import (
	"fmt"
	"os"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var setStatusCmd = &cobra.Command{
	Use:   "set-status <title> <status>",
	Short: "Set a note's lifecycle status (active, stale, reference)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := notes.ResolveTitle(args[0])
		status := args[1]
		if status != "active" && status != "stale" && status != "reference" && status != "" {
			return fmt.Errorf("invalid status %q: must be active, stale, or reference", status)
		}
		if err := notes.SetStatus(cfg.VaultPath, title, status); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: set status of %q to %q\n", title, status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setStatusCmd)
}
