// Package openapi turns an OpenAPI 3.0/3.1 document into documentation
// pages, the way fumadocs-openapi generates an API reference: an index page,
// a folder per tag and a page per operation. The output is plain Markdown
// with our components, so the pages are stored, searched and served like any
// other page.
package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"dev.jevido/jevidocs/services/api/app/docs"
)

// Page is one generated documentation page.
type Page struct {
	Slug        string
	Title       string
	Description string
	Position    int
	Body        string
}

// DefaultPrefix is where generated pages go when no prefix is given.
const DefaultPrefix = "api-reference"

var methods = []string{"get", "post", "put", "patch", "delete", "head", "options", "trace"}

// spec wraps the decoded document for $ref resolution.
type spec struct {
	root map[string]any
	base string // reader URL of the prefix page
}

// Generate parses src (JSON or YAML) and returns the pages under prefix.
// linkBase is the reader path of the project's root (e.g. "/docs" or
// "/p/petstore"), used for links between generated pages.
func Generate(src []byte, prefix, linkBase string) ([]Page, error) {
	var raw any
	// YAML is a superset of JSON, so one decoder handles both.
	if err := yaml.Unmarshal(src, &raw); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	root, _ := normalize(raw).(map[string]any)
	if root == nil {
		return nil, errors.New("spec is empty")
	}
	if v, _ := root["openapi"].(string); !strings.HasPrefix(v, "3.") {
		return nil, errors.New("only OpenAPI 3.x documents are supported (missing or unsupported \"openapi\" version)")
	}
	prefix = docs.NormalizeSlug(prefix)
	if prefix == "" {
		prefix = DefaultPrefix
	}
	s := &spec{root: root, base: strings.TrimRight(linkBase, "/") + "/" + prefix}
	return s.pages(prefix), nil
}

type operation struct {
	method, path string
	op           map[string]any
	pathParams   []any
	slug, title  string
	tag          string
}

func (s *spec) pages(prefix string) []Page {
	info := asMap(root(s, "info"))
	title := str(info["title"])
	if title == "" {
		title = "API reference"
	}
	server := s.serverURL()

	// Collect operations grouped by their first tag, in spec order of paths.
	paths := asMap(root(s, "paths"))
	pathKeys := sortedKeys(paths)
	groups := map[string][]*operation{}
	usedSlugs := map[string]bool{}
	for _, p := range pathKeys {
		item := asMap(s.deref(paths[p], nil))
		for _, m := range methods {
			raw, ok := item[m]
			if !ok {
				continue
			}
			op := asMap(raw)
			tag := "default"
			if tags, ok := op["tags"].([]any); ok && len(tags) > 0 && str(tags[0]) != "" {
				tag = str(tags[0])
			}
			o := &operation{method: m, path: p, op: op, tag: tag, pathParams: asSlice(item["parameters"])}
			o.title = firstNonEmpty(str(op["summary"]), strings.ToUpper(m)+" "+p)
			base := docs.Slugify(str(op["operationId"]))
			if base == "" {
				base = docs.Slugify(m + " " + strings.NewReplacer("/", " ", "{", "", "}", "").Replace(p))
			}
			slug := base
			for i := 2; usedSlugs[docs.Slugify(tag)+"/"+slug]; i++ {
				slug = fmt.Sprintf("%s-%d", base, i)
			}
			usedSlugs[docs.Slugify(tag)+"/"+slug] = true
			o.slug = slug
			groups[tag] = append(groups[tag], o)
		}
	}

	// Tag order: as declared in `tags`, then any others alphabetically.
	var tagOrder []string
	tagDesc := map[string]string{}
	seen := map[string]bool{}
	for _, t := range asSlice(root(s, "tags")) {
		tm := asMap(t)
		name := str(tm["name"])
		if name == "" || seen[name] {
			continue
		}
		tagDesc[name] = str(tm["description"])
		if len(groups[name]) > 0 {
			tagOrder = append(tagOrder, name)
			seen[name] = true
		}
	}
	for _, name := range sortedKeys(groups) {
		if !seen[name] {
			tagOrder = append(tagOrder, name)
		}
	}

	var out []Page
	var idx strings.Builder
	if v := str(info["version"]); v != "" {
		fmt.Fprintf(&idx, "Version `%s`\n\n", v)
	}
	if d := str(info["description"]); d != "" {
		idx.WriteString(strings.TrimSpace(d) + "\n\n")
	}
	if servers := asSlice(root(s, "servers")); len(servers) > 0 {
		idx.WriteString("## Servers\n\n")
		for _, sv := range servers {
			m := asMap(sv)
			line := "- `" + str(m["url"]) + "`"
			if d := str(m["description"]); d != "" {
				line += ": " + oneLine(d)
			}
			idx.WriteString(line + "\n")
		}
		idx.WriteString("\n")
	}
	idx.WriteString("## Endpoints\n\n<Cards>\n")
	for _, t := range tagOrder {
		desc := firstNonEmpty(oneLine(tagDesc[t]), fmt.Sprintf("%d endpoint%s", len(groups[t]), plural(len(groups[t]))))
		fmt.Fprintf(&idx, "<Card title=\"%s\" href=\"%s\" description=\"%s\" />\n", attr(t), s.base+"/"+docs.Slugify(t), attr(desc))
	}
	idx.WriteString("</Cards>\n")
	out = append(out, Page{Slug: prefix, Title: title, Description: firstLine(str(info["description"])), Position: 100, Body: idx.String()})

	for ti, t := range tagOrder {
		tagSlug := prefix + "/" + docs.Slugify(t)
		var b strings.Builder
		if d := tagDesc[t]; d != "" {
			b.WriteString(strings.TrimSpace(d) + "\n\n")
		}
		b.WriteString("| Method | Path | Summary |\n| --- | --- | --- |\n")
		for _, o := range groups[t] {
			fmt.Fprintf(&b, "| `%s` | [`%s`](%s/%s/%s) | %s |\n", strings.ToUpper(o.method), cell(o.path), s.base, docs.Slugify(t), o.slug, cell(str(o.op["summary"])))
		}
		out = append(out, Page{Slug: tagSlug, Title: t, Description: firstLine(tagDesc[t]), Position: ti + 1, Body: b.String()})
		for oi, o := range groups[t] {
			out = append(out, Page{Slug: tagSlug + "/" + o.slug, Title: o.title,
				Description: firstLine(str(o.op["description"])), Position: oi + 1, Body: s.operationBody(o, server)})
		}
	}
	return out
}

