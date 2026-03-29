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

- ` + "`read`, `search`, `list`, `links`, `backlinks`" + ` — always safe, no confirmation needed
- ` + "`new`, `append`, `tag`, `today`, `rename`" + ` — low-risk, proceed without asking
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
note new "Title" --template meeting   # meeting, person, or project
` + "```" + `

### Append to a note
` + "```bash" + `
note append "Title" "content"
echo "content" | note append "Title"
note append "[[Title]]" "content"     # wikilink syntax works everywhere
` + "```" + `

### Read a note
` + "```bash" + `
note read "Title"
note read "[[Title]]"
` + "```" + `

### List notes
` + "```bash" + `
note list
note list --tag work
` + "```" + `

### Search notes
` + "```bash" + `
note search "query"           # fuzzy (default) — searches titles and body
note search "query" --exact   # exact substring
` + "```" + `

### Tags
` + "```bash" + `
note tag "Title"              # list tags
note tag add "Title" work     # add a tag
note tag remove "Title" work  # remove a tag
` + "```" + `

### Wikilinks
` + "```bash" + `
note links "Title"            # list all [[wikilinks]] in a note
note backlinks "Title"        # list notes that link to this one
note rename "Old" "New"       # rename + update all wikilinks across vault
` + "```" + `

### Open in Obsidian
` + "```bash" + `
note open "Title"
` + "```" + `

### Daily note
` + "```bash" + `
note today                    # create or confirm today's note (YYYY-MM-DD)
` + "```" + `

## Templates

` + "`--template meeting`" + ` — Attendees, Agenda, Notes, Action Items
` + "`--template person`" + `  — Contact, Notes, Links
` + "`--template project`" + ` — Overview, Goals, Tasks, Notes

## Pipelines

` + "```bash" + `
# Browse and read interactively
note list | fzf | xargs note read

# Save git log to a note
git log --oneline -20 | note append "Dev Log"

# Find notes linking to a topic and read them
note backlinks "some-topic" | xargs -I{} note read {}

# Chain: search → open in Obsidian
note search "meeting" --exact | head -1 | xargs note open
` + "```" + `
`
