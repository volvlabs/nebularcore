package config

import (
	"fmt"
	"time"
)

// Config holds the WebSocket module configuration.
type Config struct {
	Enabled    bool           `yaml:"enabled"`
	Server     ServerConfig   `yaml:"server"`
	Routing    RoutingConfig  `yaml:"routing"`
	Security   SecurityConfig `yaml:"security"`
	TenantMode string         `yaml:"tenantMode"`
	Events     EventsConfig   `yaml:"events"`
}

// ServerConfig holds WebSocket server settings.
type ServerConfig struct {
	Host                    string        `yaml:"host"`
	Port                    string        `yaml:"port"`
	ReadBufferSize          int           `yaml:"readBufferSize"`
	WriteBufferSize         int           `yaml:"writeBufferSize"`
	ReadDeadline            time.Duration `yaml:"readDeadline"`
	WriteDeadline           time.Duration `yaml:"writeDeadline"`
	MaxConnections          int64         `yaml:"maxConnections"`
	MaxConnectionsPerUser   int           `yaml:"maxConnectionsPerUser"`
	MaxConnectionsPerTenant int           `yaml:"maxConnectionsPerTenant"`
}

// RoutingConfig holds message routing settings.
type RoutingConfig struct {
	MaxTopicLength         int `yaml:"maxTopicLength"`
	MaxTopicsPerConnection int `yaml:"maxTopicsPerConnection"`
}

// SecurityConfig holds WebSocket security settings.
type SecurityConfig struct {
	AuthRequired bool     `yaml:"authRequired"`
	AllowOrigins []string `yaml:"allowOrigins"`
	JWTSecret    string   `yaml:"jwtSecret"`
}

// EventsConfig holds event bridging settings.
type EventsConfig struct {
	AllowedEventTypes   []string `yaml:"allowedEventTypes"`
	InternalEventPrefix string   `yaml:"internalEventPrefix"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Enabled: false,
		Server: ServerConfig{
			Host:                    "localhost",
			Port:                    "8080",
			ReadBufferSize:          4096,
			WriteBufferSize:         4096,
			ReadDeadline:            60 * time.Second,
			WriteDeadline:           60 * time.Second,
			MaxConnections:          100000,
			MaxConnectionsPerUser:   10,
			MaxConnectionsPerTenant: 50000,
		},
		Routing: RoutingConfig{
			MaxTopicLength:         256,
			MaxTopicsPerConnection: 100,
		},
		Security: SecurityConfig{
			AuthRequired: true,
		},
		TenantMode: "header",
		Events: EventsConfig{
			InternalEventPrefix: "ws:",
		},
	}
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Server.Port == "" {
		return fmt.Errorf("websocket: server.port is required")
	}
	if c.Server.MaxConnections <= 0 {
		return fmt.Errorf("websocket: server.maxConnections must be positive")
	}
	if c.Server.MaxConnectionsPerUser <= 0 {
		return fmt.Errorf("websocket: server.maxConnectionsPerUser must be positive")
	}
	if c.Server.MaxConnectionsPerTenant <= 0 {
		return fmt.Errorf("websocket: server.maxConnectionsPerTenant must be positive")
	}
	if c.Routing.MaxTopicLength <= 0 {
		return fmt.Errorf("websocket: routing.maxTopicLength must be positive")
	}
	if c.Routing.MaxTopicsPerConnection <= 0 {
		return fmt.Errorf("websocket: routing.maxTopicsPerConnection must be positive")
	}
	switch c.TenantMode {
	case "header", "path", "query":
	default:
		return fmt.Errorf("websocket: tenantMode must be one of: header, path, query")
	}
	return nil
}
