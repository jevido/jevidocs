package mcpserver

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"dev.jevido/jevidocs/services/api/app/store"
)

// Resources are addressed as jevidocs://{project}/{slug} (a page's Markdown)
// and jevidocs://{project}/llms.txt. The SDK lists only resources registered
// up front and this server is built once, so pages are exposed through
// templates rather than enumerated; list_projects and get_page_tree tell a
// client which URIs exist.
const (
	pageTemplate = "jevidocs://{project}/{+slug}"
	llmsTemplate = "jevidocs://{project}/llms.txt"
)

func registerResources(server *mcp.Server) {
	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "page",
		Title:       "Documentation page",
		Description: "A page's Markdown. slug is the page path, e.g. guides/install; use index for the project's home page.",
		MIMEType:    "text/markdown",
		URITemplate: pageTemplate,
	}, readResource)
	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "llms.txt",
		Title:       "Project llms.txt",
		Description: "The project's llms.txt index: every page with a link and description.",
		MIMEType:    "text/plain",
		URITemplate: llmsTemplate,
	}, readResource)
}

// parseURI splits jevidocs://project/slug into its parts.
func parseURI(raw string) (project, slug string, err error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "jevidocs" || u.Host == "" {
		return "", "", fmt.Errorf("not a jevidocs:// URI: %q", raw)
	}
	slug = strings.Trim(u.Path, "/")
	if slug == "index" {
		slug = ""
	}
	return u.Host, slug, nil
}

func readResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI
	project, slug, err := parseURI(uri)
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	p, err := store.ReadableProject(project, readerOf(req.Extra))
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	if slug == "llms.txt" {
		text, err := store.LLMsTxt(p)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: uri, MIMEType: "text/plain", Text: text}}}, nil
	}
	pg, err := store.FindPage(p, slug)
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
		URI: uri, MIMEType: "text/markdown", Text: store.PageMarkdown(pg, store.PageURL(p, pg.Slug)),
	}}}, nil
}
