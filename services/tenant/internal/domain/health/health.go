package health

import "time"

type Status string

const StatusHealthy Status = "healthy"

type Report struct {
	Status Status `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

type Checker struct{ service, version string }

func NewChecker(service, version string) *Checker { return &Checker{service: service, version: version} }
func (c *Checker) Liveness() Report { return Report{Status: StatusHealthy, Service: c.service, Version: c.version, Timestamp: time.Now().UTC()} }
func (c *Checker) Readiness() Report { return c.Liveness() }
