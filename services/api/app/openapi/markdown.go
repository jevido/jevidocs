package openapi

import (
	"encoding/json"
	"fmt"
	"strings"
)

// markdown writes the reference as one Markdown document for agents and
// llms.txt: descriptions stay the author's Markdown, schemas become tables.
func (b *builder) markdown(ref *Reference) string {
	s := b.s
	info := asMap(s.root["info"])
	var w strings.Builder
	if ref.Version != "" {
		fmt.Fprintf(&w, "Version `%s` (OpenAPI %s)\n\n", ref.Version, ref.OpenAPI)
	}
	if d := strings.TrimSpace(str(info["description"])); d != "" {
		w.WriteString(d + "\n\n")
	}
	if len(ref.Servers) > 0 {
		w.WriteString("## Servers\n\n")
		for _, sv := range ref.Servers {
			line := "- `" + sv.URL + "`"
			if sv.Description != "" {
				line += ": " + sv.Description
			}
			w.WriteString(line + "\n")
		}
		w.WriteString("\n")
	}
	if len(ref.SecuritySchemes) > 0 {
		w.WriteString("## Authentication\n\n")
		for _, sc := range ref.SecuritySchemes {
			fmt.Fprintf(&w, "- `%s`: %s\n", sc.Key, schemeText(sc))
		}
		w.WriteString("\n")
	}
	server := ""
	if len(ref.Servers) > 0 {
		server = strings.TrimRight(ref.Servers[0].URL, "/")
	}
	for _, t := range ref.Tags {
		fmt.Fprintf(&w, "## %s\n\n", t.Name)
		if d := strings.TrimSpace(t.raw); d != "" {
			w.WriteString(d + "\n\n")
		}
		for _, o := range t.Operations {
			b.operationMarkdown(&w, o, server)
		}
	}
	return strings.TrimSpace(w.String()) + "\n"
}

func schemeText(sc SecurityScheme) string {
	switch sc.Type {
	case "http":
		if sc.Scheme == "bearer" {
			return "HTTP bearer token (`Authorization: Bearer <token>`)"
		}
		return "HTTP " + sc.Scheme + " authentication"
	case "apiKey":
		return fmt.Sprintf("API key in the %s `%s`", sc.In, sc.Name)
	case "oauth2":
		return "OAuth 2"
	case "openIdConnect":
		return "OpenID Connect"
	}
	return sc.Type
}

