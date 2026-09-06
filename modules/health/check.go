package health

import (
	"context"
	"time"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time             `json:"timestamp"`
	Duration  time.Duration         `json:"duration"`
}

// Check represents a health check function
type Check struct {
	Name        string
	Description string
	Run         func(ctx context.Context) (*CheckResult, error)
	Interval    time.Duration
	Timeout     time.Duration
	Dependencies []string
}

// CheckStatus represents possible health check statuses
const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
	StatusDegraded  = "degraded"
	StatusCritical  = "critical"
)

// NewCheck creates a new health check
func NewCheck(name string, run func(ctx context.Context) (*CheckResult, error)) *Check {
	return &Check{
		Name: name,
		Run:  run,
	}
}

// WithDescription adds a description to the check
func (c *Check) WithDescription(description string) *Check {
	c.Description = description
	return c
}

// WithInterval sets the check interval
func (c *Check) WithInterval(interval time.Duration) *Check {
	c.Interval = interval
	return c
}

// WithTimeout sets the check timeout
func (c *Check) WithTimeout(timeout time.Duration) *Check {
	c.Timeout = timeout
	return c
}

// WithDependencies sets the check dependencies
func (c *Check) WithDependencies(dependencies ...string) *Check {
	c.Dependencies = dependencies
	return c
}
