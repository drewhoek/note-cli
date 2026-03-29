# Vault Features Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add session bootstrap, vault health, and improved retrieval features to make note-cli work reliably as persistent memory for Claude.

**Architecture:** Three tracks built bottom-up — frontmatter extensions first (foundation), then vault health commands, then retrieval improvements, then the `note context` command that ties everything together. All new functionality follows existing patterns: stdlib only, errors returned up the stack, status messages to stderr, output to stdout.

**Tech Stack:** Go 1.23, cobra, stdlib only (os, filepath, strings, time, fmt)

---

## File Map

| File | Change |
|------|--------|
| `internal/notes/frontmatter.go` | Add `pinned bool` + `status string` to `noteMeta`; update `splitNote`, `joinNote` |
| `internal/notes/notes.go` | Add `Pin`, `Unpin`, `SetStatus`, `Archive` |
| `internal/notes/search.go` | Add `NoteResult` type, `SearchDetailed`, `extractSnippet`, `Related`; update `List` signature; add scoring boost |
| `internal/notes/notes_test.go` | Tests for Pin, Unpin, SetStatus, Archive |
| `internal/notes/search_test.go` | Tests for Related, SearchDetailed with snippets, List status filter |
| `cmd/pin.go` | `note pin <title>` / `note unpin <title>` |
| `cmd/archive.go` | `note archive <title>` |
| `cmd/set_status.go` | `note set-status <title> <status>` |
| `cmd/related.go` | `note related <title>` |
| `cmd/context.go` | `note context` |
| `cmd/list.go` | Add `--status` and `--include-archived` flags; update `List` call |
| `cmd/search.go` | Add `--verbose` and `--include-archived` flags; call `SearchDetailed` when verbose |
| `~/.claude/skills/note.md` | Add `note context` to command reference; update session-start instruction |

---

## Task 1: Extend frontmatter for pinned and status

**Files:**
- Modify: `internal/notes/frontmatter.go`
- Test: `internal/notes/notes_test.go`

- [ ] **Step 1: Write failing tests**

Add to `internal/notes/notes_test.go`:

```go
func TestPinRoundtrip(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "My Note")

	if err := Pin(vault, "My Note"); err != nil {
		t.Fatalf("Pin: %v", err)
	}
	content, err := Read(vault, "My Note")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.Contains(content, "pinned: true") {
		t.Errorf("expected pinned: true in frontmatter, got:\n%s", content)
	}

	if err := Unpin(vault, "My Note"); err != nil {
		t.Fatalf("Unpin: %v", err)
	}
	content, err = Read(vault, "My Note")
	if err != nil {
		t.Fatalf("Read after Unpin: %v", err)
	}
	if strings.Contains(content, "pinned:") {
		t.Errorf("expected pinned field removed after Unpin, got:\n%s", content)
	}
}

func TestSetStatus(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "My Note")

	if err := SetStatus(vault, "My Note", "stale"); err != nil {
		t.Fatalf("SetStatus: %v", err)
	}
	content, err := Read(vault, "My Note")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.Contains(content, "status: stale") {
		t.Errorf("expected status: stale in frontmatter, got:\n%s", content)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/notes/... -run "TestPinRoundtrip|TestSetStatus" -v
```

Expected: FAIL — `Pin`, `Unpin`, `SetStatus` undefined.

- [ ] **Step 3: Add pinned and status to noteMeta**

Replace the `noteMeta` struct and update `splitNote` and `joinNote` in `internal/notes/frontmatter.go`:

```go
type noteMeta struct {
	date   string
	tags   []string
	pinned bool
	status string
}
```

In `splitNote`, add two new cases inside the `for _, line := range strings.Split(header, "\n")` switch:

```go
case strings.HasPrefix(line, "pinned: "):
    meta.pinned = strings.TrimPrefix(line, "pinned: ") == "true"
case strings.HasPrefix(line, "status: "):
    meta.status = strings.TrimPrefix(line, "status: ")
```

Replace `joinNote` entirely:

```go
func joinNote(meta noteMeta, body string) string {
	tags := "[]"
	if len(meta.tags) > 0 {
		tags = "[" + strings.Join(meta.tags, ", ") + "]"
	}
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("date: " + meta.date + "\n")
	sb.WriteString("tags: " + tags + "\n")
	if meta.pinned {
		sb.WriteString("pinned: true\n")
	}
	if meta.status != "" {
		sb.WriteString("status: " + meta.status + "\n")
	}
	sb.WriteString("---\n")
	sb.WriteString(body)
	return sb.String()
}
```

