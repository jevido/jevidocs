package docs

import (
	"html"
	"regexp"
	"strings"
)

// Components are written JSX-style on their own lines, the way fumadocs
// authors write MDX:
//
//	<Callout type="warn" title="Heads up">
//	Body with **markdown**.
//	</Callout>
//
// There is no JSX runtime here. expandComponents rewrites every component
// line into the final HTML before Markdown parsing, surrounded by blank lines
// so CommonMark treats the tag as its own HTML block and still parses the
// Markdown between the tags. Every emitted line starts with a block-level tag
// (div, details, figure), which is what keeps that rule working.

var componentNames = "Callout|Cards|Card|Tabs|Tab|Steps|Step|Accordions|Accordion|Files|Folder|File"

var (
	// <Name attrs> or <Name attrs/> alone on a line.
	openTagRe = regexp.MustCompile(`^\s*<(` + componentNames + `)\b((?:[^>"']|"[^"]*"|'[^']*'|\{[^}]*\})*?)\s*(/?)>\s*$`)
	// </Name> alone on a line.
	closeTagRe = regexp.MustCompile(`^\s*</(` + componentNames + `)>\s*$`)
	// <Name attrs>inline</Name> on one line.
	inlineTagRe = regexp.MustCompile(`^\s*<(` + componentNames + `)\b((?:[^>"']|"[^"]*"|'[^']*'|\{[^}]*\})*?)>(.*)</(` + componentNames + `)>\s*$`)
	attrRe      = regexp.MustCompile(`([A-Za-z_][\w-]*)(?:=(?:"([^"]*)"|'([^']*)'|\{([^}]*)\}))?`)
	quotedRe    = regexp.MustCompile(`["']([^"']*)["']`)
	fenceRe     = regexp.MustCompile("^\\s*(```+|~~~+)")
)

type attrs map[string]string

func parseAttrs(s string) attrs {
	out := attrs{}
	for _, m := range attrRe.FindAllStringSubmatch(s, -1) {
		val := m[2] + m[3]
		if m[4] != "" {
			val = strings.TrimSpace(m[4])
		}
		if m[2] == "" && m[3] == "" && m[4] == "" && !strings.Contains(m[0], "=") {
			val = "true"
		}
		out[m[1]] = val
	}
	return out
}

func (a attrs) get(k string) string { return html.EscapeString(a[k]) }

