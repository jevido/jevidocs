// Package docs turns Markdown pages into what the docs UI renders: HTML with
// fumadocs-style components, a table of contents, searchable sections, and
// the page tree. It is pure Go with no database or framework access, so both
// the REST controllers and the MCP tools use it the same way.
package docs

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// RenderVersion changes whenever the output of Render changes for the same
// input. Stored pages from an older version are re-rendered on start.
const RenderVersion = 3

// TocItem is one heading in a page's table of contents.
type TocItem struct {
	Depth int    `json:"depth"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Section is a heading and the text under it, for heading-level search. The
// first section (before any heading) has an empty ID.
type Section struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

// Rendered is the output of Render.
type Rendered struct {
	HTML     string
	Toc      []TocItem
	Sections []Section
	Plain    string
}

// FrontMatter is the metadata block at the top of a Markdown file.
type FrontMatter struct {
	Title       string
	Description string
	Icon        string
	Section     string
	Position    int
	HasPosition bool
	Root        bool
}

// SplitFrontMatter separates a leading `---` block from the body. Values are
// simple `key: value` lines, optionally quoted; nothing else of YAML is
// needed for page metadata.
func SplitFrontMatter(src string) (FrontMatter, string) {
	var fm FrontMatter
	src = strings.TrimPrefix(src, "\ufeff")
	norm := strings.ReplaceAll(src, "\r\n", "\n")
	if !strings.HasPrefix(norm, "---\n") {
		return fm, norm
	}
	end := strings.Index(norm[4:], "\n---")
	if end < 0 {
		return fm, norm
	}
	head := norm[4 : 4+end]
	body := norm[4+end+4:]
	body = strings.TrimLeft(body, "\n")
	for _, line := range strings.Split(head, "\n") {
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		if len(v) >= 2 && (v[0] == '"' && v[len(v)-1] == '"' || v[0] == '\'' && v[len(v)-1] == '\'') {
			v = v[1 : len(v)-1]
		}
		switch strings.TrimSpace(strings.ToLower(k)) {
		case "title":
			fm.Title = v
		case "description":
			fm.Description = v
		case "icon":
			fm.Icon = v
		case "section":
			fm.Section = v
		case "root":
			fm.Root = v == "true" || v == "yes"
		case "position", "order":
			if n, err := strconv.Atoi(v); err == nil {
				fm.Position, fm.HasPosition = n, true
			}
		}
	}
	return fm, body
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM, extension.Footnote),
	goldmark.WithParserOptions(parser.WithAutoHeadingID(), parser.WithASTTransformers(util.Prioritized(alertTransformer{}, 100))),
	goldmark.WithRendererOptions(
		// Components expand to raw HTML (components.go). Page bodies are
		// written by admins with a token, the same trust as the site itself.
		gmhtml.WithUnsafe(),
		renderer.WithNodeRenderers(util.Prioritized(&docsRenderer{}, 100)),
	),
)

// Render converts a page body (without front matter) to HTML and extracts
// its table of contents and search sections.
func Render(body string) (Rendered, error) {
	src := []byte(expandComponents(body))
	ctx := parser.NewContext(parser.WithIDs(newSlugIDs()))
	doc := md.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))

	var out Rendered
	walkHeadings(doc, src, &out)

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, src, doc); err != nil {
		return out, fmt.Errorf("render markdown: %w", err)
	}
	out.HTML = buf.String()
	out.Sections = sectionsFromHTML(out.HTML)
	var plain []string
	for _, s := range out.Sections {
		plain = append(plain, s.Title, s.Text)
	}
	out.Plain = strings.TrimSpace(strings.Join(plain, "\n"))
	return out, nil
}

func walkHeadings(doc ast.Node, src []byte, out *Rendered) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !entering || !ok {
			return ast.WalkContinue, nil
		}
		id, _ := h.AttributeString("id")
		idStr := string(asBytes(id))
		if h.Level >= 2 && h.Level <= 4 && idStr != "" {
			out.Toc = append(out.Toc, TocItem{Depth: h.Level, Title: nodeText(h, src), URL: "#" + idStr})
		}
		return ast.WalkSkipChildren, nil
	})
}

func asBytes(v any) []byte {
	switch t := v.(type) {
	case []byte:
		return t
	case string:
		return []byte(t)
	}
	return nil
}

func nodeText(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := c.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(src))
			if t.SoftLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(t.Value)
		case *ast.CodeSpan:
			for cc := t.FirstChild(); cc != nil; cc = cc.NextSibling() {
				if tt, ok := cc.(*ast.Text); ok {
					b.Write(tt.Segment.Value(src))
				}
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(b.String())
}

// slugIDs produces GitHub-style heading IDs, unique within a page.
type slugIDs struct{ seen map[string]int }

func newSlugIDs() *slugIDs { return &slugIDs{seen: map[string]int{}} }

func (s *slugIDs) Generate(value []byte, _ ast.NodeKind) []byte {
	id := Slugify(string(value))
	if id == "" {
		id = "heading"
	}
	if n, ok := s.seen[id]; ok {
		s.seen[id] = n + 1
		id = fmt.Sprintf("%s-%d", id, n+1)
	} else {
		s.seen[id] = 0
	}
	return []byte(id)
}

func (s *slugIDs) Put(value []byte) { s.seen[string(value)] = 0 }

// Slugify lowercases s and keeps letters, digits and dashes.
func Slugify(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			dash = false
		case r == ' ' || r == '-' || r == '_':
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// docsRenderer overrides headings (anchor links) and fenced code (Chroma
// highlighting in a titled figure).
type docsRenderer struct{}

func (r *docsRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.heading)
	reg.Register(ast.KindFencedCodeBlock, r.fencedCode)
	reg.Register(ast.KindBlockquote, r.blockquote)
	reg.Register(ast.KindImage, r.image)
}

func (r *docsRenderer) heading(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	h := n.(*ast.Heading)
	id, _ := h.AttributeString("id")
	idStr := html.EscapeString(string(asBytes(id)))
	if entering {
		fmt.Fprintf(w, `<h%d id="%s">`, h.Level, idStr)
		if h.Level > 1 {
			fmt.Fprintf(w, `<a class="fd-anchor" href="#%s">`, idStr)
		}
		return ast.WalkContinue, nil
	}
	if h.Level > 1 {
		_, _ = w.WriteString("</a>")
	}
	fmt.Fprintf(w, "</h%d>\n", h.Level)
	return ast.WalkContinue, nil
}

var metaAttrRe = regexp.MustCompile(`(\w+)=(?:"([^"]*)"|'([^']*)'|(\S+))`)

