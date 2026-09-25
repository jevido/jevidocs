// Package openapi turns an OpenAPI 3.0/3.1 document into a native API
// reference page, the way Scalar presents one: a structured Reference (tags,
// operations, parameters, schemas, examples, security schemes and servers)
// that the docs site lays out and makes interactive, plus Markdown and search
// sections so the page is searched, served in llms.txt and read over MCP like
// any other page.
package openapi

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultPrefix is where an imported reference goes when no prefix is given.
const DefaultPrefix = "api-reference"

var methods = []string{"get", "post", "put", "patch", "delete", "head", "options", "trace"}

// spec wraps the decoded document for $ref resolution. order remembers the
// key order of every decoded mapping, so tags, paths, properties and
// responses keep the order the author wrote them in.
type spec struct {
	root  map[string]any
	order map[uintptr][]string
}

// parse decodes src (JSON or YAML) and checks it is OpenAPI 3.x.
func parse(src []byte) (*spec, error) {
	var node yaml.Node
	// YAML is a superset of JSON, so one decoder handles both.
	if err := yaml.Unmarshal(src, &node); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	s := &spec{order: map[uintptr][]string{}}
	v, err := s.convert(&node, 0)
	if err != nil {
		return nil, err
	}
	root, _ := v.(map[string]any)
	if root == nil {
		return nil, errors.New("spec is empty")
	}
	if v, _ := root["openapi"].(string); !strings.HasPrefix(v, "3.") {
		return nil, errors.New("only OpenAPI 3.x documents are supported (missing or unsupported \"openapi\" version)")
	}
	s.root = root
	return s, nil
}

// convert turns a YAML node into plain maps, slices and scalars, recording
// each mapping's key order.
func (s *spec) convert(n *yaml.Node, depth int) (any, error) {
	if depth > 200 {
		return nil, errors.New("spec is nested too deeply")
	}
	switch n.Kind {
	case yaml.DocumentNode:
		if len(n.Content) == 0 {
			return nil, nil
		}
		return s.convert(n.Content[0], depth+1)
	case yaml.AliasNode:
		return s.convert(n.Alias, depth+1)
	case yaml.SequenceNode:
		out := make([]any, 0, len(n.Content))
		for _, c := range n.Content {
			v, err := s.convert(c, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case yaml.MappingNode:
		m := make(map[string]any, len(n.Content)/2)
		keys := make([]string, 0, len(n.Content)/2)
		for i := 0; i+1 < len(n.Content); i += 2 {
			k := n.Content[i].Value
			v, err := s.convert(n.Content[i+1], depth+1)
			if err != nil {
				return nil, err
			}
			if _, dup := m[k]; !dup {
				keys = append(keys, k)
			}
			m[k] = v
		}
		s.order[reflect.ValueOf(m).Pointer()] = keys
		return m, nil
	case yaml.ScalarNode:
		var v any
		if err := n.Decode(&v); err != nil {
			return nil, fmt.Errorf("parse spec: line %d: %w", n.Line, err)
		}
		return v, nil
	}
	return nil, nil
}

// keys lists m's keys in document order (sorted for maps built here).
func (s *spec) keys(m map[string]any) []string {
	if len(m) == 0 {
		return nil
	}
	// A map built and dropped here can free its address for another, so
	// the recorded order only counts when it names exactly m's keys.
	if k, ok := s.order[reflect.ValueOf(m).Pointer()]; ok && len(k) == len(m) {
		all := true
		for _, key := range k {
			if _, in := m[key]; !in {
				all = false
				break
			}
		}
		if all {
			return k
		}
	}
	return sortedKeys(m)
}

func (s *spec) remember(m map[string]any, keys []string) {
	s.order[reflect.ValueOf(m).Pointer()] = keys
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
	if v, ok := m["const"]; ok {
		return v
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
		case "binary":
			return "@file"
		}
		return "string"
	case "integer":
		return 1
	case "number":
		return 1.5
	case "boolean":
		return true
	case "array":
		return []any{s.example(m["items"], copySeen(seen), depth+1)}
	}
	props := asMap(m["properties"])
	obj := orderedObject{}
	for _, k := range s.keys(props) {
		obj = append(obj, field{k, s.example(props[k], copySeen(seen), depth+1)})
	}
	return obj
}

// merge folds allOf members into one schema, keeping property order.
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
	var order []string
	add := func(pm map[string]any) {
		for _, k := range s.keys(pm) {
			if _, ok := props[k]; !ok {
				order = append(order, k)
			}
			props[k] = pm[k]
		}
	}
	var req []any
	for _, a := range alls {
		am := s.merge(asMap(s.deref(a, seen)), seen)
		add(asMap(am["properties"]))
		req = append(req, asSlice(am["required"])...)
		for _, k := range []string{"type", "description", "title", "additionalProperties"} {
			if out[k] == nil && am[k] != nil {
				out[k] = am[k]
			}
		}
	}
	add(asMap(m["properties"]))
	req = append(req, asSlice(m["required"])...)
	s.remember(props, order)
	out["properties"] = props
	out["required"] = req
	if out["type"] == nil && len(props) > 0 {
		out["type"] = "object"
	}
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
			cur = asMap(cur)[unescapePointer(part)]
		}
		if cur == nil {
			return map[string]any{}
		}
		v = cur
	}
	return map[string]any{}
}

func unescapePointer(s string) string { return strings.NewReplacer("~1", "/", "~0", "~").Replace(s) }

// orderedObject is a JSON object that keeps its field order, so generated
// examples list properties the way the schema does.
type orderedObject []field

type field struct {
	k string
	v any
}

func (o orderedObject) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i, f := range o {
		if i > 0 {
			b.WriteByte(',')
		}
		k, err := jsonMarshal(f.k)
		if err != nil {
			return nil, err
		}
		v, err := jsonMarshal(f.v)
		if err != nil {
			return nil, err
		}
		b.Write(k)
		b.WriteByte(':')
		b.Write(v)
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}

// ordered copies a value from the spec with its objects as orderedObjects,
// so examples serialize in the order they were written.
func (s *spec) ordered(v any) any {
	switch t := v.(type) {
	case map[string]any:
		o := make(orderedObject, 0, len(t))
		for _, k := range s.keys(t) {
			o = append(o, field{k, s.ordered(t[k])})
		}
		return o
	case []any:
		out := make([]any, len(t))
		for i, x := range t {
			out[i] = s.ordered(x)
		}
		return out
	}
	return v
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

func nullable(m map[string]any) bool {
	if b, _ := m["nullable"].(bool); b {
		return true
	}
	for _, x := range asSlice(m["type"]) {
		if str(x) == "null" {
			return true
		}
	}
	return false
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
