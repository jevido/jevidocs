package mcpserver

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPingOverStreamableHTTP(t *testing.T) {
	srv := httptest.NewServer(Handler(New()))
	defer srv.Close()

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             srv.URL,
		DisableStandaloneSSE: true,
	}, nil)
	require.NoError(t, err)
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	var names []string
	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
	}
	assert.Contains(t, names, "ping")
	assert.Contains(t, names, "search_docs")
	assert.Contains(t, names, "create_page")

	assert.Contains(t, names, "get_llms_txt")
	assert.Contains(t, names, "list_versions")

	templates, err := session.ListResourceTemplates(ctx, nil)
	require.NoError(t, err)
	var uris []string
	for _, tpl := range templates.ResourceTemplates {
		uris = append(uris, tpl.URITemplate)
	}
	assert.ElementsMatch(t, []string{pageTemplate, llmsTemplate}, uris)

	prompts, err := session.ListPrompts(ctx, nil)
	require.NoError(t, err)
	var promptNames []string
	for _, pr := range prompts.Prompts {
		promptNames = append(promptNames, pr.Name)
	}
	assert.ElementsMatch(t, []string{"summarize_page", "write_page"}, promptNames)

	got, err := session.GetPrompt(ctx, &mcp.GetPromptParams{Name: "write_page", Arguments: map[string]string{"project": "demo", "topic": "deploys"}})
	require.NoError(t, err)
	require.Len(t, got.Messages, 1)
	assert.Contains(t, got.Messages[0].Content.(*mcp.TextContent).Text, "<Callout")

	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "ping"})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, map[string]any{"message": "pong"}, result.StructuredContent)
}

func TestParseURI(t *testing.T) {
	p, s, err := parseURI("jevidocs://demo/guides/install")
	require.NoError(t, err)
	assert.Equal(t, "demo", p)
	assert.Equal(t, "guides/install", s)
	_, s, _ = parseURI("jevidocs://demo/index")
	assert.Equal(t, "", s)
	_, _, err = parseURI("https://demo/x")
	assert.Error(t, err)
}
