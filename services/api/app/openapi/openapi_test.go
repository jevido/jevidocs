package openapi

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func build(t *testing.T, file string) Result {
	t.Helper()
	src, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Build(src)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestBuildPetstore(t *testing.T) {
	res := build(t, "testdata/petstore.yaml")
	ref := res.Reference
	if ref.Title != "Swagger Petstore" || ref.Version != "1.0.0" || res.Title != "Swagger Petstore" {
		t.Errorf("info: %q %q", ref.Title, ref.Version)
	}
	if len(ref.Servers) != 1 || ref.Servers[0].URL != "https://petstore.example.com/v1" {
		t.Errorf("servers: %+v", ref.Servers)
	}
	if len(ref.Tags) != 2 || ref.Tags[0].ID != "tag/pets" || ref.Tags[1].Name != "default" {
		t.Fatalf("tags: %+v", ref.Tags)
	}
	pets := ref.Tags[0].Operations
	// Document order, not alphabetical: GET /pets, POST /pets, GET /pets/{petId}.
	if len(pets) != 3 || pets[0].OperationID != "listPets" || pets[1].OperationID != "createPet" || pets[2].ID != "tag/pets/GET/pets/{petId}" {
		t.Fatalf("operations: %+v", pets)
	}
	show := pets[2]
	if len(show.Parameters) != 1 || show.Parameters[0].Name != "petId" || !show.Parameters[0].Required {
		t.Errorf("path parameter not merged: %+v", show.Parameters)
	}
	if got := show.Responses[0].Content[0].Schema; got == nil || got.Ref != "Pet" {
		t.Errorf("response schema should refer to Pet: %+v", got)
	}
	// Pet is allOf(NewPet, {...}) merged, with its recursive parent kept as a
	// reference, and properties in declaration order.
	pet := ref.Schemas["Pet"]
	if pet == nil || pet.Type != "object" {
		t.Fatalf("Pet: %+v", pet)
	}
	var names []string
	for _, p := range pet.Properties {
		names = append(names, p.Name)
		if p.Name == "parent" && p.Schema.Ref != "Pet" {
			t.Errorf("parent: %+v", p.Schema)
		}
		if p.Name == "id" && !p.Required {
			t.Error("id should be required")
		}
	}
	if strings.Join(names, ",") != "name,tag,owner,id,parent" {
		t.Errorf("Pet properties: %v", names)
	}
	create := pets[1]
	if create.RequestBody == nil || !create.RequestBody.Required || create.RequestBody.Content[0].Schema.Ref != "NewPet" {
		t.Fatalf("request body: %+v", create.RequestBody)
	}
	ex, _ := json.Marshal(create.RequestBody.Content[0].Examples[0].Value)
	if !strings.HasPrefix(string(ex), `{"name":"Rex","tag":"dog","owner":{"email":"user@example.com"}}`) {
		t.Errorf("example: %s", ex)
	}
	del := ref.Tags[1].Operations[0]
	if !del.Deprecated || del.ID != "tag/default/DELETE/pets/{petId}" {
		t.Errorf("delete: %+v", del)
	}
	if len(ref.Models) != 3 || ref.Models[0].ID != "model/NewPet" {
		t.Errorf("models: %+v", ref.Models)
	}

	// The whole reference serializes (recursive schemas are references).
	if _, err := json.Marshal(ref); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"## pets", "### Create a pet", "`POST /pets`", "| `owner.email` | `string(email)` | no |",
		`curl -X POST "https://petstore.example.com/v1/pets"`, "> **Deprecated.**"} {
		if !strings.Contains(res.Markdown, want) {
			t.Errorf("markdown missing %q:\n%s", want, res.Markdown)
		}
	}
	ids := map[string]string{}
	for _, s := range res.Sections {
		ids[s.ID] = s.Title
	}
	if ids["tag/pets/POST/pets"] != "Create a pet" || ids["model/Pet"] != "Pet" || ids["tag/pets"] != "pets" {
		t.Errorf("sections: %v", ids)
	}
	if !strings.Contains(res.Plain, "Everything about your pets") {
		t.Errorf("plain: %s", res.Plain)
	}
}

func TestSecurityAndConstraints(t *testing.T) {
	res, err := Build([]byte(`
openapi: 3.1.0
info: {title: T, version: "1"}
security: [{bearer: []}]
components:
  securitySchemes:
    bearer: {type: http, scheme: Bearer, bearerFormat: JWT}
    key: {type: apiKey, in: header, name: X-Key}
paths:
  /open:
    get:
      security: []
      responses: {"200": {description: ok}}
  /items:
    get:
      parameters:
        - {name: limit, in: query, schema: {type: integer, minimum: 1, maximum: 100, default: 20}}
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema: {type: [string, "null"], maxLength: 5}
`))
	if err != nil {
		t.Fatal(err)
	}
	ref := res.Reference
	if len(ref.SecuritySchemes) != 2 || ref.SecuritySchemes[0].Scheme != "bearer" {
		t.Errorf("schemes: %+v", ref.SecuritySchemes)
	}
	ops := ref.Tags[0].Operations
	if len(ops[0].Security) != 0 || strings.Join(ops[1].Security, ",") != "bearer" {
		t.Errorf("security: %v / %v", ops[0].Security, ops[1].Security)
	}
	limit := ops[1].Parameters[0]
	if limit.Example != 20 || strings.Join(limit.Schema.Constraints, ",") != "min: 1,max: 100" {
		t.Errorf("limit: %+v %+v", limit, limit.Schema)
	}
	sc := ops[1].Responses[0].Content[0].Schema
	if sc.Type != "string" || !sc.Nullable || sc.Constraints[0] != "max length: 5" {
		t.Errorf("nullable: %+v", sc)
	}
}

func TestRejectsSwagger2(t *testing.T) {
	if _, err := Build([]byte(`{"swagger":"2.0"}`)); err == nil {
		t.Error("expected error for Swagger 2.0")
	}
}

// The spec of our own public API stays importable.
func TestJevidocsSpec(t *testing.T) {
	res := build(t, "../../content/openapi/jevidocs.json")
	n := 0
	for _, tag := range res.Reference.Tags {
		n += len(tag.Operations)
	}
	if len(res.Reference.Tags) != 3 || n != 7 {
		t.Errorf("got %d tags, %d operations", len(res.Reference.Tags), n)
	}
}