var formatter = chromahtml.New(chromahtml.WithClasses(true), chromahtml.PreventSurroundingPre(true))

func (r *docsRenderer) fencedCode(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	fc := n.(*ast.FencedCodeBlock)
	info := ""
	if fc.Info != nil {
		info = string(fc.Info.Segment.Value(src))
	}
	meta := parseCodeMeta(info)
	var code bytes.Buffer
	for i := 0; i < fc.Lines().Len(); i++ {
		seg := fc.Lines().At(i)
		code.Write(seg.Value(src))
	}

	// Diagrams are drawn in the browser (apps/site loads Mermaid lazily).
	if strings.EqualFold(meta.lang, "mermaid") {
		fmt.Fprintf(w, "<div class=\"fd-mermaid\"><pre class=\"fd-mermaid-src\">%s</pre></div>\n", html.EscapeString(code.String()))
		return ast.WalkSkipChildren, nil
	}

	fmt.Fprintf(w, `<figure class="fd-codeblock" data-lang="%s"`, html.EscapeString(meta.lang))
	if meta.lineNumbers {
		_, _ = w.WriteString(` data-line-numbers`)
	}
	if hasFocus(meta.lang, code.String()) {
		_, _ = w.WriteString(` data-has-focus`)
	}
	_ = w.WriteByte('>')
	if meta.title != "" {
		fmt.Fprintf(w, `<figcaption class="fd-codeblock-title">%s</figcaption>`, html.EscapeString(meta.title))
	}
	_, _ = w.WriteString(`<pre class="chroma"><code>`)
	writeCode(w, meta, code.String())
	_, _ = w.WriteString("</code></pre></figure>\n")
	return ast.WalkSkipChildren, nil
}

// blockquote renders GitHub alerts (tagged by alertTransformer) as callouts.
func (r *docsRenderer) blockquote(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	kind := ""
	if v, ok := n.AttributeString("data-alert"); ok {
		kind = string(asBytes(v))
	}
	if kind == "" {
		if entering {
			_, _ = w.WriteString("<blockquote>\n")
		} else {
			_, _ = w.WriteString("</blockquote>\n")
		}
		return ast.WalkContinue, nil
	}
	if entering {
		fmt.Fprintf(w, `<div class="fd-callout" data-type="%s"><div class="fd-callout-title">%s</div><div class="fd-callout-body">`+"\n", alertTypes[kind], alertTitles[kind])
	} else {
		_, _ = w.WriteString("</div></div>\n")
	}
	return ast.WalkContinue, nil
}

// image adds lazy loading; the site zooms images on click.
func (r *docsRenderer) image(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	img := n.(*ast.Image)
	_, _ = w.WriteString(`<img src="`)
	_, _ = w.Write(util.EscapeHTML(util.URLEscape(img.Destination, true)))
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.WriteString(html.EscapeString(nodeText(img, src)))
	_ = w.WriteByte('"')
	if len(img.Title) > 0 {
		_, _ = w.WriteString(` title="`)
		_, _ = w.Write(util.EscapeHTML(img.Title))
		_ = w.WriteByte('"')
	}
	_, _ = w.WriteString(` loading="lazy" decoding="async">`)
	return ast.WalkSkipChildren, nil
}

// langAliases maps fence languages Chroma does not know to close relatives.
var langAliases = map[string]string{"mdx": "markdown", "md": "markdown", "svelte": "html", "vue": "html", "env": "bash", "sh": "bash", "shell": "bash", "console": "bash"}

// ChromaCSS returns the stylesheet for the highlight classes in one style,
// for clients that want the exact Chroma palette.
func ChromaCSS(style string) string {
	var b bytes.Buffer
	_ = formatter.WriteCSS(&b, styles.Get(style))
	return b.String()
}
