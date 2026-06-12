# Design: Environments (Postman-style, project-scoped)

**Date:** 2026-06-12
**Status:** Approved (design)

## Goal

Add Postman-style environments (e.g. `local`, `dev`, `test`, `prod`) to a `gon`
workspace. An environment is a named set of variables plus a base URL. Requests
reference variables via `{{var}}` substitution. Environments are scoped to the
project (the workspace folder), so they behave like Postman "collection
variables" at the project level. Switching the active environment changes the
base URL and variable values used for every request.

## Decisions (settled during brainstorming)

- **`base_url` becomes a per-environment value.** Each environment owns its own
  `base_url`. The workspace no longer treats `base_url` as the source of truth.
- **Workspace `config` (default headers/query/path) stays shared.** Its values
  may contain `{{var}}`, resolved from the active environment (e.g.
  `Authorization: Bearer {{token}}`, where `token` differs per env). This keeps
  defaults DRY.
- **Environments are stored as separate files** under `environments/<name>.yml`,
  consistent with the existing collections/requests layout — the whole folder
  stays a self-contained, shareable artifact.
- **Active environment selection precedence:** `--env` flag (per command) →
  locally persisted choice (`.gon/active-env`, gitignored) → fallback rules.
  There is intentionally **no** `default_environment` committed to
  `workspace.yml`; the active choice is per-developer.
- **Resolution happens at the command layer**, and the resolved
  `*domain.Environment` is passed into the service (hexagonal: command builds
  input + context, service executes).
- **(A) Undefined variables fail fast.** If any `{{var}}` remains unresolved
  after substitution, the request fails with a message naming the missing
  variables — preventing a raw `{{token}}` from being sent to prod.
- **(B) `gon init` scaffolds `environments/local.yml`** (with the previous
  default `base_url`) and marks `local` active, so a fresh workspace works out
  of the box.
- **(C) Variable values are edited by hand in the YAML file**, consistent with
  how `request new` / `collection init` scaffold a file the user then edits.
  `env set` / `env show` are out of scope for this iteration.

## Data model

```yaml
# environments/dev.yml
name: dev
base_url: https://api.dev.example.com
variables:
  token: abc123
  user_id: "42"
```

- **`domain.Environment`** (new): `Name string`, `BaseURL string`,
  `Variables map[string]string`.
- **`domain.Workspace`**: `BaseURL` is kept as a **deprecated fallback** field
  (so older `workspace.yml` files still read), but `gon init` stops writing it
  and the active environment is the source of truth.
- **Workspace `config`** (headers/query/path) is unchanged structurally; its
  values may contain `{{var}}`.

## Variable substitution

- Syntax: `{{name}}`, tolerating inner whitespace (`{{ name }}`). Names match
  `[A-Za-z0-9_.-]+`. Single pass — no nested/recursive resolution.
- Applied inside `HttpService.Execute`, **after** `ApplyDefaults`, to:
  - the final resolved URL,
  - every header value,
  - every query value,
  - the request body bytes.
- **Fail-fast:** after substitution, if any `{{...}}` placeholder remains, the
  request fails with an error listing the unresolved variable names. (When no
  environment is active, see fallback rules — substitution is a no-op and a URL
  containing `{{...}}` would surface the same fail-fast error.)

## Storage & local state

- **Definitions:** `environments/<name>.yml` at the workspace root.
- **Active environment:** `.gon/active-env` — a single line holding the active
  environment name. `.gon/` is already gitignored, so the choice is local and
  per-developer. The `.gon/` directory is created on demand.

## Active-environment resolution (precedence)

`EnvironmentService.Resolve(ctx, root, flagEnv)` returns the active
`*domain.Environment`:

1. If `flagEnv` (`--env <name>`) is set → load that environment (error if it
   does not exist).
2. Else if `.gon/active-env` names an existing environment → load it.
3. Fallback when neither is set:
   - Exactly one environment exists → use it.
   - Zero environments exist → return `nil` (requests run without an
     environment; `ResolveURL` falls back to the deprecated `workspace.BaseURL`).
   - More than one environment and none active → error:
     `no active environment; run 'env use <name>' or pass --env`.

## Architecture (hexagonal) & call-flow changes

Direction stays inward toward `core`. Resolution is performed by the command
adapter, which passes the resolved environment into the services.

**New / changed files:**

