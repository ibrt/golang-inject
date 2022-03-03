# golang-inject
[![Go Reference](https://pkg.go.dev/badge/github.com/ibrt/golang-inject.svg)](https://pkg.go.dev/github.com/ibrt/golang-inject)
![CI](https://github.com/ibrt/golang-inject/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/ibrt/golang-inject/branch/main/graph/badge.svg?token=BQVP881F9Z)](https://codecov.io/gh/ibrt/golang-inject)

Tiny context-based dependency injection framework for Go.

### Basic Example

This basic example contains three modules illustrating how to use the dependency framework in a project.

#### Config Module

```go
package config

import (
    "context"

    "github.com/ibrt/golang-inject/injectz"
)

type contextKey int

const (
    cfgContextKey contextKey = iota
)

type Config struct {
    DatabaseURL string
}

func Initializer(ctx context.Context) (injectz.Injector, injectz.Releaser) {
    cfg := &Config{
        DatabaseURL: "...", // e.g. read from env var
    }
    return NewSingletonInjector(cfg), injectz.NewNoopReleaser()
}

func NewSingletonInjector(cfg *Config) injectz.Injector {
    return injectz.NewSingletonInjector(cfgContextKey, cfg)
}

func Get(ctx context.Context) *Config {
    return ctx.Value(cfgContextKey).(*Config)
}
```

#### Database Module

```go
package database

import (
    "context"
    "database/sql"

    "project/modules/config"
	
    "github.com/ibrt/golang-inject/injectz"
)

type contextKey int

const (
    dbContextKey contextKey = iota
)

var (
    _ Database = &databaseImpl{}
)

type User struct {
    ID        string
    FirstName string
    LastName  string
}

type Database interface {
    GetUser(ctx context.Context, id string) (*User, error)
}

type databaseImpl struct {
    sqlDB *sql.DB
}

func (d *databaseImpl) GetUser(ctx context.Context, id string) (*User, error) {
    // e.g. run a query using d.sqlDB
    return &User{ID: id}, nil
}

func Initializer(ctx context.Context) (injectz.Injector, injectz.Releaser) {
    sqlDB, err := sql.Open("mysql", config.Get(ctx).DatabaseURL)
    if err != nil {
        panic(err)
    }

    db := &databaseImpl{
        sqlDB: sqlDB,
    }

    return NewSingletonInjector(db), injectz.NewCloseReleaser(sqlDB)
}

func NewSingletonInjector(db Database) injectz.Injector {
    return injectz.NewSingletonInjector(dbContextKey, db)
}

func Get(ctx context.Context) Database {
    return ctx.Value(dbContextKey).(Database)
}
```

#### Request ID Module

```go
package request

import (
    "context"
    "math/rand"

    "github.com/ibrt/golang-inject/injectz"
)

type contextKey int

const (
    requestIDContextKey contextKey = iota
)

func Initializer(ctx context.Context) (injectz.Injector, injectz.Releaser) {
    injector := func(ctx context.Context) context.Context {
        return context.WithValue(ctx, requestIDContextKey, newID())
    }
    return injector, injectz.NewNoopReleaser()
}

func newID() string {
    const length = 8
    const chars = "abcdefghijklmnopqrstuvwxyz1234567890"
    b := make([]byte, length)
    _, _ = rand.Read(b)
    for i := 0; i < length; i++ {
        b[i] = chars[int(b[i])%len(chars)]
    }
    return string(b)
}

func Get(ctx context.Context) string {
    return ctx.Value(requestIDContextKey).(string)
}
```

#### Main

```go
package main

import (
    "log"
    "net/http"

    "project/modules/config"
    "project/modules/database"
    "project/modules/request"
	
    "github.com/ibrt/golang-inject/injectz"
)

func main() {
    injector, releaser := injectz.Initialize(
        config.Initializer, 
        database.Initializer, 
        request.Initializer)
    defer releaser()

    middleware := injectz.NewMiddleware(injector)
    mux := http.NewServeMux()
    mux.Handle("/", middleware(http.HandlerFunc(handler)))
    _ = http.ListenAndServe(":3000", mux)
}

func handler(w http.ResponseWriter, r *http.Request) {
    // use the request ID module (note that the injector generates a new ID each time)
    requestID := request.Get(r.Context())
    log.Print(requestID, " ", r.URL.String())

    // use the database module
    user, err := database.Get(r.Context()).GetUser(r.Context(), "some-id")
    // ...

    _, _ = w.Write([]byte("response"))
}
```

### Advanced Techniques

This section describes some advanced techniques that improve module APIs.

#### Context Passthrough

As you can see in the example above, the API for using the Database module is a bit cumbersome because the context needs
to be passed twice. A possible improvement - to be used with care - is to cache the context passed to get and use it
whenever a method on Database is called instead.

```go
package database

// ...

type DatabaseWithContext interface {
    GetUser(id string) (*User, error)
}

type databaseWithContextImpl struct {
    ctx context.Context
    db  Database
}

func (d *databaseWithContextImpl) GetUser(id string) (*User, error) {
    return d.db.GetUser(d.ctx, id)
}

// ...

func GetCtx(ctx context.Context) DatabaseWithContext {
    return &databaseWithContextImpl{
        ctx: ctx,
        db:  Get(ctx),
    }
}

// Old API: database.Get(ctx).GetUser(ctx, "id")
// New API: database.GetCtx(ctx).GetUser("id")
```

### Developers

Contributions are welcome, please check in on proposed implementation before sending a PR. You can validate your changes using the `./test.sh` script.
