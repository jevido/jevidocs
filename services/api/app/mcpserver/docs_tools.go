package mcpserver

import (
	"context"
	"errors"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"dev.jevido/jevidocs/services/api/app/docs"
	"dev.jevido/jevidocs/services/api/app/models"
	"dev.jevido/jevidocs/services/api/app/store"
)

type projectArg struct {
	Project string `json:"project" jsonschema:"project slug, e.g. jevidocs"`
}

// Projects and trees are recursive types, which the SDK cannot derive a
// JSON schema for, so they go out untyped.
type listProjectsOut struct {
	Projects any `json:"projects"`
}

type treeOut struct {
	Tree any `json:"tree"`
}

type readPageArgs struct {
	Project string `json:"project" jsonschema:"project slug"`
	Locale  string `json:"locale,omitempty" jsonschema:"language, e.g. nl; default language when omitted"`
	Slug    string `json:"slug" jsonschema:"page slug, e.g. guides/install; empty for the index page"`
}

type readPageOut struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Markdown    string `json:"markdown"`
}

type searchArgs struct {
	Project string `json:"project" jsonschema:"project slug"`
	Locale  string `json:"locale,omitempty" jsonschema:"language, e.g. nl; default language when omitted"`
	Query   string `json:"query" jsonschema:"search words"`
}

type searchOut struct {
	Results []store.SearchResult `json:"results"`
}

