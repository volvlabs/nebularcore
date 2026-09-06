****# NebularCore Migration System

NebularCore uses [golang-migrate](https://github.com/golang-migrate/migrate) to provide a flexible migration system that supports both module-level and custom migrations, with proper ordering and dependency management.

## Core Concepts

### 1. Migration Sources

Migrations are managed through `Source` objects that define:

```go
type Source struct {
    Path     string   // Path to migration files
    Priority int      // Higher priority runs first
    Exclude  []string // Files to exclude
}
```

Each module can define multiple migration sources with different priorities, allowing for proper sequencing of migrations.

### 2. Migration Files

Migrations follow the golang-migrate format:
- Named as `{version}_{description}.{up|down}.sql`
- Paired up/down files for each migration
- Versions ensure correct ordering
- SQL-based for database schema changes

### 3. Migration Priorities

Migrations are executed in order of priority:
1. Higher priority sources run first (e.g., priority 100)
2. Lower priority sources run second (e.g., priority 50)
3. Within each source, migrations run in version order

## Implementation Examples

### 1. Basic Module Migrations

```go
// module/migrations/000001_init.up.sql
CREATE TABLE IF NOT EXISTS settings (
    id UUID PRIMARY KEY,
    key VARCHAR(255) UNIQUE,
    value TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

// module/migrations/000001_init.down.sql
DROP TABLE IF EXISTS settings;

// module/module.go
func (m *Module) GetMigrationSources(projectRoot string) []migrationRunner.Source {
    return []migrationRunner.Source{
        {
            Path:     fmt.Sprintf("file://%s/modules/mymodule/migrations", projectRoot),
            Priority: 50,
            Exclude:  []string{},
        },
    }
}
```

### 2. Multiple Migration Sources

```go
// For modules that need to handle both core and extension migrations
func (m *Module) GetMigrationSources(projectRoot string) []migrationRunner.Source {
    sources := []migrationRunner.Source{
        {
            // Core migrations run first
            Path:     fmt.Sprintf("file://%s/modules/mymodule/migrations/core", projectRoot),
            Priority: 100,
            Exclude:  []string{},
        },
        {
            // Extension migrations run second
            Path:     fmt.Sprintf("file://%s/modules/mymodule/migrations/extensions", projectRoot),
            Priority: 50,
            Exclude:  []string{},
        },
    }
    return sources
}
```

func (m *TenantModule) GetMigrations(ctx context.Context) ([]Migration, error) {
    // Base migrations in public schema
    migrations := []Migration{
        {
            Name:    "001_tenant_tables.up.sql",
            Version: 1,
            Module:  "tenant",
            IsUp:    true,
            Content: `
                CREATE TABLE tenants (
                    id UUID PRIMARY KEY,
                    code VARCHAR(50) UNIQUE,
                    schema_name VARCHAR(50) UNIQUE
                );
            `,
        },
    }
    
    // Add tenant-specific schema migrations
    tenantMigrations := []Migration{
        {
            Name:    "001_tenant_data.up.sql",
            Version: 1,
            Module:  "tenant",
            IsUp:    true,
            Content: `
                CREATE TABLE tenant_settings (
                    id UUID PRIMARY KEY,
                    tenant_id UUID NOT NULL,
                    key VARCHAR(100),
                    value TEXT
                );
            `,
        },
    }
    
    migrations = append(migrations, tenantMigrations...)
    return migrations, nil
}
```

## Project Organization

### Directory Structure

```
your-module/
├── migrations/
│   ├── public/
│   │   ├── 001_init.up.sql
│   │   └── 001_init.down.sql
│   └── tenant/
│       ├── 001_tenant_tables.up.sql
│       └── 001_tenant_tables.down.sql
├── module.go
└── migrations.go
```

### Migration Manager Usage

```go
func InitializeMigrations(app *Application) error {
    manager := &MigrationManager{
        modules:      []MigrationProvider{},
        projectDir:   app.ProjectDir(),
        customizeDir: filepath.Join(app.ProjectDir(), "migrations"),
    }
    
    // Register module migrations
    manager.RegisterModule(NewAuthModule())
    manager.RegisterModule(NewTenantModule())
    
    // Run migrations
    return manager.RunMigrations(context.Background())
}
```

## Best Practices

1. **Version Numbering**
   - Use sequential numbers (001, 002, etc.)
   - Include descriptive names
   - Keep versions unique within a module

2. **Migration Content**
   - Use SQL for database changes
   - Include both up and down migrations
   - Keep migrations idempotent
   - Use transactions where appropriate

3. **Module Organization**
   - Separate public and tenant migrations
   - Group related changes
   - Document dependencies
   - Include migration tests

4. **Customization**
   - Use `CustomizeMigration` for project-specific changes
   - Keep customizations minimal
   - Document required customizations
   - Provide example customizations

## Advanced Features

### 1. Dynamic Migrations

```go
func (m *Module) GetMigrations(ctx context.Context) ([]Migration, error) {
    migrations := []Migration{}
    
    // Read from filesystem
    files, err := os.ReadDir(m.MigrationDir())
    if err != nil {
        return nil, err
    }
    
    for _, file := range files {
        content, err := os.ReadFile(filepath.Join(m.MigrationDir(), file.Name()))
        if err != nil {
            return nil, err
        }
        
        migrations = append(migrations, Migration{
            Name:    file.Name(),
            Content: string(content),
            Module:  m.Name(),
            IsUp:    strings.HasSuffix(file.Name(), ".up.sql"),
        })
    }
    
    return migrations, nil
}
```

### 2. Conditional Migrations

```go
func (m *Module) GetMigrations(ctx context.Context) ([]Migration, error) {
    migrations := []Migration{}
    
    // Add base migrations
    migrations = append(migrations, m.getBaseMigrations()...)
    
    // Add feature-specific migrations
    if m.config.Features.EnableAudit {
        migrations = append(migrations, m.getAuditMigrations()...)
    }
    
    return migrations, nil
}
```

### 3. Migration Dependencies

```go
type MigrationDependency struct {
    Module  string
    Version uint
}

type Migration struct {
    // ... existing fields ...
    Dependencies []MigrationDependency
}

func (m *Module) GetMigrations(ctx context.Context) ([]Migration, error) {
    return []Migration{
        {
            Name:    "002_user_roles.up.sql",
            Version: 2,
            Dependencies: []MigrationDependency{
                {Module: "auth", Version: 1},  // Requires auth tables
                {Module: "tenant", Version: 1}, // Requires tenant tables
            },
            Content: `CREATE TABLE user_roles ...`,
        },
    }
}
```

## Common Patterns

1. **Schema Creation**
```sql
-- In up migration
CREATE SCHEMA IF NOT EXISTS {{.SchemaName}};

-- In down migration
DROP SCHEMA IF EXISTS {{.SchemaName}} CASCADE;
```

2. **Table Creation**
```sql
-- In up migration
CREATE TABLE IF NOT EXISTS {{.SchemaName}}.table_name (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- In down migration
DROP TABLE IF EXISTS {{.SchemaName}}.table_name;
```

3. **Data Migration**
```sql
-- In up migration
INSERT INTO {{.SchemaName}}.settings (key, value)
SELECT 'default_theme', 'light'
WHERE NOT EXISTS (
    SELECT 1 FROM {{.SchemaName}}.settings WHERE key = 'default_theme'
);

-- In down migration
DELETE FROM {{.SchemaName}}.settings WHERE key = 'default_theme';
```
