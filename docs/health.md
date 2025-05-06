# Health Module

The health module provides comprehensive health checking capabilities for your NebularCore application. It allows you to monitor the health of your application, its dependencies, and custom components.

## Configuration

The health module can be configured in your application's configuration file:

```yaml
modules:
  health:
    enabled: true           # Enable/disable health checks
    path: "/health"        # Base path for health endpoints
    interval: "30s"        # Default check interval
    timeout: "5s"          # Check timeout
    initialDelay: "5s"     # Delay before starting checks
```

## Features

1. System Health Monitoring
   - Runtime statistics (goroutines, memory)
   - Automatic system metrics collection
2. Database Connectivity Checks
   - Connection pool statistics
   - Ping tests
3. Custom Health Checks
   - Flexible check registration
   - Custom intervals
   - Detailed status reporting
4. Status Aggregation
   - Hierarchical status levels
   - Custom status determination
5. Multiple Endpoints
   - Liveness probe
   - Readiness probe
   - Detailed status
- Prometheus metrics integration

## Basic Usage

### 1. Register Module

```go
package main

import (
    "log"

    "gitlab.com/jideobs/nebularcore/core"
    "gitlab.com/jideobs/nebularcore/modules/health"
)

func main() {
    app, err := core.NewApp("config.yml")
    if err != nil {
        log.Fatal(err)
    }

    // Register health module
    app.RegisterModule(health.NewModule())

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Configure Module

```yaml
# config.yml
modules:
  health:
    enabled: true
    path: "/health"              # Health check endpoint path
    interval: "30s"             # Check interval
    timeout: "5s"               # Check timeout
    initialDelay: "5s"         # Initial delay before first check
    prometheus:
      enabled: true
      path: "/metrics"
```

## Health Checks

### 1. Default Checks

The health module includes several built-in checks:

- **System Resources**
  - CPU usage
  - Memory usage
  - Disk space
  - Network connectivity

- **Database**
  - Connection status
  - Query response time
  - Connection pool stats

- **Cache**
  - Redis connection
  - Cache hit rate
  - Memory usage

### 2. Custom Health Checks

You can add custom health checks to monitor specific components:

```go
package main

import (
    "context"
    "gitlab.com/jideobs/nebularcore/modules/health"
)

func main() {
    healthModule := health.NewModule()

    // Add custom check
    healthModule.AddCheck("external-api", func(ctx context.Context) error {
        // Check external API health
        return checkExternalAPI()
    })

    // Add check with details
    healthModule.AddCheckWithDetails("queue", func(ctx context.Context) (*health.CheckResult, error) {
        return &health.CheckResult{
            Status: "healthy",
            Details: map[string]interface{}{
                "queueSize": 42,
                "processRate": "100/s",
                "errorRate": "0.1%",
            },
        }, nil
    })
}
```

## API Reference

### Health Check Endpoint

```http
GET /health
```

Response:
```json
{
    "status": "healthy",
    "timestamp": "2025-05-02T22:41:10Z",
    "version": "1.0.0",
    "checks": {
        "system": {
            "status": "healthy",
            "details": {
                "cpu": "25%",
                "memory": "60%",
                "disk": "45%"
            }
        },
        "database": {
            "status": "healthy",
            "details": {
                "responseTime": "20ms",
                "connections": 5,
                "maxConnections": 20
            }
        },
        "external-api": {
            "status": "healthy",
            "details": {
                "latency": "150ms"
            }
        }
    }
}
```

### Prometheus Metrics

When enabled, metrics are available at `/metrics`:

```text
# HELP nebularcore_health_check_status Health check status (0=unhealthy, 1=healthy)
# TYPE nebularcore_health_check_status gauge
nebularcore_health_check_status{check="system"} 1
nebularcore_health_check_status{check="database"} 1

# HELP nebularcore_health_check_duration_seconds Health check duration in seconds
# TYPE nebularcore_health_check_duration_seconds histogram
nebularcore_health_check_duration_seconds_bucket{check="system",le="0.1"} 1
```

## Advanced Configuration

### 1. Check Dependencies

```go
healthModule.AddDependencyCheck("cache", []string{"redis"}, func(ctx context.Context) error {
    // Check cache health after redis check
    return checkCache()
})
```

### 2. Conditional Checks

```go
healthModule.AddConditionalCheck("backup", func(ctx context.Context) (bool, error) {
    // Only run during maintenance window
    return isMaintenanceWindow(), nil
}, func(ctx context.Context) error {
    return checkBackupSystem()
})
```

### 3. Custom Status Aggregation

```go
healthModule.SetStatusAggregator(func(checks map[string]*health.CheckResult) string {
    // Custom logic to determine overall status
    for _, check := range checks {
        if check.Status == "critical" {
            return "critical"
        }
    }
    return "healthy"
})
```

## Endpoints

The health module exposes three endpoints:

1. **GET /health/live**
   - Basic liveness check
   - Returns 200 if application is running
   - Response includes uptime and version

2. **GET /health/ready**
   - Readiness check with all health results
   - Returns 200 if healthy, 503 if unhealthy
   - Includes check results and status

3. **GET /health/status**
   - Detailed health status
   - Full check results with metrics
   - System and database statistics

Example response:
```json
{
    "status": "healthy",
    "uptime": "1h2m3s",
    "timestamp": "2025-05-03T00:19:55+01:00",
    "version": "1.0.0",
    "latency": "1.234ms",
    "checks": {
        "system": {
            "status": "healthy",
            "details": {
                "goroutines": 12,
                "memory": {
                    "alloc": 1234567,
                    "sys": 4567890
                }
            }
        },
        "database": {
            "status": "healthy",
            "details": {
                "openConnections": 5,
                "inUse": 2
            }
        }
    }
}
```

## Best Practices

1. **Check Design**
   - Keep checks lightweight and fast
   - Use appropriate intervals (default: 30s)
   - Set reasonable timeouts (default: 5s)
   - Handle transient failures gracefully

2. **Configuration**
   - Adjust intervals based on importance
   - Set appropriate initial delay
   - Configure timeout based on check complexity

3. **Monitoring**
   - Monitor /health/ready for overall health
   - Set up alerts for critical check failures
   - Track check latency and error rates
   - Use proper logging for failures

4. **Security**
   - Secure health endpoints in production
   - Consider rate limiting
   - Restrict sensitive information in responses
   - Limit detailed information in production
   - Use authentication for sensitive metrics

4. **Performance**
   - Cache check results when appropriate
   - Use concurrent checks
   - Implement circuit breakers
   - Monitor resource usage
