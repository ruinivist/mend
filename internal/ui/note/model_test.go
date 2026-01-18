package note

import (
	"reflect"
	"testing"
)

// tests hint extraction from text
func TestExtractHints(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single bold hint",
			content:  "This is a **hint**.",
			expected: []string{"hint"},
		},
		{
			name:     "multiple bold hints",
			content:  "**one** and **two**",
			expected: []string{"one", "two"},
		},
		{
			name:     "underscore hints",
			content:  "__one__ and __two__",
			expected: []string{"one", "two"},
		},
		{
			name:     "mixed hints",
			content:  "**one** and __two__",
			expected: []string{"one", "two"},
		},
		{
			name:     "no hints",
			content:  "just plain text",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractHints(tt.content)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ExtractHints() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// tests section parsing from markdown
func TestParseSections(t *testing.T) {
	content := []byte(`# Title 1
Content 1 with **hint1**.

# Title 2
Content 2 with __hint2__.
`)

	sections := ParseSections(content)

	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	// verify first section
	if sections[0].Title != "# Title 1" {
		t.Errorf("expected title '# Title 1', got '%s'", sections[0].Title)
	}
	if sections[0].Content != "Content 1 with **hint1**." {
		t.Errorf("expected content match for section 1, got '%s'", sections[0].Content)
	}
	if len(sections[0].Hints) != 1 || sections[0].Hints[0] != "hint1" {
		t.Error("failed to extract hint from section 1")
	}

	// verify second section
	if sections[1].Title != "# Title 2" {
		t.Errorf("expected title '# Title 2', got '%s'", sections[1].Title)
	}
	if sections[1].Content != "Content 2 with __hint2__." {
		t.Errorf("expected content match for section 2, got '%s'", sections[1].Content)
	}
	if len(sections[1].Hints) != 1 || sections[1].Hints[0] != "hint2" {
		t.Error("failed to extract hint from section 2")
	}
}

// tests parsing with no headers
func TestParseSectionsNoHeaders(t *testing.T) {
	content := []byte(`Just some content without any headers.
It has **one hint**.`)

	sections := ParseSections(content)

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}

	if sections[0].Title != "no title" {
		t.Errorf("expected 'no title', got '%s'", sections[0].Title)
	}
	if len(sections[0].Hints) != 1 || sections[0].Hints[0] != "one hint" {
		t.Error("failed to extract hint")
	}
}