type writePageArgs struct {
	Project     string `json:"project" jsonschema:"project slug"`
	Slug        string `json:"slug" jsonschema:"page slug path, e.g. guides/install; empty for the index"`
	Title       string `json:"title,omitempty" jsonschema:"page title (required when creating unless the body has front matter)"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Position    int    `json:"position,omitempty" jsonschema:"sort order among siblings"`
	Section     string `json:"section,omitempty" jsonschema:"sidebar separator label for top-level pages"`
	Body        string `json:"body,omitempty" jsonschema:"Markdown body; supports Callout, Cards, Tabs, Steps, Accordion, Files components"`
	Published   *bool  `json:"published,omitempty"`
}

type pageRefArgs struct {
	Project string `json:"project" jsonschema:"project slug"`
	Slug    string `json:"slug" jsonschema:"page slug"`
}

type writeOut struct {
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

var errAuth = errors.New("this tool needs an admin API token: send Authorization: Bearer <token>")

// authorize checks the bearer token on the MCP HTTP request.
func authorize(req *mcp.CallToolRequest) (models.User, error) {
	if req.Extra == nil || req.Extra.Header == nil {
		return models.User{}, errAuth
	}
	h := req.Extra.Header.Get("Authorization")
	if len(h) < 8 || !strings.EqualFold(h[:7], "bearer ") {
		return models.User{}, errAuth
	}
	u, _, err := store.UserForToken(h[7:])
	if err != nil {
		return u, errAuth
	}
	if !store.HasRole(u, store.RoleEditor) {
		return u, errors.New("this tool needs an editor or admin token; this token's user is a " + store.RoleOf(u))
	}
	return u, nil
}

func registerDocsTools(server *mcp.Server) {
	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true}

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_projects", Description: "List the public documentation projects.", Annotations: readOnly,
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listProjectsOut, error) {
		ps, err := store.ReadableProjects(readerOf(req.Extra))
		return nil, listProjectsOut{Projects: ps}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_page_tree", Description: "Get a project's page tree (sidebar): folders, pages and separators with their slugs.", Annotations: readOnly,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in projectArg) (*mcp.CallToolResult, treeOut, error) {
		p, err := store.ReadableProject(in.Project, readerOf(req.Extra))
		if err != nil {
			return nil, treeOut{}, err
		}
		t, err := store.Tree(p)
		return nil, treeOut{Tree: t}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "read_page", Description: "Read one documentation page as Markdown.", Annotations: readOnly,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in readPageArgs) (*mcp.CallToolResult, readPageOut, error) {
		p, err := store.ReadableProject(in.Project, readerOf(req.Extra))
		if err != nil {
			return nil, readPageOut{}, err
		}
		p = store.WithLocale(p, in.Locale)
		pg, err := store.FindPage(p, in.Slug)
		if err != nil {
			return nil, readPageOut{}, err
		}
		return nil, readPageOut{Title: pg.Title, Description: pg.Description, URL: store.PageURL(p, pg.Slug), Markdown: store.MarkdownOf(pg)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "search_docs", Description: "Full-text search a project's documentation. Returns pages and headings with snippets.", Annotations: readOnly,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in searchArgs) (*mcp.CallToolResult, searchOut, error) {
		p, err := store.ReadableProject(in.Project, readerOf(req.Extra))
		if err != nil {
			return nil, searchOut{}, err
		}
		p = store.WithLocale(p, in.Locale)
		res, err := store.Search(p, in.Query, 15)
		return nil, searchOut{Results: res}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "create_page", Description: "Create a documentation page. Requires an admin API token.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in writePageArgs) (*mcp.CallToolResult, writeOut, error) {
		u, err := authorize(req)
		if err != nil {
			return nil, writeOut{}, err
		}
		p, err := store.FindProject(in.Project, true)
		if err != nil {
			return nil, writeOut{}, err
		}
		pg, err := store.SavePage(p, store.PageInput{Slug: in.Slug, Title: in.Title, Description: in.Description,
			Icon: in.Icon, Position: in.Position, Section: in.Section, Body: in.Body, Published: in.Published}, nil)
		if err != nil {
			return nil, writeOut{}, err
		}
		store.MarkEditedBy(pg.ID, u.ID)
		return nil, writeOut{Slug: pg.Slug, URL: store.PageURL(p, pg.Slug)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_page", Description: "Update a documentation page by slug. Omitted fields keep their value. Requires an admin API token.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in writePageArgs) (*mcp.CallToolResult, writeOut, error) {
		u, err := authorize(req)
		if err != nil {
			return nil, writeOut{}, err
		}
		p, err := store.FindProject(in.Project, true)
		if err != nil {
			return nil, writeOut{}, err
		}
		cur, err := findAnyPage(p, in.Slug)
		if err != nil {
			return nil, writeOut{}, err
		}
		upd := store.PageInput{Slug: cur.Slug, Title: pick(in.Title, cur.Title), Description: pick(in.Description, cur.Description),
			Icon: pick(in.Icon, cur.Icon), Position: cur.Position, Section: pick(in.Section, cur.Section), Body: pick(in.Body, cur.Body), Published: in.Published}
		if in.Position != 0 {
			upd.Position = in.Position
		}
		pg, err := store.SavePage(p, upd, &cur)
		if err != nil {
			return nil, writeOut{}, err
		}
		store.MarkEditedBy(pg.ID, u.ID)
		return nil, writeOut{Slug: pg.Slug, URL: store.PageURL(p, pg.Slug)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "delete_page", Description: "Delete a documentation page by slug. Requires an admin API token.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(true)},
	}, func(ctx context.Context, req *mcp.CallToolRequest, in pageRefArgs) (*mcp.CallToolResult, writeOut, error) {
		if _, err := authorize(req); err != nil {
			return nil, writeOut{}, err
		}
		p, err := store.FindProject(in.Project, true)
		if err != nil {
			return nil, writeOut{}, err
		}
		cur, err := findAnyPage(p, in.Slug)
		if err != nil {
			return nil, writeOut{}, err
		}
		return nil, writeOut{Slug: cur.Slug}, store.DeletePage(p, cur)
	})
}

func findAnyPage(p models.Project, slug string) (models.Page, error) {
	pages, err := store.AdminPages(p)
	if err != nil {
		return models.Page{}, err
	}
	want := docs.NormalizeSlug(slug)
	for _, pg := range pages {
		if pg.Slug == want {
			return store.FindPageByID(p, pg.ID)
		}
	}
	return models.Page{}, store.ErrNotFound
}

func pick(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

func ptr[T any](v T) *T { return &v }