- [ ] **Step 4: Add Pin, Unpin, SetStatus to notes.go**

Add to `internal/notes/notes.go`:

```go
// Pin marks a note as pinned in frontmatter.
func Pin(vaultPath, title string) error {
	content, err := Read(vaultPath, title)
	if err != nil {
		return err
	}
	meta, body := splitNote(content)
	meta.pinned = true
	return os.WriteFile(resolvePath(vaultPath, title), []byte(joinNote(meta, body)), 0644)
}

// Unpin removes the pinned flag from a note's frontmatter.
func Unpin(vaultPath, title string) error {
	content, err := Read(vaultPath, title)
	if err != nil {
		return err
	}
	meta, body := splitNote(content)
	meta.pinned = false
	return os.WriteFile(resolvePath(vaultPath, title), []byte(joinNote(meta, body)), 0644)
}

// SetStatus sets the status field in a note's frontmatter.
// Valid values: active, stale, reference. Pass "" to clear.
func SetStatus(vaultPath, title, status string) error {
	content, err := Read(vaultPath, title)
	if err != nil {
		return err
	}
	meta, body := splitNote(content)
	meta.status = status
	return os.WriteFile(resolvePath(vaultPath, title), []byte(joinNote(meta, body)), 0644)
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/notes/... -run "TestPinRoundtrip|TestSetStatus" -v
```

Expected: PASS

- [ ] **Step 6: Run full test suite**

```bash
go test ./... -v
```

Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/notes/frontmatter.go internal/notes/notes.go internal/notes/notes_test.go
git commit -m "feat: extend frontmatter with pinned and status fields"
```

---

## Task 2: Archive command

**Files:**
- Modify: `internal/notes/notes.go`, `internal/notes/notes_test.go`
- Create: `cmd/archive.go`

- [ ] **Step 1: Write failing test**

Add to `internal/notes/notes_test.go`:

```go
func TestArchive(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "Old Note")

	if err := Archive(vault, "Old Note"); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// Original file should be gone from vault root
	if _, err := os.Stat(filepath.Join(vault, "old-note.md")); err == nil {
		t.Error("expected old-note.md to be gone from vault root after archive")
	}

	// File should exist in archive/ subdir
	if _, err := os.Stat(filepath.Join(vault, "archive", "old-note.md")); err != nil {
		t.Errorf("expected archive/old-note.md to exist: %v", err)
	}
}

