package docs

import (
	"strings"
	"testing"
)

func TestRenderComponents(t *testing.T) {
	src := "# Title\n\nIntro text.\n\n## Install it\n\n<Callout type=\"warn\" title=\"Careful\">\nThis is **bold**.\n</Callout>\n\n<Tabs items={['npm', 'bun']}>\n<Tab value=\"npm\">\n\n```sh title=\"terminal\"\nnpm i x\n```\n\n</Tab>\n<Tab value=\"bun\">bun add x</Tab>\n</Tabs>\n\n<Cards>\n<Card title=\"A\" href=\"/docs/a\" description=\"first\" />\n\n<Card title=\"B\" href=\"/docs/b\">second</Card>\n</Cards>\n\n### Deeper\n\n```go\nfunc main() {}\n```\n\n<Files>\n<Folder name=\"app\" defaultOpen>\n<File name=\"main.go\" />\n</Folder>\n</Files>\n"
	r, err := Render(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<div class="fd-callout" data-type="warn"><div class="fd-callout-title">Careful</div>`,
		`<strong>bold</strong>`,
		`<button type="button" role="tab" class="fd-tab-trigger" data-tab="npm" data-active>npm</button>`,
		`<div class="fd-tab" role="tabpanel" data-value="bun">`,
		`<figcaption class="fd-codeblock-title">terminal</figcaption>`,
		`<a class="fd-card" href="/docs/b"><div class="fd-card-title">B</div><div class="fd-card-desc">second</div></a>`,
		`<h2 id="install-it"><a class="fd-anchor" href="#install-it">Install it</a></h2>`,
		`<details class="fd-folder" open><summary>app</summary>`,
		`<div class="fd-file">main.go</div>`,
		`class="kd"`,
	} {
		if !strings.Contains(r.HTML, want) {
			t.Errorf("missing %q in\n%s", want, r.HTML)
		}
	}
	if strings.Contains(r.HTML, "<p><a class=\"fd-card\"") {
		t.Errorf("card wrapped in a paragraph:\n%s", r.HTML)
	}
	if len(r.Toc) != 2 || r.Toc[0].URL != "#install-it" || r.Toc[1].Depth != 3 {
		t.Errorf("toc = %+v", r.Toc)
	}
	if len(r.Sections) < 2 || r.Sections[1].ID != "install-it" {
		t.Errorf("sections = %+v", r.Sections)
	}
}

func TestComponentsInsideCodeAreLeftAlone(t *testing.T) {
	r, _ := Render("```mdx\n<Callout>hi</Callout>\n```\n")
	if strings.Contains(r.HTML, "fd-callout\"") {
		t.Errorf("expanded inside code fence: %s", r.HTML)
	}
}

func TestFrontMatter(t *testing.T) {
	fm, body := SplitFrontMatter("---\ntitle: \"Hello\"\ndescription: World\nposition: 3\n---\n\n# Body\n")
	if fm.Title != "Hello" || fm.Description != "World" || fm.Position != 3 || !fm.HasPosition {
		t.Errorf("fm = %+v", fm)
	}
	if body != "\n# Body\n" && body != "# Body\n" {
		t.Errorf("body = %q", body)
	}
}

func TestTree(t *testing.T) {
	tree := BuildTree("Docs", []PageMeta{
		{Slug: "", Title: "Introduction", Position: 0},
		{Slug: "guides/install", Title: "Install", Position: 1},
		{Slug: "guides", Title: "Guides", Position: 2, Section: "Learn"},
		{Slug: "guides/deploy", Title: "Deploy", Position: 2},
		{Slug: "quickstart", Title: "Quickstart", Position: 1},
		{Slug: "reference/api/rest", Title: "REST", Position: 5},
	})
	names := []string{}
	for _, n := range tree.Children {
		names = append(names, n.Type+":"+n.Name)
	}
	got := strings.Join(names, ",")
	want := "page:Introduction,page:Quickstart,separator:Learn,folder:Guides,folder:Reference"
	if got != want {
		t.Errorf("tree = %s, want %s", got, want)
	}
	flat := tree.Flatten()
	var slugs []string
	for _, l := range flat {
		slugs = append(slugs, l.Slug)
	}
	if strings.Join(slugs, ",") != ",quickstart,guides,guides/install,guides/deploy,reference/api/rest" {
		t.Errorf("flat = %v", slugs)
	}
	prev, next := tree.Neighbours("guides")
	if prev.Slug != "quickstart" || next.Slug != "guides/install" {
		t.Errorf("neighbours = %v %v", prev, next)
	}
	crumbs := tree.Breadcrumbs("reference/api/rest", "REST")
	if len(crumbs) != 3 || crumbs[1].Name != "Api" {
		t.Errorf("crumbs = %+v", crumbs)
	}
}

func TestNormalizeSlug(t *testing.T) {
	for in, want := range map[string]string{"/index": "", "guides/index.md": "guides", "/A/b/": "a/b"} {
		if got := NormalizeSlug(in); got != want {
			t.Errorf("NormalizeSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLongerFenceKeepsInnerFence(t *testing.T) {
	r, _ := Render("````mdx\n```js\nx\n```\n<Callout>hi</Callout>\n````\n")
	if strings.Contains(r.HTML, `class="fd-callout"`) {
		t.Errorf("expanded inside a four-backtick fence: %s", r.HTML)
	}
}

func TestSectionTextKeepsInlinePunctuation(t *testing.T) {
	r, _ := Render("## A\n\nUse `x`, then **y**.\n\n- one\n- two\n")
	if got := r.Sections[0].Text; got != "Use x, then y. one two" {
		t.Errorf("text = %q", got)
	}
}

func TestGitHubAlert(t *testing.T) {
	r, _ := Render("> [!WARNING]\n> Back up **first**.\n\n> plain quote\n")
	if !strings.Contains(r.HTML, `<div class="fd-callout" data-type="warn"><div class="fd-callout-title">Warning</div>`) ||
		!strings.Contains(r.HTML, "<strong>first</strong>") || strings.Contains(r.HTML, "[!WARNING]") {
		t.Errorf("alert html:\n%s", r.HTML)
	}
	if !strings.Contains(r.HTML, "<blockquote>") {
		t.Errorf("plain quote lost:\n%s", r.HTML)
	}
}

func TestCodeLineFeatures(t *testing.T) {
	r, _ := Render("```ts {2} lineNumbers\nconst a = 1 // [!code ++]\nconst b = 2\nconst c = 3 // [!code focus]\n```\n")
	for _, want := range []string{
		`data-line-numbers`, `data-has-focus`,
		`<span class="line" data-diff="add">`,
		`<span class="line" data-highlighted>`,
		`<span class="line" data-focus>`,
	} {
		if !strings.Contains(r.HTML, want) {
			t.Errorf("missing %q in\n%s", want, r.HTML)
		}
	}
	if strings.Contains(r.HTML, "[!code") {
		t.Errorf("notation not stripped:\n%s", r.HTML)
	}
}

func TestMermaidAndLazyImages(t *testing.T) {
	r, _ := Render("```mermaid\ngraph TD; A-->B\n```\n\n![alt text](/a.png \"T\")\n")
	if !strings.Contains(r.HTML, `<div class="fd-mermaid"><pre class="fd-mermaid-src">graph TD; A--&gt;B`) {
		t.Errorf("mermaid:\n%s", r.HTML)
	}
	if !strings.Contains(r.HTML, `<img src="/a.png" alt="alt text" title="T" loading="lazy"`) {
		t.Errorf("image:\n%s", r.HTML)
	}
}
