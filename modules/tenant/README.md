# Tenant Module

The tenant module provides multi-tenancy support for NebularCore applications using PostgreSQL schema-based isolation. Each tenant gets its own PostgreSQL schema, ensuring data isolation between tenants.

## Features

- Schema-based tenant isolation
- Tenant-aware repositories for data access
- Automatic schema creation and management
- Middleware for tenant context injection
- Support for tenant-bound models

## Installation

The tenant module is included in NebularCore by default. No additional installation is required.

## Usage

### 1. Module Registration

While the tenant module can be used standalone, registering it as a NebularCore module provides several important benefits:

- **Lifecycle Management**: Proper initialization, shutdown, and resource cleanup
- **Configuration Management**: Automatic loading and validation of tenant settings
- **Dependency Resolution**: Ensures correct module initialization order
- **Migration Control**: Automated schema creation and migration sequencing
- **Integration Points**: Event hooks, middleware registration, and service discovery

To register the tenant module:

```go
package main

import (
    "gitlab.com/jideobs/nebularcore/core"
    "gitlab.com/jideobs/nebularcore/modules/tenant"
)

func main() {
    app := core.NewApp()
    
    // Register tenant module
    tenantModule := tenant.NewModule()
    app.RegisterModule(tenantModule)
    
    // Initialize app after registering all modules
    if err := app.Initialize(); err != nil {
        panic(err)
    }
}
```

### 2. Configuration

Add tenant configuration to your config file:

```yaml
tenant:
  default_schema: "public"  # Default schema for non-tenant operations
  migrations_dir: "modules/tenant/migrations"  # Directory containing tenant migrations
```

### 3. Creating Tenant-Bound Models

Models that should be tenant-aware must implement the `model.TenantBound` interface:

```go
package myapp

import "gitlab.com/jideobs/nebularcore/core/model"

type MyModel struct {
    ID        uint   `gorm:"primarykey"`
    Name      string
    TenantID  string
}

// Implement TenantBound interface
func (MyModel) IsTenantBound() bool {
    return true
}
```

### 4. Creating Tenant-Aware Repositories

Use the `TenantAwareRepository` to create repositories for tenant-bound models:

```go
package myapp

import (
    "gitlab.com/jideobs/nebularcore/modules/tenant"
    "gorm.io/gorm"
)

type MyModelRepository struct {
    *tenant.TenantAwareRepository[MyModel]
}

func NewMyModelRepository(db *gorm.DB, schema string) *MyModelRepository {
    return &MyModelRepository{
        TenantAwareRepository: tenant.NewTenantAware[MyModel](db, schema),
    }
}
```

### 5. Using Tenant-Aware Repositories

```go
func (r *MyModelRepository) CreateModel(ctx context.Context, model *MyModel) error {
    return r.Create(ctx, model)
}

func (r *MyModelRepository) FindModels(ctx context.Context) ([]MyModel, error) {
    var models []MyModel
    err := r.Find(ctx, &models)
    return models, err
}
```

### 6. Middleware Usage

Add the tenant middleware to your HTTP routes to automatically extract tenant information:

```go
package main

import (
    "github.com/gin-gonic/gin"
    "gitlab.com/jideobs/nebularcore/modules/tenant"
)

func setupRoutes(r *gin.Engine) {
    // Add tenant middleware
    r.Use(tenant.Middleware())
    
    // Your routes here
    r.GET("/api/v1/models", handleGetModels)
}
```

### 7. Migrations

Create tenant-specific migrations in your migrations directory:

```sql
-- 001_tenant.up.sql
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    schema_name VARCHAR(50) UNIQUE NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Function to create tenant schemas
CREATE OR REPLACE FUNCTION public.create_tenant_schema(schema_name text)
RETURNS void AS $$
BEGIN
    EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);
END;
$$ LANGUAGE plpgsql;

-- 001_tenant.down.sql
DROP FUNCTION IF EXISTS public.create_tenant_schema(text);
DROP TABLE IF EXISTS tenants;
```

### 8. Accessing Current Tenant

Get tenant information from the context in your handlers:

```go
func handleGetModels(c *gin.Context) {
    tenantID, tenantCode, schema, ok := tenant.GetTenantFromContext(c)
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No tenant context"})
        return
    }
    
    // Use the tenant information
    repo := NewMyModelRepository(db, schema)
    models, err := repo.FindModels(c)
    // ...
}
```

## Best Practices

1. **Always Use Context**: Pass the context through all repository operations to maintain tenant scoping.
2. **Model Design**: Keep tenant-specific fields (like TenantID) in your models for additional validation.
3. **Schema Management**: Use the provided schema creation functions instead of creating schemas manually.
4. **Error Handling**: Always check for tenant context before performing database operations.
5. **Testing**: Test your repositories with different tenant contexts to ensure proper isolation.

## Dependencies

- PostgreSQL 9.6 or later
- GORM v1.25.0 or later
- Gin Web Framework

## Standalone Usage

While registering as a NebularCore module is recommended, you can use the tenant module independently:

```go
package main

import (
    "context"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "gitlab.com/jideobs/nebularcore/modules/tenant"
)

func main() {
    // Setup your own DB connection
    db, err := gorm.Open(/* your db config */)
    if err != nil {
        panic(err)
    }

    // Create repositories directly
    repo := tenant.NewTenantAware[MyModel](db, "tenant_schema")

    // Use middleware manually
    r := gin.New()
    r.Use(tenant.Middleware())

    // Handle tenant operations yourself
    r.POST("/api/models", func(c *gin.Context) {
        model := &MyModel{}
        if err := repo.Create(c, model); err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, model)
    })
}
```

However, you'll need to handle:
- Schema migrations manually
- Tenant lifecycle events
- Configuration management
- Resource cleanup
- Module dependencies

## Contributing

Please refer to the NebularCore contribution guidelines when contributing to this module.