func TestArchiveMissing(t *testing.T) {
	vault := t.TempDir()
	err := Archive(vault, "nonexistent")
	if err == nil {
		t.Fatal("expected error archiving missing note, got nil")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/notes/... -run "TestArchive" -v
```

Expected: FAIL — `Archive` undefined.

- [ ] **Step 3: Add Archive to notes.go**

Add to `internal/notes/notes.go` (also add `"path/filepath"` to the import if not already present — it already is):

```go
// Archive moves a note to the archive/ subdirectory inside the vault.
// Archived notes are excluded from list, search, and context by default.
func Archive(vaultPath, title string) error {
	src := resolvePath(vaultPath, title)
	if _, err := os.Stat(src); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("note %q not found", title)
	}
	archiveDir := filepath.Join(vaultPath, "archive")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return err
	}
	dst := filepath.Join(archiveDir, Slug(title)+".md")
	return os.Rename(src, dst)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/notes/... -run "TestArchive" -v
```

Expected: PASS

- [ ] **Step 5: Create cmd/archive.go**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <title>",
	Short: "Move a note to the archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := notes.ResolveTitle(args[0])
		if err := notes.Archive(cfg.VaultPath, title); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: archived %q\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}
```

- [ ] **Step 6: Run full test suite**

```bash
go test ./... -v
```

Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/notes/notes.go internal/notes/notes_test.go cmd/archive.go
git commit -m "feat: add note archive command"
```

---

## Task 3: Pin/Unpin and SetStatus commands

**Files:**
- Create: `cmd/pin.go`, `cmd/set_status.go`

- [ ] **Step 1: Create cmd/pin.go**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var pinCmd = &cobra.Command{
	Use:   "pin <title>",
	Short: "Pin a note so it appears in context output with full content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := notes.ResolveTitle(args[0])
		if err := notes.Pin(cfg.VaultPath, title); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: pinned %q\n", title)
		return nil
	},
}

var unpinCmd = &cobra.Command{
	Use:   "unpin <title>",
	Short: "Unpin a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := notes.ResolveTitle(args[0])
		if err := notes.Unpin(cfg.VaultPath, title); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: unpinned %q\n", title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pinCmd)
	rootCmd.AddCommand(unpinCmd)
}
```

- [ ] **Step 2: Create cmd/set_status.go**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var setStatusCmd = &cobra.Command{
	Use:   "set-status <title> <status>",
	Short: "Set a note's lifecycle status (active, stale, reference)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := notes.ResolveTitle(args[0])
		status := args[1]
		if status != "active" && status != "stale" && status != "reference" && status != "" {
			return fmt.Errorf("invalid status %q: must be active, stale, or reference", status)
		}
		if err := notes.SetStatus(cfg.VaultPath, title, status); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "note: set status of %q to %q\n", title, status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setStatusCmd)
}
```

- [ ] **Step 3: Build to verify no compile errors**

```bash
go build ./...
```

Expected: no output (success)

- [ ] **Step 4: Smoke test the commands**

```bash
note pin "drew hoek"
note set-status "drew hoek" active
note read "drew hoek" | head -8
```

Expected: frontmatter shows `pinned: true` and `status: active`.

- [ ] **Step 5: Commit**

```bash
git add cmd/pin.go cmd/set_status.go
git commit -m "feat: add pin, unpin, and set-status commands"
```

---

## Task 4: Update List with status filter and include-archived flag

**Files:**
- Modify: `internal/notes/search.go`, `internal/notes/search_test.go`, `cmd/list.go`

- [ ] **Step 1: Write failing tests**

Add to `internal/notes/search_test.go`:

```go
func TestListByStatus(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "Active Note")
	mustCreate(t, vault, "Stale Note")

	if err := SetStatus(vault, "Stale Note", "stale"); err != nil {
		t.Fatalf("SetStatus: %v", err)
	}

	titles, err := List(vault, "", "stale", false)
	if err != nil {
		t.Fatalf("List with status: %v", err)
	}
	if len(titles) != 1 || titles[0] != "stale-note" {
		t.Errorf("List(status=stale) = %v, want [stale-note]", titles)
	}
}

func TestListIncludeArchived(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "Active Note")
	mustCreate(t, vault, "Old Note")

	if err := Archive(vault, "Old Note"); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// Without flag: archived note excluded
	titles, err := List(vault, "", "", false)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, title := range titles {
		if title == "old-note" {
			t.Error("archived note should be excluded from List by default")
		}
	}

	// With flag: archived note included
	all, err := List(vault, "", "", true)
	if err != nil {
		t.Fatalf("List includeArchived: %v", err)
	}
	found := false
	for _, title := range all {
		if title == "old-note" {
			found = true
		}
	}
	if !found {
		t.Error("expected archived note in List with includeArchived=true")
	}
}
```

- [ ] **Step 2: Also update the existing TestList and TestListByTag calls** (they will fail due to signature change)

In `internal/notes/search_test.go`, update existing calls:

```go
// TestList: change
titles, err := List(vault, "")
// to
titles, err := List(vault, "", "", false)

// TestListByTag: change
titles, err := List(vault, "work")
// to
titles, err := List(vault, "work", "", false)
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
go test ./internal/notes/... -run "TestList" -v
```

Expected: FAIL — `List` called with wrong number of arguments.

- [ ] **Step 4: Update List signature in search.go**

Replace the `List` function in `internal/notes/search.go`:

```go
// List returns all note slugs in the vault, optionally filtered by tag and/or status.
// Pass includeArchived=true to also include notes in the archive/ subdirectory.
func List(vaultPath, tag, status string, includeArchived bool) ([]string, error) {
	var dirs []string
	dirs = append(dirs, vaultPath)
	if includeArchived {
		dirs = append(dirs, filepath.Join(vaultPath, "archive"))
	}

	var titles []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			slug := strings.TrimSuffix(e.Name(), ".md")
			notePath := filepath.Join(dir, e.Name())

			if tag != "" || status != "" {
				data, err := os.ReadFile(notePath)
				if err != nil {
					continue
				}
				meta, _ := splitNote(string(data))
				if tag != "" {
					found := false
					for _, t := range meta.tags {
						if t == tag {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}
				if status != "" && meta.status != status {
					continue
				}
			}
			titles = append(titles, slug)
		}
	}
	return titles, nil
}
```

- [ ] **Step 5: Update cmd/list.go**

Replace `cmd/list.go` entirely:

```go
package cmd

