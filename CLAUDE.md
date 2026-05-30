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
    domain/                  ← domain entities (Workspace, Config)
    port/driving/            ← input port interfaces (e.g. HttpService, WorkspaceService)
    port/driven/             ← output port interfaces (e.g. HttpOutput, Formatter[T], WorkspaceRepository)
    service/                 ← implements driving ports
    formatter/               ← implements driven Formatter[T] port
    payload/                 ← data structs shared across layers (HttpExecuteInput/Output)
  adapter/
    command/                 ← driving adapters — CLI commands via urfave/cli/v3
    output/                  ← driven adapters — formats and prints responses
    repository/              ← driven adapters — YAML persistence (workspace)
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

`gon workspace init` creates `.gon/workspace.yaml` in the current directory. The workspace name is derived from the folder name (converted to kebab-case). The YAML is written via `WorkspaceRepository` (adapter/repository) using `WorkspaceModel` (adapter/model) as the serialization layer — domain structs are never marshalled directly.

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
