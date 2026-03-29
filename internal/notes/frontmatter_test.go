package notes

import (
	"testing"
)

func TestSplitJoteRoundtrip(t *testing.T) {
	content := "---\ndate: 2026-01-01\ntags: [work, meeting]\n---\n\n# My Note\n\nBody here.\n"
	meta, body := splitNote(content)

	if meta.date != "2026-01-01" {
		t.Errorf("date = %q, want %q", meta.date, "2026-01-01")
	}
	if len(meta.tags) != 2 || meta.tags[0] != "work" || meta.tags[1] != "meeting" {
		t.Errorf("tags = %v, want [work meeting]", meta.tags)
	}

	reassembled := joinNote(meta, body)
	if reassembled != content {
		t.Errorf("roundtrip failed:\ngot:  %q\nwant: %q", reassembled, content)
	}
}

func TestSplitNoteEmptyTags(t *testing.T) {
	content := "---\ndate: 2026-01-01\ntags: []\n---\n\n# Note\n"
	meta, _ := splitNote(content)
	if len(meta.tags) != 0 {
		t.Errorf("expected empty tags, got %v", meta.tags)
	}
}

func TestSplitNoteNoFrontmatter(t *testing.T) {
	content := "just a body\n"
	meta, body := splitNote(content)
	if meta.date != "" || len(meta.tags) != 0 {
		t.Errorf("expected empty meta for plain file, got %+v", meta)
	}
	if body != content {
		t.Errorf("body = %q, want %q", body, content)
	}
}

func TestAddRemoveTag(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "My Note")

	added, err := AddTag(vault, "My Note", "work")
	if err != nil {
		t.Fatalf("AddTag: %v", err)
	}
	if !added {
		t.Error("expected added=true on first add")
	}

	// adding again should be a no-op
	added, err = AddTag(vault, "My Note", "work")
	if err != nil {
		t.Fatalf("AddTag duplicate: %v", err)
	}
	if added {
		t.Error("expected added=false on duplicate add")
	}

	tags, err := ReadTags(vault, "My Note")
	if err != nil {
		t.Fatalf("ReadTags: %v", err)
	}
	if len(tags) != 1 || tags[0] != "work" {
		t.Errorf("tags = %v, want [work]", tags)
	}

	removed, err := RemoveTag(vault, "My Note", "work")
	if err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}
	if !removed {
		t.Error("expected removed=true")
	}

	tags, err = ReadTags(vault, "My Note")
	if err != nil {
		t.Fatalf("ReadTags after remove: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected empty tags after remove, got %v", tags)
	}
}

func TestRemoveTagNotPresent(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "My Note")

	removed, err := RemoveTag(vault, "My Note", "nonexistent")
	if err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}
	if removed {
		t.Error("expected removed=false for tag not present")
	}
}

func TestBacklinks(t *testing.T) {
	vault := t.TempDir()
	mustCreate(t, vault, "Target Note")
	mustCreate(t, vault, "Note A")
	mustCreate(t, vault, "Note B")

	// Note A links to Target Note by title, Note B by slug
	if err := Append(vault, "Note A", "See [[Target Note]] for more."); err != nil {
		t.Fatalf("Append A: %v", err)
	}
	if err := Append(vault, "Note B", "See [[target-note]] for more."); err != nil {
		t.Fatalf("Append B: %v", err)
	}

	results, err := Backlinks(vault, "Target Note")
	if err != nil {
		t.Fatalf("Backlinks: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Backlinks = %v, want [note-a note-b]", results)
	}
}
