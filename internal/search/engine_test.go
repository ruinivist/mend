// THIS FILE IS AYE EYE GENERATED
package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsWordMatch(t *testing.T) {
	tests := []struct {
		name   string
		target string
		query  string
		want   bool
	}{
		{"exact word", "hello world", "hello", true},
		{"word in middle", "the hello world", "hello", true},
		{"word at end", "say hello", "hello", true},
		{"partial match", "helloworld", "hello", false},
		{"suffix match", "worldhello", "hello", false},
		{"underscore is word char", "my_hello_world", "hello", false}, // underscore joins words
		{"no match", "goodbye", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := strings.Index(tt.target, tt.query)
			if idx == -1 {
				if tt.want {
					t.Errorf("isWordMatch(%q, %q) expected true but query not found", tt.target, tt.query)
				}
				return
			}
			got := isWordMatch(tt.query, tt.target, idx)
			if got != tt.want {
				t.Errorf("isWordMatch(%q, %q, %d) = %v, want %v", tt.query, tt.target, idx, got, tt.want)
			}
		})
	}
}

func TestSearchEngine(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "search_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	engine := NewSearchEngine()

	// Manually add files for testing
	engine.files = append(engine.files, fileEntry{
		path:         filepath.Join(tmpDir, "folder1", "notes.md"),
		relativePath: "folder1/notes",
		fileName:     "notes",
		content:      "Hello world test content",
	})
	engine.files = append(engine.files, fileEntry{
		path:         filepath.Join(tmpDir, "folder1", "ideas.md"),
		relativePath: "folder1/ideas",
		fileName:     "ideas",
		content:      "Some ideas here",
	})
	engine.files = append(engine.files, fileEntry{
		path:         filepath.Join(tmpDir, "readme.md"),
		relativePath: "readme",
		fileName:     "readme",
		content:      "This is the readme",
	})

	t.Run("search by title exact", func(t *testing.T) {
		results := engine.Search("notes")
		if len(results) == 0 {
			t.Error("expected at least one result for 'notes'")
		}
		if len(results) > 0 && results[0].FileName != "notes" {
			t.Errorf("expected first result to be 'notes', got '%s'", results[0].FileName)
		}
	})

	t.Run("search by title partial", func(t *testing.T) {
		results := engine.Search("not")
		if len(results) == 0 {
			t.Error("expected at least one result for 'not'")
		}
	})

	t.Run("search by content", func(t *testing.T) {
		results := engine.Search("hello")
		if len(results) == 0 {
			t.Error("expected at least one result for 'hello'")
		}
	})

	t.Run("word match scores higher", func(t *testing.T) {
		// "world" is a word in content
		results := engine.Search("world")
		if len(results) == 0 {
			t.Error("expected at least one result")
		}
	})

	t.Run("no results", func(t *testing.T) {
		results := engine.Search("xyznonexistent")
		if len(results) != 0 {
			t.Errorf("expected no results, got %d", len(results))
		}
	})

	t.Run("empty query", func(t *testing.T) {
		results := engine.Search("")
		if results != nil {
			t.Errorf("expected nil for empty query, got %v", results)
		}
	})
}

func TestExtractSnippet(t *testing.T) {
	content := "This is the start of a long piece of content that has a match somewhere in the middle."

	snippet := extractSnippet(content, 40, 20)
	if snippet == "" {
		t.Error("expected non-empty snippet")
	}
}
