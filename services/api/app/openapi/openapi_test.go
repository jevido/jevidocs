package openapi

import (
	"os"
	"strings"
	"testing"

	"dev.jevido/jevidocs/services/api/app/docs"
)

func TestGeneratePetstore(t *testing.T) {
	src, err := os.ReadFile("testdata/petstore.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pages, err := Generate(src, "", "/p/petstore")
	if err != nil {
		t.Fatal(err)
	}
	bySlug := map[string]Page{}
	for _, p := range pages {
		bySlug[p.Slug] = p
	}
	for _, want := range []string{"api-reference", "api-reference/pets", "api-reference/pets/listpets",
		"api-reference/pets/createpet", "api-reference/pets/showpetbyid", "api-reference/default", "api-reference/default/delete-pets-petid"} {
		if _, ok := bySlug[want]; !ok {
			t.Errorf("missing page %s; have %v", want, keys(bySlug))
		}
	}
	idx := bySlug["api-reference"]
	if !strings.Contains(idx.Body, `href="/p/petstore/api-reference/pets"`) || !strings.Contains(idx.Body, "petstore.example.com") {
		t.Errorf("index body:\n%s", idx.Body)
	}
	create := bySlug["api-reference/pets/createpet"].Body
	for _, want := range []string{`data-method="post"`, "| `owner.email` | `string(email)` | no |", "| `name` | `string` | yes |",
		`"name": "Rex"`, `<Tabs items="curl,JavaScript,Go">`, "curl -X POST \"https://petstore.example.com/v1/pets\""} {
		if !strings.Contains(create, want) {
			t.Errorf("createPet missing %q:\n%s", want, create)
		}
	}
	show := bySlug["api-reference/pets/showpetbyid"].Body
	if !strings.Contains(show, "| `petId` | path | `string` | yes |") || !strings.Contains(show, "| `id` | `integer(int64)` | yes |") {
		t.Errorf("showPetById:\n%s", show)
	}
	list := bySlug["api-reference/pets/listpets"].Body
	if !strings.Contains(list, "### 200") || !strings.Contains(list, "`[].name`") && !strings.Contains(list, "`name`") {
		t.Errorf("listPets:\n%s", list)
	}
	if !strings.Contains(bySlug["api-reference/default/delete-pets-petid"].Body, `title="Deprecated"`) {
		t.Error("deprecated callout missing")
	}
	// Every page renders, and the components expand.
	for _, p := range pages {
		r, err := docs.Render(p.Body)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(r.HTML, "&lt;Tab") || strings.Contains(r.HTML, "<Tabs") || strings.Contains(r.HTML, "<Card ") {
			t.Errorf("%s: unexpanded component in\n%s", p.Slug, r.HTML)
		}
		if strings.Contains(p.Body, "fd-api-method") && !strings.Contains(r.HTML, `<span class="fd-api-method"`) {
			t.Errorf("%s: method badge lost:\n%s", p.Slug, r.HTML)
		}
	}
}

func TestRejectsSwagger2(t *testing.T) {
	if _, err := Generate([]byte(`{"swagger":"2.0"}`), "", "/docs"); err == nil {
		t.Error("expected error for Swagger 2.0")
	}
}

func keys(m map[string]Page) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

// The spec of our own public API stays importable.
func TestJevidocsSpec(t *testing.T) {
	src, err := os.ReadFile("../../content/openapi/jevidocs.json")
	if err != nil {
		t.Fatal(err)
	}
	pages, err := Generate(src, "api-reference", "/docs")
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1+3+7 {
		t.Errorf("got %d pages", len(pages))
	}
	for _, p := range pages {
		if _, err := docs.Render(p.Body); err != nil {
			t.Errorf("%s: %v", p.Slug, err)
		}
	}
}
