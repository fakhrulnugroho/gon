# gon

An interactive HTTP client for terminal lovers.

`gon` runs as a REPL in your terminal — send HTTP requests, inspect colorized responses, and keep your workflow in the shell without leaving it.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Platform Support](#platform-support)
- [Usage](#usage) — [interactive](#interactive-mode) / [one-shot](#one-shot-mode)
- [Commands](#commands)
- [Options](#options)
- [Examples](#examples)
- [Workspaces](#workspaces)
- [Collections & Saved Requests](#collections--saved-requests)
- [Environments](#environments)
- [Project Layout](#project-layout)
- [Output](#output)
- [Contributing](#contributing)
- [License](#license)

## Features

- Interactive REPL with per-workspace command history
- One-shot mode for scripting and quick use
- HTTP methods: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`
- Custom request headers via `--header` and query params via `--query`
- JSON request body via `--json`
- Selectable output verbosity: `--minimal`, `--normal`, `--full`
- Per-request `--timeout`
- Workspaces — default headers, query params, and base path applied to every request
- Environments — named variable sets (e.g. `local`, `dev`, `prod`) each with their own base URL; switch with `env use` or `--env`
- Collections & saved requests — store requests as YAML files, organize them in nested folders, and run them by path with hierarchical config defaults
- Color-coded HTTP status codes and execution time
- Syntax-highlighted, pretty-printed JSON responses

## Installation

### Via install script (Linux & macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/fakhrulnugroho/gon/main/install.sh | bash
```

The script will:
- Detect your OS and architecture
- Download the latest release binary
- Install it to `~/.gon/bin/`
- Add it to your `PATH` via `.bashrc` or `.zshrc`

After installation, reload your shell:

```bash
source ~/.bashrc   # or ~/.zshrc
```

### Build from source

Requires **Go 1.25+**.

```bash
git clone https://github.com/fakhrulnugroho/gon.git
cd gon
go build -o gon ./cmd
```

### Uninstall

The install script places the binary in `~/.gon/bin/` and adds that directory to
your `PATH`. To remove it:

```bash
rm -rf ~/.gon/bin
```

Then delete the `export PATH=".../.gon/bin:$PATH"` line the script appended to your
`~/.bashrc` or `~/.zshrc`.

## Quick Start

A five-minute tour from an empty folder to a saved, reusable request.

**1. Create a project folder and initialize a workspace.**

```bash
mkdir my-api && cd my-api
gon init
```

This writes `workspace.yml` (shared request defaults) and `environments/local.yml`
(your first environment, marked active). The workspace name is derived from the
folder, so this one is called `my-api`.

**2. Point the `local` environment at your API.** Edit `environments/local.yml`:

```yaml
name: local
base_url: https://jsonplaceholder.typicode.com
variables:
  token: dev-secret-123
```

`base_url` is where relative requests are sent; `variables` are values you can
reference anywhere with `{{name}}`.

**3. Send your first request.** Start the REPL and use a relative URL — it resolves
against the environment's `base_url`:

```
$ gon
gon(my-api)> get /todos/1

200 OK (84ms)

{
  "userId": 1,
  "id": 1,
  "title": "delectus aut autem",
  "completed": false
}
```

**4. Save a request you'll reuse.** Scaffold it, then edit the generated file:

```
gon(my-api)> request new todos/create --method POST
```

```yaml
# todos/create.yml
name: create
method: POST
url: /todos
headers:
  Authorization: Bearer {{token}}   # resolved from the active environment
body:
  json:
    title: Buy milk
    completed: false
```

**5. Run it by name — any time, from anyone who clones the folder:**

```
gon(my-api)> run todos/create
```

That's the whole loop: `init` → set an environment → send requests → save the
useful ones → `run` them by path. The next sections cover each piece in depth.

## Platform Support

| OS      | amd64 | arm64 |
|---------|-------|-------|
| Linux   | ✓     | ✓     |
| macOS   | ✓     | ✓     |
| Windows | ✓     | —     |

## Usage

### Interactive mode

Start `gon` without any arguments to enter the REPL:

```
$ gon
gon — An interactive HTTP client for terminal lovers
Type 'help' for available commands

gon> _
```

The REPL supports command history. Outside a workspace it is stored at
`/tmp/gon.history`; inside a workspace it is kept per-workspace under
`.cache/<workspace-name>.history`, so each project has its own history. Press `^C`
to interrupt or type `exit` to quit. When you're in a workspace the prompt shows
its name, e.g. `gon(my-api)>`.

### One-shot mode

Pass a command directly as arguments to run it and exit immediately:

```bash
gon get https://api.example.com/users
```

This is useful for scripting or quick one-off requests.

---

## Commands

### HTTP Commands

```
get    <url> [options]    Send an HTTP GET request
post   <url> [options]    Send an HTTP POST request
put    <url> [options]    Send an HTTP PUT request
patch  <url> [options]    Send an HTTP PATCH request
delete <url> [options]    Send an HTTP DELETE request
```

### Workspace Commands

```
init    Create a workspace.yml in the current directory
```

### Environment Commands

These commands require an initialized workspace (`init`).

```
env new  <name>    Scaffold environments/<name>.yml; edit it to set base_url and variables
env list           List all environments (active one marked with *)
env use  <name>    Switch the active environment
```

### Collection Commands

These commands require an initialized workspace (`init`) — collections and
requests are stored at the workspace root, alongside `workspace.yml`. Without one
they exit with `no gon workspace found, run 'init' first`.

```
run        <path> [options]    Run a saved request by path (e.g. run auth/login)
collection init <name>         Create a collection folder (nesting allowed, e.g. auth/admin)
request    new <path> [--method]  Scaffold a new request file (e.g. request new auth/login --method POST)
```

### Common Commands

```
help      Print available commands
version   Print version info
clear     Clear the terminal screen
exit      Exit the application
```

---

## Options

Options are appended after the URL and can be combined freely. They apply to the
HTTP commands (`get`, `post`, …) and to `run`.

### `--header "Key: Value"`

Add a custom request header in `"Key: Value"` format.

```bash
get https://api.example.com/users --header "Authorization: Bearer token123"
```

Multiple headers can be added by repeating the flag:

```bash
get https://api.example.com/users \
  --header "Authorization: Bearer token123" \
  --header "Accept: application/json"
```

### `--query "Key=Value"`

Add a URL query parameter in `"Key=Value"` format. Repeat the flag for more.

```bash
get https://api.example.com/users --query "page=1" --query "limit=20"
```

### `--json <json-string>`

Set the request body to a JSON string. Automatically adds `Content-Type: application/json`.

```bash
post https://api.example.com/users --json '{"name":"Alice","email":"alice@example.com"}'
```

### `--timeout <duration>`

Set the request timeout (default `30s`). Accepts Go duration strings like `5s`, `1500ms`, `1m`.

```bash
get https://api.example.com/slow --timeout 5s
```

### `--env <name>`

Override the active environment for a single request. Useful for one-off calls
against a different target without changing your persisted active selection.

```bash
gon get /users --env prod
```

### Output verbosity

Choose how much of the exchange is printed. Defaults to `--normal`.

- `--minimal` — status code and response body only (no headers)
- `--normal` — status code, response headers, and body (this is the default)
- `--full` — everything `--normal` shows, plus an echo of the request that was
  actually sent (method, URL, merged headers, and body) above a separator

If you pass more than one, `--minimal` wins over `--full`, which wins over
`--normal`.

```bash
get https://api.example.com/users --full
```

---

## Examples

**Simple GET request:**
```
gon> get https://jsonplaceholder.typicode.com/todos/1
```

**POST with JSON body:**
```
gon> post https://jsonplaceholder.typicode.com/posts --json '{"title":"foo","body":"bar","userId":1}'
```

**GET with custom header:**
```
gon> get https://api.example.com/profile --header "Authorization: Bearer mytoken"
```

**GET with query parameters:**
```
gon> get https://api.example.com/users --query "page=1" --query "limit=20"
```

**PATCH with JSON body and header:**
```
gon> patch https://api.example.com/users/42 --json '{"name":"Bob"}' --header "Authorization: Bearer mytoken"
```

**DELETE a resource:**
```
gon> delete https://api.example.com/users/42
```

**Run a saved request with an override:**
```
gon> run auth/login --json '{"username":"admin","password":"secret"}'
```

**One-shot from the shell:**
```bash
gon post https://api.example.com/login --json '{"username":"admin","password":"secret"}'
```

---

## Workspaces

Run `gon init` in a directory to create a `workspace.yml` and a default
`environments/local.yml`. A workspace gives every request shared defaults —
headers, query params, and a base path — applied automatically, so you don't
repeat the same auth header or query string on every call. The base URL lives in
the active environment (see [Environments](#environments) below). The directory
itself becomes the home for your environments, collections, and saved requests,
so `collection init`, `request new`, and `run` all require it. Because everything
is plain files at the workspace root, the whole folder is a self-contained,
shareable artifact — hand it to a frontend teammate or commit it as its own repo.

```yaml
name: my-project
config:
  path: /v1
  headers:
    Authorization: Bearer {{token}}
  query:
    api_key: abc123
```

With the workspace above, a relative request:

```
gon> get /users
```

is sent to `<active-environment-base_url>/v1/users` with the `Authorization`
header (resolved from the active environment's `token` variable) and `api_key`
query parameter already attached.

- **Relative URLs** resolve to the active environment's `base_url` + `config.path` + the request path.
  Absolute `http(s)://` URLs are used as-is and ignore the workspace.
- **Defaults are overridable** — a per-request `--header` or `--query` for the same
  key wins over the workspace default; the default is dropped, not duplicated.
- **`{{var}}` in config** — workspace `config` values may contain `{{name}}`
  placeholders that resolve from the active environment.
- When you're inside a workspace the REPL prompt shows its name, e.g.
  `gon(my-project)>`, and command history is kept per-workspace under `.cache/`.

---

## Collections & Saved Requests

Save requests to YAML files so you can run them by name instead of retyping the
URL, headers, and body each time. Everything lives at the workspace root, so run
`gon init` first. Requests are plain files; a folder becomes a **collection** when
it holds a `collection.yml` that defines shared defaults.

### Scaffold a request

```
gon> request new auth/login --method POST
```

This creates `auth/login.yml` and, if needed, a `collection.yml` for every
folder along the path (`auth/` here). Edit the generated file to fill in the
URL, headers, query, and body:

```yaml
# auth/login.yml
name: login
method: POST
url: /login
headers:
  Accept: application/json
query:
  verbose: "true"
body:
  json:
    username: admin
    password: secret
```

A request body can be one of `json`, `raw` (with an optional `contentType`), or
`form` — only one kind may be set.

### Create a collection

```
gon> collection init auth/admin
```

Creates the nested folders and a `collection.yml` in each. A
`collection.yml` holds defaults applied to every request beneath it:

```yaml
# auth/collection.yml
name: auth
config:
  path: /auth
  headers:
    Authorization: Bearer my-token
  query:
    api_key: abc123
```

### Run a request

```
gon> run auth/login
```

`run` loads the request, layers on the config from each enclosing collection
(and the workspace), and sends it. Resolution rules:

- **Paths are additive** — each collection's `config.path` is prefixed
  outermost-first, then the request's own `url`, so `auth/login` above resolves
  to `<base_url>/auth/login`. Absolute `http(s)://` URLs bypass this.
- **Inner collections win** — when collections set the same header or query key,
  the nearest collection to the request takes precedence.
- **CLI flags override everything** — `--header`, `--query`, and `--json` passed
  to `run` replace the saved/collection values for that key.

```
gon> run auth/login --header "X-Trace: 1" --full
```

---

## Environments

Environments are project-scoped, named sets of variables plus a base URL
(`local`, `dev`, `test`, `prod`). `gon init` creates `environments/local.yml`
and marks it active.

```bash
gon env new dev             # creates environments/dev.yml — edit to set base_url and variables
gon env list                # list environments; active one is marked with *
gon env use dev             # switch the active environment
gon get /users --env prod   # override the active environment for one call
```

```yaml
# environments/dev.yml
name: dev
base_url: https://api.dev.example.com
variables:
  token: abc123
  user_id: "42"
```

Requests (and workspace `config` defaults) reference variables with `{{name}}` in
the URL, header values, query values, and body — for example
`Authorization: Bearer {{token}}` or `/users/{{user_id}}`. Values resolve from
the active environment at request time; an unresolved `{{var}}` fails the request
with an error listing the missing variable names.

The active selection is persisted in `.gon/active-env` (gitignored), so each
developer on the team can choose their own environment independently.

**Precedence** — when determining which environment to use:
1. `--env <name>` flag (per-call override)
2. Persisted active selection (`env use`)
3. The sole environment, if exactly one exists
4. Error — if multiple environments exist and none is active, `gon` asks you to
   run `env use <name>` or pass `--env`.

---

## Project Layout

Everything `gon` creates is plain text at the workspace root, so the whole folder
is a self-contained, shareable artifact — commit it to git, or hand it to a
teammate and they get your environments, collections, and saved requests intact.

```
my-api/
├── workspace.yml            # workspace name + shared config defaults (path, headers, query)
├── environments/
│   ├── local.yml            # an environment: base_url + variables
│   └── prod.yml
├── auth/                    # a collection (any folder with a collection.yml)
│   ├── collection.yml       # defaults applied to every request beneath it
│   └── login.yml            # a saved request
├── todos/
│   └── create.yml
├── .gon/
│   └── active-env           # per-developer active environment (gitignored)
└── .cache/
    └── my-api.history       # per-workspace REPL history (gitignored)
```

| Path | Purpose | Commit to git? |
|------|---------|----------------|
| `workspace.yml` | Workspace name and shared `config` defaults | ✅ Yes |
| `environments/*.yml` | Named variable sets, each with a `base_url` | ✅ Yes |
| `<path>/collection.yml` | Defaults shared by requests in that folder | ✅ Yes |
| `<path>/<name>.yml` | A saved request | ✅ Yes |
| `.gon/active-env` | Which environment *you* have selected | ❌ No (gitignored) |
| `.cache/*.history` | Your REPL command history | ❌ No (gitignored) |

`gon init` scaffolds a `.gitignore` that excludes `.gon/` and `.cache/`, so the
per-developer files never get committed — each teammate keeps their own active
environment and history.

---

## Output

Every response is displayed with:

- **HTTP status** — color-coded by class:
  - 🟢 `2xx` — green (success)
  - 🔵 `3xx` — blue (redirect)
  - 🟡 `4xx` — yellow (client error)
  - 🔴 `5xx` — red (server error)

- **Execution time** — color-coded by speed:
  - 🟢 `< 100ms` — green
  - 🟡 `100–499ms` — yellow
  - 🔴 `≥ 500ms` — red

- **Response body** — pretty-printed and syntax-highlighted JSON (keys, strings, numbers, booleans, and nulls each in a distinct color).

Example output:
```
200 OK (42ms)

{
  "id": 1,
  "title": "foo",
  "completed": false
}
```

---

## Contributing

Contributions are welcome. To work on `gon` locally:

```bash
git clone https://github.com/fakhrulnugroho/gon.git
cd gon

go build -o gon ./cmd     # build
go test ./...             # run the test suite
go run ./cmd              # run the REPL from source
```

`gon` follows a **hexagonal architecture** (ports & adapters) — the dependency
direction always points inward toward `core`, and domain structs are never
marshalled directly (adapters map them through serialization models). Before
adding a command, read [`CLAUDE.md`](./CLAUDE.md), which documents the layer
boundaries and the step-by-step recipe for wiring a new command end to end.

Please keep new code covered by tests (`go test ./...` must pass) and run
`gofmt` before opening a pull request.

---

## License

MIT
