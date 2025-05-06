# Migration CLI Guide

NebularCore provides a powerful CLI for managing database migrations, supporting both public schema and tenant-specific migrations.

## Commands

```bash
nebularcore migrate [command] [args]
```

### Available Commands

1. **Run Public Schema Migrations**
   ```bash
   # Run all pending migrations
   nebularcore migrate up
   
   # Rollback last N migrations
   nebularcore migrate down N
   ```

2. **Tenant-Specific Migrations**
   ```bash
   # Run migrations for a specific tenant
   nebularcore migrate tenant [schema_name] [command]
   
   # Examples:
   nebularcore migrate tenant tenant_1 up
   nebularcore migrate tenant tenant_1 down 2
   ```

3. **All Tenants Migration**
   ```bash
   # Run migrations for all tenant schemas
   nebularcore migrate all-tenants
   ```

4. **Create New Migration**
   ```bash
   # Create new migration files
   nebularcore migrate create [migration_name]
   
   # Example:
   nebularcore migrate create add_user_table
   # Creates:
   # - [timestamp]_add_user_table.up.sql
   # - [timestamp]_add_user_table.down.sql
   ```

## Migration Directory Structure

```
your-project/
├── migrations/
│   ├── public/
│   │   ├── [timestamp]_init.up.sql
│   │   └── [timestamp]_init.down.sql
│   └── tenant/
│       ├── [timestamp]_products.up.sql
│       └── [timestamp]_products.down.sql
```

## Example Usage

### 1. Initialize Project with Public Schema

```bash
# Create initial migration
nebularcore migrate create init_public_schema

# Edit migrations/[timestamp]_init_public_schema.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE
);

# Run the migration
nebularcore migrate up
```

### 2. Create New Tenant

```bash
# Create tenant-specific tables
nebularcore migrate create tenant_tables

# Edit migrations/[timestamp]_tenant_tables.up.sql
CREATE TABLE products (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    tenant_id UUID NOT NULL
);

# Run for specific tenant
nebularcore migrate tenant tenant_1 up
```

### 3. Rollback Migrations

```bash
# Rollback last migration in public schema
nebularcore migrate down 1

# Rollback last 2 migrations in tenant schema
nebularcore migrate tenant tenant_1 down 2
```

## Implementation Details

The migration system uses:
- `golang-migrate` for migration execution
- PostgreSQL schema-based isolation
- Transactional migrations
- Version tracking per schema

### Migration Runner

The core migration functionality is implemented in `tools/migrate/Runner`, which provides:

1. **Schema Support**
   ```go
   runner, err := migrate.NewRunner(migrationsDir, connectionString)
   schemaRunner, err := runner.WithSchema("tenant_1")
   ```

2. **Flexible Commands**
   ```go
   // Run migrations up
   runner.Run("up")
   
   // Rollback specific number
   runner.Run("down", "2")
   ```

3. **Error Handling**
   - Migration failures are logged
   - Transactions ensure consistency
   - Schema-specific error reporting

## Best Practices

1. **Migration Organization**
   - Keep public/tenant migrations separate
   - Use descriptive migration names
   - Include both up/down migrations

2. **Schema Management**
   - Use `all-tenants` for bulk updates
   - Test migrations in development first
   - Back up before major migrations

3. **Error Recovery**
   - Keep rollback migrations updated
   - Test rollback procedures
   - Monitor migration logs

4. **Performance**
   - Run large migrations off-peak
   - Use batching for data migrations
   - Monitor lock contention

## Common Scenarios

### 1. Adding a New Feature

```bash
# Create migration
nebularcore migrate create add_feature_x

# Run for all tenants
nebularcore migrate all-tenants
```

### 2. Emergency Rollback

```bash
# Rollback specific tenant
nebularcore migrate tenant problem_tenant down 1

# Verify fix and apply to others
nebularcore migrate all-tenants
```

### 3. Schema Updates

```bash
# Create schema change
nebularcore migrate create update_product_schema

# Apply selectively
nebularcore migrate tenant test_tenant up
# Verify changes
nebularcore migrate all-tenants
```
