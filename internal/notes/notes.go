package notes

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Slug converts a note title or filename to a lowercase hyphenated slug.
// "My Meeting Notes" → "my-meeting-notes"
// "my-note.md"       → "my-note"
func Slug(title string) string {
	s := strings.ToLower(title)
	s = strings.TrimSuffix(s, ".md")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

// resolvePath returns the absolute file path for a note given a title, slug, or filename.
func resolvePath(vaultPath, title string) string {
	return filepath.Join(vaultPath, Slug(title)+".md")
}

// ErrNoteExists is returned by Create and CreateDaily when the note file already exists.
var ErrNoteExists = errors.New("note already exists")

// built-in templates for note new --template
var noteTemplates = map[string]string{
	"meeting": "\n## Attendees\n\n## Agenda\n\n## Notes\n\n## Action Items\n",
	"person":  "\n## Contact\n\n## Notes\n\n## Links\n",
	"project": "\n## Overview\n\n## Goals\n\n## Tasks\n\n## Notes\n",
}

// Create creates a new note with YAML frontmatter. Returns ErrNoteExists if the note already exists.
func Create(vaultPath, title string) error {
	return create(vaultPath, title, "")
}

// CreateWithTemplate creates a new note with a named template body.
// Valid templates: meeting, person, project. Unknown names fall back to blank.
func CreateWithTemplate(vaultPath, title, template string) error {
	return create(vaultPath, title, noteTemplates[template])
}

func create(vaultPath, title, body string) error {
	path := resolvePath(vaultPath, title)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %q", ErrNoteExists, title)
	}
	content := fmt.Sprintf("---\ndate: %s\ntags: []\n---\n\n# %s\n%s",
		time.Now().Format("2006-01-02"), title, body)
	return os.WriteFile(path, []byte(content), 0644)
}

// ResolveTitle strips [[wikilink]] brackets if present.
// "[[My Note]]" → "My Note", "my-note" → "my-note"
func ResolveTitle(s string) string {
	s = strings.TrimPrefix(s, "[[")
	s = strings.TrimSuffix(s, "]]")
	return s
}

// Rename renames a note file and updates all [[wikilinks]] pointing to it across the vault.
func Rename(vaultPath, oldTitle, newTitle string) error {
	oldPath := resolvePath(vaultPath, oldTitle)
	newPath := resolvePath(vaultPath, newTitle)

	if _, err := os.Stat(oldPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("note %q not found", oldTitle)
	}
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("%w: %q", ErrNoteExists, newTitle)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	oldSlug := Slug(oldTitle)
	newSlug := Slug(newTitle)
	entries, err := os.ReadDir(vaultPath)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(vaultPath, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		updated := string(data)
		updated = strings.ReplaceAll(updated, "[["+oldTitle+"]]", "[["+newTitle+"]]")
		updated = strings.ReplaceAll(updated, "[["+oldSlug+"]]", "[["+newSlug+"]]")
		if updated != string(data) {
			if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateDaily creates a daily note with a richer template. Returns ErrNoteExists if it already exists.
func CreateDaily(vaultPath, title string) error {
	path := resolvePath(vaultPath, title)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %q", ErrNoteExists, title)
	}
	content := fmt.Sprintf("---\ndate: %s\ntags: [daily]\n---\n\n# %s\n\n## Notes\n\n## Tasks\n",
		time.Now().Format("2006-01-02"), title)
	return os.WriteFile(path, []byte(content), 0644)
}

// Read returns the full content of a note. Accepts a title, slug, or filename.
func Read(vaultPath, title string) (string, error) {
	path := resolvePath(vaultPath, title)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("note %q not found", title)
		}
		return "", err
	}
	return string(data), nil
}

// Append adds content to an existing note on a new line. Returns an error if the note does not exist.
func Append(vaultPath, title, content string) error {
	path := resolvePath(vaultPath, title)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("note %q not found", title)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "\n%s", content)
	return err
}
