/*
Simple exact string match engine.
*/
package search

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const defaultContextLen = 40

// SearchResult represents a single search match
type SearchResult struct {
	Path         string
	RelativePath string
	FileName     string
	Snippet      string
	Score        int
	IsFolder     bool
}

type SearchEngine struct {
	files      []fileEntry
	isIndexing bool
}

// internal struct used while indexing
type fileEntry struct {
	path         string
	relativePath string
	fileName     string
	content      string
	isFolder     bool
}

func NewSearchEngine() *SearchEngine {
	return &SearchEngine{
		files: make([]fileEntry, 0),
	}
}

func (e *SearchEngine) IsIndexing() bool {
	return e.isIndexing
}

// tea cmd for indexing
func StartIndexing(engine *SearchEngine, rootPath string) tea.Cmd {
	return func() tea.Msg {
		engine.isIndexing = true
		engine.files = make([]fileEntry, 0)

		// file walker
		filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip errors
			}

			// Skip hidden files/folders
			if strings.HasPrefix(info.Name(), ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip root path itself
			if path == rootPath {
				return nil
			}

			// Compute relative path from root
			relPath, _ := filepath.Rel(rootPath, path)

			if info.IsDir() {
				// Index folder
				engine.files = append(engine.files, fileEntry{
					path:         path,
					relativePath: relPath,
					fileName:     info.Name(),
					content:      "",
					isFolder:     true,
				})
				return nil
			}

			// file handling
			// Only index .md files
			if !strings.HasSuffix(info.Name(), ".md") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			fileName := strings.TrimSuffix(info.Name(), ".md")
			relPathDisplay := strings.TrimSuffix(relPath, ".md")

			engine.files = append(engine.files, fileEntry{
				path:         path,
				relativePath: relPathDisplay,
				fileName:     fileName,
				content:      strings.ToLower(string(content)),
				isFolder:     false,
			})

			return nil
		})

		engine.isIndexing = false
		return nil
	}
}

func (e *SearchEngine) Search(query string) []SearchResult {
	if query == "" || e.isIndexing {
		return nil
	}

	// to make case insensitive, we don't do it while reading so that
	// display is case correct
	queryLower := strings.ToLower(query)
	results := make([]SearchResult, 0)

	for _, file := range e.files {
		fileNameLower := strings.ToLower(file.fileName)
		contentLower := strings.ToLower(file.content)

		// Check title matches, covers fodler name as well
		matchPos := strings.Index(fileNameLower, queryLower)
		if matchPos >= 0 {
			score := 100 // base title match score
			if fileNameLower == queryLower {
				score += 100 // exact match bonus
			} else if isWordMatch(queryLower, fileNameLower, matchPos) {
				score += 50 // word boundary bonus
			}
			results = append(results, SearchResult{
				Path:         file.path,
				RelativePath: file.relativePath,
				FileName:     file.fileName,
				Snippet:      "",
				Score:        score,
				IsFolder:     file.isFolder,
			})
			continue // else double-count
		}

		// Check content matches on files
		if !file.isFolder {
			matchPos := strings.Index(contentLower, queryLower)
			if matchPos >= 0 {
				score := 10 // base content match score
				if isWordMatch(queryLower, contentLower, matchPos) {
					score += 20 // word boundary bonus
				}
				snippet := extractSnippet(file.content, matchPos, defaultContextLen) // todo: form window width
				results = append(results, SearchResult{
					Path:         file.path,
					RelativePath: file.relativePath,
					FileName:     file.fileName,
					Snippet:      snippet,
					Score:        score,
					IsFolder:     false,
				})
			}
		}
	}

	slices.SortFunc(results, func(a, b SearchResult) int {
		return b.Score - a.Score
	})
	return results
}

// isWordMatch checks if query appears as a complete word in target
func isWordMatch(query, target string, indexInTarget int) bool {
	// Go language note: these ^ strings are not getting copied again like C++, they are just references
	isSpace := func(pos int) bool {
		return pos < 0 || pos >= len(target) || target[pos] == ' ' || target[pos] == '\n'
	}

	return isSpace(indexInTarget-1) && isSpace(indexInTarget+len(query))
}

// extractSnippet extracts context around a match position
func extractSnippet(content string, pos, contextLen int) string {
	start := pos - contextLen
	if start < 0 {
		start = 0
	}
	end := pos + contextLen
	if end > len(content) {
		end = len(content)
	}

	// Adjust to word boundaries
	for start > 0 && content[start] != ' ' && content[start] != '\n' {
		start--
	}
	for end < len(content) && content[end] != ' ' && content[end] != '\n' {
		end++
	}

	snippet := strings.TrimSpace(content[start:end])
	snippet = strings.ReplaceAll(snippet, "\n", " ")

	// Add ellipsis
	prefix := ""
	suffix := ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(content) {
		suffix = "..."
	}

	return prefix + snippet + suffix
}
