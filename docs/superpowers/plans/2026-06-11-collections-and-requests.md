# Collections & Requests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add saveable, version-controllable HTTP requests organized into nested collection folders, runnable by path (`run auth/login`), with Bruno-style inherited shared config.

**Architecture:** Hexagonal (ports & adapters), matching the existing workspace feature. A folder becomes a *collection* by containing `collection.yml` (shared `domain.Config`); every other `.yml` is a single *request*. Running merges values inner→outer (request flags > request file > nearest collection > … > workspace) by reusing the existing additive `ApplyDefaults` merge, then delegates the actual HTTP call to the existing `HttpService`.

**Tech Stack:** Go 1.25, `urfave/cli/v3` (commands), `gopkg.in/yaml.v3` (persistence), `iancoleman/strcase` (kebab-case names), `stretchr/testify` (tests).

---

## File Structure

**New domain (`internal/core/domain/`):**
- `request.go` — `Request`, `RequestBody`, `BodyKind`; body encoding + `ToInput`.
- `collection.go` — `Collection` (wraps existing `Config`).
- `config.go` — *modified*: extract header/query merge into `Config.ApplyDefaults`.
- `workspace.go` — *modified*: `Workspace.ApplyDefaults` delegates to `Config.ApplyDefaults`.

**New serialization models (`internal/adapter/model/`):**
- `request_model.go` — YAML ↔ `domain.Request`, body validation (exactly one body kind).
- `collection_model.go` — YAML ↔ `domain.Collection` (reuses `ConfigModel`).

**New driven ports (`internal/core/port/driven/`):**
- `request_repository.go` — load request + ancestor collection chain; save; exists.
- `collection_repository.go` — save collection; exists.

**New driven adapters (`internal/adapter/repository/`):**
- `request_repository.go` — filesystem load/save + ancestor walk.
- `collection_repository.go` — filesystem save + exists.

**New driving ports (`internal/core/port/driving/`):**
- `request_service.go` — `Run` + `Create` (scaffold request).
- `collection_service.go` — `Create` (scaffold collection).

**New driving services (`internal/core/service/`):**
- `request_service.go` — merge precedence + delegate to `HttpService`; scaffold.
- `collection_service.go` — scaffold nested collections.

**New driving adapters (`internal/adapter/command/`):**
- `run_command.go`, `collection_init_command.go`, `request_new_command.go`.

**Wiring:**
- `cmd/main.go` — *modified*: construct new repos/services, add a "Collections" command group.

---

## Task 1: Extract `Config.ApplyDefaults` (reuse-ready merge)

**Files:**
- Modify: `internal/core/domain/config.go`
- Modify: `internal/core/domain/workspace.go:24-49` (the `ApplyDefaults` method)
- Test: `internal/core/domain/config_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `internal/core/domain/config_test.go`:

```go
package domain

import (
	"testing"

	"gon/internal/core/payload"

	"github.com/stretchr/testify/assert"
)

