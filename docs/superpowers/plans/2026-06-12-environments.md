# Environments Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Postman-style, project-scoped environments (`local`/`dev`/`test`/`prod`) to a `gon` workspace, where each environment owns a `base_url` plus named variables that requests reference via `{{var}}` substitution.

**Architecture:** Hexagonal (ports & adapters). A new `domain.Environment` carries `base_url` + variables and does `{{var}}` substitution. A new `EnvironmentRepository` (driven) persists `environments/<name>.yml` files and a gitignored `.gon/active-env` state file. A new `EnvironmentService` (driving) creates/lists/activates environments and resolves the active one with precedence `--env flag > .gon/active-env > single-env fallback > error`. The command layer resolves the active environment and passes `*domain.Environment` into `HttpService.Execute` / `RequestService.Run`, which apply substitution and fail fast on any unresolved `{{var}}`.

**Tech Stack:** Go, `gopkg.in/yaml.v3`, `github.com/urfave/cli/v3`, `github.com/iancoleman/strcase`, `github.com/stretchr/testify`.

---

## File Structure

**New files:**
- `internal/core/domain/environment.go` — `Environment` struct, `Substitute`, `FindPlaceholders`; plus a `ResolveURL` free function.
- `internal/core/domain/environment_test.go`
- `internal/core/port/driven/environment_repository.go` — `EnvironmentRepository` interface.
- `internal/core/port/driving/environment_service.go` — `EnvironmentService` interface.
- `internal/core/service/environment_service.go` — implements `EnvironmentService`.
- `internal/core/service/environment_service_test.go`
- `internal/adapter/model/environment_model.go` — `EnvironmentModel` ↔ domain.
- `internal/adapter/model/environment_model_test.go`
- `internal/adapter/repository/environment_repository.go` — YAML + active-state persistence.
- `internal/adapter/repository/environment_repository_test.go`
- `internal/adapter/command/env_command.go` — `env new/list/use` CLI.

**Modified files:**
- `internal/core/domain/workspace.go` — `BaseURL` documented as deprecated fallback; `ResolveURL` delegates to the new free function.
- `internal/core/port/driving/http_service.go` — `Execute` gains `env` param.
- `internal/core/port/driving/request_service.go` — `Run` gains `env` param.
- `internal/core/service/http_service.go` — substitution + fail-fast + `env` param.
- `internal/core/service/http_service_test.go` — update call sites, add substitution tests.
- `internal/core/service/request_service.go` — thread `env` into `Execute`.
- `internal/core/service/request_service_test.go` — update call sites + fake.
- `internal/core/service/workspace_service.go` — inject `EnvironmentRepository`; scaffold `environments/local.yml` + active on init; stop writing `base_url`.
- `internal/core/service/workspace_service_test.go` — update constructor + assertions.
- `internal/adapter/command/http_command.go` — `--env` flag; resolve + pass env.
- `internal/adapter/command/run_command.go` — `--env` flag; resolve + pass env.
- `cmd/main.go` — wire environment repo/service; REPL prompt reflects active env.
- `README.md`, `CLAUDE.md` — document environments.

---

## Task 1: Domain — `ResolveURL` free function

Extract URL resolution into a reusable free function so the HTTP service can resolve against an environment's `base_url` (not just the workspace's).

**Files:**
- Modify: `internal/core/domain/workspace.go`
- Test: `internal/core/domain/workspace_test.go` (existing tests must keep passing)

- [ ] **Step 1: Refactor `workspace.go` to add the free function and delegate**

Replace the body of `internal/core/domain/workspace.go` with:

```go
package domain

import (
	"strings"

	"gon/internal/core/payload"
)

type Workspace struct {
	Name string
	// BaseURL is a deprecated fallback. The active Environment's base_url is the
	// source of truth; BaseURL is only consulted when no environment supplies one
	// (e.g. older workspaces, or when no environments exist yet).
	BaseURL string
	Config  Config
}

// ResolveURL builds the absolute request URL from a base URL, a config path, and
// the request path. An absolute http(s) request path bypasses resolution.
func ResolveURL(baseURL, configPath, requestPath string) string {
	if strings.HasPrefix(requestPath, "http://") || strings.HasPrefix(requestPath, "https://") {
		return requestPath
	}
	return baseURL + configPath + requestPath
}

func (w *Workspace) ResolveURL(path string) string {
	return ResolveURL(w.BaseURL, w.Config.Path, path)
}

// ApplyDefaults merges the workspace's configured default headers and query
// parameters into input. Per-request values always win.
func (w *Workspace) ApplyDefaults(input *payload.HttpExecuteInput) {
	w.Config.ApplyDefaults(input)
}
```

- [ ] **Step 2: Run the existing domain tests to verify they still pass**

Run: `go test ./internal/core/domain/ -run TestWorkspaceResolveURL -v`
Expected: PASS (the method now delegates; behavior is unchanged).

- [ ] **Step 3: Commit**

```bash
git add internal/core/domain/workspace.go
git commit -m "refactor: extract ResolveURL free function in domain"
```

---

## Task 2: Domain — `Environment` with substitution

**Files:**
- Create: `internal/core/domain/environment.go`
- Test: `internal/core/domain/environment_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/core/domain/environment_test.go`:

```go
package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentSubstitute(t *testing.T) {
	env := &Environment{
		Name:      "dev",
		Variables: map[string]string{"token": "abc123", "user_id": "42"},
	}

	t.Run("replaces a known placeholder", func(t *testing.T) {
		assert.Equal(t, "Bearer abc123", env.Substitute("Bearer {{token}}"))
	})

	t.Run("tolerates inner whitespace", func(t *testing.T) {
		assert.Equal(t, "abc123", env.Substitute("{{ token }}"))
	})

	t.Run("replaces multiple placeholders", func(t *testing.T) {
		assert.Equal(t, "abc123/42", env.Substitute("{{token}}/{{user_id}}"))
	})

	t.Run("leaves unknown placeholders intact", func(t *testing.T) {
		assert.Equal(t, "{{missing}}", env.Substitute("{{missing}}"))
	})

	t.Run("nil environment returns input unchanged", func(t *testing.T) {
		var nilEnv *Environment
		assert.Equal(t, "{{token}}", nilEnv.Substitute("{{token}}"))
	})
}

func TestFindPlaceholders(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, FindPlaceholders("{{a}}/{{ b }}"))
	assert.Empty(t, FindPlaceholders("no placeholders here"))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/domain/ -run 'TestEnvironmentSubstitute|TestFindPlaceholders' -v`
Expected: FAIL — `undefined: Environment` / `undefined: FindPlaceholders`.

- [ ] **Step 3: Write the implementation**

Create `internal/core/domain/environment.go`:

```go
package domain

import "regexp"

// Environment is a named, project-scoped set of variables plus a base URL.
// Requests reference its variables via {{name}} substitution.
type Environment struct {
	Name      string
	BaseURL   string
	Variables map[string]string
}

// placeholderPattern matches {{name}} with optional inner whitespace. Names may
// contain letters, digits, underscore, dot, and dash.
var placeholderPattern = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_.-]+)\s*\}\}`)

