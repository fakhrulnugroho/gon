---
name: code-review
description: >
  Review Golang codebases that follow hexagonal architecture (ports & adapters).
  Use this skill whenever the user wants to: review Go code, audit a repository structure,
  check architectural compliance, find violations of hexagonal/clean architecture in Go,
  evaluate domain logic separation, assess adapter implementations, review port interfaces,
  check dependency direction, or get feedback on Go code quality in a layered architecture.
  Trigger even when the user pastes Go code snippets expecting architectural feedback,
  or asks things like "review my repo", "is my structure correct", or "what's wrong with this code".
---

# Code Review — Golang Hexagonal Architecture

This skill guides a thorough review of Go codebases using hexagonal architecture
(a.k.a. Ports & Adapters). Use it as a checklist and analysis guide, not as a rigid
template to follow word for word.

---

## 1. Codebase Orientation

Before diving into details, do a quick orientation:

```bash
# View directory structure
find . -type f -name "*.go" | head -60
tree -L 4 --dirsfirst   # if available

# Check module path and dependencies
cat go.mod

# Find entry points
ls cmd/ main.go
```

Identify the directory convention in use. Hexagonal architecture in Go typically follows
one of two conventions:

**Convention A — Explicit Ports & Adapters:**
```
internal/
  domain/          ← core business logic, entities, value objects
  ports/           ← interfaces (input ports & output ports)
    input/
    output/
  adapters/        ← concrete implementations
    primary/       ← driving adapters (HTTP, gRPC, CLI)
    secondary/     ← driven adapters (DB, queue, ext. service)
  application/     ← use cases / application services
cmd/               ← entry points (main.go per app)
```

**Convention B — Layer-based (more common in Go):**
```
internal/
  domain/          ← entities, value objects, domain errors
  repository/      ← output port interfaces + implementations
  service/         ← application services / use cases
  handler/         ← HTTP/gRPC handlers (primary adapters)
  infrastructure/  ← DB, cache, messaging implementations
cmd/
```

Note which convention is used, then continue the review based on that structure.

---

## 2. Dependency Rule Check (Most Critical)

The core rule of hexagonal architecture: **dependencies must point inward** (toward the domain), never outward.

```
Primary Adapters → Application/Use Cases → Domain ← Secondary Adapters (via ports)
```

### How to Check Dependency Direction

```bash
# Look for violations (domain importing outer layers)
grep -rn "import" internal/domain/ | grep -v "_test.go"

# Check if domain imports infrastructure or adapters
grep -rn '".*infrastructure' internal/domain/
grep -rn '".*handler'        internal/domain/
grep -rn '".*repository'     internal/domain/   # careful: interfaces ok, implementations not
grep -rn '".*adapter'        internal/domain/
```

**Red flags:**
- `domain/` imports `infrastructure/`, `handler/`, or `adapter/`
- `service/` imports `handler/` or an HTTP framework directly
- Use case knows it's backed by PostgreSQL, Redis, etc.

**Green flags:**
- `domain/` only imports the standard library or other domain packages
- `service/` only imports interfaces from `ports/` or `repository/`
- `handler/` imports `service/` but not `infrastructure/`

---

## 3. Domain Layer Review

Target files: `internal/domain/**`

### 3.1 Entities & Value Objects

```go
// GOOD: Entity with behavior, not just a data holder
type Order struct {
    id          OrderID
    items       []OrderItem
    status      OrderStatus
    totalAmount Money
}

func (o *Order) AddItem(item OrderItem) error { ... }  // behavior lives in the entity
func (o *Order) Confirm() error { ... }

// BAD: Anemic domain model — entity with no behavior
type Order struct {
    ID     int
    Status string
}
// all logic lives in the service, not the entity
```

Checklist:
- [ ] Entities have a clear identity (value objects vs entities are distinguished)
- [ ] Value objects are immutable
- [ ] Domain errors are defined in the domain layer (`ErrOrderNotFound`, etc.)
- [ ] No dependencies on DB, HTTP, or external libraries in the domain
- [ ] Business rules live in entities/value objects, not in services

### 3.2 Domain Events (if applicable)

```go
// GOOD: Domain event defined in the domain layer
type OrderConfirmed struct {
    OrderID   OrderID
    OccurredAt time.Time
}
```

---

## 4. Port Interface Review

Target files: `internal/ports/` or interfaces spread across `repository/`, `service/`

### 4.1 Input Ports (Driving Side)

Interfaces that define what the application can do from the outside.

