package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Output a structured vault summary for use as Claude session context",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		return runContext(cfg.VaultPath)
	},
}

type noteInfo struct {
	slug    string
	date    string
	tags    []string
	pinned  bool
	status  string
	content string
}

func runContext(vaultPath string) error {
	entries, err := os.ReadDir(vaultPath)
	if err != nil {
		return err
	}

	var pinned []noteInfo
	var dailies []noteInfo
	var active []noteInfo
	tagCounts := map[string]int{}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		data, err := os.ReadFile(filepath.Join(vaultPath, e.Name()))
		if err != nil {
			continue
		}
		content := string(data)

		meta, err := notes.ReadMeta(vaultPath, slug)
		if err != nil {
			continue
		}

		for _, t := range meta.Tags {
			tagCounts[t]++
		}

		info := noteInfo{
			slug:    slug,
			date:    meta.Date,
			tags:    meta.Tags,
			pinned:  meta.Pinned,
			status:  meta.Status,
			content: content,
		}

		if info.pinned {
			pinned = append(pinned, info)
		} else if hasTag(meta.Tags, "daily") {
			dailies = append(dailies, info)
		} else if info.status != "stale" {
			active = append(active, info)
		}
	}

	// Sort dailies and active notes by date descending
	sort.Slice(dailies, func(i, j int) bool { return dailies[i].date > dailies[j].date })
	sort.Slice(active, func(i, j int) bool { return active[i].date > active[j].date })

	fmt.Println("# Vault Context")

	if len(pinned) > 0 {
		fmt.Println("\n## Pinned Notes")
		for _, n := range pinned {
			fmt.Printf("\n### %s\n", n.slug)
			fmt.Println(n.content)
		}
	}

	if len(dailies) > 0 {
		fmt.Println("\n## Recent Daily Notes")
		cap := 7
		if len(dailies) < cap {
			cap = len(dailies)
		}
		for _, n := range dailies[:cap] {
			fmt.Printf("- %s\n", n.slug)
		}
	}

	if len(active) > 0 {
		fmt.Println("\n## Active Notes")
		cap := 10
		if len(active) < cap {
			cap = len(active)
		}
		for _, n := range active[:cap] {
			tagStr := ""
			if len(n.tags) > 0 {
				tagStr = " (tags: " + strings.Join(n.tags, ", ") + ")"
			}
			fmt.Printf("- %s%s\n", n.slug, tagStr)
		}
	}

	if len(tagCounts) > 0 {
		fmt.Println("\n## All Tags")
		type tagCount struct {
			tag   string
			count int
		}
		var tc []tagCount
		for t, c := range tagCounts {
			tc = append(tc, tagCount{t, c})
		}
		sort.Slice(tc, func(i, j int) bool { return tc[i].tag < tc[j].tag })
		parts := make([]string, len(tc))
		for i, t := range tc {
			parts[i] = fmt.Sprintf("%s (%d)", t.tag, t.count)
		}
		fmt.Println(strings.Join(parts, ", "))
	}

	return nil
}

func hasTag(tags []string, target string) bool {
	for _, t := range tags {
		if t == target {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(contextCmd)
}
