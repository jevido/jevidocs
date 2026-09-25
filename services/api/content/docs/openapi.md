---
title: OpenAPI
description: A native, interactive API reference from an OpenAPI 3 document, in the style of Scalar.
position: 34
section: Reference
---

jevidocs turns an **OpenAPI 3.0 or 3.1** document (JSON or YAML) into an
API reference page laid out like [Scalar](https://scalar.com): the whole API
on one page, every operation in two columns with its documentation on the
left and request and response samples on the right, and a request client to
try it. See it on this site's own [API reference](/docs/api/reference).

The reference is still a page. It sits in the sidebar, shows up in search (per
operation), and agents get it as Markdown in `llms.txt`, from `page.md` and
over MCP.

## What the page shows

- **Introduction:** title, version, the `OAS` version and the description.
  Next to it, pickers for the **server** (with server variables), the
  **authentication** scheme and credentials, and the **client** language for
  code samples.
- **Tags** as sections, in the order the spec declares them. Operations
  without a tag go under `default`.
- **Operations**, each with:
  - the summary, method and path, and a **Deprecated** badge when it is
  - the description (CommonMark, rendered like any page)
  - path, query, header and cookie **parameters**, with types, `required`,
    constraints (`min`, `max length`, `pattern`, ...), allowed values,
    defaults and examples
  - the **request body** per content type, as nested attributes: objects fold
    open with *Show child attributes*, `oneOf`/`anyOf` get a switcher,
    `allOf` is merged
  - the **responses** with their descriptions, headers and schemas
  - a **request sample** in cURL, JavaScript, Python, Go, PHP or raw HTTP,
    built from the chosen server and credentials, plus the spec's own
    `x-codeSamples`; named request examples get a picker
  - a **response sample** per status code, from the spec's `example` or
    `examples`, or generated from the schema
  - **Test Request**, which opens a client to edit parameters, headers and
    the body, send the request from the browser and read the status, timing,
    headers and body (`Ctrl`/`⌘` + `Enter` sends)
- **Models:** every schema in `components.schemas`, collapsible.

In the sidebar the page lists its tags, operations (with their methods) and
models, and follows the one you are reading.

<Callout type="info" title="Credentials stay in the tab">
Tokens and keys entered under Authentication live in the browser tab's
session storage only. They are never sent to jevidocs, only to the API
you test. Test Request is a plain browser `fetch`, so the API must allow
requests from the docs site's origin (CORS).
</Callout>

## Importing

Importing makes the page at a prefix (default `api-reference`) the reference.
The page's title and description default to the spec's `info.title` and
`info.summary`.

<Tabs items="Admin,CLI,HTTP,Files">
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
{ "created": 1, "updated": 0, "unchanged": 0, "deleted": 0 }
```

</Tab>
<Tab value="Files">

In a folder you sync (`jevidocs push`, a GitHub source), a file named
`<slug>.openapi.yaml`, `.openapi.yml` or `.openapi.json` is an OpenAPI page:

<Files>
<Folder name="docs" defaultOpen>
<File name="index.md" />
<File name="reference.openapi.yaml" />
</Folder>
</Files>

`jevidocs pull` and the admin's export write OpenAPI pages back the same way.

</Tab>
</Tabs>

An example spec, for jevidocs' own public API, lives in the repository at
`services/api/content/openapi/jevidocs.json`.

## Re-importing

Import again whenever the spec changes: the page is updated in place and keeps
the title, description and position you gave it in the admin. Pages
**below** the prefix are deleted (earlier versions of jevidocs generated a
page per operation there). Pages elsewhere are never touched, so
hand-written guides can live next to the reference.

In the admin, the page's body is the OpenAPI document itself; editing it
there works too, and its preview shows the Markdown agents receive.

## Linking to an operation

Each tag, operation and model has an anchor, Scalar-style:

| What | Anchor |
| ---- | ------ |
| Tag | `#tag/pets` |
| Operation | `#tag/pets/GET/pets/{petId}` |
| Model | `#model/Pet` |

## Limits

- Swagger 2.0 documents are rejected; convert them to OpenAPI 3 first.
- External `$ref`s (other files or URLs) are not fetched.
- Webhooks and callbacks are not shown yet.
- OAuth 2 and OpenID Connect schemes take a ready access token; there is no
  authorization flow in the page.
