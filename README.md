# gon

An interactive HTTP client for terminal lovers.

`gon` runs as a REPL in your terminal — send HTTP requests, inspect colorized responses, and keep your workflow in the shell without leaving it.

## Features

- Interactive REPL with per-workspace command history
- One-shot mode for scripting and quick use
- HTTP methods: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`
- Custom request headers via `--header` and query params via `--query`
- JSON request body via `--json`
- Selectable output verbosity: `--minimal`, `--normal`, `--full`
- Per-request `--timeout`
- Workspaces — a project base URL plus default headers, query params, and base path applied to every request
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

The REPL supports command history (stored at `/tmp/gon.history`). Press `^C` to interrupt or type `exit` to quit.

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

### Output verbosity

Choose how much of the exchange is printed. Defaults to `--normal`.

- `--minimal` — status code and headers only
- `--normal` — status code, headers, and body
- `--full` — request details plus the full response

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

Run `gon init` in a directory to create a `workspace.yml`. A workspace gives every
request a base URL plus defaults that are applied automatically — so you don't
repeat the same host, auth header, or query string on every call. The directory
itself becomes the home for your collections and saved requests, so
`collection init`, `request new`, and `run` all require it. Because everything is
plain files at the workspace root, the whole folder is a self-contained, shareable
artifact — hand it to a frontend teammate or commit it as its own repo.

```yaml
name: my-project
base_url: https://api.example.com
config:
  path: /v1
  headers:
    Authorization: Bearer my-token
  query:
    api_key: abc123
```

With the workspace above, a relative request:

```
gon> get /users
```

is sent to `https://api.example.com/v1/users` with the `Authorization` header and
`api_key` query parameter already attached.

- **Relative URLs** resolve to `base_url` + `config.path` + the request path.
  Absolute `http(s)://` URLs are used as-is and ignore the workspace.
- **Defaults are overridable** — a per-request `--header` or `--query` for the same
  key wins over the workspace default; the default is dropped, not duplicated.
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

## License

MIT
