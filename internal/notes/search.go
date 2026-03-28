package notes

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// parseTags reads the YAML frontmatter of a note file and returns its tags.
func parseTags(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	inFrontmatter := false
	var tags []string

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		if lineNum == 1 && line == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && line == "---" {
			break
		}
		if inFrontmatter && strings.HasPrefix(line, "tags:") {
			tagStr := strings.TrimPrefix(line, "tags:")
			tagStr = strings.TrimSpace(tagStr)
			tagStr = strings.Trim(tagStr, "[]")
			for _, t := range strings.Split(tagStr, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tags = append(tags, t)
				}
			}
		}
	}
	return tags, scanner.Err()
}

// List returns all note slugs in the vault, optionally filtered by tag.
func List(vaultPath, tag string) ([]string, error) {
	entries, err := os.ReadDir(vaultPath)
	if err != nil {
		return nil, err
	}
	var titles []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		if tag != "" {
			tags, err := parseTags(filepath.Join(vaultPath, e.Name()))
			if err != nil {
				continue
			}
			found := false
			for _, t := range tags {
				if t == tag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		titles = append(titles, slug)
	}
	return titles, nil
}

type searchResult struct {
	slug  string
	score float64
}

// Search searches note titles and content. With exact=true, uses substring match.
// With exact=false, uses bigram similarity scoring and returns results ranked by score.
func Search(vaultPath, query string, exact bool) ([]string, error) {
	entries, err := os.ReadDir(vaultPath)
	if err != nil {
		return nil, err
	}

	var results []searchResult
	queryLower := strings.ToLower(query)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		data, err := os.ReadFile(filepath.Join(vaultPath, e.Name()))
		if err != nil {
			continue
		}
		contentLower := strings.ToLower(string(data))
		slugLower := strings.ToLower(slug)

		if exact {
			if strings.Contains(slugLower, queryLower) || strings.Contains(contentLower, queryLower) {
				results = append(results, searchResult{slug: slug, score: 1.0})
			}
		} else {
			titleScore := bigramSimilarity(query, slug)
			contentScore := bigramSimilarity(query, string(data))
			score := titleScore
			if contentScore > score {
				score = contentScore
			}
			if score > 0.1 {
				results = append(results, searchResult{slug: slug, score: score})
			}
		}
	}

	if !exact {
		sort.Slice(results, func(i, j int) bool {
			return results[i].score > results[j].score
		})
	}

	slugs := make([]string, len(results))
	for i, r := range results {
		slugs[i] = r.slug
	}
	return slugs, nil
}
