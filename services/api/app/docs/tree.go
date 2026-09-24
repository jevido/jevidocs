package docs

import (
	"sort"
	"strings"
)

// PageMeta is what the tree needs to know about a page.
type PageMeta struct {
	Slug        string
	Title       string
	Icon        string
	Position    int
	Section     string
	Description string
	Root        bool
}

// Node is one entry of a page tree, shaped like fumadocs' PageTree.
type Node struct {
	Type        string  `json:"type"` // page | folder | separator
	Name        string  `json:"name"`
	Slug        string  `json:"slug,omitempty"`
	Icon        string  `json:"icon,omitempty"`
	Index       *Node   `json:"index,omitempty"`
	Children    []*Node `json:"children,omitempty"`
	DefaultOpen bool    `json:"defaultOpen,omitempty"`
	// Root folders are shown as sidebar tabs, each with its own tree.
	Root        bool   `json:"root,omitempty"`
	Description string `json:"description,omitempty"`

	position int
	path     string
	section  string
}

// Tree is the root of a page tree.
type Tree struct {
	Name     string  `json:"name"`
	Children []*Node `json:"children"`
}

// Link is a page reference used for previous/next navigation.
type Link struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

// Crumb is one step of a breadcrumb trail.
type Crumb struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

// BuildTree arranges pages into folders by slug path. See the README's
// "Page tree rules".
func BuildTree(name string, pages []PageMeta) Tree {
	root := &Node{Type: "folder"}
	folders := map[string]*Node{"": root}
	var rootIndex *Node

	var folderFor func(path string) *Node
	folderFor = func(path string) *Node {
		if f, ok := folders[path]; ok {
			return f
		}
		parentPath, last := splitSlug(path)
		parent := folderFor(parentPath)
		f := &Node{Type: "folder", Name: humanize(last), path: path, position: 1 << 30}
		folders[path] = f
		parent.Children = append(parent.Children, f)
		return f
	}

	sorted := append([]PageMeta(nil), pages...)
	sort.SliceStable(sorted, func(i, j int) bool { return depth(sorted[i].Slug) < depth(sorted[j].Slug) })

	// Folder index pages first, so later children find their folder named.
	isFolder := map[string]bool{}
	for _, p := range sorted {
		parent, _ := splitSlug(p.Slug)
		for parent != "" {
			isFolder[parent] = true
			parent, _ = splitSlug(parent)
		}
	}

	for _, p := range sorted {
		page := &Node{Type: "page", Name: p.Title, Slug: p.Slug, Icon: p.Icon, position: p.Position, path: p.Slug, section: p.Section}
		if p.Slug == "" {
			rootIndex = page
			continue
		}
		if isFolder[p.Slug] {
			f := folderFor(p.Slug)
			f.Name, f.Icon, f.position, f.section = p.Title, p.Icon, p.Position, p.Section
			f.Root = p.Root && depth(p.Slug) == 1
			if f.Root {
				f.Description = p.Description
			}
			f.Index = &Node{Type: "page", Name: p.Title, Slug: p.Slug, Icon: p.Icon}
			continue
		}
		parentPath, _ := splitSlug(p.Slug)
		parent := folderFor(parentPath)
		parent.Children = append(parent.Children, page)
	}

	// Folders without an index sort by their earliest child.
	var fix func(n *Node) int
	fix = func(n *Node) int {
		min := n.position
		for _, c := range n.Children {
			pos := c.position
			if c.Type == "folder" {
				pos = fix(c)
			}
			if pos < min {
				min = pos
			}
		}
		if n.Index == nil && n.Type == "folder" {
			n.position = min
		}
		sortNodes(n.Children)
		return n.position
	}
	fix(root)

	children := root.Children
	if rootIndex != nil {
		children = append([]*Node{rootIndex}, children...)
	}
	return Tree{Name: name, Children: withSeparators(children)}
}

func sortNodes(nodes []*Node) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].position != nodes[j].position {
			return nodes[i].position < nodes[j].position
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})
}

func withSeparators(nodes []*Node) []*Node {
	var out []*Node
	current := ""
	for _, n := range nodes {
		if n.section != "" && n.section != current {
			out = append(out, &Node{Type: "separator", Name: n.section})
		}
		if n.section != "" {
			current = n.section
		}
		out = append(out, n)
	}
	return out
}

// Flatten lists pages in reading order: a folder's index, then its children.
func (t Tree) Flatten() []Link {
	var out []Link
	var walk func(nodes []*Node)
	walk = func(nodes []*Node) {
		for _, n := range nodes {
			switch n.Type {
			case "page":
				out = append(out, Link{Title: n.Name, Slug: n.Slug})
			case "folder":
				if n.Index != nil {
					out = append(out, Link{Title: n.Index.Name, Slug: n.Index.Slug})
				}
				walk(n.Children)
			}
		}
	}
	walk(t.Children)
	return out
}

// Neighbours returns the pages before and after slug in reading order.
func (t Tree) Neighbours(slug string) (prev, next *Link) {
	flat := t.Flatten()
	for i, l := range flat {
		if l.Slug == slug {
			if i > 0 {
				prev = &flat[i-1]
			}
			if i+1 < len(flat) {
				next = &flat[i+1]
			}
		}
	}
	return prev, next
}

// Breadcrumbs lists the folders leading to slug, then the page itself.
func (t Tree) Breadcrumbs(slug, title string) []Crumb {
	var trail []Crumb
	var find func(nodes []*Node, acc []Crumb) bool
	find = func(nodes []*Node, acc []Crumb) bool {
		for _, n := range nodes {
			if n.Type == "folder" {
				c := Crumb{Name: n.Name}
				if n.Index != nil {
					c.Slug = n.Index.Slug
				}
				if n.Index != nil && n.Index.Slug == slug {
					trail = append(acc, c)
					return true
				}
				if find(n.Children, append(append([]Crumb(nil), acc...), c)) {
					return true
				}
			}
			if n.Type == "page" && n.Slug == slug {
				trail = append(acc, Crumb{Name: n.Name, Slug: n.Slug})
				return true
			}
		}
		return false
	}
	if !find(t.Children, nil) {
		return []Crumb{{Name: title, Slug: slug}}
	}
	return trail
}

// MarkOpen sets DefaultOpen on the folders that contain slug.
func (t Tree) MarkOpen(slug string) {
	var walk func(nodes []*Node) bool
	walk = func(nodes []*Node) bool {
		found := false
		for _, n := range nodes {
			if n.Type == "page" && n.Slug == slug {
				found = true
			}
			if n.Type == "folder" {
				if (n.Index != nil && n.Index.Slug == slug) || walk(n.Children) {
					n.DefaultOpen = true
					found = true
				}
			}
		}
		return found
	}
	walk(t.Children)
}

func splitSlug(s string) (parent, last string) {
	i := strings.LastIndex(s, "/")
	if i < 0 {
		return "", s
	}
	return s[:i], s[i+1:]
}

func depth(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "/") + 1
}

func humanize(s string) string {
	s = strings.NewReplacer("-", " ", "_", " ").Replace(s)
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// NormalizeSlug trims slashes and a trailing "index" and lowercases.
func NormalizeSlug(s string) string {
	s = strings.Trim(strings.TrimSpace(s), "/")
	s = strings.TrimSuffix(s, ".md")
	s = strings.TrimSuffix(s, ".mdx")
	if s == "index" {
		return ""
	}
	s = strings.TrimSuffix(s, "/index")
	return strings.ToLower(s)
}
