# NebularCore - GitHub Copilot Context

## Project Overview

**NebularCore** is a modular Go backend framework inspired by [PocketBase](https://pocketbase.io). It provides a robust foundation for building scalable backend services with features like authentication, health monitoring, database migrations, multi-tenancy, and event-driven architecture.

- **Language**: Go 1.24+
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin)
- **ORM**: [GORM](https://gorm.io)
- **Database Support**: PostgreSQL, SQLite, CloudSQL (Postgres)
- **Migration Tool**: [golang-migrate/migrate](https://github.com/golang-migrate/migrate)
- **Event System**: [Watermill](https://github.com/ThreeDotsLabs/watermill)
- **Authorization**: [Casbin](https://casbin.org/)
- **CLI**: [Cobra](https://github.com/spf13/cobra)

---

## Architecture Overview

### Core Application (`core/`)

The framework follows a modular architecture with a central `App` interface:

```
core/
├── app.go          # App interface definition
├── base.go         # Default App implementation (baseApp)
├── config/         # Configuration loading and validation
├── migration_runner/ # Database migration handling
├── module/         # Module system (interface + registry)
├── repository/     # Generic repository pattern
└── utils/          # Utility functions
```

#### Key Types

- **`App[T config.Settings]`** - Generic interface for the application core
- **`baseApp[T]`** - Default implementation handling bootstrap, shutdown, module registration
- **`Options[T]`** - Configuration options (ConfigPath, EnvPrefix)

### Module System (`core/module/`)

All features are implemented as modules that implement the `Module` interface:

```go
type Module interface {
    Name() string
    Version() string
    Dependencies() []string
    Namespace() ModuleNamespace  // PublicNamespace or TenantNamespace

    Initialize(ctx context.Context, db *gorm.DB, router *gin.Engine) error
    NewConfig() any
    Configure(config any) error
    Shutdown(ctx context.Context) error

    ProvidesMigrations() bool
    MigrationsDir() string
    GetMigrationSources(projectRoot string) []migration_runner.Source
}
```

#### Module Namespaces
- **`PublicNamespace`** - Modules operating in the public schema
- **`TenantNamespace`** - Modules operating in tenant-specific schemas

#### Module Registry
- Handles dependency resolution and initialization order
- Separate tracking for public and tenant modules
- Topological sort for dependencies

---

## Built-in Modules (`modules/`)

### Auth Module (`modules/auth/`)
- JWT-based authentication
- Multiple backends: local, Google OAuth
- OTP support
- Password management with bcrypt
- Social authentication (Google, Apple, Facebook)
- Event emission for auth events
- Embedded SQL migrations via `//go:embed`

### Event Module (`modules/event/`)
- Pub/sub event system using Watermill
- Async event publishing
- Handler registration with routing

### Health Module (`modules/health/`)
- System health checks
- Database connectivity checks
- Periodic health monitoring

### Storage Module (`modules/storage/`)
- Multi-provider storage abstraction
- Supports: Local, S3, GCS
- Unified API for file operations

### Tenant Module (`modules/tenant/`)
- Multi-tenancy support
- Schema-per-tenant isolation
- GORM callbacks for automatic tenant scoping
- Middleware for tenant extraction from headers

### Control Center Module (`modules/controlcenter/`)
- Administrative endpoints

---

## Migration System (`core/migration_runner/`)

Custom migration runner built on top of golang-migrate with advanced features:

### Source Types
- **`embedSource`** - Reads migrations from embedded `embed.FS`
- **`filteredSource`** - Wraps a source and excludes specific migrations
- **`chainedSource`** - Combines multiple sources with primary/fallback pattern

### Source Configuration
```go
type Source struct {
    Path     string    // Migration path or identifier
    Priority int       // Higher priority = checked first
    Exclude  []string  // Migration files to exclude
    FS       fs.FS     // Optional embedded filesystem
}
```

### Migration Flow
1. Sources are sorted by priority (descending)
2. Each source can be filtered for exclusions
3. Sources are chained: primary → fallback
4. Each module has its own migrations table: `schema_migrations_{module_name}`

---

## Configuration System (`core/config/`)

YAML-based configuration with validation:

```yaml
core:
  environment: development|staging|production|test
  debug: false
  database:
    driver: postgres|sqlite|cloudsqlpostgres
    host: localhost
    port: "5432"
    name: dbname
    username: user
    password: pass
    sslmode: disable
  server:
    host: localhost
    port: "8080"
    readTimeout: 30s
    writeTimeout: 30s
    shutdownTimeout: 10s

modules:
  auth:
    # auth-specific config
  storage:
    provider: local|s3|gcs
    # provider-specific config

project:
  # custom project settings (implements Settings interface)
```

### Settings Interface
```go
type Settings interface {
    Validate() error
    IsProduction() bool
}
```

---

## Repository Pattern (`core/repository/`)

Generic repository with GORM:

```go
type Repository[T any] interface {
    Create(ctx context.Context, model *T) error
    First(ctx context.Context, dest *T, conds ...interface{}) error
    Find(ctx context.Context, dest *[]T, conds ...interface{}) error
    Update(ctx context.Context, model *T, attrs interface{}) error
    Delete(ctx context.Context, model *T) error
    Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
    Query(ctx context.Context) *gorm.DB
    DB() *gorm.DB
}
```

---

## CLI Commands (`cmd/`)

### Serve Command
- Bootstraps the application
- Starts the HTTP server
- Handles graceful shutdown

### Migrate Command
- `migrate create [module] [name]` - Create new migration file
- `migrate up` - Run all pending migrations for all modules

---

## Code Conventions

### Error Handling
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`
- Use zerolog for logging: `log.Err(err).Msg("message")`

### Module Implementation Pattern
```go
type Module struct {
    name    string
    version string
    config  *Config
    // ... dependencies
}

func New() *Module {
    return &Module{
        name:    "modulename",
        version: "1.0.0",
        config:  DefaultConfig(),
    }
}

// Implement all Module interface methods...
```

### Factory Pattern for Extensible Models
- Used in auth module for custom user models
- `UserFactory` interface allows projects to define custom user types

### API Response Format
```go
type ApiResponsePayload struct {
    Status  bool     `json:"status"`
    Message string   `json:"message"`
    Data    any      `json:"data,omitempty"`
    Errors  []string `json:"errors,omitempty"`
}
```

---

## Testing

- Test files follow `*_test.go` convention
- Mock files in `mocks/` subdirectories
- Test data in `testdata/` directories
- Uses `testify` for assertions

---

## Directory Structure Summary

```
nebularcore/
├── nebularcore.go      # Main entry point, NebularCore[T] wrapper
├── go.mod              # Go module definition
├── cmd/                # CLI commands (serve, migrate)
├── core/               # Framework core
│   ├── app.go          # App interface
│   ├── base.go         # App implementation
│   ├── config/         # Configuration system
│   ├── migration_runner/ # Migration system
│   ├── model/          # Base models
│   ├── module/         # Module system
│   ├── repository/     # Repository pattern
│   └── utils/          # Utilities
├── modules/            # Built-in modules
│   ├── auth/           # Authentication
│   ├── controlcenter/  # Admin endpoints
│   ├── event/          # Event system
│   ├── health/         # Health checks
│   ├── storage/        # File storage
│   └── tenant/         # Multi-tenancy
├── models/             # Shared models
│   └── responses/      # API response types
├── tools/              # Utility packages
│   ├── auth/           # Auth providers
│   ├── common/         # Reflection helpers
│   ├── filesystem/     # FS utilities
│   ├── handlers/       # Error handlers
│   ├── httpclient/     # HTTP client
│   ├── security/       # JWT, OTP, passwords
│   ├── test/           # Test utilities
│   ├── types/          # Common types
│   └── validation/     # Input validation
├── errors/             # Error definitions (empty)
├── docs/               # Documentation
└── examples/           # Example project
```

---

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/gin-gonic/gin` | HTTP router and middleware |
| `gorm.io/gorm` | ORM |
| `github.com/golang-migrate/migrate/v4` | Database migrations |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/rs/zerolog` | Structured logging |
| `github.com/ThreeDotsLabs/watermill` | Event-driven architecture |
| `github.com/casbin/casbin/v2` | Authorization |
| `github.com/golang-jwt/jwt/v4` | JWT tokens |
| `github.com/pquerna/otp` | OTP generation |
| `github.com/stretchr/testify` | Testing assertions |

---

## Common Patterns

### Creating a New Module
1. Create directory under `modules/`
2. Implement `Module` interface
3. Add configuration struct if needed
4. Add migrations in `migrations/` subdirectory (use `//go:embed`)
5. Register with `app.RegisterModule()`

### Adding Migrations
1. Use `migrate create [module] [name]` CLI
2. Creates numbered `.up.sql` and `.down.sql` files
3. Migrations are automatically discovered via `GetMigrationSources()`

### Extending Auth with Custom User Model
1. Create custom user struct embedding base fields
2. Implement `UserFactory` interface
3. Create custom repository
4. Pass to auth module: `auth.New(eventBus).WithUserRepository(repo)`

---

## Important Notes

- **Generic Types**: The framework uses Go generics extensively (`App[T config.Settings]`)
- **Embedded Migrations**: Modules embed migrations using `//go:embed migrations/*.sql`
- **Module Dependencies**: Declare dependencies via `Dependencies()` for proper init order
- **Tenant Isolation**: Use `TenantNamespace` for tenant-scoped modules
- **Event-Driven**: Use the event module for decoupled communication between modules
