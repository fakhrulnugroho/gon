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
    port/driving/            ← input port interfaces (e.g. HttpService)
    port/driven/             ← output port interfaces (e.g. HttpOutput, Formatter[T])
    service/                 ← implements driving ports
    formatter/               ← implements driven Formatter[T] port
    payload/                 ← data structs shared across layers (HttpExecuteInput/Output)
  adapter/
    command/                 ← driving adapters — CLI commands via urfave/cli/v3
    output/                  ← driven adapters — formats and prints responses
  utility/                   ← ANSI color helpers, JSON pretty-printer
  version/                   ← Version/OS/Arch vars, injected via -ldflags at build
```

### Key flow

`HttpCommand` (adapter/command) → `HttpService.Execute` (core/service) → `HttpOutput.Format` (adapter/output)

- `HttpCommand` parses CLI flags, builds `HttpExecuteInput`, calls the service, then delegates rendering to `HttpOutput`.
- `HttpService` makes the real HTTP call and returns `HttpExecuteOutput`.
- `HttpOutput` holds two formatters (`Formatter[[]byte]` for JSON bodies, `Formatter[map[string]string]` for headers) and three display modes: `0`=minimal (status only), `1`=normal (headers + body), `2`=full (request echo + response).

### Adding a new command

1. Add a driving port interface under `internal/core/port/driving/` if new domain logic is needed.
2. Implement the service in `internal/core/service/`.
3. Create the CLI command adapter in `internal/adapter/command/` and wire it in `cmd/main.go`.

### Version injection

`internal/version/version.go` declares `var Version = "dev"`. CI overrides it at build time with `-ldflags "-X 'gon/internal/version.Version=<tag>'"`.

## Release

Releases are published via GitHub Actions (`.github/workflows/release.yml`). Use the `/release-notes` skill to generate and publish a new release.
