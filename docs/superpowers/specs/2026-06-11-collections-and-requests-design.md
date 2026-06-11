# Collections & Requests — Design

**Date:** 2026-06-11
**Status:** Approved (design), pending implementation plan
**Project:** `gon` — interactive terminal HTTP client

## Summary

Add **collections** (folders) and **named request files** to `gon`, so that
HTTP requests can be saved as version-controllable `.yml` files and re-run by
name. The model is Bruno-inspired: a folder becomes a collection by containing a
`collection.yml` holding shared config; each other `.yml` file in the folder is a
single request. Collections nest, and shared config is inherited from the
workspace down through every collection to the request.

References: Postman (collections/requests), Bruno (one file per request,
folder-level config), OpenAPI (declarative, documentation-friendly schema).

## Goals

- Save requests as plain `.yml` files at the project root, committable as living
  API documentation.
- Group requests into nested collections with shared (inherited) config.
- Run a saved request by path: `run auth/login`.
- Scaffold collections and requests from the CLI.
- Reuse the existing workspace `Config` merge semantics and `HttpService` /
  `HttpOutput` — no new HTTP or merge machinery.

## Non-Goals (deferred)

- Environments and `{{var}}` substitution. The schema leaves room for `{{var}}`
  placeholders (values stay plain strings), but no substitution engine is built
  now.
- Tests / assertions / scripting (no mini test-runner).
- `multipart/form-data` bodies (only `json`, `raw`, `form`/urlencoded for now).

## File Layout & Discovery

Collections live at the **project root** (the directory containing `.gon/`).

```
my-project/
  .gon/
    workspace.yaml          # base_url + global config (existing)
  auth/
    collection.yml          # shared config for the auth folder
    login.yml               # a request
    register.yml
    admin/
      collection.yml        # shared config for auth/admin
      impersonate.yml
  users/
    collection.yml
    get-user.yml
```

Discovery rules:

- A folder is a **collection** iff it contains `collection.yml`. Folders without
  it (e.g. `internal/`, `cmd/`) are ignored.
- A **request** is any `*.yml` / `*.yaml` file in a collection folder **except**
  `collection.yml` (reserved; cannot be `run` as a request).
- `run auth/login` resolves `<root>/auth/login.yml`, where `<root>` is the
  directory holding `.gon/`. The `.yml` extension is implied (accept both with
  and without extension in the argument).
- Nesting is recursive; each folder level may carry its own `collection.yml`.

## Schemas

### `collection.yml`

Shared config only — structurally identical to the `config` block of
`workspace.yaml`, so the domain reuses `domain.Config`.

```yaml
name: Auth                    # optional, documentation
description: Authentication endpoints
config:
  path: /auth                 # path prefix, joined during URL resolution
  headers:
    X-Client: gon
  query:
    debug: "1"
```

All fields optional. An empty/absent `config` is a no-op.

### Request file (e.g. `login.yml`)

```yaml
name: Login                   # metadata; defaults to the file name when absent
description: Authenticate a user and return a token
method: POST
url: /login                   # relative -> resolved; absolute http(s):// -> bypass
headers:
  Accept: application/json
query:
  remember: "true"
body:
  json:                       # exactly one of: json | raw | form
    email: user@example.com
    password: secret
```

`body` supports exactly one of:

- `json:` — a YAML mapping serialized to JSON; auto `Content-Type: application/json`.
- `raw:` — a raw string; optional sibling `contentType:` sets the header.
- `form:` — a string→string map encoded as `application/x-www-form-urlencoded`.

Content-Type is auto-set per body type unless `headers` already provides one.
Specifying more than one body kind is a validation error.

## Precedence & Merge

When running `auth/admin/impersonate`, headers and query are merged so that the
**innermost / most specific wins** — consistent with the existing
`Config.ApplyDefaults`, which only adds a key when it is not already present.

```
CLI run flags  >  request file  >  collection.yml (admin)  >  collection.yml (auth)  >  workspace.yaml
```

Implementation reuses existing machinery rather than adding merge logic:

1. Build `HttpExecuteInput` from the request file (method, url, headers, query, body).
2. Apply CLI override flags onto that input (so they are already present, thus win).
3. Walk `collection.yml` files from the **nearest parent up to the root**,
   calling `ApplyDefaults` for each — inner before outer, so inner values win.
4. `HttpService.Execute` then applies the workspace defaults last and resolves
   the URL (existing behavior).

