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

	titles, err := List(vault, "", "", false)
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

	titles, err := List(vault, "work", "", false)
	if err != nil {
		t.Fatalf("List with tag: %v", err)
	}
	if len(titles) != 1 || titles[0] != "work-note" {
		t.Errorf("List(work) = %v, want [work-note]", titles)
	}
}

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

func TestOutlinks(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "Source")
	if err := Append(vault, "Source", "See [[Target A]] and [[target-b]]."); err != nil {
		t.Fatalf("Append: %v", err)
	}
	links, err := Outlinks(vault, "Source")
	if err != nil {
		t.Fatalf("Outlinks: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("Outlinks = %v, want 2 results", links)
	}
}

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

func writeNoteWithTag(t *testing.T, vault, slug, tag string) {
	t.Helper()
	path := vault + "/" + slug + ".md"
	content := "---\ndate: 2026-01-01\ntags: [" + tag + "]\n---\n\n# " + slug + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeNoteWithTag: %v", err)
	}
}
