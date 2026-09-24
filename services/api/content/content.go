// Package content embeds jevidocs' own documentation, synced into the
// "jevidocs" project on every start (see store.SyncProject).
package content

import (
	"embed"
	"io/fs"
)

//go:embed all:docs
var files embed.FS

//go:embed openapi/jevidocs.json
var apiSpec []byte

// APISpec is the OpenAPI document of jevidocs' public API, imported into
// the docs as the API reference.
func APISpec() []byte { return apiSpec }

// Docs is the documentation tree, rooted at content/docs.
func Docs() fs.FS {
	sub, err := fs.Sub(files, "docs")
	if err != nil {
		panic(err)
	}
	return sub
}
