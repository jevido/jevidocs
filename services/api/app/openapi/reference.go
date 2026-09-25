package openapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"dev.jevido/jevidocs/services/api/app/docs"
)

var jsonMarshal = json.Marshal

// Reference is the structured API reference the docs site renders. Every
// description is HTML rendered from the spec's Markdown. Schemas refer to
// named component schemas with Ref, resolved against Schemas, so shared and
// recursive schemas are stored once.
type Reference struct {
	OpenAPI         string             `json:"openapi"`
	Title           string             `json:"title"`
	Version         string             `json:"version"`
	Description     string             `json:"description"`
	Servers         []Server           `json:"servers"`
	SecuritySchemes []SecurityScheme   `json:"security_schemes"`
	Tags            []Tag              `json:"tags"`
	Models          []Model            `json:"models"`
	Schemas         map[string]*Schema `json:"schemas"`
}

type Server struct {
	URL         string           `json:"url"`
	Description string           `json:"description,omitempty"`
	Variables   []ServerVariable `json:"variables,omitempty"`
}

type ServerVariable struct {
	Name        string   `json:"name"`
	Default     string   `json:"default"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description,omitempty"`
}

// SecurityScheme is one entry of components.securitySchemes; Key is its
// name there, which operations list in Security.
type SecurityScheme struct {
	Key          string      `json:"key"`
	Type         string      `json:"type"`             // http, apiKey, oauth2, openIdConnect, mutualTLS
	Scheme       string      `json:"scheme,omitempty"` // http: bearer, basic, ...
	BearerFormat string      `json:"bearer_format,omitempty"`
	In           string      `json:"in,omitempty"` // apiKey: header, query, cookie
	Name         string      `json:"name,omitempty"`
	Description  string      `json:"description,omitempty"`
	OpenIDURL    string      `json:"openid_url,omitempty"`
	Flows        []OAuthFlow `json:"flows,omitempty"`
}

type OAuthFlow struct {
	Type             string  `json:"type"`
	AuthorizationURL string  `json:"authorization_url,omitempty"`
	TokenURL         string  `json:"token_url,omitempty"`
	Scopes           []Scope `json:"scopes,omitempty"`
}

type Scope struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Tag groups operations. ID is the element id on the page ("tag/pets").
type Tag struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Operations  []Operation `json:"operations"`

	raw string
}

// Operation is one method on one path. ID is its element id on the page,
// Scalar-style: "tag/pets/GET/pets/{petId}".
type Operation struct {
	ID          string       `json:"id"`
	Method      string       `json:"method"`
	Path        string       `json:"path"`
	OperationID string       `json:"operation_id,omitempty"`
	Summary     string       `json:"summary"`
	Description string       `json:"description,omitempty"`
	Deprecated  bool         `json:"deprecated,omitempty"`
	Security    []string     `json:"security"` // scheme keys that authorize it; empty = none
	Parameters  []Parameter  `json:"parameters"`
	RequestBody *RequestBody `json:"request_body,omitempty"`
	Responses   []Response   `json:"responses"`
	CodeSamples []CodeSample `json:"code_samples,omitempty"`

	src      map[string]any
	rawDesc  string
	rawItems []any
}

