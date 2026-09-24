# cli

`jevidocs`, the command-line tool for docs as code: keep a folder of
Markdown in git and push it to a jevidocs project, or pull a project down to
files. Standard library only.

```sh
go install dev.jevido/jevidocs/apps/cli@latest   # or: go build -o jevidocs .
jevidocs init docs
JEVIDOCS_TOKEN=jd_... jevidocs push docs -project my-docs -prune
jevidocs pull out -project my-docs
jevidocs search "install" -project jevidocs
```

`-api` / `JEVIDOCS_API` default to https://api.jevidocs.jevido.app. Tokens
come from the admin's "API tokens" page. Files map to slugs the same way as
the API's own `content/` folder: `guides/index.md` is the `guides` folder page.
