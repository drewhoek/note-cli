package cmd

import (
	"fmt"
	"os"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag <title>",
	Short: "List, add, or remove tags on a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		tags, err := notes.ReadTags(cfg.VaultPath, args[0])
		if err != nil {
			return err
		}
		for _, t := range tags {
			fmt.Println(t)
		}
		return nil
	},
}

var tagAddCmd = &cobra.Command{
	Use:   "add <title> <tag>",
	Short: "Add a tag to a note",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		added, err := notes.AddTag(cfg.VaultPath, args[0], args[1])
		if err != nil {
			return err
		}
		if added {
			fmt.Fprintf(os.Stderr, "note: added tag %q to %q\n", args[1], args[0])
		} else {
			fmt.Fprintf(os.Stderr, "note: %q already has tag %q\n", args[0], args[1])
		}
		return nil
	},
}

var tagRemoveCmd = &cobra.Command{
	Use:   "remove <title> <tag>",
	Short: "Remove a tag from a note",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		removed, err := notes.RemoveTag(cfg.VaultPath, args[0], args[1])
		if err != nil {
			return err
		}
		if removed {
			fmt.Fprintf(os.Stderr, "note: removed tag %q from %q\n", args[1], args[0])
		} else {
			fmt.Fprintf(os.Stderr, "note: %q did not have tag %q\n", args[0], args[1])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
	tagCmd.AddCommand(tagAddCmd)
	tagCmd.AddCommand(tagRemoveCmd)
}
