// Command jevidocs keeps a folder of Markdown in sync with a jevidocs
// project: docs as code, with the admin and MCP still able to read the
// result.
//
//	jevidocs init   [dir]             scaffold a docs folder
//	jevidocs push   [dir] -project p  upload every .md file (optionally prune)
//	jevidocs pull   [dir] -project p  download every page as .md files
//	jevidocs search <query> -project p
//	jevidocs openapi <spec> -project p [-prefix api-reference]
//
// The API address and token come from -api / -token or JEVIDOCS_API /
// JEVIDOCS_TOKEN. Tokens are created in the admin under "API tokens".
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const defaultAPI = "https://api.jevidocs.jevido.app"

type client struct {
	api   string
	token string
	http  *http.Client
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "init":
		err = runInit(args)
	case "push":
		err = runPush(args)
	case "pull":
		err = runPull(args)
	case "search":
		err = runSearch(args)
	case "openapi":
		err = runOpenAPI(args)
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "jevidocs:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `jevidocs: sync Markdown folders with a jevidocs project

Usage:
  jevidocs init   [dir]
  jevidocs push   [dir] -project <slug> [-prune]
  jevidocs pull   [dir] -project <slug>
  jevidocs search <query> -project <slug>
  jevidocs openapi <spec.json|yaml> -project <slug> [-prefix api-reference]

Flags for push, pull, search and openapi:
  -api    API address (env JEVIDOCS_API, default `+defaultAPI+`)
  -token  API token (env JEVIDOCS_TOKEN; not needed for search)
`)
}

// commonFlags parses flags that may appear before or after the positional
// argument, since `jevidocs push docs -project x` reads naturally.
func commonFlags(name string, args []string, extra func(*flag.FlagSet)) (*client, string, string, error) {
	fsFlags := flag.NewFlagSet(name, flag.ContinueOnError)
	api := fsFlags.String("api", envOr("JEVIDOCS_API", defaultAPI), "API address")
	token := fsFlags.String("token", os.Getenv("JEVIDOCS_TOKEN"), "API token")
	project := fsFlags.String("project", "", "project slug")
	if extra != nil {
		extra(fsFlags)
	}
	var positional []string
	for len(args) > 0 {
		if err := fsFlags.Parse(args); err != nil {
			return nil, "", "", err
		}
		args = fsFlags.Args()
		if len(args) > 0 {
			positional = append(positional, args[0])
			args = args[1:]
		}
	}
	c := &client{api: strings.TrimRight(*api, "/"), token: *token, http: &http.Client{Timeout: 60 * time.Second}}
	return c, *project, strings.Join(positional, " "), nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func (c *client) do(method, path string, body, out any) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.api+path, r)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		var e struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &e) == nil && e.Error != "" {
			return fmt.Errorf("%s %s: %s (%d)", method, path, e.Error, res.StatusCode)
		}
		return fmt.Errorf("%s %s: HTTP %d", method, path, res.StatusCode)
	}
	if out != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

func runPush(args []string) error {
	var prune bool
	c, project, dir, err := commonFlags("push", args, func(f *flag.FlagSet) {
		f.BoolVar(&prune, "prune", false, "delete pages that have no file")
	})
	if err != nil {
		return err
	}
	if project == "" || c.token == "" {
		return errors.New("push needs -project and a token (-token or JEVIDOCS_TOKEN)")
	}
	if dir == "" {
		dir = "."
	}
	files := map[string]string{}
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") && p != dir {
			return filepath.SkipDir
		}
		ext := filepath.Ext(p)
		if d.IsDir() || (ext != ".md" && ext != ".mdx") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		files[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .md files in %s", dir)
	}
	var res struct{ Created, Updated, Unchanged, Deleted int }
	if err := c.do("PUT", "/api/admin/projects/"+url.PathEscape(project)+"/sync", map[string]any{"files": files, "prune": prune}, &res); err != nil {
		return err
	}
	fmt.Printf("pushed %d files to %s: %d created, %d updated, %d unchanged, %d deleted\n",
		len(files), project, res.Created, res.Updated, res.Unchanged, res.Deleted)
	return nil
}

type adminPage struct {
	ID          int    `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Position    int    `json:"position"`
	Section     string `json:"section"`
	Published   bool   `json:"published"`
	Body        string `json:"body"`
}

