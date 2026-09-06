package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// handleStatus handles detailed health status requests
func (m *Module) handleStatus(c *gin.Context) {
	start := time.Now()

	m.mu.RLock()
	checks := make(map[string]*CheckResult)
	for name, result := range m.results {
		checks[name] = result
	}
	m.mu.RUnlock()

	status := HealthStatus{
		Status:    m.aggregateStatus(checks),
		Uptime:    time.Since(m.startTime).String(),
		Timestamp: time.Now(),
		Version:   m.version,
		Latency:   time.Since(start),
		Checks:    checks,
	}

	if status.Status == StatusHealthy {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusServiceUnavailable, status)
	}
}
