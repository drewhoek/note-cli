package notes

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// bigrams returns a frequency map of overlapping 2-character substrings.
// "hello" → {"he":1, "el":1, "ll":1, "lo":1}
func bigrams(s string) map[string]int {
	s = strings.ToLower(s)
	m := make(map[string]int)
	for i := 0; i < len(s)-1; i++ {
		m[s[i:i+2]]++
	}
	return m
}

// bigramSimilarity returns a score between 0.0 and 1.0 representing how similar
// two strings are based on shared bigrams. 1.0 = identical, 0.0 = nothing in common.
// Formula: (2 × shared bigrams) / (total bigrams in a + total bigrams in b)
func bigramSimilarity(a, b string) float64 {
	// Return 0 early if either string is shorter than 2 characters
	if len(a) < 2 || len(b) < 2 {
		return 0
	}

	// Get the bigram maps for both strings using bigrams()
	aBigram := bigrams(a)
	bBigram := bigrams(b)

	// Count the intersection (shared bigrams, taking the minimum frequency for each)
	intersection := 0
	for bigram, aCount := range aBigram { // this is foreach syntax basically
		if bCount, ok := bBigram[bigram]; ok { // finding a value in a map returns the value and a boolean value of it was found or not, 0 for the count and false for ok if not found
			intersection += min(aCount, bCount)
		}
	}

	// Count the total bigrams across both strings
	total := 0
	for _, v := range aBigram {
		total += v
	}
	for _, v := range bBigram {
		total += v
	}

	return float64(2*intersection) / float64(total)
}

// parseTags returns the tags from a note file at the given path.
func parseTags(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	meta, _ := splitNote(string(data))
	return meta.tags, nil
}

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

			// Always read frontmatter: we need it to filter by tag, status, or
			// to apply the default behaviour of hiding stale notes.
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
			if status != "" {
				// Explicit status filter: only show notes matching that status.
				if meta.status != status {
					continue
				}
			} else if dir == vaultPath {
				// Default (active vault only): exclude stale notes.
				// Notes in the archive dir are shown as-is when includeArchived is set.
				if meta.status == "stale" {
					continue
				}
			}
			titles = append(titles, slug)
		}
	}
	return titles, nil
}

type searchResult struct {
	slug  string
	score float64
}

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
				meta, _ := splitNote(content)
				if meta.status == "stale" && dir == vaultPath {
					continue
				}
				if strings.Contains(slugLower, queryLower) || strings.Contains(contentLower, queryLower) {
					score = 1.0
				}
			} else {
				meta, body := splitNote(content)
				// Exclude stale notes from default fuzzy search (consistent with List behavior).
				// Archived dirs are not filtered — caller opted in with includeArchived.
				if meta.status == "stale" && dir == vaultPath {
					continue
				}
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

// Search searches note slugs only.
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

// Outlinks returns all [[wikilink]] targets found in the given note.
func Outlinks(vaultPath, title string) ([]string, error) {
	content, err := Read(vaultPath, title)
	if err != nil {
		return nil, err
	}
	var links []string
	rest := content
	for {
		start := strings.Index(rest, "[[")
		if start == -1 {
			break
		}
		rest = rest[start+2:]
		end := strings.Index(rest, "]]")
		if end == -1 {
			break
		}
		links = append(links, rest[:end])
		rest = rest[end+2:]
	}
	sort.Strings(links)
	return links, nil
}

// Backlinks returns the slugs of all notes that contain a [[wikilink]] to the given title.
// Matches both [[Title]] and [[slug]] forms.
func Backlinks(vaultPath, title string) ([]string, error) {
	slug := Slug(title)
	targets := []string{"[[" + title + "]]", "[[" + slug + "]]"}

	entries, err := os.ReadDir(vaultPath)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		noteSlug := strings.TrimSuffix(e.Name(), ".md")
		if noteSlug == slug {
			continue // skip the note itself
		}
		data, err := os.ReadFile(filepath.Join(vaultPath, e.Name()))
		if err != nil {
			continue
		}
		content := string(data)
		for _, target := range targets {
			if strings.Contains(content, target) {
				results = append(results, noteSlug)
				break
			}
		}
	}

	sort.Strings(results)
	return results, nil
}

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
