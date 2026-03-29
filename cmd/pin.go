package cmd

import (
	"fmt"
	"os"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var pinCmd = &cobra.Command{
	Use:   "pin <title>",
	Short: "Pin a note so it appears in context output with full content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := notes.ResolveTitle(args[0])
		if err := notes.Pin(cfg.VaultPath, title); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: pinned %q\n", title)
		return nil
	},
}

var unpinCmd = &cobra.Command{
	Use:   "unpin <title>",
	Short: "Unpin a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := notes.ResolveTitle(args[0])
		if err := notes.Unpin(cfg.VaultPath, title); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: unpinned %q\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pinCmd)
	rootCmd.AddCommand(unpinCmd)
}