import (
	"fmt"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var listTag string
var listStatus string
var listIncludeArchived bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes, optionally filtered by tag or status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		titles, err := notes.List(cfg.VaultPath, listTag, listStatus, listIncludeArchived)
		if err != nil {
			return err
		}
		for _, t := range titles {
			fmt.Println(t)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listTag, "tag", "", "filter by tag")
	listCmd.Flags().StringVar(&listStatus, "status", "", "filter by status (active, stale, reference)")
	listCmd.Flags().BoolVar(&listIncludeArchived, "include-archived", false, "include notes in the archive/ subdirectory")
}
```

- [ ] **Step 6: Run tests to verify they pass**

```bash
go test ./... -v
```

Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/notes/search.go internal/notes/search_test.go cmd/list.go
git commit -m "feat: add --status and --include-archived filters to note list"
```

---

## Task 5: Related command

**Files:**
- Modify: `internal/notes/search.go`, `internal/notes/search_test.go`
- Create: `cmd/related.go`

- [ ] **Step 1: Write failing test**

Add to `internal/notes/search_test.go`:

```go
func TestRelated(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "Source")
	mustCreate(t, vault, "Target")
	mustCreate(t, vault, "Tagged")
	mustCreate(t, vault, "Unrelated")

	// Source links to Target
	if err := Append(vault, "Source", "See [[Target]]."); err != nil {
		t.Fatalf("Append: %v", err)
	}
	// Source and Tagged share a tag
	if _, err := AddTag(vault, "Source", "work"); err != nil {
		t.Fatalf("AddTag Source: %v", err)
	}
	if _, err := AddTag(vault, "Tagged", "work"); err != nil {
		t.Fatalf("AddTag Tagged: %v", err)
	}

	related, err := Related(vault, "Source")
	if err != nil {
		t.Fatalf("Related: %v", err)
	}

	slugs := map[string]bool{}
	for _, r := range related {
		slugs[r] = true
	}

	if !slugs["target"] {
		t.Error("expected 'target' in related (wikilink)")
	}
	if !slugs["tagged"] {
		t.Error("expected 'tagged' in related (shared tag)")
	}
	if slugs["unrelated"] {
		t.Error("expected 'unrelated' NOT in related")
	}
	// Wikilink results should come before tag results
	if len(related) >= 2 && related[0] != "target" {
		t.Errorf("expected wikilink results first, got %v", related)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/notes/... -run "TestRelated" -v
```

Expected: FAIL — `Related` undefined.

- [ ] **Step 3: Add Related to search.go**

Add at the end of `internal/notes/search.go`. Also add `"path/filepath"` to imports if not already there (it already is):

```go
// Related returns slugs of notes connected to the given note by wikilinks (highest rank)
// or shared tags (lower rank). The note itself is excluded.
func Related(vaultPath, title string) ([]string, error) {
	slug := Slug(title)
	seen := map[string]bool{slug: true}
	var results []string

	// Wikilink connections (highest rank): notes this note links to + notes that link back
	outlinks, _ := Outlinks(vaultPath, title)
	backlinks, _ := Backlinks(vaultPath, title)
	for _, l := range append(outlinks, backlinks...) {
		lSlug := Slug(l)
		if !seen[lSlug] {
			seen[lSlug] = true
			results = append(results, lSlug)
		}
	}

	// Shared tag connections (lower rank)
	myTags, err := ReadTags(vaultPath, title)
	if err != nil {
		return results, nil
	}
	tagSet := make(map[string]bool, len(myTags))
	for _, t := range myTags {
		tagSet[t] = true
	}

	entries, err := os.ReadDir(vaultPath)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		noteSlug := strings.TrimSuffix(e.Name(), ".md")
		if seen[noteSlug] {
			continue
		}
		tags, _ := parseTags(filepath.Join(vaultPath, e.Name()))
		for _, t := range tags {
			if tagSet[t] {
				results = append(results, noteSlug)
				seen[noteSlug] = true
				break
			}
		}
	}
	return results, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/notes/... -run "TestRelated" -v
```

Expected: PASS

- [ ] **Step 5: Create cmd/related.go**

