package edge_tts

import (
	"testing"
)

func TestParseMarkdownToText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean blocks",
			input:    "# Hello World\nThis is a paragraph.\n\nAnother paragraph.",
			expected: "Hello World This is a paragraph. Another paragraph.",
		},
		{
			name:     "ignore blockquote",
			input:    "Intro.\n\n> This quote should be ignored.\n\nOutro.",
			expected: "Intro. Outro.",
		},
		{
			name:     "ignore code block",
			input:    "Text before.\n```go\nfunc main() {}\n```\nText after.",
			expected: "Text before. Text after.",
		},
		{
			name:     "inline formatting",
			input:    "This has **bold** and *italic* and `code` span.",
			expected: "This has bold and italic and code span.",
		},
		{
			name:     "links",
			input:    "Click [here](https://example.com) to visit.",
			expected: "Click here to visit.",
		},
		{
			name:     "lists with newlines",
			input:    "- Item 1\n- Item 2\n\nNew para",
			expected: "Item 1 Item 2 New para",
		},
		{
			name:     "citations and literals",
			input:    "Text with citation [2, 3].\\n\\nAnd some **bold**.",
			expected: "Text with citation . And some bold.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMarkdownToText(tt.input)
			if result != tt.expected {
				t.Errorf("ParseMarkdownToText() = %q, want %q", result, tt.expected)
			}
		})
	}
}
