# Vault Features Design

**Date:** 2026-03-29
**Status:** Approved

## Problem

Four friction points prevent the vault from working reliably as persistent memory for Claude:

1. Claude must be manually told to check the vault at the start of every session
2. Notes accumulate without structure — no lifecycle, no way to demote stale content
3. Search returns slugs with no explanation of why a note matched
4. No way to surface related notes from a known starting point

## Solution Overview

Three feature tracks, in priority order:

1. **Session Bootstrap** — `note context` command + skill update
2. **Vault Health** — `note pin`, `note unpin`, `note archive`, `note set-status`
3. **Better Retrieval** — search snippets, `note related`, search scoring boost

---

## Track 1 — Session Bootstrap

### `note context`

A new command that outputs a structured markdown summary of the vault, designed to be consumed by Claude at session start.

**Output format:**
```
# Vault Context

## Pinned Notes
- <title> (tags: ...)
  <full note content>

## Recent Daily Notes
- 2026-03-29
- 2026-03-28

## All Tags
<tag> (<count>), ...

## Recent Notes
- <title> (date: <date>)
```

**Behavior:**
- Pinned notes are listed first with their full content included
- Non-pinned notes appear as title + date only
- Daily notes (tag: `daily`) are shown separately, most recent first, capped at 7
- Recent notes sorted by frontmatter `date`, capped at 10
- Output goes to stdout — pipeable, readable via `note context`

**Skill update:**
The global skill file (`~/.claude/skills/note.md`) is updated to instruct Claude to run `note context` automatically at the start of every session and use the output to orient itself before responding.

---

## Track 2 — Vault Health

### Frontmatter fields added

- `pinned: true` — marks a note as always-relevant
- `status: active | stale | reference` — lifecycle stage

### Commands

**`note pin <title>`**
Sets `pinned: true` in frontmatter. Note appears in `note context` with full content.

**`note unpin <title>`**
Removes `pinned` field from frontmatter.

**`note archive <title>`**
Moves the file to `archive/` subdirectory inside the vault. Archived notes are excluded from `note list`, `note search`, and `note context` by default. An `--include-archived` flag on `list` and `search` opts back in.

**`note set-status <title> <status>`**
Sets the `status` frontmatter field. Valid values: `active`, `stale`, `reference`. `note list` gains a `--status` filter flag. `note context` excludes `stale` notes by default.

### Principle
Nothing is deleted — only demoted. Vault history stays intact.

---

## Track 3 — Better Retrieval

### `note search --verbose`

Adds a `--verbose` flag to the existing `search` command. When set, prints a short excerpt around the matching text beneath each result slug. For fuzzy matches, shows the highest-scoring sentence. Makes results interpretable without reading every hit.

### `note related <title>`

New command. Finds notes connected to the given note by:
1. Direct wikilinks (highest rank) — notes that link to this one or that this one links to
2. Shared tags (lower rank) — notes sharing one or more tags

Returns a ranked list of slugs. Implemented within existing `search.go` logic using `Outlinks`, `Backlinks`, and tag data already available.

### Search scoring boost

Fuzzy search currently weights all notes equally. Two adjustments:
- Pinned notes get a score multiplier (e.g. ×1.5)
- Notes with a more recent frontmatter `date` get a small recency boost

This keeps important, current notes surfacing above old or stale ones.

---

## Architecture Notes

- All new frontmatter fields (`pinned`, `status`) handled in `internal/notes/frontmatter.go` — `noteMeta` struct gains two new fields, `joinNote` emits them conditionally
- `note archive` uses `os.MkdirAll` + `os.Rename` — same pattern as `Rename`
- `note context` lives in `cmd/context.go`, reads vault via existing `List`, `Read`, tag/pin data
- `note related` lives in `cmd/related.go`, composes `Outlinks` + `Backlinks` + tag intersection from `search.go`
- No new dependencies

---

## Out of Scope

- Full-text indexing or inverted index — stdlib file scan is sufficient at vault sizes under ~1000 notes
- Note deletion — archive is the intended alternative
- Multi-vault support
