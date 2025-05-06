# Multi-Tenant Migration Guide

This guide explains how to handle migrations in a multi-tenant environment using NebularCore.

## Migration Types

1. **Public Schema Migrations**
   - Core system tables (users, roles, permissions)
   - Tenant management tables
   - Global configuration tables

2. **Tenant Schema Migrations**
   - Tenant-specific data tables
   - Tenant-scoped application tables

## Directory Structure

```
your-project/
├── migrations/
│   ├── public/
│   │   ├── 001_init.up.sql
│   │   ├── 001_init.down.sql
│   │   ├── 002_users.up.sql
│   │   └── 002_users.down.sql
│   └── tenant/
│       ├── 001_tenant_tables.up.sql
│       ├── 001_tenant_tables.down.sql
│       ├── 002_tenant_data.up.sql
│       └── 002_tenant_data.down.sql
├── modules/
│   ├── auth/
│   │   └── migrations/
│   │       ├── public/
│   │       │   └── 001_auth_tables.up.sql
│   │       └── tenant/
│   │           └── 001_tenant_auth.up.sql
│   └── tenant/
│       └── migrations/
│           └── public/
│               └── 001_tenant.up.sql
```

## Migration Configuration

```yaml
migrations:
  # Global migration settings
  auto_migrate: true
  version_table: "schema_migrations"
  
  # Public schema settings
  public:
    enabled: true
    paths:
      - "migrations/public"
      - "modules/*/migrations/public"
  
  # Tenant schema settings
  tenant:
    enabled: true
    paths:
      - "migrations/tenant"
      - "modules/*/migrations/tenant"
```

## Migration Order

1. **Public Schema First**
   ```sql
   -- migrations/public/001_init.up.sql
   CREATE TABLE IF NOT EXISTS tenants (
       id UUID PRIMARY KEY,
       code VARCHAR(50) UNIQUE,
       schema_name VARCHAR(50) UNIQUE
   );

   -- Create schema management functions
   CREATE OR REPLACE FUNCTION create_tenant_schema(schema_name text)
   RETURNS void AS $$
   BEGIN
       EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', schema_name);
   END;
   $$ LANGUAGE plpgsql;
   ```

2. **Module Public Migrations**
   ```sql
   -- modules/auth/migrations/public/001_auth_tables.up.sql
   CREATE TABLE IF NOT EXISTS users (
       id UUID PRIMARY KEY,
       email VARCHAR(255) UNIQUE
   );
   ```

3. **Tenant Schema Creation**
   ```sql
   -- Automatically created for each tenant
   SELECT create_tenant_schema('tenant_1');
   SELECT create_tenant_schema('tenant_2');
   ```

4. **Tenant-Specific Migrations**
   ```sql
   -- migrations/tenant/001_tenant_tables.up.sql
   -- This runs for EACH tenant schema
   CREATE TABLE IF NOT EXISTS products (
       id UUID PRIMARY KEY,
       name VARCHAR(255),
       tenant_id UUID NOT NULL
   );
   ```

## Usage Examples

### 1. Running All Migrations

```go
package main

import (
    "gitlab.com/jideobs/nebularcore/core"
    "gitlab.com/jideobs/nebularcore/modules/tenant"
)

func main() {
    app := core.NewApp()
    
    // Register modules
    tenantModule := tenant.NewModule()
    app.RegisterModule(tenantModule)
    
    // Run migrations
    if err := app.Migrate(); err != nil {
        panic(err)
    }
}
```

### 2. Running Project-Specific Migrations

```go
package main

import (
    "context"
    "gitlab.com/jideobs/nebularcore/core/migration"
)

func main() {
    // Create migrator
    migrator := migration.NewMigrator(db)
    
    // Run public schema migrations
    if err := migrator.RunPublic(ctx, "migrations/public"); err != nil {
        panic(err)
    }
    
    // Get all tenants
    tenants, err := tenantRepo.FindAll(ctx)
    if err != nil {
        panic(err)
    }
    
    // Run tenant migrations for each tenant
    for _, t := range tenants {
        if err := migrator.RunTenant(ctx, t.SchemaName, "migrations/tenant"); err != nil {
            panic(err)
        }
    }
}
```

### 3. Creating a New Tenant with Migrations