func (s *spec) serverURL() string {
	for _, sv := range asSlice(root(s, "servers")) {
		if u := str(asMap(sv)["url"]); u != "" {
			return strings.TrimRight(u, "/")
		}
	}
	return "https://api.example.com"
}

func (s *spec) operationBody(o *operation, server string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<div class=\"fd-api-endpoint\"><span class=\"fd-api-method\" data-method=\"%s\">%s</span><code class=\"fd-api-path\">%s</code></div>\n\n",
		o.method, strings.ToUpper(o.method), htmlEscape(o.path))
	if dep, _ := o.op["deprecated"].(bool); dep {
		b.WriteString("<Callout type=\"warn\" title=\"Deprecated\">\nThis operation is deprecated and may be removed.\n</Callout>\n\n")
	}
	if d := str(o.op["description"]); d != "" {
		b.WriteString(strings.TrimSpace(d) + "\n\n")
	}

	// Parameters: path-level ones, overridden by operation-level ones.
	type param struct {
		name, in, typ, desc string
		req                 bool
	}
	var params []param
	idx := map[string]int{}
	for _, raw := range append(append([]any{}, o.pathParams...), asSlice(o.op["parameters"])...) {
		pm := asMap(s.deref(raw, nil))
		name, in := str(pm["name"]), str(pm["in"])
		if name == "" {
			continue
		}
		req, _ := pm["required"].(bool)
		p := param{name, in, s.typeName(pm["schema"], nil), oneLine(str(pm["description"])), req || in == "path"}
		if i, ok := idx[in+":"+name]; ok {
			params[i] = p
		} else {
			idx[in+":"+name] = len(params)
			params = append(params, p)
		}
	}
	if len(params) > 0 {
		b.WriteString("## Parameters\n\n| Name | In | Type | Required | Description |\n| --- | --- | --- | --- | --- |\n")
		for _, p := range params {
			fmt.Fprintf(&b, "| `%s` | %s | `%s` | %s | %s |\n", cell(p.name), p.in, cell(p.typ), yesNo(p.req), cell(p.desc))
		}
		b.WriteString("\n")
	}

	var bodyExample any
	hasBody := false
	if rb := asMap(s.deref(o.op["requestBody"], nil)); len(rb) > 0 {
		ct, media := firstMedia(asMap(rb["content"]))
		b.WriteString("## Request body\n\n")
		if ct != "" {
			fmt.Fprintf(&b, "Content type: `%s`", ct)
			if req, _ := rb["required"].(bool); req {
				b.WriteString(" (required)")
			}
			b.WriteString("\n\n")
		}
		if d := str(rb["description"]); d != "" {
			b.WriteString(strings.TrimSpace(d) + "\n\n")
		}
		if schema := media["schema"]; schema != nil {
			b.WriteString(s.schemaTable(schema))
			hasBody = true
			bodyExample = firstNonNil(media["example"], exampleFromExamples(media["examples"]), s.example(schema, nil, 0))
		}
	}

	if resps := asMap(o.op["responses"]); len(resps) > 0 {
		b.WriteString("## Responses\n\n")
		for _, code := range sortedKeys(resps) {
			r := asMap(s.deref(resps[code], nil))
			fmt.Fprintf(&b, "### %s\n\n", code)
			if d := str(r["description"]); d != "" {
				b.WriteString(strings.TrimSpace(d) + "\n\n")
			}
			ct, media := firstMedia(asMap(r["content"]))
			if schema := media["schema"]; schema != nil {
				fmt.Fprintf(&b, "Content type: `%s`\n\n", ct)
				b.WriteString(s.schemaTable(schema))
				ex := firstNonNil(media["example"], exampleFromExamples(media["examples"]), s.example(schema, nil, 0))
				if j, err := json.MarshalIndent(ex, "", "  "); err == nil {
					fmt.Fprintf(&b, "```json title=\"Example response\"\n%s\n```\n\n", j)
				}
			}
		}
	}

	b.WriteString("## Example request\n\n")
	b.WriteString(codeSamples(o.method, server+o.path, hasBody, bodyExample))
	return b.String()
}

