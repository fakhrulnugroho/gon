# Improvement Options

After every review, provide **2–3 concrete improvement options**. Each option must represent a
meaningfully different approach — not just the same idea with different wording.

Options should differ along dimensions like:
- How much refactoring is required (minimal vs full restructure)
- Abstraction level (concrete vs interface-based)
- Scope (single function vs whole package)
- Tradeoff profile (simplicity vs testability, speed vs readability)

---

## Format

```
---
## 🛠️ Improvement Options

### Option 1: [Short descriptive name]
**Approach**: [1–2 sentences describing what changes]

**Why this**: [Reason to choose this option — what problem it solves best]

**Code example** (if applicable):
```go
// before / after snippet
```

**⚖️ Tradeoffs**
| | |
|---|---|
| ✅ Pros | [What you gain] |
| ❌ Cons | [What you give up or risk] |
| 🎯 Best when | [Specific context where this option wins] |

---

### Option 2: [Short descriptive name]
... (same structure)

---

### Option 3: [Short descriptive name] *(optional — only if genuinely different)*
... (same structure)

---

### 💬 Recommendation
[Pick the one you'd actually ship and say why in 1–2 sentences.
Be direct. Don't hedge with "it depends" unless you truly need more context.]
```

---

## Rules for Writing Options

1. **Always include a recommendation** — don't leave the user to figure it out alone
2. **Show code when possible** — abstract advice is forgettable, code is actionable
3. **Name options clearly** — "Use concrete type", "Extract interface", "Inline and simplify" — not "Option A/B/C"
4. **Tradeoffs must be honest** — if an option has a real downside, say it; don't oversell
5. **3 options max** — if you can't find 3 meaningfully different approaches, 2 is fine
6. **Options must be ranked implicitly** — put the one you'd recommend first, or mark it clearly

---

## Example: Error Handling Pattern

Scenario: user is returning raw DB errors from their service layer.

```
## 🛠️ Improvement Options

### Option 1: Wrap errors inline with context
**Approach**: Use `fmt.Errorf("getting user %d: %w", id, err)` directly at the call site.

**Why this**: Lowest effort, keeps code flat, errors are still inspectable with `errors.Is`.

```go
// before
return nil, err

// after
return nil, fmt.Errorf("getting user %d: %w", id, err)
```

**⚖️ Tradeoffs**
| | |
|---|---|
| ✅ Pros | Minimal change, readable, no new types needed |
| ❌ Cons | No structured error metadata (HTTP status, error code) |
| 🎯 Best when | Internal services, CLIs, simple APIs |

---

### Option 2: Define sentinel errors per domain
**Approach**: Declare `var ErrUserNotFound = errors.New("user not found")` in the package,
return it from the repo, check with `errors.Is` in the handler.

**Why this**: Callers can branch on error type without string matching.

```go
// repository.go
var ErrNotFound = errors.New("not found")

func (r *repo) GetUser(id int) (*User, error) {
    if notFound {
        return nil, ErrNotFound
    }
}

// handler.go
if errors.Is(err, user.ErrNotFound) {
    http.Error(w, "not found", 404)
}
```

**⚖️ Tradeoffs**
| | |
|---|---|
| ✅ Pros | Clean branching, idiomatic, no reflection |
| ❌ Cons | Adds a few exported vars per package |
| 🎯 Best when | HTTP APIs that need to map errors to status codes |

---

### Option 3: Custom error type with HTTP status
**Approach**: Create an `AppError` struct carrying a status code and message, return it from service.

**Why this**: Single place to map business errors to HTTP responses.

```go
type AppError struct {
    Status  int
    Message string
    Err     error
}
func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }
```

**⚖️ Tradeoffs**
| | |
|---|---|
| ✅ Pros | Centralized HTTP mapping, good for large APIs |
| ❌ Cons | More boilerplate, couples service layer to HTTP concerns |
| 🎯 Best when | APIs with many error types that all need different status codes |

---

### 💬 Recommendation
Go with **Option 2** (sentinel errors). It's idiomatic, practical for most Go HTTP services,
and doesn't couple your service layer to HTTP status codes like Option 3 does.
Only move to Option 3 if you have 10+ distinct error types and tired of writing the same switch.
```

---

## Example: Interface vs Concrete Type

Scenario: user added a `UserRepository` interface but only has one Postgres implementation.

```
## 🛠️ Improvement Options

### Option 1: Remove the interface, use concrete struct directly
**Approach**: Delete the interface, inject `*PostgresUserRepo` directly into the service.

**Why this**: YAGNI — you don't have a second implementation, so the interface adds no value.

```go
// before
type UserRepository interface {
    GetByID(id int) (*User, error)
}

// after — just use the struct
type UserService struct {
    repo *PostgresUserRepo
}
```

**⚖️ Tradeoffs**
| | |
|---|---|
| ✅ Pros | Simpler, less indirection, faster to navigate |
| ❌ Cons | Harder to mock in tests without a real DB |
| 🎯 Best when | Early-stage projects, CLIs, scripts, or if you use integration tests |

---

### Option 2: Keep interface, define it in the consumer (service), not the repo
**Approach**: Move the interface declaration to `service.go`, not `repository.go`.

**Why this**: Go interfaces are satisfied implicitly — define them where they're used,
not where they're implemented. This is the idiomatic Go way.

```go
// service.go
type userRepo interface {  // unexported — only this package needs it
    GetByID(id int) (*User, error)
}

type UserService struct {
    repo userRepo
}
```

**⚖️ Tradeoffs**
| | |
|---|---|
| ✅ Pros | Enables mocking, idiomatic, keeps interface minimal |
| ❌ Cons | Still adds a layer of indirection |
| 🎯 Best when | You want unit tests without a real DB |

---

### 💬 Recommendation
**Option 1** if you're early-stage or using integration tests.
**Option 2** if you need unit tests and want to mock the DB layer.
Don't keep the interface in the repository package — that's a Java habit, not Go.
```
