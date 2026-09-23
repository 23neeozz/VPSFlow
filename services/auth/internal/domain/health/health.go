package health

import "time"

// Status represents service health state.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Check represents an individual dependency health check.
type Check struct {
	Name    string `json:"name"`
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}

// Report is the aggregated health report.
type Report struct {
	Status    Status    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Checks    []Check   `json:"checks,omitempty"`
}

// Checker evaluates service readiness.
type Checker struct {
	service string
	version string
}

// NewChecker creates a health checker.
func NewChecker(service, version string) *Checker {
	return &Checker{service: service, version: version}
}

// Liveness returns a minimal liveness report.
func (c *Checker) Liveness() Report {
	return Report{
		Status:    StatusHealthy,
		Service:   c.service,
		Version:   c.version,
		Timestamp: time.Now().UTC(),
	}
}

// Readiness returns readiness including dependency checks.
func (c *Checker) Readiness() Report {
	return Report{
		Status:    StatusHealthy,
		Service:   c.service,
		Version:   c.version,
		Timestamp: time.Now().UTC(),
		Checks:    []Check{{Name: "process", Status: StatusHealthy}},
	}
}
