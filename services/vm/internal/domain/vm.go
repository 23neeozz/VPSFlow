package domain

import "time"

const (
	StatusPending      = "pending"
	StatusProvisioning = "provisioning"
	StatusRunning      = "running"
	StatusStopping     = "stopping"
	StatusStopped      = "stopped"
	StatusStarting     = "starting"
	StatusDeleting     = "deleting"
	StatusDeleted      = "deleted"
	StatusError        = "error"
)

type VirtualMachine struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	ProjectID    string     `json:"project_id,omitempty"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	HypervisorID string     `json:"hypervisor_id,omitempty"`
	DomainName   string     `json:"domain_name,omitempty"`
	VCPUs        int        `json:"vcpus"`
	MemoryMB     int        `json:"memory_mb"`
	DiskGB       int        `json:"disk_gb"`
	ImageRef     string     `json:"image_ref,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type Operation struct {
	ID             string    `json:"id"`
	VMID           string    `json:"vm_id"`
	TenantID       string    `json:"tenant_id"`
	OperationType  string    `json:"operation_type"`
	IdempotencyKey string    `json:"idempotency_key"`
	Status         string    `json:"status"`
	CommandID      string    `json:"command_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
