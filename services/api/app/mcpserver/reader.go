package mcpserver

import (
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"dev.jevido/jevidocs/services/api/app/store"
)

// readerOf lets read tools and resources see private projects the token's
// user (or a share token in X-Jevidocs-Share) may read.
func readerOf(extra *mcp.RequestExtra) store.Reader {
	if extra == nil || extra.Header == nil {
		return store.Reader{}
	}
	bearer := ""
	if h := extra.Header.Get("Authorization"); len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		bearer = strings.TrimSpace(h[7:])
	}
	return store.ReaderFrom(bearer, extra.Header.Get("X-Jevidocs-Share"))
}
