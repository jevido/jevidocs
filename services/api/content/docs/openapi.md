---
title: OpenAPI
description: Generate an API reference from an OpenAPI 3 document, like fumadocs-openapi.
position: 34
section: Reference
---

jevidocs can turn an **OpenAPI 3.0 or 3.1** document (JSON or YAML) into
documentation pages, the way `fumadocs-openapi` does. The generated pages are
ordinary Markdown pages: they show up in the sidebar, in search, in
`llms.txt`, and agents can read them over MCP.

## What gets generated

Under a prefix (default `api-reference`) the importer writes:

<Files>
<Folder name="api-reference" defaultOpen>
<File name="index: API title, version, description, servers, a card per tag" />
<Folder name="pets" defaultOpen>
<File name="index: the tag's description and a table of its operations" />
<File name="listpets: one page per operation" />
<File name="createpet" />
</Folder>
<Folder name="default">
<File name="operations without a tag" />
</Folder>
</Folder>
</Files>

Each operation page has:

- the method and path, as a coloured badge
- a **Deprecated** callout when the operation is deprecated
- the summary and description
- a parameters table (name, location, type, required, description), with
  path-level parameters merged in
- the request body as a property table, nested objects flattened with dot
  paths (`owner.email`), up to four levels deep
- every response with its description, schema table and an example body
- example requests in **curl**, **JavaScript** and **Go** tabs

Slugs come from `operationId`, or from the method and path when there is none.
Examples come from the schema's `example`, `examples`, `default` or first
`enum` value, and otherwise from placeholders for its type. Local `$ref`s
(`#/components/...`) are resolved, `allOf` is merged, and recursive schemas are
cut off where they repeat.

## Importing

<Tabs items="Admin,CLI,HTTP">
<Tab value="Admin">

Open a project in the [admin](/docs/admin), click **Import OpenAPI**, choose
a file or paste the document, and pick a prefix.

</Tab>
<Tab value="CLI">

```sh title="terminal"
JEVIDOCS_TOKEN=jd_... jevidocs openapi openapi.yaml -project my-api -prefix api-reference
```

</Tab>
<Tab value="HTTP">

```sh title="terminal"
curl -X POST https://api.jevidocs.jevido.app/api/admin/projects/my-api/openapi \
  -H "Authorization: Bearer $JEVIDOCS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$(jq -Rs '{spec: ., prefix: "api-reference"}' openapi.yaml)"
```

The answer counts what changed:

```json
{ "created": 7, "updated": 0, "unchanged": 0, "deleted": 0 }
```

</Tab>
</Tabs>

An example spec, for jevidocs' own public API, lives in the repository at
`services/api/content/openapi/jevidocs.json`.

## Re-importing

Import again whenever the spec changes. Pages under the prefix are created,
updated or left alone, and pages the spec no longer produces are deleted.
Pages **outside** the prefix are never touched, so hand-written guides can
live next to the generated reference.

<Callout type="warn" title="Edits are overwritten">
Changes made to generated pages in the admin are replaced on the next import.
Put extra prose in the spec's `description` fields, or in pages outside the
prefix.
</Callout>

## Limits

- Swagger 2.0 documents are rejected; convert them to OpenAPI 3 first.
- External `$ref`s (other files or URLs) are not fetched.
- The import runs inside the API's request timeout, so very large specs may
  need splitting by prefix.
