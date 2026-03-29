package cmd

import (
	"fmt"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var relatedCmd = &cobra.Command{
	Use:   "related <title>",
	Short: "Find notes connected by wikilinks or shared tags",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := notes.ResolveTitle(args[0])
		results, err := notes.Related(cfg.VaultPath, title)
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
	rootCmd.AddCommand(relatedCmd)
}
