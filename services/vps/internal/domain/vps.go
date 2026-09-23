package domain

import "time"

type Instance struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	VMID             string     `json:"vm_id"`
	Name             string     `json:"name"`
	OwnerUserID      string     `json:"owner_user_id"`
	OwnerEmail       string     `json:"owner_email"`
	AssignedByUserID string     `json:"assigned_by_user_id"`
	Status           string     `json:"status"`
	VCPUs            int        `json:"vcpus"`
	MemoryMB         int        `json:"memory_mb"`
	DiskGB           int        `json:"disk_gb"`
	ErrorMessage     string     `json:"error_message,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type ConsoleSession struct {
	SessionID string `json:"session_id"`
	Token     string `json:"token"`
	ProxyURL  string `json:"proxy_url"`
	ExpiresAt string `json:"expires_at"`
}
