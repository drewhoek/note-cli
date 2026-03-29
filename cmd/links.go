package cmd

import (
	"fmt"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var linksCmd = &cobra.Command{
	Use:   "links <title>",
	Short: "List all [[wikilinks]] in a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		results, err := notes.Outlinks(cfg.VaultPath, notes.ResolveTitle(args[0]))
		if err != nil {
			return err
		}
		for _, l := range results {
			fmt.Println(l)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(linksCmd)
}