```go
package cmd

import (
	"fmt"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var relatedCmd = &cobra.Command{
	Use:   "related <title>",
	Short: "Find notes connected by wikilinks or shared tags",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		title := notes.ResolveTitle(args[0])
		results, err := notes.Related(cfg.VaultPath, title)
		if err != nil {
			return err
		}
		for _, r := range results {
			fmt.Println(r)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(relatedCmd)
}
```

- [ ] **Step 6: Run full test suite**

```bash
go test ./... -v
```

Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/notes/search.go internal/notes/search_test.go cmd/related.go
git commit -m "feat: add note related command"
```

---

## Task 6: Verbose search with snippets and scoring boost

**Files:**
- Modify: `internal/notes/search.go`, `internal/notes/search_test.go`, `cmd/search.go`

- [ ] **Step 1: Write failing tests**

Add to `internal/notes/search_test.go`:

```go
func TestSearchDetailed_Snippet(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "Meeting Notes")
	if err := Append(vault, "Meeting Notes", "We discussed the quarterly budget review."); err != nil {
		t.Fatalf("Append: %v", err)
	}

	results, err := SearchDetailed(vault, "quarterly", true, false)
	if err != nil {
		t.Fatalf("SearchDetailed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Snippet == "" {
		t.Error("expected non-empty snippet in verbose search")
	}
	if !strings.Contains(results[0].Snippet, "quarterly") {
		t.Errorf("expected snippet to contain query word, got: %q", results[0].Snippet)
	}
}

func TestSearchDetailed_PinnedBoost(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "meeting alpha")
	mustCreate(t, vault, "meeting beta")

	if err := Pin(vault, "meeting beta"); err != nil {
		t.Fatalf("Pin: %v", err)
	}

	results, err := SearchDetailed(vault, "meeting", false, false)
	if err != nil {
		t.Fatalf("SearchDetailed: %v", err)
	}
	if len(results) < 2 {
		t.Fatal("expected at least 2 results")
	}
	if results[0].Slug != "meeting-beta" {
		t.Errorf("expected pinned note to rank first, got %v", results[0].Slug)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/notes/... -run "TestSearchDetailed" -v
```

Expected: FAIL — `SearchDetailed` and `NoteResult` undefined.

- [ ] **Step 3: Add NoteResult, extractSnippet, and SearchDetailed to search.go**

Add after the existing `searchResult` type in `internal/notes/search.go`. Also add `"time"` to imports:

```go
// NoteResult is a search result with an optional context snippet.
type NoteResult struct {
	Slug    string
	Snippet string
}

// extractSnippet returns a short excerpt from content around the first occurrence of query.
// Returns the first 100 characters of content if query is not found.
func extractSnippet(content, query string) string {
	lower := strings.ToLower(content)
	queryLower := strings.ToLower(query)
	idx := strings.Index(lower, queryLower)

	var raw string
	if idx == -1 {
		if len(content) > 100 {
			raw = content[:100]
		} else {
			raw = content
		}
		return strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
	}

	start := idx - 50
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + 50
	if end > len(content) {
		end = len(content)
	}

	snippet := strings.ReplaceAll(content[start:end], "\n", " ")
	snippet = strings.TrimSpace(snippet)
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet += "..."
	}
	return snippet
}