func TestConfigApplyDefaults(t *testing.T) {
	t.Run("injects header and query when absent", func(t *testing.T) {
		c := Config{
			Headers: map[string]string{"Authorization": "Bearer token"},
			Query:   map[string]string{"debug": "1"},
		}
		input := &payload.HttpExecuteInput{}

		c.ApplyDefaults(input)

		assert.Equal(t, []string{"Bearer token"}, input.Headers["Authorization"])
		assert.Equal(t, []string{"1"}, input.Query["debug"])
	})

	t.Run("existing key wins and is not duplicated", func(t *testing.T) {
		c := Config{Headers: map[string]string{"authorization": "Bearer default"}}
		input := &payload.HttpExecuteInput{
			Headers: map[string][]string{"Authorization": {"Bearer override"}},
		}

		c.ApplyDefaults(input)

		assert.Equal(t, []string{"Bearer override"}, input.Headers["Authorization"])
	})

	t.Run("empty config is a no-op", func(t *testing.T) {
		c := Config{}
		input := &payload.HttpExecuteInput{}

		c.ApplyDefaults(input)

		assert.Empty(t, input.Headers)
		assert.Empty(t, input.Query)
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/domain/ -run TestConfigApplyDefaults -v`
Expected: FAIL — `c.ApplyDefaults undefined (type Config has no field or method ApplyDefaults)`

- [ ] **Step 3: Add `ApplyDefaults` to `Config`**

Replace the contents of `internal/core/domain/config.go` with:

```go
package domain

import (
	"net/textproto"

	"gon/internal/core/payload"
)

type Config struct {
	Path    string
	Query   map[string]string
	Headers map[string]string
}

// ApplyDefaults merges this config's default headers and query parameters into
// input. It is additive: a default is added only when input does not already
// supply that key, so more specific values always win.
func (c *Config) ApplyDefaults(input *payload.HttpExecuteInput) {
	for key, value := range c.Headers {
		canonical := textproto.CanonicalMIMEHeaderKey(key)
		if _, ok := input.Headers[canonical]; ok {
			continue
		}
		if input.Headers == nil {
			input.Headers = make(map[string][]string)
		}
		input.Headers[canonical] = append(input.Headers[canonical], value)
	}

	for key, value := range c.Query {
		if _, ok := input.Query[key]; ok {
			continue
		}
		if input.Query == nil {
			input.Query = make(map[string][]string)
		}
		input.Query[key] = append(input.Query[key], value)
	}
}
```

- [ ] **Step 4: Make `Workspace.ApplyDefaults` delegate**

In `internal/core/domain/workspace.go`, replace the entire `ApplyDefaults` method (and remove now-unused imports `net/textproto` and `strings` only if `strings` is unused — `strings` is still used by `ResolveURL`, so keep it; `net/textproto` is no longer needed here):

```go
// ApplyDefaults merges the workspace's configured default headers and query
// parameters into input. Per-request values always win.
func (w *Workspace) ApplyDefaults(input *payload.HttpExecuteInput) {
	w.Config.ApplyDefaults(input)
}
```

Then fix imports in `workspace.go`: the file should import only `strings` and `gon/internal/core/payload` (remove `net/textproto`).

- [ ] **Step 5: Run all domain tests**

Run: `go test ./internal/core/domain/ -v`
Expected: PASS — both `TestConfigApplyDefaults` and the existing `TestWorkspaceApplyDefaults` / `TestWorkspaceResolveURL` pass.

- [ ] **Step 6: Commit**

```bash
git add internal/core/domain/config.go internal/core/domain/workspace.go internal/core/domain/config_test.go
git commit -m "refactor: extract Config.ApplyDefaults for reuse by collections"
```

---

## Task 2: `Request` domain entity — body encoding + `ToInput`

**Files:**
- Create: `internal/core/domain/request.go`
- Test: `internal/core/domain/request_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/core/domain/request_test.go`:

```go
package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestBodyEncode(t *testing.T) {
	t.Run("none yields no body", func(t *testing.T) {
		data, ct, err := RequestBody{Kind: BodyNone}.Encode()
		require.NoError(t, err)
		assert.Nil(t, data)
		assert.Equal(t, "", ct)
	})

	t.Run("json marshals and sets content type", func(t *testing.T) {
		data, ct, err := RequestBody{Kind: BodyJSON, JSON: map[string]any{"a": 1}}.Encode()
		require.NoError(t, err)
		assert.JSONEq(t, `{"a":1}`, string(data))
		assert.Equal(t, "application/json", ct)
	})

	t.Run("raw passes through with given content type", func(t *testing.T) {
		data, ct, err := RequestBody{Kind: BodyRaw, Raw: "hello", ContentType: "text/plain"}.Encode()
		require.NoError(t, err)
		assert.Equal(t, "hello", string(data))
		assert.Equal(t, "text/plain", ct)
	})

	t.Run("form url-encodes and sets content type", func(t *testing.T) {
		data, ct, err := RequestBody{Kind: BodyForm, Form: map[string]string{"a": "b"}}.Encode()
		require.NoError(t, err)
		assert.Equal(t, "a=b", string(data))
		assert.Equal(t, "application/x-www-form-urlencoded", ct)
	})
}

func TestRequestToInput(t *testing.T) {
	t.Run("builds input and auto-sets content type", func(t *testing.T) {
		r := Request{
			Method:  "POST",
			URL:     "/login",
			Headers: map[string][]string{"Accept": {"application/json"}},
			Query:   map[string][]string{"remember": {"true"}},
			Body:    RequestBody{Kind: BodyJSON, JSON: map[string]any{"email": "a@b.com"}},
		}

		input, err := r.ToInput()

		require.NoError(t, err)
		assert.Equal(t, "POST", input.Method)
		assert.Equal(t, "/login", input.URL)
		assert.Equal(t, []string{"application/json"}, input.Headers["Accept"])
		assert.Equal(t, []string{"true"}, input.Query["remember"])
		assert.JSONEq(t, `{"email":"a@b.com"}`, string(input.Body))
		assert.Equal(t, []string{"application/json"}, input.Headers["Content-Type"])
	})

	t.Run("does not override an explicit content type header", func(t *testing.T) {
		r := Request{
			Method:  "POST",
			Headers: map[string][]string{"Content-Type": {"application/vnd.api+json"}},
			Body:    RequestBody{Kind: BodyJSON, JSON: map[string]any{"a": 1}},
		}

		input, err := r.ToInput()

		require.NoError(t, err)
		assert.Equal(t, []string{"application/vnd.api+json"}, input.Headers["Content-Type"])
	})

	t.Run("initializes non-nil header and query maps", func(t *testing.T) {
		input, err := Request{Method: "GET", URL: "/ping"}.ToInput()
		require.NoError(t, err)
		assert.NotNil(t, input.Headers)
		assert.NotNil(t, input.Query)
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/domain/ -run "TestRequestBodyEncode|TestRequestToInput" -v`
Expected: FAIL — undefined `RequestBody`, `BodyNone`, `Request`, etc.

- [ ] **Step 3: Create the domain entity**

Create `internal/core/domain/request.go`:

```go
package domain

import (
	"encoding/json"
	"net/textproto"
	"net/url"

	"gon/internal/core/payload"
)

type BodyKind int

const (
	BodyNone BodyKind = iota
	BodyJSON
	BodyRaw
	BodyForm
)

type RequestBody struct {
	Kind        BodyKind
	JSON        any
	Raw         string
	ContentType string
	Form        map[string]string
}

// Encode returns the wire bytes for the body and the content type that should
// be used when the request does not already specify one.
func (b RequestBody) Encode() ([]byte, string, error) {
	switch b.Kind {
	case BodyJSON:
		data, err := json.Marshal(b.JSON)
		if err != nil {
			return nil, "", err
		}
		return data, "application/json", nil
	case BodyRaw:
		return []byte(b.Raw), b.ContentType, nil
	case BodyForm:
		values := url.Values{}
		for key, value := range b.Form {
			values.Set(key, value)
		}
		return []byte(values.Encode()), "application/x-www-form-urlencoded", nil
	default:
		return nil, "", nil
	}
}

type Request struct {
	Name        string
	Description string
	Method      string
	URL         string
	Headers     map[string][]string
	Query       map[string][]string
	Body        RequestBody
}

// ToInput builds an HttpExecuteInput from the request, encoding the body and
// auto-setting Content-Type when the request does not already provide one.
func (r *Request) ToInput() (*payload.HttpExecuteInput, error) {
	data, contentType, err := r.Body.Encode()
	if err != nil {
		return nil, err
	}

	headers := make(map[string][]string, len(r.Headers))
	for key, values := range r.Headers {
		headers[textproto.CanonicalMIMEHeaderKey(key)] = append([]string(nil), values...)
	}
	query := make(map[string][]string, len(r.Query))
	for key, values := range r.Query {
		query[key] = append([]string(nil), values...)
	}

	if contentType != "" {
		if _, ok := headers["Content-Type"]; !ok {
			headers["Content-Type"] = []string{contentType}
		}
	}

	return &payload.HttpExecuteInput{
		Method:  r.Method,
		URL:     r.URL,
		Headers: headers,
		Query:   query,
		Body:    data,
	}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/domain/ -run "TestRequestBodyEncode|TestRequestToInput" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/core/domain/request.go internal/core/domain/request_test.go
git commit -m "feat: add Request domain entity with body encoding"
```

---

## Task 3: `Collection` domain entity

**Files:**
- Create: `internal/core/domain/collection.go`

- [ ] **Step 1: Create the entity**

Create `internal/core/domain/collection.go`:

```go
package domain

// Collection is a folder of requests sharing a Config. The Config defaults are
// inherited by every request in the folder (and, for nested folders, by child
// collections too).
type Collection struct {
	Name        string
	Description string
	Config      Config
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/core/domain/`
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/core/domain/collection.go
git commit -m "feat: add Collection domain entity"
```

---

## Task 4: Serialization models (`RequestModel`, `CollectionModel`)

**Files:**
- Create: `internal/adapter/model/request_model.go`
- Create: `internal/adapter/model/collection_model.go`
- Test: `internal/adapter/model/request_model_test.go`
- Test: `internal/adapter/model/collection_model_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/adapter/model/request_model_test.go`:

```go
package model

import (
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRequestModelToDomain(t *testing.T) {
	t.Run("maps fields, canonicalizes headers, parses json body", func(t *testing.T) {
		raw := []byte(`
name: Login
description: Authenticate a user
method: post
url: /login
headers:
  accept: application/json
query:
  remember: "true"
body:
  json:
    email: a@b.com
`)
		var m RequestModel
		require.NoError(t, yaml.Unmarshal(raw, &m))

		req, err := m.ToDomain()

		require.NoError(t, err)
		assert.Equal(t, "Login", req.Name)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "/login", req.URL)
		assert.Equal(t, []string{"application/json"}, req.Headers["Accept"])
		assert.Equal(t, []string{"true"}, req.Query["remember"])
		assert.Equal(t, domain.BodyJSON, req.Body.Kind)
	})

	t.Run("rejects more than one body kind", func(t *testing.T) {
		var m RequestModel
		require.NoError(t, yaml.Unmarshal([]byte("body:\n  raw: hi\n  form:\n    a: b\n"), &m))

		_, err := m.ToDomain()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "only one body")
	})

	t.Run("no body yields BodyNone", func(t *testing.T) {
		var m RequestModel
		require.NoError(t, yaml.Unmarshal([]byte("method: get\nurl: /ping\n"), &m))

		req, err := m.ToDomain()

		require.NoError(t, err)
		assert.Equal(t, domain.BodyNone, req.Body.Kind)
	})

	t.Run("round-trips through FromDomain", func(t *testing.T) {
		req := domain.Request{
			Name:    "Get User",
			Method:  "GET",
			URL:     "/users/1",
			Headers: map[string][]string{"Accept": {"application/json"}},
		}
		m := NewRequestModelFromDomain(req)
		back, err := m.ToDomain()
		require.NoError(t, err)
		assert.Equal(t, "GET", back.Method)
		assert.Equal(t, []string{"application/json"}, back.Headers["Accept"])
	})
}
```

Create `internal/adapter/model/collection_model_test.go`:

```go
package model

import (
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCollectionModel(t *testing.T) {
	raw := []byte(`
name: Auth
description: Authentication endpoints
config:
  path: /auth
  headers:
    X-Client: gon
`)
	var m CollectionModel
	require.NoError(t, yaml.Unmarshal(raw, &m))

	c := m.ToDomain()

	assert.Equal(t, "Auth", c.Name)
	assert.Equal(t, "/auth", c.Config.Path)
	assert.Equal(t, "gon", c.Config.Headers["X-Client"])

	back := NewCollectionModelFromDomain(*c)
	assert.Equal(t, "Auth", back.Name)
	assert.Equal(t, "/auth", back.Config.Path)
	_ = domain.Collection{}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/adapter/model/ -run "TestRequestModelToDomain|TestCollectionModel" -v`
Expected: FAIL — undefined `RequestModel`, `CollectionModel`.

- [ ] **Step 3: Create `request_model.go`**

Create `internal/adapter/model/request_model.go`:

```go
package model

import (
	"fmt"
	"net/textproto"
	"strings"

	"gon/internal/core/domain"
)

type BodyModel struct {
	JSON        any               `yaml:"json,omitempty"`
	Raw         string            `yaml:"raw,omitempty"`
	ContentType string            `yaml:"contentType,omitempty"`
	Form        map[string]string `yaml:"form,omitempty"`
}

type RequestModel struct {
	Name        string            `yaml:"name,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Method      string            `yaml:"method,omitempty"`
	URL         string            `yaml:"url,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty"`
	Query       map[string]string `yaml:"query,omitempty"`
	Body        *BodyModel        `yaml:"body,omitempty"`
}

func NewRequestModelFromDomain(request domain.Request) *RequestModel {
	headers := make(map[string]string, len(request.Headers))
	for key, values := range request.Headers {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	query := make(map[string]string, len(request.Query))
	for key, values := range request.Query {
		if len(values) > 0 {
			query[key] = values[0]
		}
	}

	m := &RequestModel{
		Name:        request.Name,
		Description: request.Description,
		Method:      request.Method,
		URL:         request.URL,
		Headers:     headers,
		Query:       query,
	}
	switch request.Body.Kind {
	case domain.BodyJSON:
		m.Body = &BodyModel{JSON: request.Body.JSON}
	case domain.BodyRaw:
		m.Body = &BodyModel{Raw: request.Body.Raw, ContentType: request.Body.ContentType}
	case domain.BodyForm:
		m.Body = &BodyModel{Form: request.Body.Form}
	}
	return m
}

func (m *RequestModel) ToDomain() (*domain.Request, error) {
	body, err := m.Body.toDomain()
	if err != nil {
		return nil, err
	}

	headers := make(map[string][]string, len(m.Headers))
	for key, value := range m.Headers {
		canonical := textproto.CanonicalMIMEHeaderKey(key)
		headers[canonical] = []string{value}
	}
	query := make(map[string][]string, len(m.Query))
	for key, value := range m.Query {
		query[key] = []string{value}
	}

	return &domain.Request{
		Name:        m.Name,
		Description: m.Description,
		Method:      strings.ToUpper(m.Method),
		URL:         m.URL,
		Headers:     headers,
		Query:       query,
		Body:        body,
	}, nil
}

func (b *BodyModel) toDomain() (domain.RequestBody, error) {
	if b == nil {
		return domain.RequestBody{Kind: domain.BodyNone}, nil
	}

	set := 0
	if b.JSON != nil {
		set++
	}
	if b.Raw != "" {
		set++
	}
	if len(b.Form) > 0 {
		set++
	}
	if set > 1 {
		return domain.RequestBody{}, fmt.Errorf("invalid body: only one body kind (json, raw, form) may be set")
	}

	switch {
	case b.JSON != nil:
		return domain.RequestBody{Kind: domain.BodyJSON, JSON: b.JSON}, nil
	case b.Raw != "":
		return domain.RequestBody{Kind: domain.BodyRaw, Raw: b.Raw, ContentType: b.ContentType}, nil
	case len(b.Form) > 0:
		return domain.RequestBody{Kind: domain.BodyForm, Form: b.Form}, nil
	default:
		return domain.RequestBody{Kind: domain.BodyNone}, nil
	}
}
```

- [ ] **Step 4: Create `collection_model.go`**

Create `internal/adapter/model/collection_model.go`:

```go
package model

import "gon/internal/core/domain"

type CollectionModel struct {
	Name        string      `yaml:"name,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Config      ConfigModel `yaml:"config,omitempty"`
}

func NewCollectionModelFromDomain(collection domain.Collection) *CollectionModel {
	return &CollectionModel{
		Name:        collection.Name,
		Description: collection.Description,
		Config:      *NewConfigModelFromDomain(collection.Config),
	}
}

func (m *CollectionModel) ToDomain() *domain.Collection {
	return &domain.Collection{
		Name:        m.Name,
		Description: m.Description,
		Config:      m.Config.ToDomain(),
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/adapter/model/ -v`
Expected: PASS (new tests plus existing workspace/config model tests).

- [ ] **Step 6: Commit**

```bash
git add internal/adapter/model/request_model.go internal/adapter/model/collection_model.go internal/adapter/model/request_model_test.go internal/adapter/model/collection_model_test.go
git commit -m "feat: add request and collection serialization models"
```

---

## Task 5: Driven port interfaces

**Files:**
- Create: `internal/core/port/driven/request_repository.go`
- Create: `internal/core/port/driven/collection_repository.go`

- [ ] **Step 1: Create `request_repository.go`**

Create `internal/core/port/driven/request_repository.go`:

```go
package driven

import (
	"context"

	"gon/internal/core/domain"
)

// RequestRepository loads and persists request files. Load also returns the
// chain of collections from the request's own folder up to the project root,
// ordered nearest-first (the request's folder first, the outermost folder last).
type RequestRepository interface {
	Load(ctx context.Context, root string, requestPath string) (*domain.Request, []domain.Collection, error)
	Save(ctx context.Context, root string, requestPath string, request domain.Request) error
	Exists(ctx context.Context, root string, requestPath string) (bool, error)
}
```

- [ ] **Step 2: Create `collection_repository.go`**

Create `internal/core/port/driven/collection_repository.go`:

```go
package driven

import (
	"context"

	"gon/internal/core/domain"
)

type CollectionRepository interface {
	Save(ctx context.Context, root string, collectionPath string, collection domain.Collection) error
	Exists(ctx context.Context, root string, collectionPath string) (bool, error)
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/core/port/driven/`
Expected: no output (success)

- [ ] **Step 4: Commit**

```bash
git add internal/core/port/driven/request_repository.go internal/core/port/driven/collection_repository.go
git commit -m "feat: add request and collection driven ports"
```

---

## Task 6: Filesystem repositories

**Files:**
- Create: `internal/adapter/repository/request_repository.go`
- Create: `internal/adapter/repository/collection_repository.go`
- Test: `internal/adapter/repository/request_repository_test.go`
- Test: `internal/adapter/repository/collection_repository_test.go`

Resolution rules: `requestPath` is relative to `root` and may omit the extension
(`auth/login` → tries `auth/login.yml` then `auth/login.yaml`). Collections are
discovered by walking from the request's directory up to (and including) `root`,
reading `collection.yml` where present, nearest-first.

- [ ] **Step 1: Write the failing tests**

Create `internal/adapter/repository/request_repository_test.go`:

```go
package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}

func TestRequestRepositoryLoad(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "auth", "collection.yml"), "name: Auth\nconfig:\n  path: /auth\n")
	writeFile(t, filepath.Join(root, "auth", "admin", "collection.yml"), "name: Admin\nconfig:\n  path: /admin\n")
	writeFile(t, filepath.Join(root, "auth", "admin", "impersonate.yml"), "method: post\nurl: /impersonate\n")

	repo := NewRequestRepository()
	req, collections, err := repo.Load(context.Background(), root, "auth/admin/impersonate")

	require.NoError(t, err)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "/impersonate", req.URL)
	require.Len(t, collections, 2)
	assert.Equal(t, "Admin", collections[0].Name) // nearest first
	assert.Equal(t, "Auth", collections[1].Name)
}

func TestRequestRepositoryLoadMissing(t *testing.T) {
	repo := NewRequestRepository()
	_, _, err := repo.Load(context.Background(), t.TempDir(), "nope/missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request not found")
}

func TestRequestRepositoryLoadRejectsCollectionFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "auth", "collection.yml"), "name: Auth\n")
	repo := NewRequestRepository()
	_, _, err := repo.Load(context.Background(), root, "auth/collection")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reserved")
}

func TestRequestRepositorySaveAndExists(t *testing.T) {
	root := t.TempDir()
	repo := NewRequestRepository()

	exists, err := repo.Exists(context.Background(), root, "auth/login")
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, os.MkdirAll(filepath.Join(root, "auth"), 0755))
	err = repo.Save(context.Background(), root, "auth/login", domainGetRequest())
	require.NoError(t, err)

	exists, err = repo.Exists(context.Background(), root, "auth/login")
	require.NoError(t, err)
	assert.True(t, exists)

	_, err = os.Stat(filepath.Join(root, "auth", "login.yml"))
	require.NoError(t, err)
}
```

Add this helper at the bottom of `internal/adapter/repository/request_repository_test.go`:

```go
func domainGetRequest() domain.Request {
	return domain.Request{Method: "GET", URL: "/login"}
}
```

And add the import for `domain` at the top of that test file:

```go
import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
```

Create `internal/adapter/repository/collection_repository_test.go`:

```go
package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectionRepositorySaveAndExists(t *testing.T) {
	root := t.TempDir()
	repo := NewCollectionRepository()

	exists, err := repo.Exists(context.Background(), root, "auth")
	require.NoError(t, err)
	assert.False(t, exists)

	err = repo.Save(context.Background(), root, "auth", domain.Collection{Name: "auth"})
	require.NoError(t, err)

	exists, err = repo.Exists(context.Background(), root, "auth")
	require.NoError(t, err)
	assert.True(t, exists)

	_, err = os.Stat(filepath.Join(root, "auth", "collection.yml"))
	require.NoError(t, err)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/adapter/repository/ -run "TestRequestRepository|TestCollectionRepository" -v`
Expected: FAIL — undefined `NewRequestRepository`, `NewCollectionRepository`.

- [ ] **Step 3: Create `request_repository.go`**

Create `internal/adapter/repository/request_repository.go`:

```go
package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gon/internal/adapter/model"
	"gon/internal/core/domain"
	"gon/internal/core/port/driven"

	"gopkg.in/yaml.v3"
)

const collectionFileName = "collection.yml"

type requestRepository struct{}

func NewRequestRepository() driven.RequestRepository {
	return &requestRepository{}
}

// resolveFile returns the on-disk path of a request file, trying .yml then
// .yaml. It returns ok=false when neither exists.
func resolveFile(root, requestPath string) (string, bool) {
	clean := filepath.Clean(requestPath)
	for _, ext := range []string{".yml", ".yaml"} {
		candidate := filepath.Join(root, clean) + ext
		if hasExtension(clean) {
			candidate = filepath.Join(root, clean)
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
		if hasExtension(clean) {
			break
		}
	}
	return "", false
}

func hasExtension(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".yml" || ext == ".yaml"
}

func (r *requestRepository) Load(ctx context.Context, root string, requestPath string) (*domain.Request, []domain.Collection, error) {
	clean := filepath.Clean(requestPath)
	base := strings.TrimSuffix(filepath.Base(clean), filepath.Ext(clean))
	if base == "collection" {
		return nil, nil, fmt.Errorf("%q is a reserved collection file, not a request", requestPath)
	}

	file, ok := resolveFile(root, requestPath)
	if !ok {
		return nil, nil, fmt.Errorf("request not found: %s", requestPath)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, err
	}
	var requestModel model.RequestModel
	if err := yaml.Unmarshal(data, &requestModel); err != nil {
		return nil, nil, fmt.Errorf("error parsing %s: %w", file, err)
	}
	request, err := requestModel.ToDomain()
	if err != nil {
		return nil, nil, fmt.Errorf("error in %s: %w", file, err)
	}

	collections, err := loadCollectionChain(root, filepath.Dir(file))
	if err != nil {
		return nil, nil, err
	}
	return request, collections, nil
}

// loadCollectionChain walks from dir up to and including root, collecting any
// collection.yml found, nearest-first.
func loadCollectionChain(root, dir string) ([]domain.Collection, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	current, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	var collections []domain.Collection
	for {
		file := filepath.Join(current, collectionFileName)
		if data, err := os.ReadFile(file); err == nil {
			var collectionModel model.CollectionModel
			if err := yaml.Unmarshal(data, &collectionModel); err != nil {
				return nil, fmt.Errorf("error parsing %s: %w", file, err)
			}
			collections = append(collections, *collectionModel.ToDomain())
		}
		if current == absRoot {
			break
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return collections, nil
}

func (r *requestRepository) Save(ctx context.Context, root string, requestPath string, request domain.Request) error {
	clean := filepath.Clean(requestPath)
	if !hasExtension(clean) {
		clean += ".yml"
	}
	target := filepath.Join(root, clean)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(model.NewRequestModelFromDomain(request))
	if err != nil {
		return err
	}
	return os.WriteFile(target, data, 0644)
}

func (r *requestRepository) Exists(ctx context.Context, root string, requestPath string) (bool, error) {
	_, ok := resolveFile(root, requestPath)
	return ok, nil
}
```

- [ ] **Step 4: Create `collection_repository.go`**

Create `internal/adapter/repository/collection_repository.go`:

```go
package repository

import (
	"context"
	"os"
	"path/filepath"

	"gon/internal/adapter/model"
	"gon/internal/core/domain"
	"gon/internal/core/port/driven"

	"gopkg.in/yaml.v3"
)

type collectionRepository struct{}

func NewCollectionRepository() driven.CollectionRepository {
	return &collectionRepository{}
}

func (r *collectionRepository) Save(ctx context.Context, root string, collectionPath string, collection domain.Collection) error {
	dir := filepath.Join(root, filepath.Clean(collectionPath))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(model.NewCollectionModelFromDomain(collection))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, collectionFileName), data, 0644)
}

func (r *collectionRepository) Exists(ctx context.Context, root string, collectionPath string) (bool, error) {
	file := filepath.Join(root, filepath.Clean(collectionPath), collectionFileName)
	if _, err := os.Stat(file); err == nil {
		return true, nil
	}
	return false, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/adapter/repository/ -v`
Expected: PASS (new tests plus existing workspace repository tests).

- [ ] **Step 6: Commit**

```bash
git add internal/adapter/repository/request_repository.go internal/adapter/repository/collection_repository.go internal/adapter/repository/request_repository_test.go internal/adapter/repository/collection_repository_test.go
git commit -m "feat: add filesystem request and collection repositories"
```

---

## Task 7: Driving port interfaces

**Files:**
- Create: `internal/core/port/driving/request_service.go`
- Create: `internal/core/port/driving/collection_service.go`

- [ ] **Step 1: Create `request_service.go`**

Create `internal/core/port/driving/request_service.go`:

```go
package driving

import (
	"context"

	"gon/internal/core/payload"
)

type RequestService interface {
	// Run executes a saved request. overrides (may be nil) carries per-execution
	// header/query/body values that take precedence over the request file.
	Run(ctx context.Context, root string, requestPath string, overrides *payload.HttpExecuteInput) (*payload.HttpExecuteOutput, error)
	// Create scaffolds a new request file with the given method.
	Create(ctx context.Context, root string, requestPath string, method string) error
}
```

- [ ] **Step 2: Create `collection_service.go`**

Create `internal/core/port/driving/collection_service.go`:

```go
package driving

import "context"

type CollectionService interface {
	// Create scaffolds a collection (and any missing ancestor collections).
	Create(ctx context.Context, root string, collectionPath string) error
}
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./internal/core/port/driving/`
Expected: no output (success)

- [ ] **Step 4: Commit**

```bash
git add internal/core/port/driving/request_service.go internal/core/port/driving/collection_service.go
git commit -m "feat: add request and collection driving ports"
```

---

## Task 8: `RequestService` — merge precedence + delegate

**Files:**
- Create: `internal/core/service/request_service.go`
- Test: `internal/core/service/request_service_test.go`

The service builds the input from the request file, lets `overrides` replace
those values, applies the collection chain (nearest→root, additive), prefixes
the collection paths onto a relative URL, then delegates to `HttpService`
(which applies the workspace defaults and resolves the base URL).

- [ ] **Step 1: Write the failing test**

Create `internal/core/service/request_service_test.go`:

```go
package service

import (
	"context"
	"testing"

	"gon/internal/core/domain"
	"gon/internal/core/payload"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRequestRepo returns canned request + collections and records nothing else.
type fakeRequestRepo struct {
	request     domain.Request
	collections []domain.Collection
}

func (f *fakeRequestRepo) Load(ctx context.Context, root, requestPath string) (*domain.Request, []domain.Collection, error) {
	r := f.request
	return &r, f.collections, nil
}
func (f *fakeRequestRepo) Save(ctx context.Context, root, requestPath string, request domain.Request) error {
	return nil
}
func (f *fakeRequestRepo) Exists(ctx context.Context, root, requestPath string) (bool, error) {
	return false, nil
}

// captureHttpService records the input it was given and returns an empty output.
type captureHttpService struct{ input *payload.HttpExecuteInput }

func (c *captureHttpService) Execute(ctx context.Context, input *payload.HttpExecuteInput) (*payload.HttpExecuteOutput, error) {
	c.input = input
	return &payload.HttpExecuteOutput{StatusCode: 200}, nil
}

func TestRequestServiceRun(t *testing.T) {
	t.Run("merges request, collections, and prefixes paths; overrides win", func(t *testing.T) {
		repo := &fakeRequestRepo{
			request: domain.Request{
				Method:  "POST",
				URL:     "/login",
				Headers: map[string][]string{"Accept": {"application/json"}},
			},
			collections: []domain.Collection{
				{Name: "Admin", Config: domain.Config{Path: "/admin", Headers: map[string]string{"X-Inner": "1"}}},
				{Name: "Auth", Config: domain.Config{Path: "/auth", Headers: map[string]string{"X-Outer": "1", "Accept": "text/plain"}}},
			},
		}
		http := &captureHttpService{}
		svc := NewRequestService(repo, nil, http)

		overrides := &payload.HttpExecuteInput{Headers: map[string][]string{"X-Inner": {"override"}}}
		_, err := svc.Run(context.Background(), "/root", "auth/admin/impersonate", overrides)

		require.NoError(t, err)
		got := http.input
		// path: outermost (/auth) then nearest (/admin) then request url
		assert.Equal(t, "/auth/admin/login", got.URL)
		// override beats collection default
		assert.Equal(t, []string{"override"}, got.Headers["X-Inner"])
		// request header beats collection default for the same key
		assert.Equal(t, []string{"application/json"}, got.Headers["Accept"])
		// outer collection default still applied
		assert.Equal(t, []string{"1"}, got.Headers["X-Outer"])
	})

	t.Run("absolute url bypasses collection path prefixing", func(t *testing.T) {
		repo := &fakeRequestRepo{
			request:     domain.Request{Method: "GET", URL: "https://other.com/x"},
			collections: []domain.Collection{{Config: domain.Config{Path: "/auth"}}},
		}
		http := &captureHttpService{}
		svc := NewRequestService(repo, nil, http)

		_, err := svc.Run(context.Background(), "/root", "auth/x", nil)

		require.NoError(t, err)
		assert.Equal(t, "https://other.com/x", http.input.URL)
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/service/ -run TestRequestServiceRun -v`
Expected: FAIL — undefined `NewRequestService`.

- [ ] **Step 3: Create `request_service.go`**

Create `internal/core/service/request_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"path"
	"strings"

	"gon/internal/core/domain"
	"gon/internal/core/payload"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"

	"github.com/iancoleman/strcase"
)

type requestService struct {
	requestRepository    driven.RequestRepository
	collectionRepository driven.CollectionRepository
	httpService          driving.HttpService
}

func NewRequestService(
	requestRepository driven.RequestRepository,
	collectionRepository driven.CollectionRepository,
	httpService driving.HttpService,
) driving.RequestService {
	return &requestService{
		requestRepository:    requestRepository,
		collectionRepository: collectionRepository,
		httpService:          httpService,
	}
}

func (s *requestService) Run(ctx context.Context, root string, requestPath string, overrides *payload.HttpExecuteInput) (*payload.HttpExecuteOutput, error) {
	request, collections, err := s.requestRepository.Load(ctx, root, requestPath)
	if err != nil {
		return nil, err
	}

	input, err := request.ToInput()
	if err != nil {
		return nil, err
	}

	applyOverrides(input, overrides)

	// Collection defaults: nearest-first so inner collections win (additive).
	for i := range collections {
		collections[i].Config.ApplyDefaults(input)
	}

	input.URL = prefixCollectionPaths(input.URL, collections)

	return s.httpService.Execute(ctx, input)
}

// applyOverrides copies per-execution values onto input, replacing existing
// keys so the override always wins.
func applyOverrides(input *payload.HttpExecuteInput, overrides *payload.HttpExecuteInput) {
	if overrides == nil {
		return
	}
	if input.Headers == nil {
		input.Headers = make(map[string][]string)
	}
	for key, values := range overrides.Headers {
		input.Headers[key] = values
	}
	if input.Query == nil {
		input.Query = make(map[string][]string)
	}
	for key, values := range overrides.Query {
		input.Query[key] = values
	}
	if overrides.Body != nil {
		input.Body = overrides.Body
	}
}

// prefixCollectionPaths prepends each collection's configured path, outermost
// first, to a relative URL. Absolute URLs are returned unchanged.
func prefixCollectionPaths(url string, collections []domain.Collection) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	var prefix string
	// collections is nearest-first; iterate in reverse for outermost-first.
	for i := len(collections) - 1; i >= 0; i-- {
		prefix += collections[i].Config.Path
	}
	return prefix + url
}

func (s *requestService) Create(ctx context.Context, root string, requestPath string, method string) error {
	exists, err := s.requestRepository.Exists(ctx, root, requestPath)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("request already exists: %s", requestPath)
	}

	// Ensure the parent folder is a collection.
	parent := path.Dir(filepath_ToSlash(requestPath))
	if parent != "." && parent != "" {
		ok, err := s.collectionRepository.Exists(ctx, root, parent)
		if err != nil {
			return err
		}
		if !ok {
			name := strcase.ToKebab(path.Base(parent))
			if err := s.collectionRepository.Save(ctx, root, parent, domain.Collection{Name: name}); err != nil {
				return err
			}
		}
	}

	name := strcase.ToKebab(strings.TrimSuffix(path.Base(filepath_ToSlash(requestPath)), ".yml"))
	request := domain.Request{
		Name:   name,
		Method: strings.ToUpper(method),
		URL:    "/",
	}
	return s.requestRepository.Save(ctx, root, requestPath, request)
}

// filepath_ToSlash normalizes OS separators to forward slashes so path.Dir /
// path.Base behave consistently regardless of platform.
func filepath_ToSlash(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/service/ -run TestRequestServiceRun -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/core/service/request_service.go internal/core/service/request_service_test.go
git commit -m "feat: add RequestService with inherited config merge"
```

---

## Task 9: `CollectionService` + `RequestService.Create` scaffolding test

**Files:**
- Create: `internal/core/service/collection_service.go`
- Test: `internal/core/service/collection_service_test.go`
- Test: `internal/core/service/request_service_test.go` (add a scaffolding test)

- [ ] **Step 1: Write the failing tests**

Create `internal/core/service/collection_service_test.go`:

```go
package service

import (
	"context"
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingCollectionRepo records Save calls and answers Exists from a set.
type recordingCollectionRepo struct {
	existing map[string]bool
	saved    []string
}

func (r *recordingCollectionRepo) Save(ctx context.Context, root, collectionPath string, c domain.Collection) error {
	r.saved = append(r.saved, collectionPath)
	if r.existing == nil {
		r.existing = map[string]bool{}
	}
	r.existing[collectionPath] = true
	return nil
}
func (r *recordingCollectionRepo) Exists(ctx context.Context, root, collectionPath string) (bool, error) {
	return r.existing[collectionPath], nil
}

func TestCollectionServiceCreate(t *testing.T) {
	t.Run("creates nested collections including ancestors", func(t *testing.T) {
		repo := &recordingCollectionRepo{}
		svc := NewCollectionService(repo)

		err := svc.Create(context.Background(), "/root", "auth/admin")

		require.NoError(t, err)
		assert.Equal(t, []string{"auth", "auth/admin"}, repo.saved)
	})

	t.Run("errors when target already exists", func(t *testing.T) {
		repo := &recordingCollectionRepo{existing: map[string]bool{"auth": true}}
		svc := NewCollectionService(repo)

		err := svc.Create(context.Background(), "/root", "auth")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}
```

Append a scaffolding test to `internal/core/service/request_service_test.go` (reusing `recordingCollectionRepo` from the file above, which is in the same package):

```go
func TestRequestServiceCreate(t *testing.T) {
	repo := &fakeRequestRepo{}
	collections := &recordingCollectionRepo{}
	svc := NewRequestService(repo, collections, nil)

	err := svc.Create(context.Background(), "/root", "auth/login", "post")

	require.NoError(t, err)
	// parent collection auth had to be created
	assert.Equal(t, []string{"auth"}, collections.saved)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/core/service/ -run "TestCollectionServiceCreate|TestRequestServiceCreate" -v`
Expected: FAIL — undefined `NewCollectionService`.

- [ ] **Step 3: Create `collection_service.go`**

Create `internal/core/service/collection_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"path"
	"strings"

	"gon/internal/core/domain"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"

	"github.com/iancoleman/strcase"
)

type collectionService struct {
	collectionRepository driven.CollectionRepository
}

func NewCollectionService(collectionRepository driven.CollectionRepository) driving.CollectionService {
	return &collectionService{collectionRepository: collectionRepository}
}

func (s *collectionService) Create(ctx context.Context, root string, collectionPath string) error {
	normalized := strings.Trim(filepath_ToSlash(collectionPath), "/")
	if normalized == "" {
		return fmt.Errorf("collection path is required")
	}

	segments := strings.Split(normalized, "/")
	for i := range segments {
		sub := strings.Join(segments[:i+1], "/")
		isTarget := i == len(segments)-1

		exists, err := s.collectionRepository.Exists(ctx, root, sub)
		if err != nil {
			return err
		}
		if exists {
			if isTarget {
				return fmt.Errorf("collection already exists: %s", sub)
			}
			continue
		}

		name := strcase.ToKebab(path.Base(sub))
		if err := s.collectionRepository.Save(ctx, root, sub, domain.Collection{Name: name}); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/core/service/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/core/service/collection_service.go internal/core/service/collection_service_test.go internal/core/service/request_service_test.go
git commit -m "feat: add CollectionService and request scaffolding"
```

---

## Task 10: CLI commands (`run`, `collection init`, `request new`)

**Files:**
- Create: `internal/adapter/command/run_command.go`
- Create: `internal/adapter/command/collection_init_command.go`
- Create: `internal/adapter/command/request_new_command.go`

The `run` command reuses the existing `parseHeaders`, `parseQuery`, and
`resolveMode` helpers in `internal/adapter/command/http_command.go` (same
package).

- [ ] **Step 1: Create `run_command.go`**

Create `internal/adapter/command/run_command.go`:

```go
package command

import (
	"context"
	"os"
	"time"

	"gon/internal/core/payload"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"

	"github.com/urfave/cli/v3"
)

func RunCommand(requestService driving.RequestService, httpOutput driven.HttpOutput) *cli.Command {
	return &cli.Command{
		Name:      "run",
		Usage:     "Run a saved request by path (e.g. run auth/login)",
		ArgsUsage: "<path>",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "path", UsageText: "Path to the saved request, e.g. auth/login"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			headers, err := parseHeaders(cmd.StringSlice("header"))
			if err != nil {
				return err
			}
			query, err := parseQuery(cmd.StringSlice("query"))
			if err != nil {
				return err
			}

			overrides := &payload.HttpExecuteInput{Headers: headers, Query: query}
			if cmd.String("json") != "" {
				overrides.Body = []byte(cmd.String("json"))
				if _, ok := headers["Content-Type"]; !ok {
					headers["Content-Type"] = append(headers["Content-Type"], "application/json")
				}
			}

			if d := cmd.Duration("timeout"); d > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, d)
				defer cancel()
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			result, err := requestService.Run(ctx, cwd, cmd.StringArg("path"), overrides)
			if err != nil {
				return err
			}
			httpOutput.Format(overrides, result, resolveMode(cmd))
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "json", Value: "", Usage: "JSON body override for the request"},
			&cli.StringSliceFlag{Name: "header", Usage: `HTTP header override in "Key: Value" format, can be repeated`},
			&cli.StringSliceFlag{Name: "query", Usage: `HTTP query override in "Key=Value" format, can be repeated`},
			&cli.BoolFlag{Name: "minimal", Usage: "Minimal output, only print status code and headers"},
			&cli.BoolFlag{Name: "normal", Usage: "Normal output, print status code, headers, and body"},
			&cli.BoolFlag{Name: "full", Usage: "Full output, print request and response details"},
			&cli.DurationFlag{Name: "timeout", Value: 30 * time.Second, Usage: "Request timeout duration"},
		},
	}
}
```

> Note: `httpOutput.Format` is passed `overrides` as the echoed input. The
> merged request is not echoed because the merge happens inside the service; if
> a fully-merged `--full` echo is wanted later, the service can be extended to
> return the resolved input. This is acceptable for the first version.

- [ ] **Step 2: Create `collection_init_command.go`**

Create `internal/adapter/command/collection_init_command.go`:

```go
package command

import (
	"context"
	"fmt"
	"os"

	"gon/internal/core/port/driving"

	"github.com/urfave/cli/v3"
)

func CollectionInitCommand(collectionService driving.CollectionService) *cli.Command {
	return &cli.Command{
		Name:      "collection",
		Usage:     "Manage collections",
		ArgsUsage: "init <name>",
		Commands: []*cli.Command{
			{
				Name:      "init",
				Usage:     "Create a new collection folder (nesting allowed, e.g. auth/admin)",
				ArgsUsage: "<name>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name", UsageText: "Collection path, e.g. auth or auth/admin"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					name := cmd.StringArg("name")
					if name == "" {
						return fmt.Errorf("collection name is required")
					}
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					if err := collectionService.Create(ctx, cwd, name); err != nil {
						return err
					}
					fmt.Printf("Collection '%s' created\n", name)
					return nil
				},
			},
		},
	}
}
```

- [ ] **Step 3: Create `request_new_command.go`**

Create `internal/adapter/command/request_new_command.go`:

```go
package command

import (
	"context"
	"fmt"
	"os"

	"gon/internal/core/port/driving"

	"github.com/urfave/cli/v3"
)

func RequestNewCommand(requestService driving.RequestService) *cli.Command {
	return &cli.Command{
		Name:      "request",
		Usage:     "Manage saved requests",
		ArgsUsage: "new <path>",
		Commands: []*cli.Command{
			{
				Name:      "new",
				Usage:     "Scaffold a new request file (e.g. request new auth/login --method POST)",
				ArgsUsage: "<path>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "path", UsageText: "Request path, e.g. auth/login"},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "method", Value: "GET", Usage: "HTTP method for the new request"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					path := cmd.StringArg("path")
					if path == "" {
						return fmt.Errorf("request path is required")
					}
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					if err := requestService.Create(ctx, cwd, path, cmd.String("method")); err != nil {
						return err
					}
					fmt.Printf("Request '%s' created\n", path)
					return nil
				},
			},
		},
	}
}
```

- [ ] **Step 4: Verify it compiles**

Run: `go build ./internal/adapter/command/`
Expected: no output (success)

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/command/run_command.go internal/adapter/command/collection_init_command.go internal/adapter/command/request_new_command.go
git commit -m "feat: add run, collection init, and request new commands"
```

---

## Task 11: Wire commands into `cmd/main.go`

**Files:**
- Modify: `cmd/main.go:22-57` (inside `cli_app`)

- [ ] **Step 1: Construct the new repos and services**

In `cmd/main.go`, inside `cli_app`, after the existing `workspaceService` line, add:

```go
	requestRepository := repository.NewRequestRepository()
	collectionRepository := repository.NewCollectionRepository()
	requestService := service.NewRequestService(requestRepository, collectionRepository, httpService)
	collectionService := service.NewCollectionService(collectionRepository)
```

- [ ] **Step 2: Add a Collections command group**

In `cmd/main.go`, after the `workspaceCommands` slice declaration, add:

```go
	collectionCommands := []*cli.Command{
		command.RunCommand(requestService, httpOutput),
		command.CollectionInitCommand(collectionService),
		command.RequestNewCommand(requestService),
	}
```

Then add the group to the `groups` slice (insert before `{Name: "Common", ...}`):

```go
	groups := []command.CommandGroup{
		{Name: "HTTP Commands", Commands: httpCommands},
		{Name: "Workspace", Commands: workspaceCommands},
		{Name: "Collections", Commands: collectionCommands},
		{Name: "Common", Commands: utilityCommands},
	}
```

Because a group was inserted, update the help-command index: the line
`groups[2].Commands = utilityCommands` must become `groups[3].Commands = utilityCommands`.

Finally, include the new commands in the final `commands` slice. Replace:

```go
	commands := append(httpCommands, workspaceCommands...)
	commands = append(commands, utilityCommands...)
```

with:

```go
	commands := append(httpCommands, workspaceCommands...)
	commands = append(commands, collectionCommands...)
	commands = append(commands, utilityCommands...)
```

- [ ] **Step 3: Build and verify the whole project**

Run: `go build -o gon ./cmd && go test ./...`
Expected: build succeeds; all tests PASS.

- [ ] **Step 4: Manual smoke test**

Run:

```bash
cd "$(mktemp -d)" && go run -C /home/fakhrulnugroho/work/opensource/gon ./cmd collection init auth \
  && go run -C /home/fakhrulnugroho/work/opensource/gon ./cmd request new auth/login --method POST \
  && cat auth/collection.yml auth/login.yml
```

Expected: `auth/collection.yml` and `auth/login.yml` are created and printed.
(Note: `run auth/login` needs a real workspace + reachable server, so the smoke
test covers scaffolding only.)

- [ ] **Step 5: Commit**

```bash
git add cmd/main.go
git commit -m "feat: wire collections and requests commands into REPL and CLI"
```

---

## Self-Review Notes

- **Spec coverage:** file layout (Task 6 resolution), `collection.yml` schema (Task 4), request schema + json/raw/form (Tasks 2, 4), precedence/merge (Tasks 1, 8), URL prefixing + absolute bypass (Task 8), `run` command (Tasks 10–11), scaffolding `collection init` + `request new` with auto-created parent (Tasks 9–11), hexagonal mapping (Tasks 2–11), validation/errors (Tasks 4, 6, 8, 9). The `{{var}}` substitution and tests/assertions are intentionally deferred per the spec.
- **Type consistency:** `RequestService.Run/Create`, `CollectionService.Create`, `RequestRepository.Load/Save/Exists`, `CollectionRepository.Save/Exists`, `Config.ApplyDefaults`, `Request.ToInput`, `RequestBody.Encode`, and `BodyKind` constants are used with identical signatures across tasks.
- **Known simplification:** `run --full` echoes the override input, not the fully merged request (documented inline in Task 10). Revisit only if needed.
