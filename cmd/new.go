package cmd

import (
	"fmt"
	"os"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <title>",
	Short: "Create a new note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := notes.Create(cfg.VaultPath, args[0]); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: created %q\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