Because `ApplyDefaults` is additive (never overwrites an existing key), applying
inner→outer→workspace yields exactly the precedence above with no new code path.

### URL resolution

```
final = BaseURL + workspace.config.path + (collection.config.path, outer -> inner) + request.url
```

An absolute `http(s)://` URL in the request bypasses all prefixing (matching the
existing `Workspace.ResolveURL` rule). The collection path prefixes extend the
existing resolution; the workspace base/path remain the outermost segments.

## `run` Command

```
run <path>                       # run auth/login
run <path> --header "K: V"       # per-execution override (repeatable)
run <path> --query k=v           # per-execution override (repeatable)
run <path> --json '{...}'        # override request body
run <path> --minimal|--normal|--full
run <path> --timeout 10s
```

Rendering goes through the existing `HttpOutput`; `--full` echoes the fully
merged request (since merge mutates the shared input, the echo reflects all
inherited defaults). Output modes, flag parsing (`parseHeaders`/`parseQuery`),
and timeout handling mirror `HttpCommand`.

## Scaffolding Commands

```
collection init <name>           # create <name>/collection.yml (nested ok: auth/admin)
request new <path> --method POST # create <path>.yml skeleton
```

- `collection init` creates the folder and a `collection.yml` whose `name`
  defaults to the folder name converted to kebab-case (reusing the workspace
  name-derivation helper). Nested paths create intermediate `collection.yml`
  files as needed.
- `request new auth/login --method GET` writes `auth/login.yml` with a commented
  skeleton (method, url, headers, query, body). If the parent folder lacks a
  `collection.yml`, it is auto-created.
- Both refuse to overwrite an existing file (error, suggest editing instead).

## Hexagonal Mapping

| Layer | Addition |
|---|---|
| `core/domain` | `Request`, `RequestBody`, `Collection` (Collection = `Name`, `Description`, `Config`, reusing `domain.Config`). |
| `core/port/driving` | `RequestService` — `Run(ctx, path string, overrides *payload.HttpExecuteInput) (*payload.HttpExecuteOutput, error)`; plus scaffolding methods or a separate `CollectionService` for `init`/`new`. |
| `core/port/driven` | `RequestRepository` — load a request file and walk its ancestor `collection.yml` files from leaf to root; create files for scaffolding. |
| `core/service` | `RequestService` impl: load request → build input → apply overrides → apply collection + workspace defaults → delegate to `HttpService.Execute`. |
| `adapter/command` | `run_command.go`, `collection_init_command.go`, `request_new_command.go`. |
| `adapter/repository` | `request_repository.go` — read files, traverse collection ancestry, scaffold. |
| `adapter/model` | `request_model.go`, `collection_model.go` — YAML tags and domain↔model mapping (domain is never marshalled directly). Reuse `ConfigModel`. |

`RequestService` depends on the existing `HttpService` to execute. Exact wiring
(driving→driving vs. an output port) is decided in the implementation plan;
`cmd/main.go` wires the new commands into a new "Collections" command group.

## Domain Sketch

```go
// domain
type Request struct {
    Name        string
    Description string
    Method      string
    URL         string
    Headers     map[string][]string
    Query       map[string][]string
    Body        RequestBody
}

type BodyKind int // None | JSON | Raw | Form

type RequestBody struct {
    Kind        BodyKind
    JSON        any               // marshalled to JSON when Kind == JSON
    Raw         string
    ContentType string            // optional, for Raw
    Form        map[string]string
}

type Collection struct {
    Name        string
    Description string
    Config      Config // reuse existing domain.Config
}
```

## Validation & Errors

- Unknown request path / missing file → clear "request not found" error.
- `collection.yml` requested as a run target → reserved-name error.
- More than one body kind set → validation error naming the conflict.
- Malformed YAML → surfaced with the file path.
- Scaffolding into an existing file → refuse, suggest editing.

## Testing Strategy

- Domain: body-kind validation, Collection→Config merge ordering, URL prefix
  joining (table tests, mirroring existing `workspace_test.go`).
- Repository: temp-dir fixtures for load + ancestor traversal; scaffolding writes
  and refusal-to-overwrite.
- Service: precedence/merge across workspace + nested collections + request +
  CLI overrides, asserting the final `HttpExecuteInput` (reuse the assertion
  style in `workspace_service_test.go`).
- Command: `run` flag parsing/override and output-mode selection.
```
