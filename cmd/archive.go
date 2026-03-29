package cmd

import (
	"fmt"
	"os"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <title>",
	Short: "Move a note to the archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := notes.ResolveTitle(args[0])
		if err := notes.Archive(cfg.VaultPath, title); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: archived %q\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}