// Substitute replaces {{name}} placeholders in s with this environment's
// variables in a single pass. Unknown placeholders are left intact so callers
// can detect them.
func (e *Environment) Substitute(s string) string {
	if e == nil {
		return s
	}
	return placeholderPattern.ReplaceAllStringFunc(s, func(match string) string {
		name := placeholderPattern.FindStringSubmatch(match)[1]
		if v, ok := e.Variables[name]; ok {
			return v
		}
		return match
	})
}

// FindPlaceholders returns the variable names of every {{name}} placeholder in s,
// in order of appearance (duplicates included).
func FindPlaceholders(s string) []string {
	matches := placeholderPattern.FindAllStringSubmatch(s, -1)
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, m[1])
	}
	return names
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/domain/ -run 'TestEnvironmentSubstitute|TestFindPlaceholders' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/domain/environment.go internal/core/domain/environment_test.go
git commit -m "feat: add Environment domain entity with variable substitution"
```

---

## Task 3: Driven port — `EnvironmentRepository` interface

**Files:**
- Create: `internal/core/port/driven/environment_repository.go`

- [ ] **Step 1: Write the interface**

Create `internal/core/port/driven/environment_repository.go`:

```go
package driven

import (
	"context"

	"gon/internal/core/domain"
)

// EnvironmentRepository persists environment definition files
// (environments/<name>.yml) and the locally-selected active environment
// (.gon/active-env, gitignored).
type EnvironmentRepository interface {
	Save(ctx context.Context, root string, environment domain.Environment) error
	Load(ctx context.Context, root string, name string) (*domain.Environment, error)
	List(ctx context.Context, root string) ([]string, error)
	Exists(ctx context.Context, root string, name string) (bool, error)
	// ReadActive returns the active environment name, or "" if none is set.
	ReadActive(ctx context.Context, root string) (string, error)
	WriteActive(ctx context.Context, root string, name string) error
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/core/port/driven/`
Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add internal/core/port/driven/environment_repository.go
git commit -m "feat: add EnvironmentRepository driven port"
```

---

## Task 4: Adapter model — `EnvironmentModel`

**Files:**
- Create: `internal/adapter/model/environment_model.go`
- Test: `internal/adapter/model/environment_model_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/adapter/model/environment_model_test.go`:

```go
package model

import (
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentModelRoundTrip(t *testing.T) {
	env := domain.Environment{
		Name:      "dev",
		BaseURL:   "https://api.dev.example.com",
		Variables: map[string]string{"token": "abc123"},
	}

	m := NewEnvironmentModelFromDomain(env)
	assert.Equal(t, "dev", m.Name)
	assert.Equal(t, "https://api.dev.example.com", m.BaseURL)
	assert.Equal(t, "abc123", m.Variables["token"])

	got := m.ToDomain()
	assert.Equal(t, env, *got)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/model/ -run TestEnvironmentModelRoundTrip -v`
Expected: FAIL — `undefined: NewEnvironmentModelFromDomain`.

- [ ] **Step 3: Write the implementation**

Create `internal/adapter/model/environment_model.go`:

```go
package model

import "gon/internal/core/domain"

type EnvironmentModel struct {
	Name      string            `yaml:"name,omitempty"`
	BaseURL   string            `yaml:"base_url,omitempty"`
	Variables map[string]string `yaml:"variables,omitempty"`
}

func NewEnvironmentModelFromDomain(environment domain.Environment) *EnvironmentModel {
	return &EnvironmentModel{
		Name:      environment.Name,
		BaseURL:   environment.BaseURL,
		Variables: environment.Variables,
	}
}

func (m *EnvironmentModel) ToDomain() *domain.Environment {
	return &domain.Environment{
		Name:      m.Name,
		BaseURL:   m.BaseURL,
		Variables: m.Variables,
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/model/ -run TestEnvironmentModelRoundTrip -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/model/environment_model.go internal/adapter/model/environment_model_test.go
git commit -m "feat: add EnvironmentModel serialization layer"
```

---

## Task 5: Adapter repository — `environmentRepository`

Persists `environments/<name>.yml` files and the `.gon/active-env` state file.

**Files:**
- Create: `internal/adapter/repository/environment_repository.go`
- Test: `internal/adapter/repository/environment_repository_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/adapter/repository/environment_repository_test.go`:

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

func TestEnvironmentRepositorySaveLoadExistsList(t *testing.T) {
	root := t.TempDir()
	repo := NewEnvironmentRepository()
	ctx := context.Background()

	exists, err := repo.Exists(ctx, root, "dev")
	require.NoError(t, err)
	assert.False(t, exists)

	env := domain.Environment{
		Name:      "dev",
		BaseURL:   "https://api.dev.example.com",
		Variables: map[string]string{"token": "abc123"},
	}
	require.NoError(t, repo.Save(ctx, root, env))

	// file is written under environments/<name>.yml
	_, err = os.Stat(filepath.Join(root, "environments", "dev.yml"))
	require.NoError(t, err)

	exists, err = repo.Exists(ctx, root, "dev")
	require.NoError(t, err)
	assert.True(t, exists)

	got, err := repo.Load(ctx, root, "dev")
	require.NoError(t, err)
	assert.Equal(t, env, *got)

	require.NoError(t, repo.Save(ctx, root, domain.Environment{Name: "prod", BaseURL: "https://api.example.com"}))
	names, err := repo.List(ctx, root)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"dev", "prod"}, names)
}

func TestEnvironmentRepositoryListEmpty(t *testing.T) {
	root := t.TempDir()
	repo := NewEnvironmentRepository()

	names, err := repo.List(context.Background(), root)
	require.NoError(t, err)
	assert.Empty(t, names)
}

func TestEnvironmentRepositoryActiveState(t *testing.T) {
	root := t.TempDir()
	repo := NewEnvironmentRepository()
	ctx := context.Background()

	active, err := repo.ReadActive(ctx, root)
	require.NoError(t, err)
	assert.Equal(t, "", active)

	require.NoError(t, repo.WriteActive(ctx, root, "dev"))

	// active state lives under the gitignored .gon directory
	_, err = os.Stat(filepath.Join(root, ".gon", "active-env"))
	require.NoError(t, err)

	active, err = repo.ReadActive(ctx, root)
	require.NoError(t, err)
	assert.Equal(t, "dev", active)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/repository/ -run TestEnvironmentRepository -v`
Expected: FAIL — `undefined: NewEnvironmentRepository`.

- [ ] **Step 3: Write the implementation**

Create `internal/adapter/repository/environment_repository.go`:

```go
package repository

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gon/internal/adapter/model"
	"gon/internal/core/domain"
	"gon/internal/core/port/driven"

	"gopkg.in/yaml.v3"
)

const (
	environmentsDirName = "environments"
	activeEnvDirName    = ".gon"
	activeEnvFileName   = "active-env"
)

type environmentRepository struct{}

