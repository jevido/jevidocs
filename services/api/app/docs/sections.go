package docs

import (
	"strings"

	"golang.org/x/net/html"
)

// sectionsFromHTML splits rendered HTML at h2-h4 headings into plain-text
// sections. Working on the output rather than the Markdown means component
// markup and code blocks are indexed as the reader sees them.
func sectionsFromHTML(src string) []Section {
	doc, err := html.Parse(strings.NewReader(src))
	if err != nil {
		return nil
	}
	sections := []Section{{}}
	var text strings.Builder
	flush := func() {
		sections[len(sections)-1].Text = collapse(text.String())
		text.Reset()
	}
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h2", "h3", "h4":
				flush()
				sections = append(sections, Section{ID: attr(n, "id"), Title: collapse(innerText(n))})
				return
			case "button", "script", "style":
				return
			}
		}
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
			text.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	flush()
	if sections[0].Text == "" {
		sections = sections[1:]
	}
	return sections
}

func innerText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }
