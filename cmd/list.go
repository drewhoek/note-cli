package cmd

import (
	"fmt"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var listTag string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes, optionally filtered by tag",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		titles, err := notes.List(cfg.VaultPath, listTag)
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
}
