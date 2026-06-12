# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o gon ./cmd

# Build with version embedded (mirrors CI)
go build -ldflags "-s -w -X 'gon/internal/version.Version=v1.x.x'" -o gon ./cmd

# Run tests
go test ./...

# Run a specific test
go test ./internal/... -run TestName

# Run (REPL)
go run ./cmd

# Run one-shot
go run ./cmd get https://example.com
```

## Architecture

`gon` follows **hexagonal architecture** (ports & adapters). Dependency direction always points inward toward `core`.

```
cmd/main.go                  ← wires everything, runs REPL or one-shot
internal/
  core/
    domain/                  ← domain entities (Workspace, Config, Environment)
    port/driving/            ← input port interfaces (e.g. HttpService, WorkspaceService, EnvironmentService)
    port/driven/             ← output port interfaces (e.g. HttpOutput, Formatter[T], WorkspaceRepository, EnvironmentRepository)
    service/                 ← implements driving ports
    formatter/               ← implements driven Formatter[T] port
    payload/                 ← data structs shared across layers (HttpExecuteInput/Output)
  adapter/
    command/                 ← driving adapters — CLI commands via urfave/cli/v3
    output/                  ← driven adapters — formats and prints responses
    repository/              ← driven adapters — YAML persistence (workspace, environments, collections, requests)
    model/                   ← serialization models for YAML (domain → model mapping)
  utility/                   ← ANSI color helpers, JSON pretty-printer
  version/                   ← Version/OS/Arch vars, injected via -ldflags at build
```

### Key flow

`HttpCommand` (adapter/command) → `HttpService.Execute` (core/service) → `HttpOutput.Format` (adapter/output)

- `HttpCommand` parses CLI flags, builds `HttpExecuteInput`, calls the service, then delegates rendering to `HttpOutput`.
- `HttpService` makes the real HTTP call and returns `HttpExecuteOutput`.
- `HttpOutput` holds two formatters (`Formatter[[]byte]` for JSON bodies, `Formatter[map[string]string]` for headers) and three display modes: `0`=minimal (status only), `1`=normal (headers + body), `2`=full (request echo + response).

### Workspace

`gon init` creates `workspace.yml` in the current directory (the CLI command is named `init`, registered under the "Workspace" help group). The workspace name is derived from the folder name (converted to kebab-case). The YAML is written via `WorkspaceRepository` (adapter/repository) using `WorkspaceModel` (adapter/model) as the serialization layer — domain structs are never marshalled directly.

The workspace `Config` (default `headers`, `query`, and base `path`) is applied to every request by `Workspace.ApplyDefaults` (domain), called from `HttpService.Execute`. `ResolveURL` resolves a relative request URL to `BaseURL + Config.Path + path`; absolute `http(s)://` URLs bypass it. Per-request `--header`/`--query` flags take precedence over the workspace defaults (the default for a colliding key is dropped, not duplicated). Because `ApplyDefaults` mutates the shared `HttpExecuteInput`, the `--full` display echoes the merged request.

### Collections & requests

Collections and saved requests live **at the workspace root**, alongside `workspace.yml` — so the whole folder is a self-contained, shareable artifact. `collection init auth/admin` writes `auth/admin/collection.yml`; `request new auth/login` writes `auth/login.yml`. The repositories receive the project root (cwd) and join request/collection paths directly onto it. `RequestRepository.Load` walks the collection chain from the request's folder up to the workspace root (nearest-first).

All three operations — `collection init`, `request new`, and `run` — require an initialized workspace. `CollectionService`/`RequestService` are injected with `WorkspaceRepository` and call the shared `ensureWorkspace` guard (core/service) first; when `workspace.yml` is absent (`WorkspaceRepository.Exists` returns false) they fail with `no gon workspace found, run 'init' first`.

### Environments

`domain.Environment` holds `Name`, `BaseURL`, and `Variables` (a `map[string]string`). It exposes `Substitute(s string) string` (replaces every `{{name}}` placeholder it knows; unknown placeholders are left intact so callers can detect them — a nil `*Environment` returns the input unchanged) and `FindPlaceholders(s string) []string` (returns the names of any remaining `{{name}}` placeholders). A `ResolveURL(baseURL, configPath, requestPath string)` free function in `domain` resolves a relative URL against a base URL and config path; `Workspace.ResolveURL` delegates to it. Fail-fast on unresolved variables is the HTTP service's responsibility, not `Substitute`'s: after substitution it collects leftover placeholders via `FindPlaceholders` and errors before the request is sent.

Storage: each environment lives in `environments/<name>.yml` at the workspace root, written and read through `EnvironmentRepository` (adapter/repository) using `EnvironmentModel` (adapter/model) as the serialization layer. The per-developer active selection is persisted in `.gon/active-env` (gitignored) via `ReadActive`/`WriteActive` on the repository — not in `workspace.yml`.

`EnvironmentService.Resolve` applies the following precedence to pick the active environment:
1. An explicit `--env <name>` flag passed by the command adapter.
2. The persisted active selection from `.gon/active-env`.
3. The sole environment when exactly one exists.
4. `(nil, nil)` when zero environments are present (workspace-less call is still valid).
5. An error when multiple environments exist and none is active (tells the user to run `env use` or pass `--env`).

Resolution happens in the command adapter layer. The resolved `*domain.Environment` is threaded into `HttpService.Execute(ctx, input, env)` and `RequestService.Run(..., env)`, which call `env.Substitute` on the URL, each header value, each query value, and the body; any unresolved `{{var}}` causes a fast-fail error before the HTTP call is made.

Workspace `config` defaults stay shared in `workspace.yml` and may contain `{{var}}` placeholders — they are substituted with the same `env.Substitute` pass after `Workspace.ApplyDefaults` merges them into `HttpExecuteInput`. `workspace.BaseURL` is retained as a deprecated fallback for workspaces created before the environments feature; the active environment's `base_url` is the authoritative source of truth. `gon init` no longer writes `base_url` into `workspace.yml`; instead it scaffolds `environments/local.yml` and marks `local` active.

The `env` command group (`env new`, `env list`/`ls`, `env use`) is registered under an "Environments" help group in `cmd/main.go`.

### Adding a new command

1. Define domain entities in `internal/core/domain/` if new data types are needed.
2. Add a driving port interface under `internal/core/port/driving/`.
3. Add driven port interfaces under `internal/core/port/driven/` for any external dependencies (output, persistence).
4. Implement the service in `internal/core/service/`.
5. Implement driven adapters: output formatters in `adapter/output/`, repositories in `adapter/repository/` (with a model in `adapter/model/` for serialization).
6. Create the CLI command adapter in `internal/adapter/command/` and wire everything in `cmd/main.go`.

### Version injection

`internal/version/version.go` declares `var Version = "dev"`. CI overrides it at build time with `-ldflags "-X 'gon/internal/version.Version=<tag>'"`.

## Release

Releases are published via GitHub Actions (`.github/workflows/release.yml`). Use the `/release-notes` skill to generate and publish a new release.
