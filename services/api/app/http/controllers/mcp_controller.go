package controllers

import (
	nethttp "net/http"

	"github.com/goravel/framework/contracts/http"

	"dev.jevido/jevidocs/services/api/app/mcpserver"
)

// McpController mounts the hosted MCP server on a Goravel route.
type McpController struct {
	handler nethttp.Handler
}

func NewMcpController() *McpController {
	return &McpController{handler: mcpserver.Handler(mcpserver.New())}
}

// Serve hands the raw request to the MCP transport, which writes the response
// (JSON or an event stream) itself. Returning nil tells Goravel not to write
// anything on top of it.
func (r *McpController) Serve(ctx http.Context) http.Response {
	r.handler.ServeHTTP(ctx.Response().Writer(), ctx.Request().Origin())
	return nil
}
