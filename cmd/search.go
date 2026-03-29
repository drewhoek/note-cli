package cmd

import (
	"fmt"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var searchExact bool
var searchVerbose bool
var searchIncludeArchived bool

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search note titles and content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		results, err := notes.SearchDetailed(cfg.VaultPath, args[0], searchExact, searchIncludeArchived)
		if err != nil {
			return err
		}
		for _, r := range results {
			if searchVerbose {
				fmt.Printf("%s\n  %s\n\n", r.Slug, r.Snippet)
			} else {
				fmt.Println(r.Slug)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolVar(&searchExact, "exact", false, "use exact substring matching instead of fuzzy")
	searchCmd.Flags().BoolVar(&searchVerbose, "verbose", false, "show a content snippet for each result")
	searchCmd.Flags().BoolVar(&searchIncludeArchived, "include-archived", false, "include archived notes in results")
}
