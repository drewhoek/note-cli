package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var appendCmd = &cobra.Command{
	Use:   "append <title> [content]",
	Short: "Append content to a note (reads from stdin if content is omitted)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var content string
		if len(args) == 2 {
			content = args[1]
		} else {
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			content = strings.Join(lines, "\n")
		}

		if err := notes.Append(cfg.VaultPath, args[0], content); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: appended to %q\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(appendCmd)
}
