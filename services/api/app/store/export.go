package store

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// Export zips every page of p (all locales) as Markdown files with front
// matter, laid out like `jevidocs pull`, plus project.json with settings.
func Export(p models.Project) ([]byte, error) {
	var pages []models.Page
	if err := facades.Orm().Query().Where("project_id", p.ID).Order("slug asc").Get(&pages); err != nil {
		return nil, err
	}
	parents := map[string]bool{}
	for _, pg := range pages {
		for s := pg.Slug; strings.Contains(s, "/"); {
			s = s[:strings.LastIndex(s, "/")]
			parents[s] = true
		}
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	settings, _ := json.MarshalIndent(ViewProject(p), "", "  ")
	w, err := zw.Create("project.json")
	if err != nil {
		return nil, err
	}
	_, _ = w.Write(settings)
	for _, pg := range pages {
		name := pg.Slug
		switch {
		case pg.Slug == "":
			name = "index"
		case parents[pg.Slug]:
			name = pg.Slug + "/index"
		}
		if pg.Locale != "" {
			name += "." + pg.Locale
		}
		w, err := zw.Create(name + ".md")
		if err != nil {
			return nil, err
		}
		_, _ = w.Write([]byte(exportFile(pg)))
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func exportFile(pg models.Page) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", fmQuote(pg.Title))
	if pg.Description != "" {
		fmt.Fprintf(&b, "description: %s\n", fmQuote(pg.Description))
	}
	if pg.Icon != "" {
		fmt.Fprintf(&b, "icon: %s\n", pg.Icon)
	}
	if pg.Position != 0 {
		fmt.Fprintf(&b, "position: %d\n", pg.Position)
	}
	if pg.Section != "" {
		fmt.Fprintf(&b, "section: %s\n", fmQuote(pg.Section))
	}
	if pg.Root {
		b.WriteString("root: true\n")
	}
	if !pg.Published {
		b.WriteString("# unpublished\n")
	}
	b.WriteString("---\n\n")
	b.WriteString(strings.TrimLeft(pg.Body, "\n"))
	if !strings.HasSuffix(pg.Body, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

func fmQuote(s string) string {
	if strings.ContainsAny(s, ":#\"'") {
		return strconv.Quote(s)
	}
	return s
}
