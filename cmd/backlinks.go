package cmd

import (
	"fmt"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var backlinksCmd = &cobra.Command{
	Use:   "backlinks <title>",
	Short: "List notes that link to this note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		results, err := notes.Backlinks(cfg.VaultPath, args[0])
		if err != nil {
			return err
		}
		for _, r := range results {
			fmt.Println(r)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backlinksCmd)
}
