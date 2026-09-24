# cli

`jevidocs`, the command-line tool for docs as code: keep a folder of
Markdown in git and push it to a jevidocs project, or pull a project down to
files. Standard library only.

Download a binary from https://jevidocs.jevido.app/downloads/ (for example
`jevidocs-linux-amd64`, `jevidocs-darwin-arm64`, `jevidocs-windows-amd64.exe`;
`checksums.txt` lists SHA-256 sums), or build it: `go build -o jevidocs .`
in this directory.

```sh
jevidocs init docs
JEVIDOCS_TOKEN=jd_... jevidocs push docs -project my-docs -prune
jevidocs pull out -project my-docs
jevidocs search "install" -project jevidocs
```

`-api` / `JEVIDOCS_API` default to https://api.jevidocs.jevido.app. Tokens
come from the admin's "API tokens" page. Files map to slugs the same way as
the API's own `content/` folder: `guides/index.md` is the `guides` folder page.
