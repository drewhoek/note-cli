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

// Create creates a new note with YAML frontmatter. Returns ErrNoteExists if the note already exists.
func Create(vaultPath, title string) error {
	path := resolvePath(vaultPath, title)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %q", ErrNoteExists, title)
	}
	content := fmt.Sprintf("---\ndate: %s\ntags: []\n---\n\n# %s\n",
		time.Now().Format("2006-01-02"), title)
	return os.WriteFile(path, []byte(content), 0644)
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
