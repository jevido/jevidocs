package store

import (
	"encoding/json"
	"strings"

	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/models"
	"dev.jevido/jevidocs/services/api/app/openapi"
)

// Page kinds. A Markdown page's body is Markdown; an OpenAPI page's body is
// an OpenAPI 3 document (JSON or YAML) that the site shows as an API
// reference.
const (
	KindMarkdown = ""
	KindOpenAPI  = "openapi"
)

// rendered is what a page derives from its body.
type rendered struct {
	docs.Rendered
	API      string // OpenAPI pages: the reference as JSON
	Markdown string // OpenAPI pages: Markdown for agents and llms.txt
	// Title and Summary come from an OpenAPI document's info, as defaults
	// for the page's title and description.
	Title, Summary string
}

// renderBody derives HTML, TOC, search sections and, for OpenAPI pages, the
// reference from a page body (without front matter).
func renderBody(kind, body string) (rendered, error) {
	switch kind {
	case KindMarkdown:
		r, err := docs.Render(body)
		return rendered{Rendered: r}, err
	case KindOpenAPI:
		res, err := openapi.Build([]byte(body))
		if err != nil {
			return rendered{}, ValidationError{err.Error()}
		}
		ref, err := json.Marshal(res.Reference)
		if err != nil {
			return rendered{}, err
		}
		return rendered{Rendered: docs.Rendered{Sections: res.Sections, Plain: res.Plain}, API: string(ref),
			Markdown: res.Markdown, Title: res.Title, Summary: res.Summary}, nil
	}
	return rendered{}, ValidationError{`kind must be "" (Markdown) or "openapi"`}
}

// apply stores r's derived columns on pg.
func (r rendered) apply(pg *models.Page) {
	pg.HTML = r.HTML
	pg.Plain = r.Plain
	toc, _ := json.Marshal(nonNil(r.Toc))
	pg.Toc = string(toc)
	secs, _ := json.Marshal(nonNil(r.Sections))
	pg.Sections = string(secs)
	pg.API = r.API
	pg.Markdown = r.Markdown
	pg.RenderVersion = docs.RenderVersion
}

// MarkdownOf is a page's content as Markdown: the body of a Markdown page,
// the generated text of an OpenAPI page.
func MarkdownOf(pg models.Page) string {
	if pg.Kind == KindOpenAPI {
		return pg.Markdown
	}
	return pg.Body
}

// plainText drops a byte order mark and uses \n line endings, as
// docs.SplitFrontMatter does for Markdown.
func plainText(s string) string {
	return strings.ReplaceAll(strings.TrimPrefix(s, "\ufeff"), "\r\n", "\n")
}

// openAPIFileSuffixes mark spec files in synced folders and exports:
// reference.openapi.yaml is the OpenAPI page at slug "reference".
var openAPIFileSuffixes = []string{".openapi.json", ".openapi.yaml", ".openapi.yml"}

// openAPIFile reports whether file is an OpenAPI page's source and returns
// it renamed to the .md it stands for, so slug and locale rules apply alike.
func openAPIFile(file string) (string, bool) {
	for _, s := range openAPIFileSuffixes {
		if strings.HasSuffix(file, s) {
			return strings.TrimSuffix(file, s) + ".md", true
		}
	}
	return file, false
}
