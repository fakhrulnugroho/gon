# gon

An interactive HTTP client for terminal lovers.

`gon` runs as a REPL in your terminal — send HTTP requests, inspect colorized responses, and keep your workflow in the shell without leaving it.

## Features

- Interactive REPL with command history
- One-shot mode for scripting and quick use
- HTTP methods: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`
- Custom request headers via `--header`
- JSON request body via `--json`
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

### Common Commands

```
help      Print available commands
version   Print version info
clear     Clear the terminal screen
exit      Exit the application
```

---

## Options

Options are appended after the URL and can be combined freely.

### `--header <key> <value>`

Add a custom request header.

```bash
get https://api.example.com/users --header Authorization "Bearer token123"
```

Multiple headers can be added by repeating the flag:

```bash
get https://api.example.com/users \
  --header Authorization "Bearer token123" \
  --header Accept "application/json"
```

### `--json <json-string>`

Set the request body to a JSON string. Automatically adds `Content-Type: application/json`.

```bash
post https://api.example.com/users --json '{"name":"Alice","email":"alice@example.com"}'
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
gon> get https://api.example.com/profile --header Authorization "Bearer mytoken"
```

**PATCH with JSON body and header:**
```
gon> patch https://api.example.com/users/42 --json '{"name":"Bob"}' --header Authorization "Bearer mytoken"
```

**DELETE a resource:**
```
gon> delete https://api.example.com/users/42
```

**One-shot from the shell:**
```bash
gon post https://api.example.com/login --json '{"username":"admin","password":"secret"}'
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
