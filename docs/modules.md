# NebularCore Module System

NebularCore uses a modular architecture that allows you to create, configure, and register custom modules. This guide explains how to create and integrate modules into your NebularCore application.

## Core Concepts

### 1. Module Interface

Every module must implement the `Module` interface:

```go
type Module interface {
    Name() string
    Version() string
    Dependencies() []string
    Initialize(app core.App) error
    Configure(config interface{}) error
    Shutdown() error
}
```

### 2. Module Configuration

Each module can define its own configuration structure:

```go
type ModuleConfig struct {
    // Module-specific configuration fields
    Setting1 string `yaml:"setting1"`
    Setting2 int    `yaml:"setting2"`
    
    // Optional embedded configs
    DatabaseConfig `yaml:"database,omitempty"`
    APIConfig     `yaml:"api,omitempty"`
}
```

## Creating a Module

### 1. Basic Module Structure

```go
// modules/mymodule/module.go
package mymodule

type Module struct {
    app    core.App
    config *Config
}

func NewModule() *Module {
    return &Module{}
}

func (m *Module) Name() string {
    return "mymodule"
}

func (m *Module) Version() string {
    return "1.0.0"
}

func (m *Module) Dependencies() []string {
    return []string{"auth"} // List module dependencies
}

func (m *Module) Initialize(app core.App) error {
    m.app = app
    return nil
}

func (m *Module) Configure(config interface{}) error {
    if cfg, ok := config.(*Config); ok {
        m.config = cfg
        return nil
    }
    return fmt.Errorf("invalid configuration type")
}

func (m *Module) Shutdown() error {
    // Cleanup resources
    return nil
}
```

### 2. Module Configuration

```go
// modules/mymodule/config/schema.go
package config

type Config struct {
    // Module configuration
    Setting1 string `yaml:"setting1" validate:"required"`
    Setting2 int    `yaml:"setting2" default:"100"`
    
    // Optional database config
    Database struct {
        Host     string `yaml:"host"`
        Port     string `yaml:"port"`
        Username string `yaml:"username"`
        Password string `yaml:"password"`
    } `yaml:"database,omitempty"`
}
```

### 3. Module Routes

```go
// modules/mymodule/routes.go
package mymodule

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
    group := r.Group("/mymodule")
    {
        group.GET("/status", m.handleStatus)
        group.POST("/items", m.handleCreateItem)
        group.GET("/items/:id", m.handleGetItem)
    }
}

func (m *Module) handleStatus(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "ok",
        "version": m.Version(),
    })
}
```

### 4. Module Services

```go
// modules/mymodule/service.go
package mymodule

type Service struct {
    db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
    return &Service{db: db}
}

func (s *Service) DoSomething() error {
    // Business logic
    return nil
}
```

## Using a Module

### 1. Register Module

```go
package main

import (
    "github.com/volvlabs/nebularcore/core"
    "github.com/volvlabs/nebularcore/modules/mymodule"
)

func main() {
    app, err := core.NewApp("config.yml")
    if err != nil {
        log.Fatal(err)
    }

    // Register module
    app.RegisterModule(mymodule.NewModule())

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Configure Module

```yaml
# config.yml
core:
  environment: development
  database:
    driver: postgres
    host: localhost
    port: "5432"

modules:
  mymodule:
    setting1: "value1"
    setting2: 200
    database:
      host: localhost
      port: "5432"
      username: postgres
      password: password
```

## Best Practices

1. **Dependency Management**
   - Clearly define module dependencies
   - Use semantic versioning
   - Handle dependency conflicts gracefully

2. **Configuration**
   - Use validation tags for required fields
   - Provide sensible defaults
   - Document all configuration options

3. **Error Handling**
   - Return descriptive errors
   - Log important operations
   - Handle cleanup in Shutdown()

4. **Testing**
   ```go
   func TestModule(t *testing.T) {
       module := NewModule()
       
       // Test configuration
       cfg := &Config{
           Setting1: "test",
           Setting2: 100,
       }
       
       err := module.Configure(cfg)
       assert.NoError(t, err)
       
       // Test initialization
       app := mock.NewApp()
       err = module.Initialize(app)
       assert.NoError(t, err)
   }
   ```

## Module Lifecycle

1. **Registration**: Module is registered with the application
2. **Configuration**: Module config is parsed and validated
3. **Dependencies**: Dependencies are checked and resolved
4. **Initialization**: Module is initialized with app context
5. **Running**: Module is active and handling requests
6. **Shutdown**: Module performs cleanup when app stops

## Common Patterns

### 1. Database Migrations

```go
func (m *Module) GetMigrationSources(projectRoot string) []migrationRunner.Source {
    return []migrationRunner.Source{
        {
            Path:     fmt.Sprintf("file://%s/modules/mymodule/migrations", projectRoot),
            Priority: 50,
        },
    }
}
```

### 2. Middleware

```go
func (m *Module) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Module-specific middleware logic
        c.Next()
    }
}
```

### 3. Event Handling

```go
func (m *Module) HandleEvent(event core.Event) error {
    switch event.Type {
    case "user.created":
        return m.handleUserCreated(event)
    case "data.updated":
        return m.handleDataUpdated(event)
    }
    return nil
}
```
