# NebularCore

Project name _"Suggesting a vast, foundational core like a nebula, symbolizing the birthplace of new projects."_

NebularCore is a modular Go backend framework created with inspiration from [PocketBase](https://pocketbase.io). It provides a robust foundation for building scalable backend services with features like authentication, health monitoring, database migrations, and more.

## Features

- 🔐 Authentication with JWT, OTP, and social providers
- 🏥 Health monitoring and system status
- 📦 Database support for PostgreSQL and SQLite
- 🔄 Flexible migration system
- 🛠️ Modular architecture
- 🌐 Built on Gin web framework

## Quick Start

```go
package main

import (
    "log"

    "github.com/volvlabs/nebularcore/core"
    "github.com/volvlabs/nebularcore/modules/auth"
    "github.com/volvlabs/nebularcore/modules/health"
)

func main() {
    // Initialize app with config
    app, err := core.NewApp("config.yml")
    if err != nil {
        log.Fatal(err)
    }

    // Register modules
    app.RegisterModule(auth.NewModule())
    app.RegisterModule(health.NewModule())

    // Start the server
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

```yaml
# config.yml
core:
  environment: development
  database:
    driver: postgres
    host: localhost
    username: postgres
    password: password
    name: nebularcore
    port: "5432"
    sslmode: disable
  server:
    port: "8888"
    host: localhost

modules:
    moduleA:
        ...
    moduleB:
        ...
```

## Documentation

- [Migrations Guide](docs/migrations.md)
- [Module System](docs/modules.md)
- [Authentication](docs/auth.md)
- [Examples](examples/)

## License

(c) Volvlabs 2023 - 2024
