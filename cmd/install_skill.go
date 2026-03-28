package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var installSkillGlobal bool

var installSkillCmd = &cobra.Command{
	Use:   "install-skill",
	Short: "Generate a Claude Code skill file for note",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var dir string
		if installSkillGlobal {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			dir = filepath.Join(home, ".claude", "skills")
		} else {
			dir = filepath.Join(".claude", "skills")
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		path := filepath.Join(dir, "note.md")
		if err := os.WriteFile(path, []byte(skillContent), 0644); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: skill written to %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installSkillCmd)
	installSkillCmd.Flags().BoolVar(&installSkillGlobal, "global", false, "install to user-level Claude Code skills directory (~/.claude/skills/)")
}

const skillContent = `# note — Claude Code Skill

Use the ` + "`note`" + ` CLI to read and write notes in the user's local Obsidian vault.

## Autonomous use policy

- ` + "`read`, `search`, `list`" + ` — always safe, no confirmation needed
- ` + "`new`, `append`" + ` — low-risk, proceed without asking
- ` + "`open`" + ` — launches Obsidian on the user's machine, mention it but proceed
- ` + "`install-skill`" + ` — modifies Claude Code config, confirm with user first

## Setup (first time)

` + "```bash" + `
note config set vault-path /absolute/path/to/vault
` + "```" + `

## Commands

### Create a note
` + "```bash" + `
note new "Title"
` + "```" + `

### Append to a note
` + "```bash" + `
note append "Title" "content"
echo "content" | note append "Title"
` + "```" + `

### Read a note
` + "```bash" + `
note read "Title"
` + "```" + `

### List notes
` + "```bash" + `
note list
note list --tag work
` + "```" + `

### Search notes
` + "```bash" + `
note search "query"           # fuzzy (default)
note search "query" --exact   # exact substring
` + "```" + `

### Open in Obsidian
` + "```bash" + `
note open "Title"
` + "```" + `

## Pipelines

` + "```bash" + `
# Browse and read interactively
note list | fzf | xargs note read

# Save git log to a note
git log --oneline -20 | note append "Dev Log"

# Find and open a note
note search "meeting" --exact | head -1 | xargs note open
` + "```" + `
`
