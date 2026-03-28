package cmd

import (
	"fmt"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <title>",
	Short: "Print the content of a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		content, err := notes.Read(cfg.VaultPath, args[0])
		if err != nil {
			return err
		}
		fmt.Print(content)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
