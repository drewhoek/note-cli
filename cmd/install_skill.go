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

- ` + "`read`, `search`, `list`, `links`, `backlinks`, `context`" + ` — always safe, no confirmation needed
- ` + "`new`, `append`, `tag`, `today`, `rename`, `pin`, `unpin`, `set-status`, `archive`, `related`" + ` — low-risk, proceed without asking
- ` + "`open`" + ` — launches Obsidian on the user's machine, mention it but proceed
- ` + "`install-skill`" + ` — modifies Claude Code config, confirm with user first

## Session Start

At the start of every session, run ` + "`note context`" + ` and use the output to orient yourself before responding. Pinned notes contain critical context about the user and active projects.

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

### Vault context (session bootstrap)
` + "```bash" + `
note context                  # structured vault summary — run at session start
` + "```" + `

### List notes
` + "```bash" + `
note list
note list --tag work
note list --status active              # filter by status: active, stale, reference
note list --include-archived           # include archived notes
` + "```" + `

### Search notes
` + "```bash" + `
note search "query"                    # fuzzy (default) — searches titles and body
note search "query" --exact            # exact substring
note search "query" --verbose          # show content snippet for each result
note search "query" --include-archived # include archived notes
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

### Pin / archive notes
` + "```bash" + `
note pin "Title"              # pin a note — appears in context with full content
note unpin "Title"            # unpin a note
note archive "Title"          # move to archive/ (excluded from list/search/context)
note set-status "Title" active      # set lifecycle status: active, stale, reference
note set-status "Title" stale
note set-status "Title" reference
` + "```" + `

### Related notes
` + "```bash" + `
note related "Title"          # find notes connected by wikilinks or shared tags
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