func (b *builder) operationMarkdown(w *strings.Builder, o Operation, server string) {
	fmt.Fprintf(w, "### %s\n\n`%s %s`\n\n", o.Summary, strings.ToUpper(o.Method), o.Path)
	if o.Deprecated {
		w.WriteString("> **Deprecated.** This operation may be removed.\n\n")
	}
	if d := strings.TrimSpace(o.rawDesc); d != "" {
		w.WriteString(d + "\n\n")
	}
	if len(o.Security) > 0 {
		fmt.Fprintf(w, "Authentication: %s\n\n", "`"+strings.Join(o.Security, "` or `")+"`")
	}
	if len(o.Parameters) > 0 {
		w.WriteString("**Parameters**\n\n| Name | In | Type | Required | Description |\n| --- | --- | --- | --- | --- |\n")
		all := append(append([]any{}, o.rawItems...), asSlice(o.src["parameters"])...)
		types := map[string]string{}
		for _, raw := range all {
			pm := asMap(b.s.deref(raw, nil))
			types[str(pm["in"])+":"+str(pm["name"])] = b.s.typeName(pm["schema"], nil)
		}
		for _, p := range o.Parameters {
			fmt.Fprintf(w, "| `%s` | %s | `%s` | %s | %s |\n", cell(p.Name), p.In, cell(types[p.In+":"+p.Name]), yesNo(p.Required), cell(b.md(rawParamDesc(b.s, all, p)).Plain))
		}
		w.WriteString("\n")
	}
	var bodyExample any
	hasBody := false
	if rb := asMap(b.s.deref(o.src["requestBody"], nil)); len(rb) > 0 {
		ct, media := firstMedia(b.s, asMap(rb["content"]))
		w.WriteString("**Request body**")
		if ct != "" {
			fmt.Fprintf(w, " (`%s`", ct)
			if req, _ := rb["required"].(bool); req {
				w.WriteString(", required")
			}
			w.WriteString(")")
		}
		w.WriteString("\n\n")
		if d := strings.TrimSpace(str(rb["description"])); d != "" {
			w.WriteString(d + "\n\n")
		}
		if schema := media["schema"]; schema != nil {
			w.WriteString(b.schemaTable(schema))
			hasBody = true
			bodyExample = firstNonNil(media["example"], exampleFromExamples(b.s, media["examples"]), b.s.example(schema, nil, 0))
		}
	}
	if resps := asMap(o.src["responses"]); len(resps) > 0 {
		w.WriteString("**Responses**\n\n")
		for _, code := range b.s.keys(resps) {
			r := asMap(b.s.deref(resps[code], nil))
			fmt.Fprintf(w, "- `%s`: %s\n", code, oneLine(str(r["description"])))
		}
		w.WriteString("\n")
		for _, code := range b.s.keys(resps) {
			r := asMap(b.s.deref(resps[code], nil))
			ct, media := firstMedia(b.s, asMap(r["content"]))
			if schema := media["schema"]; schema != nil {
				fmt.Fprintf(w, "Response `%s` (`%s`):\n\n", code, ct)
				w.WriteString(b.schemaTable(schema))
				ex := firstNonNil(media["example"], exampleFromExamples(b.s, media["examples"]), b.s.example(schema, nil, 0))
				if j, err := json.MarshalIndent(ex, "", "  "); err == nil {
					fmt.Fprintf(w, "```json\n%s\n```\n\n", j)
				}
			}
		}
	}
	w.WriteString("Example request:\n\n```bash\n" + curl(o.Method, server+o.Path, hasBody, bodyExample) + "\n```\n\n")
}

func rawParamDesc(s *spec, all []any, p Parameter) string {
	desc := ""
	for _, raw := range all {
		pm := asMap(s.deref(raw, nil))
		if str(pm["name"]) == p.Name && str(pm["in"]) == p.In {
			desc = str(pm["description"])
		}
	}
	return desc
}

func curl(method, url string, hasBody bool, example any) string {
	c := fmt.Sprintf("curl -X %s \"%s\"", strings.ToUpper(method), url)
	if hasBody {
		if j, err := json.MarshalIndent(example, "", "  "); err == nil {
			c += fmt.Sprintf(" \\\n  -H \"Content-Type: application/json\" \\\n  -d '%s'", strings.ReplaceAll(string(j), "'", `'\''`))
		}
	}
	return c
}

// schemaTable renders an object's properties (flattened with dot paths) or,
// for a non-object schema, a single line with its type.
func (b *builder) schemaTable(schema any) string {
	s := b.s
	type row struct {
		name, typ, desc string
		req             bool
	}
	var rows []row
	var walk func(sc any, path string, depth int, seen map[string]bool)
	walk = func(sc any, path string, depth int, seen map[string]bool) {
		m := s.merge(asMap(s.deref(sc, seen)), seen)
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
		for _, k := range s.keys(props) {
			name := k
			if path != "" {
				name = path + "." + k
			}
			pm := asMap(s.deref(props[k], copySeen(seen)))
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
	var w strings.Builder
	w.WriteString("| Field | Type | Required | Description |\n| --- | --- | --- | --- |\n")
	for _, r := range rows {
		fmt.Fprintf(&w, "| `%s` | `%s` | %s | %s |\n", cell(r.name), cell(r.typ), yesNo(r.req), cell(r.desc))
	}
	w.WriteString("\n")
	return w.String()
}

func cell(s string) string { return strings.ReplaceAll(oneLine(s), "|", `\|`) }

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
