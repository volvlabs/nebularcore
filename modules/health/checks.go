package health

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"gorm.io/gorm"
)

// AddCheck adds a new health check
func (m *Module) AddCheck(name string, check *Check) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.checks[name]; exists {
		return fmt.Errorf("check %s already exists", name)
	}

	m.checks[name] = check
	return nil
}

// RemoveCheck removes a health check
func (m *Module) RemoveCheck(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.checks, name)
	delete(m.results, name)
}

// systemCheck returns a check for system health
func (m *Module) systemCheck() *Check {
	return NewCheck("system", func(ctx context.Context) (*CheckResult, error) {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		return &CheckResult{
			Name:   "system",
			Status: StatusHealthy,
			Details: map[string]interface{}{
				"goroutines": runtime.NumGoroutine(),
				"memory": map[string]interface{}{
					"alloc":      memStats.Alloc,
					"totalAlloc": memStats.TotalAlloc,
					"sys":        memStats.Sys,
					"numGC":      memStats.NumGC,
				},
			},
			Timestamp: time.Now(),
		}, nil
	}).WithInterval(time.Minute)
}

// databaseCheck returns a check for database connectivity
func (m *Module) databaseCheck(db *gorm.DB) *Check {
	return NewCheck("database", func(ctx context.Context) (*CheckResult, error) {
		start := time.Now()
		sqlDB, err := db.DB()
		if err != nil {
			return &CheckResult{
				Name:      "database",
				Status:    StatusUnhealthy,
				Error:     err.Error(),
				Timestamp: time.Now(),
			}, nil
		}

		err = sqlDB.PingContext(ctx)
		duration := time.Since(start)

		if err != nil {
			return &CheckResult{
				Name:      "database",
				Status:    StatusUnhealthy,
				Error:     err.Error(),
				Duration:  duration,
				Timestamp: time.Now(),
			}, nil
		}

		stats := sqlDB.Stats()
		return &CheckResult{
			Name:   "database",
			Status: StatusHealthy,
			Details: map[string]interface{}{
				"openConnections": stats.OpenConnections,
				"inUse":          stats.InUse,
				"idle":           stats.Idle,
				"maxOpen":        stats.MaxOpenConnections,
				"waitCount":      stats.WaitCount,
				"waitDuration":   stats.WaitDuration,
			},
			Duration:  duration,
			Timestamp: time.Now(),
		}, nil
	}).WithInterval(30 * time.Second)
}

// runChecks starts the health check scheduler
func (m *Module) runChecks() {
	if !m.config.Enabled {
		return
	}

	// Initial delay before starting checks
	time.Sleep(m.config.InitialDelay)

	ticker := time.NewTicker(m.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.shutdown:
			return
		case <-ticker.C:
			m.executeChecks()
		}
	}
}

// executeChecks runs all registered health checks
func (m *Module) executeChecks() {
	m.mu.RLock()
	checks := make(map[string]*Check)
	for name, check := range m.checks {
		checks[name] = check
	}
	m.mu.RUnlock()

	for name, check := range checks {
		if check.Interval > 0 {
			// Skip if custom interval hasn't elapsed
			m.mu.RLock()
			lastResult, exists := m.results[name]
			m.mu.RUnlock()
			if exists && time.Since(lastResult.Timestamp) < check.Interval {
				continue
			}
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
		result, err := check.Run(ctx)
		cancel()

		if err != nil {
			result = &CheckResult{
				Name:      name,
				Status:    StatusUnhealthy,
				Error:     err.Error(),
				Timestamp: time.Now(),
			}
		}

		m.mu.Lock()
		m.results[name] = result
		m.mu.Unlock()
	}
}

// aggregateStatus determines the overall system status based on check results
func (m *Module) aggregateStatus(checks map[string]*CheckResult) string {
	if len(checks) == 0 {
		return StatusHealthy
	}

	hasDegraded := false
	for _, result := range checks {
		switch result.Status {
		case StatusCritical:
			return StatusCritical
		case StatusUnhealthy:
			return StatusUnhealthy
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}
