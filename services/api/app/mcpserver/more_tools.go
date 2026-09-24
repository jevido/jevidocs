package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"dev.jevido/jevidocs/services/api/app/store"
)

type llmsOut struct {
	Text string `json:"text"`
}

type versionsOut struct {
	Versions []store.VersionLink `json:"versions"`
}

func registerMoreTools(server *mcp.Server) {
	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true}

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_llms_txt", Description: "Get a project's llms.txt: every page in reading order with its Markdown URL and description.", Annotations: readOnly,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in projectArg) (*mcp.CallToolResult, llmsOut, error) {
		p, err := store.FindProject(in.Project, false)
		if err != nil {
			return nil, llmsOut{}, err
		}
		text, err := store.LLMsTxt(p)
		return nil, llmsOut{Text: text}, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_versions", Description: "List the versions of a project (other projects in its version group). Empty when it has none.", Annotations: readOnly,
	}, func(ctx context.Context, req *mcp.CallToolRequest, in projectArg) (*mcp.CallToolResult, versionsOut, error) {
		p, err := store.FindProject(in.Project, false)
		if err != nil {
			return nil, versionsOut{}, err
		}
		vs, err := store.Versions(p)
		if vs == nil {
			vs = []store.VersionLink{}
		}
		return nil, versionsOut{Versions: vs}, err
	})
}
