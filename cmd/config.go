package cmd

import (
	"fmt"
	"os"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage note configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set vault-path <path>",
	Short: "Set the path to your Obsidian vault",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "vault-path" {
			return fmt.Errorf("unknown config key %q", args[0])
		}
		vaultPath := args[1]
		if _, err := os.Stat(vaultPath); err != nil {
			return fmt.Errorf("vault path does not exist: %s", vaultPath)
		}
		cfg := &config.Config{VaultPath: vaultPath}
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: vault-path set to %s\n", vaultPath)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get vault-path",
	Short: "Get a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "vault-path" {
			return fmt.Errorf("unknown config key %q", args[0])
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Println(cfg.VaultPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
}
