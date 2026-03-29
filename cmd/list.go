package cmd

import (
	"fmt"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var listTag string
var listStatus string
var listIncludeArchived bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes, optionally filtered by tag or status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		titles, err := notes.List(cfg.VaultPath, listTag, listStatus, listIncludeArchived)
		if err != nil {
			return err
		}
		for _, t := range titles {
			fmt.Println(t)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listTag, "tag", "", "filter by tag")
	listCmd.Flags().StringVar(&listStatus, "status", "", "filter by status (active, stale, reference)")
	listCmd.Flags().BoolVar(&listIncludeArchived, "include-archived", false, "include notes in the archive/ subdirectory")
}
