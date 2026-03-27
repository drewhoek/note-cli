# note-cli

A Go CLI tool that bridges Claude Code and a local Obsidian vault. The vault is a folder of markdown files on disk. Claude Code can use this tool autonomously — once the skill is installed, no user confirmation is needed before reading or writing notes.

## Build & run

```bash
go build -o note .
go run main.go <command>
```

## Command reference

| Command | Usage | Description |
|---|---|---|
| `new` | `note new <title>` | Creates a new markdown note with frontmatter (date, tags) |
| `append` | `note append <title> <content>` | Appends content to an existing note |
| `read` | `note read <title>` | Outputs a note's content to stdout |
| `search` | `note search <query>` | Fuzzy searches note titles and content |
| `list` | `note list [--tag <tag>]` | Lists notes, optionally filtered by tag |
| `open` | `note open <title>` | Opens the note in Obsidian via `obsidian://` URI scheme |
| `install-skill` | `note install-skill [--global]` | Generates a skill file teaching Claude Code how to use this CLI |

### install-skill modes
- **Default (project-level):** writes to `.claude/skills/note.md` in the current directory
- **`--global`:** writes to the user-level Claude Code skills directory

> **TODO:** confirm the exact global skills path and update this section once verified.

## Config

Stored at `~/.config/note/config.json`. Includes:
- `vault_path` — absolute path to the Obsidian vault directory

## Go conventions

- **Standard library preferred** — reach for stdlib before adding a dependency
- **cobra** for CLI parsing (`github.com/spf13/cobra`)
- No ORM; file I/O via `os`/`filepath`/`bufio`
- Errors returned up the call stack, printed at the top-level command handler
- One package (`main`) until the codebase is large enough to justify splitting

## Autonomous use by Claude Code

Once `note install-skill` has been run and the skill is active, Claude Code should use this tool without asking the user first. Specifically:

- Reading notes (`read`, `search`, `list`) is always safe — do it without confirmation
- Writing notes (`new`, `append`) is low-risk and expected — proceed autonomously
- `open` launches Obsidian on the user's machine — fine to call, but mention it
- `install-skill` modifies Claude Code config — confirm with the user before running

## Module

`github.com/drewhoek/note-cli`