type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"` // path, query, header, cookie
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Deprecated  bool    `json:"deprecated,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
	Example     any     `json:"example,omitempty"`
}

type RequestBody struct {
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Content     []Media `json:"content"`
}

type Media struct {
	Type     string    `json:"type"`
	Schema   *Schema   `json:"schema,omitempty"`
	Examples []Example `json:"examples,omitempty"`
}

type Example struct {
	Name    string `json:"name"`
	Summary string `json:"summary,omitempty"`
	Value   any    `json:"value"`
}

type Response struct {
	Status      string      `json:"status"`
	Description string      `json:"description,omitempty"`
	Headers     []Parameter `json:"headers,omitempty"`
	Content     []Media     `json:"content,omitempty"`
}

// CodeSample is an author-written example from x-codeSamples.
type CodeSample struct {
	Lang   string `json:"lang"`
	Label  string `json:"label,omitempty"`
	Source string `json:"source"`
}

// Model is a named component schema, listed at the end of the reference.
type Model struct {
	ID   string `json:"id"` // "model/Pet"
	Name string `json:"name"`
}

// Schema is a JSON Schema reduced to what a reader needs. A schema with Ref
// set is a pointer to Reference.Schemas[Ref] and has no other fields.
type Schema struct {
	Ref                  string     `json:"ref,omitempty"`
	Type                 string     `json:"type,omitempty"`
	Format               string     `json:"format,omitempty"`
	Title                string     `json:"title,omitempty"`
	Description          string     `json:"description,omitempty"`
	Enum                 []any      `json:"enum,omitempty"`
	Default              any        `json:"default,omitempty"`
	Example              any        `json:"example,omitempty"`
	Nullable             bool       `json:"nullable,omitempty"`
	ReadOnly             bool       `json:"read_only,omitempty"`
	WriteOnly            bool       `json:"write_only,omitempty"`
	Deprecated           bool       `json:"deprecated,omitempty"`
	Constraints          []string   `json:"constraints,omitempty"`
	Properties           []Property `json:"properties,omitempty"`
	AdditionalProperties *Schema    `json:"additional_properties,omitempty"`
	Items                *Schema    `json:"items,omitempty"`
	Composition          string     `json:"composition,omitempty"` // oneOf or anyOf
	Variants             []*Schema  `json:"variants,omitempty"`
	// Truncated marks a schema cut off because it nested too deeply.
	Truncated bool `json:"truncated,omitempty"`
}

type Property struct {
	Name     string  `json:"name"`
	Required bool    `json:"required,omitempty"`
	Schema   *Schema `json:"schema"`
}

// Result is everything a page derives from one OpenAPI document.
type Result struct {
	Reference *Reference
	// Markdown is the whole reference as one Markdown document, for
	// llms.txt, page.md and MCP.
	Markdown string
	Sections []docs.Section
	Plain    string
	// Title and Summary default the page's title and description.
	Title   string
	Summary string
}

// maxDepth bounds inline schema nesting; named schemas restart the count.
const maxDepth = 24

// Build parses src (JSON or YAML) and derives the reference.
func Build(src []byte) (Result, error) {
	s, err := parse(src)
	if err != nil {
		return Result{}, err
	}
	b := &builder{s: s, schemas: map[string]*Schema{}, building: map[string]bool{}, rendered: map[string]docs.Rendered{}}
	ref := b.reference()
	info := asMap(s.root["info"])
	res := Result{Reference: ref, Title: ref.Title, Summary: firstNonEmpty(str(info["summary"]), firstLine(str(info["description"])))}
	res.Markdown = b.markdown(ref)
	res.Sections = b.sections(ref)
	var plain []string
	for _, sec := range res.Sections {
		plain = append(plain, sec.Title, sec.Text)
	}
	res.Plain = strings.TrimSpace(strings.Join(plain, "\n"))
	return res, nil
}

type builder struct {
	s        *spec
	schemas  map[string]*Schema
	building map[string]bool
	rendered map[string]docs.Rendered
}

// md renders a spec description (CommonMark) once per distinct text.
func (b *builder) md(src string) docs.Rendered {
	src = strings.TrimSpace(src)
	if src == "" {
		return docs.Rendered{}
	}
	if r, ok := b.rendered[src]; ok {
		return r
	}
	r, err := docs.Render(src)
	if err != nil {
		r = docs.Rendered{HTML: "<p>" + htmlEscape(src) + "</p>", Plain: src}
	}
	b.rendered[src] = r
	return r
}

func (b *builder) html(src string) string { return b.md(src).HTML }

func (b *builder) reference() *Reference {
	s := b.s
	info := asMap(s.root["info"])
	ref := &Reference{
		OpenAPI:     str(s.root["openapi"]),
		Title:       firstNonEmpty(str(info["title"]), "API reference"),
		Version:     str(info["version"]),
		Description: b.html(str(info["description"])),
		Servers:     []Server{},
		Tags:        []Tag{},
		Models:      []Model{},
		Schemas:     b.schemas,
	}
	for _, raw := range asSlice(s.root["servers"]) {
		m := asMap(raw)
		if str(m["url"]) == "" {
			continue
		}
		sv := Server{URL: str(m["url"]), Description: oneLine(str(m["description"]))}
		vars := asMap(m["variables"])
		for _, k := range s.keys(vars) {
			vm := asMap(vars[k])
			v := ServerVariable{Name: k, Default: str(vm["default"]), Description: oneLine(str(vm["description"]))}
			for _, e := range asSlice(vm["enum"]) {
				v.Enum = append(v.Enum, str(e))
			}
			sv.Variables = append(sv.Variables, v)
		}
		ref.Servers = append(ref.Servers, sv)
	}

	schemes := asMap(asMap(s.root["components"])["securitySchemes"])
	for _, k := range s.keys(schemes) {
		m := asMap(s.deref(schemes[k], nil))
		sc := SecurityScheme{Key: k, Type: str(m["type"]), Scheme: strings.ToLower(str(m["scheme"])), BearerFormat: str(m["bearerFormat"]),
			In: str(m["in"]), Name: str(m["name"]), Description: b.html(str(m["description"])), OpenIDURL: str(m["openIdConnectUrl"])}
		flows := asMap(m["flows"])
		for _, ft := range s.keys(flows) {
			fm := asMap(flows[ft])
			f := OAuthFlow{Type: ft, AuthorizationURL: str(fm["authorizationUrl"]), TokenURL: str(fm["tokenUrl"])}
			scopes := asMap(fm["scopes"])
			for _, name := range s.keys(scopes) {
				f.Scopes = append(f.Scopes, Scope{Name: name, Description: str(scopes[name])})
			}
			sc.Flows = append(sc.Flows, f)
		}
		ref.SecuritySchemes = append(ref.SecuritySchemes, sc)
	}
	if ref.SecuritySchemes == nil {
		ref.SecuritySchemes = []SecurityScheme{}
	}

	ref.Tags = b.tags()

	comps := asMap(asMap(s.root["components"])["schemas"])
	for _, name := range s.keys(comps) {
		b.component(name)
		ref.Models = append(ref.Models, Model{ID: "model/" + name, Name: name})
	}
	return ref
}

// tags groups operations by their first tag: declared tags first, in
// declaration order, then the others as they appear.
func (b *builder) tags() []Tag {
	s := b.s
	byName := map[string]*Tag{}
	var order []string
	declared := map[string]map[string]any{}
	for _, t := range asSlice(s.root["tags"]) {
		tm := asMap(t)
		if name := str(tm["name"]); name != "" && declared[name] == nil {
			declared[name] = tm
			order = append(order, name)
		}
	}
	tag := func(name string) *Tag {
		if t, ok := byName[name]; ok {
			return t
		}
		tm := declared[name]
		if tm == nil {
			order = append(order, name)
		}
		t := &Tag{Name: firstNonEmpty(str(asMap(tm)["x-displayName"]), name), raw: str(asMap(tm)["description"])}
		t.Description = b.html(t.raw)
		byName[name] = t
		return t
	}

	rootSecurity, hasRootSecurity := s.root["security"]
	paths := asMap(s.root["paths"])
	for _, p := range s.keys(paths) {
		item := asMap(s.deref(paths[p], nil))
		for _, m := range methods {
			raw, ok := item[m]
			if !ok {
				continue
			}
			om := asMap(raw)
			name := "default"
			if tags := asSlice(om["tags"]); len(tags) > 0 && str(tags[0]) != "" {
				name = str(tags[0])
			}
			t := tag(name)
			op := Operation{Method: m, Path: p, OperationID: str(om["operationId"]),
				Summary: firstNonEmpty(str(om["summary"]), strings.ToUpper(m)+" "+p), rawDesc: str(om["description"]), src: om, rawItems: asSlice(item["parameters"])}
			op.Description = b.html(op.rawDesc)
			op.Deprecated, _ = om["deprecated"].(bool)
			sec, has := om["security"]
			if !has && hasRootSecurity {
				sec = rootSecurity
			}
			op.Security = securityKeys(sec)
			op.Parameters = b.parameters(append(append([]any{}, asSlice(item["parameters"])...), asSlice(om["parameters"])...))
			if rb := asMap(s.deref(om["requestBody"], nil)); len(rb) > 0 {
				body := &RequestBody{Description: b.html(str(rb["description"])), Content: b.content(rb["content"])}
				body.Required, _ = rb["required"].(bool)
				op.RequestBody = body
			}
			resps := asMap(om["responses"])
			for _, code := range s.keys(resps) {
				rm := asMap(s.deref(resps[code], nil))
				r := Response{Status: code, Description: b.html(str(rm["description"])), Content: b.content(rm["content"])}
				hdrs := asMap(rm["headers"])
				for _, h := range s.keys(hdrs) {
					hm := asMap(s.deref(hdrs[h], nil))
					hm = withName(hm, h, "header")
					r.Headers = append(r.Headers, b.parameter(hm))
				}
				op.Responses = append(op.Responses, r)
			}
			if op.Responses == nil {
				op.Responses = []Response{}
			}
			for _, key := range []string{"x-codeSamples", "x-code-samples"} {
				for _, cs := range asSlice(om[key]) {
					cm := asMap(cs)
					if src := str(cm["source"]); src != "" {
						op.CodeSamples = append(op.CodeSamples, CodeSample{Lang: str(cm["lang"]), Label: str(cm["label"]), Source: src})
					}
				}
			}
			t.Operations = append(t.Operations, op)
		}
	}

	out := []Tag{}
	used := map[string]bool{}
	for _, name := range order {
		t := byName[name]
		if t == nil || len(t.Operations) == 0 {
			continue
		}
		base := "tag/" + firstNonEmpty(docs.Slugify(name), "default")
		t.ID = base
		for i := 2; used[t.ID]; i++ {
			t.ID = fmt.Sprintf("%s-%d", base, i)
		}
		used[t.ID] = true
		for i := range t.Operations {
			o := &t.Operations[i]
			o.ID = t.ID + "/" + strings.ToUpper(o.Method) + o.Path
		}
		out = append(out, *t)
	}
	return out
}

func withName(m map[string]any, name, in string) map[string]any {
	out := make(map[string]any, len(m)+2)
	for k, v := range m {
		out[k] = v
	}
	out["name"], out["in"] = name, in
	return out
}

func securityKeys(v any) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, req := range asSlice(v) {
		for k := range asMap(req) {
			if !seen[k] {
				seen[k] = true
				out = append(out, k)
			}
		}
	}
	return out
}

// parameters merges path-level and operation-level parameters; a later one
// with the same name and location replaces the earlier.
func (b *builder) parameters(raw []any) []Parameter {
	out := []Parameter{}
	idx := map[string]int{}
	for _, r := range raw {
		pm := asMap(b.s.deref(r, nil))
		if str(pm["name"]) == "" {
			continue
		}
		p := b.parameter(pm)
		key := p.In + ":" + p.Name
		if i, ok := idx[key]; ok {
			out[i] = p
		} else {
			idx[key] = len(out)
			out = append(out, p)
		}
	}
	return out
}

func (b *builder) parameter(pm map[string]any) Parameter {
	p := Parameter{Name: str(pm["name"]), In: str(pm["in"]), Description: b.html(str(pm["description"]))}
	p.Required, _ = pm["required"].(bool)
	p.Required = p.Required || p.In == "path"
	p.Deprecated, _ = pm["deprecated"].(bool)
	sc := pm["schema"]
	if sc == nil {
		if _, media := firstMedia(b.s, asMap(pm["content"])); media != nil {
			sc = media["schema"]
		}
	}
	if sc != nil {
		p.Schema = b.schema(sc, 0)
	}
	switch {
	case pm["example"] != nil:
		p.Example = pm["example"]
	case len(asMap(pm["examples"])) > 0:
		p.Example = exampleFromExamples(b.s, pm["examples"])
	case sc != nil:
		m := b.s.merge(asMap(b.s.deref(sc, map[string]bool{})), map[string]bool{})
		p.Example = firstNonNil(m["example"], first(asSlice(m["examples"])), m["default"], first(asSlice(m["enum"])))
	}
	p.Example = b.s.ordered(p.Example)
	return p
}

func (b *builder) content(v any) []Media {
	c := asMap(v)
	var out []Media
	for _, ct := range b.s.keys(c) {
		mm := asMap(c[ct])
		media := Media{Type: ct}
		if sc := mm["schema"]; sc != nil {
			media.Schema = b.schema(sc, 0)
		}
		if ex, ok := mm["example"]; ok {
			media.Examples = append(media.Examples, Example{Name: "Example", Value: b.s.ordered(ex)})
		}
		exs := asMap(mm["examples"])
		for _, name := range b.s.keys(exs) {
			em := asMap(b.s.deref(exs[name], nil))
			if val, ok := em["value"]; ok {
				media.Examples = append(media.Examples, Example{Name: name, Summary: oneLine(str(em["summary"])), Value: b.s.ordered(val)})
			}
		}
		if len(media.Examples) == 0 && mm["schema"] != nil {
			media.Examples = []Example{{Name: "Example", Value: b.s.example(mm["schema"], nil, 0)}}
		}
		out = append(out, media)
	}
	return out
}

// component builds a named schema once. A schema that refers to itself
// finds its name already building and keeps the reference.
func (b *builder) component(name string) {
	if _, ok := b.schemas[name]; ok || b.building[name] {
		return
	}
	b.building[name] = true
	raw := asMap(asMap(b.s.root["components"])["schemas"])[name]
	sc := b.schema(raw, 0)
	if sc.Ref == name { // a schema that is only a reference to itself
		sc = &Schema{Type: "any"}
	}
	b.schemas[name] = sc
	delete(b.building, name)
}

const componentPrefix = "#/components/schemas/"

func (b *builder) schema(v any, depth int) *Schema {
	m := asMap(v)
	if ref := str(m["$ref"]); ref != "" {
		if name, ok := strings.CutPrefix(ref, componentPrefix); ok && !strings.Contains(name, "/") {
			name = unescapePointer(name)
			if asMap(asMap(b.s.root["components"])["schemas"])[name] != nil {
				b.component(name)
				return &Schema{Ref: name}
			}
		}
		if depth >= maxDepth {
			return &Schema{Truncated: true}
		}
		return b.schema(b.s.deref(v, map[string]bool{}), depth+1)
	}
	if depth >= maxDepth {
		return &Schema{Truncated: true}
	}
	if alls := asSlice(m["allOf"]); len(alls) > 0 {
		// A lone allOf member (the usual way to add a description to a
		// $ref) keeps its reference.
		if len(alls) == 1 && len(asMap(m["properties"])) == 0 {
			inner := b.schema(alls[0], depth+1)
			if d := str(m["description"]); d != "" && inner.Ref != "" {
				return &Schema{Composition: "allOf", Variants: []*Schema{inner}, Description: b.html(d)}
			}
			return inner
		}
		m = b.s.merge(m, map[string]bool{})
	}
	out := &Schema{
		Type:        typeOf(m),
		Format:      str(m["format"]),
		Title:       str(m["title"]),
		Description: b.html(str(m["description"])),
		Enum:        asSlice(m["enum"]),
		Default:     b.s.ordered(m["default"]),
		Example:     b.s.ordered(firstNonNil(m["example"], first(asSlice(m["examples"])))),
		Nullable:    nullable(m),
		Constraints: constraints(m),
	}
	if c, ok := m["const"]; ok && out.Enum == nil {
		out.Enum = []any{c}
		if out.Type == "" {
			switch c.(type) {
			case string:
				out.Type = "string"
			case bool:
				out.Type = "boolean"
			case int, int64, float64:
				out.Type = "number"
			}
		}
	}
	out.ReadOnly, _ = m["readOnly"].(bool)
	out.WriteOnly, _ = m["writeOnly"].(bool)
	out.Deprecated, _ = m["deprecated"].(bool)
	for _, k := range []string{"oneOf", "anyOf"} {
		if alts := asSlice(m[k]); len(alts) > 0 {
			out.Composition = k
			for _, a := range alts {
				out.Variants = append(out.Variants, b.schema(a, depth+1))
			}
			break
		}
	}
	props := asMap(m["properties"])
	required := map[string]bool{}
	for _, r := range asSlice(m["required"]) {
		required[str(r)] = true
	}
	for _, k := range b.s.keys(props) {
		out.Properties = append(out.Properties, Property{Name: k, Required: required[k], Schema: b.schema(props[k], depth+1)})
	}
	if items, ok := m["items"]; ok && items != nil {
		out.Items = b.schema(items, depth+1)
		if out.Type == "" {
			out.Type = "array"
		}
	}
	switch ap := m["additionalProperties"].(type) {
	case map[string]any:
		out.AdditionalProperties = b.schema(ap, depth+1)
	case bool:
		if ap {
			out.AdditionalProperties = &Schema{}
		}
	}
	if out.Type == "" && (len(props) > 0 || out.AdditionalProperties != nil) {
		out.Type = "object"
	}
	return out
}

// constraints lists validation keywords as short phrases ("min: 1").
func constraints(m map[string]any) []string {
	var out []string
	num := func(k string) (string, bool) {
		v, ok := m[k]
		if !ok || v == nil {
			return "", false
		}
		if _, isBool := v.(bool); isBool {
			return "", false
		}
		return fmt.Sprint(v), true
	}
	exclusive := func(k string) bool { b, _ := m[k].(bool); return b } // 3.0 style
	if v, ok := num("minimum"); ok {
		if exclusive("exclusiveMinimum") {
			out = append(out, "> "+v)
		} else {
			out = append(out, "min: "+v)
		}
	}
	if v, ok := num("exclusiveMinimum"); ok {
		out = append(out, "> "+v)
	}
	if v, ok := num("maximum"); ok {
		if exclusive("exclusiveMaximum") {
			out = append(out, "< "+v)
		} else {
			out = append(out, "max: "+v)
		}
	}
	if v, ok := num("exclusiveMaximum"); ok {
		out = append(out, "< "+v)
	}
	for _, c := range []struct{ key, label string }{
		{"multipleOf", "multiple of"}, {"minLength", "min length"}, {"maxLength", "max length"},
		{"minItems", "min items"}, {"maxItems", "max items"}, {"minProperties", "min properties"}, {"maxProperties", "max properties"},
	} {
		if v, ok := num(c.key); ok {
			out = append(out, c.label+": "+v)
		}
	}
	if p := str(m["pattern"]); p != "" {
		out = append(out, "pattern: "+p)
	}
	if u, _ := m["uniqueItems"].(bool); u {
		out = append(out, "unique items")
	}
	return out
}

func firstMedia(s *spec, content map[string]any) (string, map[string]any) {
	if m, ok := content["application/json"]; ok {
		return "application/json", asMap(m)
	}
	for _, k := range s.keys(content) {
		return k, asMap(content[k])
	}
	return "", nil
}

func exampleFromExamples(s *spec, v any) any {
	m := asMap(v)
	for _, k := range s.keys(m) {
		if val, ok := asMap(s.deref(m[k], nil))["value"]; ok {
			return val
		}
	}
	return nil
}

func first(v []any) any {
	if len(v) == 0 {
		return nil
	}
	return v[0]
}

func firstNonNil(v ...any) any {
	for _, x := range v {
		if x != nil {
			return x
		}
	}
	return nil
}

func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

// sections are the search index: the introduction, each tag, operation and
// model, keyed by the element ids the site gives them.
func (b *builder) sections(ref *Reference) []docs.Section {
	info := asMap(b.s.root["info"])
	out := []docs.Section{{Text: b.md(str(info["description"])).Plain}}
	for _, t := range ref.Tags {
		out = append(out, docs.Section{ID: t.ID, Title: t.Name, Text: b.md(t.raw).Plain})
		for _, o := range t.Operations {
			text := []string{strings.ToUpper(o.Method) + " " + o.Path, b.md(o.rawDesc).Plain}
			for _, p := range o.Parameters {
				text = append(text, p.Name)
			}
			out = append(out, docs.Section{ID: o.ID, Title: o.Summary, Text: strings.TrimSpace(strings.Join(text, " "))})
		}
	}
	comps := asMap(asMap(b.s.root["components"])["schemas"])
	for _, m := range ref.Models {
		raw := asMap(comps[m.Name])
		text := []string{b.md(str(raw["description"])).Plain}
		for _, p := range b.s.keys(asMap(raw["properties"])) {
			text = append(text, p)
		}
		out = append(out, docs.Section{ID: m.ID, Title: m.Name, Text: strings.TrimSpace(strings.Join(text, " "))})
	}
	return out
}