func codeSamples(method, url string, hasBody bool, example any) string {
	up := strings.ToUpper(method)
	body := ""
	if hasBody {
		if j, err := json.MarshalIndent(example, "", "  "); err == nil {
			body = string(j)
		}
	}
	var curl, js, gosrc strings.Builder
	fmt.Fprintf(&curl, "curl -X %s \"%s\"", up, url)
	if body != "" {
		fmt.Fprintf(&curl, " \\\n  -H \"Content-Type: application/json\" \\\n  -d '%s'", strings.ReplaceAll(body, "'", `'\''`))
	}
	fmt.Fprintf(&js, "const res = await fetch(%q, {\n  method: %q,\n", url, up)
	if body != "" {
		fmt.Fprintf(&js, "  headers: { 'Content-Type': 'application/json' },\n  body: JSON.stringify(%s),\n", indent(body, "  "))
	}
	js.WriteString("})\nconst data = await res.json()")
	gosrc.WriteString("package main\n\nimport (\n\t\"fmt\"\n\t\"io\"\n\t\"net/http\"\n")
	if body != "" {
		gosrc.WriteString("\t\"strings\"\n")
	}
	gosrc.WriteString(")\n\nfunc main() {\n")
	if body != "" {
		fmt.Fprintf(&gosrc, "\tbody := strings.NewReader(`%s`)\n\treq, _ := http.NewRequest(%q, %q, body)\n\treq.Header.Set(\"Content-Type\", \"application/json\")\n",
			strings.ReplaceAll(body, "`", "'"), up, url)
	} else {
		fmt.Fprintf(&gosrc, "\treq, _ := http.NewRequest(%q, %q, nil)\n", up, url)
	}
	gosrc.WriteString("\tres, err := http.DefaultClient.Do(req)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\tdefer res.Body.Close()\n\tb, _ := io.ReadAll(res.Body)\n\tfmt.Println(string(b))\n}")

	var b strings.Builder
	b.WriteString("<Tabs items=\"curl,JavaScript,Go\">\n")
	for _, t := range []struct{ name, lang, code string }{{"curl", "bash", curl.String()}, {"JavaScript", "js", js.String()}, {"Go", "go", gosrc.String()}} {
		fmt.Fprintf(&b, "<Tab value=\"%s\">\n\n```%s title=\"%s\"\n%s\n```\n\n</Tab>\n", t.name, t.lang, t.name, t.code)
	}
	b.WriteString("</Tabs>\n")
	return b.String()
}

