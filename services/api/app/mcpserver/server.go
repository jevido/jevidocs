// Package mcpserver is the hosted MCP server. Agents use it to read and
// search documentation, and, with an admin API token, to write pages. Tools
// stay thin: they validate input and call the same store code the REST
// routes use.
package mcpserver

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New builds the MCP server with every hosted tool registered.
func New() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "jevidocs",
		Version: "0.1.0",
	}, &mcp.ServerOptions{
		Instructions: "jevidocs hosts documentation sites. Use list_projects to find a project, " +
			"get_page_tree for its structure, search_docs to find pages, and read_page for a page's Markdown. " +
			"Pages are also resources: jevidocs://{project}/{slug} and jevidocs://{project}/llms.txt. " +
			"create_page, update_page and delete_page need an admin API token as a Bearer token.",
	})

	registerPing(server)
	registerDocsTools(server)
	registerMoreTools(server)
	registerResources(server)
	registerPrompts(server)

	return server
}

// Handler serves the server over MCP's streamable HTTP transport.
//
// Stateless: every tool call is a self-contained request, so no session state
// lives in the API process and it can restart or scale without breaking
// clients.
//
// JSONResponse: Goravel wraps every route in a global timeout middleware
// (http.request_timeout) that buffers the response. A flushed event stream
// escapes that buffer before its headers are copied, so clients receive SSE
// without a Content-Type and reject it. Plain JSON responses are written once
// and survive intact. Consequence: each tool call must finish within the
// request timeout.
func Handler(server *mcp.Server) http.Handler {
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
}

type pingOutput struct {
	Message string `json:"message" jsonschema:"always \"pong\""`
}

// registerPing adds a no-op tool so clients can check they are connected.
func registerPing(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ping",
		Description: "Check that the jevidocs server is reachable.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, pingOutput, error) {
		return nil, pingOutput{Message: "pong"}, nil
	})
}