// SearchDetailed searches note titles and content, returning NoteResult with optional snippets.
// With exact=true, uses substring match. With exact=false, uses bigram similarity with
// a 1.5x boost for pinned notes and a 1.1x boost for notes dated within the last 30 days.
func SearchDetailed(vaultPath, query string, exact, includeArchived bool) ([]NoteResult, error) {
	var dirs []string
	dirs = append(dirs, vaultPath)
	if includeArchived {
		dirs = append(dirs, filepath.Join(vaultPath, "archive"))
	}

	type scoredResult struct {
		slug    string
		score   float64
		content string
	}
	var scored []scoredResult
	queryLower := strings.ToLower(query)

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			slug := strings.TrimSuffix(e.Name(), ".md")
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			content := string(data)
			contentLower := strings.ToLower(content)
			slugLower := strings.ToLower(slug)

			var score float64
			if exact {
				if strings.Contains(slugLower, queryLower) || strings.Contains(contentLower, queryLower) {
					score = 1.0
				}
			} else {
				meta, body := splitNote(content)
				titleScore := bigramSimilarity(query, slug)
				contentScore := bigramSimilarity(query, body)
				score = titleScore
				if contentScore > score {
					score = contentScore
				}
				if score > 0.1 {
					if meta.pinned {
						score *= 1.5
					}
					if t, err := time.Parse("2006-01-02", meta.date); err == nil {
						if time.Since(t).Hours()/24 < 30 {
							score *= 1.1
						}
					}
				}
			}

			if score > 0 {
				scored = append(scored, scoredResult{slug: slug, score: score, content: content})
			}
		}
	}

	if !exact {
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].score > scored[j].score
		})
	}

	results := make([]NoteResult, len(scored))
	for i, r := range scored {
		results[i] = NoteResult{
			Slug:    r.slug,
			Snippet: extractSnippet(r.content, query),
		}
	}
	return results, nil
}
```

- [ ] **Step 4: Update the existing Search function to delegate to SearchDetailed**

Replace the existing `Search` function in `internal/notes/search.go`:

```go
// Search searches note titles and content, returning slugs only.
// Use SearchDetailed for snippet support and scoring boosts.
func Search(vaultPath, query string, exact bool) ([]string, error) {
	results, err := SearchDetailed(vaultPath, query, exact, false)
	if err != nil {
		return nil, err
	}
	slugs := make([]string, len(results))
	for i, r := range results {
		slugs[i] = r.Slug
	}
	return slugs, nil
}
```

- [ ] **Step 5: Add "time" to search.go imports**

The imports block in `internal/notes/search.go` should be:

```go
import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)
```

- [ ] **Step 6: Run tests to verify they pass**

```bash
go test ./internal/notes/... -run "TestSearchDetailed|TestSearch" -v
```

Expected: PASS

- [ ] **Step 7: Update cmd/search.go to support --verbose**

Replace `cmd/search.go` entirely:

```go
package cmd

import (
	"fmt"

	"github.com/drewhoek/note-cli/internal/config"
	"github.com/drewhoek/note-cli/internal/notes"
	"github.com/spf13/cobra"
)

