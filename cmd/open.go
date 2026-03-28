package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <title>",
	Short: "Open a note in Obsidian",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		vaultName := filepath.Base(cfg.VaultPath)
		slug := notes.Slug(args[0])
		uri := fmt.Sprintf("obsidian://open?vault=%s&file=%s",
			url.QueryEscape(vaultName),
			url.QueryEscape(slug),
		)
		fmt.Fprintf(os.Stderr, "note: opening %q in Obsidian\n", args[0])
		return openURI(uri)
	},
}

func openURI(uri string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", uri).Run()
	case "windows":
		return exec.Command("cmd", "/c", "start", "", uri).Run()
	default:
		return exec.Command("xdg-open", uri).Run()
	}
}

func init() {
	rootCmd.AddCommand(openCmd)
}
