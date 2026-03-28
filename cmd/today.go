package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Open or create today's daily note",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := time.Now().Format("2006-01-02")
		err = notes.CreateDaily(cfg.VaultPath, title)
		if err != nil && !errors.Is(err, notes.ErrNoteExists) {
			return err
		}
		fmt.Fprintln(os.Stderr, "note: "+title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(todayCmd)
}
