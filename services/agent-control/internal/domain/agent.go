package domain

import "time"

const (
	SessionActive = "active"
	SessionClosed = "closed"

	CommandPending   = "pending"
	CommandRunning   = "running"
	CommandCompleted = "completed"
	CommandFailed    = "failed"
)

type Session struct {
	ID              string    `json:"session_id"`
	HypervisorID    string    `json:"hypervisor_id"`
	AgentVersion    string    `json:"agent_version"`
	NodeName        string    `json:"node_name"`
	Status          string    `json:"status"`
	LastHeartbeatAt time.Time `json:"last_heartbeat_at"`
	CreatedAt       time.Time `json:"created_at"`
}

type Command struct {
	ID           string     `json:"command_id"`
	HypervisorID string     `json:"hypervisor_id"`
	TenantID     string     `json:"tenant_id"`
	Type         string     `json:"type"`
	Payload      []byte     `json:"payload"`
	Status       string     `json:"status"`
	Result       []byte     `json:"result,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}