func NewEnvironmentRepository() driven.EnvironmentRepository {
	return &environmentRepository{}
}

func (r *environmentRepository) envFile(root, name string) string {
	return filepath.Join(root, environmentsDirName, name+".yml")
}

func (r *environmentRepository) Save(ctx context.Context, root string, environment domain.Environment) error {
	dir := filepath.Join(root, environmentsDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(model.NewEnvironmentModelFromDomain(environment))
	if err != nil {
		return err
	}
	return os.WriteFile(r.envFile(root, environment.Name), data, 0644)
}

func (r *environmentRepository) Load(ctx context.Context, root string, name string) (*domain.Environment, error) {
	data, err := os.ReadFile(r.envFile(root, name))
	if err != nil {
		return nil, err
	}
	var m model.EnvironmentModel
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *environmentRepository) List(ctx context.Context, root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, environmentsDirName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		base := e.Name()
		if strings.HasSuffix(base, ".yml") {
			names = append(names, strings.TrimSuffix(base, ".yml"))
		} else if strings.HasSuffix(base, ".yaml") {
			names = append(names, strings.TrimSuffix(base, ".yaml"))
		}
	}
	return names, nil
}

func (r *environmentRepository) Exists(ctx context.Context, root string, name string) (bool, error) {
	if _, err := os.Stat(r.envFile(root, name)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *environmentRepository) ReadActive(ctx context.Context, root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, activeEnvDirName, activeEnvFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (r *environmentRepository) WriteActive(ctx context.Context, root string, name string) error {
	dir := filepath.Join(root, activeEnvDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, activeEnvFileName), []byte(name+"\n"), 0644)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapter/repository/ -run TestEnvironmentRepository -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/repository/environment_repository.go internal/adapter/repository/environment_repository_test.go
git commit -m "feat: add environment repository for YAML files and active state"
```

---

## Task 6: Driving port — `EnvironmentService` interface

**Files:**
- Create: `internal/core/port/driving/environment_service.go`

- [ ] **Step 1: Write the interface**

Create `internal/core/port/driving/environment_service.go`:

```go
package driving

import (
	"context"

	"gon/internal/core/domain"
)

type EnvironmentService interface {
	// Create scaffolds a new environment file. Requires an initialized workspace.
	Create(ctx context.Context, root string, name string) error
	// List returns all environment names and the active environment name ("" if none).
	List(ctx context.Context, root string) (names []string, active string, err error)
	// Use marks name as the active environment for this project (local state).
	Use(ctx context.Context, root string, name string) error
	// Resolve returns the active environment. Precedence: override (--env flag) >
	// persisted active state > the sole environment if exactly one exists.
	// Returns (nil, nil) when no environments exist. Returns an error when
	// multiple exist and none is selected, or when a named environment is missing.
	Resolve(ctx context.Context, root string, override string) (*domain.Environment, error)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/core/port/driving/`
Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add internal/core/port/driving/environment_service.go
git commit -m "feat: add EnvironmentService driving port"
```

---

## Task 7: Service — `environmentService`

**Files:**
- Create: `internal/core/service/environment_service.go`
- Test: `internal/core/service/environment_service_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/core/service/environment_service_test.go`:

```go
package service

import (
	"context"
	"testing"

	"gon/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeEnvRepo is an in-memory EnvironmentRepository test double.
type fakeEnvRepo struct {
	envs   map[string]domain.Environment
	active string
}

func newFakeEnvRepo() *fakeEnvRepo {
	return &fakeEnvRepo{envs: map[string]domain.Environment{}}
}

func (f *fakeEnvRepo) Save(_ context.Context, _ string, environment domain.Environment) error {
	f.envs[environment.Name] = environment
	return nil
}
func (f *fakeEnvRepo) Load(_ context.Context, _ string, name string) (*domain.Environment, error) {
	env, ok := f.envs[name]
	if !ok {
		return nil, assertMissing(name)
	}
	e := env
	return &e, nil
}
func (f *fakeEnvRepo) List(_ context.Context, _ string) ([]string, error) {
	names := make([]string, 0, len(f.envs))
	for n := range f.envs {
		names = append(names, n)
	}
	return names, nil
}
func (f *fakeEnvRepo) Exists(_ context.Context, _ string, name string) (bool, error) {
	_, ok := f.envs[name]
	return ok, nil
}
func (f *fakeEnvRepo) ReadActive(_ context.Context, _ string) (string, error) { return f.active, nil }
func (f *fakeEnvRepo) WriteActive(_ context.Context, _ string, name string) error {
	f.active = name
	return nil
}

func assertMissing(name string) error { return &missingEnvErr{name} }

type missingEnvErr struct{ name string }

func (e *missingEnvErr) Error() string { return "environment not found: " + e.name }

func TestEnvironmentServiceResolve(t *testing.T) {
	ctx := context.Background()

	t.Run("override flag wins", func(t *testing.T) {
		repo := newFakeEnvRepo()
		repo.envs["dev"] = domain.Environment{Name: "dev", BaseURL: "https://dev"}
		repo.envs["prod"] = domain.Environment{Name: "prod", BaseURL: "https://prod"}
		repo.active = "dev"
		svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

		got, err := svc.Resolve(ctx, "/root", "prod")
		require.NoError(t, err)
		assert.Equal(t, "prod", got.Name)
	})

	t.Run("falls back to persisted active", func(t *testing.T) {
		repo := newFakeEnvRepo()
		repo.envs["dev"] = domain.Environment{Name: "dev"}
		repo.envs["prod"] = domain.Environment{Name: "prod"}
		repo.active = "dev"
		svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

		got, err := svc.Resolve(ctx, "/root", "")
		require.NoError(t, err)
		assert.Equal(t, "dev", got.Name)
	})

	t.Run("single environment is used when none active", func(t *testing.T) {
		repo := newFakeEnvRepo()
		repo.envs["local"] = domain.Environment{Name: "local"}
		svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

		got, err := svc.Resolve(ctx, "/root", "")
		require.NoError(t, err)
		assert.Equal(t, "local", got.Name)
	})

	t.Run("zero environments returns nil", func(t *testing.T) {
		repo := newFakeEnvRepo()
		svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

		got, err := svc.Resolve(ctx, "/root", "")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("multiple environments and none active errors", func(t *testing.T) {
		repo := newFakeEnvRepo()
		repo.envs["dev"] = domain.Environment{Name: "dev"}
		repo.envs["prod"] = domain.Environment{Name: "prod"}
		svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

		_, err := svc.Resolve(ctx, "/root", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no active environment")
	})
}

func TestEnvironmentServiceCreateRequiresWorkspace(t *testing.T) {
	repo := newFakeEnvRepo()
	svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: false})

	err := svc.Create(context.Background(), "/root", "dev")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no gon workspace found")
}

func TestEnvironmentServiceCreateAndUse(t *testing.T) {
	ctx := context.Background()
	repo := newFakeEnvRepo()
	svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

	require.NoError(t, svc.Create(ctx, "/root", "Dev"))
	_, ok := repo.envs["dev"]
	assert.True(t, ok, "name should be kebab-cased")

	require.NoError(t, svc.Use(ctx, "/root", "dev"))
	assert.Equal(t, "dev", repo.active)

	require.Error(t, svc.Use(ctx, "/root", "missing"))
}

func TestEnvironmentServiceList(t *testing.T) {
	ctx := context.Background()
	repo := newFakeEnvRepo()
	repo.envs["dev"] = domain.Environment{Name: "dev"}
	repo.active = "dev"
	svc := NewEnvironmentService(repo, &mockWorkspaceRepository{existsResponse: true})

	names, active, err := svc.List(ctx, "/root")
	require.NoError(t, err)
	assert.Equal(t, []string{"dev"}, names)
	assert.Equal(t, "dev", active)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/service/ -run TestEnvironmentService -v`
Expected: FAIL — `undefined: NewEnvironmentService`.

- [ ] **Step 3: Write the implementation**

Create `internal/core/service/environment_service.go`:

```go
package service

import (
	"context"
	"fmt"

	"gon/internal/core/domain"
	"gon/internal/core/port/driven"
	"gon/internal/core/port/driving"

	"github.com/iancoleman/strcase"
)

type environmentService struct {
	environmentRepository driven.EnvironmentRepository
	workspaceRepository   driven.WorkspaceRepository
}

func NewEnvironmentService(environmentRepository driven.EnvironmentRepository, workspaceRepository driven.WorkspaceRepository) driving.EnvironmentService {
	return &environmentService{
		environmentRepository: environmentRepository,
		workspaceRepository:   workspaceRepository,
	}
}

func (s *environmentService) Create(ctx context.Context, root string, name string) error {
	if err := ensureWorkspace(ctx, s.workspaceRepository, root); err != nil {
		return err
	}
	normalized := strcase.ToKebab(name)
	if normalized == "" {
		return fmt.Errorf("environment name is required")
	}
	exists, err := s.environmentRepository.Exists(ctx, root, normalized)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("environment already exists: %s", normalized)
	}
	env := domain.Environment{
		Name:      normalized,
		BaseURL:   "https://api.example.com",
		Variables: map[string]string{},
	}
	return s.environmentRepository.Save(ctx, root, env)
}

func (s *environmentService) List(ctx context.Context, root string) ([]string, string, error) {
	names, err := s.environmentRepository.List(ctx, root)
	if err != nil {
		return nil, "", err
	}
	active, err := s.environmentRepository.ReadActive(ctx, root)
	if err != nil {
		return nil, "", err
	}
	return names, active, nil
}

func (s *environmentService) Use(ctx context.Context, root string, name string) error {
	if err := ensureWorkspace(ctx, s.workspaceRepository, root); err != nil {
		return err
	}
	exists, err := s.environmentRepository.Exists(ctx, root, name)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("environment not found: %s", name)
	}
	return s.environmentRepository.WriteActive(ctx, root, name)
}

func (s *environmentService) Resolve(ctx context.Context, root string, override string) (*domain.Environment, error) {
	if override != "" {
		return s.environmentRepository.Load(ctx, root, override)
	}
	active, err := s.environmentRepository.ReadActive(ctx, root)
	if err != nil {
		return nil, err
	}
	if active != "" {
		return s.environmentRepository.Load(ctx, root, active)
	}
	names, err := s.environmentRepository.List(ctx, root)
	if err != nil {
		return nil, err
	}
	switch len(names) {
	case 0:
		return nil, nil
	case 1:
		return s.environmentRepository.Load(ctx, root, names[0])
	default:
		return nil, fmt.Errorf("no active environment; run 'env use <name>' or pass --env")
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/service/ -run TestEnvironmentService -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/service/environment_service.go internal/core/service/environment_service_test.go
git commit -m "feat: add environment service with active-env resolution"
```

---

## Task 8: HTTP service — env param, substitution, fail-fast

Changing `Execute`'s signature breaks the `HttpService` interface, all callers, and existing tests. This task updates them all together so the package compiles.

**Files:**
- Modify: `internal/core/port/driving/http_service.go`
- Modify: `internal/core/service/http_service.go`
- Modify: `internal/core/service/http_service_test.go`
- Modify: `internal/core/service/request_service.go:63` (caller — temporary `nil` until Task 9)

- [ ] **Step 1: Update the port interface**

Replace `internal/core/port/driving/http_service.go`:

```go
package driving

import (
	"context"

	"gon/internal/core/domain"
	"gon/internal/core/payload"
)

type HttpService interface {
	Execute(ctx context.Context, input *payload.HttpExecuteInput, env *domain.Environment) (*payload.HttpExecuteOutput, error)
}
```

- [ ] **Step 2: Add substitution tests (these will fail to compile until Step 4)**

Add to `internal/core/service/http_service_test.go` (append these two tests at the end of the file):

```go
func TestHttpServiceExecuteSubstitutesVariables(t *testing.T) {
	var gotPath, gotHeader, gotQuery string
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHeader = r.Header.Get("Authorization")
		gotQuery = r.URL.Query().Get("uid")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	env := &domain.Environment{
		Name:      "dev",
		BaseURL:   server.URL,
		Variables: map[string]string{"token": "secret", "uid": "42"},
	}
	svc := NewHttpService(nil, server.Client())

	input := &payload.HttpExecuteInput{
		Method:  http.MethodPost,
		URL:     "/users/{{uid}}",
		Headers: map[string][]string{"Authorization": {"Bearer {{token}}"}},
		Query:   map[string][]string{"uid": {"{{uid}}"}},
		Body:    []byte(`{"id":"{{uid}}"}`),
	}
	_, err := svc.Execute(context.Background(), input, env)

	require.NoError(t, err)
	assert.Equal(t, "/users/42", gotPath)
	assert.Equal(t, "Bearer secret", gotHeader)
	assert.Equal(t, "42", gotQuery)
	assert.Equal(t, `{"id":"42"}`, string(gotBody))
}

func TestHttpServiceExecuteFailsOnUnresolvedVariable(t *testing.T) {
	env := &domain.Environment{Name: "dev", BaseURL: "https://api.example.com"}
	svc := NewHttpService(nil, http.DefaultClient)

	_, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
		Method:  http.MethodGet,
		URL:     "/users",
		Headers: map[string][]string{"Authorization": {"Bearer {{token}}"}},
	}, env)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unresolved variable")
	assert.Contains(t, err.Error(), "token")
}
```

- [ ] **Step 3: Update existing test call sites to pass `nil` env**

In `internal/core/service/http_service_test.go`, every existing call is `svc.Execute(context.Background(), <input>)`. Add a trailing `, nil` argument to each. There are 11 existing calls (in `TestHttpServiceExecuteSuccess`, `...ResolvesWorkspaceURL`, `...AbsoluteURLBypassesWorkspace`, `...AppliesWorkspaceDefaults`, `...RequestOverridesWorkspaceDefaults`, `...ForwardsHeaders`, `...EncodesQuery`, `...SendsBody`, `...NilBody`, `...InvalidMethod`, `...TransportError`).

Example — change:

```go
out, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
	Method: http.MethodGet,
	URL:    server.URL,
})
```

to:

```go
out, err := svc.Execute(context.Background(), &payload.HttpExecuteInput{
	Method: http.MethodGet,
	URL:    server.URL,
}, nil)
```

- [ ] **Step 4: Implement the new `Execute`**

Replace `internal/core/service/http_service.go`:

```go
package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"gon/internal/core/domain"
	"gon/internal/core/payload"
	"gon/internal/core/port/driving"
)

type httpService struct {
	workspace  *domain.Workspace
	httpClient *http.Client
}

func NewHttpService(workspace *domain.Workspace, httpClient *http.Client) driving.HttpService {
	return &httpService{
		workspace:  workspace,
		httpClient: httpClient,
	}
}

func (s *httpService) Execute(ctx context.Context, input *payload.HttpExecuteInput, env *domain.Environment) (*payload.HttpExecuteOutput, error) {
	start := time.Now()

	base := ""
	configPath := ""
	if env != nil {
		base = env.BaseURL
	}
	if s.workspace != nil {
		s.workspace.ApplyDefaults(input)
		configPath = s.workspace.Config.Path
		if base == "" {
			base = s.workspace.BaseURL // deprecated fallback
		}
	}

	url := domain.ResolveURL(base, configPath, input.URL)
	url = substituteInput(url, input, env)

	if missing := unresolvedVariables(url, input); len(missing) > 0 {
		return nil, fmt.Errorf("unresolved variables: %s", strings.Join(missing, ", "))
	}

	var requestBody io.Reader
	if input.Body != nil {
		requestBody = bytes.NewReader(input.Body)
	}

	req, err := http.NewRequestWithContext(ctx, input.Method, url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("error building request : %w", err)
	}

	for key, values := range input.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if len(input.Query) > 0 {
		q := req.URL.Query()
		for key, values := range input.Query {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("response error: %w", err)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	result := payload.HttpExecuteOutput{
		Body:       responseBody,
		StatusCode: res.StatusCode,
		Headers:    map[string][]string(res.Header),
		Metadata: payload.Metadata{
			ExecutionTime: time.Since(start),
			ContentType:   res.Header.Get("Content-Type"),
			ContentLength: res.ContentLength,
		},
	}

	return &result, nil
}

// substituteInput resolves {{var}} placeholders in the URL and, in place, across
// every header value, query value, and the body. It returns the resolved URL.
func substituteInput(url string, input *payload.HttpExecuteInput, env *domain.Environment) string {
	if env == nil {
		return url
	}
	url = env.Substitute(url)
	for _, values := range input.Headers {
		for i := range values {
			values[i] = env.Substitute(values[i])
		}
	}
	for _, values := range input.Query {
		for i := range values {
			values[i] = env.Substitute(values[i])
		}
	}
	if input.Body != nil {
		input.Body = []byte(env.Substitute(string(input.Body)))
	}
	return url
}

// unresolvedVariables returns the sorted, de-duplicated names of any {{var}}
// placeholders still present in the URL, headers, query, or body.
func unresolvedVariables(url string, input *payload.HttpExecuteInput) []string {
	seen := map[string]struct{}{}
	add := func(s string) {
		for _, name := range domain.FindPlaceholders(s) {
			seen[name] = struct{}{}
		}
	}
	add(url)
	for _, values := range input.Headers {
		for _, v := range values {
			add(v)
		}
	}
	for _, values := range input.Query {
		for _, v := range values {
			add(v)
		}
	}
	if input.Body != nil {
		add(string(input.Body))
	}
	if len(seen) == 0 {
		return nil
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
```

- [ ] **Step 5: Update the `request_service.go` caller to compile**

In `internal/core/service/request_service.go`, line ~63, change:

```go
	result, err := s.httpService.Execute(ctx, input)
```

to (temporary — Task 9 threads the real env):

```go
	result, err := s.httpService.Execute(ctx, input, nil)
```

- [ ] **Step 6: Run the service tests**

Run: `go test ./internal/core/service/ -run TestHttpService -v`
Expected: PASS (including the two new substitution tests).

- [ ] **Step 7: Commit**

```bash
git add internal/core/port/driving/http_service.go internal/core/service/http_service.go internal/core/service/http_service_test.go internal/core/service/request_service.go
git commit -m "feat: apply environment variable substitution in http service"
```

---

## Task 9: Request service — thread env through `Run`

**Files:**
- Modify: `internal/core/port/driving/request_service.go`
- Modify: `internal/core/service/request_service.go`
- Modify: `internal/core/service/request_service_test.go`

- [ ] **Step 1: Update the port interface**

Replace `internal/core/port/driving/request_service.go`:

```go
package driving

import (
	"context"

	"gon/internal/core/domain"
	"gon/internal/core/payload"
)

type RequestService interface {
	// Run executes a saved request. overrides (may be nil) carries per-execution
	// header/query/body values that take precedence over the request file. env
	// (may be nil) is the active environment used for {{var}} substitution.
	Run(ctx context.Context, root string, requestPath string, overrides *payload.HttpExecuteInput, env *domain.Environment) (*payload.HttpExecuteInput, *payload.HttpExecuteOutput, error)
	// Create scaffolds a new request file with the given method.
	Create(ctx context.Context, root string, requestPath string, method string) error
}
```

- [ ] **Step 2: Update `request_service.go` implementation**

In `internal/core/service/request_service.go`, change the `Run` signature and the `Execute` call. The method header becomes:

```go
func (s *requestService) Run(ctx context.Context, root string, requestPath string, overrides *payload.HttpExecuteInput, env *domain.Environment) (*payload.HttpExecuteInput, *payload.HttpExecuteOutput, error) {
```

and the `Execute` call (previously `s.httpService.Execute(ctx, input, nil)` from Task 8) becomes:

```go
	result, err := s.httpService.Execute(ctx, input, env)
```

- [ ] **Step 3: Update `request_service_test.go`**

The fake `captureHttpService.Execute` must match the new interface, and the three `svc.Run(...)` calls need a trailing `env` argument.

Change the fake's `Execute` method:

```go
func (c *captureHttpService) Execute(ctx context.Context, input *payload.HttpExecuteInput, env *domain.Environment) (*payload.HttpExecuteOutput, error) {
	c.input = input
	return &payload.HttpExecuteOutput{StatusCode: 200}, nil
}
```

Update the three `Run` call sites by appending `, nil`:
- `svc.Run(context.Background(), "/root", "auth/admin/impersonate", overrides)` → `..., overrides, nil)`
- `svc.Run(context.Background(), "/root", "auth/x", nil)` → `..., "auth/x", nil, nil)`
- `svc.Run(context.Background(), "/root", "auth/x", nil)` (in `TestRequestServiceRequiresWorkspace`) → `..., "auth/x", nil, nil)`

Then add a test verifying env is threaded into `Execute` (append to the file):

```go
func TestRequestServiceRunSubstitutesEnv(t *testing.T) {
	repo := &fakeRequestRepo{
		request: domain.Request{Method: "GET", URL: "/users/{{uid}}"},
	}
	http := &captureHttpService{}
	svc := NewRequestService(repo, nil, &mockWorkspaceRepository{existsResponse: true}, http)

	env := &domain.Environment{Name: "dev", Variables: map[string]string{"uid": "42"}}
	_, _, err := svc.Run(context.Background(), "/root", "users/get", nil, env)

	require.NoError(t, err)
	// the captured input still holds the raw placeholder; substitution happens
	// inside HttpService.Execute, but the env must reach it unchanged.
	assert.Equal(t, "/users/{{uid}}", http.input.URL)
}
```

> Note: the fake `captureHttpService` does not perform substitution, so this test asserts the env is accepted and the call succeeds. Real substitution is covered by `TestHttpServiceExecuteSubstitutesVariables` in Task 8.

- [ ] **Step 4: Run the service tests**

Run: `go test ./internal/core/service/ -run TestRequestService -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/port/driving/request_service.go internal/core/service/request_service.go internal/core/service/request_service_test.go
git commit -m "feat: thread active environment through request Run"
```

---

## Task 10: Commands — `--env` flag on HTTP and run

`HttpCommand` and `RunCommand` must resolve the active environment (honoring `--env`) and pass it to the service. Both gain an `EnvironmentService` dependency.

**Files:**
- Modify: `internal/adapter/command/http_command.go`
- Modify: `internal/adapter/command/run_command.go`
- Modify: `internal/adapter/command/http_command_test.go` (only if it constructs `HttpCommand` — see Step 5)

- [ ] **Step 1: Update `HttpCommand` to resolve and pass env**

In `internal/adapter/command/http_command.go`:

Change the function signature:

```go
func HttpCommand(method string, httpService driving.HttpService, environmentService driving.EnvironmentService, httpOutput driven.HttpOutput) *cli.Command {
```

Add `"os"` to the imports. Inside the `Action`, after building `input` and before `result, err := httpService.Execute(...)`, resolve the env:

```go
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				env, err := environmentService.Resolve(ctx, cwd, cmd.String("env"))
				if err != nil {
					return err
				}

				mode := resolveMode(cmd)

				result, err := httpService.Execute(ctx, input, env)
```

(Replace the existing `mode := resolveMode(cmd)` / `result, err := httpService.Execute(ctx, input)` lines with the block above.)

Add the `--env` flag to the `Flags` slice:

```go
				&cli.StringFlag{
					Name:  "env",
					Usage: "Environment to use for this request (overrides the active one)",
				},
```

- [ ] **Step 2: Update `RunCommand` to resolve and pass env**

In `internal/adapter/command/run_command.go`:

Change the function signature:

```go
func RunCommand(requestService driving.RequestService, environmentService driving.EnvironmentService, httpOutput driven.HttpOutput) *cli.Command {
```

Inside the `Action`, after computing `cwd`, resolve the env and pass it to `Run`:

```go
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}

				env, err := environmentService.Resolve(ctx, cwd, cmd.String("env"))
				if err != nil {
					return err
				}

				merged, result, err := requestService.Run(ctx, cwd, cmd.StringArg("path"), overrides, env)
				if err != nil {
					return err
				}
				httpOutput.Format(merged, result, resolveMode(cmd))
				return nil
```

Add the `--env` flag to the `Flags` slice:

```go
				&cli.StringFlag{Name: "env", Usage: "Environment to use for this request (overrides the active one)"},
```

- [ ] **Step 3: Verify the command package compiles (callers in cmd/main.go are updated in Task 13)**

Run: `go build ./internal/adapter/command/`
Expected: FAIL with `not enough arguments in call to HttpCommand` only inside test files, OR success. If the command package itself builds, proceed. (The wiring in `cmd/main.go` is fixed in Task 13.)

- [ ] **Step 4: Check the existing command test for `HttpCommand` construction**

Run: `grep -n "HttpCommand(" internal/adapter/command/http_command_test.go`
Expected: lists any call sites that construct `HttpCommand`.

- [ ] **Step 5: If Step 4 found call sites, update them**

For each `HttpCommand(method, httpService, httpOutput)` call in the test, insert a stub environment service argument. First add this fake to the top of `internal/adapter/command/http_command_test.go` (below the imports — add `"gon/internal/core/domain"` to imports):

```go
type stubEnvService struct{}

func (stubEnvService) Create(context.Context, string, string) error { return nil }
func (stubEnvService) List(context.Context, string) ([]string, string, error) {
	return nil, "", nil
}
func (stubEnvService) Use(context.Context, string, string) error { return nil }
func (stubEnvService) Resolve(context.Context, string, string) (*domain.Environment, error) {
	return nil, nil
}
```

Then change each `HttpCommand(method, httpService, httpOutput)` to `HttpCommand(method, httpService, stubEnvService{}, httpOutput)`.

> If Step 4 found no call sites (the test exercises only `parseHeaders`/`parseQuery`/`resolveMode`), skip this step.

- [ ] **Step 6: Run the command tests**

Run: `go test ./internal/adapter/command/ -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/adapter/command/http_command.go internal/adapter/command/run_command.go internal/adapter/command/http_command_test.go
git commit -m "feat: add --env flag to http and run commands"
```

---

## Task 11: Command — `env` CLI (new/list/use)

**Files:**
- Create: `internal/adapter/command/env_command.go`

- [ ] **Step 1: Write the command**

Create `internal/adapter/command/env_command.go`:

```go
package command

import (
	"context"
	"fmt"
	"os"

	"gon/internal/core/port/driving"

	"github.com/urfave/cli/v3"
)

func EnvCommand(environmentService driving.EnvironmentService) *cli.Command {
	return &cli.Command{
		Name:      "env",
		Usage:     "Manage environments",
		ArgsUsage: "new|list|use <name>",
		Commands: []*cli.Command{
			{
				Name:      "new",
				Usage:     "Create a new environment (e.g. env new dev)",
				ArgsUsage: "<name>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name", UsageText: "Environment name, e.g. dev"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					name := cmd.StringArg("name")
					if name == "" {
						return fmt.Errorf("environment name is required")
					}
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					if err := environmentService.Create(ctx, cwd, name); err != nil {
						return err
					}
					fmt.Printf("Environment '%s' created\n", name)
					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List environments (active marked with *)",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					names, active, err := environmentService.List(ctx, cwd)
					if err != nil {
						return err
					}
					if len(names) == 0 {
						fmt.Println("No environments. Create one with 'env new <name>'")
						return nil
					}
					for _, name := range names {
						marker := " "
						if name == active {
							marker = "*"
						}
						fmt.Printf("%s %s\n", marker, name)
					}
					return nil
				},
			},
			{
				Name:      "use",
				Usage:     "Set the active environment (e.g. env use prod)",
				ArgsUsage: "<name>",
				Arguments: []cli.Argument{
					&cli.StringArg{Name: "name", UsageText: "Environment name to activate"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					name := cmd.StringArg("name")
					if name == "" {
						return fmt.Errorf("environment name is required")
					}
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					if err := environmentService.Use(ctx, cwd, name); err != nil {
						return err
					}
					fmt.Printf("Active environment set to '%s'\n", name)
					return nil
				},
			},
		},
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/adapter/command/`
Expected: success (the `cmd/main.go` wiring is added in Task 13).

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/command/env_command.go
git commit -m "feat: add env command (new/list/use)"
```

---

## Task 12: Workspace init — scaffold `local` environment

`gon init` should stop writing `base_url` to `workspace.yml`, and instead scaffold `environments/local.yml` and mark `local` active, so a fresh workspace works out of the box.

**Files:**
- Modify: `internal/core/service/workspace_service.go`
- Modify: `internal/core/service/workspace_service_test.go`

- [ ] **Step 1: Update the failing test**

In `internal/core/service/workspace_service_test.go`:

Replace `TestWorkspaceServiceCreate` and `TestWorkspaceServiceCreatePropagatesError` with versions that inject a fake environment repository and assert the scaffold. (The `fakeEnvRepo` type already exists from Task 7 in the same package — reuse it.)

```go
func TestWorkspaceServiceCreate(t *testing.T) {
	repo := &mockWorkspaceRepository{}
	envRepo := newFakeEnvRepo()
	svc := NewWorkspaceService(repo, envRepo)

	err := svc.Create(context.Background(), "/home/user/My Project")
	require.NoError(t, err)

	assert.Equal(t, 1, repo.saveCalls)
	assert.Equal(t, "/home/user/My Project", repo.savedDir)
	assert.Equal(t, "my-project", repo.saved.Name)
	// base_url is no longer written to the workspace; it lives in the environment.
	assert.Equal(t, "", repo.saved.BaseURL)
	assert.Equal(t, domain.Config{}, repo.saved.Config)

	// a 'local' environment is scaffolded and marked active.
	local, ok := envRepo.envs["local"]
	require.True(t, ok)
	assert.Equal(t, "https://api.example.com", local.BaseURL)
	assert.Equal(t, "local", envRepo.active)
}

func TestWorkspaceServiceCreatePropagatesError(t *testing.T) {
	sentinel := errors.New("disk full")
	repo := &mockWorkspaceRepository{saveErr: sentinel}
	svc := NewWorkspaceService(repo, newFakeEnvRepo())

	err := svc.Create(context.Background(), "/tmp/proj")
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/service/ -run TestWorkspaceServiceCreate -v`
Expected: FAIL — `not enough arguments in call to NewWorkspaceService`.

- [ ] **Step 3: Update `workspace_service.go`**

Replace the struct, constructor, and `Create` in `internal/core/service/workspace_service.go` (keep `ensureWorkspace` and `getFolderName` unchanged):

```go
type workspaceService struct {
	workspaceRepository   driven.WorkspaceRepository
	environmentRepository driven.EnvironmentRepository
}

func NewWorkspaceService(repo driven.WorkspaceRepository, environmentRepository driven.EnvironmentRepository) driving.WorkspaceService {
	return &workspaceService{workspaceRepository: repo, environmentRepository: environmentRepository}
}

func (s *workspaceService) Create(ctx context.Context, directory string) error {
	workspace := domain.Workspace{
		Name:   getFolderName(directory),
		Config: domain.Config{},
	}
	if err := s.workspaceRepository.Save(ctx, directory, workspace); err != nil {
		return err
	}

	local := domain.Environment{
		Name:      "local",
		BaseURL:   "https://api.example.com",
		Variables: map[string]string{},
	}
	if err := s.environmentRepository.Save(ctx, directory, local); err != nil {
		return err
	}
	return s.environmentRepository.WriteActive(ctx, directory, "local")
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/service/ -run TestWorkspaceServiceCreate -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/service/workspace_service.go internal/core/service/workspace_service_test.go
git commit -m "feat: scaffold local environment on workspace init"
```

---

## Task 13: Wiring — `cmd/main.go`

Wire the environment repository/service and update all changed command/service constructors. Make the REPL prompt reflect the active environment.

**Files:**
- Modify: `cmd/main.go`

- [ ] **Step 1: Update `cli_app` wiring**

In `cmd/main.go`, inside `cli_app`, after the existing repository/service construction, add the environment wiring and update the constructors that changed:

Replace:

```go
	workspaceRepository := repository.NewWorkspaceRepository()
	workspaceService := service.NewWorkspaceService(workspaceRepository)

	requestRepository := repository.NewRequestRepository()
	collectionRepository := repository.NewCollectionRepository()
	requestService := service.NewRequestService(requestRepository, collectionRepository, workspaceRepository, httpService)
	collectionService := service.NewCollectionService(collectionRepository, workspaceRepository)
```

with:

```go
	workspaceRepository := repository.NewWorkspaceRepository()
	environmentRepository := repository.NewEnvironmentRepository()
	environmentService := service.NewEnvironmentService(environmentRepository, workspaceRepository)
	workspaceService := service.NewWorkspaceService(workspaceRepository, environmentRepository)

	requestRepository := repository.NewRequestRepository()
	collectionRepository := repository.NewCollectionRepository()
	requestService := service.NewRequestService(requestRepository, collectionRepository, workspaceRepository, httpService)
	collectionService := service.NewCollectionService(collectionRepository, workspaceRepository)
```

- [ ] **Step 2: Update command construction**

Replace the `httpCommands` slice so each `HttpCommand` receives `environmentService`:

```go
	httpCommands := []*cli.Command{
		command.HttpCommand(strings.ToLower(http.MethodGet), httpService, environmentService, httpOutput),
		command.HttpCommand(strings.ToLower(http.MethodPost), httpService, environmentService, httpOutput),
		command.HttpCommand(strings.ToLower(http.MethodPut), httpService, environmentService, httpOutput),
		command.HttpCommand(strings.ToLower(http.MethodDelete), httpService, environmentService, httpOutput),
		command.HttpCommand(strings.ToLower(http.MethodPatch), httpService, environmentService, httpOutput),
	}
```

Replace the `collectionCommands` slice so `RunCommand` receives `environmentService` and the new `EnvCommand` is registered:

```go
	collectionCommands := []*cli.Command{
		command.RunCommand(requestService, environmentService, httpOutput),
		command.CollectionInitCommand(collectionService),
		command.RequestNewCommand(requestService),
	}

	environmentCommands := []*cli.Command{
		command.EnvCommand(environmentService),
	}
```

- [ ] **Step 3: Register the Environments group and include it in `commands`**

In the `groups` slice, add an Environments group (place it after Collections):

```go
	groups := []command.CommandGroup{
		{Name: "HTTP Commands", Commands: httpCommands},
		{Name: "Workspace", Commands: workspaceCommands},
		{Name: "Collections", Commands: collectionCommands},
		{Name: "Environments", Commands: environmentCommands},
		{Name: "Common", Commands: utilityCommands},
	}
```

Because a group index changed, update the help-group index. Replace:

```go
	helpCmd := command.HelpCommand(groups)
	utilityCommands = append(utilityCommands, helpCmd)
	groups[3].Commands = utilityCommands
```

with:

```go
	helpCmd := command.HelpCommand(groups)
	utilityCommands = append(utilityCommands, helpCmd)
	groups[4].Commands = utilityCommands
```

And include the environment commands in the flat `commands` slice. Replace:

```go
	commands := append(httpCommands, workspaceCommands...)
	commands = append(commands, collectionCommands...)
	commands = append(commands, utilityCommands...)
```

with:

```go
	commands := append(httpCommands, workspaceCommands...)
	commands = append(commands, collectionCommands...)
	commands = append(commands, environmentCommands...)
	commands = append(commands, utilityCommands...)
```

- [ ] **Step 4: Make the REPL prompt reflect the active environment**

In `repl()`, replace the prompt-building block:

```go
	if workspace != nil {
		prompt = utility.ColorInfo("gon(" + workspace.Name + ")> ")
		cacheDirectory := filepath.Join(cwd, ".cache")
		os.Mkdir(cacheDirectory, 0755)
		historyFile = filepath.Join(cacheDirectory, workspace.Name+".history")
	}
```

with:

```go
	environmentRepository := repository.NewEnvironmentRepository()
	buildPrompt := func() string {
		if workspace == nil {
			return utility.ColorInfo("gon> ")
		}
		active, _ := environmentRepository.ReadActive(context.Background(), cwd)
		label := workspace.Name
		if active != "" {
			label = workspace.Name + ":" + active
		}
		return utility.ColorInfo("gon(" + label + ")> ")
	}

	if workspace != nil {
		prompt = buildPrompt()
		cacheDirectory := filepath.Join(cwd, ".cache")
		os.Mkdir(cacheDirectory, 0755)
		historyFile = filepath.Join(cacheDirectory, workspace.Name+".history")
	}
```

Then, in the REPL `for` loop, after a command runs, refresh the prompt so `env use` is reflected. Replace:

```go
		if err := gon_app.Run(context.Background(), append([]string{"gon"}, args...)); err != nil {
			fmt.Println(err)
		}
		fmt.Println()
```

with:

```go
		if err := gon_app.Run(context.Background(), append([]string{"gon"}, args...)); err != nil {
			fmt.Println(err)
		}
		fmt.Println()
		rl.SetPrompt(buildPrompt())
```

- [ ] **Step 5: Build the whole project**

Run: `go build ./...`
Expected: success (no output).

- [ ] **Step 6: Run the full test suite**

Run: `go test ./...`
Expected: all packages PASS.

- [ ] **Step 7: Manual smoke test**

```bash
go build -o gon ./cmd
mkdir -p /tmp/gon-env-smoke && cd /tmp/gon-env-smoke
/home/fakhrulnugroho/work/opensource/gon/gon init
cat environments/local.yml
cat .gon/active-env
/home/fakhrulnugroho/work/opensource/gon/gon env new dev
/home/fakhrulnugroho/work/opensource/gon/gon env list
/home/fakhrulnugroho/work/opensource/gon/gon env use dev
/home/fakhrulnugroho/work/opensource/gon/gon env list
```

Expected:
- `environments/local.yml` contains `name: local` and `base_url: https://api.example.com`.
- `.gon/active-env` contains `local`.
- `env list` shows `* local` then ` dev`; after `env use dev` it shows ` local` and `* dev`.

- [ ] **Step 8: Commit**

```bash
cd /home/fakhrulnugroho/work/opensource/gon
git add cmd/main.go
git commit -m "feat: wire environments and reflect active env in REPL prompt"
```

---

## Task 14: Documentation

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Document environments in `README.md`**

Add an "Environments" section after the collections/requests documentation. Cover:
- `gon init` scaffolds `environments/local.yml` and marks it active.
- `env new <name>` creates `environments/<name>.yml`; edit the file to set `base_url` and `variables`.
- `env list` / `env use <name>` to view and switch the active environment.
- `--env <name>` on `get`/`post`/`put`/`delete`/`patch`/`run` overrides the active environment for one call.
- `{{var}}` substitution in URLs, header values, query values, and body, resolved from the active environment; unresolved variables fail the request.
- The active environment is stored in `.gon/active-env` (gitignored, per-developer).

Example block to include:

````markdown
## Environments

Environments are project-scoped, named sets of variables plus a base URL
(`local`, `dev`, `test`, `prod`). `gon init` creates `environments/local.yml`
and marks it active.

```bash
gon env new dev          # creates environments/dev.yml
gon env list             # active is marked with *
gon env use dev          # switch the active environment
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

Requests reference variables with `{{name}}` in the URL, header values, query
values, and body — for example `Authorization: Bearer {{token}}` or
`/users/{{user_id}}`. Values resolve from the active environment; an unresolved
`{{var}}` fails the request. The active selection lives in `.gon/active-env`
(gitignored), so each developer chooses their own.
````

- [ ] **Step 2: Document the Environments architecture in `CLAUDE.md`**

Add a "### Environments" subsection under "## Architecture" (after "### Collections & requests"). Summarize:
- `domain.Environment` (`Name`, `BaseURL`, `Variables`) + `Substitute`/`FindPlaceholders` and the `ResolveURL` free function.
- Storage: `environments/<name>.yml` via `EnvironmentRepository`; active selection in `.gon/active-env` (gitignored) via `ReadActive`/`WriteActive`.
- `EnvironmentService.Resolve` precedence: `--env` flag > active state > sole environment > error.
- Resolution happens in the command layer; the resolved `*domain.Environment` is passed into `HttpService.Execute(ctx, input, env)` / `RequestService.Run(..., env)`, which apply `{{var}}` substitution and fail fast on unresolved variables.
- `workspace.BaseURL` is a deprecated fallback; the active environment's `base_url` is the source of truth; `gon init` scaffolds `environments/local.yml`.

- [ ] **Step 3: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: document environments feature"
```

---

## Self-Review Notes

- **Spec coverage:** data model (Task 2/4), separate `environments/<name>.yml` storage (Task 5), shared workspace config with `{{var}}` (Task 8 substitution applies to merged headers/query), substitution scope URL/headers/query/body (Task 8), fail-fast on undefined vars (Task 8), active-env precedence flag→local-file→single-env-fallback→error (Task 7), `.gon/active-env` gitignored state (Task 5), resolve-at-command-pass-to-service (Tasks 8–10), `env new/list/use` (Task 11), `--env` flag (Task 10), `gon init` scaffolds `local` (Task 12), REPL prompt `gon(proj:env)>` (Task 13), docs (Task 14). All covered.
- **Out of scope (per spec):** `env set`/`env show`, nested variable resolution, secret encryption, committed `default_environment` — intentionally not implemented.
- **Signature consistency:** `Execute(ctx, input, env)` and `Run(ctx, root, path, overrides, env)` are updated in the port, implementation, all callers, and all tests within Tasks 8–10 and 13, keeping each package compilable at its commit.
