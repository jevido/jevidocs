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

	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "ping"})
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, map[string]any{"message": "pong"}, result.StructuredContent)
}