func runPull(args []string) error {
	c, project, dir, err := commonFlags("pull", args, nil)
	if err != nil {
		return err
	}
	if project == "" || c.token == "" {
		return errors.New("pull needs -project and a token (-token or JEVIDOCS_TOKEN)")
	}
	if dir == "" {
		dir = project
	}
	var pages []adminPage
	base := "/api/admin/projects/" + url.PathEscape(project) + "/pages"
	if err := c.do("GET", base, nil, &pages); err != nil {
		return err
	}
	// Slugs that have children are folders; their page is folder/index.md.
	parents := map[string]bool{}
	for _, p := range pages {
		for s := p.Slug; strings.Contains(s, "/"); {
			s = s[:strings.LastIndex(s, "/")]
			parents[s] = true
		}
	}
	for _, p := range pages {
		var full adminPage
		if err := c.do("GET", base+"/"+strconv.Itoa(p.ID), nil, &full); err != nil {
			return err
		}
		name := p.Slug + ".md"
		if p.Slug == "" {
			name = "index.md"
		} else if parents[p.Slug] {
			name = p.Slug + "/index.md"
		}
		target := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(withFrontMatter(full)), 0o644); err != nil {
			return err
		}
		fmt.Println("wrote", target)
	}
	return nil
}

func withFrontMatter(p adminPage) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", quote(p.Title))
	if p.Description != "" {
		fmt.Fprintf(&b, "description: %s\n", quote(p.Description))
	}
	if p.Icon != "" {
		fmt.Fprintf(&b, "icon: %s\n", p.Icon)
	}
	if p.Position != 0 {
		fmt.Fprintf(&b, "position: %d\n", p.Position)
	}
	if p.Section != "" {
		fmt.Fprintf(&b, "section: %s\n", quote(p.Section))
	}
	b.WriteString("---\n\n")
	b.WriteString(strings.TrimLeft(p.Body, "\n"))
	if !strings.HasSuffix(p.Body, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

func quote(s string) string {
	if strings.ContainsAny(s, ":#\"'") {
		return strconv.Quote(s)
	}
	return s
}

func runSearch(args []string) error {
	c, project, q, err := commonFlags("search", args, nil)
	if err != nil {
		return err
	}
	if project == "" || q == "" {
		return errors.New("search needs a query and -project")
	}
	var res []struct {
		Type, Title, URL string
		PageTitle        string `json:"page_title"`
	}
	if err := c.do("GET", "/api/projects/"+url.PathEscape(project)+"/search?q="+url.QueryEscape(q), nil, &res); err != nil {
		return err
	}
	for _, r := range res {
		label := r.Title
		if r.Type == "heading" {
			label = r.PageTitle + " › " + r.Title
		}
		fmt.Printf("%-50s %s\n", label, r.URL)
	}
	if len(res) == 0 {
		fmt.Println("no results")
	}
	return nil
}

var scaffold = map[string]string{
	"index.md": `---
title: Introduction
description: Welcome to your documentation.
---

Welcome! Edit this file, then run ` + "`jevidocs push`" + `.

<Cards>
<Card title="Quick start" href="./quick-start" description="Get going in a minute" />
<Card title="Guides" href="./guides" description="Everything else" />
</Cards>
`,
	"quick-start.md": `---
title: Quick start
description: From zero to published.
position: 1
---

<Steps>
<Step>

### Write

Add Markdown files to this folder. Folders become sidebar sections.

</Step>
<Step>

### Push

` + "```sh title=\"terminal\"\njevidocs push . -project my-docs\n```" + `

</Step>
</Steps>

<Callout type="info">
Front matter sets the title, description, position and section of a page.
</Callout>
`,
	"guides/index.md": `---
title: Guides
position: 2
---

Guides live in this folder.
`,
}

func runInit(args []string) error {
	dir := "docs"
	if len(args) > 0 {
		dir = args[0]
	}
	for name, body := range scaffold {
		target := filepath.Join(dir, filepath.FromSlash(name))
		if _, err := os.Stat(target); err == nil {
			fmt.Println("exists, skipped", target)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
			return err
		}
		fmt.Println("wrote", target)
	}
	fmt.Printf("\nNext: create a project in the admin, then\n  JEVIDOCS_TOKEN=... jevidocs push %s -project <slug>\n", dir)
	return nil
}

// runOpenAPI generates API reference pages from an OpenAPI 3 document.
func runOpenAPI(args []string) error {
	var prefix string
	c, project, file, err := commonFlags("openapi", args, func(f *flag.FlagSet) {
		f.StringVar(&prefix, "prefix", "api-reference", "slug prefix for the generated pages")
	})
	if err != nil {
		return err
	}
	if project == "" || c.token == "" || file == "" {
		return errors.New("openapi needs a spec file, -project and a token (-token or JEVIDOCS_TOKEN)")
	}
	spec, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	var res struct{ Created, Updated, Unchanged, Deleted int }
	if err := c.do("POST", "/api/admin/projects/"+url.PathEscape(project)+"/openapi", map[string]any{"spec": string(spec), "prefix": prefix}, &res); err != nil {
		return err
	}
	fmt.Printf("imported %s into %s/%s: %d created, %d updated, %d unchanged, %d deleted\n",
		file, project, prefix, res.Created, res.Updated, res.Unchanged, res.Deleted)
	return nil
}
