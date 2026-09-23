package domain

import "time"

// Permission represents an atomic platform permission.
type Permission struct {
	Name        string
	Description string
	Resource    string
}

// Role represents a tenant-scoped role.
type Role struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	IsSystem    bool
	Permissions []string
	CreatedAt   time.Time
}

// PolicyDecision is the result of a policy evaluation.
type PolicyDecision struct {
	Allowed bool
	Reason  string
}