- `internal/core/domain/`
  - `environment.go` (new): `Environment` struct, `Substitute(s string)`
    helper, and URL resolution (`base_url` + `config.Path` + path, absolute
    `http(s)://` bypasses).
  - `workspace.go` (changed): `BaseURL` documented as deprecated fallback; URL
    resolution updated to prefer the environment's `base_url`.
- `internal/core/port/driving/`
  - `environment_service.go` (new): `EnvironmentService` (`Create`, `List`,
    `Use`, `Resolve`).
  - `http_service.go` (changed): `Execute(ctx, input, env)`.
  - `request_service.go` (changed): `Run(ctx, root, path, overrides, env)`.
- `internal/core/port/driven/`
  - `environment_repository.go` (new): `EnvironmentRepository` —
    `Save`, `Load`, `List`, `Exists`, plus `ReadActive` / `WriteActive` for the
    `.gon/active-env` state.
- `internal/core/service/`
  - `environment_service.go` (new): implements the driving port; `Create` calls
    the shared `ensureWorkspace` guard.
  - `http_service.go` (changed): apply substitution + accept env.
  - `request_service.go` (changed): thread env through to `Execute`.
- `internal/adapter/model/`
  - `environment_model.go` (new): `EnvironmentModel` ↔ `domain.Environment`.
- `internal/adapter/repository/`
  - `environment_repository.go` (new): YAML persistence for env files and the
    active-env state file.
- `internal/adapter/command/`
  - `env_command.go` (new): `env` parent command with `new`, `list`, `use`.
  - `http_command.go` (changed): add `--env` flag; resolve and pass env.
  - `run_command.go` (changed): add `--env` flag; resolve and pass env.
- `cmd/main.go` (changed): wire `EnvironmentRepository` / `EnvironmentService`;
  REPL re-reads `.gon/active-env` after each command to refresh the prompt.
- `internal/core/service/workspace_service.go` (changed): `Create` stops writing
  `base_url`; instead scaffolds `environments/local.yml` and marks it active.

**Execute flow (HTTP):**
`HttpCommand` resolves env via `EnvironmentService.Resolve` (honoring `--env`) →
`HttpService.Execute(ctx, input, env)` → `ApplyDefaults` (shared config) →
resolve URL using `env.BaseURL` + `config.Path` → substitute `{{var}}` across
URL/headers/query/body (fail-fast on leftovers) → `HttpOutput.Format`.

**Run flow:** `RunCommand` resolves env → `RequestService.Run(..., env)` loads
the request + collection chain → calls `HttpService.Execute(ctx, merged, env)`.

## CLI / REPL

New help group **"Environments"**, parent command `env`:

- `env new <name>` — scaffold `environments/<name>.yml` (base_url placeholder +
  empty `variables`). Requires an initialized workspace (`ensureWorkspace`).
- `env list` (alias `ls`) — list environment names, marking the active one with
  `*`.
- `env use <name>` — set the active environment (writes `.gon/active-env`,
  verifying the env exists). In the REPL, the prompt becomes `gon(proj:dev)>`.
- `--env <name>` flag added to `get`/`post`/`put`/`delete`/`patch` and `run`
  for a per-command override.

`gon init` scaffolds `environments/local.yml` and marks `local` active.

The REPL rebuilds nothing on switch: the command layer reads the active env per
call. After each command, the REPL re-reads `.gon/active-env` to refresh the
prompt string.

## Testing

- **domain** (`environment_test.go`): `Substitute` for defined, undefined,
  multiple, and whitespaced placeholders; URL resolution with env `base_url`
  and absolute-URL bypass.
- **repository** (`environment_repository_test.go`): save/load/list/exists
  round-trip for env files; `ReadActive`/`WriteActive` round-trip.
- **service** (`environment_service_test.go`): `Resolve` precedence (flag >
  active > single-env fallback > error on multiple/none); zero-env → nil;
  `Create` requires a workspace.
- **http_service** (`http_service_test.go`): substitution applied to
  URL/headers/query/body; fail-fast error on missing variable.
- **request_service** (`request_service_test.go`): env threaded through `Run`.
- **command**: `env new` / `list` / `use` behavior and `--env` flag parsing.

## Out of scope (future)

- `env set <name> KEY=VALUE` and `env show <name>`.
- Nested/recursive variable resolution.
- Secret-specific handling (encryption, separate secret files).
- A committed `default_environment` in `workspace.yml`.
