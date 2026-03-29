package cmd

import (
	"fmt"
	"os"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <old-title> <new-title>",
	Short: "Rename a note and update all wikilinks to it",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := notes.Rename(cfg.VaultPath, args[0], args[1]); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: renamed %q → %q\n", args[0], args[1])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