```go
func CreateTenant(ctx context.Context, code string) error {
    // 1. Create tenant record in public schema
    tenant := &tenant.Tenant{
        Code: code,
        SchemaName: fmt.Sprintf("tenant_%s", code),
    }
    if err := tenantRepo.Create(ctx, tenant); err != nil {
        return err
    }
    
    // 2. Create tenant schema
    if err := db.Exec(fmt.Sprintf("CREATE SCHEMA %s", tenant.SchemaName)).Error; err != nil {
        return err
    }
    
    // 3. Run all tenant migrations
    migrator := migration.NewMigrator(db)
    paths := []string{
        "migrations/tenant",
        "modules/*/migrations/tenant",
    }
    
    for _, path := range paths {
        if err := migrator.RunTenant(ctx, tenant.SchemaName, path); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 4. Module-Specific Migrations

```go
// modules/auth/module.go
func (m *Module) GetMigrations() []migration.Migration {
    return []migration.Migration{
        // Public schema migrations
        {
            Name: "001_auth_tables",
            Path: "modules/auth/migrations/public",
            Type: migration.TypePublic,
        },
        // Tenant schema migrations
        {
            Name: "001_tenant_auth",
            Path: "modules/auth/migrations/tenant",
            Type: migration.TypeTenant,
        },
    }
}
```

## Best Practices

1. **Migration Versioning**
   - Use sequential version numbers (001, 002, etc.)
   - Include descriptive names
   - Always create down migrations
   - Use transactions where possible

2. **Schema Management**
   - Keep tenant schemas identical
   - Use schema creation functions
   - Handle schema deletion carefully
   - Backup before migrations

3. **Module Organization**
   - Separate public/tenant migrations
   - Use clear directory structure
   - Document dependencies
   - Version control all migrations

4. **Error Handling**
   - Roll back failed migrations
   - Log all migration operations
   - Handle partial failures
   - Implement retry mechanisms

5. **Performance**
   - Batch similar migrations
   - Use appropriate indexes
   - Consider migration timing
   - Monitor migration duration

## Single-Schema Projects (No Multi-tenancy)

If your project doesn't require multi-tenancy, you can use a simpler setup with just the public schema:

### Directory Structure
```
your-project/
├── migrations/
│   ├── 001_init.up.sql
│   ├── 001_init.down.sql
│   ├── 002_users.up.sql
│   └── 002_users.down.sql
└── modules/
    └── auth/
        └── migrations/
            ├── 001_auth_tables.up.sql
            └── 001_auth_tables.down.sql
```

### Configuration
```yaml
migrations:
  auto_migrate: true
  version_table: "schema_migrations"
  paths:
    - "migrations"
    - "modules/*/migrations"
```

### Running Migrations
```go
package main

import (
    "context"
    "gitlab.com/jideobs/nebularcore/core"
    "gitlab.com/jideobs/nebularcore/core/migration"
)

func main() {
    app := core.NewApp()
    
    // No need to register tenant module
    if err := app.Initialize(); err != nil {
        panic(err)
    }
    
    // Run all migrations in public schema
    if err := app.Migrate(); err != nil {
        panic(err)
    }
}
```

### Using Base Repository
```go
package myapp

import (
    "gitlab.com/jideobs/nebularcore/core/repository"
    "gorm.io/gorm"
)

type User struct {
    ID   uint
    Name string
}

type UserRepository struct {
    *repository.BaseRepository[User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{
        BaseRepository: repository.NewBase[User](db),
    }
}
```

### Example Migration
```sql
-- migrations/001_init.up.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes
CREATE INDEX idx_users_name ON users(name);
```

### Benefits of Single-Schema
- Simpler database structure
- No schema management overhead
- Easier backup and restore
- Better query performance
- Simpler deployment process

## Common Scenarios

1. **Adding a New Table**
   ```sql
   -- If public table:
   -- migrations/public/003_settings.up.sql
   CREATE TABLE public.settings (...);
   
   -- If tenant table:
   -- migrations/tenant/003_products.up.sql
   CREATE TABLE products (...);
   ```

2. **Modifying Existing Tables**
   ```sql
   -- migrations/tenant/004_add_product_category.up.sql
   ALTER TABLE products ADD COLUMN category_id UUID;
   ```

3. **Data Migrations**
   ```sql
   -- migrations/tenant/005_populate_categories.up.sql
   INSERT INTO categories (id, name) VALUES
   (gen_random_uuid(), 'Electronics'),
   (gen_random_uuid(), 'Books');
   ```

4. **Schema Changes**
   ```sql
   -- Handle with care, usually in transaction
   BEGIN;
   ALTER TABLE products 
   ADD COLUMN temp_status VARCHAR(50);
   
   UPDATE products 
   SET temp_status = 
       CASE status 
           WHEN 1 THEN 'active'
           WHEN 0 THEN 'inactive'
       END;
   
   ALTER TABLE products 
   DROP COLUMN status,
   RENAME COLUMN temp_status TO status;
   COMMIT;
   ```
