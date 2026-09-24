package docs

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/yuin/goldmark/util"
)

// Code blocks follow shiki's transformers, as fumadocs uses them:
//
//	```ts title="app.ts" {1,3-4} lineNumbers
//	const a = 1 // [!code highlight]
//	const b = 2 // [!code ++]
//	```
//
// Every line becomes <span class="line"> with data-highlighted, data-diff
// ("add" | "remove") or data-focus, separated by newlines so copying the
// code keeps its text intact.

type codeMeta struct {
	lang        string
	title       string
	highlight   map[int]bool
	lineNumbers bool
}

var (
	rangeRe    = regexp.MustCompile(`\{([\d,\s-]+)\}`)
	notationRe = regexp.MustCompile(`\s*(?://|#|--|;|<!--|/\*)\s*\[!code\s+(highlight|hl|\+\+|--|focus)\]\s*(?:-->|\*/)?\s*$`)
)

func parseCodeMeta(info string) codeMeta {
	m := codeMeta{highlight: map[int]bool{}}
	info = strings.TrimSpace(info)
	lang, _, _ := strings.Cut(info, " ")
	if i := strings.IndexByte(lang, '{'); i >= 0 {
		lang = lang[:i]
	}
	if strings.Contains(lang, "=") {
		lang = ""
	}
	m.lang = lang
	for _, a := range metaAttrRe.FindAllStringSubmatch(info, -1) {
		if a[1] == "title" {
			m.title = a[2] + a[3] + a[4]
		}
	}
	// Ranges only outside quoted values, e.g. not inside title="{x}".
	bare := metaAttrRe.ReplaceAllString(info, "")
	if r := rangeRe.FindStringSubmatch(bare); r != nil {
		for _, part := range strings.Split(r[1], ",") {
			part = strings.TrimSpace(part)
			from, to, isRange := strings.Cut(part, "-")
			a, err := strconv.Atoi(strings.TrimSpace(from))
			if err != nil {
				continue
			}
			b := a
			if isRange {
				if b, err = strconv.Atoi(strings.TrimSpace(to)); err != nil {
					continue
				}
			}
			for i := a; i <= b && i-a < 10000; i++ {
				m.highlight[i] = true
			}
		}
	}
	for _, f := range strings.Fields(bare) {
		if f == "lineNumbers" || f == "showLineNumbers" {
			m.lineNumbers = true
		}
	}
	return m
}

type lineFlags struct {
	highlight bool
	diff      string
	focus     bool
}

// stripNotations removes `// [!code ...]` comments and returns the flags
// per (1-based) line.
func stripNotations(code string) (string, map[int]lineFlags) {
	flags := map[int]lineFlags{}
	lines := strings.Split(code, "\n")
	for i, l := range lines {
		m := notationRe.FindStringSubmatchIndex(l)
		if m == nil {
			continue
		}
		f := flags[i+1]
		switch l[m[2]:m[3]] {
		case "highlight", "hl":
			f.highlight = true
		case "++":
			f.diff = "add"
		case "--":
			f.diff = "remove"
		case "focus":
			f.focus = true
		}
		flags[i+1] = f
		lines[i] = l[:m[0]]
	}
	return strings.Join(lines, "\n"), flags
}

// writeCode writes the highlighted lines of code.
func writeCode(w util.BufWriter, meta codeMeta, code string) {
	code = strings.TrimSuffix(code, "\n")
	flags := map[int]lineFlags{}
	if notations(meta.lang) {
		code, flags = stripNotations(code)
	}

	lang := meta.lang
	if alias, ok := langAliases[strings.ToLower(lang)]; ok {
		lang = alias
	}
	var lexer chroma.Lexer
	if lang != "" {
		lexer = lexers.Get(lang)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	var lines [][]chroma.Token
	if it, err := chroma.Coalesce(lexer).Tokenise(nil, code+"\n"); err == nil {
		lines = chroma.SplitTokensIntoLines(it.Tokens())
	} else {
		for _, l := range strings.Split(code, "\n") {
			lines = append(lines, []chroma.Token{{Type: chroma.Text, Value: l}})
		}
	}

	for i, toks := range lines {
		n := i + 1
		f := flags[n]
		if i > 0 {
			_ = w.WriteByte('\n')
		}
		_, _ = w.WriteString(`<span class="line"`)
		if f.highlight || meta.highlight[n] {
			_, _ = w.WriteString(` data-highlighted`)
		}
		if f.diff != "" {
			fmt.Fprintf(w, ` data-diff="%s"`, f.diff)
		}
		if f.focus {
			_, _ = w.WriteString(` data-focus`)
		}
		_ = w.WriteByte('>')
		for _, t := range toks {
			v := strings.TrimRight(t.Value, "\n")
			if v == "" {
				continue
			}
			cls := chroma.StandardTypes[t.Type]
			if cls == "" {
				_, _ = w.WriteString(html.EscapeString(v))
				continue
			}
			fmt.Fprintf(w, `<span class="%s">%s</span>`, cls, html.EscapeString(v))
		}
		_, _ = w.WriteString(`</span>`)
	}
}

// hasFocus reports whether any line is focused, so the figure can dim the
// others.
func hasFocus(lang, code string) bool {
	if !notations(lang) {
		return false
	}
	_, flags := stripNotations(code)
	for _, f := range flags {
		if f.focus {
			return true
		}
	}
	return false
}

// notations are off for Markdown fences, so docs can show `[!code ...]`
// examples without them being applied.
func notations(lang string) bool {
	switch strings.ToLower(lang) {
	case "md", "mdx", "markdown":
		return false
	}
	return true
}
