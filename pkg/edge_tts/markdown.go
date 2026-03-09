package edge_tts

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// ParseMarkdownToText extracts plain text from Markdown, ignoring blockquotes,
// code blocks, and other non-speech elements. It replaces newlines with spaces.
func ParseMarkdownToText(md string) string {
	b := []byte(md)
	reader := text.NewReader(b)
	p := goldmark.DefaultParser()
	doc := p.Parse(reader)

	var buf bytes.Buffer

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			// Add a space after block elements (e.g., paragraphs, headings)
			// to ensure smooth transition between blocks.
			if n.Type() == ast.TypeBlock {
				buf.WriteByte(' ')
			}
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindBlockquote, ast.KindFencedCodeBlock, ast.KindCodeBlock, ast.KindHTMLBlock:
			// Ignore these blocks entirely
			return ast.WalkSkipChildren, nil
		case ast.KindText:
			if t, ok := n.(*ast.Text); ok {
				val := t.Segment.Value(b)
				buf.Write(val)
			}
		case ast.KindString:
			if s, ok := n.(*ast.String); ok {
				buf.Write(s.Value)
			}
		case ast.KindAutoLink:
			if l, ok := n.(*ast.AutoLink); ok {
				buf.Write(l.URL(b))
			}
		}

		return ast.WalkContinue, nil
	})

	res := buf.String()

	// Remove citations like [1], [2, 3] usually found in Wikipedia texts
	reCitations := regexp.MustCompile(`\[.*?\]`)
	res = reCitations.ReplaceAllString(res, " ")

	// Remove remaining markdown characters that could impact speech fluency
	reSpecial := regexp.MustCompile(`[*#_~]+`)
	res = reSpecial.ReplaceAllString(res, " ")

	// Replace all newlines, literal '\n' and multiple spaces with a single space
	res = strings.ReplaceAll(res, "\\n", " ")
	res = strings.ReplaceAll(res, "\n", " ")
	return strings.Join(strings.Fields(res), " ")
}