// schemaTable renders an object's properties (flattened with dot paths) or,
// for a non-object schema, a single line with its type.
func (s *spec) schemaTable(schema any) string {
	type row struct {
		name, typ, desc string
		req             bool
	}
	var rows []row
	var walk func(sc any, path string, depth int, seen map[string]bool)
	walk = func(sc any, path string, depth int, seen map[string]bool) {
		m := asMap(s.deref(sc, seen))
		m = s.merge(m, seen)
		if items, ok := m["items"]; ok && typeOf(m) == "array" {
			m = s.merge(asMap(s.deref(items, seen)), seen)
			if path != "" {
				path += "[]"
			}
		}
		props := asMap(m["properties"])
		if len(props) == 0 || depth >= 4 {
			return
		}
		required := map[string]bool{}
		for _, r := range asSlice(m["required"]) {
			required[str(r)] = true
		}
		for _, k := range sortedKeys(props) {
			name := k
			if path != "" {
				name = path + "." + k
			}
			pm := asMap(s.deref(props[k], seen))
			desc := oneLine(str(pm["description"]))
			if enum := asSlice(pm["enum"]); len(enum) > 0 {
				var vs []string
				for _, e := range enum {
					vs = append(vs, fmt.Sprint(e))
				}
				desc = strings.TrimSpace(desc + " One of: " + strings.Join(vs, ", ") + ".")
			}
			rows = append(rows, row{name, s.typeName(props[k], seen), desc, required[k]})
			walk(props[k], name, depth+1, copySeen(seen))
		}
	}
	walk(schema, "", 0, map[string]bool{})
	if len(rows) == 0 {
		return fmt.Sprintf("Type: `%s`\n\n", s.typeName(schema, nil))
	}
	var b strings.Builder
	b.WriteString("| Field | Type | Required | Description |\n| --- | --- | --- | --- |\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| `%s` | `%s` | %s | %s |\n", cell(r.name), cell(r.typ), yesNo(r.req), cell(r.desc))
	}
	b.WriteString("\n")
	return b.String()
}

// typeName is a short human type: string(date-time), Pet[], object, ...
func (s *spec) typeName(sc any, seen map[string]bool) string {
	m := asMap(sc)
	if ref := str(m["$ref"]); ref != "" {
		name := ref[strings.LastIndex(ref, "/")+1:]
		resolved := asMap(s.deref(sc, copySeen(seen)))
		if t := typeOf(resolved); t != "" && t != "object" {
			return t
		}
		return name
	}
	for _, k := range []string{"oneOf", "anyOf"} {
		if alts := asSlice(m[k]); len(alts) > 0 {
			var names []string
			for _, a := range alts {
				names = append(names, s.typeName(a, seen))
			}
			return strings.Join(names, " | ")
		}
	}
	if alls := asSlice(m["allOf"]); len(alls) == 1 {
		return s.typeName(alls[0], seen)
	}
	t := typeOf(m)
	switch t {
	case "array":
		return s.typeName(m["items"], seen) + "[]"
	case "":
		if len(asMap(m["properties"])) > 0 || len(asSlice(m["allOf"])) > 0 {
			return "object"
		}
		return "any"
	}
	if f := str(m["format"]); f != "" {
		return t + "(" + f + ")"
	}
	return t
}

// example builds an example value: example, default, first enum, then a
// placeholder by type.
func (s *spec) example(sc any, seen map[string]bool, depth int) any {
	if seen == nil {
		seen = map[string]bool{}
	}
	m := s.merge(asMap(s.deref(sc, seen)), seen)
	if v, ok := m["example"]; ok {
		return v
	}
	if ex := asSlice(m["examples"]); len(ex) > 0 {
		return ex[0]
	}
	if v, ok := m["default"]; ok {
		return v
	}
	if e := asSlice(m["enum"]); len(e) > 0 {
		return e[0]
	}
	if depth > 5 {
		return nil
	}
	for _, k := range []string{"oneOf", "anyOf"} {
		if alts := asSlice(m[k]); len(alts) > 0 {
			return s.example(alts[0], copySeen(seen), depth+1)
		}
	}
	switch typeOf(m) {
	case "string":
		switch str(m["format"]) {
		case "date-time":
			return "2026-01-01T00:00:00Z"
		case "date":
			return "2026-01-01"
		case "email":
			return "user@example.com"
		case "uuid":
			return "3fa85f64-5717-4562-b3fc-2c963f66afa6"
		case "uri", "url":
			return "https://example.com"
		}
		return "string"
	case "integer":
		return 0
	case "number":
		return 0.0
	case "boolean":
		return true
	case "array":
		return []any{s.example(m["items"], copySeen(seen), depth+1)}
	}
	props := asMap(m["properties"])
	obj := map[string]any{}
	for _, k := range sortedKeys(props) {
		obj[k] = s.example(props[k], copySeen(seen), depth+1)
	}
	return obj
}