```go
// GOOD: Input port defined at the application layer, not in the handler
type OrderService interface {
    CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*Order, error)
    GetOrder(ctx context.Context, id OrderID) (*Order, error)
}

// BAD: Handler depends on a concrete struct
type OrderHandler struct {
    svc *orderServiceImpl  // should use the interface instead
}
```

### 4.2 Output Ports (Driven Side)

Interfaces that the domain/application requires from infrastructure.

```go
// GOOD: Output port defined on the consumer side (domain/application), implemented in the adapter
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id OrderID) (*Order, error)
}

// BAD: Interface leaks implementation details
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id OrderID) (*Order, error)
    ExecSQL(query string, args ...interface{}) error  // ← leaks SQL detail
}
```

Checklist:
- [ ] Interfaces are defined on the consumer side (domain/application), not the implementor
- [ ] Interfaces do not leak technical details (SQL, HTTP, Redis commands, etc.)
- [ ] Interfaces are as small as possible — Interface Segregation Principle
- [ ] `context.Context` is the first parameter on all I/O operations

---

## 5. Application / Use Case Layer Review

Target files: `internal/service/`, `internal/application/`, `internal/usecase/`

```go
// GOOD: Use case orchestrates, contains no business logic
func (s *orderService) ConfirmOrder(ctx context.Context, id OrderID) error {
    order, err := s.repo.FindByID(ctx, id)   // uses output port
    if err != nil {
        return err
    }
    if err := order.Confirm(); err != nil {  // business rule stays in the entity
        return err
    }
    if err := s.repo.Save(ctx, order); err != nil {
        return err
    }
    s.eventBus.Publish(order.Events()...)    // domain events
    return nil
}

// BAD: Business logic leaks into the use case
func (s *orderService) ConfirmOrder(ctx context.Context, id OrderID) error {
    order, _ := s.repo.FindByID(ctx, id)
    if order.Status == "pending" && order.TotalAmount > 0 {  // ← business rule leaking here
        order.Status = "confirmed"
        s.repo.Save(ctx, order)
    }
    return nil
}
```

Checklist:
- [ ] Use cases only orchestrate — they contain no business rules
- [ ] All dependencies are injected via constructor (Constructor Injection)
- [ ] Return types use domain types, not DTOs or response structs
- [ ] Error wrapping is consistent (`fmt.Errorf("confirmOrder: %w", err)`)
- [ ] Transaction management lives in the use case, not in the repository

---

## 6. Adapter Review

### 6.1 Primary Adapters (HTTP Handler, gRPC, CLI)

Target files: `internal/handler/`, `internal/adapters/primary/`

```go
// GOOD: Thin handler, delegates to the use case
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respond(w, http.StatusBadRequest, err)
        return
    }
    cmd := req.toCommand()                           // mapping happens in the adapter layer
    order, err := h.svc.CreateOrder(r.Context(), cmd)
    if err != nil {
        h.handleError(w, err)
        return
    }
    h.respond(w, http.StatusCreated, toResponse(order))
}

// BAD: Fat handler containing business logic
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // business validation in the handler
    // data manipulation in the handler
    // direct DB access from the handler
}
```

Checklist:
- [ ] Handlers contain no business logic
- [ ] Request/response mapping happens in the handler (not in domain/service)
- [ ] Domain error → HTTP status code mapping happens in the handler
- [ ] Middleware (auth, logging, tracing) is separate from handler logic
- [ ] Dependency on use case is via interface, not concrete struct

### 6.2 Secondary Adapters (Repository, Messaging, External API)

Target files: `internal/infrastructure/`, `internal/adapters/secondary/`

```go
// GOOD: Clean repository implementation
type postgresOrderRepository struct {
    db *sql.DB
}

func (r *postgresOrderRepository) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
    row := r.db.QueryRowContext(ctx, "SELECT ... FROM orders WHERE id = $1", id)
    return r.scan(row)    // mapping from DB row to domain entity
}

// BAD: Leaks SQL to the layer above
func (r *postgresOrderRepository) FindByRawSQL(ctx context.Context, query string) (*domain.Order, error) { ... }
```

Checklist:
- [ ] Implementation fully satisfies the output port interface
- [ ] Mapping from DB model / external model to domain entity happens in the adapter
- [ ] No domain logic in the adapter (only translate + call infra)
- [ ] Infrastructure errors are wrapped into relevant domain errors
- [ ] DB structs (models) are separate from domain entities

---

## 7. Dependency Injection & Wiring

Target files: `cmd/`, `main.go`, or a dedicated `wire.go` / `container.go`