// list reads `items="a,b"` or `items={['a', 'b']}`.
func (a attrs) list(k string) []string {
	v := a[k]
	if v == "" {
		return nil
	}
	if strings.ContainsAny(v, `'"`) {
		var out []string
		for _, m := range quotedRe.FindAllStringSubmatch(v, -1) {
			out = append(out, m[1])
		}
		return out
	}
	var out []string
	for _, p := range strings.Split(strings.Trim(v, "[]"), ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (a attrs) truthy(k string) bool {
	v, ok := a[k]
	return ok && v != "false"
}

type frame struct {
	name   string
	tabs   []string // Tabs: the tab values in order
	tabIdx int
}

// compact components hold only HTML; blank lines inside them would end the
// HTML block and turn the rest into paragraphs, so they are dropped.
func compact(name string) bool {
	return name == "Cards" || name == "Files" || name == "Folder"
}

func expandComponents(src string) string {
	lines := strings.Split(src, "\n")
	var out []string
	var stack []*frame
	fence := ""

	emit := func(s string, block bool) {
		inCompact := len(stack) > 0 && compact(stack[len(stack)-1].name)
		if block && !inCompact {
			out = append(out, "", s, "")
			return
		}
		out = append(out, s)
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Leave fenced code alone, including component examples inside it.
		if m := fenceRe.FindStringSubmatch(line); m != nil {
			marker := m[1]
			if fence == "" {
				fence = marker[:1]
			} else if strings.HasPrefix(strings.TrimSpace(line), strings.Repeat(fence, 3)) && strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), fence)) == "" {
				fence = ""
			}
			out = append(out, line)
			continue
		}
		if fence != "" {
			out = append(out, line)
			continue
		}

		if m := inlineTagRe.FindStringSubmatch(line); m != nil && m[1] == m[4] {
			name, a, inner := m[1], parseAttrs(m[2]), m[3]
			f := &frame{name: name}
			if name == "Tabs" && len(a.list("items")) == 0 {
				f.tabs = nil
			}
			parent := top(stack)
			if name == "Card" {
				if a["description"] == "" {
					a["description"] = inner
				}
				emit(renderCard(a), true)
				continue
			}
			if name == "File" {
				emit(`<div class="fd-file">`+html.EscapeString(firstNonEmpty(a["name"], inner))+`</div>`, true)
				continue
			}
			emit(openHTML(name, a, f, parent), true)
			out = append(out, inner)
			emit(closeHTML(name), true)
			continue
		}

		if m := openTagRe.FindStringSubmatch(line); m != nil {
			name, a, self := m[1], parseAttrs(m[2]), m[3] == "/"
			parent := top(stack)
			switch {
			case name == "Card":
				if self {
					emit(renderCard(a), true)
					continue
				}
				// <Card ...> body </Card>: the body is the description.
				var body []string
				for i+1 < len(lines) && !closeTagRe.MatchString(lines[i+1]) {
					i++
					body = append(body, strings.TrimSpace(lines[i]))
				}
				i++ // the closing tag
				if a["description"] == "" {
					a["description"] = strings.TrimSpace(strings.Join(body, " "))
				}
				emit(renderCard(a), true)
				continue
			case name == "File":
				emit(`<div class="fd-file">`+a.get("name")+`</div>`, true)
				continue
			case self:
				f := &frame{name: name}
				emit(openHTML(name, a, f, parent)+closeHTML(name), true)
				continue
			}
			f := &frame{name: name}
			if name == "Tabs" {
				f.tabs = a.list("items")
				if len(f.tabs) == 0 {
					f.tabs = lookaheadTabs(lines[i+1:])
				}
			}
			emit(openHTML(name, a, f, parent), true)
			stack = append(stack, f)
			continue
		}

		if m := closeTagRe.FindStringSubmatch(line); m != nil {
			name := m[1]
			// Pop to the matching frame; tolerate sloppy nesting.
			for j := len(stack) - 1; j >= 0; j-- {
				if stack[j].name == name {
					stack = stack[:j]
					break
				}
			}
			emit(closeHTML(name), true)
			continue
		}

		if strings.TrimSpace(line) == "" && len(stack) > 0 && compact(stack[len(stack)-1].name) {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func top(stack []*frame) *frame {
	if len(stack) == 0 {
		return nil
	}
	return stack[len(stack)-1]
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func lookaheadTabs(lines []string) []string {
	var tabs []string
	depth := 0
	for _, l := range lines {
		if m := openTagRe.FindStringSubmatch(l); m != nil && m[3] != "/" {
			if m[1] == "Tabs" {
				depth++
			}
			if m[1] == "Tab" && depth == 0 {
				tabs = append(tabs, parseAttrs(m[2])["value"])
			}
		} else if m := inlineTagRe.FindStringSubmatch(l); m != nil && m[1] == "Tab" && depth == 0 {
			tabs = append(tabs, parseAttrs(m[2])["value"])
		} else if m := closeTagRe.FindStringSubmatch(l); m != nil && m[1] == "Tabs" {
			if depth == 0 {
				break
			}
			depth--
		}
	}
	return tabs
}

var calloutTypes = map[string]string{
	"info": "info", "note": "info", "warn": "warn", "warning": "warn",
	"error": "error", "danger": "error", "success": "success", "idea": "idea", "tip": "idea",
}

func openHTML(name string, a attrs, f, parent *frame) string {
	switch name {
	case "Callout":
		t := calloutTypes[strings.ToLower(a["type"])]
		if t == "" {
			t = "info"
		}
		s := `<div class="fd-callout" data-type="` + t + `">`
		if a["title"] != "" {
			s += `<div class="fd-callout-title">` + a.get("title") + `</div>`
		}
		return s + `<div class="fd-callout-body">`
	case "Cards":
		return `<div class="fd-cards">`
	case "Tabs":
		var b strings.Builder
		b.WriteString(`<div class="fd-tabs"><div class="fd-tabs-list" role="tablist">`)
		for i, t := range f.tabs {
			b.WriteString(`<button type="button" role="tab" class="fd-tab-trigger" data-tab="` + html.EscapeString(t) + `"`)
			if i == 0 {
				b.WriteString(` data-active`)
			}
			b.WriteString(`>` + html.EscapeString(t) + `</button>`)
		}
		b.WriteString(`</div>`)
		return b.String()
	case "Tab":
		value := a["value"]
		active := false
		if parent != nil && parent.name == "Tabs" {
			if value == "" && parent.tabIdx < len(parent.tabs) {
				value = parent.tabs[parent.tabIdx]
			}
			active = parent.tabIdx == 0
			parent.tabIdx++
		}
		s := `<div class="fd-tab" role="tabpanel" data-value="` + html.EscapeString(value) + `"`
		if active {
			s += ` data-active`
		}
		return s + `>`
	case "Steps":
		return `<div class="fd-steps">`
	case "Step":
		return `<div class="fd-step">`
	case "Accordions":
		return `<div class="fd-accordions">`
	case "Accordion":
		s := `<details class="fd-accordion"`
		if a["id"] != "" {
			s += ` id="` + a.get("id") + `"`
		}
		return s + `><summary>` + a.get("title") + `</summary><div class="fd-accordion-body">`
	case "Files":
		return `<div class="fd-files">`
	case "Folder":
		s := `<details class="fd-folder"`
		if a.truthy("defaultOpen") {
			s += ` open`
		}
		return s + `><summary>` + a.get("name") + `</summary><div class="fd-folder-body">`
	}
	return `<div>`
}

func closeHTML(name string) string {
	switch name {
	case "Callout", "Accordion":
		if name == "Accordion" {
			return `</div></details>`
		}
		return `</div></div>`
	case "Folder":
		return `</div></details>`
	}
	return `</div>`
}

func renderCard(a attrs) string {
	tag, end := `<div class="fd-card">`, `</div>`
	if a["href"] != "" {
		tag, end = `<a class="fd-card" href="`+a.get("href")+`">`, `</a>`
	}
	s := tag + `<div class="fd-card-title">` + a.get("title") + `</div>`
	if a["description"] != "" {
		s += `<div class="fd-card-desc">` + a.get("description") + `</div>`
	}
	// Wrapped so the line starts with a block tag (see the package comment).
	return `<div class="fd-card-slot">` + s + end + `</div>`
}