var searchExact bool
var searchVerbose bool
var searchIncludeArchived bool

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search note titles and content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		results, err := notes.SearchDetailed(cfg.VaultPath, args[0], searchExact, searchIncludeArchived)
		if err != nil {
			return err
		}
		for _, r := range results {
			if searchVerbose {
				fmt.Printf("%s\n  %s\n\n", r.Slug, r.Snippet)
			} else {
				fmt.Println(r.Slug)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().BoolVar(&searchExact, "exact", false, "use exact substring matching instead of fuzzy")
	searchCmd.Flags().BoolVar(&searchVerbose, "verbose", false, "show a content snippet for each result")
	searchCmd.Flags().BoolVar(&searchIncludeArchived, "include-archived", false, "include archived notes in results")
}
```

- [ ] **Step 8: Run full test suite**

```bash
go test ./... -v
```

Expected: all PASS

- [ ] **Step 9: Smoke test verbose search**

```bash
note search "cli" --verbose
```

Expected: results with indented snippet beneath each slug.

- [ ] **Step 10: Commit**

```bash
git add internal/notes/search.go internal/notes/search_test.go cmd/search.go
git commit -m "feat: add verbose search with snippets and pinned/recency scoring boost"
```

---

## Task 7: note context command

**Files:**
- Create: `cmd/context.go`

- [ ] **Step 1: Create cmd/context.go**

```go
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
		// Use unexported splitNote via the notes package indirectly — read tags and pin state via exported functions
		tags, _ := notes.ReadTags(vaultPath, slug)
		for _, t := range tags {
			tagCounts[t]++
		}

		// Determine pin/status by reading the file directly here
		// We'll embed a small parser inline since splitNote is unexported
		info := parseNoteInfo(slug, content, tags)

		if info.pinned {
			pinned = append(pinned, info)
		} else if hasTag(tags, "daily") {
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

func parseNoteInfo(slug, content string, tags []string) noteInfo {
	info := noteInfo{slug: slug, content: content, tags: tags}
	if !strings.HasPrefix(content, "---\n") {
		return info
	}
	rest := content[4:]
	end := strings.Index(rest, "\n---\n")
	if end == -1 {
		return info
	}
	header := rest[:end]
	for _, line := range strings.Split(header, "\n") {
		switch {
		case strings.HasPrefix(line, "date: "):
			info.date = strings.TrimPrefix(line, "date: ")
		case strings.HasPrefix(line, "pinned: "):
			info.pinned = strings.TrimPrefix(line, "pinned: ") == "true"
		case strings.HasPrefix(line, "status: "):
			info.status = strings.TrimPrefix(line, "status: ")
		}
	}
	return info
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
```

- [ ] **Step 2: Build to verify no compile errors**

```bash
go build ./...
```

Expected: no output (success). If there's a `var _ = time.Now` unused warning, remove that line — `time` is imported but not used directly. Remove `"time"` from the import block if the build fails on it.

- [ ] **Step 3: Smoke test**

```bash
note context
```

Expected: structured markdown output with pinned notes (full content), recent dailies, active notes, tag summary.

- [ ] **Step 4: Run full test suite**

```bash
go test ./... -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/context.go
git commit -m "feat: add note context command for session bootstrap"
```

---

## Task 8: Update skill file

**Files:**
- Modify: `~/.claude/skills/note.md`
- Modify: `cmd/install_skill.go`

- [ ] **Step 1: Read current skill content**

```bash
cat ~/.claude/skills/note.md
```

- [ ] **Step 2: Add new commands and session-start instruction to the skill**

The skill file needs two additions:

1. A **session-start rule** near the top of the autonomous use section:
```
## Session Start
At the start of every session, run `note context` and use the output to orient yourself before responding. Pinned notes contain critical context about the user and active projects.
```

2. New commands in the command reference table:
```
| `context`    | `note context`                       | Output structured vault summary for session bootstrap       |
| `pin`        | `note pin <title>`                   | Pin a note so it appears in context with full content       |
| `unpin`      | `note unpin <title>`                 | Unpin a note                                                |
| `archive`    | `note archive <title>`               | Move a note to archive/ (excluded from list/search/context) |
| `set-status` | `note set-status <title> <status>`   | Set lifecycle status: active, stale, reference              |
| `related`    | `note related <title>`               | Find notes connected by wikilinks or shared tags            |
| `search`     | `note search <query> [--verbose]`    | Fuzzy search; --verbose shows content snippets              |
```

- [ ] **Step 3: Update skillContent in cmd/install_skill.go**

Read `cmd/install_skill.go`, then update the `skillContent` string to match the updated skill file content from Step 2.

- [ ] **Step 4: Reinstall the skill**

```bash
note install-skill --global
```

- [ ] **Step 5: Build and run full test suite**

```bash
go build ./... && go test ./... -v
```

Expected: all PASS, binary builds cleanly.

- [ ] **Step 6: Commit**

```bash
git add cmd/install_skill.go
git commit -m "feat: update skill file with new commands and session-start instruction"
```

---

## Task 9: Final integration smoke test

- [ ] **Step 1: Pin a note and run context**

```bash
note pin "note-cli project"
note context
```

Expected: `note-cli project` appears under `## Pinned Notes` with full content.

- [ ] **Step 2: Test set-status filtering**

```bash
note set-status "Welcome" stale
note list
```

Expected: `Welcome` excluded from list (status is stale, context excludes stale by default).

```bash
note list --status stale
```

Expected: `welcome` appears.

- [ ] **Step 3: Test related**

```bash
note related "note-cli project"
```

Expected: `drew-hoek` appears (wikilink connection from earlier).

- [ ] **Step 4: Test verbose search**

```bash
note search "obsidian" --verbose
```

Expected: results with indented snippets.

- [ ] **Step 5: Test archive**

```bash
note archive "Welcome"
note list
```

Expected: `welcome` gone from list.

```bash
note list --include-archived
```

Expected: `welcome` appears again.

- [ ] **Step 6: Final full test run**

```bash
go test ./... -race -v
```

Expected: all PASS, no race conditions.

- [ ] **Step 7: Open a PR**

```bash
git push origin feature/vault-features
gh pr create --title "feat: session bootstrap, vault health, and retrieval improvements" --body "$(cat <<'EOF'
## Summary
- `note context`: structured vault summary for Claude session bootstrap
- `note pin`/`note unpin`: mark notes as always-relevant
- `note archive`: move notes out of active vault without deleting
- `note set-status`: lifecycle tracking (active/stale/reference)
- `note related`: find connected notes via wikilinks and shared tags
- `note search --verbose`: show content snippets with results
- Search scoring boost for pinned and recently-dated notes
- Updated skill file with session-start instruction
EOF
)"
```