```go
// GOOD: All wiring in main/cmd, dependencies are explicit
func main() {
    db         := postgres.NewDB(cfg.Database)
    orderRepo  := postgres.NewOrderRepository(db)    // secondary adapter
    orderSvc   := service.NewOrderService(orderRepo) // use case
    orderHandler := http.NewOrderHandler(orderSvc)   // primary adapter

    router := chi.NewRouter()
    router.Post("/orders", orderHandler.CreateOrder)
    ...
}
```

Checklist:
- [ ] No `init()` functions hiding dependencies
- [ ] No global variables for dependencies (singletons via DI, not globals)
- [ ] Constructors accept interfaces, not concrete types (except in `cmd/`)
- [ ] If using a DI framework (wire, dig, fx) — confirm it doesn't leak into the domain layer

---

## 8. Error Handling Patterns

```go
// GOOD: Sentinel errors defined in the domain
var (
    ErrOrderNotFound  = errors.New("order not found")
    ErrOrderConfirmed = errors.New("order already confirmed")
)

// GOOD: Error wrapping with context
func (s *service) GetOrder(ctx context.Context, id OrderID) (*Order, error) {
    order, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("GetOrder %s: %w", id, err)
    }
    return order, nil
}

// GOOD: Error mapping in the handler
func (h *handler) handleError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, domain.ErrOrderNotFound):
        h.respond(w, http.StatusNotFound, err)
    case errors.Is(err, domain.ErrOrderConfirmed):
        h.respond(w, http.StatusConflict, err)
    default:
        h.respond(w, http.StatusInternalServerError, err)
    }
}
```

Checklist:
- [ ] Sentinel errors are defined in the domain layer
- [ ] Use `fmt.Errorf("...: %w", err)` for wrapping, not string concatenation
- [ ] Use `errors.Is` / `errors.As` for inspection, not string matching
- [ ] HTTP error mapping lives in the adapter layer, not in domain/service
- [ ] `panic` is reserved for truly unrecoverable state, not flow control

---

## 9. Testing Review

```bash
# Check coverage per package
go test ./... -cover

# Find test files
find . -name "*_test.go" | head -40
```

Checklist:
- [ ] Domain layer is tested without any mocks (pure unit tests)
- [ ] Use cases are tested with mocked output ports (testify/mock, gomock, or manual mocks)
- [ ] Handlers are tested with `httptest` — no real DB required
- [ ] Repository/adapters are tested with integration tests (real DB, testcontainers, etc.)
- [ ] Test naming follows: `TestFunctionName_Scenario_ExpectedResult`

```go
// GOOD: Domain test, no mocks needed
func TestOrder_Confirm_WhenPending_ShouldSucceed(t *testing.T) {
    order := NewOrder(...)
    err := order.Confirm()
    assert.NoError(t, err)
    assert.Equal(t, StatusConfirmed, order.Status())
}

// GOOD: Use case test with mock
func TestOrderService_ConfirmOrder_WhenOrderNotFound_ShouldReturnError(t *testing.T) {
    repo := &mockOrderRepo{}
    repo.On("FindByID", mock.Anything, orderID).Return(nil, domain.ErrOrderNotFound)
    svc := NewOrderService(repo)
    err := svc.ConfirmOrder(ctx, orderID)
    assert.ErrorIs(t, err, domain.ErrOrderNotFound)
}
```

---

## 10. Go-Specific Best Practices

### Naming & Package Structure

```go
// GOOD: Short, lowercase package names, no underscores
package order
package postgres
package http   // fine under handler/

// BAD
package orderService
package order_repository
```

Checklist:
- [ ] Package names are not redundant with the parent directory (`order/order.go` is fine, but `order.OrderService` becomes redundant)
- [ ] Go interfaces are typically 1–3 methods — if larger, consider splitting
- [ ] Exported names are self-explanatory without needing long comments
- [ ] Entity fields are unexported for proper encapsulation

### Context Propagation

```go
// GOOD: context is always the first parameter
func (r *repo) FindByID(ctx context.Context, id OrderID) (*Order, error)

// BAD: context stored in a struct
type Service struct { ctx context.Context }
```

### Goroutines & Concurrency

- [ ] Every launched goroutine has a clear owner
- [ ] Goroutine leaks are prevented — use `context.Done()` or `errgroup`
- [ ] Shared state is protected by a mutex or communicated via channels

---

## 11. Review Output Format

There are two output modes depending on what the user asks for.

---

### Mode A — Standard Code Review (default)

