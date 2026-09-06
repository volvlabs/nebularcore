# Authentication Module

NebularCore's authentication module provides a flexible and secure user authentication system with support for multiple authentication backends and customizable user models.

## Features

- Multiple authentication backends (Local, OAuth, Social)
- JWT-based authentication
- Password policy enforcement
- Custom user models
- Migration system for user schemas
- OTP support
- API key authentication

## Basic Usage

### 1. Default User Model

By default, the auth module provides a base user model with essential fields:

```go
// Default User Model
type User struct {
    ID                UUID       `gorm:"primaryKey;type:uuid"`
    Email             string     `gorm:"uniqueIndex"`
    PhoneNumber       string     `gorm:"uniqueIndex"`
    Username          string     `gorm:"uniqueIndex"`
    Password          string
    Metadata          JSON
    Active            bool       `gorm:"default:true"`
    PasswordResetToken string
    Token             string
    CreatedAt         time.Time
    UpdatedAt         time.Time
    LastLoginAt       *time.Time
    PasswordResetAt   *time.Time
    DeletedAt         *time.Time
}
```

### 2. Configure Default Setup

```yaml
# config.yml
modules:
  auth:
    backends:
      - local
    jwt:
      secret: your-secret-key
      expiry: 24h
    passwordPolicy:
      minLength: 8
      requireUpper: true
      requireLower: true
      requireNumber: true
      requireSpecial: true
```

### 3. Default Authentication Flow

1. **Register**
   ```http
   POST /api/auth/register
   Content-Type: application/json
   
   {
     "email": "user@example.com",
     "password": "securePassword123!"
   }
   ```

2. **Login**
   ```http
   POST /api/auth/login
   Content-Type: application/json
   
   {
     "email": "user@example.com",
     "password": "securePassword123!"
   }
   ```

## Custom User Models

For applications requiring additional user fields, you can extend the base user model:

### 1. Define Custom Model

```go
// models/user.go
type CustomUser struct {
    models.User        // Embed base user model
    FirstName    string
    LastName     string
    DateOfBirth  *time.Time
    Address      string
    CompanyName  string
    Department   string
    Role         string
}
```

### 2. Create Custom Migration

```sql
-- migrations/000001_custom_user.up.sql
CREATE TABLE IF NOT EXISTS users (
    -- Base fields
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    phone_number VARCHAR(255) UNIQUE,
    username VARCHAR(255) UNIQUE,
    password VARCHAR(255),
    metadata JSONB,
    active BOOLEAN DEFAULT true,
    password_reset_token VARCHAR(255),
    token VARCHAR(255),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    last_login_at TIMESTAMP,
    password_reset_at TIMESTAMP,
    deleted_at TIMESTAMP,
    -- Custom fields
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    date_of_birth TIMESTAMP,
    address TEXT,
    company_name VARCHAR(255),
    department VARCHAR(255),
    role VARCHAR(255)
);

-- migrations/000001_custom_user.down.sql
DROP TABLE IF EXISTS users;
```

### 3. Configure Custom Setup

```yaml
# config.yml
modules:
  auth:
    userMigrationScriptPath: "./migrations"  # Point to custom migrations
    backends:
      - local
    jwt:
      secret: your-secret-key
      expiry: 24h
    passwordPolicy:
      minLength: 8
      requireUpper: true
      requireLower: true
      requireNumber: true
      requireSpecial: true
```

### 4. Custom Authentication Flow
1. **Login (Same as Default)**
   ```http
   POST /api/auth/login
   Content-Type: application/json
   
   {
     "email": "user@example.com",
     "password": "securePassword123!"
   }
   ```

2. **Get User Profile (Returns Custom Fields)**
   ```http
   GET /api/auth/me
   Authorization: Bearer <jwt_token>
   ```

## API Reference

### Authentication Endpoints

- `POST /auth/login` - Login with credentials
- `POST /auth/logout` - Logout current user
- `POST /auth/refresh` - Refresh JWT token
- `POST /auth/password/reset` - Request password reset
- `POST /auth/password/change` - Change password

### Social Authentication

- `GET /auth/social/{provider}` - Initiate social login
- `GET /auth/social/{provider}/callback` - Social login callback

### API Key Authentication

- `POST /auth/apikey` - Generate API key
- `DELETE /auth/apikey/{id}` - Revoke API key
- `GET /auth/apikey` - List API keys