// merge folds allOf members into one schema.
func (s *spec) merge(m map[string]any, seen map[string]bool) map[string]any {
	alls := asSlice(m["allOf"])
	if len(alls) == 0 {
		return m
	}
	out := map[string]any{}
	for k, v := range m {
		if k != "allOf" {
			out[k] = v
		}
	}
	props := map[string]any{}
	for k, v := range asMap(m["properties"]) {
		props[k] = v
	}
	var req []any
	req = append(req, asSlice(m["required"])...)
	for _, a := range alls {
		am := s.merge(asMap(s.deref(a, seen)), seen)
		for k, v := range asMap(am["properties"]) {
			props[k] = v
		}
		req = append(req, asSlice(am["required"])...)
		if out["type"] == nil && am["type"] != nil {
			out["type"] = am["type"]
		}
		if out["description"] == nil && am["description"] != nil {
			out["description"] = am["description"]
		}
	}
	out["properties"] = props
	out["required"] = req
	return out
}

// deref follows local "$ref"s. seen guards against cycles: a ref already on
// the current path resolves to an empty schema.
func (s *spec) deref(v any, seen map[string]bool) any {
	for i := 0; i < 32; i++ {
		m, ok := v.(map[string]any)
		if !ok {
			return v
		}
		ref := str(m["$ref"])
		if ref == "" {
			return v
		}
		if seen != nil {
			if seen[ref] {
				return map[string]any{}
			}
			seen[ref] = true
		}
		if !strings.HasPrefix(ref, "#/") {
			return map[string]any{"description": "External reference " + ref}
		}
		var cur any = s.root
		for _, part := range strings.Split(ref[2:], "/") {
			part = strings.NewReplacer("~1", "/", "~0", "~").Replace(part)
			cur = asMap(cur)[part]
		}
		if cur == nil {
			return map[string]any{}
		}
		v = cur
	}
	return map[string]any{}
}

func root(s *spec, k string) any { return s.root[k] }

func firstMedia(content map[string]any) (string, map[string]any) {
	if m, ok := content["application/json"]; ok {
		return "application/json", asMap(m)
	}
	for _, k := range sortedKeys(content) {
		return k, asMap(content[k])
	}
	return "", nil
}

func exampleFromExamples(v any) any {
	m := asMap(v)
	for _, k := range sortedKeys(m) {
		if val, ok := asMap(m[k])["value"]; ok {
			return val
		}
	}
	return nil
}

func typeOf(m map[string]any) string {
	switch t := m["type"].(type) {
	case string:
		return t
	case []any: // 3.1: ["string", "null"]
		for _, x := range t {
			if str(x) != "null" {
				return str(x)
			}
		}
	}
	return ""
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	if m == nil {
		return map[string]any{}
	}
	return m
}

func asSlice(v any) []any { s, _ := v.([]any); return s }

func str(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	}
	return fmt.Sprint(v)
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func copySeen(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func firstNonNil(v ...any) any {
	for _, x := range v {
		if x != nil {
			return x
		}
	}
	return nil
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\n"); i >= 0 {
		s = s[:i]
	}
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

func cell(s string) string { return strings.ReplaceAll(oneLine(s), "|", `\|`) }

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func indent(s, pad string) string { return strings.ReplaceAll(s, "\n", "\n"+pad) }

func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

// normalize turns YAML's map[any]any (from unquoted keys like `200:`) into
// map[string]any all the way down.
func normalize(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, x := range t {
			t[k] = normalize(x)
		}
		return t
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, x := range t {
			m[fmt.Sprint(k)] = normalize(x)
		}
		return m
	case []any:
		for i, x := range t {
			t[i] = normalize(x)
		}
		return t
	}
	return v
}

// attr makes s safe inside a double-quoted component attribute.
func attr(s string) string { return strings.ReplaceAll(oneLine(s), `"`, "'") }