Use this when the user simply asks for a code review. Each finding must be **concise** and
cover exactly four points: **What, Why, How, Where**. No more, no less.

```
#### [CRITICAL|MAJOR|MINOR|SUGGESTION] <short title>

- **What:** One sentence describing the violation found.
- **Why:** One sentence explaining why this is a problem (architecture principle or risk).
- **How:** One or two sentences on how to fix it, with a brief inline code snippet if helpful.
- **Where:** Exact file path and line number(s). e.g. `internal/service/order_service.go:42-58`
```

**Example:**

```
#### [CRITICAL] Domain layer imports PostgreSQL driver

- **What:** `internal/domain/order.go` directly imports `"github.com/lib/pq"`.
- **Why:** The domain must not know about any infrastructure — this breaks the dependency rule and makes the core untestable in isolation.
- **How:** Define an `OrderRepository` interface in the domain/application layer and inject the concrete Postgres implementation from `cmd/main.go`.
- **Where:** `internal/domain/order.go:5`
```

**Severity guide:**
- **CRITICAL** — Violates the core dependency rule or risks data corruption
- **MAJOR** — Logic in the wrong layer, very low test coverage, goroutine/memory leak
- **MINOR** — Style, naming, minor separation of concerns issue
- **SUGGESTION** — Optional improvement, not urgent

After listing all findings, close with a short **Summary** (3–5 sentences max) covering:
overall architecture health, the most urgent item to fix, and any positive patterns worth keeping.

---

### Mode B — GitHub Issue (when user asks to create issues)

Use this when the user asks to turn findings into GitHub issues (e.g. "create issues for these",
"buat jadi github issue", "open issues"). Each finding becomes one issue with significantly
more detail than Mode A, enough for any engineer to pick it up and act on it independently.

**Issue body structure:**

```markdown
## Summary
One paragraph describing the problem, its location, and its impact on the codebase.

## Problem Details
Explain the violation in depth. Include the current code snippet that demonstrates the issue.

```go
// current problematic code with full context
```

## Why This Matters
Explain the architectural principle being violated, the concrete risk (e.g. "breaks testability",
"couples domain to PostgreSQL", "prevents swapping implementations"), and any cascading effects
on other layers.

## Recommended Fix
Step-by-step instructions for the engineer. Include the target code after the fix.

```go
// recommended code after the fix
```

If the fix involves moving code across files, list each file and the change needed.

## Acceptance Criteria
- [ ] Concrete, checkable condition 1
- [ ] Concrete, checkable condition 2
- [ ] Unit/integration test added or updated to cover the fix
- [ ] No regression in existing tests (`go test ./...` passes)

## References
- Relevant hexagonal architecture principle or Go best practice (with link if applicable)
- Related files or packages beyond the primary location
```

**GitHub CLI command to create each issue:**

```bash
gh issue create \
  --title "[CRITICAL] Domain layer imports PostgreSQL driver" \
  --label "architecture,technical-debt" \
  --body "$(cat <<'EOF'
## Summary
...

## Problem Details
...
EOF
)"
```

**Labels to use** (create them first if they don't exist):
- `architecture` — hexagonal/clean architecture violations
- `technical-debt` — code that needs refactoring
- `bug` — if the finding can cause incorrect runtime behavior
- `testing` — missing or wrong test strategy
- `suggestion` — optional improvements

**Batch creation tip:** Generate all issue commands in a single shell script so the user can
run them at once:

```bash
#!/bin/bash
# review-issues.sh — generated by code-review skill

gh issue create --title "[CRITICAL] ..." --label "architecture,technical-debt" --body "..."
gh issue create --title "[MAJOR] ..."   --label "technical-debt" --body "..."
gh issue create --title "[MINOR] ..."   --label "technical-debt" --body "..."
```

---

## 12. Quick Reference — Common Violations

| Violation | Example | Fix |
|-----------|---------|-----|
| Domain imports infra | `domain/` imports `"database/sql"` | Use an interface, inject from outside |
| Fat handler | HTTP handler contains business validation | Move to entity/use case |
| Anemic domain | Entity is only getters/setters | Add behavior to the entity |
| Leaky abstraction | Repository interface exposes SQL | Use semantic methods (`FindByStatus`) |
| DI inside domain | `service.New()` creates its own dependencies | Wire everything in `cmd/main.go` |
| Global state | `var db *sql.DB` at package level | Inject via constructor |
| No error wrapping | `return err` with no context | `fmt.Errorf("op: %w", err)` |
| Service knows HTTP | Service returns `http.StatusCode` | Return domain error, map in handler |
