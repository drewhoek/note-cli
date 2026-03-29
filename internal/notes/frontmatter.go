package notes

import (
	"os"
	"strings"
)

// noteMeta holds the parsed frontmatter fields we care about.
type noteMeta struct {
	date   string
	tags   []string
	pinned bool
	status string
}

// splitNote parses a note file into its frontmatter metadata and body text.
// If the file doesn't start with ---, the whole content is treated as body.
func splitNote(content string) (noteMeta, string) {
	var meta noteMeta
	if !strings.HasPrefix(content, "---\n") {
		return meta, content
	}
	rest := content[4:] // skip opening ---\n
	end := strings.Index(rest, "\n---\n")
	if end == -1 {
		return meta, content
	}
	header := rest[:end]
	body := rest[end+5:] // skip \n---\n

	for _, line := range strings.Split(header, "\n") {
		switch {
		case strings.HasPrefix(line, "date: "):
			meta.date = strings.TrimPrefix(line, "date: ")
		case strings.HasPrefix(line, "tags: "):
			meta.tags = parseTagList(strings.TrimPrefix(line, "tags: "))
		case strings.HasPrefix(line, "pinned: "):
			meta.pinned = strings.TrimPrefix(line, "pinned: ") == "true"
		case strings.HasPrefix(line, "status: "):
			meta.status = strings.TrimPrefix(line, "status: ")
		}
	}
	return meta, body
}

// joinNote reassembles a note from its frontmatter metadata and body.
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

// parseTagList parses a YAML flow sequence like "[]" or "[work, meeting]".
func parseTagList(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return []string{}
	}
	var tags []string
	for _, t := range strings.Split(s, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

// ReadTags returns the tags from a note's frontmatter.
func ReadTags(vaultPath, title string) ([]string, error) {
	content, err := Read(vaultPath, title)
	if err != nil {
		return nil, err
	}
	meta, _ := splitNote(content)
	return meta.tags, nil
}

// AddTag adds a tag to a note if not already present.
// Returns false if the tag was already there (no-op).
func AddTag(vaultPath, title, tag string) (bool, error) {
	content, err := Read(vaultPath, title)
	if err != nil {
		return false, err
	}
	meta, body := splitNote(content)
	for _, t := range meta.tags {
		if t == tag {
			return false, nil
		}
	}
	meta.tags = append(meta.tags, tag)
	return true, os.WriteFile(resolvePath(vaultPath, title), []byte(joinNote(meta, body)), 0644)
}

// RemoveTag removes a tag from a note.
// Returns false if the tag was not present (no-op).
func RemoveTag(vaultPath, title, tag string) (bool, error) {
	content, err := Read(vaultPath, title)
	if err != nil {
		return false, err
	}
	meta, body := splitNote(content)
	filtered := meta.tags[:0]
	for _, t := range meta.tags {
		if t != tag {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) == len(meta.tags) {
		return false, nil
	}
	meta.tags = filtered
	return true, os.WriteFile(resolvePath(vaultPath, title), []byte(joinNote(meta, body)), 0644)
}
