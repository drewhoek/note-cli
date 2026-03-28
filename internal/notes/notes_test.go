package notes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Meeting Notes", "my-meeting-notes"},
		{"my-meeting", "my-meeting"},
		{"my-meeting.md", "my-meeting"},
		{"  spaces  ", "--spaces--"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Slug(tt.input)
			if got != tt.want {
				t.Errorf("Slug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	vault := t.TempDir()

	if err := Create(vault, "My Note"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	path := filepath.Join(vault, "my-note.md")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# My Note") {
		t.Errorf("expected h1 title in content, got:\n%s", content)
	}
	if !strings.Contains(content, "date:") {
		t.Errorf("expected date frontmatter, got:\n%s", content)
	}
	if !strings.Contains(content, "tags: []") {
		t.Errorf("expected tags frontmatter, got:\n%s", content)
	}
}

// mustCreate is a test helper that calls Create and fails the test if it returns an error.
func mustCreate(t *testing.T, vault, title string) {
	t.Helper()
	if err := Create(vault, title); err != nil {
		t.Fatalf("mustCreate(%q): %v", title, err)
	}
}

func TestCreateDuplicate(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "My Note")

	err := Create(vault, "My Note")
	if err == nil {
		t.Fatal("expected error when creating duplicate note, got nil")
	}
}

func TestRead(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "My Note")

	content, err := Read(vault, "My Note")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !strings.Contains(content, "# My Note") {
		t.Errorf("expected title in content, got:\n%s", content)
	}
}

func TestReadTitleVariants(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "My Note")

	variants := []string{"My Note", "my-note", "my-note.md"}
	for _, v := range variants {
		t.Run(v, func(t *testing.T) {
			_, err := Read(vault, v)
			if err != nil {
				t.Errorf("Read(%q): %v", v, err)
			}
		})
	}
}

func TestReadMissing(t *testing.T) {
	vault := t.TempDir()
	_, err := Read(vault, "nonexistent")
	if err == nil {
		t.Fatal("expected error reading missing note, got nil")
	}
}

func TestAppend(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "My Note")

	if err := Append(vault, "My Note", "appended content"); err != nil {
		t.Fatalf("Append: %v", err)
	}

	content, err := Read(vault, "My Note")
	if err != nil {
		t.Fatalf("Read after Append: %v", err)
	}
	if !strings.Contains(content, "appended content") {
		t.Errorf("expected appended content, got:\n%s", content)
	}
}

func TestAppendMissing(t *testing.T) {
	vault := t.TempDir()
	err := Append(vault, "nonexistent", "content")
	if err == nil {
		t.Fatal("expected error appending to missing note, got nil")
	}
}
