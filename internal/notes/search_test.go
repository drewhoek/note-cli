package notes

import (
	"os"
	"testing"
)

func TestBigramSimilarity(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"meeting", "meeting", true},
		{"meting", "meeting", true},
		{"hello", "world", false},
	}
	for _, tt := range tests {
		score := bigramSimilarity(tt.a, tt.b)
		got := score > 0.1
		if got != tt.want {
			t.Errorf("bigramSimilarity(%q, %q) = %.2f, wantPositive=%v", tt.a, tt.b, score, tt.want)
		}
	}
}

func TestList(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "Alpha")
	mustCreate(t, vault, "Beta")
	mustCreate(t, vault, "Gamma")

	titles, err := List(vault, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(titles) != 3 {
		t.Errorf("List() returned %d titles, want 3", len(titles))
	}
}

func TestListByTag(t *testing.T) {
	vault := t.TempDir()
	writeNoteWithTag(t, vault, "work-note", "work")
	writeNoteWithTag(t, vault, "personal-note", "personal")

	titles, err := List(vault, "work")
	if err != nil {
		t.Fatalf("List with tag: %v", err)
	}
	if len(titles) != 1 || titles[0] != "work-note" {
		t.Errorf("List(work) = %v, want [work-note]", titles)
	}
}

func TestSearchExact(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "Meeting Notes")
	mustCreate(t, vault, "Shopping List")

	results, err := Search(vault, "meeting", true)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 || results[0] != "meeting-notes" {
		t.Errorf("Search(exact meeting) = %v, want [meeting-notes]", results)
	}
}

func TestSearchFuzzy(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "Meeting Notes")
	mustCreate(t, vault, "Shopping List")

	results, err := Search(vault, "meting", false)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Error("fuzzy search for 'meting' returned no results, expected to find 'meeting-notes'")
	}
}

func writeNoteWithTag(t *testing.T, vault, slug, tag string) {
	t.Helper()
	path := vault + "/" + slug + ".md"
	content := "---\ndate: 2026-01-01\ntags: [" + tag + "]\n---\n\n# " + slug + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeNoteWithTag: %v", err)
	}
}
