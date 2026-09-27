// Package content embeds jevidocs' own documentation, synced into the
// "jevidocs" project on every start (see store.SyncProject).
package content

import (
	"embed"
	"io/fs"
)

//go:embed all:docs
var files embed.FS

// Docs is the documentation tree, rooted at content/docs.
func Docs() fs.FS {
	sub, err := fs.Sub(files, "docs")
	if err != nil {
		panic(err)
	}
	return sub
}
