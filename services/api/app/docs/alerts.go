package docs

import (
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// GitHub alert syntax, rendered as a Callout:
//
//	> [!WARNING]
//	> Back up first.
//
// The transformer tags the blockquote and removes the marker; the renderer
// (blockquote in markdown.go) writes the callout markup.

var alertRe = regexp.MustCompile(`^\s*\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]`)

var alertTypes = map[string]string{
	"NOTE": "info", "TIP": "idea", "IMPORTANT": "info", "WARNING": "warn", "CAUTION": "error",
}

var alertTitles = map[string]string{
	"NOTE": "Note", "TIP": "Tip", "IMPORTANT": "Important", "WARNING": "Warning", "CAUTION": "Caution",
}

type alertTransformer struct{}

func (alertTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	src := reader.Source()
	var quotes []*ast.Blockquote
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if q, ok := n.(*ast.Blockquote); ok && entering {
			quotes = append(quotes, q)
		}
		return ast.WalkContinue, nil
	})
	for _, q := range quotes {
		para, ok := q.FirstChild().(*ast.Paragraph)
		if !ok || para.Lines().Len() == 0 {
			continue
		}
		first := para.Lines().At(0)
		m := alertRe.FindSubmatchIndex(first.Value(src))
		if m == nil {
			continue
		}
		kind := string(first.Value(src)[m[2]:m[3]])
		end := first.Start + m[1]
		for c := para.FirstChild(); c != nil; {
			next := c.NextSibling()
			if t, ok := c.(*ast.Text); ok {
				switch {
				case t.Segment.Stop <= end:
					para.RemoveChild(para, c)
				case t.Segment.Start < end:
					t.Segment = t.Segment.WithStart(end)
				}
			} else if c.Kind() != ast.KindText {
				// Links etc. never overlap the marker; stop at the first one
				// that starts after it.
				if seg := firstSegment(c); seg.Start >= end {
					break
				}
			}
			c = next
		}
		if para.ChildCount() == 0 {
			q.RemoveChild(q, para)
		} else if t, ok := para.FirstChild().(*ast.Text); ok {
			// The marker usually sits on its own line.
			if len(t.Segment.Value(src)) == 0 || isSpace(t.Segment.Value(src)) {
				t.SetSoftLineBreak(false)
			}
		}
		q.SetAttributeString("data-alert", []byte(kind))
	}
}

func firstSegment(n ast.Node) text.Segment {
	for c := n.FirstChild(); c != nil; c = c.FirstChild() {
		if t, ok := c.(*ast.Text); ok {
			return t.Segment
		}
	}
	return text.NewSegment(1<<30, 1<<30)
}

func isSpace(b []byte) bool {
	for _, c := range b {
		if c != ' ' && c != '\t' {
			return false
		}
	}
	return true
}
